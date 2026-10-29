[CmdletBinding()]
param()

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
$repoRoot = Split-Path -Parent $PSScriptRoot
$testRoot = Join-Path $repoRoot ('dist\build-dist-test-' + [guid]::NewGuid().ToString('N'))
$fixture = Join-Path $testRoot 'fixture'
$utf8 = [System.Text.UTF8Encoding]::new($false)
$failures = 0

function Assert-Check {
    param([bool]$Condition, [string]$Message)
    if ($Condition) { Write-Host "  [OK] $Message" }
    else { $script:failures++; Write-Host "  [FAIL] $Message" }
}

function Invoke-FixtureBuild {
    param([string]$Name, [switch]$Install)
    $arguments = @('-NoProfile', '-ExecutionPolicy', 'Bypass', '-File', ('"' + (Join-Path $fixture 'tools\build_dist.ps1') + '"'))
    if ($Install) { $arguments += '-Install' }
    $process = Start-Process powershell -WindowStyle Hidden -Wait -PassThru -ArgumentList $arguments -RedirectStandardOutput (Join-Path $testRoot "$Name.log") -RedirectStandardError (Join-Path $testRoot "$Name.err.log")
    return $process.ExitCode
}

New-Item -ItemType Directory -Path (Join-Path $fixture 'tools'), (Join-Path $fixture 'cmd\cc-session') -Force | Out-Null
Copy-Item -LiteralPath (Join-Path $PSScriptRoot 'build_dist.ps1') -Destination (Join-Path $fixture 'tools\build_dist.ps1')
[System.IO.File]::WriteAllText((Join-Path $fixture 'go.mod'), "module buildfixture`ngo 1.22`n", $utf8)
[System.IO.File]::WriteAllText((Join-Path $fixture 'cmd\cc-session\main.go'), "package main`nvar version, commit string`nfunc main() {}`n", $utf8)
$savedLocalAppData = $env:LOCALAPPDATA
$savedPath = $env:PATH
$savedFailArch = $env:CC_SESSION_TEST_FAIL_ARCH
try {
    $env:LOCALAPPDATA = Join-Path $testRoot 'local-app-data'
    $installed = Join-Path $env:LOCALAPPDATA 'cc-session\cc-session.exe'
    New-Item -ItemType Directory -Path (Split-Path -Parent $installed) -Force | Out-Null
    [System.IO.File]::WriteAllText($installed, 'existing installation sentinel', $utf8)
    $before = (Get-FileHash -LiteralPath $installed).Hash
    $code = Invoke-FixtureBuild -Name 'default'
    Assert-Check ($code -eq 0) 'default build exits zero'
    Assert-Check ((Get-FileHash -LiteralPath $installed).Hash -eq $before) 'default build preserves installed SHA256'

    Add-Type -AssemblyName System.IO.Compression.FileSystem
    $dist = Join-Path $fixture 'dist'
    $lines = @(Get-Content -LiteralPath (Join-Path $dist 'checksums.txt'))
    Assert-Check ($lines.Count -eq 2) 'checksums contain exactly two archives'
    foreach ($arch in @('amd64', 'arm64')) {
        $zipName = "cc-session-reader_0.1.82_windows_$arch.zip"
        $zipPath = Join-Path $dist $zipName
        $hash = (Get-FileHash -LiteralPath $zipPath).Hash.ToLowerInvariant()
        Assert-Check ($lines -contains "$hash  $zipName") "$arch checksum matches archive"
        $archive = [System.IO.Compression.ZipFile]::OpenRead($zipPath)
        try {
            Assert-Check ($archive.Entries.Count -eq 1 -and $archive.Entries[0].FullName -eq 'cc-session.exe') "$arch zip contains only cc-session.exe"
            $stream = $archive.Entries[0].Open()
            $bytes = [System.IO.MemoryStream]::new()
            try {
                $stream.CopyTo($bytes)
                $data = $bytes.ToArray()
                $peOffset = [BitConverter]::ToInt32($data, 0x3c)
                $machine = [BitConverter]::ToUInt16($data, $peOffset + 4)
                $expected = if ($arch -eq 'amd64') { 0x8664 } else { 0xaa64 }
                Assert-Check ($machine -eq $expected) "$arch archive has correct PE architecture"
            } finally { $stream.Dispose(); $bytes.Dispose() }
        } finally { $archive.Dispose() }
    }

    $code = Invoke-FixtureBuild -Name 'install' -Install
    Assert-Check ($code -eq 0) 'explicit install exits zero'
    $nativeArch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString()
    $nativeZip = if ($nativeArch -eq 'Arm64') { 'arm64' } else { 'amd64' }
    $extract = Join-Path $testRoot 'native-package'
    Expand-Archive -LiteralPath (Join-Path $dist "cc-session-reader_0.1.82_windows_$nativeZip.zip") -DestinationPath $extract
    Assert-Check ((Get-FileHash -LiteralPath $installed).Hash -eq (Get-FileHash -LiteralPath (Join-Path $extract 'cc-session.exe')).Hash) 'explicit install matches native package'

    # A compiler failure with stale artifacts must never be reported as success.
    $goPath = (Get-Command go -CommandType Application).Source
    $shimDir = Join-Path $testRoot 'shim'
    New-Item -ItemType Directory -Path $shimDir | Out-Null
    $quotedGoPath = $goPath.Replace("'", "''")
    $shim = "if (`$env:GOARCH -eq `$env:CC_SESSION_TEST_FAIL_ARCH) { exit 23 }`r`n& '$quotedGoPath' @args`r`nexit `$LASTEXITCODE`r`n"
    [System.IO.File]::WriteAllText((Join-Path $shimDir 'go.ps1'), $shim, $utf8)
    $env:PATH = "$shimDir;$savedPath"
    foreach ($arch in @('amd64', 'arm64')) {
        $env:CC_SESSION_TEST_FAIL_ARCH = $arch
        $installedHash = (Get-FileHash -LiteralPath $installed).Hash
        $checksumHash = (Get-FileHash -LiteralPath (Join-Path $dist 'checksums.txt')).Hash
        Copy-Item -LiteralPath (Join-Path $extract 'cc-session.exe') -Destination (Join-Path $dist 'cc-session_arm64.exe') -Force
        $code = Invoke-FixtureBuild -Name "failure-$arch"
        Assert-Check ($code -ne 0) "$arch build failure exits nonzero despite stale artifacts"
        $errorText = Get-Content -LiteralPath (Join-Path $testRoot "failure-$arch.err.log") -Raw
        Assert-Check ($errorText -like "*go build windows/$arch failed with exit code 23*") "$arch failure reports compiler exit code 23"
        Assert-Check ((Get-FileHash -LiteralPath $installed).Hash -eq $installedHash) "$arch build failure preserves installation"
        Assert-Check ((Get-FileHash -LiteralPath (Join-Path $dist 'checksums.txt')).Hash -eq $checksumHash) "$arch build failure does not regenerate checksums"
        $code = Invoke-FixtureBuild -Name "failure-install-$arch" -Install
        Assert-Check ($code -ne 0 -and (Get-FileHash -LiteralPath $installed).Hash -eq $installedHash) "$arch build failure with explicit install preserves installation"
    }
} finally {
    $env:LOCALAPPDATA = $savedLocalAppData
    $env:PATH = $savedPath
    $env:CC_SESSION_TEST_FAIL_ARCH = $savedFailArch
}
Write-Host "Logs: $testRoot"
if ($failures -gt 0) { throw "$failures build distribution checks failed" }
Write-Host 'BUILD DIST TEST GREEN'
