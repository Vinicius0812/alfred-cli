# Alfred Configuration v1

`alfred.yaml` is a strict, versioned contract. Unknown fields and ambiguous
YAML constructs are rejected so that a typo cannot silently change execution.

## Discovery and trust boundary

Without `--config`, Alfred searches the current directory and its parents for
the first `alfred.yaml`. The directory containing that file is the project root.

```bash
alfred --config path/to/alfred.yaml config validate
```

Paths used by workflows, Docker settings, and `extends` are resolved against
the project root and checked after symbolic-link resolution. They cannot escape
the project root. `extends` may leave the root only when the user passes
`--allow-external-extends`; use that flag solely for a reviewed, trusted preset.
Remote configuration URLs are not supported.

## Parsing guardrails

- maximum size per YAML file: 1 MiB;
- maximum configuration graph: 64 local YAML files;
- maximum YAML depth: 64;
- maximum YAML nodes: 10,000;
- exactly one YAML document per file;
- mappings must use string keys;
- duplicate keys, aliases, merge keys, and custom tags are rejected;
- inheritance cycles and missing files are errors;
- each inherited file is loaded once.

## Merge behavior

Extended files are applied in declaration order, followed by the project file.
Maps merge recursively; scalar and list values replace earlier values. Workflow
maps merge by workflow name. Extended files may be partial, but the final
configuration must contain `version` and `project.name`.

## Top-level schema

Only these keys are valid:

```text
version, extends, project, commit, docker, workflows, policies
```

### `version`

Required integer. Version 1 accepts only `1`.

### `extends`

Optional list of local YAML paths.

```yaml
extends:
  - ./config/alfred.company.yaml
  - ./config/alfred.team.yaml
```

### `project`

```yaml
project:
  name: example-api
```

`name` is a required, non-empty string in the final configuration.

### `commit`

The commit section is typed and validated for the future commit adapter.

```yaml
commit:
  enabled: true
  convention: conventional
  allowed_types: [feat, fix, docs, refactor, test, chore]
  scopes: [api, web]
  protected_branches: [main]
  require_scope: false
  emoji: false
  confirm_commit: true
  confirm_push: true
```

| Field | Type | Default |
| --- | --- | --- |
| `enabled` | boolean | `true` |
| `convention` | `conventional` | `conventional` |
| `allowed_types` | string list | Conventional defaults |
| `scopes` | string list | empty |
| `protected_branches` | string list | `main`, `master` |
| `require_scope` | boolean | `false` |
| `emoji` | boolean | `false` |
| `confirm_commit` | boolean | `true` |
| `confirm_push` | boolean | `true` |

### `docker`

The Docker section is typed and path-validated for the future Compose adapter.

```yaml
docker:
  enabled: true
  compose_files: [compose.yaml]
  profiles: [development]
  env_files: [.env]
  project_name: example-api
```

`compose_files` and `env_files` must stay within the project root, including
after symlink resolution.

### `workflows`

Workflow names may contain letters, numbers, underscores, and hyphens. Each
workflow requires a non-empty description and at least one step.

```yaml
workflows:
  verify:
    description: Verify the project
    working_directory: .
    environment:
      REPORT_TOKEN: ${REPORT_TOKEN}
    steps:
      - name: Run tests
        command: go
        args: [test, ./...]
        working_directory: .
        environment:
          GOFLAGS: -mod=readonly
        risk: read
        timeout: 10m
        confirm: false
        continue_on_error: false
```

Workflow fields:

| Field | Type | Required |
| --- | --- | --- |
| `description` | string | yes |
| `working_directory` | project-local path | no |
| `environment` | string map | no |
| `steps` | step list | yes |

Structured step fields:

| Field | Type | Default |
| --- | --- | --- |
| `name` | string | executable name |
| `command` | string | required unless using explicit shell mode |
| `args` | string list | empty |
| `working_directory` | project-local path | workflow directory |
| `environment` | string map | empty |
| `risk` | risk enum | `local-write` |
| `timeout` | positive Go duration | `10m` |
| `confirm` | boolean | `false` |
| `continue_on_error` | boolean | `false` |

`command` and `args` are passed directly to the process API. Shell operators,
redirection, interpolation, and pipelines have no special meaning.

Shell execution is an exceptional compatibility mode:

```yaml
policies:
  allow_shell_steps: true

workflows:
  exceptional:
    description: A reviewed shell-only operation
    steps:
      - name: Reviewed pipeline
        run: go test ./... | tee test.log
        shell: true
        confirm: true
        risk: local-write
```

A shell step must set all three values: `run`, `shell: true`, and
`confirm: true`; the final policy must also set `allow_shell_steps: true`.
`command` and `run` are mutually exclusive.

Environment names must match `[A-Za-z_][A-Za-z0-9_]*`. Values may reference an
existing variable with `${NAME}`. An undefined variable is an error, while a
defined empty variable is valid. Resolved values are passed to the child
process but omitted from plans and masked from captured process output.

### `policies`

```yaml
policies:
  require_clean_worktree: true
  confirm_destructive_actions: true
  confirm_network_actions: true
  confirm_git_actions: true
  allow_shell_steps: false
  allow_non_interactive: false
  audit: true
```

| Field | Default | Effect |
| --- | --- | --- |
| `require_clean_worktree` | `false` | Blocks `git-write` steps when `git status --porcelain` is non-empty |
| `confirm_destructive_actions` | `true` | Confirms `destructive` steps |
| `confirm_network_actions` | `true` | Confirms `network` steps |
| `confirm_git_actions` | `true` | Confirms `git-write` steps |
| `allow_shell_steps` | `false` | Enables explicitly declared shell steps |
| `allow_non_interactive` | `false` | Allows `--yes` to answer required confirmations |
| `audit` | `true` | Writes `.alfred/runs/<id>.json` |

Risk values are `read`, `local-write`, `network`, `git-write`, and
`destructive`. Risk is a declared security contract, not automatic command
analysis; reviewers must verify that the selected class matches the command.

## Inspect and validate

```bash
alfred config validate
alfred config validate --output json
alfred config explain
alfred config explain --output json
```

`config explain` shows the merged typed model and every loaded file. It retains
environment references such as `${TOKEN}` and does not print resolved values.
