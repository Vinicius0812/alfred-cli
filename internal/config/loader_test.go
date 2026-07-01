package config

import (
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
