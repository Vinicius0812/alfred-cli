package workflow

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Vinicius0812/alfred-cli/internal/config"
)

type PlanOptions struct {
	LookupEnv      func(string) (string, bool)
	DefaultTimeout time.Duration
	Now            func() time.Time
}

func BuildPlan(result config.Result, workflowName string, opts PlanOptions) (Plan, error) {
	if len(result.Errors) > 0 {
		return Plan{}, fmt.Errorf("configuration is invalid: %s", result.Errors.Error())
	}
	configured, ok := result.Config.Workflows[workflowName]
	if !ok {
		return Plan{}, fmt.Errorf("%w: %s", ErrWorkflowNotFound, workflowName)
	}
	if opts.LookupEnv == nil {
		opts.LookupEnv = func(string) (string, bool) { return "", false }
	}
	if opts.DefaultTimeout <= 0 {
		opts.DefaultTimeout = 10 * time.Minute
	}
	if opts.Now == nil {
		opts.Now = time.Now
	}

	configJSON, err := json.Marshal(result.Config)
	if err != nil {
		return Plan{}, fmt.Errorf("could not hash configuration: %w", err)
	}
	hash := sha256.Sum256(configJSON)

	createdAt := opts.Now().UTC()
	plan := Plan{
		ID:          newRunID(createdAt),
		Workflow:    workflowName,
		Description: configured.Description,
		ConfigFile:  result.RootFile,
		ConfigHash:  hex.EncodeToString(hash[:]),
		ProjectRoot: result.RootDir,
		CreatedAt:   createdAt,
		Policies:    result.Config.Policies,
	}

	workflowDir := result.RootDir
	if configured.WorkingDirectory != "" {
		workflowDir, err = config.ResolveProjectPath(result.RootDir, configured.WorkingDirectory)
		if err != nil {
			return Plan{}, fmt.Errorf("invalid workflow directory: %w", err)
		}
	}

	workflowEnv, err := resolveEnvironment(configured.Environment, opts.LookupEnv)
	if err != nil {
		return Plan{}, err
	}
	secretSet := map[string]bool{}
	for _, value := range workflowEnv {
		if value != "" {
			secretSet[value] = true
		}
	}

	for index, step := range configured.Steps {
		stepEnv := cloneEnvironment(workflowEnv)
		resolvedOverrides, resolveErr := resolveEnvironment(step.Environment, opts.LookupEnv)
		if resolveErr != nil {
			return Plan{}, fmt.Errorf("step %d environment: %w", index+1, resolveErr)
		}
		for key, value := range resolvedOverrides {
			stepEnv[key] = value
			if value != "" {
				secretSet[value] = true
			}
		}

		workingDir := workflowDir
		if step.WorkingDirectory != "" {
			workingDir, err = config.ResolveProjectPath(result.RootDir, step.WorkingDirectory)
			if err != nil {
				return Plan{}, fmt.Errorf("step %d directory: %w", index+1, err)
			}
		}

		timeout, err := step.ParsedTimeout(opts.DefaultTimeout)
		if err != nil {
			return Plan{}, fmt.Errorf("step %d timeout: %w", index+1, err)
		}
		planned := PlannedStep{
			Index:            index + 1,
			Name:             step.Name,
			WorkingDirectory: workingDir,
			EnvironmentKeys:  sortedKeys(stepEnv),
			Risk:             step.Risk,
			ContinueOnError:  step.ContinueOnError,
			Timeout:          timeout.String(),
			environment:      stepEnv,
		}

		if step.Command != "" {
			planned.Command = step.Command
			planned.Args = append([]string(nil), step.Args...)
			if strings.ContainsAny(step.Command, `/\`) {
				planned.Command, err = config.ResolveProjectPath(result.RootDir, step.Command)
				if err != nil {
					return Plan{}, fmt.Errorf("step %d command: %w", index+1, err)
				}
			}
		} else {
			planned.ShellCommand = step.Run
		}
		if planned.Name == "" {
			if planned.Command != "" {
				planned.Name = filepath.Base(planned.Command)
			} else {
				planned.Name = "shell"
			}
		}

		planned.RequiresConfirmation, planned.ConfirmationReason = confirmationFor(step, result.Config.Policies)
		plan.Steps = append(plan.Steps, planned)
	}

	for value := range secretSet {
		plan.secretValues = append(plan.secretValues, value)
	}
	sort.Slice(plan.secretValues, func(i, j int) bool { return len(plan.secretValues[i]) > len(plan.secretValues[j]) })
	return plan, nil
}

func confirmationFor(step config.Step, policies config.Policies) (bool, string) {
	if step.Shell {
		return true, "shell command"
	}
	if step.Confirm {
		return true, "step policy"
	}
	switch step.Risk {
	case config.RiskNetwork:
		return policies.ConfirmNetworkActions, "network access"
	case config.RiskGitWrite:
		return policies.ConfirmGitActions, "Git state change"
	case config.RiskDestructive:
		return policies.ConfirmDestructiveActions, "destructive operation"
	default:
		return false, ""
	}
}

func resolveEnvironment(values map[string]string, lookupEnv func(string) (string, bool)) (map[string]string, error) {
	resolved := make(map[string]string, len(values))
	for key, value := range values {
		var resolutionErr error
		resolved[key] = environmentReferencePattern.ReplaceAllStringFunc(value, func(match string) string {
			name := environmentReferencePattern.FindStringSubmatch(match)[1]
			replacement, ok := lookupEnv(name)
			if !ok {
				resolutionErr = fmt.Errorf("environment variable %s is not defined", name)
				return ""
			}
			return replacement
		})
		if resolutionErr != nil {
			return nil, resolutionErr
		}
	}
	return resolved, nil
}

var environmentReferencePattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

func cloneEnvironment(values map[string]string) map[string]string {
	clone := make(map[string]string, len(values))
	for key, value := range values {
		clone[key] = value
	}
	return clone
}

func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func newRunID(now time.Time) string {
	random := make([]byte, 4)
	if _, err := rand.Read(random); err != nil {
		return now.UTC().Format("20060102T150405.000000000Z")
	}
	return now.UTC().Format("20060102T150405.000000000Z") + "-" + hex.EncodeToString(random)
}
