[CmdletBinding()]
param(
    [switch]$Installer
)

$ErrorActionPreference = 'Stop'
$projectRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
$projectLocationPushed = $false
$verificationStarted = $false
$buildOutputsStarted = $false
$artifactsPublished = $false
$buildRunId = [guid]::NewGuid().ToString('N')
$artifactBaseline = $null
$artifactRoot = Join-Path $projectRoot 'artifacts'
$goToolBinRoot = Join-Path $artifactRoot 'tools\go\bin'
$env:GOCACHE = Join-Path $artifactRoot 'cache\go-build'
$goModCacheRoot = Join-Path $artifactRoot 'cache\go-mod'
$env:GOMODCACHE = $goModCacheRoot
$env:GOTMPDIR = Join-Path $artifactRoot 'cache\go-tmp'
$env:npm_config_cache = Join-Path $artifactRoot 'cache\npm'
New-Item -ItemType Directory -Force -Path $goModCacheRoot, $env:GOTMPDIR | Out-Null
$goModCacheBoundary = Join-Path $goModCacheRoot 'go.mod'
if (-not (Test-Path -LiteralPath $goModCacheBoundary)) {
    @('module majucau.local/artifact-cache', '', 'go 1.26.0') |
        Set-Content -LiteralPath $goModCacheBoundary -Encoding utf8
}
$wailsLocal = Join-Path $goToolBinRoot 'wails.exe'
$wailsCommand = if ($env:MAJUCAU_WAILS) { $env:MAJUCAU_WAILS } elseif (Test-Path -LiteralPath $wailsLocal) { $wailsLocal } else { (Get-Command wails -ErrorAction Stop).Source }
$goCommand = if ($env:MAJUCAU_GO) { $env:MAJUCAU_GO } else { (Get-Command go -ErrorAction Stop).Source }
$env:Path = "$(Split-Path -Parent $goCommand);$env:Path"

function Invoke-NativeChecked {
    param(
        [Parameter(Mandatory)] [string]$Command,
        [Parameter(Mandatory)] [string[]]$Arguments
    )
    & $Command @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "Comando nativo falhou ($LASTEXITCODE): $Command $($Arguments -join ' ')"
    }
}

Import-Module (Join-Path $PSScriptRoot 'SafeCleanup.psm1') -Force
Import-Module (Join-Path $PSScriptRoot 'BuildFailureRetention.psm1') -Force

try {
    $verificationStarted = $true
    & (Join-Path $PSScriptRoot 'verify.ps1')

    $artifactBaseline = New-MajucauArtifactBaseline -ProjectRoot $projectRoot -RunId $buildRunId
    Push-Location $projectRoot
    $projectLocationPushed = $true
    $buildOutputsStarted = $true
    Invoke-NativeChecked $wailsCommand @('build', '-clean', '-trimpath')
    Invoke-NativeChecked $goCommand @('build', '-trimpath', '-o', (Join-Path $projectRoot 'build\bin\majucau-worker.exe'), './cmd/worker')
    Invoke-NativeChecked $goCommand @('build', '-trimpath', '-o', (Join-Path $projectRoot 'build\bin\installer-helper.exe'), './cmd/installer-helper')

    if ($Installer) {
        Invoke-NativeChecked $wailsCommand @('build', '-s', '-skipbindings', '-trimpath')
        $artifactsPublished = $true
        & (Join-Path $PSScriptRoot 'consolidate-artifacts.ps1') -KeepSourceOutputs
        & (Join-Path $PSScriptRoot 'write-release-manifest.ps1')
        & (Join-Path $PSScriptRoot 'prepare-webview2.ps1')
        $makensisLocal = Join-Path $artifactRoot 'tools\nsis-3.13\Bin\makensis.exe'
        $makensisRoots = @()
        if ($env:ProgramFiles) { $makensisRoots += (Join-Path -Path $env:ProgramFiles -ChildPath 'NSIS') }
        $programFilesX86 = [Environment]::GetEnvironmentVariable('ProgramFiles(x86)')
        if ($programFilesX86) { $makensisRoots += (Join-Path -Path $programFilesX86 -ChildPath 'NSIS') }
        $makensisInstalled = $makensisRoots |
            ForEach-Object { Join-Path $_ 'makensis.exe' } |
            Where-Object { Test-Path -LiteralPath $_ } |
            Select-Object -First 1
        $makensisCommand = if ($env:MAJUCAU_MAKENSIS) {
            $env:MAJUCAU_MAKENSIS
        } elseif (Test-Path -LiteralPath $makensisLocal) {
            $makensisLocal
        } elseif ($makensisInstalled) {
            $makensisInstalled
        } else {
            (Get-Command makensis -ErrorAction Stop).Source
        }
        Push-Location (Join-Path $projectRoot 'build\windows\installer')
        try {
            Invoke-NativeChecked $makensisCommand @('project.nsi')
        }
        finally {
            Pop-Location
        }
        $installerArtifact = Get-ChildItem -LiteralPath (Join-Path $projectRoot 'build\bin') -Filter '*-installer.exe' -File |
            Select-Object -First 1
        if (-not $installerArtifact) {
            throw 'O Wails não produziu o instalador NSIS. Verifique se makensis está instalado e disponível no PATH.'
        }
    } else {
        $artifactsPublished = $true
        & (Join-Path $PSScriptRoot 'consolidate-artifacts.ps1') -KeepSourceOutputs
        & (Join-Path $PSScriptRoot 'write-release-manifest.ps1')
    }

    & (Join-Path $PSScriptRoot 'consolidate-artifacts.ps1')
    & (Join-Path $PSScriptRoot 'package-portable.ps1')
    try {
        Complete-MajucauArtifactBaseline -ProjectRoot $projectRoot -Baseline $artifactBaseline
    }
    catch {
        Write-Warning "O build terminou, mas a cópia temporária precisa de revisão manual: $($artifactBaseline.Root). Consulte artifacts/retained/cleanup-register.md."
    }
    $artifactBaseline = $null
}
catch {
    $retentionError = $null
    if ($verificationStarted -or $buildOutputsStarted) {
        try {
            $retainedOutputs = Preserve-MajucauFailedBuildOutputs -ProjectRoot $projectRoot -RunId $buildRunId -IncludePublishedArtifacts:$artifactsPublished
            if ($retainedOutputs.Moved.Count -gt 0 -or $retainedOutputs.NotMoved.Count -gt 0) {
                Write-Warning "A execução falhou. As saídas encontradas foram registradas em artifacts/retained/cleanup-register.md."
            }
        }
        catch {
            $retentionError = $_
        }
    }
    if ($null -ne $artifactBaseline) {
        try {
            if ($artifactsPublished) {
                $restoration = Restore-MajucauArtifactBaseline -ProjectRoot $projectRoot -Baseline $artifactBaseline
                if (-not $restoration.Restored) {
                    $retentionError = 'A versão anterior permanece preservada em artifacts/cache; confira artifacts/retained/cleanup-register.md.'
                }
            } else {
                Complete-MajucauArtifactBaseline -ProjectRoot $projectRoot -Baseline $artifactBaseline
            }
        }
        catch {
            $retentionError = $_
        }
    }
    if ($null -ne $retentionError) {
        throw 'A execução falhou. A recuperação/limpeza automática não terminou; preserve os caminhos em artifacts/cache e consulte artifacts/retained/cleanup-register.md antes de qualquer limpeza.'
    }
    throw
}
finally {
    if ($projectLocationPushed) {
        Pop-Location
    }
}
