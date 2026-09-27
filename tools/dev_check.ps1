[CmdletBinding()]
param(
    [switch]$Quick
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$repoRoot = Split-Path -Parent $PSScriptRoot
Set-Location -LiteralPath $repoRoot
$env:PYTHONUTF8 = '1'
$env:PYTHONIOENCODING = 'utf-8'

function Assert-Utf8File {
    param([string]$Path)

    $strictUtf8 = [System.Text.UTF8Encoding]::new($false, $true)
    $resolved = (Resolve-Path -LiteralPath $Path).Path
    $text = $strictUtf8.GetString([System.IO.File]::ReadAllBytes($resolved))
    if ($text.Contains([char]0) -or $text.Contains([char]0xFFFD)) {
        throw "Invalid or corrupted text detected in $Path"
    }
    return $text
}

function Get-SkillName {
    param([string]$Path)

    $text = Assert-Utf8File -Path $Path
    if ($text -notmatch '(?ms)\A---\s*\r?\n.*?^name:\s*([^\r\n]+)\s*$.*?^---\s*$') {
        throw "Missing valid YAML frontmatter name in $Path"
    }
    return $Matches[1].Trim()
}

Write-Host '==> Checking PowerShell syntax'
$parseTokens = $null
$parseErrors = $null
[System.Management.Automation.Language.Parser]::ParseFile(
    (Resolve-Path -LiteralPath 'install.ps1').Path,
    [ref]$parseTokens,
    [ref]$parseErrors
) | Out-Null
if ($parseErrors.Count -gt 0) {
    throw ($parseErrors | ForEach-Object Message | Out-String)
}
Write-Host '  [OK] install.ps1 parses'

Write-Host '==> Checking canonical Skills'
$agentSkill = 'skills\agent-handoff\SKILL.md'
$agentName = Get-SkillName -Path $agentSkill
if ($agentName -ne 'agent-handoff') {
    throw "Skill name mismatch: expected agent-handoff, found $agentName"
}
if (-not (Test-Path -LiteralPath 'skills\agent-handoff\agents\openai.yaml')) {
    throw 'Missing skills\agent-handoff\agents\openai.yaml'
}
$duplicateSkill = Get-ChildItem -LiteralPath skills -Recurse -Filter SKILL.md |
    Where-Object { $_.FullName -ne (Resolve-Path -LiteralPath $agentSkill).Path } |
    Where-Object { (Get-SkillName -Path $_.FullName) -eq $agentName }
if ($duplicateSkill) {
    throw "Duplicate Skill name '$agentName': $($duplicateSkill.FullName -join ', ')"
}
Write-Host '  [OK] agent-handoff is unique and folder/name aligned'

Write-Host '==> Checking bilingual public documentation'
$pairs = @(
    @('README.md', 'README.en.md'),
    @('SKILL.md', 'SKILL.en.md'),
    @('docs\benchmark.md', 'docs\benchmark.zh-TW.md'),
    @('docs\adr-001-collapse-cc-session-calls.md', 'docs\adr-001-collapse-cc-session-calls.zh-TW.md'),
    @('docs\adr-002-tool-compression-optimization.md', 'docs\adr-002-tool-compression-optimization.zh-TW.md'),
    @('docs\adr-003-tool-result-status-and-summaries.md', 'docs\adr-003-tool-result-status-and-summaries.zh-TW.md'),
    @('docs\adr-004-failure-retention-strategy.md', 'docs\adr-004-failure-retention-strategy.zh-TW.md'),
    @('docs\adr-005-collapse-retry-loops-and-reads.md', 'docs\adr-005-collapse-retry-loops-and-reads.zh-TW.md'),
    @('docs\adr-007-format-changes-measured.md', 'docs\adr-007-format-changes-measured.en.md'),
    @('docs\adr-008-harness-event-drift.md', 'docs\adr-008-harness-event-drift.en.md'),
    @('docs\adr-009-prompt-source-field.md', 'docs\adr-009-prompt-source-field.en.md')
)
foreach ($pair in $pairs) {
    foreach ($path in $pair) {
        if (-not (Test-Path -LiteralPath $path)) { throw "Missing bilingual document: $path" }
        [void](Assert-Utf8File -Path $path)
    }
}
Write-Host "  [OK] $($pairs.Count) public document pairs are present and valid UTF-8"

Write-Host '==> Checking Windows-only automation'
$workflowText = Get-ChildItem -LiteralPath '.github\workflows' -File |
    ForEach-Object { Get-Content -LiteralPath $_.FullName -Raw -Encoding UTF8 }
if (($workflowText -join "`n") -match 'runs-on:\s*(ubuntu|macos)') {
    throw 'A non-Windows GitHub Actions runner remains configured.'
}
$releaseConfig = Get-Content -LiteralPath '.goreleaser.yaml' -Raw -Encoding UTF8
if ($releaseConfig -notmatch '(?m)^\s*- windows\s*$' -or $releaseConfig -match '(?m)^\s*- (linux|darwin)\s*$') {
    throw '.goreleaser.yaml must build Windows only.'
}
Write-Host '  [OK] workflows and release configuration are Windows-only'

Write-Host '==> Verifying local binary'
$binPath = Join-Path $env:LOCALAPPDATA 'cc-session\cc-session.exe'
if (Test-Path -LiteralPath $binPath) {
    & $binPath --help | Out-Null
    if ($LASTEXITCODE -ne 0) { throw "cc-session --help failed with exit code $LASTEXITCODE" }
    Write-Host '  [OK] cc-session binary is runnable'
} else {
    Write-Warning "  [SKIP] Local cc-session binary not found in $binPath"
}

if (-not $Quick) {
    Write-Host '==> Running Go build and tests'
    $go = Get-Command go -ErrorAction SilentlyContinue
    if (-not $go) { throw 'Go is required for the full development gate.' }
    & $go.Source test ./...
    if ($LASTEXITCODE -ne 0) { throw "go test failed with exit code $LASTEXITCODE" }
    & $go.Source build ./...
    if ($LASTEXITCODE -ne 0) { throw "go build failed with exit code $LASTEXITCODE" }
    Write-Host '  [OK] Go tests and build passed'
}

Write-Host '==> Checking upstream baseline'
if (Get-Command python -ErrorAction SilentlyContinue) {
    New-Item -ItemType Directory -Path dist -Force | Out-Null
    python tools/check_upstream_updates.py --strict --output dist/upstream-review-report.md
    if ($LASTEXITCODE -ne 0) { throw "Upstream check failed with exit code $LASTEXITCODE" }
} else {
    Write-Warning 'Python not found, skipping check_upstream_updates.py'
}

git diff --check
if ($LASTEXITCODE -ne 0) { throw "git diff --check failed with exit code $LASTEXITCODE" }

Write-Host 'WINDOWS DEV CHECK GREEN'
