[CmdletBinding()]
param(
    [switch]$Install
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

$repoRoot = Split-Path -Parent $PSScriptRoot
Set-Location -LiteralPath $repoRoot

$distDir = Join-Path $repoRoot "dist"
if (-not (Test-Path -LiteralPath $distDir)) {
    New-Item -ItemType Directory -Force -Path $distDir | Out-Null
}

$env:GOFLAGS = "-buildvcs=false"
$env:GOOS = "windows"

$gitSha = "none"
try {
    $gitSha = (git rev-parse HEAD).Trim()
} catch {
    $gitSha = "none"
}
$ldFlags = "-s -w -X main.version=v0.1.82 -X main.commit=$gitSha"

Write-Host "==> Building Windows amd64 binary (commit: $gitSha)"
$env:GOARCH = "amd64"
$amd64Bin = Join-Path $distDir "cc-session.exe"
go build -ldflags $ldFlags -o $amd64Bin ./cmd/cc-session
if ($LASTEXITCODE -ne 0) { throw "go build windows/amd64 failed with exit code $LASTEXITCODE" }

Write-Host "==> Packaging Windows amd64 zip"
$amd64Zip = Join-Path $distDir "cc-session-reader_0.1.82_windows_amd64.zip"
Compress-Archive -Path $amd64Bin -DestinationPath $amd64Zip -Force

Write-Host "==> Building Windows arm64 binary"
$env:GOARCH = "arm64"
$arm64Dir = Join-Path $distDir "arm64"
New-Item -ItemType Directory -Force -Path $arm64Dir | Out-Null
$arm64Bin = Join-Path $arm64Dir "cc-session.exe"
go build -ldflags $ldFlags -o $arm64Bin ./cmd/cc-session
if ($LASTEXITCODE -ne 0) { throw "go build windows/arm64 failed with exit code $LASTEXITCODE" }

Write-Host "==> Packaging Windows arm64 zip"
$arm64Zip = Join-Path $distDir "cc-session-reader_0.1.82_windows_arm64.zip"
Compress-Archive -Path $arm64Bin -DestinationPath $arm64Zip -Force

Write-Host "==> Generating checksums.txt"
$hashAmd64 = (Get-FileHash -LiteralPath $amd64Zip -Algorithm SHA256).Hash.ToLowerInvariant()
$hashArm64 = (Get-FileHash -LiteralPath $arm64Zip -Algorithm SHA256).Hash.ToLowerInvariant()

$checksumContent = "$hashAmd64  cc-session-reader_0.1.82_windows_amd64.zip`n$hashArm64  cc-session-reader_0.1.82_windows_arm64.zip`n"
$checksumFile = Join-Path $distDir "checksums.txt"
[System.IO.File]::WriteAllText($checksumFile, $checksumContent, [System.Text.Encoding]::UTF8)

if ($Install) {
    $nativeBin = switch ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture) {
        'X64' { $amd64Bin }
        'Arm64' { $arm64Bin }
        default { throw "Unsupported installation architecture: $([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture)" }
    }
    $localAppDir = Join-Path $env:LOCALAPPDATA "cc-session"
    New-Item -ItemType Directory -Force -Path $localAppDir | Out-Null
    $installedBin = Join-Path $localAppDir "cc-session.exe"
    Write-Host "==> Installing native binary to $installedBin"
    Copy-Item -LiteralPath $nativeBin -Destination $installedBin -Force -ErrorAction Stop
}

Write-Host "BUILD AND PACKAGE SUCCESSFUL"
