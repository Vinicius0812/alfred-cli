package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/Vinicius0812/alfred-cli/internal/config"
)

var runIDPattern = regexp.MustCompile(`^[0-9]{8}T[0-9]{6}\.[0-9]{9}Z(?:-[a-f0-9]{8})?$`)

func auditDirectory(projectRoot string) (string, error) {
	directory, err := config.ResolveProjectPath(projectRoot, filepath.Join(".alfred", "runs"))
	if err != nil {
		return "", fmt.Errorf("invalid audit directory: %w", err)
	}
	return directory, nil
}

func WriteAudit(projectRoot string, execution Execution) error {
	if !runIDPattern.MatchString(execution.ID) {
		return fmt.Errorf("invalid run id")
	}
	directory, err := auditDirectory(projectRoot)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return fmt.Errorf("could not create audit directory: %w", err)
	}
	content, err := json.MarshalIndent(execution, "", "  ")
	if err != nil {
		return fmt.Errorf("could not encode audit record: %w", err)
	}
	content = append(content, '\n')
	target := filepath.Join(directory, execution.ID+".json")
	temporary, err := os.CreateTemp(directory, ".audit-*.tmp")
	if err != nil {
		return fmt.Errorf("could not create audit record: %w", err)
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(content); err != nil {
		temporary.Close()
		return fmt.Errorf("could not write audit record: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return fmt.Errorf("could not sync audit record: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("could not close audit record: %w", err)
	}
	if err := os.Rename(temporaryName, target); err != nil {
		return fmt.Errorf("could not publish audit record: %w", err)
	}
	return nil
}

func LoadAudit(projectRoot, id string) (Execution, error) {
	if !runIDPattern.MatchString(id) {
		return Execution{}, fmt.Errorf("invalid run id")
	}
	directory, err := auditDirectory(projectRoot)
	if err != nil {
		return Execution{}, err
	}
	path := filepath.Join(directory, id+".json")
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Execution{}, fmt.Errorf("audit record not found: %s", id)
		}
		return Execution{}, fmt.Errorf("could not read audit record: %w", err)
	}
	var execution Execution
	if err := json.Unmarshal(content, &execution); err != nil {
		return Execution{}, fmt.Errorf("invalid audit record: %w", err)
	}
	return execution, nil
}

func ListAudits(projectRoot string) ([]Execution, error) {
	directory, err := auditDirectory(projectRoot)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(directory)
	if os.IsNotExist(err) {
		return []Execution{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("could not list audit records: %w", err)
	}
	var executions []Execution
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		id := entry.Name()[:len(entry.Name())-len(".json")]
		if !runIDPattern.MatchString(id) {
			continue
		}
		execution, err := LoadAudit(projectRoot, id)
		if err != nil {
			return nil, err
		}
		executions = append(executions, execution)
	}
	sort.Slice(executions, func(i, j int) bool {
		return executions[i].StartedAt.After(executions[j].StartedAt)
	})
	return executions, nil
}
