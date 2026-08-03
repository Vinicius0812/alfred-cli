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

### Changed

- GitHub Actions are pinned to full commit SHAs and use least-privilege job
  permissions.
- Published release assets are no longer overwritten by workflow reruns.

## v0.1.0

First planned public preview release.
