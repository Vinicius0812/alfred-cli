package workflow

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Vinicius0812/alfred-cli/internal/config"
)

func TestBuildPlanCreatesAuditableStructuredSteps(t *testing.T) {
	root := t.TempDir()
	result := config.Result{
		RootFile: filepath.Join(root, "alfred.yaml"),
		RootDir:  root,
		Config: config.Config{
			Version: 1,
			Project: config.ProjectConfig{Name: "demo"},
			Policies: config.Policies{
				ConfirmNetworkActions: true,
				Audit:                 true,
			},
			Workflows: map[string]config.Workflow{
				"verify": {
					Description: "Verify project",
					Environment: map[string]string{"TOKEN": "${TOKEN}"},
					Steps: []config.Step{{
						Name:    "tests",
						Command: "go",
						Args:    []string{"test", "./..."},
						Risk:    config.RiskNetwork,
					}},
				},
			},
		},
	}

	plan, err := BuildPlan(result, "verify", PlanOptions{
		LookupEnv: func(key string) (string, bool) { return "top-secret", key == "TOKEN" },
		Now:       func() time.Time { return time.Date(2026, 8, 12, 12, 0, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatal(err)
	}
	if plan.Workflow != "verify" || len(plan.Steps) != 1 {
		t.Fatalf("unexpected plan: %#v", plan)
	}
	step := plan.Steps[0]
	if step.Command != "go" || strings.Join(step.Args, " ") != "test ./..." {
		t.Fatalf("unexpected structured command: %#v", step)
	}
	if !step.RequiresConfirmation || step.ConfirmationReason != "network access" {
		t.Fatalf("network confirmation not applied: %#v", step)
	}
	if len(plan.secretValues) != 1 || plan.secretValues[0] != "top-secret" {
		t.Fatalf("secret not tracked for redaction")
	}
}

func TestBuildPlanRejectsUnknownWorkflow(t *testing.T) {
	_, err := BuildPlan(config.Result{Config: config.Config{Workflows: map[string]config.Workflow{}}}, "missing", PlanOptions{})
	if !errors.Is(err, ErrWorkflowNotFound) {
		t.Fatalf("expected ErrWorkflowNotFound, got %v", err)
	}
}

func TestExecuteDryRunWritesAudit(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 8, 12, 12, 0, 0, 0, time.UTC)
	plan := Plan{
		ID:          "20260812T120000.000000000Z-deadbeef",
		Workflow:    "verify",
		ConfigFile:  filepath.Join(root, "alfred.yaml"),
		ConfigHash:  strings.Repeat("a", 64),
		ProjectRoot: root,
		Policies:    config.Policies{Audit: true},
		Steps:       []PlannedStep{{Index: 1, Name: "tests", Risk: config.RiskRead}},
	}

	execution, err := Execute(context.Background(), plan, ExecuteOptions{
		DryRun:        true,
		Now:           func() time.Time { return now },
		AlfredVersion: "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	if execution.Outcome != "planned" || execution.Steps[0].Outcome != "planned" {
		t.Fatalf("unexpected dry run: %#v", execution)
	}
	loaded, err := LoadAudit(root, plan.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.ConfigHash != plan.ConfigHash || !loaded.DryRun {
		t.Fatalf("unexpected audit: %#v", loaded)
	}
}

func TestExecuteRunsStructuredCommandAndRedactsSecrets(t *testing.T) {
	var stdout bytes.Buffer
	plan := helperPlan(t, "echo-secret")
	plan.secretValues = []string{"classified"}
	plan.Steps[0].environment["SECRET"] = "classified"

	execution, err := Execute(context.Background(), plan, ExecuteOptions{
		Stdout:        &stdout,
		Stderr:        &stdout,
		AlfredVersion: "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	if execution.Outcome != "success" {
		t.Fatalf("unexpected execution: %#v", execution)
	}
	if strings.Contains(stdout.String(), "classified") || !strings.Contains(stdout.String(), "[REDACTED]") {
		t.Fatalf("secret was not redacted: %q", stdout.String())
	}
}

func TestExecuteHonorsTimeout(t *testing.T) {
	plan := helperPlan(t, "sleep")
	plan.Steps[0].Timeout = "20ms"

	execution, err := Execute(context.Background(), plan, ExecuteOptions{AlfredVersion: "test"})
	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected timeout, got execution=%#v err=%v", execution, err)
	}
	if execution.Outcome != "failed" {
		t.Fatalf("unexpected outcome: %s", execution.Outcome)
	}
}

func TestExecuteBlocksYesWithoutPolicy(t *testing.T) {
	plan := helperPlan(t, "success")
	plan.Steps[0].RequiresConfirmation = true

	execution, err := Execute(context.Background(), plan, ExecuteOptions{AssumeYes: true})
	if err == nil || !strings.Contains(err.Error(), "allow_non_interactive") {
		t.Fatalf("expected non-interactive policy error, got %v", err)
	}
	if execution.Outcome != "blocked" {
		t.Fatalf("unexpected outcome: %s", execution.Outcome)
	}
}

func TestExecuteContinuesOnlyWhenConfigured(t *testing.T) {
	plan := helperPlan(t, "failure")
	plan.Steps[0].ContinueOnError = true
	second := plan.Steps[0]
	second.Index = 2
	second.Name = "success"
	second.Args = []string{"-test.run=TestWorkflowHelperProcess", "--", "success"}
	plan.Steps = append(plan.Steps, second)

	execution, err := Execute(context.Background(), plan, ExecuteOptions{AlfredVersion: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if execution.Outcome != "success_with_warnings" || execution.Steps[0].Outcome != "failed" || execution.Steps[1].Outcome != "success" {
		t.Fatalf("unexpected continued execution: %#v", execution)
	}
}

func TestExecuteHonorsParentCancellation(t *testing.T) {
	plan := helperPlan(t, "sleep")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	execution, err := Execute(ctx, plan, ExecuteOptions{AlfredVersion: "test"})
	if err == nil || !strings.Contains(err.Error(), "canceled") || execution.Outcome != "failed" {
		t.Fatalf("expected canceled execution, got execution=%#v err=%v", execution, err)
	}
}

func TestExecuteNeverContinuesAfterParentCancellation(t *testing.T) {
	plan := helperPlan(t, "sleep")
	plan.Steps[0].ContinueOnError = true
	second := plan.Steps[0]
	second.Index = 2
	second.Name = "must-not-run"
	second.Args = []string{"-test.run=TestWorkflowHelperProcess", "--", "success"}
	plan.Steps = append(plan.Steps, second)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	execution, err := Execute(ctx, plan, ExecuteOptions{AlfredVersion: "test"})
	if err == nil || !strings.Contains(err.Error(), "canceled") {
		t.Fatalf("expected canceled execution, got execution=%#v err=%v", execution, err)
	}
	if execution.Outcome != "failed" || len(execution.Steps) != 1 {
		t.Fatalf("cancellation was incorrectly continued: %#v", execution)
	}
}

func TestStructuredArgumentsAreNotInterpretedByShell(t *testing.T) {
	plan := helperPlan(t, "echo-arg")
	marker := filepath.Join(plan.ProjectRoot, "should-not-exist")
	argument := "; touch " + marker
	plan.Steps[0].Args = []string{"-test.run=TestWorkflowHelperProcess", "--", "echo-arg", argument}
	var stdout bytes.Buffer

	_, err := Execute(context.Background(), plan, ExecuteOptions{Stdout: &stdout, AlfredVersion: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), argument) {
		t.Fatalf("argument was not passed literally: %q", stdout.String())
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("structured argument was interpreted by a shell")
	}
}

func TestRedactingWriterMasksSecretsSplitAcrossWritesWithoutNewlines(t *testing.T) {
	var output bytes.Buffer
	writer := newRedactingWriter(&output, []string{"top-secret"})
	if _, err := writer.Write([]byte("value=top-")); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "top-") {
		t.Fatalf("partial secret was emitted too early: %q", output.String())
	}
	if _, err := writer.Write([]byte("secret;done")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Flush(); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "top-secret") || !strings.Contains(output.String(), "[REDACTED]") {
		t.Fatalf("secret was not redacted across writes: %q", output.String())
	}
}

func TestExecuteRedactsSensitiveInheritedEnvironment(t *testing.T) {
	for _, testCase := range []struct {
		name  string
		value string
	}{
		{name: "API_TOKEN", value: "inherited-secret"},
		{name: "AWS_ACCESS_KEY_ID", value: "inherited-access-key"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var stdout bytes.Buffer
			plan := helperPlan(t, "echo-inherited")
			plan.Steps[0].Args = []string{"-test.run=TestWorkflowHelperProcess", "--", "echo-inherited", testCase.name}
			execution, err := Execute(context.Background(), plan, ExecuteOptions{
				Stdout:        &stdout,
				Environ:       append(os.Environ(), testCase.name+"="+testCase.value),
				AlfredVersion: "test",
			})
			if err != nil || execution.Outcome != "success" {
				t.Fatalf("unexpected execution: %#v err=%v", execution, err)
			}
			if strings.Contains(stdout.String(), testCase.value) || !strings.Contains(stdout.String(), "[REDACTED]") {
				t.Fatalf("inherited credential was not redacted: %q", stdout.String())
			}
		})
	}
}

func TestLoadAuditRejectsTraversal(t *testing.T) {
	_, err := LoadAudit(t.TempDir(), "../../secrets")
	if err == nil || !strings.Contains(err.Error(), "invalid run id") {
		t.Fatalf("expected traversal rejection, got %v", err)
	}
}

func TestWriteAuditRejectsSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, ".alfred")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	execution := Execution{ID: "20260812T120000.000000000Z-deadbeef"}
	err := WriteAudit(root, execution)
	if err == nil || !strings.Contains(err.Error(), "project root") {
		t.Fatalf("expected audit symlink escape rejection, got %v", err)
	}
}

func helperPlan(t *testing.T, action string) Plan {
	t.Helper()
	return Plan{
		ID:          newRunID(time.Now()),
		Workflow:    "helper",
		ProjectRoot: t.TempDir(),
		Policies:    config.Policies{Audit: false},
		Steps: []PlannedStep{{
			Index:            1,
			Name:             action,
			Command:          os.Args[0],
			Args:             []string{"-test.run=TestWorkflowHelperProcess", "--", action},
			WorkingDirectory: t.TempDir(),
			Risk:             config.RiskRead,
			Timeout:          "5s",
			environment:      map[string]string{"GO_WANT_WORKFLOW_HELPER": "1"},
		}},
	}
}

func TestWorkflowHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_WORKFLOW_HELPER") != "1" {
		return
	}
	separator := -1
	for i, arg := range os.Args {
		if arg == "--" {
			separator = i
			break
		}
	}
	if separator == -1 || separator+1 >= len(os.Args) {
		os.Exit(2)
	}
	switch os.Args[separator+1] {
	case "success":
		os.Exit(0)
	case "echo-secret":
		fmt.Println(os.Getenv("SECRET"))
		os.Exit(0)
	case "sleep":
		time.Sleep(2 * time.Second)
		os.Exit(0)
	case "failure":
		os.Exit(7)
	case "echo-arg":
		if separator+2 >= len(os.Args) {
			os.Exit(2)
		}
		fmt.Println(os.Args[separator+2])
		os.Exit(0)
	case "echo-inherited":
		if separator+2 >= len(os.Args) {
			os.Exit(2)
		}
		fmt.Println(os.Getenv(os.Args[separator+2]))
		os.Exit(0)
	default:
		os.Exit(3)
	}
}
