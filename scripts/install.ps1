#requires -Version 5.1
param(
    [string]$Version = "",
    [string]$Repo = "baselhusam/bareai-cli",
    [string]$InstallDir = "$env:LOCALAPPDATA\Programs\bareai",
    [switch]$AddToPath
)

$ErrorActionPreference = "Stop"

function Get-LatestVersion {
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
    return $release.tag_name
}

function Get-Arch {
    switch ($env:PROCESSOR_ARCHITECTURE) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default { throw "Unsupported architecture: $($env:PROCESSOR_ARCHITECTURE)" }
    }
}

function Verify-Checksum {
    param(
        [string]$FilePath,
        [string]$ChecksumsPath
    )
    $name = [IO.Path]::GetFileName($FilePath)
    $line = Get-Content $ChecksumsPath | Where-Object { $_ -match "\s+$name$" } | Select-Object -First 1
    if (-not $line) {
        throw "Checksum entry not found for $name"
    }
    $expected = ($line -split '\s+')[0].ToLower()
    $hash = Get-FileHash -Path $FilePath -Algorithm SHA256
    if ($hash.Hash.ToLower() -ne $expected) {
        throw "Checksum mismatch for $name"
    }
}

if (-not $Version) {
    $Version = Get-LatestVersion
}
if (-not $Version) {
    throw "Could not determine release version"
}

$ver = $Version.TrimStart('v')
$arch = Get-Arch
$archive = "bareai_${ver}_windows_${arch}.zip"
$base = "https://github.com/$Repo/releases/download/$Version"
$tmpdir = Join-Path $env:TEMP ("bareai-install-" + [guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $tmpdir -Force | Out-Null

try {
    Write-Host "Installing bareai $Version for windows/$arch..."
    $checksumsPath = Join-Path $tmpdir "checksums.txt"
    $archivePath = Join-Path $tmpdir $archive
    Invoke-WebRequest -Uri "$base/checksums.txt" -OutFile $checksumsPath
    Invoke-WebRequest -Uri "$base/$archive" -OutFile $archivePath
    Verify-Checksum -FilePath $archivePath -ChecksumsPath $checksumsPath

    Expand-Archive -Path $archivePath -DestinationPath $tmpdir -Force
    $binary = Join-Path $tmpdir "bareai.exe"
    if (-not (Test-Path $binary)) {
        throw "bareai.exe not found in archive"
    }

    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    $dest = Join-Path $InstallDir "bareai.exe"
    Copy-Item -Path $binary -Destination $dest -Force

    if ($AddToPath) {
        $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
        if ($userPath -notlike "*$InstallDir*") {
            [Environment]::SetEnvironmentVariable("Path", "$userPath;$InstallDir", "User")
            $env:Path = "$env:Path;$InstallDir"
        }
    }

    Write-Host "Installed to $dest"
    & $dest version
}
finally {
    Remove-Item -Recurse -Force $tmpdir -ErrorAction SilentlyContinue
}
