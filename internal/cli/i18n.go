package cli

import (
	"fmt"
	"strings"
)

type language string

const (
	languageEnglish    language = "en"
	languagePortuguese language = "pt-BR"
)

func resolveLanguage(value, environmentValue string) (language, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		normalized = strings.ToLower(strings.TrimSpace(environmentValue))
	}

	switch normalized {
	case "", "en", "en-us", "en_us", "english":
		return languageEnglish, nil
	case "pt", "pt-br", "pt_br", "portuguese", "portugues", "português":
		return languagePortuguese, nil
	default:
		return "", fmt.Errorf("unsupported language %q; use en or pt-BR", normalized)
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

func localizeDiagnostic(value string, lang language) string {
	if lang != languagePortuguese {
		return value
	}
	replacements := []struct {
		from string
		to   string
	}{
		{"configuration file not found", "arquivo de configuração não encontrado"},
		{"alfred.yaml not found in current directory or parents", "alfred.yaml não encontrado no diretório atual ou nos diretórios superiores"},
		{"could not read configuration file", "não foi possível ler o arquivo de configuração"},
		{"could not resolve configuration file", "não foi possível resolver o arquivo de configuração"},
		{"configuration is invalid", "a configuração é inválida"},
		{"requires a value", "exige um valor"},
		{"requires a path", "exige um caminho"},
		{"must not be empty", "não pode ficar vazio"},
		{"output must be text or json", "a saída deve ser text ou json"},
		{"timeout must be a positive Go duration such as 30s or 5m", "o timeout deve ser uma duração Go positiva, como 30s ou 5m"},
		{"could not build workflow plan", "não foi possível construir o plano do workflow"},
		{"could not hash configuration", "não foi possível calcular o hash da configuração"},
		{"invalid workflow directory", "diretório do workflow inválido"},
		{"invalid audit directory", "diretório de auditoria inválido"},
		{"invalid run id", "ID de execução inválido"},
		{"audit record not found", "registro de auditoria não encontrado"},
		{"could not list audit records", "não foi possível listar os registros de auditoria"},
		{"could not read audit record", "não foi possível ler o registro de auditoria"},
		{"invalid audit record", "registro de auditoria inválido"},
		{"could not create audit directory", "não foi possível criar o diretório de auditoria"},
		{"could not encode audit record", "não foi possível codificar o registro de auditoria"},
		{"could not create audit record", "não foi possível criar o registro de auditoria"},
		{"could not write audit record", "não foi possível gravar o registro de auditoria"},
		{"could not sync audit record", "não foi possível sincronizar o registro de auditoria"},
		{"could not close audit record", "não foi possível fechar o registro de auditoria"},
		{"could not publish audit record", "não foi possível publicar o registro de auditoria"},
		{"audit failed", "a auditoria falhou"},
		{"path escapes the project root", "o caminho sai da raiz do projeto"},
		{"could not resolve project root", "não foi possível resolver a raiz do projeto"},
		{"environment variable", "variável de ambiente"},
		{"is not defined", "não está definida"},
		{"shell steps are disabled by policy", "etapas de shell estão desabilitadas pela política"},
		{"step requires interactive confirmation", "a etapa exige confirmação interativa"},
		{"operation declined", "operação recusada"},
		{"step canceled", "etapa cancelada"},
		{"step timed out after", "a etapa excedeu o limite de"},
		{"command exited with code", "o comando terminou com o código"},
		{"could not inspect Git worktree", "não foi possível inspecionar o diretório de trabalho Git"},
		{"Git worktree is not clean", "o diretório de trabalho Git possui alterações"},
		{"unknown preset", "preset desconhecido"},
		{"available presets", "presets disponíveis"},
		{"could not generate configuration", "não foi possível gerar a configuração"},
		{"could not encode configuration", "não foi possível codificar a configuração"},
		{"could not inspect", "não foi possível inspecionar"},
		{"could not create", "não foi possível criar"},
		{"workflow not found", "workflow não encontrado"},
	}
	for _, replacement := range replacements {
		value = strings.ReplaceAll(value, replacement.from, replacement.to)
	}
	return value
}

func localizeConfirmationReason(value string, lang language) string {
	if lang != languagePortuguese {
		return value
	}
	switch value {
	case "shell command":
		return "comando de shell"
	case "step policy":
		return "política da etapa"
	case "network access":
		return "acesso à rede"
	case "Git state change":
		return "alteração do estado do Git"
	case "destructive operation":
		return "operação destrutiva"
	default:
		return value
	}
}

func localizeOutcome(value string, lang language) string {
	if lang != languagePortuguese {
		return value
	}
	switch value {
	case "planned":
		return "planejada"
	case "success":
		return "sucesso"
	case "success_with_warnings":
		return "sucesso com avisos"
	case "failed":
		return "falha"
	case "blocked":
		return "bloqueada"
	case "declined":
		return "recusada"
	default:
		return value
	}
}

var englishMessages = map[string]string{
	"help": `Alfred CLI - secure, auditable development automation

Usage:
  alfred [global options] <command>

Commands:
  menu                         Open the interactive menu
  init [--preset name]         Create an alfred.yaml safely
  doctor                       Inspect the project environment
  config validate              Validate the resolved configuration
  config explain [--output]    Show the resolved configuration
  workflow list [--output]     List configured workflows
  run <workflow> [options]     Plan or execute a workflow
  run list [--output]          List local audit records
  run show <id> [--output]     Show one audit record
  completion <shell>           Generate bash, zsh, or PowerShell completion
  version                      Show the Alfred version

Global options:
  --config <path>              Select a configuration file
  --lang <en|pt-BR>            Select the interface language
  --allow-external-extends     Trust presets outside the project root
  --no-color                   Disable terminal colors

Environment:
  ALFRED_LANG                  Default interface language`,
	"config_usage":              "Usage: alfred [--config path] config <validate|explain> [--output text|json]",
	"config_valid":              "Everything is in order, Master Bruce. Configuration valid: %s",
	"config_valid_generic":      "Everything is in order, Master Bruce. Configuration valid.",
	"config_missing_hint":       "Hint: run 'alfred init' to create one in this directory.",
	"source":                    "Source: %s",
	"loaded_files":              "Loaded files:",
	"resolved_config":           "Resolved configuration:",
	"unknown_command":           "alfred: unknown command %q",
	"unknown_config":            "alfred: unknown config command %q",
	"config_extra_arg":          "alfred: config validate does not accept positional arguments",
	"cwd_error":                 "alfred: could not determine the current directory: %v",
	"init_created":              "Your configuration is ready, Master Bruce: %s (preset %s).",
	"init_exists":               "alfred: %s already exists; use --force to overwrite",
	"init_usage":                "Usage: alfred init [--project name] [--preset basic|go-secure] [--force]",
	"doctor_title":              "Alfred Doctor — allow me to inspect the estate, Master Bruce.",
	"doctor_usage":              "Usage: alfred doctor",
	"doctor_ok_config":          "OK configuration: %s",
	"doctor_fail_config":        "FAIL configuration: %s",
	"doctor_ok_tool":            "OK tool found: %s",
	"doctor_fail_tool":          "FAIL tool missing: %s",
	"doctor_valid":              "The environment is ready, Master Bruce.",
	"doctor_invalid":            "I found matters requiring your attention, Master Bruce.",
	"flag_requires_value":       "%s requires a value",
	"unsupported_language":      "unsupported language %q; use en or pt-BR",
	"unknown_argument":          "alfred: unknown argument %q",
	"unknown_init_argument":     "alfred: unknown init argument %q",
	"unknown_workflow_argument": "alfred: unknown workflow argument %q",
	"output_invalid":            "alfred: output must be text or json",
	"workflow_usage":            "Usage: alfred workflow list [--output text|json]",
	"workflow_none":             "No workflows are configured.",
	"workflow_title":            "These routines are at your disposal, Master Bruce:",
	"workflow_item":             "  %s - %s (%d steps)",
	"workflow_item_one":         "  %s - %s (1 step)",
	"run_usage":                 "Usage: alfred run <workflow> [--dry-run] [--yes] [--timeout duration] [--output text|json]\n       alfred run list|show <id> [--output text|json]",
	"run_plan":                  "I have prepared the %s plan, Master Bruce (%d steps):",
	"run_plan_one":              "I have prepared the %s plan, Master Bruce (1 step):",
	"run_step":                  "  %d. %s [%s] %s (timeout %s)",
	"run_shell":                 "shell command",
	"run_confirm":               "Step %d (%s) requires confirmation: %s. Shall I proceed, Master Bruce? [y/N]: ",
	"run_finished":              "The task is complete, Master Bruce. Run %s finished with outcome: %s",
	"run_audit":                 "Audit record: .alfred/runs/%s.json",
	"run_error":                 "alfred: workflow failed: %v",
	"run_not_found":             "alfred: workflow %q was not found",
	"run_list_title":            "Local audit records:",
	"run_list_empty":            "No local audit records found.",
	"run_list_item":             "  %s  %-22s  %s",
	"run_show_title":            "Audit record %s",
	"completion_usage":          "Usage: alfred completion <bash|zsh|powershell>",
	"completion_unknown":        "alfred: unsupported shell %q; use bash, zsh, or powershell",
	"menu_title":                "How may I assist you, Master Bruce?",
	"menu_validate":             "1) Validate configuration",
	"menu_doctor":               "2) Run doctor",
	"menu_init":                 "3) Initialize project",
	"menu_workflows":            "4) List workflows",
	"menu_run":                  "5) Run a workflow",
	"menu_version":              "6) Show version",
	"menu_exit":                 "0) Exit",
	"menu_prompt":               "Choose an option: ",
	"menu_workflow_prompt":      "Workflow name: ",
	"menu_invalid":              "Invalid menu option.",
	"menu_goodbye":              "Until next time, Master Bruce.",
}

var portugueseMessages = map[string]string{
	"help": `Alfred CLI - automação de desenvolvimento segura e auditável

Uso:
  alfred [opções globais] <comando>

Comandos:
  menu                         Abre o menu interativo
  init [--preset nome]         Cria um alfred.yaml com segurança
  doctor                       Inspeciona o ambiente do projeto
  config validate              Valida a configuração resolvida
  config explain [--output]    Mostra a configuração resolvida
  workflow list [--output]     Lista os workflows configurados
  run <workflow> [opções]      Planeja ou executa um workflow
  run list [--output]          Lista registros locais de auditoria
  run show <id> [--output]     Mostra um registro de auditoria
  completion <shell>           Gera completion para bash, zsh ou PowerShell
  version                      Mostra a versão do Alfred

Opções globais:
  --config <caminho>           Seleciona um arquivo de configuração
  --lang <en|pt-BR>            Seleciona o idioma da interface
  --allow-external-extends     Confia em presets externos ao projeto
  --no-color                   Desabilita cores no terminal

Ambiente:
  ALFRED_LANG                  Idioma padrão da interface`,
	"config_usage":              "Uso: alfred [--config caminho] config <validate|explain> [--output text|json]",
	"config_valid":              "Tudo em ordem, mestre Bruce. Configuração válida: %s",
	"config_valid_generic":      "Tudo em ordem, mestre Bruce. Configuração válida.",
	"config_missing_hint":       "Dica: execute 'alfred init' para criar uma configuração neste diretório.",
	"source":                    "Origem: %s",
	"loaded_files":              "Arquivos carregados:",
	"resolved_config":           "Configuração resolvida:",
	"unknown_command":           "alfred: comando desconhecido %q",
	"unknown_config":            "alfred: subcomando de config desconhecido %q",
	"config_extra_arg":          "alfred: config validate não aceita argumentos posicionais",
	"cwd_error":                 "alfred: não foi possível determinar o diretório atual: %v",
	"init_created":              "Sua configuração está pronta, mestre Bruce: %s (preset %s).",
	"init_exists":               "alfred: %s já existe; use --force para sobrescrever",
	"init_usage":                "Uso: alfred init [--project nome] [--preset basic|go-secure] [--force]",
	"doctor_title":              "Diagnóstico do Alfred — permita-me inspecionar o ambiente, mestre Bruce.",
	"doctor_usage":              "Uso: alfred doctor",
	"doctor_ok_config":          "OK configuração: %s",
	"doctor_fail_config":        "FALHA configuração: %s",
	"doctor_ok_tool":            "OK ferramenta encontrada: %s",
	"doctor_fail_tool":          "FALHA ferramenta ausente: %s",
	"doctor_valid":              "O ambiente está pronto, mestre Bruce.",
	"doctor_invalid":            "Encontrei pontos que exigem sua atenção, mestre Bruce.",
	"flag_requires_value":       "%s exige um valor",
	"unsupported_language":      "idioma não suportado %q; use en ou pt-BR",
	"unknown_argument":          "alfred: argumento desconhecido %q",
	"unknown_init_argument":     "alfred: argumento de init desconhecido %q",
	"unknown_workflow_argument": "alfred: argumento de workflow desconhecido %q",
	"output_invalid":            "alfred: a saída deve ser text ou json",
	"workflow_usage":            "Uso: alfred workflow list [--output text|json]",
	"workflow_none":             "Nenhum workflow está configurado.",
	"workflow_title":            "Estas rotinas estão à sua disposição, mestre Bruce:",
	"workflow_item":             "  %s - %s (%d etapas)",
	"workflow_item_one":         "  %s - %s (1 etapa)",
	"run_usage":                 "Uso: alfred run <workflow> [--dry-run] [--yes] [--timeout duração] [--output text|json]\n     alfred run list|show <id> [--output text|json]",
	"run_plan":                  "Preparei o plano %s, mestre Bruce (%d etapas):",
	"run_plan_one":              "Preparei o plano %s, mestre Bruce (1 etapa):",
	"run_step":                  "  %d. %s [%s] %s (limite %s)",
	"run_shell":                 "comando de shell",
	"run_confirm":               "A etapa %d (%s) exige confirmação: %s. Devo prosseguir, mestre Bruce? [s/N]: ",
	"run_finished":              "A tarefa está concluída, mestre Bruce. Execução %s finalizada com resultado: %s",
	"run_audit":                 "Registro de auditoria: .alfred/runs/%s.json",
	"run_error":                 "alfred: falha no workflow: %v",
	"run_not_found":             "alfred: o workflow %q não foi encontrado",
	"run_list_title":            "Registros locais de auditoria:",
	"run_list_empty":            "Nenhum registro local de auditoria foi encontrado.",
	"run_list_item":             "  %s  %-22s  %s",
	"run_show_title":            "Registro de auditoria %s",
	"completion_usage":          "Uso: alfred completion <bash|zsh|powershell>",
	"completion_unknown":        "alfred: shell %q não suportado; use bash, zsh ou powershell",
	"menu_title":                "Como posso ajudá-lo, mestre Bruce?",
	"menu_validate":             "1) Validar configuração",
	"menu_doctor":               "2) Executar diagnóstico",
	"menu_init":                 "3) Inicializar projeto",
	"menu_workflows":            "4) Listar workflows",
	"menu_run":                  "5) Executar um workflow",
	"menu_version":              "6) Mostrar versão",
	"menu_exit":                 "0) Sair",
	"menu_prompt":               "Escolha uma opção: ",
	"menu_workflow_prompt":      "Nome do workflow: ",
	"menu_invalid":              "Opção de menu inválida.",
	"menu_goodbye":              "Até a próxima, mestre Bruce.",
}
