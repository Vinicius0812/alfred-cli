# Alfred Release Roadmap

This roadmap defines product contracts, not calendar dates. Security and
compatibility gates take precedence over shipping a numbered release.

## v0.2.0 — secure workflow engine

The v0.2.0 release candidate is the first version intended to execute project
workflows. Its required scope is:

- strict typed configuration and safe local inheritance;
- YAML resource and ambiguity guardrails;
- project-root and symlink containment;
- structured command execution without a shell by default;
- exceptional shell mode requiring schema, policy, and user opt-in;
- risk classes, policy-driven confirmation, clean-worktree protection, and a
  controlled `--yes` path;
- dry-run planning, timeouts, Ctrl+C cancellation, streaming output, and secret
  masking;
- local atomic audit records and text/JSON inspection;
- basic and `go-secure` presets;
- bilingual commands, interactive Alfred identity, and shell completions;
- checksums, per-target CycloneDX SBOMs, provenance attestations, and published
  binary/installer smoke tests;
- complete configuration, command, workflow, architecture, security, and
  release documentation.

Release gates still external to a local checkout are a green GitHub Actions run
for the release commit, protected-default-branch policy, and successful
post-publication jobs for the tagged assets.

## v0.3.x — commit workflow

- manual Conventional Commit construction remains available without AI;
- diff selection and sensitive-file filtering occur before provider use;
- protected-branch and clean-worktree policies are enforced;
- commit and push are distinct confirmed operations;
- optional AI suggestions are provider-agnostic and review-before-use;
- local models are supported without requiring a hosted provider;
- remote diff transmission requires explicit user configuration and consent;
- credentials remain in user environment or user-level storage, never project
  YAML or audit records.

Hosted free tiers and model availability must be researched at implementation
time because their limits and terms change.

## v0.4.x — Docker Compose adapter

- typed `docker compose` planning from the existing configuration contract;
- direct executable/argument invocation through the workflow guardrails;
- explicit project, profile, Compose file, and environment-file selection;
- risk classification for build, pull, up, down, and destructive cleanup;
- dry-run, confirmation, timeout, redaction, and audit parity with workflows.

## v0.5.x and later previews

- user feedback-driven schema additions with migrations;
- richer doctor checks and machine-readable diagnostics;
- auditable preset provenance without automatic remote execution;
- packaging integrations only when they preserve checksum and attestation
  verification;
- accessibility and terminal compatibility improvements to the interactive
  experience.

## v1.0.0 — stable defensive automation contract

Version 1.0.0 should be published only after the following are demonstrated in
real projects on Windows, Linux, and macOS:

- stable configuration v1, CLI command names, JSON fields, outcome values, and
  exit codes;
- documented migration and deprecation policy;
- workflow, commit, and Docker adapters share the same planning, confirmation,
  execution, redaction, and auditing model;
- no remote preset execution or remote AI data transfer without explicit trust
  decisions;
- cross-platform integration tests and post-release installation tests;
- protected release branch, pinned Actions, dependency scanning, per-target
  SBOMs, checksums, and verifiable provenance;
- threat model, security reporting process, and recovery procedure for a
  compromised release;
- documentation sufficient for users and contributors to reproduce every
  supported operation.

Offensive execution, persistence, credential collection, evasion, and covert
remote control remain outside Alfred's product scope.
