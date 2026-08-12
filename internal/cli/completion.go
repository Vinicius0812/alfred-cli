package cli

import "fmt"

func (a *application) runCompletion(args []string, msg messages) int {
	if len(args) != 1 || args[0] == "-h" || args[0] == "--help" {
		fmt.Fprintln(a.stdout, msg.text("completion_usage"))
		if len(args) == 1 {
			return exitSuccess
		}
		return exitUsage
	}
	switch args[0] {
	case "bash":
		fmt.Fprint(a.stdout, bashCompletion)
	case "zsh":
		fmt.Fprint(a.stdout, zshCompletion)
	case "powershell", "pwsh":
		fmt.Fprint(a.stdout, powershellCompletion)
	default:
		fmt.Fprintln(a.stderr, msg.text("completion_unknown", args[0]))
		return exitUsage
	}
	return exitSuccess
}

const bashCompletion = `# bash completion for Alfred CLI
_alfred_complete() {
  local current previous
  current="${COMP_WORDS[COMP_CWORD]}"
  previous="${COMP_WORDS[COMP_CWORD-1]}"
  case "$previous" in
    --lang) COMPREPLY=( $(compgen -W "en pt-BR" -- "$current") ); return ;;
    --output) COMPREPLY=( $(compgen -W "text json" -- "$current") ); return ;;
    --preset) COMPREPLY=( $(compgen -W "basic go-secure" -- "$current") ); return ;;
    completion) COMPREPLY=( $(compgen -W "bash zsh powershell" -- "$current") ); return ;;
  esac
  COMPREPLY=( $(compgen -W "menu init doctor config workflow run completion version help --config --lang --allow-external-extends --no-color" -- "$current") )
}
complete -F _alfred_complete alfred
`

const zshCompletion = `#compdef alfred
_arguments \
  '--config[configuration file]:file:_files' \
  '--lang[interface language]:(en pt-BR)' \
  '--allow-external-extends[trust external presets]' \
  '--no-color[disable colors]' \
  '1:command:(menu init doctor config workflow run completion version help)' \
  '*::argument:_files'
`

const powershellCompletion = `Register-ArgumentCompleter -Native -CommandName alfred -ScriptBlock {
  param($wordToComplete, $commandAst, $cursorPosition)
  'menu','init','doctor','config','workflow','run','completion','version','help',
  '--config','--lang','--allow-external-extends','--no-color' |
    Where-Object { $_ -like "$wordToComplete*" } |
    ForEach-Object { [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_) }
}
`
