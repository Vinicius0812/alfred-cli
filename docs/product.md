# Alfred Product Scope

## Product Statement

Alfred is an open source CLI for developers who want one predictable interface
for project setup, environment diagnostics, standardized commits, Docker
Compose, and repeatable local workflows.

It is useful without company configuration and extensible enough for teams to
share policies without maintaining a fork.

## Design Principles

1. **Useful by default**: common developer workflows should require little or
   no configuration.
2. **Explicit execution**: Alfred does not pull, push, delete, prune, or run a
   workflow unless the user asks it to.
3. **Configuration over forks**: organization rules belong in presets and
   project configuration.
4. **Portable installation**: users receive a native binary and do not need a
   language runtime.
5. **Inspectable behavior**: commands can explain what they will execute, and
   destructive actions require confirmation.
6. **Stable contracts**: configuration is versioned and validated before use.
7. **Secrets stay external**: configuration references environment variables;
   it does not store credentials.

## Target Users

- individual developers working across multiple stacks;
- teams standardizing local setup and commit practices;
- companies sharing development policies across repositories;
- maintainers who want repeatable project commands without bespoke shell docs.

## Implemented secure automation baseline

The `v0.2.0` scope contains:

- `alfred init` to create an `alfred.yaml`;
- `alfred config validate` and `config explain` for merged configuration;
- `alfred doctor` to inspect required tools, files, and project conditions;
- `alfred workflow list` and `alfred run <workflow>` with planning, direct
  process execution, risk-based confirmation, cancellation, and auditing;
- `alfred run list|show` for local audit inspection;
- local preset inheritance through `extends`, project-root restricted by
  default;
- the `go-secure` development preset;
- English and Brazilian Portuguese interaction and shell completions.

The typed `commit` and `docker` sections preserve the intended public contract,
but their command adapters remain roadmap items after the secure workflow
engine has been exercised in real projects.

## AI-assisted commits

Alfred's `commit` command should support generating Conventional Commit
messages from the current Git diff, inspired by the previous UNI CLI workflow.

Because Alfred is open source and intended for broad community use, AI support
must be optional and provider-agnostic:

- users can write commit messages manually without configuring any AI provider;
- AI-generated messages are proposed for review before commit execution;
- Alfred must not send diffs to a remote provider unless the user explicitly
  enables that behavior;
- credentials must come from environment variables or user-level configuration,
  never from project files committed to a repository;
- provider selection should be configurable so free or local options can be
  evaluated before choosing a default recommendation;
- OpenRouter can be supported as one provider, but must not be hardcoded as the
  only path.

When this feature is implemented, evaluate currently available free or
low-friction options for open source users, including:

- local models through tools such as Ollama or compatible local runtimes;
- free-tier hosted APIs where terms allow this use case;
- OpenRouter free models, when available;
- bring-your-own-key providers compatible with an OpenAI-style API.

The exact provider list should be researched at implementation time because
free tiers, model availability, and terms change frequently.

## Out of scope for the current release

- remote presets downloaded automatically;
- binary or in-process plugin systems;
- user and role authorization;
- secrets management;
- opinionated merge or release flows;
- Docker image building APIs beyond invoking Docker Compose;
- a full-screen TUI or plugin runtime;
- replacing Git, Docker, or project package managers.

These features may be added after real usage demonstrates the need.

## Core And Organization Boundary

The core owns:

- configuration discovery, parsing, merging, and validation;
- command execution and output;
- Git and Docker Compose adapters;
- workflow orchestration;
- safety prompts and dry-run support.

Organizations own:

- protected branches;
- allowed commit types and scopes;
- required project checks;
- Docker Compose files and profiles;
- reusable workflows;
- environment-specific conventions.

UNI CLI should be treated as a source of proven workflows, not as the public
configuration model. UNI-specific paths, repositories, users, and branch names
must not be hardcoded into Alfred.

## Initial Success Criteria

- installation requires only one Alfred binary;
- a new user can configure a project in under five minutes;
- all configured behavior can be inspected before execution;
- the same schema supports Go, Node.js, PHP, and company presets;
- invalid configuration fails with a field-specific, actionable message;
- core commands behave consistently on Windows, Linux, and macOS.
