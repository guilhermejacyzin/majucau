[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$projectRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
$artifactRoot = Join-Path $projectRoot 'artifacts'
$goToolBinRoot = Join-Path $artifactRoot 'tools\go\bin'
$env:GOCACHE = Join-Path $artifactRoot 'cache\go-build'
$goModCacheRoot = Join-Path $artifactRoot 'cache\go-mod'
$env:GOMODCACHE = $goModCacheRoot
$env:GOTMPDIR = Join-Path $artifactRoot 'cache\go-tmp'
$env:npm_config_cache = Join-Path $artifactRoot 'cache\npm'
$env:GOBIN = $goToolBinRoot
$goModCacheBoundary = Join-Path $goModCacheRoot 'go.mod'
New-Item -ItemType Directory -Force -Path $goModCacheRoot, $env:GOTMPDIR, $goToolBinRoot | Out-Null
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

Write-Host 'Instalando as ferramentas Go nas versões fixadas pelo projeto...'
$goCliTools = @(
    'github.com/wailsapp/wails/v2/cmd/wails@v2.13.0',
    'github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1',
    'golang.org/x/vuln/cmd/govulncheck@v1.7.0'
)
foreach ($tool in $goCliTools) {
    Invoke-NativeChecked -Command $goCommand -Arguments @('install', $tool)
}

Write-Host 'Instalando dependências frontend pelo lockfile...'
Push-Location (Join-Path $projectRoot 'frontend')
try {
    Invoke-NativeChecked -Command 'npm' -Arguments @('ci')
}
finally {
    Pop-Location
}

Write-Host 'Bootstrap concluído.'
