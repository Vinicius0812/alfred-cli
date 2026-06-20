# CONTEXT_UNI.md

Handoff técnico do UNI CLI para outro agente Codex continuar o projeto em outro workspace.

Data do handoff: 2026-06-14

Observação: o prompt original cita "alfred-cli" nos próximos passos, mas o projeto deste workspace é o UNI CLI. Se a intenção futura for renomear ou reaproveitar a base para um `alfred-cli`, tratar isso como iniciativa separada para não quebrar compatibilidade do comando `uni`.

## 1. Objetivo do CLI

O UNI CLI é uma automação interna em Bash para padronizar tarefas recorrentes de desenvolvimento nos projetos UNIUBE, principalmente no ecossistema AVA.

Problemas que resolve:

- Reduz comandos manuais de Git, Docker e configuração local.
- Centraliza o fluxo de commit com Conventional Commits, emojis, seleção de arquivos e confirmação de push.
- Facilita subir/derrubar Docker-FRONT e Docker-API por projeto.
- Mantém registro do projeto Docker ativo para exibir no header sem varrer Docker a cada tela.
- Automatiza merges e atualização de branches com regras de permissão.
- Adiciona sugestão de mensagem de commit via IA, com auditoria de chamadas.
- Separa comandos administrativos em `uni root`, protegidos por senha e permissões.
- Possui `uni init` para preparar ambiente padrão em `C:/Projetos`.

## 2. Stack, runtime e ferramentas

Stack principal:

- Bash, executado normalmente pelo Git Bash no Windows.
- Scripts shell modulares em `uni.sh`, `lib/*.sh` e `providers/*.sh`.
- Não há `package.json`, `pyproject.toml`, Composer ou build step próprio.

Dependências externas esperadas:

- `git`
- `docker`
- `docker-compose` ou `docker compose`
- `curl`
- `node`, usado pelos providers de IA para montar/parsear JSON
- comandos Unix disponíveis no Git Bash: `awk`, `sed`, `grep`, `tr`, `find`, `head`, `cut`, `date`
- IDE CLI opcional: `code`, `cursor` ou `antigravity`

Gerenciador de pacotes:

- Não há gerenciador de pacotes do projeto.
- Instalações dependem do ambiente Windows/Git Bash/Docker/IDE.

Comandos úteis de verificação:

```bash
# Sintaxe Bash de todos os scripts principais
for f in uni.sh lib/*.sh providers/*.sh; do bash -n "$f" || exit 1; done

# Ajuda normal
source ./uni.sh
uni help

# Ajuda root
source ./uni.sh
uni root help

# Diagnóstico operacional
source ./uni.sh
uni doctor
```

Não há comando formal de build/test automatizado. A validação atual costuma ser:

- `bash -n` nos scripts.
- Execução manual de `uni --help`, `uni init --help`, `uni doctor`.
- Testes manuais dos fluxos interativos em Git Bash.

## 3. Estrutura de pastas relevante

```text
scripts/
  uni.sh                         # entrypoint principal e parser de comandos
  CONTEXT_UNI.md                 # este handoff
  permissions.conf               # grupos de permissão
  security.conf                  # hash/salt da senha root, não versionar externamente
  backups/                       # backups datados do CLI
  lib/
    ai_commit_audit.sh           # auditoria do ai_commit
    autocomplete.sh              # autocomplete Bash
    backup.sh                    # backup datado com CHANGELOG.md
    commit_ai.sh                 # sugestão local/script de mensagens de commit
    docker.sh                    # up/down/clean/status Docker
    doctor.sh                    # diagnóstico do ambiente
    git.sh                       # branches, commit, merge e atualização
    help.sh                      # ajuda normal/root e versão
    init.sh                      # setup inicial com clones
    paths.sh                     # paths.conf, IDE e configs do ai_commit
    permissions.sh               # leitura/edição de grupos do permissions.conf
    projects.sh                  # projects.conf e cadastro de projetos
    security.sh                  # senha e autenticação root
    ui.sh                        # header, menus por seta, logs, spinner, prompts
  providers/
    ai_commit_openrouter.sh      # provider padrão via OpenRouter
    ai_commit_openai.sh          # provider alternativo via OpenAI Responses
    secrets.conf                 # secrets compartilhado, contém dados sensíveis
    secrets.conf.example         # modelo seguro sem chaves reais
    logs/
      openrouter_calls.log       # auditoria de chamadas OpenRouter
```

Arquivos gerados no usuário:

```text
~/.uni/paths.conf                # PROJECTS_ROOT, DOCKER_ROOT, IDE, ai_commit
~/.uni/projects.conf             # projetos cadastrados
~/.uni/secrets.conf              # secrets pessoais, prioridade sobre providers/secrets.conf
~/.uni/docker_state.conf         # projeto Docker ativo salvo pelo CLI
~/.uni/tmp/                      # temporários de commit, seleção e diff
```

Config local por repositório:

```text
.uni.local.conf                  # se existir no diretório atual, substitui ~/.uni/projects.conf
```

## 4. Entrypoint/bin e registro

Entrypoint real:

```bash
uni.sh
```

O arquivo define a função Bash `uni()`. Em geral, o comando `uni` funciona porque `uni.sh` é carregado no shell do usuário, por exemplo via `source /caminho/para/scripts/uni.sh` no `.bashrc`, `.bash_profile` ou wrapper interno.

Trecho essencial:

```bash
UNI_VERSION="3.0.0"

CONFIG_GLOBAL="$HOME/.uni/projects.conf"
CONFIG_LOCAL=".uni.local.conf"
UNI_SCRIPT_DIR="${UNI_SCRIPT_DIR:-$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)}"
PERMISSIONS_FILE="${UNI_PERMISSIONS_FILE:-${PERMISSIONS_FILE:-$UNI_SCRIPT_DIR/permissions.conf}}"
PATHS_FILE="$HOME/.uni/paths.conf"
UNI_LIB_DIR="$UNI_SCRIPT_DIR/lib"

for module in \
  ui.sh \
  permissions.sh \
  security.sh \
  paths.sh \
  commit_ai.sh \
  init.sh \
  docker.sh \
  projects.sh \
  git.sh \
  ai_commit_audit.sh \
  backup.sh \
  help.sh \
  doctor.sh \
  autocomplete.sh
do
  source "$UNI_LIB_DIR/$module"
done
```

Autocomplete:

```bash
lib/autocomplete.sh
```

Registra comandos principais como `init`, `up`, `down`, `branch`, `commit`, `ai_commit`, `config`, `doctor`, `backup`, `root`, etc.

## 5. Comandos, flags, exemplos e outputs esperados

### Menu principal

```bash
uni
```

Abre menu interativo com navegação por setas e atalhos numéricos.

Opções atuais:

```text
[1] Docker
[2] Branch
[3] Commit
[4] Merge
[5] Atualizar
[6] Configurações
[7] Doctor
[8] Ajuda
[0] Sair
```

Output esperado:

```text
UNI CLI v3.0.0
dir: /c/Projetos/AVA
docker: corp | front: corp_front | api: corp_api
branch: homologacao

Escolha uma opção:
```

### Init

```bash
uni init
uni init --yes
uni init --help
```

Objetivo:

- Cria `/c/Projetos`.
- Clona AVA front em `/c/Projetos/AVA`.
- Clona API em `/c/Projetos/api/server/API`.
- Clona Docker-Uniube duas vezes: `/c/Projetos/Docker-FRONT` e `/c/Projetos/Docker-API`.
- Configura `PROJECTS_ROOT=/c/Projetos` e `DOCKER_ROOT=/c/Projetos`.
- Registra projeto `ava|ava|api|/c/Projetos/AVA/`.

Repositórios usados:

```text
AVA front: https://git.uniube.br:3000/Uniube/AVA.git
API:       https://git.uniube.br:3000/Uniube/API.git
Dockers:   https://git.uniube.br:3000/DTI/Docker-Uniube.git
```

Output esperado:

```text
UNI CLI - Init

Este comando vai preparar o ambiente padrão do UNI CLI.

Pasta base         /c/Projetos
AVA front          /c/Projetos/AVA
API AVA            /c/Projetos/api/server/API
Docker FRONT       /c/Projetos/Docker-FRONT
Docker API         /c/Projetos/Docker-API

Confirmar inicialização do ambiente? (y/N):
```

Compatibilidade:

- O comando é idempotente.
- Se uma pasta já tem `.git`, o clone é ignorado.
- Se uma pasta existe e não está vazia, não sobrescreve.

### Docker

```bash
uni up <projeto>
uni down <projeto>
uni clean_docker
```

Exemplos:

```bash
uni up ava
uni down ava
uni clean_docker
```

Comportamento:

- Usa `PROJECTS_ROOT` e `DOCKER_ROOT` de `~/.uni/paths.conf`.
- Espera Docker-FRONT em `$DOCKER_ROOT/Docker-FRONT`.
- Espera Docker-API em `$DOCKER_ROOT/Docker-API` quando o projeto tem branch API.
- Ao subir projeto, derruba automaticamente o projeto ativo anterior, se houver.
- Ao subir, faz `git checkout -f <branch>`, `git pull` e `docker compose up -d`.
- Ao derrubar, faz `docker compose down`.
- Ao subir, oferece abrir diretório final na IDE atual, nova janela ou não abrir.
- Salva projeto ativo em `~/.uni/docker_state.conf`.

Output esperado:

```text
[INFO] Executando 'up' para 'ava'...
Starting 100%
[OK] Processo finalizado.
[INFO] Projeto 'ava' iniciado com sucesso.
Diretório final: /c/Projetos/AVA/
```

`uni clean_docker` executa:

```bash
docker system prune -af
```

com confirmação.

### Projetos

```bash
uni project add <nome> <front_branch> <api_branch|none> <diretorio>
uni project remove <nome>
uni project list
```

Exemplo:

```bash
uni project add ava ava api /c/Projetos/AVA/
uni project list
```

Formato de `~/.uni/projects.conf`:

```text
nome|front_branch|api_branch|diretorio_final
ava|ava|api|/c/Projetos/AVA/
```

Se não tiver API:

```text
agro|agro_front||/c/Projetos/WebProdAgro/
```

### Branch

```bash
uni branch list
uni branch switch <nome>
uni branch create <nome>
uni branch delete <nome>
```

Comportamento:

- `list`: lista branches locais e destaca a atual.
- `switch`: troca para branch local existente.
- `create`: parte de `homologacao`, faz pull, cria branch e publica com upstream.
- `delete`: bloqueia `main`, `master` e `homologacao`; pede confirmação; remove local/remoto.

Output esperado:

```text
Branches locais:
----------------------------------------
  * homologacao
    minha-feature
----------------------------------------
```

### Commit

```bash
uni commit
uni commit <tipo> "mensagem"
uni commit feat "adiciona fluxo de matrícula"
```

Tipos:

```text
feat, fix, docs, style, refactor, test, chore, perf, wip, misc
```

Emojis usados no header:

```text
feat ✨
fix 🐛
docs 📝
style 🎨
refactor ♻️
test 🧪
chore 🔧
perf 🚀
wip 🚧
misc 🔖
```

Fluxo interativo:

1. Valida que está em repositório Git.
2. Bloqueia commit direto em `homologacao`, `main` e `master` conforme permissões.
3. Coleta arquivos modificados/novos/deletados.
4. Se houver apenas 1 arquivo, ignora menu de seleção.
5. Se houver mais de 1 arquivo, abre seleção por setas/espaço para incluir/excluir arquivos do commit.
6. Arquivos fora do commit são guardados temporariamente em stash.
7. Após seleção, gera sugestões por IA/local, se configurado.
8. Usuário escolhe sugestão, edita ou insere manualmente.
9. Mostra prévia do commit.
10. Confirma commit local.
11. Pergunta se deseja fazer push.
12. Restaura arquivos excluídos do commit via stash.

Formato do commit:

```text
<emoji> <tipo> (<branch-ou-scope>): <titulo>

<corpo opcional>

[YYYY-MM-DD HH:MM] Branch: <branch>
```

Se a sugestão de IA vier com descrição detalhada, o corpo pode virar tópicos:

```text
feat|melhora fluxo de commit|ajusta seleção de arquivos; adiciona sugestões por IA; preserva stash dos arquivos excluídos
```

vira:

```text
✨ feat (minha-branch): melhora fluxo de commit

- ajusta seleção de arquivos
- adiciona sugestões por IA
- preserva stash dos arquivos excluídos

[2026-06-14 10:30] Branch: minha-branch
```

Output esperado:

```text
UNI CLI - Seleção de Arquivos

Escolha os arquivos que devem entrar no commit:

> [x] [M] view/pages/exemplo/index.php
  [ ] [M] view/pages/exemplo/script.js

Use ↑ ↓, espaço para marcar/desmarcar, A para marcar todos, N para limpar e Enter para confirmar.
Esc cancela.
```

### AI Commit

```bash
uni ai_commit logs
uni ai_commit audit
uni ai_commit help
```

Observação:

- Auditoria exige root ou usuário autorizado.
- O provider padrão é script externo configurado em `COMMIT_AI_SCRIPT`.
- Provider padrão atual: `providers/ai_commit_openrouter.sh`.
- Provider alternativo: `providers/ai_commit_openai.sh`.

Output esperado da auditoria:

```text
UNI CLI v3.0.0
Auditoria ai_commit

Log: /k/Sistema AVA/scripts/providers/logs/openrouter_calls.log

Data | Usuário | Repositório | Branch | Status | Custo | Arquivos
-------------------------------------------------------------------
...
```

Não expor chaves reais em documentação, logs ou prompts.

### Merge

```bash
uni quick_merge_to_hm
uni merge_to_master
uni merge_current_to_master
```

Atalhos conceituais:

```text
branch atual -> homologacao
homologacao -> master
branch atual -> homologacao -> master
```

Permissão:

- Restringido por `authorized_merge_users` em `permissions.conf`.

Comportamento geral:

- Valida worktree limpo.
- Faz checkout/pull das branches envolvidas.
- Em conflito, aborta merge automaticamente quando aplicável.
- Faz push após concluir fluxos restritos.

### Atualizar

```bash
uni update_current_with_hm
uni update_hm_with_main
uni update_current_with_main
```

Atalhos conceituais:

```text
branch atual <- homologacao
homologacao <- main/master
branch atual <- homologacao <- main/master
```

Permissão:

- `update_current_with_hm` pode ser executado por qualquer usuário.
- `update_hm_with_main` e `update_current_with_main` exigem permissão de merge.

### Configurações

```bash
uni config show
uni config projects_root
uni config docker_root
uni config ide
uni config ai_enabled
uni config ai_provider
uni config ai_script
uni paths edit
```

Aliases:

```bash
uni paths ...
uni configs ...
uni configuracoes ...
uni configurações ...
```

Config atual esperada:

```text
PROJECTS_ROOT=/c/Projetos
DOCKER_ROOT=/c/Projetos
IDE_NAME=vscode
COMMIT_AI_ENABLED=on
COMMIT_AI_PROVIDER=script
COMMIT_AI_SCRIPT=/.../providers/ai_commit_openrouter.sh
```

IDEs aceitas:

```text
vscode, cursor, antigravity
```

### Doctor

```bash
uni doctor
```

Valida:

- Git
- branch atual
- Docker
- Docker Compose
- `paths.conf`
- pastas de projeto/Docker
- `projects.conf`
- permissões
- IDE
- ai_commit
- secrets de IA sem expor valores

Output esperado:

```text
UNI CLI - Doctor

[OK] Git: Comando git disponível.
[WARN] Docker engine: Docker não respondeu. Verifique se o Docker Desktop está aberto.
[OK] paths.conf: ~/.uni/paths.conf

Resumo: 8 OK, 1 WARN, 0 ERRO
```

### Root/admin

```bash
uni root
uni root setup
uni root reset
uni root help
```

Comandos diretos protegidos:

```bash
uni backup [resumo]
uni ai_commit logs
```

Menu root inclui:

```text
[1] Auditoria ai_commit
[2] Backup com changelog
[3] Doctor
[4] Permissões
[5] Trocar senha root
[6] Ajuda root
[0] Voltar
```

Permissão:

- `cli_admin_users` em `permissions.conf`.
- `security.conf` guarda hash/salt da senha root.

Nota de implementação:

- `root_read_password` usa leitura visível (`read -r -p`) porque `read -s` apresentou problema no Git Bash/terminal usado. Se for reativar senha oculta, testar no terminal real do time.

### Backup

```bash
uni backup "resumo da versão estável"
```

Exige root.

Cria pasta datada em:

```text
backups/YYYY-MM-DD/
backups/YYYY-MM-DD_HHMMSS/
```

Inclui scripts/configs e `CHANGELOG.md`.

### Versão e ajuda

```bash
uni --version
uni -v
uni help
uni root help
```

## 6. Módulos internos e responsabilidades

```text
uni.sh
```

- Bootstrap dos módulos.
- Sincronização inicial de contexto com spinner.
- Menus principais.
- Parser de comandos.

```text
lib/ui.sh
```

- Cores, logs `[INFO]`, `[OK]`, `[WARN]`, `[ERRO]`.
- Header com versão, usuário, diretório, Docker ativo e branch Git.
- Menus por setas, seleção múltipla, prompts, pause, spinner e transições.

```text
lib/init.sh
```

- `uni init`.
- Cria `/c/Projetos`.
- Clona AVA, API e Docker-Uniube duas vezes.
- Configura paths e projeto `ava`.

```text
lib/paths.sh
```

- Criação/leitura/escrita de `~/.uni/paths.conf`.
- Configuração de pasta de projetos, Docker, IDE e ai_commit.
- Abertura de diretório em IDE (`code`, `cursor`, `antigravity`).

```text
lib/projects.sh
```

- Criação/leitura de `~/.uni/projects.conf`.
- `project add/remove/list`.
- Suporta `.uni.local.conf`.

```text
lib/docker.sh
```

- Resolve `docker-compose` vs `docker compose`.
- Sobe/derruba Docker-FRONT e Docker-API.
- Detecta/salva projeto ativo em `~/.uni/docker_state.conf`.
- Limpa Docker com `docker system prune -af`.
- Abre diretório final na IDE após `up`.

```text
lib/git.sh
```

- Branch list/switch/create/delete.
- Tipos de commit, emojis e assistente de commit.
- Seleção de arquivos por menu.
- Stash temporário de arquivos excluídos do commit.
- Confirmação de push.
- Fluxos de merge e atualização de branches.

Importante: `lib/git.sh` atualmente possui duplicidade aparente de `commit_code_interactive()` e `commit_code()` em pontos diferentes do arquivo. Em Bash, a última definição vence. Ao alterar fluxo de commit, edite a definição efetivamente usada no fim do arquivo e considere consolidar essa duplicidade como refactor futuro.

```text
lib/commit_ai.sh
```

- Sugestões locais ou via script externo.
- Normalização de provider `local`/`script`.
- Parser de sugestões no formato `tipo|mensagem` e `tipo|título|descrição`.
- Conversão de descrições separadas por `;` em tópicos no corpo do commit.

```text
providers/ai_commit_openrouter.sh
```

- Provider padrão.
- Chama OpenRouter Chat Completions.
- Usa `curl` e `node`.
- Registra auditoria em log.
- Tenta buscar custo em endpoint de geração quando possível.

```text
providers/ai_commit_openai.sh
```

- Provider alternativo.
- Chama OpenAI Responses API.
- Usa `curl` e `node`.

```text
lib/ai_commit_audit.sh
```

- Lê e formata log de chamadas do OpenRouter.
- Restrito por root/permissão.

```text
lib/permissions.sh
```

- Garante/cria `permissions.conf`.
- Lê grupos de permissão.
- Adiciona/remove usuários por grupo.

```text
lib/security.sh
```

- Senha root, hash SHA-256 com salt.
- `root_require_access`.
- `root setup` e `root reset`.

```text
lib/backup.sh
```

- Backup datado do CLI.
- Gera `CHANGELOG.md`.

```text
lib/doctor.sh
```

- Diagnóstico de ambiente.

```text
lib/help.sh
```

- Ajuda normal.
- Ajuda root.
- Versão.

```text
lib/autocomplete.sh
```

- Bash completion.

## 7. Variáveis, configs, credenciais e integrações

### Variáveis globais do CLI

```bash
UNI_VERSION="3.0.0"
UNI_SCRIPT_DIR
UNI_LIB_DIR
UNI_WIDTH
CONFIG_GLOBAL="$HOME/.uni/projects.conf"
CONFIG_LOCAL=".uni.local.conf"
PATHS_FILE="$HOME/.uni/paths.conf"
PERMISSIONS_FILE="$UNI_SCRIPT_DIR/permissions.conf"
```

### paths.conf

Arquivo:

```text
~/.uni/paths.conf
```

Campos:

```bash
PROJECTS_ROOT=/c/Projetos
DOCKER_ROOT=/c/Projetos
IDE_NAME=vscode
COMMIT_AI_ENABLED=on
COMMIT_AI_PROVIDER=script
COMMIT_AI_SCRIPT=/caminho/para/providers/ai_commit_openrouter.sh
```

### projects.conf

Arquivo:

```text
~/.uni/projects.conf
```

Formato:

```text
nome|front_branch|api_branch|diretorio_final
```

### permissions.conf

Arquivo:

```text
permissions.conf
```

Grupos:

```bash
authorized_merge_users=
authorized_homologacao_commit_users=
authorized_mainline_commit_users=
cli_admin_users=
authorized_ai_commit_audit_users=
```

Não é necessário listar nomes reais no handoff externo. Manter o arquivo real no workspace interno.

### security.conf

Arquivo:

```text
security.conf
```

Guarda hash/salt da senha root. Não deve conter senha em texto puro.

### Secrets de IA

Arquivos:

```text
providers/secrets.conf
providers/secrets.conf.example
~/.uni/secrets.conf
```

Prioridade recomendada:

1. variáveis `UNI_*`/ambiente
2. `~/.uni/secrets.conf` do usuário
3. `providers/secrets.conf` compartilhado

Nunca registrar chave real em documentação, commit, Notion, handoff ou output.

Variáveis OpenRouter:

```bash
OPENROUTER_API_KEY=""
OPENROUTER_MODEL="openai/gpt-4o-mini"
OPENROUTER_ENDPOINT="https://openrouter.ai/api/v1/chat/completions"
OPENROUTER_APP_NAME="UNI CLI"
OPENROUTER_SITE_URL=""
UNI_AI_COMMIT_AUDIT_LOG=""
```

Variáveis OpenAI:

```bash
OPENAI_API_KEY=""
OPENAI_MODEL="gpt-5.2"
OPENAI_ENDPOINT="https://api.openai.com/v1/responses"
```

## 8. Fluxo de autenticação e permissões

Permissões normais:

- São baseadas no usuário retornado por `whoami`.
- `permissions.conf` define os grupos.

Root:

- `uni root` exige senha root e/ou usuário em `cli_admin_users`.
- `uni root setup` cria/troca senha.
- `uni root reset` permite redefinir senha para usuários em `cli_admin_users`.
- `UNI_ROOT_AUTHENTICATED=1` é usado como flag em memória durante a sessão/função.

Fluxos protegidos:

- Merges restritos: `authorized_merge_users`.
- Commit direto em homologação: `authorized_homologacao_commit_users`.
- Commit direto em main/master: `authorized_mainline_commit_users`.
- Auditoria ai_commit: `cli_admin_users` ou fallback legado `authorized_ai_commit_audit_users`.
- Backup e root menu: root.

## 9. Dependências externas/APIs usadas

Git interno:

- `git.uniube.br:3000`
- Repositórios do `uni init`:
  - `https://git.uniube.br:3000/Uniube/AVA.git`
  - `https://git.uniube.br:3000/Uniube/API.git`
  - `https://git.uniube.br:3000/DTI/Docker-Uniube.git`

Docker:

- Docker Desktop/engine precisa estar ativo.
- `docker-compose` ou `docker compose`.

OpenRouter:

- Endpoint padrão: `https://openrouter.ai/api/v1/chat/completions`
- Endpoint de custo/generation derivado no provider.
- Modelo padrão: `openai/gpt-4o-mini`.

OpenAI:

- Endpoint padrão: `https://api.openai.com/v1/responses`
- Modelo padrão no arquivo: `gpt-5.2`.

IDE:

- VS Code: `code`
- Cursor: `cursor`
- Antigravity: `antigravity`

## 10. Decisões de arquitetura importantes

- Projeto é Bash puro, sem build, para rodar direto no Git Bash dos devs.
- Módulos em `lib/*.sh` são carregados explicitamente por `uni.sh`.
- Menus interativos usam navegação por setas e atalhos numéricos; `0` é reservado para voltar/sair.
- Header deve ser rápido: Docker ativo vem de `~/.uni/docker_state.conf`, não de varredura Docker a cada render.
- `uni_bootstrap_context` faz uma sincronização inicial com spinner para atualizar contexto de Docker/branch.
- `uni root` separa comandos administrativos para reduzir risco de uso acidental no menu normal.
- `uni ai_commit logs` é protegido por root/permissão.
- IA é modular: `COMMIT_AI_PROVIDER=script` chama `COMMIT_AI_SCRIPT`; dá para trocar provider sem mexer no fluxo de commit.
- Provider retorna sugestões como texto simples, uma por linha, para manter integração robusta com Bash.
- Seleção de arquivos de commit usa arquivos temporários em `~/.uni/tmp`.
- Arquivos excluídos do commit são stashed temporariamente para evitar erro/conflito no push.
- `uni init` não sobrescreve diretórios existentes.

## 11. Estado atual

Funciona:

- Menu principal e submenus por setas.
- Atalhos numéricos com `0` para voltar/sair.
- Header com diretório, usuário, Docker ativo e branch Git quando aplicável.
- Docker up/down/clean/status e abertura em IDE.
- Detecção/salvamento de Docker ativo.
- Projetos via `projects.conf`.
- Branch list/switch/create/delete.
- Commit interativo com seleção de arquivos.
- Commit direto `uni commit feat "mensagem"`.
- Sugestão de commit via IA.
- Corpo de commit detalhado com tópicos quando a IA retorna `tipo|título|descrição`.
- Auditoria de ai_commit.
- Root menu com senha/permissões.
- Doctor.
- Backup root.
- `uni init`.

Incompleto ou pontos frágeis:

- Não há suíte automatizada de testes.
- CLI não está versionado por Git neste workspace; backups datados são o controle atual.
- Há duplicidade de funções de commit em `lib/git.sh`; consolidar para reduzir risco.
- `security.conf` existe no workspace e deve ser tratado como sensível.
- `providers/secrets.conf` pode conter chave real; nunca copiar para docs externas.
- `read -s` para senha foi evitado por problema no terminal; senha root pode aparecer enquanto digita.
- Mensagens exibidas por PowerShell podem parecer com acentuação quebrada dependendo da codepage, mesmo com arquivos sem BOM. Manter UTF-8 sem BOM.
- `uni init` usa caminho `/c/Projetos/api/server/API`; confirmar se o padrão final deveria ser `server` mesmo.
- `uni init` não faz pull/update se repositório já existe; apenas pula.
- Custos de OpenRouter dependem do retorno da API/generation; quando indisponível, auditoria pode mostrar `-`.

Bugs conhecidos ou riscos:

- Editar somente a primeira definição duplicada de `commit_code()` em `lib/git.sh` pode não surtir efeito porque Bash usa a última definição.
- Operações de merge/pull/push são destrutivas o bastante para exigir cuidado; não remover confirmações/permissões.
- `docker_control up` chama `docker_control down` internamente quando há projeto ativo; evitar mudanças que criem recursão indevida.
- `git checkout -f` nos dockers descarta alterações locais nos repositórios Docker-FRONT/API por design.

## 12. Próximos passos desejados

Para UNI CLI:

- Consolidar duplicidade em `lib/git.sh` (`commit_code_interactive`/`commit_code`).
- Criar testes automatizados mínimos com fixtures em Bash, por exemplo `tests/*.bats` ou scripts shell simples.
- Adicionar modo `--dry-run` para `uni init`, merges e Docker.
- Melhorar segurança da senha root com input oculto quando terminal suportar.
- Adicionar rotação/limpeza de logs de auditoria do ai_commit.
- Adicionar comando root para validar/editar `permissions.conf` com backup antes de salvar.
- Criar `README.md` a partir deste handoff.
- Formalizar instalação do comando `uni` com script que adiciona `source` no perfil do Git Bash.
- Validar `uni init` em máquina limpa.
- Adicionar verificação de remote nos clones existentes do `uni init`.
- Criar backup automático antes de alterações grandes no CLI.

Para eventual "alfred-cli":

- Se a intenção for renomear ou derivar o projeto, criar uma branch/pasta separada.
- Definir nome final do binário (`alfred`, `uni`, alias duplo).
- Separar strings e branding para evitar hardcode de `UNI CLI`.
- Definir compatibilidade: `uni` deve continuar funcionando se usuários atuais dependem dele.

## 13. Trechos essenciais

### Parser principal de comandos

```bash
uni() {
  init_ui

  case "$1" in
    up|down)
      docker_control "$1" "$2"
      ;;

    clean_docker)
      docker_clean
      ;;

    init)
      shift
      uni_init_command "$@"
      ;;

    project)
      case "$2" in
        add) project_add "$3" "$4" "$5" "$6" ;;
        remove) project_remove "$3" ;;
        list) project_list ;;
        *) log_error "Uso: uni project add|remove|list" ;;
      esac
      ;;

    branch)
      case "$2" in
        list) list_branches ;;
        switch) switch_branch "$3" ;;
        create) create_branch "$3" ;;
        delete) delete_branch "$3" ;;
        *) uni_help ;;
      esac
      ;;

    commit)
      shift
      commit_code "$@"
      ;;

    ai_commit)
      shift
      case "${1:-logs}" in
        logs|log|audit|auditoria)
          root_require_access || return 1
          ;;
      esac
      ai_commit_audit_command "$@"
      ;;

    paths|config|configs|configuracoes|configurações)
      shift
      paths_command "$@"
      ;;

    doctor)
      doctor_command
      ;;

    backup)
      shift
      root_require_access || return 1
      backup_command "$@"
      ;;

    root)
      shift
      case "${1:-menu}" in
        setup) root_setup_password ;;
        reset) root_reset_password ;;
        help|-h|--help) uni_root_help ;;
        menu|"") uni_root_menu ;;
        *) log_error "Uso: uni root [setup|reset]"; return 1 ;;
      esac
      ;;

    help|-h|--help)
      uni_help
      ;;

    "")
      uni_bootstrap_context
      uni_main_menu
      ;;

    --version|-v)
      uni_version
      ;;

    *)
      uni_help
      ;;
  esac
}
```

### paths.conf default

```bash
PROJECTS_ROOT=/c/Projetos
DOCKER_ROOT=/c/Projetos
IDE_NAME=vscode
COMMIT_AI_ENABLED=on
COMMIT_AI_PROVIDER=script
COMMIT_AI_SCRIPT=/caminho/para/providers/ai_commit_openrouter.sh
```

### projects.conf default

```text
ava|ava|api|/c/Projetos/AVA/
```

### permissions.conf grupos

```bash
authorized_merge_users=
authorized_homologacao_commit_users=
authorized_mainline_commit_users=
cli_admin_users=
authorized_ai_commit_audit_users=
```

### Formato provider de AI commit

```text
tipo|mensagem
tipo|título breve|descrição detalhada
```

Regras atuais:

- 3 a 4 sugestões.
- Pelo menos uma `wip` quando implementação parecer incompleta.
- Usar apenas arquivos incluídos no commit.
- Nunca mencionar arquivos fora do commit.
- Quando houver muitas modificações, usar título breve e descrição detalhada.
- Descrição detalhada pode conter 2 a 4 tópicos separados por `;`.

### Comando init

```bash
uni init [--yes]
```

Pontos-chave:

```bash
UNI_INIT_PROJECTS_ROOT_DEFAULT="/c/Projetos"
UNI_INIT_AVA_REPOSITORY="https://git.uniube.br:3000/Uniube/AVA.git"
UNI_INIT_API_REPOSITORY="https://git.uniube.br:3000/Uniube/API.git"
UNI_INIT_DOCKER_REPOSITORY="https://git.uniube.br:3000/DTI/Docker-Uniube.git"
```

### Auditoria ai_commit

Log padrão:

```text
providers/logs/openrouter_calls.log
```

Pode ser sobrescrito por:

```bash
UNI_AI_COMMIT_AUDIT_LOG=/caminho/custom.log
```

## 14. Cuidados para não quebrar compatibilidade

- Preservar comando público `uni`.
- Preservar aliases `paths`, `config`, `configs`, `configuracoes`, `configurações`.
- Preservar formato de `projects.conf`.
- Preservar grupos de `permissions.conf`.
- Não mover `providers/ai_commit_openrouter.sh` sem atualizar `get_default_commit_ai_script`.
- Não expor secrets reais.
- Manter UTF-8 sem BOM.
- Evitar `git reset --hard`, `rm -rf` ou limpeza destrutiva sem confirmação explícita.
- Não remover pausas em mensagens de erro/sucesso/warning se o fluxo for interativo.
- Manter `0` como voltar/sair nos menus.
- Manter `Esc` cancelando menus quando suportado pela função de UI.
- Validar em Git Bash real, porque alguns comportamentos de teclado/senha diferem do PowerShell.
- Ao mexer em Docker, lembrar que projeto ativo é cacheado em `~/.uni/docker_state.conf`.
- Ao mexer no commit, validar seleção múltipla, stash temporário e restauração.
- Ao mexer em IA, validar fallback local/script e auditoria.

## 15. Checklist para o próximo agente

1. Rodar `bash -n` em `uni.sh`, `lib/*.sh`, `providers/*.sh`.
2. Ler `lib/git.sh` com atenção antes de alterar commit por causa de duplicidade de funções.
3. Não abrir/copiar `providers/secrets.conf` para documentação pública.
4. Testar `uni init --help`, `uni help`, `uni root help`.
5. Testar `uni doctor` em ambiente real.
6. Se alterar menus, testar setas, Enter, Esc e atalhos numéricos.
7. Se alterar Docker, testar `uni up <projeto>` e `uni down <projeto>` com projeto ativo.
8. Se alterar ai_commit, testar provider `local` e `script`.
9. Se alterar permissões/root, testar com usuário autorizado e não autorizado.
10. Criar backup datado antes de grandes mudanças.
