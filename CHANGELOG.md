# Changelog

All notable changes to Alfred CLI will be documented in this file.

The format follows a simple release-based structure.

## Unreleased

### Added

- Initial Go CLI entrypoint.
- `alfred config validate` for validating `alfred.yaml`.
- Local `extends` support for configuration presets.
- `alfred init` for generating an initial project configuration.
- `alfred doctor` for basic local environment checks.
- `alfred menu` as an initial text-based interactive menu.
- English and Brazilian Portuguese CLI messages via `--lang` and `ALFRED_LANG`.
- GitHub Actions release workflow for Windows, Linux, and macOS binaries.
- Windows PowerShell and Linux/macOS shell installer scripts.
- Product guidance for future optional, provider-agnostic AI commit generation.
- Pull request and push CI with formatting, tests, race detection, static
  analysis, workflow linting, build verification, and Go vulnerability
  scanning.
- Dependabot configuration for Go modules and GitHub Actions.
- Strict release version and tag validation.
- SHA-256 verification in Windows, Linux, and macOS installers.
- Windows ARM64 release artifacts.
- Versioned installer assets in GitHub releases.
- A security policy covering support, reporting, and trust boundaries.
- Typed configuration models and strict workflow step contracts.
- YAML duplicate-key, alias, custom-tag, merge-key, size, depth, and node
  guardrails, plus a bounded local inheritance graph.
- Project-root and symbolic-link containment for configuration and execution
  paths.
- Structured `command` and `args` workflow execution without a shell by
  default.
- Explicit, policy-gated shell steps for exceptional compatibility needs.
- Risk classification, interactive confirmation, clean-worktree checks,
  timeouts, Ctrl+C cancellation, and controlled non-interactive confirmation.
- Dry-run planning and stable text or JSON output.
- Streaming secret redaction and atomic local audit records with `run list` and
  `run show` inspection.
- `config explain`, `workflow list`, and Bash, Zsh, and PowerShell completion.
- `basic` and `go-secure` initialization presets.
- Alfred's bat identity and bilingual interactive dialogue.
- A documented interface identity contract that keeps Alfred's personality out
  of stable JSON and automation output.
- CycloneDX SBOMs for every release target, Sigstore/GitHub build-provenance
  attestations, and post-publication Linux and Windows smoke tests.

### Changed

- GitHub Actions are pinned to full commit SHAs and use least-privilege job
  permissions.
- Published release assets are no longer overwritten by workflow reruns.
- Workflow examples now use direct structured execution and explicit risk
  classes.
- Configuration errors and missing-file guidance are localized for Portuguese.
- Sensitive inherited environment values, including access keys, are redacted
  from subprocess output and audit diagnostics.
- `alfred init` derives the project name from the current directory and honors
  an explicit global `--config` output path.
- Parent or user cancellation always stops execution, including steps marked
  with `continue_on_error`.

## v0.1.0

First planned public preview release.
