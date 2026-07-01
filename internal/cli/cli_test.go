package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunVersion(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"--version"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), Version) {
		t.Fatalf("expected version output, got %q", stdout.String())
	}
}

func TestRunConfigValidate(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	configPath := filepath.Join("..", "..", "examples", "go", "alfred.yaml")

	code := Run([]string{"--config", configPath, "config", "validate"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "configuration valid: example-go-api") {
		t.Fatalf("unexpected stdout: %s", stdout.String())
	}
}

func TestRunConfigValidatePortuguese(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	configPath := filepath.Join("..", "..", "examples", "go", "alfred.yaml")

	code := Run([]string{"--lang", "pt-BR", "--config", configPath, "config", "validate"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "configuração válida: example-go-api") {
		t.Fatalf("unexpected stdout: %s", stdout.String())
	}
}

func TestRunConfigValidateMissingFile(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"--config", "missing.yaml", "config", "validate"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "configuration file not found") {
		t.Fatalf("unexpected stderr: %s", stderr.String())
	}
}

func TestRunInitCreatesConfig(t *testing.T) {
	tempDir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Fatal(err)
		}
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"init", "--project", "demo-api"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr=%s", code, stderr.String())
	}

	content, err := os.ReadFile(filepath.Join(tempDir, "alfred.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "name: demo-api") {
		t.Fatalf("unexpected config:\n%s", string(content))
	}
}

func TestRunDoctorValidatesConfigAndTools(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "alfred.yaml")
	if err := os.WriteFile(configPath, []byte(`version: 1

project:
  name: doctor-demo
`), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"--config", configPath, "doctor"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "Alfred Doctor") {
		t.Fatalf("unexpected stdout: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "OK config") {
		t.Fatalf("unexpected stdout: %s", stdout.String())
	}
}
