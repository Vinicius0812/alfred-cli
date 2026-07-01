package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Vinicius0812/alfred-cli/internal/config"
)

var Version = "0.1.0-dev"

func Run(args []string, stdout, stderr io.Writer) int {
	global, command, err := parseArgs(args)
	msg := newMessages(global.lang)
	if err != nil {
		fmt.Fprintf(stderr, "alfred: %v\n", err)
		return 2
	}

	if len(command) == 0 {
		printHelp(stdout, msg)
		return 0
	}

	switch command[0] {
	case "-h", "--help", "help":
		printHelp(stdout, msg)
		return 0
	case "-v", "--version", "version":
		fmt.Fprintf(stdout, "alfred %s\n", Version)
		return 0
	case "menu":
		return runMenu(global, msg, stdout, stderr)
	case "init":
		return runInit(command[1:], global, msg, stdout, stderr)
	case "doctor":
		return runDoctor(global, msg, stdout, stderr)
	case "config":
		return runConfig(command[1:], global, msg, stdout, stderr)
	default:
		fmt.Fprintf(stderr, msg.text("unknown_command")+"\n\n", command[0])
		printHelp(stderr, msg)
		return 2
	}
}

type globalOptions struct {
	configPath string
	lang       language
}

func parseArgs(args []string) (globalOptions, []string, error) {
	opts := globalOptions{lang: resolveLanguage("")}
	command := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--config":
			if i+1 >= len(args) {
				return opts, nil, fmt.Errorf("--config requires a path")
			}
			i++
			opts.configPath = args[i]
		case "--lang", "--language":
			if i+1 >= len(args) {
				return opts, nil, fmt.Errorf("--lang requires a value")
			}
			i++
			opts.lang = resolveLanguage(args[i])
		default:
			command = append(command, arg)
		}
	}

	return opts, command, nil
}

func runConfig(args []string, global globalOptions, msg messages, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		fmt.Fprintln(stdout, msg.text("config_usage"))
		return 0
	}

	if args[0] != "validate" {
		fmt.Fprintf(stderr, msg.text("unknown_config")+"\n", args[0])
		return 2
	}

	if len(args) > 1 {
		fmt.Fprintln(stderr, msg.text("config_extra_arg"))
		return 2
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, msg.text("cwd_error")+"\n", err)
		return 1
	}

	result, err := config.LoadAndValidate(config.Options{
		ConfigPath: global.configPath,
		WorkingDir: cwd,
		Getenv:     os.Getenv,
	})
	if err != nil {
		fmt.Fprintf(stderr, "alfred: %v\n", err)
		return 1
	}

	if len(result.Errors) > 0 {
		for _, validationErr := range result.Errors {
			fmt.Fprintln(stderr, validationErr.Error())
		}
		return 1
	}

	if result.ProjectName != "" {
		fmt.Fprintln(stdout, msg.text("config_valid", result.ProjectName))
	} else {
		fmt.Fprintln(stdout, msg.text("config_valid_generic"))
	}
	fmt.Fprintln(stdout, msg.text("source", result.RootFile))
	return 0
}

func runInit(args []string, global globalOptions, msg messages, stdout, stderr io.Writer) int {
	opts := initOptions{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 >= len(args) {
				fmt.Fprintf(stderr, "alfred: %s\n", msg.text("flag_requires_value", "--project"))
				return 2
			}
			i++
			opts.projectName = args[i]
		case "--force":
			opts.force = true
		case "-h", "--help", "help":
			fmt.Fprintln(stdout, msg.text("init_usage"))
			return 0
		default:
			fmt.Fprintf(stderr, msg.text("unknown_init_argument")+"\n", args[i])
			return 2
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, msg.text("cwd_error")+"\n", err)
		return 1
	}
	if opts.projectName == "" {
		opts.projectName = filepath.Base(cwd)
	}

	target := global.configPath
	if target == "" {
		target = filepath.Join(cwd, "alfred.yaml")
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(cwd, target)
	}

	if _, err := os.Stat(target); err == nil && !opts.force {
		fmt.Fprintf(stderr, msg.text("init_exists")+"\n", target)
		return 1
	} else if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(stderr, "alfred: %v\n", err)
		return 1
	}

	content := initialConfig(opts.projectName)
	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
		fmt.Fprintf(stderr, "alfred: %v\n", err)
		return 1
	}

	fmt.Fprintln(stdout, msg.text("init_created", target))
	return 0
}

type initOptions struct {
	projectName string
	force       bool
}

func initialConfig(projectName string) string {
	return fmt.Sprintf(`version: 1

project:
  name: %s

workflows:
  test:
    description: Run tests
    steps:
      - run: go test ./...

policies:
  confirm_destructive_actions: true
`, projectName)
}

func runDoctor(global globalOptions, msg messages, stdout, stderr io.Writer) int {
	fmt.Fprintln(stdout, msg.text("doctor_title"))

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, msg.text("cwd_error")+"\n", err)
		return 1
	}

	result, err := config.LoadAndValidate(config.Options{
		ConfigPath: global.configPath,
		WorkingDir: cwd,
		Getenv:     os.Getenv,
	})
	if err != nil {
		fmt.Fprintln(stdout, msg.text("doctor_fail_config", err.Error()))
		fmt.Fprintln(stdout, msg.text("doctor_invalid"))
		return 1
	}
	if len(result.Errors) > 0 {
		fmt.Fprintln(stdout, msg.text("doctor_fail_config", result.Errors.Error()))
		fmt.Fprintln(stdout, msg.text("doctor_invalid"))
		return 1
	}

	fmt.Fprintln(stdout, msg.text("doctor_ok_config", result.RootFile))

	ok := true
	tools := append([]string{"git"}, result.RequiredTools...)
	for _, tool := range uniqueStrings(tools) {
		if _, err := exec.LookPath(tool); err != nil {
			fmt.Fprintln(stdout, msg.text("doctor_fail_tool", tool))
			ok = false
			continue
		}
		fmt.Fprintln(stdout, msg.text("doctor_ok_tool", tool))
	}

	if !ok {
		fmt.Fprintln(stdout, msg.text("doctor_invalid"))
		return 1
	}
	fmt.Fprintln(stdout, msg.text("doctor_valid"))
	return 0
}

func runMenu(global globalOptions, msg messages, stdout, stderr io.Writer) int {
	fmt.Fprintln(stdout, msg.text("menu_title"))
	fmt.Fprintln(stdout, msg.text("menu_validate"))
	fmt.Fprintln(stdout, msg.text("menu_doctor"))
	fmt.Fprintln(stdout, msg.text("menu_init"))
	fmt.Fprintln(stdout, msg.text("menu_version"))
	fmt.Fprintln(stdout, msg.text("menu_exit"))
	fmt.Fprint(stdout, msg.text("menu_prompt"))

	var choice string
	if _, err := fmt.Fscan(os.Stdin, &choice); err != nil {
		fmt.Fprintf(stderr, "alfred: %v\n", err)
		return 1
	}

	switch choice {
	case "1":
		return runConfig([]string{"validate"}, global, msg, stdout, stderr)
	case "2":
		return runDoctor(global, msg, stdout, stderr)
	case "3":
		return runInit(nil, global, msg, stdout, stderr)
	case "4":
		fmt.Fprintf(stdout, "alfred %s\n", Version)
		return 0
	case "0":
		return 0
	default:
		fmt.Fprintln(stderr, msg.text("menu_invalid"))
		return 2
	}
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	unique := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		unique = append(unique, value)
	}
	return unique
}

func printHelp(w io.Writer, msg messages) {
	fmt.Fprintln(w, msg.text("help"))
}
