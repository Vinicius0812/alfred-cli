# Alfred Configuration v1

## Goals

`alfred.yaml` describes project behavior, not Alfred's internal implementation.
Version 1 intentionally exposes only:

```text
version
extends
project
commit
docker
workflows
policies
```

Unknown fields are errors. This catches typos and keeps migrations explicit.

## Discovery

Alfred searches for `alfred.yaml` in the current directory and then each parent
directory until it reaches the filesystem root.

The first file found is the project configuration. Users may override discovery:

```bash
alfred --config path/to/custom.yaml <command>
```

Relative paths are resolved from the file in which they are declared.

## Configuration Layers

Lowest to highest precedence:

1. Alfred defaults;
2. global user configuration;
3. files in `extends`, in listed order;
4. project `alfred.yaml`;
5. supported `ALFRED_*` environment variables;
6. command-line flags.

The global configuration path follows the operating system convention:

```text
Linux:   $XDG_CONFIG_HOME/alfred/config.yaml
macOS:   ~/Library/Application Support/alfred/config.yaml
Windows: %AppData%\alfred\config.yaml
```

Presets provide reusable defaults, not tamper-proof policy enforcement. Because
the project file has higher precedence, organizations that require mandatory
rules must validate the resolved Alfred configuration in CI as well.

## Merge Rules

- scalar values replace earlier values;
- maps merge recursively;
- lists replace earlier lists;
- workflows merge by workflow name;
- a project may replace an inherited workflow by declaring the same name;
- `null` is not used to delete inherited values in v1.

Replacement for lists is deliberate: appending security-sensitive policy lists
can produce surprising results.

## Top-Level Fields

### `version`

Required integer. The only accepted value in this specification is `1`.
Every project configuration and extended preset must declare its own version.

### `extends`

Optional list of local YAML file paths.

```yaml
extends:
  - ./config/alfred.company.yaml
  - ./config/alfred.team.yaml
```

Rules:

- remote URLs are rejected in v1;
- cycles are errors;
- missing files are errors;
- an extended file may extend another local file;
- a file may be loaded only once in a resolved configuration graph.
- extended files may be partial and do not need to declare `project`;
- the final merged configuration must satisfy all required fields.

### `project`

```yaml
project:
  name: example-api
```

Fields:

| Field | Type | Required | Meaning |
| --- | --- | --- | --- |
| `name` | string | final config | Human-readable project identifier |

The project root is the directory containing the discovered `alfred.yaml`; it
is not configurable in v1.

### `commit`

```yaml
commit:
  enabled: true
  convention: conventional
  allowed_types:
    - feat
    - fix
    - docs
    - refactor
    - test
    - chore
  scopes:
    - api
    - web
  protected_branches:
    - main
  require_scope: false
  emoji: false
  confirm_commit: true
  confirm_push: true
```

Fields:

| Field | Type | Default |
| --- | --- | --- |
| `enabled` | boolean | `true` |
| `convention` | `conventional` | `conventional` |
| `allowed_types` | list of strings | Conventional Commit defaults |
| `scopes` | list of strings | unrestricted |
| `protected_branches` | list of strings | `[main, master]` |
| `require_scope` | boolean | `false` |
| `emoji` | boolean | `false` |
| `confirm_commit` | boolean | `true` |
| `confirm_push` | boolean | `true` |

In v1, protected branches block commits. There is no user-based bypass in the
project file. A future policy mechanism may introduce explicit exceptions.

### `docker`

```yaml
docker:
  enabled: true
  compose_files:
    - compose.yaml
  profiles:
    - development
  env_files:
    - .env
  project_name: example-api
```

Fields:

| Field | Type | Default |
| --- | --- | --- |
| `enabled` | boolean | `true` |
| `compose_files` | list of paths | auto-detect Compose default |
| `profiles` | list of strings | empty |
| `env_files` | list of paths | empty |
| `project_name` | string | unset |

Alfred delegates to `docker compose`. It does not run Git operations as a side
effect of Docker commands.

### `workflows`

Workflows are a map keyed by command name:

```yaml
workflows:
  setup:
    description: Prepare the local environment
    steps:
      - run: docker compose build
      - run: docker compose up -d
        confirm: true
```

A workflow contains:

| Field | Type | Required |
| --- | --- | --- |
| `description` | string | yes |
| `working_directory` | path | no |
| `environment` | map of strings | no |
| `steps` | list of step objects | yes |

A step contains:

| Field | Type | Default |
| --- | --- | --- |
| `name` | string | command text |
| `run` | string | required |
| `working_directory` | path | workflow/project directory |
| `environment` | map of strings | empty |
| `confirm` | boolean | `false` |
| `continue_on_error` | boolean | `false` |

Commands run through the platform shell in v1. Alfred prints the command and
working directory before execution. Workflow and step environments merge, with
the step taking precedence.

Environment values may reference existing variables:

```yaml
environment:
  REGISTRY_TOKEN: ${REGISTRY_TOKEN}
```

An unresolved variable is an error. Alfred never writes resolved secret values
to diagnostic output.

### `policies`

```yaml
policies:
  require_clean_worktree: true
  confirm_destructive_actions: true
```

Fields:

| Field | Type | Default |
| --- | --- | --- |
| `require_clean_worktree` | boolean | `false` |
| `confirm_destructive_actions` | boolean | `true` |

`require_clean_worktree` applies to Alfred operations that change Git state. It
does not prevent read-only commands such as `doctor` or `config validate`.

Destructive built-in operations always honor `confirm_destructive_actions`.
Workflow commands are user-defined and require explicit `confirm: true` when
their authors want Alfred to prompt.

## Validation

`alfred config validate` must detect:

- unsupported schema versions;
- unknown fields;
- invalid enum values and types;
- empty required strings or lists;
- missing or cyclic `extends`;
- duplicate or invalid workflow names;
- steps without `run`;
- paths escaping the project root where a command requires project-local files;
- unresolved environment references;
- commit scopes required but not configured.

Validation errors identify the source file and field path:

```text
alfred.yaml: commit.allowed_types[2]: value must not be empty
```

## Example

```yaml
version: 1

project:
  name: example-api

commit:
  convention: conventional
  allowed_types: [feat, fix, docs, refactor, test, chore]
  protected_branches: [main]
  confirm_push: true

docker:
  compose_files:
    - compose.yaml

workflows:
  test:
    description: Run tests
    steps:
      - run: go test ./...

policies:
  require_clean_worktree: true
  confirm_destructive_actions: true
```
