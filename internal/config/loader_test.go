package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadAndValidateAcceptsExampleWithExtends(t *testing.T) {
	root := repoRoot(t)
	result, err := LoadAndValidate(Options{
		ConfigPath: filepath.Join(root, "examples", "company", "alfred.yaml"),
		Getenv: func(string) string {
			return ""
		},
	})
	if err != nil {
		t.Fatalf("LoadAndValidate returned error: %v", err)
	}
	if len(result.Errors) != 0 {
		t.Fatalf("expected no validation errors, got:\n%s", result.Errors.Error())
	}
	verify := result.Config.Workflows["verify"]
	if len(verify.Steps) != 3 || verify.Steps[0].Command != "npm" {
		t.Fatalf("expected typed workflow, got %#v", verify)
	}
}

func TestLoadAndValidateRejectsUnknownField(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "alfred.yaml", `version: 1
project:
  name: demo
surprise: true
`)

	result, err := LoadAndValidate(Options{WorkingDir: dir})
	if err != nil {
		t.Fatalf("LoadAndValidate returned error: %v", err)
	}
	assertErrorContains(t, result.Errors, "surprise: unknown field")
}

func TestLoadAndValidateRejectsCyclicExtends(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "alfred.yaml", `version: 1
extends:
  - ./base.yaml
project:
  name: demo
`)
	writeFile(t, dir, "base.yaml", `version: 1
extends:
  - ./alfred.yaml
`)

	result, err := LoadAndValidate(Options{WorkingDir: dir})
	if err != nil {
		t.Fatalf("LoadAndValidate returned error: %v", err)
	}
	assertErrorContains(t, result.Errors, "cyclic extends reference detected")
}

func TestLoadAndValidateRejectsExcessiveExtendsGraph(t *testing.T) {
	dir := t.TempDir()
	for index := 0; index < maxConfigFiles+1; index++ {
		var content strings.Builder
		content.WriteString("version: 1\n")
		if index < maxConfigFiles {
			fmt.Fprintf(&content, "extends:\n  - ./layer-%d.yaml\n", index+1)
		}
		if index == 0 {
			content.WriteString("project:\n  name: demo\n")
		}
		writeFile(t, dir, fmt.Sprintf("layer-%d.yaml", index), content.String())
	}

	result, err := LoadAndValidate(Options{ConfigPath: filepath.Join(dir, "layer-0.yaml")})
	if err != nil {
		t.Fatalf("LoadAndValidate returned error: %v", err)
	}
	assertErrorContains(t, result.Errors, "configuration graph exceeds the 64 file limit")
}

func TestLoadAndValidateRejectsUnresolvedEnvironmentReference(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "alfred.yaml", `version: 1
project:
  name: demo
workflows:
  deploy:
    description: Deploy
    environment:
      TOKEN: ${DEPLOY_TOKEN}
    steps:
      - run: echo deploy
`)

	result, err := LoadAndValidate(Options{
		WorkingDir: dir,
		Getenv: func(string) string {
			return ""
		},
	})
	if err != nil {
		t.Fatalf("LoadAndValidate returned error: %v", err)
	}
	assertErrorContains(t, result.Errors, "references unresolved environment variable DEPLOY_TOKEN")
}

func TestLoadAndValidateRejectsPathsEscapingProjectRoot(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "alfred.yaml", `version: 1
project:
  name: demo
docker:
  compose_files:
    - ../compose.yaml
workflows:
  check:
    description: Check
    working_directory: ..
    steps:
      - run: go test ./...
`)

	result, err := LoadAndValidate(Options{WorkingDir: dir})
	if err != nil {
		t.Fatalf("LoadAndValidate returned error: %v", err)
	}
	assertErrorContains(t, result.Errors, "docker.compose_files[0]: must not escape the project root")
	assertErrorContains(t, result.Errors, "workflows.check.working_directory: must not escape the project root")
}

func TestLoadAndValidateRejectsDuplicateYAMLKeys(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "alfred.yaml", `version: 1
project:
  name: first
  name: second
`)

	result, err := LoadAndValidate(Options{WorkingDir: dir})
	if err != nil {
		t.Fatalf("LoadAndValidate returned error: %v", err)
	}
	assertErrorContains(t, result.Errors, "duplicate YAML key")
}

func TestLoadAndValidateRejectsYAMLAliases(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "alfred.yaml", `version: 1
project: &project
  name: demo
copy: *project
`)

	result, err := LoadAndValidate(Options{WorkingDir: dir})
	if err != nil {
		t.Fatalf("LoadAndValidate returned error: %v", err)
	}
	assertErrorContains(t, result.Errors, "YAML aliases are not supported")
}

func TestLoadAndValidateRejectsExcessiveYAMLDepth(t *testing.T) {
	tempDir := t.TempDir()
	var content strings.Builder
	for depth := 0; depth < maxYAMLDepth+2; depth++ {
		content.WriteString(strings.Repeat("  ", depth))
		fmt.Fprintf(&content, "level%d:\n", depth)
	}
	content.WriteString(strings.Repeat("  ", maxYAMLDepth+2))
	content.WriteString("value: true\n")
	writeFile(t, tempDir, "alfred.yaml", content.String())
	path := filepath.Join(tempDir, "alfred.yaml")

	result, err := LoadAndValidate(Options{ConfigPath: path, WorkingDir: tempDir})
	if err != nil {
		t.Fatal(err)
	}
	assertErrorContains(t, result.Errors, "maximum depth")
}

func TestLoadAndValidateRejectsExcessiveYAMLNodes(t *testing.T) {
	tempDir := t.TempDir()
	var content strings.Builder
	for index := 0; index < maxYAMLNodes/2+2; index++ {
		fmt.Fprintf(&content, "key%d: value\n", index)
	}
	writeFile(t, tempDir, "alfred.yaml", content.String())
	path := filepath.Join(tempDir, "alfred.yaml")

	result, err := LoadAndValidate(Options{ConfigPath: path, WorkingDir: tempDir})
	if err != nil {
		t.Fatal(err)
	}
	assertErrorContains(t, result.Errors, "node limit")
}

func TestLoadAndValidateAppliesTypedDefaults(t *testing.T) {
	tempDir := t.TempDir()
	writeFile(t, tempDir, "alfred.yaml", "version: 1\nproject:\n  name: defaults\n")
	path := filepath.Join(tempDir, "alfred.yaml")
	result, err := LoadAndValidate(Options{ConfigPath: path, WorkingDir: tempDir})
	if err != nil || len(result.Errors) != 0 {
		t.Fatalf("unexpected validation result: err=%v validation=%v", err, result.Errors)
	}
	if !result.Config.Commit.Enabled || !result.Config.Docker.Enabled || !result.Config.Policies.Audit {
		t.Fatalf("typed defaults were not applied: %#v", result.Config)
	}
}

func TestRequiredToolsIncludesEveryStructuredExecutable(t *testing.T) {
	root := repoRoot(t)
	result, err := LoadAndValidate(Options{ConfigPath: filepath.Join(root, "examples", "go", "alfred.yaml")})
	if err != nil || len(result.Errors) != 0 {
		t.Fatalf("unexpected validation result: err=%v validation=%v", err, result.Errors)
	}
	joined := strings.Join(result.RequiredTools, ",")
	if !strings.Contains(joined, "go") || !strings.Contains(joined, "gofmt") {
		t.Fatalf("missing structured tools: %v", result.RequiredTools)
	}
}

func TestLoadAndValidateEnforcesFileSizeLimit(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "alfred.yaml", "version: 1\nproject:\n  name: a-name-that-is-too-long\n")

	result, err := LoadAndValidate(Options{WorkingDir: dir, MaxConfigBytes: 16})
	if err != nil {
		t.Fatalf("LoadAndValidate returned error: %v", err)
	}
	assertErrorContains(t, result.Errors, "configuration exceeds the 16 byte limit")
}

func TestLoadAndValidateAcceptsDefinedEmptyEnvironmentVariable(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "alfred.yaml", `version: 1
project:
  name: demo
workflows:
  check:
    description: Check
    environment:
      OPTIONAL_VALUE: ${OPTIONAL_VALUE}
    steps:
      - command: go
        args: [version]
        risk: read
`)

	result, err := LoadAndValidate(Options{
		WorkingDir: dir,
		LookupEnv: func(key string) (string, bool) {
			return "", key == "OPTIONAL_VALUE"
		},
	})
	if err != nil {
		t.Fatalf("LoadAndValidate returned error: %v", err)
	}
	if len(result.Errors) != 0 {
		t.Fatalf("expected no validation errors, got:\n%s", result.Errors.Error())
	}
}

func TestLoadAndValidateRejectsInvalidEnvironmentName(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "alfred.yaml", `version: 1
project:
  name: demo
workflows:
  check:
    description: Check
    environment:
      BAD-NAME: value
    steps:
      - command: go
        args: [version]
        risk: read
`)

	result, err := LoadAndValidate(Options{WorkingDir: dir})
	if err != nil {
		t.Fatalf("LoadAndValidate returned error: %v", err)
	}
	assertErrorContains(t, result.Errors, "environment variable name must match")
}

func TestLoadAndValidateRequiresExplicitShellOptIn(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "alfred.yaml", `version: 1
project:
  name: demo
workflows:
  check:
    description: Check
    steps:
      - run: go test ./...
`)

	result, err := LoadAndValidate(Options{WorkingDir: dir})
	if err != nil {
		t.Fatalf("LoadAndValidate returned error: %v", err)
	}
	assertErrorContains(t, result.Errors, "shell: must be true for a run string")
	assertErrorContains(t, result.Errors, "confirm: must be true for a shell step")
	assertErrorContains(t, result.Errors, "policies.allow_shell_steps: must be true")
}

func TestLoadAndValidateRejectsExternalExtendsByDefault(t *testing.T) {
	parent := t.TempDir()
	projectDir := filepath.Join(parent, "project")
	if err := os.Mkdir(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, parent, "base.yaml", "version: 1\n")
	writeFile(t, projectDir, "alfred.yaml", `version: 1
extends:
  - ../base.yaml
project:
  name: demo
`)

	result, err := LoadAndValidate(Options{WorkingDir: projectDir})
	if err != nil {
		t.Fatalf("LoadAndValidate returned error: %v", err)
	}
	assertErrorContains(t, result.Errors, "must remain inside the project root")

	trusted, err := LoadAndValidate(Options{WorkingDir: projectDir, AllowExternalExtends: true})
	if err != nil {
		t.Fatalf("trusted LoadAndValidate returned error: %v", err)
	}
	if len(trusted.Errors) != 0 {
		t.Fatalf("expected trusted external preset to pass, got:\n%s", trusted.Errors.Error())
	}
}

func TestLoadAndValidateRejectsSymlinkEscape(t *testing.T) {
	projectDir := t.TempDir()
	externalDir := t.TempDir()
	link := filepath.Join(projectDir, "external")
	if err := os.Symlink(externalDir, link); err != nil {
		t.Skipf("symlink creation is unavailable: %v", err)
	}
	writeFile(t, projectDir, "alfred.yaml", `version: 1
project:
  name: demo
workflows:
  check:
    description: Check
    working_directory: ./external
    steps:
      - command: go
        args: [version]
        risk: read
`)

	result, err := LoadAndValidate(Options{WorkingDir: projectDir})
	if err != nil {
		t.Fatalf("LoadAndValidate returned error: %v", err)
	}
	assertErrorContains(t, result.Errors, "must not escape the project root")
}

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			t.Fatal("go.mod not found")
		}
		wd = parent
	}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func assertErrorContains(t *testing.T, errs ErrorList, want string) {
	t.Helper()
	for _, err := range errs {
		if strings.Contains(err.Error(), want) {
			return
		}
	}
	t.Fatalf("expected validation error containing %q, got:\n%s", want, errs.Error())
}
