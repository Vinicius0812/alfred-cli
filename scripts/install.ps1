param(
    [string] $Version = $env:ALFRED_VERSION,
    [string] $Repo = $(if ($env:ALFRED_REPO) { $env:ALFRED_REPO } else { "Vinicius0812/alfred-cli" }),
    [string] $InstallDir = $(if ($env:ALFRED_INSTALL_DIR) { $env:ALFRED_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA "Programs\Alfred\bin" }),
    [switch] $NoPathUpdate
)

$ErrorActionPreference = "Stop"

function Get-AlfredArchitecture {
    switch ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLowerInvariant()) {
        "x64" { "amd64"; break }
        "arm64" { "arm64"; break }
        default { throw "Unsupported architecture: $([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture)" }
    }
}

function Assert-AlfredRepository {
    param([string] $Value)

    if ($Value -notmatch '^[A-Za-z0-9._-]+/[A-Za-z0-9._-]+$') {
        throw "Invalid repository '$Value'. Expected owner/repository."
    }
}

function Assert-AlfredVersion {
    param([string] $Value)

    if ($Value -notmatch '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$') {
        throw "Invalid Alfred version '$Value'. Expected vMAJOR.MINOR.PATCH."
    }
}

function Get-ExpectedChecksum {
    param(
        [string] $ChecksumPath,
        [string] $FileName
    )

    $matches = @(
        Get-Content -LiteralPath $ChecksumPath | Where-Object {
            $parts = $_ -split '\s+', 2
            if ($parts.Count -ne 2) {
                return $false
            }
            $listedName = $parts[1].Trim().TrimStart('*')
            return $listedName -ceq $FileName
        }
    )

    if ($matches.Count -ne 1) {
        throw "Could not find exactly one checksum for $FileName."
    }

    $expected = ($matches[0] -split '\s+', 2)[0].ToLowerInvariant()
    if ($expected -notmatch '^[a-f0-9]{64}$') {
        throw "Invalid SHA-256 checksum for $FileName."
    }
    return $expected
}

Assert-AlfredRepository -Value $Repo

if ([string]::IsNullOrWhiteSpace($Version)) {
    $latest = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
    $Version = $latest.tag_name
}

Assert-AlfredVersion -Value $Version

$arch = Get-AlfredArchitecture
$archive = "alfred_${Version}_windows_${arch}.zip"
$releaseBaseUrl = "https://github.com/$Repo/releases/download/$Version"
$archiveUrl = "$releaseBaseUrl/$archive"
$checksumUrl = "$releaseBaseUrl/checksums.txt"
$tempDir = Join-Path ([System.IO.Path]::GetTempPath()) ("alfred-install-" + [System.Guid]::NewGuid().ToString("N"))
$zipPath = Join-Path $tempDir $archive
$checksumPath = Join-Path $tempDir "checksums.txt"

[System.IO.Directory]::CreateDirectory($tempDir) | Out-Null
[System.IO.Directory]::CreateDirectory($InstallDir) | Out-Null

try {
    Invoke-WebRequest -Uri $archiveUrl -OutFile $zipPath
    Invoke-WebRequest -Uri $checksumUrl -OutFile $checksumPath

    $expectedHash = Get-ExpectedChecksum -ChecksumPath $checksumPath -FileName $archive
    $actualHash = (Get-FileHash -LiteralPath $zipPath -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($actualHash -cne $expectedHash) {
        throw "Checksum verification failed for $archive."
    }

    Write-Host "Verified SHA-256 checksum for $archive"
    Expand-Archive -LiteralPath $zipPath -DestinationPath $tempDir -Force

    $source = Join-Path $tempDir "alfred.exe"
    if (-not (Test-Path -LiteralPath $source -PathType Leaf)) {
        throw "The downloaded archive does not contain alfred.exe."
    }

    $target = Join-Path $InstallDir "alfred.exe"
    Copy-Item -LiteralPath $source -Destination $target -Force

    $normalizedInstallDir = [System.IO.Path]::GetFullPath($InstallDir).TrimEnd('\')
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $pathEntries = @()
    if (-not [string]::IsNullOrWhiteSpace($userPath)) {
        $pathEntries = @($userPath -split ";" | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
    }

    $alreadyInPath = $false
    foreach ($entry in $pathEntries) {
        try {
            $normalizedEntry = [System.IO.Path]::GetFullPath($entry).TrimEnd('\')
            if ([System.StringComparer]::OrdinalIgnoreCase.Equals($normalizedEntry, $normalizedInstallDir)) {
                $alreadyInPath = $true
                break
            }
        }
        catch {
            # Preserve unusual existing PATH entries without treating them as Alfred's directory.
        }
    }

    if (-not $NoPathUpdate -and -not $alreadyInPath) {
        $newPath = if ([string]::IsNullOrWhiteSpace($userPath)) { $InstallDir } else { "$userPath;$InstallDir" }
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        Write-Host "Added $InstallDir to your user PATH. Open a new terminal to use alfred from anywhere."
    }

    Write-Host "Alfred installed at $target"
}
finally {
    Remove-Item -LiteralPath $tempDir -Recurse -Force -ErrorAction SilentlyContinue
}
