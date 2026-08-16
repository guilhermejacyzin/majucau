[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSScriptRoot
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
