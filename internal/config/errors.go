package config

import (
	"fmt"
	"strings"
)

type ValidationError struct {
	File    string
	Path    string
	Message string
}

func (e ValidationError) Error() string {
	if e.Path == "" {
		return fmt.Sprintf("%s: %s", e.File, e.Message)
	}
	return fmt.Sprintf("%s: %s: %s", e.File, e.Path, e.Message)
}

// Localized keeps the stable configuration path while translating common
// validation guidance for the CLI. English remains the canonical diagnostic.
func (e ValidationError) Localized(language string) string {
	if language != "pt" && language != "pt-BR" && language != "pt_br" {
		return e.Error()
	}

	message := portugueseValidationMessage(e.Message)
	if e.Path == "" {
		return fmt.Sprintf("%s: %s", e.File, message)
	}
	return fmt.Sprintf("%s: %s: %s", e.File, e.Path, message)
}

func portugueseValidationMessage(message string) string {
	translations := map[string]string{
		"is required":                            "é obrigatório",
		"is required in the final configuration": "é obrigatório na configuração final",
		"must be the integer 1":                  "deve ser o número inteiro 1",
		"must be a mapping":                      "deve ser um mapeamento",
		"must be a non-empty string":             "deve ser uma string não vazia",
		"must be a boolean":                      "deve ser um booleano",
		"must be a string":                       "deve ser uma string",
		"must be a list":                         "deve ser uma lista",
		"must be a list of strings":              "deve ser uma lista de strings",
		"must be a mapping of strings":           "deve ser um mapeamento de strings",
		"must be a list of project-local paths":  "deve ser uma lista de caminhos locais do projeto",
		"must be a list of local file paths":     "deve ser uma lista de caminhos de arquivos locais",
		"must be a non-empty project-local path": "deve ser um caminho local do projeto não vazio",
		"must be a non-empty local file path":    "deve ser um caminho de arquivo local não vazio",
		"must contain at least one step":         "deve conter pelo menos uma etapa",
		"must contain at least one scope when commit.require_scope is true":                           "deve conter pelo menos um escopo quando commit.require_scope for true",
		"must not escape the project root":                                                            "não pode sair da raiz do projeto",
		"must not be empty":                                                                           "não pode estar vazio",
		"must be conventional":                                                                        "deve ser conventional",
		"must be a Go duration such as 30s or 5m":                                                     "deve ser uma duração Go, como 30s ou 5m",
		"remote paths are not supported in v1":                                                        "caminhos remotos não são permitidos na v1",
		"remote URLs are not supported in v1":                                                         "URLs remotas não são permitidas na v1",
		"unknown field":                                                                               "campo desconhecido",
		"workflow names must be non-empty strings":                                                    "nomes de workflow devem ser strings não vazias",
		"duplicate workflow name":                                                                     "nome de workflow duplicado",
		"workflow name must contain only letters, numbers, underscores, and hyphens":                  "o nome do workflow deve conter apenas letras, números, underscores e hífens",
		"environment variable names must be non-empty strings":                                        "nomes de variáveis de ambiente devem ser strings não vazias",
		"environment variable name must match [A-Za-z_][A-Za-z0-9_]*":                                 "o nome da variável de ambiente deve corresponder a [A-Za-z_][A-Za-z0-9_]*",
		"command and run are mutually exclusive":                                                      "command e run são mutuamente exclusivos",
		"must define command with optional args, or an explicitly enabled shell run":                  "deve definir command com args opcionais ou habilitar explicitamente um run via shell",
		"must be true for a run string":                                                               "deve ser true para uma string run",
		"must be true for a shell step":                                                               "deve ser true para uma etapa de shell",
		"must be true when shell steps are configured":                                                "deve ser true quando houver etapas de shell",
		"must be false for structured command steps":                                                  "deve ser false para etapas com comando estruturado",
		"is only supported with a structured command":                                                 "só é permitido com um comando estruturado",
		"must name one executable; place parameters in args":                                          "deve nomear um executável; coloque os parâmetros em args",
		"workflow name is reserved by the run command":                                                "o nome do workflow é reservado pelo comando run",
		"must be a positive Go duration such as 30s or 5m":                                            "deve ser uma duração Go positiva, como 30s ou 5m",
		"must be one of read, local-write, network, git-write, or destructive":                        "deve ser read, local-write, network, git-write ou destructive",
		"cyclic extends reference detected":                                                           "referência cíclica em extends detectada",
		"configuration must contain exactly one YAML document":                                        "a configuração deve conter exatamente um documento YAML",
		"configuration must be a YAML mapping":                                                        "a configuração deve ser um mapeamento YAML",
		"duplicate mapping key":                                                                       "chave YAML duplicada",
		"duplicate YAML key":                                                                          "chave YAML duplicada",
		"YAML aliases are not supported":                                                              "aliases YAML não são permitidos",
		"YAML merge keys are not supported":                                                           "chaves de mesclagem YAML não são permitidas",
		"mapping keys must be strings":                                                                "chaves de mapeamento devem ser strings",
		"must remain inside the project root; use --allow-external-extends only for a trusted preset": "deve permanecer na raiz do projeto; use --allow-external-extends somente para um preset confiável",
	}
	if translated, ok := translations[message]; ok {
		return translated
	}
	if strings.HasPrefix(message, "references unresolved environment variable ") {
		return "referencia uma variável de ambiente não definida " + strings.TrimPrefix(message, "references unresolved environment variable ")
	}
	if strings.HasPrefix(message, "configuration exceeds the ") && strings.HasSuffix(message, " byte limit") {
		limit := strings.TrimSuffix(strings.TrimPrefix(message, "configuration exceeds the "), " byte limit")
		return "a configuração excede o limite de " + limit + " bytes"
	}
	if strings.HasPrefix(message, "YAML exceeds the ") && strings.HasSuffix(message, " node limit") {
		limit := strings.TrimSuffix(strings.TrimPrefix(message, "YAML exceeds the "), " node limit")
		return "o YAML excede o limite de " + limit + " nós"
	}
	if strings.HasPrefix(message, "custom YAML tag ") && strings.HasSuffix(message, " is not supported") {
		tag := strings.TrimSuffix(strings.TrimPrefix(message, "custom YAML tag "), " is not supported")
		return "a tag YAML personalizada " + tag + " não é permitida"
	}
	prefixes := []struct {
		from string
		to   string
	}{
		{"YAML exceeds the maximum depth of ", "o YAML excede a profundidade máxima de "},
		{"could not decode resolved configuration: ", "não foi possível decodificar a configuração resolvida: "},
		{"could not resolve file: ", "não foi possível resolver o arquivo: "},
		{"could not read file: ", "não foi possível ler o arquivo: "},
		{"invalid YAML: ", "YAML inválido: "},
		{"extended file not found: ", "arquivo estendido não encontrado: "},
		{"could not resolve extended file: ", "não foi possível resolver o arquivo estendido: "},
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(message, prefix.from) {
			return prefix.to + strings.TrimPrefix(message, prefix.from)
		}
	}
	return message
}

type ErrorList []ValidationError

func (l ErrorList) Error() string {
	if len(l) == 0 {
		return ""
	}

	parts := make([]string, 0, len(l))
	for _, err := range l {
		parts = append(parts, err.Error())
	}
	return strings.Join(parts, "\n")
}
