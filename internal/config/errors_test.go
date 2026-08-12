package config

import (
	"strings"
	"testing"
)

func TestValidationErrorLocalizationPreservesLocation(t *testing.T) {
	err := ValidationError{
		File:    "alfred.yaml",
		Path:    "workflows.deploy.steps[0].command",
		Message: "must name one executable; place parameters in args",
	}
	localized := err.Localized("pt-BR")
	if !strings.Contains(localized, err.File) || !strings.Contains(localized, err.Path) || !strings.Contains(localized, "executável") {
		t.Fatalf("unexpected localized error: %s", localized)
	}
	if strings.Contains(localized, "must name") {
		t.Fatalf("validation guidance was not translated: %s", localized)
	}
}

func TestDynamicValidationMessagesAreLocalized(t *testing.T) {
	tests := map[string]string{
		"configuration exceeds the 1024 byte limit":    "1024 bytes",
		"YAML exceeds the 10000 node limit":            "10000 nós",
		"YAML exceeds the maximum depth of 64":         "profundidade máxima de 64",
		"custom YAML tag \"!unsafe\" is not supported": "tag YAML personalizada \"!unsafe\" não é permitida",
		"extended file not found: company.yaml":        "arquivo estendido não encontrado: company.yaml",
	}
	for message, expected := range tests {
		t.Run(message, func(t *testing.T) {
			if localized := portugueseValidationMessage(message); !strings.Contains(localized, expected) {
				t.Fatalf("expected %q in %q", expected, localized)
			}
		})
	}
}
