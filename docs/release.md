# Release Process

Alfred releases are created from Git tags.

## Requirements

- all tests pass with `go test ./...`;
- `go vet ./...`, `go mod verify`, and `govulncheck ./...` pass;
- the release version follows a `vMAJOR.MINOR.PATCH` format, for example
  `v0.1.0`;
- `CHANGELOG.md` has an entry for the release;
- `LICENSE` is present;
- the repository has a clean working tree before tagging.

## Create a release

```bash
git tag v0.1.0
git push origin v0.1.0
```

The GitHub Actions release workflow builds:

- `windows/amd64` as `.zip`;
- `windows/arm64` as `.zip`;
- `linux/amd64` as `.tar.gz`;
- `linux/arm64` as `.tar.gz`;
- `darwin/amd64` as `.tar.gz`;
- `darwin/arm64` as `.tar.gz`.

It also publishes versioned installer scripts and a `checksums.txt` file with
SHA-256 checksums for every archive and installer. The installers verify the
archive checksum before installation.

Published release assets are immutable from the workflow's perspective. If a
release already exists for a tag, the workflow fails instead of replacing its
assets. Publish corrections under a new patch version.

Release notes are extracted from the matching `## vMAJOR.MINOR.PATCH` section
in `CHANGELOG.md`. The workflow fails if that section is absent or empty.

## Manual release

The workflow can also be started manually from GitHub Actions with a version
input such as `v0.1.0`.

The input must name an existing tag, and that tag must point to the commit being
built. Pre-release and build suffixes are not accepted by the current workflow.

## Local smoke test

Before tagging, run:

```bash
go test ./...
go vet ./...
go mod verify
go build -o bin/alfred ./cmd/alfred
./bin/alfred --version
./bin/alfred --config examples/go/alfred.yaml config validate
```

On Windows:

```powershell
go test ./...
go vet ./...
go mod verify
go build -o bin\alfred.exe .\cmd\alfred
.\bin\alfred.exe --version
.\bin\alfred.exe --config examples\go\alfred.yaml config validate
```
