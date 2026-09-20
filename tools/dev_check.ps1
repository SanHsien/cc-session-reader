 [CmdletBinding()]
 param(
     [switch]$Quick
 )
 
 Set-StrictMode -Version Latest
 $ErrorActionPreference = "Stop"
 
 $repoRoot = Split-Path -Parent $PSScriptRoot
 Set-Location -LiteralPath $repoRoot
 
 $env:PYTHONUTF8 = "1"
 $env:PYTHONIOENCODING = "utf-8"
 
 Write-Host "==> Verifying cc-session binary and skill structure"
 $binPath = Join-Path $env:LOCALAPPDATA "cc-session\cc-session.exe"
 if (Test-Path $binPath) {
     & $binPath --help | Out-Null
     Write-Host "  [OK] cc-session binary is runnable"
 } else {
     Write-Warning "  [SKIP] Local cc-session binary not found in $binPath"
 }
 
 $codexSkill = "skills\claude-handoff\SKILL.md"
 if (Test-Path $codexSkill) {
     Write-Host "  [OK] Codex skill file verified ($codexSkill)"
 } else {
     throw "Missing Codex skill file: $codexSkill"
 }
 
 Write-Host "==> Checking upstream baseline and releases"
 if (Get-Command python -ErrorAction SilentlyContinue) {
     python tools/check_upstream_updates.py --strict
     if ($LASTEXITCODE -ne 0) {
         throw "Upstream check failed with exit code $LASTEXITCODE"
     }
 } else {
     Write-Warning "Python not found, skipping check_upstream_updates.py"
 }
 
 Write-Host "WINDOWS DEV CHECK GREEN"
