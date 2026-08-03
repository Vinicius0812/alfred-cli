# Pre-release Checklist

Use this checklist before publishing an Alfred release.

## Code

- [ ] `go test ./...` passes.
- [ ] `go vet ./...` and `go mod verify` pass.
- [ ] `govulncheck ./...` reports no reachable known vulnerabilities.
- [ ] A local binary builds successfully.
- [ ] `alfred --version` prints the expected version when built with ldflags.
- [ ] `alfred config validate` succeeds against the example configs.
- [ ] `alfred doctor` behavior is understood for examples that require missing
      tools such as Docker.

## Documentation

- [ ] `README.md` documents current commands.
- [ ] `docs/release.md` matches the actual release process.
- [ ] `CHANGELOG.md` includes a non-empty `## vMAJOR.MINOR.PATCH` section for
      the release notes.
- [ ] `LICENSE` is present.

## GitHub

- [ ] The default branch is correct in install-script URLs.
- [ ] GitHub Actions are enabled.
- [ ] The repository can create releases from tags.
- [ ] The tag follows `vMAJOR.MINOR.PATCH`, for example `v0.1.0`.
- [ ] The tag points to the exact commit intended for release.
- [ ] GitHub Actions dependencies are pinned to reviewed full commit SHAs.
- [ ] The release does not already exist; published assets are never replaced.

## Smoke test after release

Windows:

```powershell
$release = Invoke-RestMethod -Uri "https://api.github.com/repos/Vinicius0812/alfred-cli/releases/latest"
$version = $release.tag_name
$installer = Join-Path $env:TEMP "install-alfred.ps1"
Invoke-WebRequest -Uri "https://github.com/Vinicius0812/alfred-cli/releases/download/$version/install.ps1" -OutFile $installer
PowerShell -NoProfile -ExecutionPolicy Bypass -File $installer -Version $version
alfred --version
alfred --lang pt-BR help
```

Linux/macOS:

```bash
version="$(curl -fsSL https://api.github.com/repos/Vinicius0812/alfred-cli/releases/latest | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n 1)"
curl -fsSLo install-alfred.sh "https://github.com/Vinicius0812/alfred-cli/releases/download/$version/install.sh"
ALFRED_VERSION="$version" sh ./install-alfred.sh
alfred --version
alfred --lang en help
```

For both platforms, confirm that the installer prints a successful SHA-256
verification before reporting the installed binary path.
