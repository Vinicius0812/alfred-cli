package config

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var environmentReferencePattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)
var environmentNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func validateKnownShape(file string, node *yaml.Node) ErrorList {
	var errs ErrorList

	topLevel := map[string]bool{
		"version": true, "extends": true, "project": true, "commit": true,
		"docker": true, "workflows": true, "policies": true,
	}
	errs = append(errs, validateAllowedKeys(file, "", node, topLevel)...)

	errs = append(errs, validateProjectShape(file, mappingValue(node, "project"))...)
	errs = append(errs, validateCommitShape(file, mappingValue(node, "commit"))...)
	errs = append(errs, validateDockerShape(file, mappingValue(node, "docker"))...)
	errs = append(errs, validateWorkflowsShape(file, mappingValue(node, "workflows"))...)
	errs = append(errs, validatePoliciesShape(file, mappingValue(node, "policies"))...)

	return errs
}

func validateFinalConfig(file, rootDir string, node *yaml.Node, lookupEnv func(string) (string, bool)) ErrorList {
	var errs ErrorList

	version := mappingValue(node, "version")
	if version == nil {
		errs = append(errs, ValidationError{File: file, Path: "version", Message: "is required"})
	} else if version.Kind != yaml.ScalarNode || version.Tag != "!!int" || version.Value != "1" {
		errs = append(errs, ValidationError{File: file, Path: "version", Message: "must be the integer 1"})
	}

	project := mappingValue(node, "project")
	if project == nil {
		errs = append(errs, ValidationError{File: file, Path: "project", Message: "is required in the final configuration"})
	} else {
		name := mappingValue(project, "name")
		if !isNonEmptyString(name) {
			errs = append(errs, ValidationError{File: file, Path: "project.name", Message: "must be a non-empty string"})
		}
	}

	errs = append(errs, validateFinalCommit(file, node)...)
	errs = append(errs, validateFinalDockerPaths(file, rootDir, node)...)
	errs = append(errs, validateFinalWorkflows(file, rootDir, node, lookupEnv)...)

	return errs
}

func validateProjectShape(file string, node *yaml.Node) ErrorList {
	if node == nil {
		return nil
	}
	if node.Kind != yaml.MappingNode {
		return ErrorList{{File: file, Path: "project", Message: "must be a mapping"}}
	}
	return validateAllowedKeys(file, "project", node, map[string]bool{"name": true})
}

func validateCommitShape(file string, node *yaml.Node) ErrorList {
	if node == nil {
		return nil
	}
	if node.Kind != yaml.MappingNode {
		return ErrorList{{File: file, Path: "commit", Message: "must be a mapping"}}
	}
	return validateAllowedKeys(file, "commit", node, map[string]bool{
		"enabled": true, "convention": true, "allowed_types": true, "scopes": true,
		"protected_branches": true, "require_scope": true, "emoji": true,
		"confirm_commit": true, "confirm_push": true,
	})
}

func validateDockerShape(file string, node *yaml.Node) ErrorList {
	if node == nil {
		return nil
	}
	if node.Kind != yaml.MappingNode {
		return ErrorList{{File: file, Path: "docker", Message: "must be a mapping"}}
	}
	return validateAllowedKeys(file, "docker", node, map[string]bool{
		"enabled": true, "compose_files": true, "profiles": true, "env_files": true, "project_name": true,
	})
}

func validatePoliciesShape(file string, node *yaml.Node) ErrorList {
	if node == nil {
		return nil
	}
	if node.Kind != yaml.MappingNode {
		return ErrorList{{File: file, Path: "policies", Message: "must be a mapping"}}
	}
	return validateAllowedKeys(file, "policies", node, map[string]bool{
		"require_clean_worktree": true, "confirm_destructive_actions": true,
		"confirm_network_actions": true, "confirm_git_actions": true,
		"allow_shell_steps": true, "allow_non_interactive": true, "audit": true,
	})
}

func validateWorkflowsShape(file string, node *yaml.Node) ErrorList {
	if node == nil {
		return nil
	}
	if node.Kind != yaml.MappingNode {
		return ErrorList{{File: file, Path: "workflows", Message: "must be a mapping"}}
	}

	var errs ErrorList
	seen := map[string]bool{}
	for i := 0; i < len(node.Content); i += 2 {
		name := node.Content[i]
		workflow := node.Content[i+1]
		workflowPath := "workflows." + name.Value

		if name.Kind != yaml.ScalarNode || name.Tag != "!!str" || strings.TrimSpace(name.Value) == "" {
			errs = append(errs, ValidationError{File: file, Path: "workflows", Message: "workflow names must be non-empty strings"})
			continue
		}
		if seen[name.Value] {
			errs = append(errs, ValidationError{File: file, Path: workflowPath, Message: "duplicate workflow name"})
		}
		seen[name.Value] = true

		if !isValidWorkflowName(name.Value) {
			errs = append(errs, ValidationError{File: file, Path: workflowPath, Message: "workflow name must contain only letters, numbers, underscores, and hyphens"})
		}
		if name.Value == "list" || name.Value == "show" {
			errs = append(errs, ValidationError{File: file, Path: workflowPath, Message: "workflow name is reserved by the run command"})
		}

		if workflow.Kind != yaml.MappingNode {
			errs = append(errs, ValidationError{File: file, Path: workflowPath, Message: "must be a mapping"})
			continue
		}
		errs = append(errs, validateAllowedKeys(file, workflowPath, workflow, map[string]bool{
			"description": true, "working_directory": true, "environment": true, "steps": true,
		})...)
		errs = append(errs, validateEnvironmentShape(file, workflowPath+".environment", mappingValue(workflow, "environment"))...)
		errs = append(errs, validateStepsShape(file, workflowPath, mappingValue(workflow, "steps"))...)
	}

	return errs
}

func validateStepsShape(file, workflowPath string, node *yaml.Node) ErrorList {
	if node == nil {
		return nil
	}
	if node.Kind != yaml.SequenceNode {
		return ErrorList{{File: file, Path: workflowPath + ".steps", Message: "must be a list"}}
	}

	var errs ErrorList
	for i, step := range node.Content {
		stepPath := fmt.Sprintf("%s.steps[%d]", workflowPath, i)
		if step.Kind != yaml.MappingNode {
			errs = append(errs, ValidationError{File: file, Path: stepPath, Message: "must be a mapping"})
			continue
		}
		errs = append(errs, validateAllowedKeys(file, stepPath, step, map[string]bool{
			"name": true, "command": true, "args": true, "run": true, "shell": true,
			"working_directory": true, "environment": true, "confirm": true,
			"continue_on_error": true, "timeout": true, "risk": true,
		})...)
		errs = append(errs, validateEnvironmentShape(file, stepPath+".environment", mappingValue(step, "environment"))...)
	}
	return errs
}

func validateEnvironmentShape(file, path string, node *yaml.Node) ErrorList {
	if node == nil {
		return nil
	}
	if node.Kind != yaml.MappingNode {
		return ErrorList{{File: file, Path: path, Message: "must be a mapping of strings"}}
	}

	var errs ErrorList
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		value := node.Content[i+1]
		if !isNonEmptyString(key) {
			errs = append(errs, ValidationError{File: file, Path: path, Message: "environment variable names must be non-empty strings"})
			continue
		}
		if !environmentNamePattern.MatchString(key.Value) {
			errs = append(errs, ValidationError{File: file, Path: path + "." + key.Value, Message: "environment variable name must match [A-Za-z_][A-Za-z0-9_]*"})
		}
		if value.Kind != yaml.ScalarNode || value.Tag != "!!str" {
			errs = append(errs, ValidationError{File: file, Path: path + "." + key.Value, Message: "must be a string"})
		}
	}
	return errs
}

func validateFinalCommit(file string, node *yaml.Node) ErrorList {
	commit := mappingValue(node, "commit")
	if commit == nil {
		return nil
	}

	var errs ErrorList
	if convention := mappingValue(commit, "convention"); convention != nil {
		if convention.Kind != yaml.ScalarNode || convention.Tag != "!!str" || convention.Value != "conventional" {
			errs = append(errs, ValidationError{File: file, Path: "commit.convention", Message: "must be conventional"})
		}
	}

	requireScope := mappingValue(commit, "require_scope")
	if requireScope != nil && requireScope.Tag != "!!bool" {
		errs = append(errs, ValidationError{File: file, Path: "commit.require_scope", Message: "must be a boolean"})
	}
	if requireScope != nil && requireScope.Value == "true" {
		if scopes := mappingValue(commit, "scopes"); !isNonEmptyStringList(scopes) {
			errs = append(errs, ValidationError{File: file, Path: "commit.scopes", Message: "must contain at least one scope when commit.require_scope is true"})
		}
	}

	errs = append(errs, validateOptionalStringList(file, "commit.allowed_types", mappingValue(commit, "allowed_types"), true)...)
	errs = append(errs, validateOptionalStringList(file, "commit.scopes", mappingValue(commit, "scopes"), false)...)
	errs = append(errs, validateOptionalStringList(file, "commit.protected_branches", mappingValue(commit, "protected_branches"), true)...)
	errs = append(errs, validateOptionalBool(file, "commit.enabled", mappingValue(commit, "enabled"))...)
	errs = append(errs, validateOptionalBool(file, "commit.emoji", mappingValue(commit, "emoji"))...)
	errs = append(errs, validateOptionalBool(file, "commit.confirm_commit", mappingValue(commit, "confirm_commit"))...)
	errs = append(errs, validateOptionalBool(file, "commit.confirm_push", mappingValue(commit, "confirm_push"))...)

	return errs
}

func validateFinalDockerPaths(file, rootDir string, node *yaml.Node) ErrorList {
	docker := mappingValue(node, "docker")
	if docker == nil {
		return nil
	}
	var errs ErrorList
	errs = append(errs, validateOptionalBool(file, "docker.enabled", mappingValue(docker, "enabled"))...)
	errs = append(errs, validateOptionalString(file, "docker.project_name", mappingValue(docker, "project_name"))...)
	errs = append(errs, validateOptionalStringList(file, "docker.profiles", mappingValue(docker, "profiles"), false)...)
	errs = append(errs, validateProjectLocalPathList(file, rootDir, "docker.compose_files", mappingValue(docker, "compose_files"))...)
	errs = append(errs, validateProjectLocalPathList(file, rootDir, "docker.env_files", mappingValue(docker, "env_files"))...)
	return errs
}

func validateFinalWorkflows(file, rootDir string, node *yaml.Node, lookupEnv func(string) (string, bool)) ErrorList {
	workflows := mappingValue(node, "workflows")
	if workflows == nil {
		return nil
	}

	var errs ErrorList
	for i := 0; i < len(workflows.Content); i += 2 {
		name := workflows.Content[i].Value
		workflow := workflows.Content[i+1]
		workflowPath := "workflows." + name
		if workflow.Kind != yaml.MappingNode {
			continue
		}

		if !isNonEmptyString(mappingValue(workflow, "description")) {
			errs = append(errs, ValidationError{File: file, Path: workflowPath + ".description", Message: "must be a non-empty string"})
		}

		steps := mappingValue(workflow, "steps")
		if steps == nil || steps.Kind != yaml.SequenceNode || len(steps.Content) == 0 {
			errs = append(errs, ValidationError{File: file, Path: workflowPath + ".steps", Message: "must contain at least one step"})
		}

		errs = append(errs, validateProjectLocalPath(file, rootDir, workflowPath+".working_directory", mappingValue(workflow, "working_directory"))...)
		errs = append(errs, validateEnvironmentReferences(file, workflowPath+".environment", mappingValue(workflow, "environment"), lookupEnv)...)

		if steps == nil || steps.Kind != yaml.SequenceNode {
			continue
		}
		for stepIndex, step := range steps.Content {
			if step.Kind != yaml.MappingNode {
				continue
			}
			stepPath := fmt.Sprintf("%s.steps[%d]", workflowPath, stepIndex)
			command := mappingValue(step, "command")
			run := mappingValue(step, "run")
			shell := mappingValue(step, "shell")
			confirm := mappingValue(step, "confirm")
			if command == nil && run == nil {
				errs = append(errs, ValidationError{File: file, Path: stepPath, Message: "must define command with optional args, or an explicitly enabled shell run"})
			}
			if command != nil && run != nil {
				errs = append(errs, ValidationError{File: file, Path: stepPath, Message: "command and run are mutually exclusive"})
			}
			if command != nil {
				errs = append(errs, validateOptionalString(file, stepPath+".command", command)...)
				if isNonEmptyString(command) && !strings.ContainsAny(command.Value, `/\`) && len(strings.Fields(command.Value)) != 1 {
					errs = append(errs, ValidationError{File: file, Path: stepPath + ".command", Message: "must name one executable; place parameters in args"})
				}
				errs = append(errs, validateOptionalStringList(file, stepPath+".args", mappingValue(step, "args"), false)...)
				if shell != nil && shell.Value == "true" {
					errs = append(errs, ValidationError{File: file, Path: stepPath + ".shell", Message: "must be false for structured command steps"})
				}
			}
			if run != nil {
				errs = append(errs, validateOptionalString(file, stepPath+".run", run)...)
				if mappingValue(step, "args") != nil {
					errs = append(errs, ValidationError{File: file, Path: stepPath + ".args", Message: "is only supported with a structured command"})
				}
				if shell == nil || shell.Tag != "!!bool" || shell.Value != "true" {
					errs = append(errs, ValidationError{File: file, Path: stepPath + ".shell", Message: "must be true for a run string"})
				}
				if confirm == nil || confirm.Tag != "!!bool" || confirm.Value != "true" {
					errs = append(errs, ValidationError{File: file, Path: stepPath + ".confirm", Message: "must be true for a shell step"})
				}
				policies := mappingValue(node, "policies")
				allowShell := mappingValue(policies, "allow_shell_steps")
				if allowShell == nil || allowShell.Tag != "!!bool" || allowShell.Value != "true" {
					errs = append(errs, ValidationError{File: file, Path: "policies.allow_shell_steps", Message: "must be true when shell steps are configured"})
				}
			}
			errs = append(errs, validateProjectLocalPath(file, rootDir, stepPath+".working_directory", mappingValue(step, "working_directory"))...)
			errs = append(errs, validateEnvironmentReferences(file, stepPath+".environment", mappingValue(step, "environment"), lookupEnv)...)
			errs = append(errs, validateOptionalBool(file, stepPath+".confirm", mappingValue(step, "confirm"))...)
			errs = append(errs, validateOptionalBool(file, stepPath+".continue_on_error", mappingValue(step, "continue_on_error"))...)
			errs = append(errs, validateOptionalBool(file, stepPath+".shell", shell)...)
			if timeoutNode := mappingValue(step, "timeout"); timeoutNode != nil {
				if !isNonEmptyString(timeoutNode) {
					errs = append(errs, ValidationError{File: file, Path: stepPath + ".timeout", Message: "must be a Go duration such as 30s or 5m"})
				} else if duration, err := time.ParseDuration(timeoutNode.Value); err != nil || duration <= 0 {
					errs = append(errs, ValidationError{File: file, Path: stepPath + ".timeout", Message: "must be a positive Go duration such as 30s or 5m"})
				}
			}
			if riskNode := mappingValue(step, "risk"); riskNode != nil {
				if !isNonEmptyString(riskNode) || !isAllowedRisk(riskNode.Value) {
					errs = append(errs, ValidationError{File: file, Path: stepPath + ".risk", Message: "must be one of read, local-write, network, git-write, or destructive"})
				}
			}
		}
	}

	policies := mappingValue(node, "policies")
	for _, key := range []string{"require_clean_worktree", "confirm_destructive_actions", "confirm_network_actions", "confirm_git_actions", "allow_shell_steps", "allow_non_interactive", "audit"} {
		errs = append(errs, validateOptionalBool(file, "policies."+key, mappingValue(policies, key))...)
	}

	return errs
}

func validateAllowedKeys(file, basePath string, node *yaml.Node, allowed map[string]bool) ErrorList {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}

	var errs ErrorList
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		path := key.Value
		if basePath != "" {
			path = basePath + "." + key.Value
		}
		if !allowed[key.Value] {
			errs = append(errs, ValidationError{File: file, Path: path, Message: "unknown field"})
		}
	}
	return errs
}

func validateOptionalBool(file, path string, node *yaml.Node) ErrorList {
	if node == nil {
		return nil
	}
	if node.Kind != yaml.ScalarNode || node.Tag != "!!bool" {
		return ErrorList{{File: file, Path: path, Message: "must be a boolean"}}
	}
	return nil
}

func validateOptionalString(file, path string, node *yaml.Node) ErrorList {
	if node == nil {
		return nil
	}
	if !isNonEmptyString(node) {
		return ErrorList{{File: file, Path: path, Message: "must be a non-empty string"}}
	}
	return nil
}

func validateOptionalStringList(file, path string, node *yaml.Node, requireNonEmptyItems bool) ErrorList {
	if node == nil {
		return nil
	}
	if node.Kind != yaml.SequenceNode {
		return ErrorList{{File: file, Path: path, Message: "must be a list of strings"}}
	}
	var errs ErrorList
	for i, item := range node.Content {
		itemPath := fmt.Sprintf("%s[%d]", path, i)
		if item.Kind != yaml.ScalarNode || item.Tag != "!!str" {
			errs = append(errs, ValidationError{File: file, Path: itemPath, Message: "must be a string"})
			continue
		}
		if requireNonEmptyItems && strings.TrimSpace(item.Value) == "" {
			errs = append(errs, ValidationError{File: file, Path: itemPath, Message: "must not be empty"})
		}
	}
	return errs
}

func validateProjectLocalPathList(file, rootDir, path string, node *yaml.Node) ErrorList {
	if node == nil {
		return nil
	}
	if node.Kind != yaml.SequenceNode {
		return ErrorList{{File: file, Path: path, Message: "must be a list of project-local paths"}}
	}
	var errs ErrorList
	for i, item := range node.Content {
		errs = append(errs, validateProjectLocalPath(file, rootDir, fmt.Sprintf("%s[%d]", path, i), item)...)
	}
	return errs
}

func validateProjectLocalPath(file, rootDir, path string, node *yaml.Node) ErrorList {
	if node == nil {
		return nil
	}
	if !isNonEmptyString(node) {
		return ErrorList{{File: file, Path: path, Message: "must be a non-empty project-local path"}}
	}
	if hasURLScheme(node.Value) {
		return ErrorList{{File: file, Path: path, Message: "remote paths are not supported in v1"}}
	}
	candidate := node.Value
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(rootDir, candidate)
	}
	if _, err := ResolveProjectPath(rootDir, candidate); err != nil {
		return ErrorList{{File: file, Path: path, Message: "must not escape the project root"}}
	}
	return nil
}

func validateEnvironmentReferences(file, path string, node *yaml.Node, lookupEnv func(string) (string, bool)) ErrorList {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	var errs ErrorList
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i].Value
		value := node.Content[i+1]
		if value.Kind != yaml.ScalarNode || value.Tag != "!!str" {
			continue
		}
		for _, match := range environmentReferencePattern.FindAllStringSubmatch(value.Value, -1) {
			if _, ok := lookupEnv(match[1]); !ok {
				errs = append(errs, ValidationError{File: file, Path: path + "." + key, Message: fmt.Sprintf("references unresolved environment variable %s", match[1])})
			}
		}
	}
	return errs
}

func isAllowedRisk(value string) bool {
	switch value {
	case RiskRead, RiskLocalWrite, RiskNetwork, RiskGitWrite, RiskDestructive:
		return true
	default:
		return false
	}
}

func isNonEmptyString(node *yaml.Node) bool {
	return node != nil && node.Kind == yaml.ScalarNode && node.Tag == "!!str" && strings.TrimSpace(node.Value) != ""
}

func isNonEmptyStringList(node *yaml.Node) bool {
	if node == nil || node.Kind != yaml.SequenceNode || len(node.Content) == 0 {
		return false
	}
	for _, item := range node.Content {
		if !isNonEmptyString(item) {
			return false
		}
	}
	return true
}

func isValidWorkflowName(name string) bool {
	if strings.TrimSpace(name) == "" {
		return false
	}
	for _, char := range name {
		if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || char == '_' || char == '-' {
			continue
		}
		return false
	}
	return true
}
