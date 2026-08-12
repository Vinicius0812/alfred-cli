package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type Options struct {
	ConfigPath           string
	WorkingDir           string
	Getenv               func(string) string
	LookupEnv            func(string) (string, bool)
	AllowExternalExtends bool
	MaxConfigBytes       int64
}

type Result struct {
	RootFile      string
	RootDir       string
	ProjectName   string
	RequiredTools []string
	Files         []string
	Config        Config
	Errors        ErrorList
}

const (
	defaultMaxConfigBytes int64 = 1 << 20
	maxConfigFiles              = 64
)

type loadedFile struct {
	path string
	node *yaml.Node
}

func LoadAndValidate(opts Options) (Result, error) {
	if opts.WorkingDir == "" {
		opts.WorkingDir = "."
	}
	if opts.LookupEnv == nil {
		if opts.Getenv != nil {
			opts.LookupEnv = func(key string) (string, bool) {
				value := opts.Getenv(key)
				return value, value != ""
			}
		} else {
			opts.LookupEnv = os.LookupEnv
		}
	}
	if opts.MaxConfigBytes <= 0 {
		opts.MaxConfigBytes = defaultMaxConfigBytes
	}

	rootFile, err := resolveRootFile(opts.ConfigPath, opts.WorkingDir)
	if err != nil {
		return Result{}, err
	}

	rootFile, err = filepath.Abs(rootFile)
	if err != nil {
		return Result{}, err
	}

	rootFile, err = filepath.EvalSymlinks(rootFile)
	if err != nil {
		return Result{}, fmt.Errorf("could not resolve configuration file: %w", err)
	}
	rootFile = filepath.Clean(rootFile)
	rootDir := filepath.Dir(rootFile)

	state := loadState{
		visiting:             map[string]bool{},
		loaded:               map[string]bool{},
		rootDir:              rootDir,
		allowExternalExtends: opts.AllowExternalExtends,
		maxConfigBytes:       opts.MaxConfigBytes,
	}

	rootNode, files, validationErrors := state.loadGraph(rootFile)
	result := Result{
		RootFile: rootFile,
		RootDir:  rootDir,
		Errors:   validationErrors,
	}
	for _, file := range files {
		result.Files = append(result.Files, file.path)
	}

	if rootNode == nil {
		return result, nil
	}

	for _, file := range files {
		result.Errors = append(result.Errors, validateKnownShape(file.path, file.node)...)
	}

	result.Errors = append(result.Errors, validateFinalConfig(rootFile, rootDir, rootNode, opts.LookupEnv)...)
	result.ProjectName = projectName(rootNode)
	result.RequiredTools = requiredTools(rootNode)
	if len(result.Errors) == 0 {
		resolved := defaultConfig()
		if err := rootNode.Decode(&resolved); err != nil {
			result.Errors = append(result.Errors, ValidationError{File: rootFile, Message: fmt.Sprintf("could not decode resolved configuration: %v", err)})
		} else {
			applyModelDefaults(&resolved)
			result.Config = resolved
		}
	}
	return result, nil
}

func resolveRootFile(configPath, workingDir string) (string, error) {
	if configPath != "" {
		if hasURLScheme(configPath) {
			return "", fmt.Errorf("remote configuration URLs are not supported in v1")
		}
		if !filepath.IsAbs(configPath) {
			configPath = filepath.Join(workingDir, configPath)
		}
		if _, err := os.Stat(configPath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return "", fmt.Errorf("configuration file not found: %s", configPath)
			}
			return "", fmt.Errorf("could not read configuration file: %w", err)
		}
		return configPath, nil
	}

	dir, err := filepath.Abs(workingDir)
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(dir, "alfred.yaml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("alfred.yaml not found in current directory or parents")
}

type loadState struct {
	visiting             map[string]bool
	loaded               map[string]bool
	files                []loadedFile
	rootDir              string
	allowExternalExtends bool
	maxConfigBytes       int64
	fileCount            int
}

func (s *loadState) loadGraph(path string) (*yaml.Node, []loadedFile, ErrorList) {
	node, errs := s.loadFile(path)
	if node == nil {
		return nil, s.files, errs
	}
	return node, s.files, errs
}

func (s *loadState) loadFile(path string) (*yaml.Node, ErrorList) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, ErrorList{{File: path, Message: err.Error()}}
	}
	absPath, err = filepath.EvalSymlinks(absPath)
	if err != nil {
		return nil, ErrorList{{File: path, Message: fmt.Sprintf("could not resolve file: %v", err)}}
	}
	absPath = filepath.Clean(absPath)

	if s.visiting[absPath] {
		return nil, ErrorList{{File: absPath, Path: "extends", Message: "cyclic extends reference detected"}}
	}
	if s.loaded[absPath] {
		return emptyMapNode(), nil
	}
	if s.fileCount >= maxConfigFiles {
		return nil, ErrorList{{File: absPath, Path: "extends", Message: fmt.Sprintf("configuration graph exceeds the %d file limit", maxConfigFiles)}}
	}
	s.fileCount++

	file, err := os.Open(absPath)
	if err != nil {
		return nil, ErrorList{{File: absPath, Message: fmt.Sprintf("could not read file: %v", err)}}
	}
	defer file.Close()

	content, err := io.ReadAll(io.LimitReader(file, s.maxConfigBytes+1))
	if err != nil {
		return nil, ErrorList{{File: absPath, Message: fmt.Sprintf("could not read file: %v", err)}}
	}
	if int64(len(content)) > s.maxConfigBytes {
		return nil, ErrorList{{File: absPath, Message: fmt.Sprintf("configuration exceeds the %d byte limit", s.maxConfigBytes)}}
	}

	var document yaml.Node
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	if err := decoder.Decode(&document); err != nil {
		return nil, ErrorList{{File: absPath, Message: fmt.Sprintf("invalid YAML: %v", err)}}
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return nil, ErrorList{{File: absPath, Message: "configuration must contain exactly one YAML document"}}
		}
		return nil, ErrorList{{File: absPath, Message: fmt.Sprintf("invalid YAML: %v", err)}}
	}

	if len(document.Content) == 0 {
		return nil, ErrorList{{File: absPath, Message: "configuration must be a YAML mapping"}}
	}

	current := document.Content[0]
	if current.Kind != yaml.MappingNode {
		return nil, ErrorList{{File: absPath, Message: "configuration must be a YAML mapping"}}
	}
	if securityErrors := validateYAMLSecurity(absPath, current); len(securityErrors) > 0 {
		return nil, securityErrors
	}

	s.visiting[absPath] = true
	defer delete(s.visiting, absPath)

	merged := emptyMapNode()
	var errs ErrorList
	extends := mappingValue(current, "extends")
	if extends != nil {
		paths, pathErrs := s.readExtends(absPath, extends)
		errs = append(errs, pathErrs...)
		for _, extendedPath := range paths {
			extendedNode, extendedErrs := s.loadFile(extendedPath)
			errs = append(errs, extendedErrs...)
			if extendedNode != nil {
				merged = mergeMaps(merged, extendedNode)
			}
		}
	}

	merged = mergeMaps(merged, current)
	s.files = append(s.files, loadedFile{path: absPath, node: current})
	s.loaded[absPath] = true

	return merged, errs
}

func (s *loadState) readExtends(file string, node *yaml.Node) ([]string, ErrorList) {
	if node.Kind != yaml.SequenceNode {
		return nil, ErrorList{{File: file, Path: "extends", Message: "must be a list of local file paths"}}
	}

	paths := make([]string, 0, len(node.Content))
	var errs ErrorList
	for i, item := range node.Content {
		path := fmt.Sprintf("extends[%d]", i)
		if item.Kind != yaml.ScalarNode || item.Tag != "!!str" || strings.TrimSpace(item.Value) == "" {
			errs = append(errs, ValidationError{File: file, Path: path, Message: "must be a non-empty local file path"})
			continue
		}
		if hasURLScheme(item.Value) {
			errs = append(errs, ValidationError{File: file, Path: path, Message: "remote URLs are not supported in v1"})
			continue
		}

		extendedPath := item.Value
		if !filepath.IsAbs(extendedPath) {
			extendedPath = filepath.Join(filepath.Dir(file), extendedPath)
		}
		if _, err := os.Stat(extendedPath); err != nil {
			errs = append(errs, ValidationError{File: file, Path: path, Message: fmt.Sprintf("extended file not found: %s", item.Value)})
			continue
		}
		resolvedPath, err := filepath.EvalSymlinks(extendedPath)
		if err != nil {
			errs = append(errs, ValidationError{File: file, Path: path, Message: fmt.Sprintf("could not resolve extended file: %v", err)})
			continue
		}
		resolvedPath = filepath.Clean(resolvedPath)
		if !s.allowExternalExtends && !pathWithin(s.rootDir, resolvedPath) {
			errs = append(errs, ValidationError{File: file, Path: path, Message: "must remain inside the project root; use --allow-external-extends only for a trusted preset"})
			continue
		}
		paths = append(paths, resolvedPath)
	}

	return paths, errs
}

func hasURLScheme(value string) bool {
	lower := strings.ToLower(strings.TrimSpace(value))
	return strings.Contains(lower, "://")
}

func emptyMapNode() *yaml.Node {
	return &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
}

func mergeMaps(base, overlay *yaml.Node) *yaml.Node {
	if base == nil {
		return cloneNode(overlay)
	}
	if overlay == nil {
		return cloneNode(base)
	}
	if base.Kind != yaml.MappingNode || overlay.Kind != yaml.MappingNode {
		return cloneNode(overlay)
	}

	out := cloneNode(base)
	for i := 0; i < len(overlay.Content); i += 2 {
		key := overlay.Content[i]
		value := overlay.Content[i+1]
		index := mappingKeyIndex(out, key.Value)
		if index == -1 {
			out.Content = append(out.Content, cloneNode(key), cloneNode(value))
			continue
		}

		existing := out.Content[index+1]
		if existing.Kind == yaml.MappingNode && value.Kind == yaml.MappingNode {
			out.Content[index+1] = mergeMaps(existing, value)
			continue
		}
		out.Content[index+1] = cloneNode(value)
	}

	return out
}

func cloneNode(node *yaml.Node) *yaml.Node {
	if node == nil {
		return nil
	}
	clone := *node
	clone.Content = make([]*yaml.Node, len(node.Content))
	for i, child := range node.Content {
		clone.Content[i] = cloneNode(child)
	}
	return &clone
}

func mappingKeyIndex(node *yaml.Node, key string) int {
	if node == nil || node.Kind != yaml.MappingNode {
		return -1
	}
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return i
		}
	}
	return -1
}

func mappingValue(node *yaml.Node, key string) *yaml.Node {
	index := mappingKeyIndex(node, key)
	if index == -1 {
		return nil
	}
	return node.Content[index+1]
}

func projectName(node *yaml.Node) string {
	project := mappingValue(node, "project")
	if project == nil {
		return ""
	}
	name := mappingValue(project, "name")
	if name == nil || name.Kind != yaml.ScalarNode || name.Tag != "!!str" {
		return ""
	}
	return strings.TrimSpace(name.Value)
}

func requiredTools(node *yaml.Node) []string {
	tools := map[string]bool{}
	docker := mappingValue(node, "docker")
	if docker != nil {
		if mappingValue(docker, "compose_files") != nil {
			tools["docker"] = true
		}
	}

	workflows := mappingValue(node, "workflows")
	if workflows != nil && workflows.Kind == yaml.MappingNode {
		for i := 0; i < len(workflows.Content); i += 2 {
			workflow := workflows.Content[i+1]
			if workflow.Kind != yaml.MappingNode {
				continue
			}
			steps := mappingValue(workflow, "steps")
			if steps == nil || steps.Kind != yaml.SequenceNode {
				continue
			}
			for _, step := range steps.Content {
				if step.Kind != yaml.MappingNode {
					continue
				}
				command := mappingValue(step, "command")
				if command == nil || command.Kind != yaml.ScalarNode || command.Tag != "!!str" {
					continue
				}
				for _, tool := range toolsFromStructuredCommand(command.Value) {
					tools[tool] = true
				}
			}
		}
	}

	values := make([]string, 0, len(tools))
	for tool := range tools {
		values = append(values, tool)
	}
	sort.Strings(values)
	return values
}

func applyModelDefaults(resolved *Config) {
	for name, workflow := range resolved.Workflows {
		if workflow.Environment == nil {
			workflow.Environment = map[string]string{}
		}
		for i := range workflow.Steps {
			if workflow.Steps[i].Environment == nil {
				workflow.Steps[i].Environment = map[string]string{}
			}
			if workflow.Steps[i].Risk == "" {
				workflow.Steps[i].Risk = RiskLocalWrite
			}
		}
		resolved.Workflows[name] = workflow
	}
}

func toolsFromStructuredCommand(command string) []string {
	command = strings.TrimSpace(command)
	if command == "" || strings.ContainsAny(command, `/\`) {
		return nil
	}
	switch command {
	case "npm":
		return []string{"node", "npm"}
	case "composer":
		return []string{"php", "composer"}
	default:
		return []string{command}
	}
}
