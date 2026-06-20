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

## MVP

The first public version contains:

- `alfred init` to create an `alfred.yaml`;
- `alfred config validate` to validate and explain merged configuration;
- `alfred doctor` to inspect required tools, files, and project conditions;
- `alfred commit` for Conventional Commits and optional push confirmation;
- `alfred docker up|down` for configured Docker Compose projects;
- `alfred run <workflow>` for explicitly declared local workflows;
- local preset inheritance through `extends`.

## Out Of Scope For V1

- remote presets downloaded automatically;
- binary or in-process plugin systems;
- user and role authorization;
- secrets management;
- AI commit providers;
- opinionated merge or release flows;
- Docker image building APIs beyond invoking Docker Compose;
- a full-screen TUI;
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
