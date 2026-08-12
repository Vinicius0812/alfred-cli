package cli

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"

	"gopkg.in/yaml.v3"
)

type outputFormat string

const (
	outputText outputFormat = "text"
	outputJSON outputFormat = "json"
)

func parseOutput(args []string) (outputFormat, []string, error) {
	format := outputText
	remaining := make([]string, 0, len(args))
	for index := 0; index < len(args); index++ {
		if args[index] != "--output" {
			remaining = append(remaining, args[index])
			continue
		}
		if index+1 >= len(args) {
			return "", nil, fmt.Errorf("--output requires a value")
		}
		index++
		format = outputFormat(args[index])
	}
	if format != outputText && format != outputJSON {
		return "", nil, fmt.Errorf("output must be text or json")
	}
	return format, remaining, nil
}

func writeJSON(w interface{ Write([]byte) (int, error) }, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func (a *application) runConfig(args []string, global globalOptions, msg messages) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Fprintln(a.stdout, msg.text("config_usage"))
		return exitSuccess
	}
	if args[0] != "validate" && args[0] != "explain" {
		fmt.Fprintln(a.stderr, msg.text("unknown_config", args[0]))
		return exitUsage
	}
	format, remaining, err := parseOutput(args[1:])
	if err != nil {
		fmt.Fprintf(a.stderr, "alfred: %s\n", localizeDiagnostic(err.Error(), global.lang))
		return exitUsage
	}
	if len(remaining) != 0 {
		if args[0] == "validate" {
			fmt.Fprintln(a.stderr, msg.text("config_extra_arg"))
		} else {
			fmt.Fprintln(a.stderr, msg.text("config_usage"))
		}
		return exitUsage
	}
	result, ok := a.loadConfig(global, msg)
	if !ok {
		return exitFailure
	}

	switch args[0] {
	case "validate":
		if format == outputJSON {
			_ = writeJSON(a.stdout, map[string]any{"valid": true, "project": result.ProjectName, "source": result.RootFile, "files": result.Files})
		} else if result.ProjectName != "" {
			fmt.Fprintln(a.stdout, msg.text("config_valid", result.ProjectName))
		} else {
			fmt.Fprintln(a.stdout, msg.text("config_valid_generic"))
		}
		return exitSuccess
	case "explain":
		if format == outputJSON {
			_ = writeJSON(a.stdout, map[string]any{"source": result.RootFile, "root": result.RootDir, "files": result.Files, "config": result.Config})
			return exitSuccess
		}
		resolved, marshalErr := yaml.Marshal(result.Config)
		if marshalErr != nil {
			fmt.Fprintf(a.stderr, "alfred: %s\n", localizeDiagnostic("could not encode configuration: "+marshalErr.Error(), global.lang))
			return exitFailure
		}
		fmt.Fprintln(a.stdout, msg.text("source", result.RootFile))
		fmt.Fprintln(a.stdout, msg.text("loaded_files"))
		for _, file := range result.Files {
			fmt.Fprintf(a.stdout, "  - %s\n", file)
		}
		fmt.Fprintln(a.stdout, msg.text("resolved_config"))
		fmt.Fprint(a.stdout, string(resolved))
		return exitSuccess
	}
	return exitSuccess
}

func (a *application) runWorkflow(args []string, global globalOptions, msg messages) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Fprintln(a.stdout, msg.text("workflow_usage"))
		return exitSuccess
	}
	if args[0] != "list" {
		fmt.Fprintln(a.stderr, msg.text("unknown_workflow_argument", args[0]))
		return exitUsage
	}
	format, remaining, err := parseOutput(args[1:])
	if err != nil || len(remaining) != 0 {
		if err != nil {
			fmt.Fprintf(a.stderr, "alfred: %s\n", localizeDiagnostic(err.Error(), global.lang))
		} else {
			fmt.Fprintln(a.stderr, msg.text("workflow_usage"))
		}
		return exitUsage
	}
	result, ok := a.loadConfig(global, msg)
	if !ok {
		return exitFailure
	}
	names := make([]string, 0, len(result.Config.Workflows))
	for name := range result.Config.Workflows {
		names = append(names, name)
	}
	sort.Strings(names)
	if format == outputJSON {
		type workflowSummary struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Steps       int    `json:"steps"`
		}
		items := make([]workflowSummary, 0, len(names))
		for _, name := range names {
			configured := result.Config.Workflows[name]
			items = append(items, workflowSummary{Name: name, Description: configured.Description, Steps: len(configured.Steps)})
		}
		_ = writeJSON(a.stdout, items)
		return exitSuccess
	}
	if len(names) == 0 {
		fmt.Fprintln(a.stdout, msg.text("workflow_none"))
		return exitSuccess
	}
	fmt.Fprintln(a.stdout, msg.text("workflow_title"))
	for _, name := range names {
		configured := result.Config.Workflows[name]
		if len(configured.Steps) == 1 {
			fmt.Fprintln(a.stdout, msg.text("workflow_item_one", name, configured.Description))
		} else {
			fmt.Fprintln(a.stdout, msg.text("workflow_item", name, configured.Description, len(configured.Steps)))
		}
	}
	return exitSuccess
}

func (a *application) runDoctor(args []string, global globalOptions, msg messages) int {
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Fprintln(a.stdout, msg.text("doctor_usage"))
		return exitSuccess
	}
	if len(args) > 0 {
		fmt.Fprintln(a.stderr, msg.text("unknown_argument", args[0]))
		return exitUsage
	}
	fmt.Fprintln(a.stdout, msg.text("doctor_title"))
	result, ok := a.loadConfig(global, msg)
	if !ok {
		fmt.Fprintln(a.stdout, msg.text("doctor_invalid"))
		return exitFailure
	}
	fmt.Fprintln(a.stdout, msg.text("doctor_ok_config", result.RootFile))
	valid := true
	for _, tool := range result.RequiredTools {
		if _, err := exec.LookPath(tool); err != nil {
			fmt.Fprintln(a.stdout, msg.text("doctor_fail_tool", tool))
			valid = false
		} else {
			fmt.Fprintln(a.stdout, msg.text("doctor_ok_tool", tool))
		}
	}
	if !valid {
		fmt.Fprintln(a.stdout, msg.text("doctor_invalid"))
		return exitFailure
	}
	fmt.Fprintln(a.stdout, msg.text("doctor_valid"))
	return exitSuccess
}
