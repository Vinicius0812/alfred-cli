package workflow

import (
	"errors"
	"time"

	"github.com/Vinicius0812/alfred-cli/internal/config"
)

var (
	ErrWorkflowNotFound = errors.New("workflow not found")
	ErrDeclined         = errors.New("operation declined")
)

type Plan struct {
	ID           string          `json:"id"`
	Workflow     string          `json:"workflow"`
	Description  string          `json:"description"`
	ConfigFile   string          `json:"config_file"`
	ConfigHash   string          `json:"config_hash"`
	ProjectRoot  string          `json:"project_root"`
	CreatedAt    time.Time       `json:"created_at"`
	Policies     config.Policies `json:"policies"`
	Steps        []PlannedStep   `json:"steps"`
	secretValues []string
}

type PlannedStep struct {
	Index                int      `json:"index"`
	Name                 string   `json:"name"`
	Command              string   `json:"command,omitempty"`
	Args                 []string `json:"args,omitempty"`
	ShellCommand         string   `json:"shell_command,omitempty"`
	WorkingDirectory     string   `json:"working_directory"`
	EnvironmentKeys      []string `json:"environment_keys,omitempty"`
	Risk                 string   `json:"risk"`
	RequiresConfirmation bool     `json:"requires_confirmation"`
	ConfirmationReason   string   `json:"confirmation_reason,omitempty"`
	ContinueOnError      bool     `json:"continue_on_error"`
	Timeout              string   `json:"timeout"`
	environment          map[string]string
}

type Execution struct {
	ID         string       `json:"id"`
	Workflow   string       `json:"workflow"`
	ConfigFile string       `json:"config_file"`
	ConfigHash string       `json:"config_hash"`
	Version    string       `json:"alfred_version"`
	DryRun     bool         `json:"dry_run"`
	StartedAt  time.Time    `json:"started_at"`
	FinishedAt time.Time    `json:"finished_at"`
	Outcome    string       `json:"outcome"`
	Error      string       `json:"error,omitempty"`
	Steps      []StepResult `json:"steps"`
}

type StepResult struct {
	Index         int           `json:"index"`
	Name          string        `json:"name"`
	Risk          string        `json:"risk"`
	StartedAt     time.Time     `json:"started_at"`
	FinishedAt    time.Time     `json:"finished_at"`
	Duration      time.Duration `json:"duration_ns"`
	ExitCode      int           `json:"exit_code"`
	Outcome       string        `json:"outcome"`
	Confirmed     bool          `json:"confirmed"`
	AutoConfirmed bool          `json:"auto_confirmed"`
	Error         string        `json:"error,omitempty"`
}
