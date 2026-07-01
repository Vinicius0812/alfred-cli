# Alfred CLI

Alfred is an open source developer assistant for automating repetitive project
tasks with predictable, versioned configuration.

The first release focuses on:

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

After the first GitHub release is published, users can install Alfred with the
provided scripts.

Windows PowerShell:

```powershell
iwr https://raw.githubusercontent.com/Vinicius0812/alfred-cli/master/scripts/install.ps1 -UseB | iex
```

Linux/macOS:

```bash
curl -fsSL https://raw.githubusercontent.com/Vinicius0812/alfred-cli/master/scripts/install.sh | sh
```

Specific versions can be installed with `ALFRED_VERSION`:

```bash
curl -fsSL https://raw.githubusercontent.com/Vinicius0812/alfred-cli/master/scripts/install.sh | ALFRED_VERSION=v0.1.0 sh
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

## License

Alfred is released under the [MIT License](LICENSE).
