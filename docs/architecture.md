# Alfred Architecture

Alfred is a Go native binary with an intentionally small dependency surface.
The CLI entry point delegates to focused internal packages.

```text
cmd/alfred
    |
    v
internal/cli ---------> internal/preset
    |                         |
    v                         v
internal/config ------> typed Config model
    |
    v
internal/workflow ----> child process + .alfred/runs
```

## Packages and responsibilities

### `cmd/alfred`

Owns process startup and exit status only. Build-time `ldflags` inject the
release version into `internal/cli.Version`.

### `internal/cli`

- parses global options and dispatches commands;
- owns English and Brazilian Portuguese human messages;
- renders the brand, help, plans, prompts, and command output;
- exposes `RunWithIO` so menu and command behavior are testable without a real
  terminal;
- converts workflow and configuration results into text or stable JSON;
- maps usage, failure, and interrupt conditions to documented exit codes;
- generates static completion scripts without executing a shell.

Human-facing personality and machine-facing stability are separate contracts.
The rules for that boundary are documented in [Interface Identity](identity.md).

### `internal/config`

- discovers the project configuration;
- applies size and YAML structure limits before model decoding;
- rejects duplicate keys, aliases, merge keys, custom tags, and multiple YAML
  documents;
- resolves and merges local `extends` graphs;
- checks root containment after symlink resolution;
- validates the strict v1 shape, values, paths, policies, and environment
  references;
- decodes the result into typed `Config`, `Workflow`, `Step`, and `Policies`
  models;
- identifies recognized external tools for `doctor`.

### `internal/preset`

Builds typed preset models and serializes them through `yaml.v3`. It never
constructs YAML through string interpolation. `basic` and `go-secure` are the
current presets.

### `internal/workflow`

- creates a run ID and a SHA-256 hash of the typed configuration;
- resolves project-local paths and workflow environment references;
- produces a serializable plan without resolved secret values;
- determines confirmation requirements from step risk and policy;
- invokes structured processes with `exec.CommandContext`;
- permits a platform shell only for an explicitly enabled shell step;
- applies per-step timeout, parent cancellation, controlled environment
  merging, streaming redaction, and `continue_on_error` behavior;
- checks a clean Git worktree before policy-protected `git-write` steps;
- atomically writes and strictly retrieves local JSON audit records.

## Execution flow

1. CLI arguments and language are validated.
2. Configuration files are discovered and parsed within resource limits.
3. The merged YAML graph is validated before decoding into typed models.
4. `BuildPlan` resolves paths, environment references, timeouts, and prompts.
5. Dry-run stops before subprocess creation; normal execution confirms each
   guarded step and starts it with a cancelable context.
6. Output passes through the streaming redactor.
7. A final execution record is atomically persisted when auditing is enabled.

## Security assumptions

- the project configuration and local presets are reviewed code;
- the current operating-system user is authorized to run the selected tools;
- tools resolved from `PATH` are trusted by the user;
- local audit files are not a cryptographically immutable ledger;
- GitHub Actions and release attestations establish provenance, not absence of
  vulnerabilities.

These boundaries keep Alfred focused on reproducible defensive development
automation rather than offensive execution or general-purpose remote agents.
