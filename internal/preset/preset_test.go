package preset

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestGoSecurePresetIsStructuredAndSerializable(t *testing.T) {
	configured, err := Build(GoSecure, "demo: project\nwith yaml")
	if err != nil {
		t.Fatal(err)
	}
	workflow := configured.Workflows["verify"]
	if len(workflow.Steps) != 4 {
		t.Fatalf("expected four secure checks, got %d", len(workflow.Steps))
	}
	for _, step := range workflow.Steps {
		if step.Command == "" || step.Run != "" || step.Shell {
			t.Fatalf("preset contains a non-structured step: %#v", step)
		}
	}
	content, err := yaml.Marshal(configured)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := yaml.Unmarshal(content, &decoded); err != nil {
		t.Fatalf("generated YAML is invalid: %v\n%s", err, content)
	}
}
