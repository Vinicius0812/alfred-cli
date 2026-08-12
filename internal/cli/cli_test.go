package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunVersion(t *testing.T) {
	stdout, stderr, code := runCLI(t, nil, "--version")
	if code != exitSuccess {
		t.Fatalf("expected exit code 0, got %d; stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, Version) {
		t.Fatalf("expected version output, got %q", stdout)
	}
}

func TestRunRejectsUnsupportedLanguage(t *testing.T) {
	_, stderr, code := runCLI(t, nil, "--lang", "fr", "help")
	if code != exitUsage {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr, "unsupported language") {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
}

func TestRunConfigValidate(t *testing.T) {
	configPath := filepath.Join("..", "..", "examples", "go", "alfred.yaml")
	stdout, stderr, code := runCLI(t, nil, "--config", configPath, "config", "validate")
	if code != exitSuccess {
		t.Fatalf("expected exit code 0, got %d; stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Configuration valid: example-go-api") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestRunConfigValidatePortuguese(t *testing.T) {
	configPath := filepath.Join("..", "..", "examples", "go", "alfred.yaml")
	stdout, stderr, code := runCLI(t, nil, "--lang", "pt-BR", "--config", configPath, "config", "validate")
	if code != exitSuccess {
		t.Fatalf("expected exit code 0, got %d; stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "mestre Bruce") || !strings.Contains(stdout, "Configuração válida: example-go-api") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestGlobalArgumentErrorUsesConfiguredLanguage(t *testing.T) {
	t.Setenv("ALFRED_LANG", "pt-BR")
	_, stderr, code := runCLI(t, nil, "--config")
	if code != exitUsage {
		t.Fatalf("expected usage failure, got %d", code)
	}
	if !strings.Contains(stderr, "exige um caminho") {
		t.Fatalf("unexpected localized error: %s", stderr)
	}
}

func TestInitRejectsEmptyProjectInPortuguese(t *testing.T) {
	_, stderr, code := runCLI(t, nil, "--lang", "pt-BR", "init", "--project", "")
	if code != exitUsage {
		t.Fatalf("expected usage failure, got %d", code)
	}
	if !strings.Contains(stderr, "não pode ficar vazio") {
		t.Fatalf("unexpected localized error: %s", stderr)
	}
}

func TestRunConfigValidateMissingFileIncludesHint(t *testing.T) {
	_, stderr, code := runCLI(t, nil, "--config", "missing.yaml", "config", "validate")
	if code != exitFailure {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr, "configuration file not found") || !strings.Contains(stderr, "alfred init") {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
}

func TestInvalidSubcommandsFailBeforeConfigurationDiscovery(t *testing.T) {
	_, configError, configCode := runCLI(t, nil, "config", "unknown")
	if configCode != exitUsage || !strings.Contains(configError, "unknown config command") {
		t.Fatalf("unexpected config result: code=%d stderr=%s", configCode, configError)
	}
	if strings.Contains(configError, "alfred.yaml not found") {
		t.Fatalf("configuration discovery masked the usage error: %s", configError)
	}

	_, runError, runCode := runCLI(t, nil, "run", "list", "extra")
	if runCode != exitUsage || !strings.Contains(runError, "Usage: alfred run") {
		t.Fatalf("unexpected run result: code=%d stderr=%s", runCode, runError)
	}
	if strings.Contains(runError, "alfred.yaml not found") {
		t.Fatalf("configuration discovery masked the usage error: %s", runError)
	}
}

func TestDoctorHelpDoesNotRequireConfiguration(t *testing.T) {
	stdout, stderr, code := runCLI(t, nil, "doctor", "--help")
	if code != exitSuccess || stderr != "" || !strings.Contains(stdout, "Usage: alfred doctor") {
		t.Fatalf("unexpected help result: code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
}

func TestRunConfigMissingFileIsLocalized(t *testing.T) {
	_, stderr, code := runCLI(t, nil, "--lang", "pt-BR", "--config", "ausente.yaml", "config", "validate")
	if code != exitFailure {
		t.Fatalf("expected failure, got %d", code)
	}
	if !strings.Contains(stderr, "arquivo de configuração não encontrado") || !strings.Contains(stderr, "Dica:") {
		t.Fatalf("unexpected localized error: %s", stderr)
	}
}

func TestRunConfigExplainJSON(t *testing.T) {
	configPath := filepath.Join("..", "..", "examples", "go", "alfred.yaml")
	stdout, stderr, code := runCLI(t, nil, "--config", configPath, "config", "explain", "--output", "json")
	if code != exitSuccess {
		t.Fatalf("expected success, got %d: %s", code, stderr)
	}
	var result struct {
		Config struct {
			Project struct {
				Name string `json:"name"`
			} `json:"project"`
		} `json:"config"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, stdout)
	}
	if result.Config.Project.Name != "example-go-api" {
		t.Fatalf("unexpected project: %+v", result)
	}
}

func TestRunInitCreatesSafePreset(t *testing.T) {
	tempDir := t.TempDir()
	withWorkingDirectory(t, tempDir)
	project := "demo: api # safe"
	stdout, stderr, code := runCLI(t, nil, "init", "--project", project, "--preset", "go-secure")
	if code != exitSuccess {
		t.Fatalf("expected exit code 0, got %d; stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "go-secure") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
	content, err := os.ReadFile(filepath.Join(tempDir, "alfred.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "govulncheck") {
		t.Fatalf("secure preset missing vulnerability scan:\n%s", content)
	}
	validateOut, validateErr, validateCode := runCLI(t, nil, "config", "validate")
	if validateCode != exitSuccess {
		t.Fatalf("generated YAML is invalid: stdout=%s stderr=%s", validateOut, validateErr)
	}
	if !strings.Contains(validateOut, project) {
		t.Fatalf("project name was not preserved: %s", validateOut)
	}
}

func TestRunInitUsesDirectoryNameAndExplicitConfigPath(t *testing.T) {
	parent := t.TempDir()
	projectDir := filepath.Join(parent, "wayne-api")
	configDir := filepath.Join(projectDir, "config")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatal(err)
	}
	withWorkingDirectory(t, projectDir)

	configPath := filepath.Join("config", "alfred.custom.yaml")
	stdout, stderr, code := runCLI(t, nil, "--config", configPath, "init")
	if code != exitSuccess {
		t.Fatalf("expected success, got code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
	content, err := os.ReadFile(filepath.Join(projectDir, configPath))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "name: wayne-api") {
		t.Fatalf("directory project name was not used:\n%s", content)
	}
	if _, err := os.Stat(filepath.Join(projectDir, "alfred.yaml")); !os.IsNotExist(err) {
		t.Fatalf("default configuration path was unexpectedly written")
	}
}

func TestRunDoctorValidatesConfigAndTools(t *testing.T) {
	tempDir := t.TempDir()
	configPath := writeConfig(t, tempDir, `version: 1
project:
  name: doctor-demo
`)
	stdout, stderr, code := runCLI(t, nil, "--config", configPath, "doctor")
	if code != exitSuccess {
		t.Fatalf("expected exit code 0, got %d; stdout=%s stderr=%s", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "Alfred Doctor") || !strings.Contains(stdout, "OK configuration") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestRunWorkflowListJSONIsSorted(t *testing.T) {
	tempDir := t.TempDir()
	configPath := writeConfig(t, tempDir, workflowConfiguration)
	stdout, stderr, code := runCLI(t, nil, "--config", configPath, "workflow", "list", "--output", "json")
	if code != exitSuccess {
		t.Fatalf("expected success, got %d: %s", code, stderr)
	}
	if strings.Index(stdout, `"name": "alpha"`) > strings.Index(stdout, `"name": "verify"`) {
		t.Fatalf("workflows are not sorted: %s", stdout)
	}
}

func TestRunDryRunWritesAuditablePlan(t *testing.T) {
	tempDir := t.TempDir()
	configPath := writeConfig(t, tempDir, workflowConfiguration)
	stdout, stderr, code := runCLI(t, nil, "--config", configPath, "run", "verify", "--dry-run", "--output", "json")
	if code != exitSuccess {
		t.Fatalf("expected success, got %d: %s", code, stderr)
	}
	var response struct {
		Plan struct {
			ID string `json:"id"`
		} `json:"plan"`
		Execution struct {
			Outcome string `json:"outcome"`
		} `json:"execution"`
	}
	if err := json.Unmarshal([]byte(stdout), &response); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, stdout)
	}
	if response.Execution.Outcome != "planned" || response.Plan.ID == "" {
		t.Fatalf("unexpected response: %+v", response)
	}
	if _, err := os.Stat(filepath.Join(tempDir, ".alfred", "runs", response.Plan.ID+".json")); err != nil {
		t.Fatalf("audit record not written: %v", err)
	}

	showOut, showErr, showCode := runCLI(t, nil, "--config", configPath, "run", "show", response.Plan.ID, "--output", "json")
	if showCode != exitSuccess || !strings.Contains(showOut, response.Plan.ID) {
		t.Fatalf("could not read audit: code=%d stdout=%s stderr=%s", showCode, showOut, showErr)
	}
}

func TestRunAcceptsOptionsBeforeWorkflowName(t *testing.T) {
	tempDir := t.TempDir()
	configPath := writeConfig(t, tempDir, workflowConfiguration)
	stdout, stderr, code := runCLI(t, nil, "--config", configPath, "run", "--dry-run", "--output", "json", "verify")
	if code != exitSuccess || !strings.Contains(stdout, `"outcome": "planned"`) {
		t.Fatalf("unexpected result: code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
}

func TestRunJSONKeepsInteractivePromptOutOfStdout(t *testing.T) {
	tempDir := t.TempDir()
	configPath := writeConfig(t, tempDir, `version: 1
project:
  name: json-confirmation
policies:
  audit: false
  confirm_network_actions: true
workflows:
  network:
    description: Confirm a network-classified step
    steps:
      - command: go
        args: [version]
        risk: network
`)
	stdout, stderr, code := runCLI(t, strings.NewReader("yes\n"), "--config", configPath, "run", "network", "--output", "json")
	if code != exitSuccess {
		t.Fatalf("expected success, got %d: stdout=%s stderr=%s", code, stdout, stderr)
	}
	var response map[string]any
	if err := json.Unmarshal([]byte(stdout), &response); err != nil {
		t.Fatalf("stdout was corrupted by interactive output: %v\n%s", err, stdout)
	}
	if !strings.Contains(stderr, "requires confirmation") {
		t.Fatalf("confirmation prompt was not sent to stderr: %s", stderr)
	}
}

func TestMenuUsesPortugueseIdentityAndIsTestable(t *testing.T) {
	stdout, stderr, code := runCLI(t, strings.NewReader("0\n"), "--lang", "pt-BR", "menu")
	if code != exitSuccess || stderr != "" {
		t.Fatalf("unexpected result: code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "ALFRED") || !strings.Contains(stdout, "Como posso ajudá-lo, mestre Bruce?") {
		t.Fatalf("identity missing from menu: %s", stdout)
	}
}

func TestCompletionSupportsThreeShells(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "powershell"} {
		stdout, stderr, code := runCLI(t, nil, "completion", shell)
		if code != exitSuccess || stdout == "" || stderr != "" {
			t.Fatalf("completion %s failed: code=%d stdout=%q stderr=%s", shell, code, stdout, stderr)
		}
	}
}

const workflowConfiguration = `version: 1
project:
  name: workflow-demo
policies:
  audit: true
workflows:
  verify:
    description: Verify the project
    steps:
      - name: Print version
        command: go
        args: [version]
        risk: read
  alpha:
    description: First workflow
    steps:
      - command: go
        args: [version]
        risk: read
`

func runCLI(t *testing.T, stdin *strings.Reader, args ...string) (string, string, int) {
	t.Helper()
	if stdin == nil {
		stdin = strings.NewReader("")
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := RunWithIO(args, stdin, &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}

func writeConfig(t *testing.T, directory, content string) string {
	t.Helper()
	path := filepath.Join(directory, "alfred.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func withWorkingDirectory(t *testing.T, directory string) {
	t.Helper()
	oldDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(directory); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
}
