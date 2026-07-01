# Release Process

Alfred releases are created from Git tags.

## Requirements

- all tests pass with `go test ./...`;
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
- `linux/amd64` as `.tar.gz`;
- `linux/arm64` as `.tar.gz`;
- `darwin/amd64` as `.tar.gz`;
- `darwin/arm64` as `.tar.gz`.

It also publishes a `checksums.txt` file with SHA-256 checksums.

## Manual release

The workflow can also be started manually from GitHub Actions with a version
input such as `v0.1.0`.

## Local smoke test

Before tagging, run:

```bash
go test ./...
go build -o bin/alfred ./cmd/alfred
./bin/alfred --version
./bin/alfred --config examples/go/alfred.yaml config validate
```

On Windows:

```powershell
go test ./...
go build -o bin\alfred.exe .\cmd\alfred
.\bin\alfred.exe --version
.\bin\alfred.exe --config examples\go\alfred.yaml config validate
```
