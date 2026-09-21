[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$projectRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
$installerPath = Join-Path $projectRoot 'build\windows\installer\project.nsi'
if (-not (Test-Path -LiteralPath $installerPath)) {
    throw "Arquivo NSIS não encontrado: $installerPath"
}

$source = Get-Content -LiteralPath $installerPath -Raw
$required = [ordered]@{
    helper_source_define = 'MAJUCAU_HELPER_SOURCE'
    helper_embedded_before_install = 'File /oname=$PLUGINSDIR\majucau-installer-helper.exe'
    diagnostics_command = 'diagnostics --output'
    preflight_data_dir = '--data-dir "${MAJUCAU_DATA_DIR}"'
    preflight_port = '--port ${MAJUCAU_PREFERRED_PORT}'
    blocked_exit_handling = '${If} $1 == 2'
    abort_on_failed_preflight = 'Quit'
    installed_helper = 'File "/oname=installer-helper.exe"'
}

$missing = @(
    $required.GetEnumerator() |
        Where-Object { $source -notlike "*$($_.Value)*" } |
        ForEach-Object { $_.Key }
)
if ($missing.Count -gt 0) {
    throw "Contrato do instalador NSIS incompleto. Ausentes: $($missing -join ', ')"
}

$initStart = $source.IndexOf('Function .onInit', [System.StringComparison]::Ordinal)
$sectionStart = $source.IndexOf('Section', [System.StringComparison]::Ordinal)
if ($initStart -lt 0 -or $sectionStart -lt 0 -or $initStart -gt $sectionStart) {
    throw 'Não foi possível localizar a ordem .onInit -> Section no instalador.'
}
$initBlock = $source.Substring($initStart, $sectionStart - $initStart)
if ($initBlock.IndexOf('diagnostics --output', [System.StringComparison]::Ordinal) -lt 0) {
    throw 'O diagnóstico precisa ocorrer dentro de .onInit, antes da Section de instalação.'
}

$result = [ordered]@{
    schema_version = '1.0'
    status = 'PASS'
    installer = 'build/windows/installer/project.nsi'
    checks = @($required.Keys)
}
$result | ConvertTo-Json -Depth 4

