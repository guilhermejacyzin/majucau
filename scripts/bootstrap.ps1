[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$projectRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
$artifactRoot = Join-Path $projectRoot 'artifacts'
$env:GOCACHE = Join-Path $artifactRoot 'cache\go-build'
$goModCacheRoot = Join-Path $artifactRoot 'cache\go-mod'
$env:GOMODCACHE = $goModCacheRoot
$env:npm_config_cache = Join-Path $artifactRoot 'cache\npm'
$goModCacheBoundary = Join-Path $goModCacheRoot 'go.mod'
New-Item -ItemType Directory -Force -Path $goModCacheRoot | Out-Null
if (-not (Test-Path -LiteralPath $goModCacheBoundary)) {
    @('module majucau.local/artifact-cache', '', 'go 1.26.0') |
        Set-Content -LiteralPath $goModCacheBoundary -Encoding utf8
}
$goCommand = if ($env:MAJUCAU_GO) { $env:MAJUCAU_GO } else { (Get-Command go -ErrorAction Stop).Source }

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

Write-Host 'Baixando dependências Go verificadas pelo go.mod/go.sum...'
Invoke-NativeChecked -Command $goCommand -Arguments @('mod', 'download')

Write-Host 'Instalando dependências frontend pelo lockfile...'
Push-Location (Join-Path $projectRoot 'frontend')
try {
    Invoke-NativeChecked -Command 'npm' -Arguments @('ci')
}
finally {
    Pop-Location
}

Write-Host 'Bootstrap concluído.'
