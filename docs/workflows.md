# Secure Workflows

Alfred workflows automate reviewed project tasks. They are not an OS sandbox:
once a step is approved, its process has the same permissions as the user who
started Alfred.

## Prefer structured execution

```yaml
workflows:
  test:
    description: Run all tests
    steps:
      - name: Go tests
        command: go
        args:
          - test
          - ./...
        risk: read
        timeout: 10m
```

Alfred calls the executable directly. A malicious-looking argument such as
`; rm -rf .` remains one literal argument instead of becoming a second shell
command. Do not wrap a structured command in `sh -c`, `bash -c`, `cmd /c`, or
`powershell -Command`; that recreates shell-injection risk outside Alfred's
explicit shell guardrails.

## Classify risk accurately

| Risk | Typical examples | Default confirmation |
| --- | --- | --- |
| `read` | tests, linters, inspection | no |
| `local-write` | compilation, formatting, generated files | no |
| `network` | dependency download, vulnerability APIs | yes |
| `git-write` | commit, tag, branch mutation | yes |
| `destructive` | deletion, prune, irreversible replacement | yes |

Set `confirm: true` for any step that deserves confirmation regardless of its
risk class. `--yes` cannot bypass a confirmation unless the project explicitly
sets `policies.allow_non_interactive: true`.

When `require_clean_worktree` is true, Alfred runs `git status --porcelain`
before the first `git-write` step and blocks the workflow if changes exist.

## Environment and secrets

```yaml
environment:
  API_TOKEN: ${API_TOKEN}
```

Variables must already exist in Alfred's environment. A defined empty value is
valid; an undefined reference blocks validation. Workflow values are inherited
by steps, and step values win.

Alfred never includes resolved values in plans or JSON audit records. Values
configured for a workflow are also masked from stdout, stderr, and recorded
error strings. Inherited environment values whose names identify tokens,
passwords, secrets, private/API keys, or credentials receive the same output
masking. Redaction is defense in depth: commands should still avoid
printing secrets, and users should never place literal credentials in YAML.

## Working directories and executables

Workflow and step directories must remain within the project root after
symbolic-link resolution. A command containing a slash is treated as a path and
receives the same restriction. A bare executable name is resolved using the
current process `PATH`.

For stronger reproducibility, use trusted tool installations, lock module or
package versions, and run Alfred from a minimally privileged account.

## Exceptional shell mode

Shell mode exists for reviewed commands that cannot be expressed as an
executable plus arguments. It is disabled by default and always interactive
unless non-interactive confirmation is explicitly allowed.

```yaml
policies:
  allow_shell_steps: true

workflows:
  report:
    description: Produce a test report through a reviewed pipeline
    steps:
      - name: Test report
        run: go test ./... | tee test.log
        shell: true
        confirm: true
        risk: local-write
        timeout: 10m
```

On Windows, shell steps use `cmd.exe /D /S /C`. On Unix-like systems they use
`/bin/sh -c`. Review quoting for every supported platform.

## Planning, cancellation, and outcomes

`alfred run NAME --dry-run` performs configuration validation, environment
resolution, path checks, confirmation classification, and plan creation. It
does not start a configured subprocess.

During execution, output is streamed in real time. Each step has a timeout;
Ctrl+C cancels the active process through Go's command context. Outcomes are:

User or parent-context cancellation always stops the workflow, even when the
active step has `continue_on_error: true`.

- `planned`: dry-run completed;
- `success`: every step succeeded;
- `success_with_warnings`: only `continue_on_error` steps failed;
- `failed`: a command or required operation failed;
- `blocked`: policy or precondition denied execution;
- `declined`: the user rejected confirmation.

## Audit records

When `policies.audit` is true, Alfred writes an atomic JSON record with mode
`0600` where supported to `.alfred/runs/<id>.json`. It includes:

- workflow, Alfred version, configuration path, and configuration hash;
- start and finish timestamps;
- dry-run status and final outcome;
- step names, risk, duration, exit code, and confirmation state;
- redacted error diagnostics.

It does not include command environment values or process output. Audit records
are local evidence, not tamper-proof signatures. Protect the project directory
if the integrity of local history matters.
