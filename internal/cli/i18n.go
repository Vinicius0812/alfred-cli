package cli

import (
	"fmt"
	"os"
	"strings"
)

type language string

const (
	languageEnglish    language = "en"
	languagePortuguese language = "pt-BR"
)

func resolveLanguage(value string) language {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		normalized = strings.ToLower(strings.TrimSpace(os.Getenv("ALFRED_LANG")))
	}

	switch normalized {
	case "pt", "pt-br", "pt_br", "portuguese", "portugues", "português":
		return languagePortuguese
	default:
		return languageEnglish
	}
}

type messages struct {
	lang language
}

func newMessages(lang language) messages {
	return messages{lang: lang}
}

func (m messages) text(key string, args ...any) string {
	catalog := englishMessages
	if m.lang == languagePortuguese {
		catalog = portugueseMessages
	}

	template, ok := catalog[key]
	if !ok {
		template = englishMessages[key]
	}
	if len(args) == 0 {
		return template
	}
	return fmt.Sprintf(template, args...)
}

var englishMessages = map[string]string{
	"help": `Alfred CLI

Usage:
  alfred [--lang en|pt-BR] [--config path] <command>

Commands:
  menu              Open the interactive menu
  init              Create an initial alfred.yaml
  doctor            Inspect the local project environment
  config validate   Validate alfred.yaml without executing workflows
  help              Show this help
  version           Show the Alfred version

Language:
  Use --lang en, --lang pt-BR, or ALFRED_LANG.`,
	"config_usage":          "Usage: alfred [--config path] config validate",
	"config_valid":          "configuration valid: %s",
	"config_valid_generic":  "configuration valid",
	"source":                "source: %s",
	"unknown_command":       "alfred: unknown command %q",
	"unknown_config":        "alfred: unknown config command %q",
	"config_extra_arg":      "alfred: config validate does not accept positional arguments",
	"cwd_error":             "alfred: could not determine current directory: %v",
	"init_created":          "created %s",
	"init_exists":           "alfred: %s already exists; use --force to overwrite",
	"init_usage":            "Usage: alfred init [--project name] [--force]",
	"doctor_title":          "Alfred Doctor",
	"doctor_ok_config":      "OK config: %s",
	"doctor_fail_config":    "FAIL config: %s",
	"doctor_ok_tool":        "OK tool found: %s",
	"doctor_fail_tool":      "FAIL tool missing: %s",
	"doctor_valid":          "environment looks ready",
	"doctor_invalid":        "environment has issues",
	"flag_requires_value":   "%s requires a value",
	"unsupported_language":  "unsupported language %q; use en or pt-BR",
	"unknown_init_argument": "alfred: unknown init argument %q",
	"menu_title":            "What do you want to do?",
	"menu_validate":         "1) Validate configuration",
	"menu_doctor":           "2) Run doctor",
	"menu_init":             "3) Initialize project",
	"menu_version":          "4) Show version",
	"menu_exit":             "0) Exit",
	"menu_prompt":           "Choose an option: ",
	"menu_invalid":          "invalid menu option",
}

var portugueseMessages = map[string]string{
	"help": `Alfred CLI

Uso:
  alfred [--lang en|pt-BR] [--config caminho] <comando>

Comandos:
  menu              Abre o menu interativo
  init              Cria um alfred.yaml inicial
  doctor            Inspeciona o ambiente local do projeto
  config validate   Valida o alfred.yaml sem executar workflows
  help              Mostra esta ajuda
  version           Mostra a versão do Alfred

Idioma:
  Use --lang en, --lang pt-BR ou ALFRED_LANG.`,
	"config_usage":          "Uso: alfred [--config caminho] config validate",
	"config_valid":          "configuração válida: %s",
	"config_valid_generic":  "configuração válida",
	"source":                "origem: %s",
	"unknown_command":       "alfred: comando desconhecido %q",
	"unknown_config":        "alfred: subcomando de config desconhecido %q",
	"config_extra_arg":      "alfred: config validate não aceita argumentos posicionais",
	"cwd_error":             "alfred: não foi possível determinar o diretório atual: %v",
	"init_created":          "%s criado",
	"init_exists":           "alfred: %s já existe; use --force para sobrescrever",
	"init_usage":            "Uso: alfred init [--project nome] [--force]",
	"doctor_title":          "Diagnóstico do Alfred",
	"doctor_ok_config":      "OK configuração: %s",
	"doctor_fail_config":    "FALHA configuração: %s",
	"doctor_ok_tool":        "OK ferramenta encontrada: %s",
	"doctor_fail_tool":      "FALHA ferramenta ausente: %s",
	"doctor_valid":          "ambiente parece pronto",
	"doctor_invalid":        "ambiente tem problemas",
	"flag_requires_value":   "%s exige um valor",
	"unsupported_language":  "idioma não suportado %q; use en ou pt-BR",
	"unknown_init_argument": "alfred: argumento de init desconhecido %q",
	"menu_title":            "O que você quer fazer?",
	"menu_validate":         "1) Validar configuração",
	"menu_doctor":           "2) Executar diagnóstico",
	"menu_init":             "3) Inicializar projeto",
	"menu_version":          "4) Mostrar versão",
	"menu_exit":             "0) Sair",
	"menu_prompt":           "Escolha uma opção: ",
	"menu_invalid":          "opção de menu inválida",
}
