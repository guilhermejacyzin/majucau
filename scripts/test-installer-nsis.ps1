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
    preflight_command = 'preflight --install-dir'
    diagnostics_command = 'diagnostics --output'
    package_verify_command = 'verify-package --package-dir'
    package_manifest_source = 'MAJUCAU_MANIFEST_SOURCE'
    package_migration_source = 'MAJUCAU_MIGRATION_SOURCE'
    package_integrity_failure = 'O pacote do instalador não passou na verificação de integridade'
    preflight_data_dir = '--data-dir "${MAJUCAU_DATA_DIR}"'
    preflight_port = '--port ${MAJUCAU_PREFERRED_PORT}'
    blocked_exit_handling = '${If} $1 == 2'
    silent_block_handling = 'IfSilent MajuauPreflightSilent MajuauPreflightInteractive'
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
$packageIndex = $initBlock.IndexOf('verify-package --package-dir', [System.StringComparison]::Ordinal)
$sectionIndex = $initBlock.IndexOf('FunctionEnd', [System.StringComparison]::Ordinal)
if ($packageIndex -lt 0 -or $sectionIndex -lt 0 -or $packageIndex -gt $sectionIndex) {
    throw 'A verificação de integridade do pacote precisa ocorrer dentro de .onInit.'
}
$preflightIndex = $initBlock.IndexOf('preflight --install-dir', [System.StringComparison]::Ordinal)
$diagnosticsIndex = $initBlock.IndexOf('diagnostics --output', [System.StringComparison]::Ordinal)
if ($preflightIndex -lt 0 -or $diagnosticsIndex -lt 0 -or $preflightIndex -gt $diagnosticsIndex) {
    throw 'O preflight precisa ocorrer antes da geração do diagnóstico.'
}

$result = [ordered]@{
    schema_version = '1.0'
    status = 'PASS'
    installer = 'build/windows/installer/project.nsi'
    checks = @($required.Keys)
}
$result | ConvertTo-Json -Depth 4
