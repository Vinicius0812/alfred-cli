package config

import "time"

const (
	RiskRead        = "read"
	RiskLocalWrite  = "local-write"
	RiskNetwork     = "network"
	RiskGitWrite    = "git-write"
	RiskDestructive = "destructive"
)

type Config struct {
	Version   int                 `yaml:"version" json:"version"`
	Extends   []string            `yaml:"extends,omitempty" json:"extends,omitempty"`
	Project   ProjectConfig       `yaml:"project" json:"project"`
	Commit    CommitConfig        `yaml:"commit,omitempty" json:"commit,omitempty"`
	Docker    DockerConfig        `yaml:"docker,omitempty" json:"docker,omitempty"`
	Workflows map[string]Workflow `yaml:"workflows,omitempty" json:"workflows,omitempty"`
	Policies  Policies            `yaml:"policies,omitempty" json:"policies,omitempty"`
}

type ProjectConfig struct {
	Name string `yaml:"name" json:"name"`
}

type CommitConfig struct {
	Enabled           bool     `yaml:"enabled,omitempty" json:"enabled"`
	Convention        string   `yaml:"convention,omitempty" json:"convention,omitempty"`
	AllowedTypes      []string `yaml:"allowed_types,omitempty" json:"allowed_types,omitempty"`
	Scopes            []string `yaml:"scopes,omitempty" json:"scopes,omitempty"`
	ProtectedBranches []string `yaml:"protected_branches,omitempty" json:"protected_branches,omitempty"`
	RequireScope      bool     `yaml:"require_scope,omitempty" json:"require_scope"`
	Emoji             bool     `yaml:"emoji,omitempty" json:"emoji"`
	ConfirmCommit     bool     `yaml:"confirm_commit,omitempty" json:"confirm_commit"`
	ConfirmPush       bool     `yaml:"confirm_push,omitempty" json:"confirm_push"`
}

type DockerConfig struct {
	Enabled      bool     `yaml:"enabled,omitempty" json:"enabled"`
	ComposeFiles []string `yaml:"compose_files,omitempty" json:"compose_files,omitempty"`
	Profiles     []string `yaml:"profiles,omitempty" json:"profiles,omitempty"`
	EnvFiles     []string `yaml:"env_files,omitempty" json:"env_files,omitempty"`
	ProjectName  string   `yaml:"project_name,omitempty" json:"project_name,omitempty"`
}

type Workflow struct {
	Description      string            `yaml:"description" json:"description"`
	WorkingDirectory string            `yaml:"working_directory,omitempty" json:"working_directory,omitempty"`
	Environment      map[string]string `yaml:"environment,omitempty" json:"environment,omitempty"`
	Steps            []Step            `yaml:"steps" json:"steps"`
}

type Step struct {
	Name             string            `yaml:"name,omitempty" json:"name,omitempty"`
	Command          string            `yaml:"command,omitempty" json:"command,omitempty"`
	Args             []string          `yaml:"args,omitempty" json:"args,omitempty"`
	Run              string            `yaml:"run,omitempty" json:"run,omitempty"`
	Shell            bool              `yaml:"shell,omitempty" json:"shell"`
	WorkingDirectory string            `yaml:"working_directory,omitempty" json:"working_directory,omitempty"`
	Environment      map[string]string `yaml:"environment,omitempty" json:"environment,omitempty"`
	Confirm          bool              `yaml:"confirm,omitempty" json:"confirm"`
	ContinueOnError  bool              `yaml:"continue_on_error,omitempty" json:"continue_on_error"`
	Timeout          string            `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Risk             string            `yaml:"risk,omitempty" json:"risk"`
}

func (s Step) ParsedTimeout(fallback time.Duration) (time.Duration, error) {
	if s.Timeout == "" {
		return fallback, nil
	}
	return time.ParseDuration(s.Timeout)
}

type Policies struct {
	RequireCleanWorktree      bool `yaml:"require_clean_worktree,omitempty" json:"require_clean_worktree"`
	ConfirmDestructiveActions bool `yaml:"confirm_destructive_actions,omitempty" json:"confirm_destructive_actions"`
	ConfirmNetworkActions     bool `yaml:"confirm_network_actions,omitempty" json:"confirm_network_actions"`
	ConfirmGitActions         bool `yaml:"confirm_git_actions,omitempty" json:"confirm_git_actions"`
	AllowShellSteps           bool `yaml:"allow_shell_steps,omitempty" json:"allow_shell_steps"`
	AllowNonInteractive       bool `yaml:"allow_non_interactive,omitempty" json:"allow_non_interactive"`
	Audit                     bool `yaml:"audit,omitempty" json:"audit"`
}

func defaultConfig() Config {
	return Config{
		Commit: CommitConfig{
			Enabled:           true,
			Convention:        "conventional",
			AllowedTypes:      []string{"feat", "fix", "docs", "refactor", "test", "chore"},
			ProtectedBranches: []string{"main", "master"},
			ConfirmCommit:     true,
			ConfirmPush:       true,
		},
		Docker: DockerConfig{Enabled: true},
		Policies: Policies{
			ConfirmDestructiveActions: true,
			ConfirmNetworkActions:     true,
			ConfirmGitActions:         true,
			Audit:                     true,
		},
		Workflows: map[string]Workflow{},
	}
}
