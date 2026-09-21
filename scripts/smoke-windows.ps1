[CmdletBinding()]
param(
    [string]$ArtifactRoot = ''
)

$ErrorActionPreference = 'Stop'
$projectRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
if ([string]::IsNullOrWhiteSpace($ArtifactRoot)) {
    $ArtifactRoot = Join-Path $projectRoot 'artifacts'
}
$windowsRoot = Join-Path $ArtifactRoot 'windows'
$smokeRoot = Join-Path $ArtifactRoot 'cache\windows-smoke-run'
$portableZip = Join-Path $windowsRoot 'majucau-windows-x64-portable.zip'
$required = @('majucau.exe', 'majucau-worker.exe', 'installer-helper.exe')

function Assert-Condition {
    param([bool]$Condition, [string]$Message)
    if (-not $Condition) {
        throw $Message
    }
}

if (Test-Path -LiteralPath $smokeRoot) {
    Remove-Item -LiteralPath $smokeRoot -Recurse -Force
}
New-Item -ItemType Directory -Force -Path $smokeRoot | Out-Null
Assert-Condition (Test-Path -LiteralPath $portableZip -PathType Leaf) "Bundle portátil ausente: $portableZip"

foreach ($name in $required) {
    Assert-Condition (Test-Path -LiteralPath (Join-Path $windowsRoot $name) -PathType Leaf) "Artefato ausente: $name"
}

$portableRoot = Join-Path $smokeRoot 'portable'
Expand-Archive -LiteralPath $portableZip -DestinationPath $portableRoot
foreach ($name in $required + @('README-PORTABLE.txt', 'SHA256SUMS.txt')) {
    Assert-Condition (Test-Path -LiteralPath (Join-Path $portableRoot $name) -PathType Leaf) "Arquivo ausente no bundle: $name"
}

$sumLines = Get-Content -LiteralPath (Join-Path $portableRoot 'SHA256SUMS.txt') |
    Where-Object { $_ -match '^(?<hash>[0-9a-fA-F]{64})\s+(?<name>[^\s]+)$' }
foreach ($line in $sumLines) {
    $expected = $Matches.hash.ToLowerInvariant()
    $name = $Matches.name
    $actual = (Get-FileHash -Algorithm SHA256 -LiteralPath (Join-Path $portableRoot $name)).Hash.ToLowerInvariant()
    Assert-Condition ($actual -eq $expected) "Hash divergente no bundle: $name"
}

$helper = Join-Path $portableRoot 'installer-helper.exe'
$installPath = Join-Path $smokeRoot 'install'
$dataPath = Join-Path $smokeRoot 'data'
$preflightPath = Join-Path $smokeRoot 'preflight.json'
$diagnosticZip = Join-Path $smokeRoot 'diagnostic.zip'
$freePath = $smokeRoot

& $helper preflight --install-dir $installPath --data-dir $dataPath --free-space-path $freePath --min-free-bytes 1 --port 55439 |
    Set-Content -LiteralPath $preflightPath -Encoding utf8
$preflightExit = $LASTEXITCODE
Assert-Condition ($preflightExit -in @(0, 2)) "Preflight retornou código inesperado: $preflightExit"
$preflight = Get-Content -LiteralPath $preflightPath -Raw | ConvertFrom-Json
Assert-Condition ($preflight.schema_version -eq '1.0') 'Preflight sem schema_version esperado.'
Assert-Condition ($preflight.status -in @('READY', 'BLOCKED')) "Status de preflight inesperado: $($preflight.status)"
Assert-Condition ($preflight.checks.platform.architecture -eq 'amd64') 'Artefato não reportou arquitetura x64.'
Assert-Condition ($preflight.checks.preferred_port.available -eq $true) 'Porta local de smoke não está disponível.'

& $helper diagnostics --output $diagnosticZip --install-dir $installPath --data-dir $dataPath --free-space-path $freePath --min-free-bytes 1 --port 55439 |
    Set-Content -LiteralPath (Join-Path $smokeRoot 'diagnostics-result.json') -Encoding utf8
$diagnosticsExit = $LASTEXITCODE
Assert-Condition ($diagnosticsExit -in @(0, 2)) "Diagnostics retornou código inesperado: $diagnosticsExit"
Assert-Condition (Test-Path -LiteralPath $diagnosticZip -PathType Leaf) 'Bundle de diagnóstico não foi criado.'

$diagnosticRoot = Join-Path $smokeRoot 'diagnostic'
Expand-Archive -LiteralPath $diagnosticZip -DestinationPath $diagnosticRoot
Assert-Condition (Test-Path -LiteralPath (Join-Path $diagnosticRoot 'diagnostic.json') -PathType Leaf) 'diagnostic.json ausente.'
Assert-Condition (Test-Path -LiteralPath (Join-Path $diagnosticRoot 'README.txt') -PathType Leaf) 'README.txt ausente.'
$diagnosticText = Get-Content -LiteralPath (Join-Path $diagnosticRoot 'diagnostic.json') -Raw
Assert-Condition ($diagnosticText -notmatch '(?i)([A-Z]:\\|[A-Z]:/|password|secret|token|dsn)') 'Diagnóstico contém dado sensível ou caminho completo.'

$workerStdout = Join-Path $smokeRoot 'worker.stdout.log'
$workerStderr = Join-Path $smokeRoot 'worker.stderr.log'
$worker = Join-Path $portableRoot 'majucau-worker.exe'
$workerProcess = Start-Process -FilePath $worker -ArgumentList '--console' -WorkingDirectory $portableRoot -WindowStyle Hidden -RedirectStandardOutput $workerStdout -RedirectStandardError $workerStderr -PassThru
Start-Sleep -Seconds 2
$workerStarted = -not $workerProcess.HasExited
if ($workerStarted) {
    Stop-Process -Id $workerProcess.Id -Force -ErrorAction SilentlyContinue
    Wait-Process -Id $workerProcess.Id -Timeout 5 -ErrorAction SilentlyContinue
}
Assert-Condition $workerStarted 'Worker não permaneceu ativo durante o smoke console.'
$workerHealth = Get-Content -LiteralPath $workerStdout -Raw | ConvertFrom-Json
Assert-Condition ($workerHealth.service -eq 'majucau-worker' -and $workerHealth.state -eq 'OK') 'Health do worker não ficou OK.'

$result = [ordered]@{
    schema_version = '1.0'
    status = 'PASS'
    preflight_status = $preflight.status
    preflight_exit_code = $preflightExit
    diagnostics_exit_code = $diagnosticsExit
    worker_health = $workerHealth.state
    portable_files = $required + @('README-PORTABLE.txt', 'SHA256SUMS.txt')
    diagnostic_entries = @('diagnostic.json', 'README.txt')
}
$resultPath = Join-Path $smokeRoot 'SMOKE-RESULT.json'
$result | ConvertTo-Json -Depth 6 | Set-Content -LiteralPath $resultPath -Encoding utf8
Write-Output "Windows smoke PASS: $resultPath"

