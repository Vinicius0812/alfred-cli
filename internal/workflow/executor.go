package workflow

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/Vinicius0812/alfred-cli/internal/config"
)

type ConfirmFunc func(PlannedStep) (bool, error)

type ExecuteOptions struct {
	DryRun        bool
	AssumeYes     bool
	Stdout        io.Writer
	Stderr        io.Writer
	Environ       []string
	Confirm       ConfirmFunc
	Now           func() time.Time
	AlfredVersion string
}

func Execute(ctx context.Context, plan Plan, opts ExecuteOptions) (Execution, error) {
	if opts.Stdout == nil {
		opts.Stdout = io.Discard
	}
	if opts.Stderr == nil {
		opts.Stderr = io.Discard
	}
	if opts.Environ == nil {
		opts.Environ = os.Environ()
	}
	if opts.Now == nil {
		opts.Now = time.Now
	}
	secrets := executionSecrets(plan.secretValues, opts.Environ)

	execution := Execution{
		ID:         plan.ID,
		Workflow:   plan.Workflow,
		ConfigFile: plan.ConfigFile,
		ConfigHash: plan.ConfigHash,
		Version:    opts.AlfredVersion,
		DryRun:     opts.DryRun,
		StartedAt:  opts.Now().UTC(),
		Outcome:    "running",
	}

	finish := func(outcome string, runErr error) (Execution, error) {
		execution.Outcome = outcome
		execution.FinishedAt = opts.Now().UTC()
		if runErr != nil {
			execution.Error = redact(runErr.Error(), secrets)
		}
		if plan.Policies.Audit {
			if auditErr := WriteAudit(plan.ProjectRoot, execution); auditErr != nil {
				if runErr != nil {
					return execution, fmt.Errorf("%v; audit failed: %w", runErr, auditErr)
				}
				return execution, auditErr
			}
		}
		return execution, runErr
	}

	if opts.DryRun {
		for _, step := range plan.Steps {
			execution.Steps = append(execution.Steps, StepResult{
				Index:      step.Index,
				Name:       step.Name,
				Risk:       step.Risk,
				StartedAt:  execution.StartedAt,
				FinishedAt: execution.StartedAt,
				Outcome:    "planned",
			})
		}
		return finish("planned", nil)
	}

	hadContinuedFailure := false
	cleanWorktreeChecked := false
	for _, step := range plan.Steps {
		stepResult := StepResult{
			Index:     step.Index,
			Name:      step.Name,
			Risk:      step.Risk,
			StartedAt: opts.Now().UTC(),
			ExitCode:  -1,
			Outcome:   "running",
		}

		if step.ShellCommand != "" && !plan.Policies.AllowShellSteps {
			runErr := fmt.Errorf("shell steps are disabled by policy")
			stepResult.Outcome = "blocked"
			stepResult.Error = runErr.Error()
			stepResult.FinishedAt = opts.Now().UTC()
			execution.Steps = append(execution.Steps, stepResult)
			return finish("blocked", runErr)
		}

		if step.RequiresConfirmation {
			if opts.AssumeYes {
				if !plan.Policies.AllowNonInteractive {
					runErr := fmt.Errorf("--yes is blocked by policies.allow_non_interactive")
					stepResult.Outcome = "blocked"
					stepResult.Error = runErr.Error()
					stepResult.FinishedAt = opts.Now().UTC()
					execution.Steps = append(execution.Steps, stepResult)
					return finish("blocked", runErr)
				}
				stepResult.Confirmed = true
				stepResult.AutoConfirmed = true
			} else {
				if opts.Confirm == nil {
					runErr := fmt.Errorf("step requires interactive confirmation")
					stepResult.Outcome = "blocked"
					stepResult.Error = runErr.Error()
					stepResult.FinishedAt = opts.Now().UTC()
					execution.Steps = append(execution.Steps, stepResult)
					return finish("blocked", runErr)
				}
				confirmed, err := opts.Confirm(step)
				if err != nil {
					stepResult.Outcome = "failed"
					stepResult.Error = redact(err.Error(), secrets)
					stepResult.FinishedAt = opts.Now().UTC()
					execution.Steps = append(execution.Steps, stepResult)
					return finish("failed", err)
				}
				if !confirmed {
					stepResult.Outcome = "declined"
					stepResult.Error = ErrDeclined.Error()
					stepResult.FinishedAt = opts.Now().UTC()
					execution.Steps = append(execution.Steps, stepResult)
					return finish("declined", ErrDeclined)
				}
				stepResult.Confirmed = true
			}
		}

		if step.Risk == config.RiskGitWrite && plan.Policies.RequireCleanWorktree && !cleanWorktreeChecked {
			if err := requireCleanWorktree(ctx, plan.ProjectRoot, opts.Environ); err != nil {
				stepResult.Outcome = "blocked"
				stepResult.Error = err.Error()
				stepResult.FinishedAt = opts.Now().UTC()
				execution.Steps = append(execution.Steps, stepResult)
				return finish("blocked", err)
			}
			cleanWorktreeChecked = true
		}

		fmt.Fprintf(opts.Stdout, "[%d/%d] %s\n", step.Index, len(plan.Steps), step.Name)
		runErr, exitCode := executeStep(ctx, step, opts, secrets)
		stepResult.ExitCode = exitCode
		stepResult.FinishedAt = opts.Now().UTC()
		stepResult.Duration = stepResult.FinishedAt.Sub(stepResult.StartedAt)
		if runErr == nil {
			stepResult.Outcome = "success"
			execution.Steps = append(execution.Steps, stepResult)
			continue
		}

		stepResult.Outcome = "failed"
		stepResult.Error = redact(runErr.Error(), secrets)
		execution.Steps = append(execution.Steps, stepResult)
		if ctx.Err() != nil {
			return finish("failed", runErr)
		}
		if step.ContinueOnError {
			hadContinuedFailure = true
			fmt.Fprintf(opts.Stderr, "step failed and execution will continue: %s\n", stepResult.Error)
			continue
		}
		return finish("failed", runErr)
	}

	if hadContinuedFailure {
		return finish("success_with_warnings", nil)
	}
	return finish("success", nil)
}

func executionSecrets(configured []string, environ []string) []string {
	seen := map[string]bool{}
	secrets := make([]string, 0, len(configured))
	for _, value := range configured {
		if value != "" && !seen[value] {
			seen[value] = true
			secrets = append(secrets, value)
		}
	}
	for _, entry := range environ {
		index := strings.IndexByte(entry, '=')
		if index <= 0 || index == len(entry)-1 || !isSensitiveEnvironmentName(entry[:index]) {
			continue
		}
		value := entry[index+1:]
		if !seen[value] {
			seen[value] = true
			secrets = append(secrets, value)
		}
	}
	sort.Slice(secrets, func(i, j int) bool { return len(secrets[i]) > len(secrets[j]) })
	return secrets
}

func isSensitiveEnvironmentName(name string) bool {
	upper := strings.ToUpper(name)
	for _, marker := range []string{
		"TOKEN",
		"SECRET",
		"PASSWORD",
		"PASSWD",
		"API_KEY",
		"ACCESS_KEY",
		"PRIVATE_KEY",
		"CREDENTIAL",
		"BEARER",
	} {
		if strings.Contains(upper, marker) {
			return true
		}
	}
	return false
}

func executeStep(parent context.Context, step PlannedStep, opts ExecuteOptions, secrets []string) (error, int) {
	timeout, err := time.ParseDuration(step.Timeout)
	if err != nil {
		return err, -1
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	var command *exec.Cmd
	if step.ShellCommand != "" {
		if runtime.GOOS == "windows" {
			command = exec.CommandContext(ctx, "cmd.exe", "/D", "/S", "/C", step.ShellCommand)
		} else {
			command = exec.CommandContext(ctx, "/bin/sh", "-c", step.ShellCommand)
		}
	} else {
		command = exec.CommandContext(ctx, step.Command, step.Args...)
	}
	command.Dir = step.WorkingDirectory
	command.Env = mergeEnvironment(opts.Environ, step.environment)
	stdout := newRedactingWriter(opts.Stdout, secrets)
	stderr := newRedactingWriter(opts.Stderr, secrets)
	command.Stdout = stdout
	command.Stderr = stderr

	err = command.Run()
	stdoutErr := stdout.Flush()
	stderrErr := stderr.Flush()
	if err == nil && stdoutErr != nil {
		err = stdoutErr
	}
	if err == nil && stderrErr != nil {
		err = stderrErr
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return fmt.Errorf("step timed out after %s", step.Timeout), -1
	}
	if errors.Is(ctx.Err(), context.Canceled) {
		return fmt.Errorf("step canceled"), -1
	}
	if err == nil {
		return nil, 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return fmt.Errorf("command exited with code %d", exitErr.ExitCode()), exitErr.ExitCode()
	}
	return redactError(err, secrets), -1
}

func requireCleanWorktree(ctx context.Context, projectRoot string, environ []string) error {
	command := exec.CommandContext(ctx, "git", "-C", projectRoot, "status", "--porcelain")
	command.Env = environ
	output, err := command.Output()
	if err != nil {
		return fmt.Errorf("could not inspect Git worktree: %w", err)
	}
	if len(strings.TrimSpace(string(output))) > 0 {
		return fmt.Errorf("Git worktree is not clean")
	}
	return nil
}

func mergeEnvironment(base []string, overrides map[string]string) []string {
	type environmentValue struct {
		key   string
		value string
	}
	values := map[string]environmentValue{}
	for _, entry := range base {
		if index := strings.IndexByte(entry, '='); index > 0 {
			key := entry[:index]
			values[canonicalEnvironmentKey(key)] = environmentValue{key: key, value: entry[index+1:]}
		}
	}
	for key, value := range overrides {
		values[canonicalEnvironmentKey(key)] = environmentValue{key: key, value: value}
	}
	keys := make([]string, 0, len(values))
	for canonicalKey := range values {
		keys = append(keys, canonicalKey)
	}
	sort.Strings(keys)
	merged := make([]string, 0, len(keys))
	for _, canonicalKey := range keys {
		entry := values[canonicalKey]
		merged = append(merged, entry.key+"="+entry.value)
	}
	return merged
}

func canonicalEnvironmentKey(key string) string {
	if runtime.GOOS == "windows" {
		return strings.ToUpper(key)
	}
	return key
}

func redactError(err error, secrets []string) error {
	if err == nil {
		return nil
	}
	return errors.New(redact(err.Error(), secrets))
}
