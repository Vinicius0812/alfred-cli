$ErrorActionPreference = "Stop"

$testRoot = Join-Path ([System.IO.Path]::GetTempPath()) ("alfred-installer-test-" + [System.Guid]::NewGuid().ToString("N"))
$fixtureDir = Join-Path $testRoot "fixture"
$fixtureArchive = Join-Path $testRoot "fixture.zip"
$installedDir = Join-Path $testRoot "installed"
$rejectedDir = Join-Path $testRoot "rejected"
$archiveName = "alfred_v9.8.7_windows_amd64.zip"

[System.IO.Directory]::CreateDirectory($fixtureDir) | Out-Null
Set-Content -LiteralPath (Join-Path $fixtureDir "alfred.exe") -Value "safe fixture" -NoNewline
Compress-Archive -LiteralPath (Join-Path $fixtureDir "alfred.exe") -DestinationPath $fixtureArchive
$fixtureHash = (Get-FileHash -LiteralPath $fixtureArchive -Algorithm SHA256).Hash.ToLowerInvariant()
$global:AlfredInstallerTestBadChecksum = $false
$global:AlfredInstallerTestFixtureHash = $fixtureHash
$global:AlfredInstallerTestFixtureArchive = $fixtureArchive
$global:AlfredInstallerTestArchiveName = $archiveName

function global:Invoke-WebRequest {
    param(
        [string] $Uri,
        [string] $OutFile
    )

    if ($Uri.EndsWith("/checksums.txt")) {
        $hash = if ($global:AlfredInstallerTestBadChecksum) { "0" * 64 } else { $global:AlfredInstallerTestFixtureHash }
        Set-Content -LiteralPath $OutFile -Value "$hash  $global:AlfredInstallerTestArchiveName"
        return
    }

    if ($Uri.EndsWith("/$global:AlfredInstallerTestArchiveName")) {
        Copy-Item -LiteralPath $global:AlfredInstallerTestFixtureArchive -Destination $OutFile
        return
    }

    throw "Unexpected test download: $Uri"
}

try {
    & (Join-Path $PSScriptRoot "install.ps1") `
        -Version "v9.8.7" `
        -Repo "example/alfred" `
        -InstallDir $installedDir `
        -NoPathUpdate

    $installedBinary = Join-Path $installedDir "alfred.exe"
    if (-not (Test-Path -LiteralPath $installedBinary -PathType Leaf)) {
        throw "Installer did not copy the verified binary."
    }
    if ((Get-Content -LiteralPath $installedBinary -Raw) -ne "safe fixture") {
        throw "Installed binary does not match the verified fixture."
    }

    $global:AlfredInstallerTestBadChecksum = $true
    $rejected = $false
    try {
        & (Join-Path $PSScriptRoot "install.ps1") `
            -Version "v9.8.7" `
            -Repo "example/alfred" `
            -InstallDir $rejectedDir `
            -NoPathUpdate
    }
    catch {
        if ($_.Exception.Message -notlike "Checksum verification failed*") {
            throw
        }
        $rejected = $true
    }

    if (-not $rejected) {
        throw "Installer accepted an archive with an invalid checksum."
    }
    if (Test-Path -LiteralPath (Join-Path $rejectedDir "alfred.exe")) {
        throw "Installer copied a binary after checksum verification failed."
    }

    Write-Host "PowerShell installer tests passed"
}
finally {
    Remove-Item Function:\Invoke-WebRequest -ErrorAction SilentlyContinue
    Remove-Variable AlfredInstallerTestBadChecksum -Scope Global -ErrorAction SilentlyContinue
    Remove-Variable AlfredInstallerTestFixtureHash -Scope Global -ErrorAction SilentlyContinue
    Remove-Variable AlfredInstallerTestFixtureArchive -Scope Global -ErrorAction SilentlyContinue
    Remove-Variable AlfredInstallerTestArchiveName -Scope Global -ErrorAction SilentlyContinue
    Remove-Item -LiteralPath $testRoot -Recurse -Force -ErrorAction SilentlyContinue
}
