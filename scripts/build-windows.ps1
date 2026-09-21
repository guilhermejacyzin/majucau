[CmdletBinding()]
param(
    [switch]$Installer
)

$ErrorActionPreference = 'Stop'
$projectRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
$artifactRoot = Join-Path $projectRoot 'artifacts'
$env:GOCACHE = Join-Path $artifactRoot 'cache\go-build'
$goModCacheRoot = Join-Path $artifactRoot 'cache\go-mod'
$env:GOMODCACHE = $goModCacheRoot
$env:npm_config_cache = Join-Path $artifactRoot 'cache\npm'
New-Item -ItemType Directory -Force -Path $goModCacheRoot | Out-Null
$goModCacheBoundary = Join-Path $goModCacheRoot 'go.mod'
if (-not (Test-Path -LiteralPath $goModCacheBoundary)) {
    @('module majucau.local/artifact-cache', '', 'go 1.26.0') |
        Set-Content -LiteralPath $goModCacheBoundary -Encoding utf8
}
$wailsCommand = if ($env:MAJUCAU_WAILS) { $env:MAJUCAU_WAILS } else { (Get-Command wails -ErrorAction Stop).Source }
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

& (Join-Path $PSScriptRoot 'verify.ps1')

Push-Location $projectRoot
try {
    Invoke-NativeChecked $wailsCommand @('build', '-clean', '-trimpath')
    Invoke-NativeChecked $goCommand @('build', '-trimpath', '-o', (Join-Path $projectRoot 'build\bin\majucau-worker.exe'), './cmd/worker')
    Invoke-NativeChecked $goCommand @('build', '-trimpath', '-o', (Join-Path $projectRoot 'build\bin\installer-helper.exe'), './cmd/installer-helper')

    if ($Installer) {
        Invoke-NativeChecked $wailsCommand @('build', '-s', '-skipbindings', '-trimpath')
        & (Join-Path $PSScriptRoot 'consolidate-artifacts.ps1') -KeepSourceOutputs
        & (Join-Path $PSScriptRoot 'write-release-manifest.ps1')
        & (Join-Path $PSScriptRoot 'prepare-webview2.ps1')
        $makensisCommand = (Get-Command makensis -ErrorAction Stop).Source
        Push-Location (Join-Path $projectRoot 'build\windows\installer')
        try {
            Invoke-NativeChecked $makensisCommand @('project.nsi')
        }
        finally {
            Pop-Location
        }
        $installer = Get-ChildItem -LiteralPath (Join-Path $projectRoot 'build\bin') -Filter '*-installer.exe' -File |
            Select-Object -First 1
        if (-not $installer) {
            throw 'O Wails não produziu o instalador NSIS. Verifique se makensis está instalado e disponível no PATH.'
        }
    } else {
        & (Join-Path $PSScriptRoot 'consolidate-artifacts.ps1') -KeepSourceOutputs
        & (Join-Path $PSScriptRoot 'write-release-manifest.ps1')
    }

    & (Join-Path $PSScriptRoot 'consolidate-artifacts.ps1')
    & (Join-Path $PSScriptRoot 'package-portable.ps1')
}
finally {
    Pop-Location
}
