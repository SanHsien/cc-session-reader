[CmdletBinding()]
param()

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

$repoRoot = Split-Path -Parent $PSScriptRoot
Set-Location -LiteralPath $repoRoot

$distDir = Join-Path $repoRoot "dist"
if (-not (Test-Path -LiteralPath $distDir)) {
    New-Item -ItemType Directory -Force -Path $distDir | Out-Null
}

$localAppDir = Join-Path $env:LOCALAPPDATA "cc-session"
if (-not (Test-Path -LiteralPath $localAppDir)) {
    New-Item -ItemType Directory -Force -Path $localAppDir | Out-Null
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

Write-Host "==> Installing amd64 binary to LocalAppData"
$installedBin = Join-Path $localAppDir "cc-session.exe"
try {
    Copy-Item -LiteralPath $amd64Bin -Destination $installedBin -Force -ErrorAction Stop
    Write-Host "  [OK] Installed to $installedBin"
} catch {
    Write-Warning "  [SKIP] Could not write to $installedBin in sandbox: $_"
}

Write-Host "==> Packaging Windows amd64 zip"
$amd64Zip = Join-Path $distDir "cc-session-reader_0.1.82_windows_amd64.zip"
if (Test-Path -LiteralPath $amd64Zip) {
    Remove-Item -LiteralPath $amd64Zip -Force
}
Compress-Archive -Path $amd64Bin -DestinationPath $amd64Zip -Force

Write-Host "==> Building Windows arm64 binary"
$env:GOARCH = "arm64"
$arm64Bin = Join-Path $distDir "cc-session_arm64.exe"
go build -ldflags $ldFlags -o $arm64Bin ./cmd/cc-session

Write-Host "==> Packaging Windows arm64 zip"
$arm64Temp = Join-Path $distDir "arm64_temp"
if (Test-Path -LiteralPath $arm64Temp) {
    Remove-Item -Recurse -Force -LiteralPath $arm64Temp
}
New-Item -ItemType Directory -Force -Path $arm64Temp | Out-Null
$arm64TempBin = Join-Path $arm64Temp "cc-session.exe"
Copy-Item -LiteralPath $arm64Bin -Destination $arm64TempBin -Force

$arm64Zip = Join-Path $distDir "cc-session-reader_0.1.82_windows_arm64.zip"
if (Test-Path -LiteralPath $arm64Zip) {
    Remove-Item -LiteralPath $arm64Zip -Force
}
Compress-Archive -Path $arm64TempBin -DestinationPath $arm64Zip -Force

Remove-Item -Recurse -Force -LiteralPath $arm64Temp
Remove-Item -Force -LiteralPath $arm64Bin

Write-Host "==> Generating checksums.txt"
$hashAmd64 = (Get-FileHash -LiteralPath $amd64Zip -Algorithm SHA256).Hash.ToLowerInvariant()
$hashArm64 = (Get-FileHash -LiteralPath $arm64Zip -Algorithm SHA256).Hash.ToLowerInvariant()

$checksumContent = "$hashAmd64  cc-session-reader_0.1.82_windows_amd64.zip`n$hashArm64  cc-session-reader_0.1.82_windows_arm64.zip`n"
$checksumFile = Join-Path $distDir "checksums.txt"
[System.IO.File]::WriteAllText($checksumFile, $checksumContent, [System.Text.Encoding]::UTF8)

Write-Host "BUILD AND PACKAGE SUCCESSFUL"
