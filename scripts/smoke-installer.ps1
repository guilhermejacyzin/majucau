[CmdletBinding()]
param(
    [Parameter(Mandatory)] [string]$InstallerPath,
    [string]$OutputRoot = ''
)

$ErrorActionPreference = 'Stop'
$projectRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
if ([string]::IsNullOrWhiteSpace($OutputRoot)) {
    $OutputRoot = Join-Path $projectRoot 'artifacts\cache\installer-smoke'
}
$OutputRoot = [System.IO.Path]::GetFullPath($OutputRoot)
$InstallerPath = (Resolve-Path -LiteralPath $InstallerPath).Path

function Assert-Condition {
    param([bool]$Condition, [string]$Message)
    if (-not $Condition) { throw $Message }
}

function Invoke-ElevatedExecutable {
    param(
        [Parameter(Mandatory)] [string]$FilePath,
        [Parameter(Mandatory)] [string[]]$Arguments,
        [Parameter(Mandatory)] [string]$TaskRoot,
        [int]$TimeoutSeconds = 180
    )

    $taskName = "MajucauSmoke-$([guid]::NewGuid().ToString('N'))"
    $scriptPath = Join-Path $TaskRoot "$taskName.ps1"
    $resultPath = Join-Path $TaskRoot "$taskName.exit"
    $errorPath = Join-Path $TaskRoot "$taskName.error"
    $fileLiteral = $FilePath.Replace("'", "''")
    $argumentLiteral = ($Arguments | ForEach-Object { "'$(($_.Replace("'", "''")))'" }) -join ', '
    $resultLiteral = $resultPath.Replace("'", "''")
    $errorLiteral = $errorPath.Replace("'", "''")
    $taskScript = @"
`$ErrorActionPreference = 'Stop'
try {
    `$process = Start-Process -FilePath '$fileLiteral' -ArgumentList @($argumentLiteral) -Wait -PassThru
    Set-Content -LiteralPath '$resultLiteral' -Value `$process.ExitCode -Encoding utf8
} catch {
    Set-Content -LiteralPath '$errorLiteral' -Value `$_.Exception.ToString() -Encoding utf8
    exit 1
}
"@
    Set-Content -LiteralPath $scriptPath -Value $taskScript -Encoding utf8

    $powerShell = (Get-Command powershell.exe -ErrorAction Stop).Source
    $action = New-ScheduledTaskAction -Execute $powerShell -Argument "-NoProfile -ExecutionPolicy Bypass -File `"$scriptPath`""
    $identity = if ($env:USERDOMAIN) { "$env:USERDOMAIN\$env:USERNAME" } else { $env:USERNAME }
    $principal = New-ScheduledTaskPrincipal -UserId $identity -LogonType Interactive -RunLevel Highest
    try {
        Register-ScheduledTask -TaskName $taskName -Action $action -Principal $principal -Force | Out-Null
        Start-ScheduledTask -TaskName $taskName
        $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
        do {
            Start-Sleep -Seconds 2
            if (Test-Path -LiteralPath $resultPath -PathType Leaf) { break }
            if (Test-Path -LiteralPath $errorPath -PathType Leaf) { break }
        } while ((Get-Date) -lt $deadline)

        if (Test-Path -LiteralPath $errorPath -PathType Leaf) {
            throw "Processo elevado falhou: $(Get-Content -LiteralPath $errorPath -Raw)"
        }
        if (-not (Test-Path -LiteralPath $resultPath -PathType Leaf)) {
            throw "Processo elevado excedeu $TimeoutSeconds segundos: $FilePath"
        }
        return [int](Get-Content -LiteralPath $resultPath -Raw).Trim()
    }
    finally {
        Unregister-ScheduledTask -TaskName $taskName -Confirm:$false -ErrorAction SilentlyContinue
        Remove-Item -LiteralPath $scriptPath, $resultPath, $errorPath -Force -ErrorAction SilentlyContinue
    }
}

if (Test-Path -LiteralPath $OutputRoot) { Remove-Item -LiteralPath $OutputRoot -Recurse -Force }
New-Item -ItemType Directory -Force -Path $OutputRoot | Out-Null
$installDir = Join-Path $OutputRoot 'app'
$dataDir = Join-Path $OutputRoot 'data'
$helper = Join-Path $installDir 'installer-helper.exe'
$preflightPath = Join-Path $OutputRoot 'preflight.json'
$resultPath = Join-Path $OutputRoot 'INSTALL-SMOKE-RESULT.json'

$installExit = Invoke-ElevatedExecutable -FilePath $InstallerPath -Arguments @('/S', "/D=$installDir") -TaskRoot $OutputRoot
Assert-Condition ($installExit -eq 0) "Instalação elevada retornou código $installExit."
foreach ($name in @('Majucau Financial Intelligence.exe', 'majucau-worker.exe', 'installer-helper.exe', 'uninstall.exe')) {
    Assert-Condition (Test-Path -LiteralPath (Join-Path $installDir $name) -PathType Leaf) "Arquivo ausente após instalação: $name"
}

& $helper preflight --install-dir $installDir --data-dir $dataDir --free-space-path $OutputRoot --min-free-bytes 1 --port 55439 |
    Set-Content -LiteralPath $preflightPath -Encoding utf8
$preflightExit = $LASTEXITCODE
Assert-Condition ($preflightExit -in @(0, 2)) "Preflight pós-instalação retornou código inesperado: $preflightExit"
$preflight = Get-Content -LiteralPath $preflightPath -Raw | ConvertFrom-Json
Assert-Condition ($preflight.schema_version -eq '1.0') 'Preflight pós-instalação sem schema.'

$uninstaller = Join-Path $installDir 'uninstall.exe'
$uninstallExit = Invoke-ElevatedExecutable -FilePath $uninstaller -Arguments @('/S') -TaskRoot $OutputRoot
Assert-Condition ($uninstallExit -eq 0) "Desinstalação elevada retornou código $uninstallExit."
Assert-Condition (-not (Test-Path -LiteralPath $installDir)) 'Diretório da aplicação permaneceu após desinstalação.'

$result = [ordered]@{
    schema_version = '1.0'
    status = 'PASS'
    install_exit_code = $installExit
    preflight_exit_code = $preflightExit
    preflight_status = $preflight.status
    uninstall_exit_code = $uninstallExit
    installer = [System.IO.Path]::GetFileName($InstallerPath)
    mode = 'scheduled-task-runlevel-highest'
}
$result | ConvertTo-Json -Depth 6 | Set-Content -LiteralPath $resultPath -Encoding utf8
Write-Output "Installer smoke PASS: $resultPath"

