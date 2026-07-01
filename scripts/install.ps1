param(
    [string] $Version = $env:ALFRED_VERSION,
    [string] $Repo = $(if ($env:ALFRED_REPO) { $env:ALFRED_REPO } else { "Vinicius0812/alfred-cli" }),
    [string] $InstallDir = $(if ($env:ALFRED_INSTALL_DIR) { $env:ALFRED_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA "Programs\Alfred\bin" })
)

$ErrorActionPreference = "Stop"

function Get-AlfredArchitecture {
    switch ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLowerInvariant()) {
        "x64" { "amd64"; break }
        "arm64" { "arm64"; break }
        default { throw "Unsupported architecture: $([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture)" }
    }
}

if ([string]::IsNullOrWhiteSpace($Version)) {
    $latest = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
    $Version = $latest.tag_name
}

if ([string]::IsNullOrWhiteSpace($Version)) {
    throw "Could not determine Alfred version."
}

$arch = Get-AlfredArchitecture
$archive = "alfred_${Version}_windows_${arch}.zip"
$url = "https://github.com/$Repo/releases/download/$Version/$archive"
$tempDir = Join-Path ([System.IO.Path]::GetTempPath()) ("alfred-install-" + [System.Guid]::NewGuid().ToString("N"))
$zipPath = Join-Path $tempDir $archive

New-Item -ItemType Directory -Path $tempDir -Force | Out-Null
New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null

try {
    Invoke-WebRequest -Uri $url -OutFile $zipPath
    Expand-Archive -Path $zipPath -DestinationPath $tempDir -Force

    $source = Join-Path $tempDir "alfred.exe"
    $target = Join-Path $InstallDir "alfred.exe"
    Copy-Item -Path $source -Destination $target -Force

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $pathEntries = @()
    if (-not [string]::IsNullOrWhiteSpace($userPath)) {
        $pathEntries = $userPath -split ";"
    }

    if ($pathEntries -notcontains $InstallDir) {
        $newPath = if ([string]::IsNullOrWhiteSpace($userPath)) { $InstallDir } else { "$userPath;$InstallDir" }
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        Write-Host "Added $InstallDir to your user PATH. Open a new terminal to use alfred from anywhere."
    }

    Write-Host "Alfred installed at $target"
}
finally {
    Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
}
