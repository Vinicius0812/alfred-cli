# Alfred Command Reference

## Global syntax

```text
alfred [global options] <command>
```

| Option | Purpose |
| --- | --- |
| `--config <path>` | Use a specific configuration file instead of discovery |
| `--lang <en\|pt-BR>` | Select the interface language (`--language` is an alias) |
| `--allow-external-extends` | Trust local presets outside the project root |
| `--no-color` | Disable terminal colors; reserved for stable scripting behavior |

`ALFRED_LANG` supplies the default language. An explicit `--lang` wins. Invalid
language names produce a usage error.

## `alfred menu`

Opens the bilingual interactive menu with Alfred's identity. The menu can
validate configuration, run the doctor, initialize a project, list workflows,
execute a workflow, or show the version. It uses the same command handlers and
security checks as non-interactive invocation.

## `alfred init`

```text
alfred init [--project name] [--preset basic|go-secure] [--force]
```

Creates `alfred.yaml` atomically with permission `0600` where supported. YAML
serialization safely quotes project names. Existing files are preserved unless
`--force` is explicit.

When `--project` is omitted, Alfred uses the current directory name. The global
`--config <path>` option selects a different output file; its parent directory
must already exist.

- `basic` creates the typed policy baseline without workflows;
- `go-secure` adds a `verify` workflow with `go test`, `go vet`, `go mod
  verify`, and `govulncheck`.

## `alfred doctor`

Loads and validates the configuration, then checks whether each bare executable
used by a structured step is available on `PATH`. Local executable paths are
validated by the configuration and at execution time. The doctor never
executes a configured workflow.
It returns failure if configuration is invalid or a required tool is missing.

## `alfred config validate`

```text
alfred config validate [--output text|json]
```

Discovers, parses, merges, and validates the entire configuration graph without
running project commands. Errors contain the source file and schema path. JSON
output is appropriate for CI and contains `valid`, `project`, `source`, and
`files`.

## `alfred config explain`

```text
alfred config explain [--output text|json]
```

Shows the resolved typed configuration, root path, primary source, and inherited
files. Environment references remain unresolved in this output so their values
cannot leak through configuration inspection.

## `alfred workflow list`

```text
alfred workflow list [--output text|json]
```

Lists workflows alphabetically with their descriptions and step counts.

## `alfred run <workflow>`

```text
alfred run <workflow> [--dry-run] [--yes] [--timeout duration] [--output text|json]
```

Builds a deterministic plan from the validated configuration and then executes
steps sequentially.

| Option | Effect |
| --- | --- |
| `--dry-run` | Records and prints the plan without starting a child process |
| `--yes`, `-y` | Answers confirmations only when `allow_non_interactive` is true |
| `--timeout <duration>` | Overrides every step timeout for this invocation |
| `--output json` | Writes a plan/execution object to stdout; subprocess output goes to stderr |

Interactive confirmation also goes to stderr in JSON mode, preserving stdout as
one valid JSON document. The text plan contains command names, arguments,
directories, risk, and timeout,
but not resolved environment values. Ctrl+C cancels the active child process
and returns exit code 130. A failed step stops the workflow unless it declares
`continue_on_error: true`; user cancellation is never treated as a continuable
failure.

## `alfred run list`

```text
alfred run list [--output text|json]
```

Lists local audit records newest first. Records live under `.alfred/runs`, a
project-local path ignored by the repository's `.gitignore`.

## `alfred run show <id>`

```text
alfred run show <id> [--output text|json]
```

Reads a single audit record. Run IDs are strictly validated before path
construction, preventing traversal through this command.

## `alfred completion`

```text
alfred completion bash
alfred completion zsh
alfred completion powershell
```

`pwsh` is accepted as an alias for `powershell`.

Prints a completion script to stdout. Installation examples:

```bash
alfred completion bash > ~/.local/share/bash-completion/completions/alfred
alfred completion zsh > ~/.zfunc/_alfred
```

```powershell
alfred completion powershell | Out-String | Invoke-Expression
```

Add the PowerShell expression to the user's profile for persistence only after
reviewing the generated script.

## `alfred version`, `help`, and flags

`alfred version` and `alfred --version` print the embedded build version.
`alfred help` and `alfred --help` print the current command summary.

## Exit codes

| Code | Meaning |
| --- | --- |
| `0` | Command completed successfully, including an accepted dry-run plan |
| `1` | Validation, policy, environment, execution, or audit failure |
| `2` | Invalid command, flag, language, workflow name, or other usage error |
| `130` | Execution canceled by an interrupt such as Ctrl+C |

JSON field names, risk values, outcome values, and exit codes remain stable in
English even when the human interface language is Portuguese.
