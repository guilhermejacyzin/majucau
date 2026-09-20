[CmdletBinding()]
param(
    [switch]$Installer
)

$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSScriptRoot
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
        $installerStartedAt = Get-Date
        Invoke-NativeChecked $wailsCommand @('build', '-s', '-skipbindings', '-trimpath', '-nsis')
        $installer = Get-ChildItem -LiteralPath (Join-Path $projectRoot 'build\bin') -Filter '*-installer.exe' -File |
            Where-Object { $_.LastWriteTime -ge $installerStartedAt } |
            Select-Object -First 1
        if (-not $installer) {
            throw 'O Wails não produziu o instalador NSIS. Verifique se makensis está instalado e disponível no PATH.'
        }
    }

    & (Join-Path $PSScriptRoot 'consolidate-artifacts.ps1')
}
finally {
    Pop-Location
}
