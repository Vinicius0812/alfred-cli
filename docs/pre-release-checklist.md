# Pre-release Checklist

Use this checklist before publishing an Alfred release.

## Code

- [ ] `go test ./...` passes.
- [ ] A local binary builds successfully.
- [ ] `alfred --version` prints the expected version when built with ldflags.
- [ ] `alfred config validate` succeeds against the example configs.
- [ ] `alfred doctor` behavior is understood for examples that require missing
      tools such as Docker.

## Documentation

- [ ] `README.md` documents current commands.
- [ ] `docs/release.md` matches the actual release process.
- [ ] `CHANGELOG.md` includes the release version.
- [ ] `LICENSE` is present.

## GitHub

- [ ] The default branch is correct in install-script URLs.
- [ ] GitHub Actions are enabled.
- [ ] The repository can create releases from tags.
- [ ] The tag follows `vMAJOR.MINOR.PATCH`, for example `v0.1.0`.

## Smoke test after release

Windows:

```powershell
iwr https://raw.githubusercontent.com/Vinicius0812/alfred-cli/master/scripts/install.ps1 -UseB | iex
alfred --version
alfred --lang pt-BR help
```

Linux/macOS:

```bash
curl -fsSL https://raw.githubusercontent.com/Vinicius0812/alfred-cli/master/scripts/install.sh | sh
alfred --version
alfred --lang en help
```
