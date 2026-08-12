# Release Process

Alfred releases are immutable, versioned GitHub releases created from tags.

## Preconditions

- the worktree is clean and the intended commit is on the protected default
  branch;
- `go test -race ./...`, `go vet ./...`, `go mod verify`, and `govulncheck`
  pass;
- GitHub Actions workflows pass `actionlint`;
- examples pass `alfred config validate`;
- `CHANGELOG.md` has a non-empty `## vMAJOR.MINOR.PATCH` section;
- the release tag follows exactly `vMAJOR.MINOR.PATCH`.

Use [Pre-release Checklist](pre-release-checklist.md) for the complete review.

## Create a release

```bash
git tag v0.2.0
git push origin v0.2.0
```

The release workflow verifies that the tag points to the checked-out commit. It
refuses to replace an existing GitHub release; corrections require a new patch
version.

## Produced targets

- Windows AMD64 and ARM64 ZIP archives;
- Linux AMD64 and ARM64 tarballs;
- macOS AMD64 and ARM64 tarballs;
- SHA-256 `checksums.txt`;
- versioned `install.ps1` and `install.sh`;
- one CycloneDX JSON SBOM (`*.cdx.json`) per target;
- release notes extracted from the exact changelog section.

Go builds use `CGO_ENABLED=0`, `-trimpath`, and stripped release symbols. The
version is injected through `ldflags` and checked on the Linux AMD64 binary
before publication.

## Provenance and smoke tests

The workflow uses GitHub's OIDC-backed `actions/attest` action, pinned to a
reviewed full commit SHA, to create Sigstore build-provenance attestations for
archives, SBOMs, checksums, and installers. The release job has only the write
permissions required for contents, attestations, OIDC, and artifact metadata.

After `gh release create`, independent Linux and Windows jobs download the
published archive and checksum, verify SHA-256, verify the attestation with
GitHub CLI, execute the extracted binary, run the published installer into a
temporary directory, and execute the installed binary.

## Verify as a consumer

```bash
sha256sum --check checksums.txt --ignore-missing
gh attestation verify alfred_v0.2.0_linux_amd64.tar.gz \
  --repo Vinicius0812/alfred-cli
```

On PowerShell:

```powershell
(Get-FileHash .\alfred_v0.2.0_windows_amd64.zip -Algorithm SHA256).Hash
gh attestation verify .\alfred_v0.2.0_windows_amd64.zip --repo Vinicius0812/alfred-cli
```

Compare the PowerShell hash with the exact entry in `checksums.txt`.

## Manual dispatch

The workflow may be started manually with an existing tag. The same tag format,
commit, tests, provenance, immutability, and smoke-test checks apply.
