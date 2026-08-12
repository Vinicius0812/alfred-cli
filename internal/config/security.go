package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	maxYAMLDepth = 64
	maxYAMLNodes = 10000
)

func validateYAMLSecurity(file string, root *yaml.Node) ErrorList {
	var errs ErrorList
	nodes := 0
	var walk func(*yaml.Node, string, int)
	walk = func(node *yaml.Node, path string, depth int) {
		if node == nil || len(errs) > 50 {
			return
		}
		nodes++
		if nodes > maxYAMLNodes {
			errs = append(errs, ValidationError{File: file, Path: path, Message: fmt.Sprintf("YAML exceeds the %d node limit", maxYAMLNodes)})
			return
		}
		if depth > maxYAMLDepth {
			errs = append(errs, ValidationError{File: file, Path: path, Message: fmt.Sprintf("YAML exceeds the maximum depth of %d", maxYAMLDepth)})
			return
		}
		if node.Kind == yaml.AliasNode || node.Alias != nil {
			errs = append(errs, ValidationError{File: file, Path: path, Message: "YAML aliases are not supported"})
			return
		}
		if strings.HasPrefix(node.Tag, "!") && !strings.HasPrefix(node.Tag, "!!") {
			errs = append(errs, ValidationError{File: file, Path: path, Message: fmt.Sprintf("custom YAML tag %q is not supported", node.Tag)})
		}

		if node.Kind == yaml.MappingNode {
			seen := map[string]bool{}
			for i := 0; i+1 < len(node.Content); i += 2 {
				key := node.Content[i]
				value := node.Content[i+1]
				keyPath := key.Value
				if path != "" {
					keyPath = path + "." + key.Value
				}
				if key.Kind != yaml.ScalarNode || key.Tag != "!!str" {
					errs = append(errs, ValidationError{File: file, Path: path, Message: "mapping keys must be strings"})
				} else if key.Value == "<<" {
					errs = append(errs, ValidationError{File: file, Path: keyPath, Message: "YAML merge keys are not supported"})
				} else if seen[key.Value] {
					errs = append(errs, ValidationError{File: file, Path: keyPath, Message: "duplicate YAML key"})
				} else {
					seen[key.Value] = true
				}
				walk(key, keyPath, depth+1)
				walk(value, keyPath, depth+1)
			}
			return
		}

		for i, child := range node.Content {
			childPath := path
			if node.Kind == yaml.SequenceNode {
				childPath = fmt.Sprintf("%s[%d]", path, i)
			}
			walk(child, childPath, depth+1)
		}
	}

	walk(root, "", 0)
	return errs
}

func pathWithin(root, candidate string) bool {
	rel, err := filepath.Rel(filepath.Clean(root), filepath.Clean(candidate))
	if err != nil {
		return false
	}
	return rel == "." || (!filepath.IsAbs(rel) && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

func ResolveProjectPath(root, value string) (string, error) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	rootResolved, err := filepath.EvalSymlinks(rootAbs)
	if err != nil {
		return "", fmt.Errorf("could not resolve project root: %w", err)
	}

	candidate := value
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(rootResolved, candidate)
	}
	candidate, err = filepath.Abs(candidate)
	if err != nil {
		return "", err
	}

	resolvedCandidate, err := resolveWithExistingAncestor(candidate)
	if err != nil {
		return "", err
	}
	if !pathWithin(rootResolved, resolvedCandidate) {
		return "", fmt.Errorf("path escapes the project root")
	}
	return filepath.Clean(resolvedCandidate), nil
}

func resolveWithExistingAncestor(path string) (string, error) {
	path = filepath.Clean(path)
	current := path
	var suffix []string

	for {
		_, err := os.Lstat(current)
		if err == nil {
			resolved, resolveErr := filepath.EvalSymlinks(current)
			if resolveErr != nil {
				return "", resolveErr
			}
			for i := len(suffix) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, suffix[i])
			}
			return filepath.Clean(resolved), nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("could not find an existing ancestor for %s", path)
		}
		suffix = append(suffix, filepath.Base(current))
		current = parent
	}
}
