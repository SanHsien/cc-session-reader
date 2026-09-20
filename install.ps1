#Requires -Version 5.1
[CmdletBinding()]
param(
    [switch]$NoSkill
)

$ErrorActionPreference = 'Stop'

$BinaryRepo = 'SanHsien/cc-session-reader'
$SkillRepo = 'SanHsien/cc-session-reader'
$InstallDir = Join-Path $env:LOCALAPPDATA 'cc-session'
$ClaudeCodeSkillDir = Join-Path $HOME '.claude\skills\cc-session'
$ClaudeHandoffSkillDir = Join-Path $HOME '.claude\skills\agent-handoff'
$CodexSkillDir = Join-Path $HOME '.codex\skills\agent-handoff'
$ClaudeDesktopPackage = Join-Path $InstallDir 'agent-handoff-claude.zip'
$CCSessionSkillUrl = "https://raw.githubusercontent.com/$SkillRepo/main/SKILL.md"
$AgentHandoffSkillUrl = "https://raw.githubusercontent.com/$SkillRepo/main/skills/agent-handoff/SKILL.md"
$AgentHandoffMetadataUrl = "https://raw.githubusercontent.com/$SkillRepo/main/skills/agent-handoff/agents/openai.yaml"

function Read-HostOrDefault {
    param([string]$Prompt, [string]$Default)

    try {
        $result = Read-Host $Prompt
        if ([string]::IsNullOrEmpty($result)) { return $Default }
        return $result
    } catch {
        Write-Host "(non-interactive: using default '$Default')"
        return $Default
    }
}

function Get-Architecture {
    $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture
    switch ($arch) {
        'X64' { return 'amd64' }
        'Arm64' { return 'arm64' }
        default { throw "Unsupported architecture: $arch" }
    }
}

function Get-LatestVersion {
    $apiUrl = "https://api.github.com/repos/$BinaryRepo/releases/latest"
    try {
        $response = Invoke-RestMethod -Uri $apiUrl -UseBasicParsing
        if (-not $response.tag_name) {
            throw 'GitHub API response did not include tag_name.'
        }
        return [string]$response.tag_name
    } catch {
        throw "Failed to fetch the latest upstream Windows release: $_"
    }
}

function Install-Binary {
    param([string]$Version, [string]$Arch)

    $versionBare = $Version.TrimStart('v')
    $zipName = "cc-session-reader_${versionBare}_windows_${Arch}.zip"
    $downloadUrl = "https://github.com/$BinaryRepo/releases/download/$Version/$zipName"
    $tmpDir = Join-Path ([System.IO.Path]::GetTempPath()) ([System.IO.Path]::GetRandomFileName())

    Write-Host "Downloading cc-session $Version for windows/$Arch from $BinaryRepo..."
    try {
        New-Item -ItemType Directory -Path $tmpDir | Out-Null
        $zipPath = Join-Path $tmpDir $zipName
        Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath -UseBasicParsing
        Expand-Archive -LiteralPath $zipPath -DestinationPath $tmpDir -Force

        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        $exeSrc = Join-Path $tmpDir 'cc-session.exe'
        if (-not (Test-Path -LiteralPath $exeSrc)) {
            throw "Release archive did not contain cc-session.exe: $downloadUrl"
        }
        Move-Item -LiteralPath $exeSrc -Destination (Join-Path $InstallDir 'cc-session.exe') -Force
        Write-Host "Installed cc-session to $InstallDir"
    } finally {
        Remove-Item -LiteralPath $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

function Update-UserPath {
    $currentPath = [Environment]::GetEnvironmentVariable('PATH', 'User')
    $dirs = @($currentPath -split ';' | Where-Object { $_ })
    if ($dirs -contains $InstallDir) { return }

    Write-Host ''
    Write-Host "Warning: $InstallDir is not in your user PATH."
    if (-not [Environment]::UserInteractive) {
        Write-Host 'Add it manually to your user PATH.'
        return
    }

    $answer = Read-HostOrDefault -Prompt "Add $InstallDir to user PATH? [Y/n]" -Default 'Y'
    if ($answer -match '^[Yy]$') {
        [Environment]::SetEnvironmentVariable('PATH', (($dirs + $InstallDir) -join ';'), 'User')
        Write-Host 'Added to user PATH. Restart your terminal to apply.'
    }
}

function Install-RemoteFile {
    param([string]$Url, [string]$Destination)

    $parent = Split-Path -Parent $Destination
    New-Item -ItemType Directory -Path $parent -Force | Out-Null
    Invoke-WebRequest -Uri $Url -OutFile $Destination -UseBasicParsing
}

function Sync-ArgumentHint {
    param([string]$SkillPath)

    $exePath = Join-Path $InstallDir 'cc-session.exe'
    if (-not (Test-Path -LiteralPath $exePath)) { return }

    try {
        $hint = & $exePath help --argument-hint 2>$null
        if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($hint) -or -not $hint.StartsWith('[')) {
            return
        }

        $updated = Get-Content -LiteralPath $SkillPath -Encoding UTF8 | ForEach-Object {
            if ($_ -match '^argument-hint:') { "argument-hint: `"$hint`"" } else { $_ }
        }
        [System.IO.File]::WriteAllLines(
            $SkillPath,
            [string[]]$updated,
            [System.Text.UTF8Encoding]::new($false)
        )
    } catch {
        Write-Warning "Could not synchronize the cc-session argument hint: $_"
    }
}

function New-ClaudeDesktopPackage {
    $tmpDir = Join-Path ([System.IO.Path]::GetTempPath()) ([System.IO.Path]::GetRandomFileName())
    try {
        $packageRoot = Join-Path $tmpDir 'agent-handoff'
        New-Item -ItemType Directory -Path $packageRoot -Force | Out-Null
        Install-RemoteFile -Url $AgentHandoffSkillUrl -Destination (Join-Path $packageRoot 'SKILL.md')
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        Compress-Archive -LiteralPath $packageRoot -DestinationPath $ClaudeDesktopPackage -Force
        Write-Host "Created Claude Desktop upload package: $ClaudeDesktopPackage"
    } finally {
        Remove-Item -LiteralPath $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

function Install-Skills {
    if ($NoSkill) { return }

    if ([Environment]::UserInteractive) {
        $answer = Read-HostOrDefault -Prompt 'Install cc-session and agent-handoff skills? [Y/n]' -Default 'Y'
        if ($answer -match '^[Nn]$') { return }
    }

    $ccSessionDestination = Join-Path $ClaudeCodeSkillDir 'SKILL.md'
    Install-RemoteFile -Url $CCSessionSkillUrl -Destination $ccSessionDestination
    Sync-ArgumentHint -SkillPath $ccSessionDestination
    Write-Host "Installed Claude Code session-reader skill: $ccSessionDestination"

    $claudeHandoffDestination = Join-Path $ClaudeHandoffSkillDir 'SKILL.md'
    Install-RemoteFile -Url $AgentHandoffSkillUrl -Destination $claudeHandoffDestination
    Write-Host "Installed Claude Code handoff skill: $claudeHandoffDestination"

    $codexHandoffDestination = Join-Path $CodexSkillDir 'SKILL.md'
    Install-RemoteFile -Url $AgentHandoffSkillUrl -Destination $codexHandoffDestination
    Install-RemoteFile -Url $AgentHandoffMetadataUrl -Destination (Join-Path $CodexSkillDir 'agents\openai.yaml')
    Write-Host "Installed Codex handoff skill: $codexHandoffDestination"

    New-ClaudeDesktopPackage
}

function Show-NextSteps {
    Write-Host ''
    Write-Host 'Getting started'
    Write-Host '  Stage 1: Summarize the current work into HANDOFF.md for the next agent.'
    Write-Host '  Stage 2: Read HANDOFF.md, verify the workspace, and continue the next step.'
    if (Test-Path -LiteralPath $ClaudeDesktopPackage) {
        Write-Host ''
        Write-Host 'Claude Desktop requires a one-time manual upload:'
        Write-Host "  Customize > Skills > + Create skill > Upload a skill > $ClaudeDesktopPackage"
    }
    Write-Host ''
    Write-Host 'Advanced CLI examples:'
    Write-Host '  cc-session list'
    Write-Host '  cc-session inherit <id>'
}

$version = Get-LatestVersion
$arch = Get-Architecture
Install-Binary -Version $version -Arch $arch
Update-UserPath
Install-Skills
Show-NextSteps
