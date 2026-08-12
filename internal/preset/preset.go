package preset

import (
	"fmt"
	"sort"

	"github.com/Vinicius0812/alfred-cli/internal/config"
)

const (
	Basic    = "basic"
	GoSecure = "go-secure"
)

func Names() []string {
	names := []string{Basic, GoSecure}
	sort.Strings(names)
	return names
}

func Build(name, project string) (config.Config, error) {
	base := config.Config{
		Version: 1,
		Project: config.ProjectConfig{Name: project},
		Policies: config.Policies{
			ConfirmDestructiveActions: true,
			ConfirmNetworkActions:     true,
			ConfirmGitActions:         true,
			Audit:                     true,
		},
		Workflows: map[string]config.Workflow{},
	}

	switch name {
	case "", Basic:
		return base, nil
	case GoSecure:
		base.Workflows["verify"] = config.Workflow{
			Description: "Run the secure Go development baseline",
			Steps: []config.Step{
				{Name: "Run tests", Command: "go", Args: []string{"test", "./..."}, Risk: config.RiskRead, Timeout: "10m"},
				{Name: "Run static analysis", Command: "go", Args: []string{"vet", "./..."}, Risk: config.RiskRead, Timeout: "5m"},
				{Name: "Verify modules", Command: "go", Args: []string{"mod", "verify"}, Risk: config.RiskRead, Timeout: "2m"},
				{Name: "Scan vulnerabilities", Command: "go", Args: []string{"run", "golang.org/x/vuln/cmd/govulncheck@v1.6.0", "./..."}, Risk: config.RiskNetwork, Timeout: "10m"},
			},
		}
		return base, nil
	default:
		return config.Config{}, fmt.Errorf("unknown preset %q; available presets: %v", name, Names())
	}
}
