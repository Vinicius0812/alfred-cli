package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Vinicius0812/alfred-cli/internal/workflow"
)

type runOptions struct {
	dryRun  bool
	yes     bool
	timeout time.Duration
	output  outputFormat
}

func parseRunOptions(args []string) (runOptions, []string, error) {
	opts := runOptions{output: outputText}
	remaining := make([]string, 0, len(args))
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--dry-run":
			opts.dryRun = true
		case "--yes", "-y":
			opts.yes = true
		case "--timeout", "--output":
			flag := args[index]
			if index+1 >= len(args) {
				return opts, nil, fmt.Errorf("%s requires a value", flag)
			}
			index++
			if flag == "--output" {
				opts.output = outputFormat(args[index])
				if opts.output != outputText && opts.output != outputJSON {
					return opts, nil, fmt.Errorf("output must be text or json")
				}
				continue
			}
			duration, err := time.ParseDuration(args[index])
			if err != nil || duration <= 0 {
				return opts, nil, fmt.Errorf("timeout must be a positive Go duration such as 30s or 5m")
			}
			opts.timeout = duration
		default:
			remaining = append(remaining, args[index])
		}
	}
	return opts, remaining, nil
}

func (a *application) runWorkflowExecution(args []string, global globalOptions, msg messages) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Fprintln(a.stdout, msg.text("run_usage"))
		return exitSuccess
	}
	if args[0] == "show" || args[0] == "list" {
		return a.runAuditCommand(args, global, msg)
	}

	opts, remaining, err := parseRunOptions(args)
	if err != nil || len(remaining) != 1 {
		if err != nil {
			fmt.Fprintf(a.stderr, "alfred: %s\n", localizeDiagnostic(err.Error(), global.lang))
		} else {
			fmt.Fprintln(a.stderr, msg.text("run_usage"))
		}
		return exitUsage
	}
	workflowName := remaining[0]
	result, ok := a.loadConfig(global, msg)
	if !ok {
		return exitFailure
	}
	plan, err := workflow.BuildPlan(result, workflowName, workflow.PlanOptions{
		LookupEnv: a.lookupEnv,
		Now:       a.now,
	})
	if errors.Is(err, workflow.ErrWorkflowNotFound) {
		fmt.Fprintln(a.stderr, msg.text("run_not_found", workflowName))
		return exitUsage
	}
	if err != nil {
		fmt.Fprintf(a.stderr, "alfred: %s\n", localizeDiagnostic("could not build workflow plan: "+err.Error(), global.lang))
		return exitFailure
	}
	if opts.timeout > 0 {
		for index := range plan.Steps {
			plan.Steps[index].Timeout = opts.timeout.String()
		}
	}

	if opts.output == outputText {
		printPlan(a.stdout, plan, msg)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	commandOutput := a.stdout
	confirmationOutput := a.stdout
	if opts.output == outputJSON {
		commandOutput = a.stderr
		confirmationOutput = a.stderr
	}
	execution, runErr := workflow.Execute(ctx, plan, workflow.ExecuteOptions{
		DryRun:        opts.dryRun,
		AssumeYes:     opts.yes,
		Stdout:        commandOutput,
		Stderr:        a.stderr,
		Environ:       a.environ(),
		Confirm:       a.confirmStep(msg, confirmationOutput),
		Now:           a.now,
		AlfredVersion: Version,
	})
	if opts.output == outputJSON {
		_ = writeJSON(a.stdout, struct {
			Plan      workflow.Plan      `json:"plan"`
			Execution workflow.Execution `json:"execution"`
		}{Plan: plan, Execution: execution})
	} else {
		fmt.Fprintln(a.stdout, msg.text("run_finished", execution.ID, localizeOutcome(execution.Outcome, global.lang)))
		if plan.Policies.Audit {
			fmt.Fprintln(a.stdout, msg.text("run_audit", execution.ID))
		}
	}
	if runErr == nil {
		return exitSuccess
	}
	fmt.Fprintln(a.stderr, msg.text("run_error", localizeDiagnostic(runErr.Error(), global.lang)))
	if errors.Is(ctx.Err(), context.Canceled) {
		return exitCanceled
	}
	return exitFailure
}

func printPlan(w interface{ Write([]byte) (int, error) }, plan workflow.Plan, msg messages) {
	if len(plan.Steps) == 1 {
		fmt.Fprintln(w, msg.text("run_plan_one", plan.Workflow))
	} else {
		fmt.Fprintln(w, msg.text("run_plan", plan.Workflow, len(plan.Steps)))
	}
	for _, step := range plan.Steps {
		command := step.Command
		if command != "" {
			arguments := make([]string, 0, len(step.Args)+1)
			arguments = append(arguments, command)
			for _, argument := range step.Args {
				arguments = append(arguments, strconv.Quote(argument))
			}
			command = strings.Join(arguments, " ")
		} else {
			command = msg.text("run_shell")
		}
		fmt.Fprintln(w, msg.text("run_step", step.Index, step.Name, step.Risk, command, step.Timeout))
	}
}

func (a *application) confirmStep(msg messages, output io.Writer) workflow.ConfirmFunc {
	return func(step workflow.PlannedStep) (bool, error) {
		fmt.Fprint(output, msg.text("run_confirm", step.Index, step.Name, localizeConfirmationReason(step.ConfirmationReason, msg.lang)))
		answer, err := a.reader.ReadString('\n')
		if err != nil && strings.TrimSpace(answer) == "" {
			return false, err
		}
		switch strings.ToLower(strings.TrimSpace(answer)) {
		case "y", "yes", "s", "sim":
			return true, nil
		default:
			return false, nil
		}
	}
}

func (a *application) runAuditCommand(args []string, global globalOptions, msg messages) int {
	format, remaining, err := parseOutput(args[1:])
	if err != nil {
		fmt.Fprintf(a.stderr, "alfred: %s\n", localizeDiagnostic(err.Error(), global.lang))
		return exitUsage
	}
	if args[0] == "list" && len(remaining) != 0 || args[0] == "show" && len(remaining) != 1 {
		fmt.Fprintln(a.stderr, msg.text("run_usage"))
		return exitUsage
	}
	result, ok := a.loadConfig(global, msg)
	if !ok {
		return exitFailure
	}

	if args[0] == "list" {
		executions, listErr := workflow.ListAudits(result.RootDir)
		if listErr != nil {
			fmt.Fprintf(a.stderr, "alfred: %s\n", localizeDiagnostic(listErr.Error(), global.lang))
			return exitFailure
		}
		if format == outputJSON {
			_ = writeJSON(a.stdout, executions)
			return exitSuccess
		}
		if len(executions) == 0 {
			fmt.Fprintln(a.stdout, msg.text("run_list_empty"))
			return exitSuccess
		}
		fmt.Fprintln(a.stdout, msg.text("run_list_title"))
		for _, execution := range executions {
			fmt.Fprintln(a.stdout, msg.text("run_list_item", execution.ID, execution.Workflow, localizeOutcome(execution.Outcome, global.lang)))
		}
		return exitSuccess
	}

	execution, loadErr := workflow.LoadAudit(result.RootDir, remaining[0])
	if loadErr != nil {
		fmt.Fprintf(a.stderr, "alfred: %s\n", localizeDiagnostic(loadErr.Error(), global.lang))
		return exitFailure
	}
	if format == outputJSON {
		_ = writeJSON(a.stdout, execution)
		return exitSuccess
	}
	fmt.Fprintln(a.stdout, msg.text("run_show_title", execution.ID))
	encoded, _ := writeIndentedJSON(execution)
	fmt.Fprint(a.stdout, encoded)
	return exitSuccess
}

func writeIndentedJSON(value any) (string, error) {
	var builder strings.Builder
	err := writeJSON(&builder, value)
	return builder.String(), err
}
