package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Vinicius0812/alfred-cli/internal/config"
)

var Version = "0.2.0-dev"

const (
	exitSuccess  = 0
	exitFailure  = 1
	exitUsage    = 2
	exitCanceled = 130
)

type application struct {
	stdin     io.Reader
	reader    *bufio.Reader
	stdout    io.Writer
	stderr    io.Writer
	lookupEnv func(string) (string, bool)
	environ   func() []string
	getwd     func() (string, error)
	now       func() time.Time
}

func Run(args []string, stdout, stderr io.Writer) int {
	return RunWithIO(args, os.Stdin, stdout, stderr)
}

func RunWithIO(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	app := application{
		stdin:     stdin,
		reader:    bufio.NewReader(stdin),
		stdout:    stdout,
		stderr:    stderr,
		lookupEnv: os.LookupEnv,
		environ:   os.Environ,
		getwd:     os.Getwd,
		now:       time.Now,
	}
	return app.run(args)
}

func (a *application) run(args []string) int {
	global, command, err := a.parseArgs(args)
	msg := newMessages(global.lang)
	if err != nil {
		fmt.Fprintf(a.stderr, "alfred: %s\n", localizeDiagnostic(err.Error(), global.lang))
		return exitUsage
	}

	if len(command) == 0 {
		printHelp(a.stdout, msg)
		return exitSuccess
	}

	switch command[0] {
	case "-h", "--help", "help":
		printHelp(a.stdout, msg)
		return exitSuccess
	case "-v", "--version", "version":
		fmt.Fprintf(a.stdout, "alfred %s\n", Version)
		return exitSuccess
	case "menu":
		return a.runMenu(global, msg)
	case "init":
		return a.runInit(command[1:], global, msg)
	case "doctor":
		return a.runDoctor(command[1:], global, msg)
	case "config":
		return a.runConfig(command[1:], global, msg)
	case "workflow":
		return a.runWorkflow(command[1:], global, msg)
	case "run":
		return a.runWorkflowExecution(command[1:], global, msg)
	case "completion":
		return a.runCompletion(command[1:], msg)
	default:
		fmt.Fprintf(a.stderr, msg.text("unknown_command")+"\n\n", command[0])
		printHelp(a.stderr, msg)
		return exitUsage
	}
}

type globalOptions struct {
	configPath           string
	lang                 language
	allowExternalExtends bool
	noColor              bool
}

func (a *application) parseArgs(args []string) (globalOptions, []string, error) {
	envLanguage, _ := a.lookupEnv("ALFRED_LANG")
	lang, err := resolveLanguage("", envLanguage)
	if err != nil {
		return globalOptions{}, nil, err
	}
	opts := globalOptions{lang: lang}
	command := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			command = append(command, args[i+1:]...)
			break
		}
		switch {
		case arg == "--config":
			if i+1 >= len(args) {
				return opts, nil, fmt.Errorf("--config requires a path")
			}
			i++
			opts.configPath = args[i]
		case strings.HasPrefix(arg, "--config="):
			opts.configPath = strings.TrimPrefix(arg, "--config=")
			if opts.configPath == "" {
				return opts, nil, fmt.Errorf("--config requires a path")
			}
		case arg == "--lang" || arg == "--language":
			if i+1 >= len(args) {
				return opts, nil, fmt.Errorf("--lang requires a value")
			}
			i++
			opts.lang, err = resolveLanguage(args[i], "")
			if err != nil {
				return opts, nil, err
			}
		case strings.HasPrefix(arg, "--lang=") || strings.HasPrefix(arg, "--language="):
			value := strings.SplitN(arg, "=", 2)[1]
			opts.lang, err = resolveLanguage(value, "")
			if err != nil {
				return opts, nil, err
			}
		case arg == "--allow-external-extends":
			opts.allowExternalExtends = true
		case arg == "--no-color":
			opts.noColor = true
		default:
			command = append(command, arg)
		}
	}

	return opts, command, nil
}

func (a *application) loadConfig(global globalOptions, msg messages) (config.Result, bool) {
	cwd, err := a.getwd()
	if err != nil {
		fmt.Fprintf(a.stderr, msg.text("cwd_error")+"\n", err)
		return config.Result{}, false
	}
	result, err := config.LoadAndValidate(config.Options{
		ConfigPath:           global.configPath,
		WorkingDir:           cwd,
		LookupEnv:            a.lookupEnv,
		AllowExternalExtends: global.allowExternalExtends,
	})
	if err != nil {
		fmt.Fprintf(a.stderr, "alfred: %s\n", localizeDiagnostic(err.Error(), global.lang))
		if strings.Contains(err.Error(), "not found") {
			fmt.Fprintln(a.stderr, msg.text("config_missing_hint"))
		}
		return config.Result{}, false
	}
	if len(result.Errors) > 0 {
		for _, validationErr := range result.Errors {
			fmt.Fprintln(a.stderr, validationErr.Localized(string(global.lang)))
		}
		return result, false
	}
	return result, true
}

func printHelp(w io.Writer, msg messages) {
	fmt.Fprintln(w, msg.text("help"))
}
