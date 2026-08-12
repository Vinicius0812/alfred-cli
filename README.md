# Alfred CLI

[![CI](https://github.com/Vinicius0812/alfred-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/Vinicius0812/alfred-cli/actions/workflows/ci.yml)

Alfred is an open source CLI for secure, explicit, and auditable development
automation. It turns project checks into versioned workflows while keeping
execution inspectable: structured commands run without a shell, sensitive
operations require policy-based confirmation, and every run can produce a
local audit record.

```text
       /\                 /\
      /  \__         __/  \
      \     \_______/     /
       \____/ ALFRED \____/
```

The interface is available in English and Brazilian Portuguese.

## Highlights

- strict, typed `alfred.yaml` configuration with actionable validation;
- duplicate-key, alias, size, depth, path, symlink, and environment guardrails;
- local preset inheritance, restricted to the project by default;
- structured `command` + `args` execution without a platform shell;
- explicit opt-in and confirmation for exceptional shell steps;
- risk classes for read, local write, network, Git write, and destructive work;
- dry-run planning, timeouts, Ctrl+C cancellation, and stable exit codes;
- secret redaction from subprocess output and audit diagnostics;
- local JSON audit records under `.alfred/runs`;
- human-readable or JSON output for automation;
- a secure Go preset and completions for Bash, Zsh, and PowerShell;
- versioned native releases with checksums, SBOMs, provenance attestations, and
  post-publication smoke tests.

## Install

Download the installer from the release you intend to trust. The installers
verify the selected binary archive against the published SHA-256 checksum
before installation.

Windows PowerShell:

```powershell
$version = "v0.2.0"
$installer = Join-Path $env:TEMP "install-alfred.ps1"
Invoke-WebRequest -Uri "https://github.com/Vinicius0812/alfred-cli/releases/download/$version/install.ps1" -OutFile $installer
PowerShell -NoProfile -ExecutionPolicy Bypass -File $installer -Version $version
```

Linux or macOS:

```bash
version="v0.2.0"
curl -fsSLo install-alfred.sh "https://github.com/Vinicius0812/alfred-cli/releases/download/$version/install.sh"
ALFRED_VERSION="$version" sh ./install-alfred.sh
```

To build from source:

```bash
go build -trimpath -o bin/alfred ./cmd/alfred
./bin/alfred --version
```

Use `bin/alfred.exe` on Windows.

## Quick start

Create a minimal configuration:

```bash
alfred init --project example-api
```

For a Go project, create the secure preset:

```bash
alfred init --project example-api --preset go-secure
alfred config validate
alfred workflow list
alfred run verify --dry-run
alfred run verify
```

The `go-secure` preset runs tests, static analysis, module verification, and
`govulncheck`. Its networked vulnerability scan is identified in the plan and
requires confirmation by default.

## Commands

```text
alfred menu
alfred init [--project name] [--preset basic|go-secure] [--force]
alfred doctor
alfred config validate [--output text|json]
alfred config explain [--output text|json]
alfred workflow list [--output text|json]
alfred run <workflow> [--dry-run] [--yes] [--timeout duration] [--output text|json]
alfred run list [--output text|json]
alfred run show <id> [--output text|json]
alfred completion <bash|zsh|powershell>
alfred version
alfred help
```

See [Command Reference](docs/commands.md) for behavior, examples, output
contracts, and exit codes.

## Configuration example

```yaml
version: 1

project:
  name: example-api

policies:
  confirm_network_actions: true
  confirm_git_actions: true
  confirm_destructive_actions: true
  audit: true

workflows:
  verify:
    description: Validate the Go project
    steps:
      - name: Run tests
        command: go
        args: [test, ./...]
        risk: read
        timeout: 10m
```

Structured arguments are passed directly to the executable. Alfred does not
concatenate them into a shell command. Review [Configuration v1](docs/configuration-v1.md)
and [Secure Workflows](docs/workflows.md) before enabling shell steps.

## Language

```bash
alfred --lang en help
alfred --lang pt-BR menu
```

Set `ALFRED_LANG=en` or `ALFRED_LANG=pt-BR` to select a default. Unsupported
language values are rejected instead of silently falling back.

## Verify a release

After checking an archive against `checksums.txt`, users with GitHub CLI can
verify that GitHub Actions produced the artifact from this repository:

```bash
gh attestation verify alfred_v0.2.0_linux_amd64.tar.gz \
  --repo Vinicius0812/alfred-cli
```

Each target also has a matching `*.cdx.json` CycloneDX SBOM in the release.
An attestation proves build provenance, not that an artifact is vulnerability
free; users should still evaluate the source, workflow, checksum, and SBOM.

## Security model

Treat `alfred.yaml` as code. Alfred applies guardrails and makes behavior
auditable, but it is not an operating-system sandbox. A confirmed command runs
with the current user's permissions. Remote presets are not supported and
external local presets require the explicit `--allow-external-extends` trust
decision.

See [Security Policy](SECURITY.md), [Secure Workflows](docs/workflows.md),
[Architecture](docs/architecture.md), [Interface Identity](docs/identity.md),
and [Release Roadmap](docs/roadmap.md).

## Roadmap

The workflow engine and secure execution baseline form the `v0.2.0` scope.
Provider-agnostic, review-before-use commit message assistance and Docker
Compose adapters remain future features; neither will bypass the same planning,
confirmation, audit, and secret-handling rules.

## License

Alfred is released under the [MIT License](LICENSE).
