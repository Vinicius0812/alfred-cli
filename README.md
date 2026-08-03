# Alfred CLI

[![CI](https://github.com/Vinicius0812/alfred-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/Vinicius0812/alfred-cli/actions/workflows/ci.yml)

Alfred is an open source developer assistant for automating repetitive project
tasks with predictable, versioned configuration.

The product direction focuses on:

- validating local development environments;
- creating standardized Conventional Commits;
- managing Docker Compose projects;
- running project workflows declared in YAML;
- allowing teams and companies to share reusable presets.

Alfred's core stays generic. Company-specific branch rules, commit policies,
Docker settings, and workflows belong in configuration, not in the executable.

## Current Commands

```text
alfred menu
alfred init
alfred doctor
alfred config validate
alfred version
alfred help
```

## Planned Commands

```text
alfred init
alfred doctor
alfred commit
alfred docker up
alfred docker down
alfred run <workflow>
alfred config validate
```

## Build from source

Alfred is implemented in Go and is intended to be distributed as a native
binary.

```bash
go build -o bin/alfred ./cmd/alfred
```

On Windows:

```powershell
go build -o bin/alfred.exe ./cmd/alfred
```

Then run:

```powershell
.\bin\alfred.exe --version
.\bin\alfred.exe config validate
```

## Install from a GitHub release

Starting with `v0.2.0`, release assets include versioned installer scripts. The
installers download `checksums.txt` and verify the selected archive's SHA-256
digest before extracting or copying the Alfred binary.

Windows PowerShell:

```powershell
$release = Invoke-RestMethod -Uri "https://api.github.com/repos/Vinicius0812/alfred-cli/releases/latest"
$version = $release.tag_name
$installer = Join-Path $env:TEMP "install-alfred.ps1"
Invoke-WebRequest -Uri "https://github.com/Vinicius0812/alfred-cli/releases/download/$version/install.ps1" -OutFile $installer
PowerShell -NoProfile -ExecutionPolicy Bypass -File $installer -Version $version
```

Linux/macOS:

```bash
version="$(curl -fsSL https://api.github.com/repos/Vinicius0812/alfred-cli/releases/latest | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n 1)"
curl -fsSLo install-alfred.sh "https://github.com/Vinicius0812/alfred-cli/releases/download/$version/install.sh"
ALFRED_VERSION="$version" sh ./install-alfred.sh
```

The current `v0.1.0` preview predates automatic checksum verification. Download
its archive and `checksums.txt` from the
[release page](https://github.com/Vinicius0812/alfred-cli/releases/tag/v0.1.0)
and compare the SHA-256 digest manually before installation.

Specific `v0.2.0+` versions can be installed by downloading the installer from
that same release tag and passing `ALFRED_VERSION`:

```bash
version="v0.2.0"
curl -fsSLo install-alfred.sh "https://github.com/Vinicius0812/alfred-cli/releases/download/$version/install.sh"
ALFRED_VERSION="$version" sh ./install-alfred.sh
```

## Language

Alfred supports English and Brazilian Portuguese CLI messages:

```bash
alfred --lang en help
alfred --lang pt-BR help
```

The language can also be selected with:

```bash
ALFRED_LANG=pt-BR alfred help
```

## Configuration

Projects are configured through `alfred.yaml`:

```yaml
version: 1

project:
  name: example-api

commit:
  convention: conventional
  confirm_push: true

docker:
  compose_files:
    - compose.yaml

workflows:
  test:
    description: Run the project test suite
    steps:
      - run: go test ./...
```

See [Product Scope](docs/product.md), [Configuration v1](docs/configuration-v1.md),
and the [examples](examples/) for the current design.

## Status

Alfred is in early implementation. The current CLI can initialize and validate
configuration, run a basic environment doctor, and be built as a native binary.
The remaining MVP commands are still under development.

## Security

Treat project configuration as code and review every workflow or preset before
using future execution commands. See the [Security Policy](SECURITY.md) for the
support policy, trust boundaries, and private reporting guidance.

## License

Alfred is released under the [MIT License](LICENSE).
