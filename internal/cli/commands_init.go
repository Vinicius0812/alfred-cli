package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Vinicius0812/alfred-cli/internal/preset"
	"gopkg.in/yaml.v3"
)

func (a *application) runInit(args []string, global globalOptions, msg messages) int {
	project := ""
	projectProvided := false
	presetName := preset.Basic
	force := false

	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "-h", "--help":
			fmt.Fprintln(a.stdout, msg.text("init_usage"))
			return exitSuccess
		case "--force":
			force = true
		case "--project", "--preset":
			flag := args[index]
			if index+1 >= len(args) {
				fmt.Fprintln(a.stderr, msg.text("flag_requires_value", flag))
				return exitUsage
			}
			index++
			if flag == "--project" {
				project = strings.TrimSpace(args[index])
				projectProvided = true
			} else {
				presetName = strings.TrimSpace(args[index])
			}
		default:
			fmt.Fprintln(a.stderr, msg.text("unknown_init_argument", args[index]))
			fmt.Fprintln(a.stderr, msg.text("init_usage"))
			return exitUsage
		}
	}

	cwd, err := a.getwd()
	if err != nil {
		fmt.Fprintln(a.stderr, msg.text("cwd_error", err))
		return exitFailure
	}
	if !projectProvided {
		project = filepath.Base(cwd)
	}
	if project == "" || project == "." || project == string(filepath.Separator) {
		fmt.Fprintln(a.stderr, "alfred: "+localizeDiagnostic("--project must not be empty", global.lang))
		return exitUsage
	}
	configuration, err := preset.Build(presetName, project)
	if err != nil {
		fmt.Fprintf(a.stderr, "alfred: %s\n", localizeDiagnostic(err.Error(), global.lang))
		return exitUsage
	}
	content, err := yaml.Marshal(configuration)
	if err != nil {
		fmt.Fprintf(a.stderr, "alfred: %s\n", localizeDiagnostic("could not generate configuration: "+err.Error(), global.lang))
		return exitFailure
	}

	target := global.configPath
	if target == "" {
		target = filepath.Join(cwd, "alfred.yaml")
	} else if !filepath.IsAbs(target) {
		target = filepath.Join(cwd, target)
	}
	target = filepath.Clean(target)
	if _, err := os.Lstat(target); err == nil && !force {
		fmt.Fprintln(a.stderr, msg.text("init_exists", target))
		return exitFailure
	} else if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(a.stderr, "alfred: %s\n", localizeDiagnostic(fmt.Sprintf("could not inspect %s: %v", target, err), global.lang))
		return exitFailure
	}
	if err := writeFileAtomically(target, content, 0o600); err != nil {
		fmt.Fprintf(a.stderr, "alfred: %s\n", localizeDiagnostic(fmt.Sprintf("could not create %s: %v", target, err), global.lang))
		return exitFailure
	}
	fmt.Fprintln(a.stdout, msg.text("init_created", target, presetName))
	return exitSuccess
}

func writeFileAtomically(target string, content []byte, mode os.FileMode) error {
	directory := filepath.Dir(target)
	temporary, err := os.CreateTemp(directory, ".alfred-config-*.tmp")
	if err != nil {
		return err
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if err := temporary.Chmod(mode); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(content); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryName, target); err != nil {
		if removeErr := os.Remove(target); removeErr != nil && !os.IsNotExist(removeErr) {
			return err
		}
		return os.Rename(temporaryName, target)
	}
	return nil
}
