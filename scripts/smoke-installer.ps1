[CmdletBinding()]
param(
    [Parameter(Mandatory)] [string]$InstallerPath,
    [string]$OutputRoot = ''
)

$ErrorActionPreference = 'Stop'
Import-Module (Join-Path $PSScriptRoot 'BoundedText.psm1') -Force
$projectRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
if ([string]::IsNullOrWhiteSpace($OutputRoot)) {
    $OutputRoot = Join-Path $projectRoot 'artifacts\cache\installer-smoke'
}
$OutputRoot = [System.IO.Path]::GetFullPath($OutputRoot)
$InstallerPath = (Resolve-Path -LiteralPath $InstallerPath).Path
$outputPathRoot = [System.IO.Path]::GetPathRoot($OutputRoot)
if ($OutputRoot.Length -gt $outputPathRoot.Length) {
    $OutputRoot = $OutputRoot.TrimEnd([char[]]@([System.IO.Path]::DirectorySeparatorChar, [System.IO.Path]::AltDirectorySeparatorChar))
}
$artifactsRoot = [System.IO.Path]::GetFullPath((Join-Path $projectRoot 'artifacts'))
$cacheOutputRoot = [System.IO.Path]::GetFullPath((Join-Path $artifactsRoot 'cache\installer-smoke'))
$verificationRoot = [System.IO.Path]::GetFullPath((Join-Path $artifactsRoot 'verification'))
$verificationOutputPrefix = Join-Path $verificationRoot 'install-smoke-'

function Assert-PlainDirectoryPath {
    param([Parameter(Mandatory)] [string]$Path)
    if (-not (Test-Path -LiteralPath $Path)) { return }
    $item = Get-Item -LiteralPath $Path -Force -ErrorAction Stop
    if (-not $item.PSIsContainer) {
        throw "O caminho reservado para o smoke não é uma pasta: $Path"
    }
    if (($item.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
        throw "O smoke não pode usar nem apagar uma pasta ligada a outro local: $Path"
    }
}

function Assert-NoReparseDescendants {
    param([Parameter(Mandatory)] [string]$Path)
    if (-not (Test-Path -LiteralPath $Path -PathType Container)) { return }
    $linkedItem = Get-ChildItem -LiteralPath $Path -Force -Recurse -ErrorAction Stop |
        Where-Object { ($_.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0 } |
        Select-Object -First 1
    if ($null -ne $linkedItem) {
        throw "O smoke não pode apagar uma pasta que contém um link para outro local: $($linkedItem.FullName)"
    }
}

$isCacheOutput = [string]::Equals($OutputRoot, $cacheOutputRoot, [System.StringComparison]::OrdinalIgnoreCase)
$verificationSuffix = ''
if ($OutputRoot.StartsWith($verificationOutputPrefix, [System.StringComparison]::OrdinalIgnoreCase)) {
    $verificationSuffix = $OutputRoot.Substring($verificationOutputPrefix.Length)
}
$isVerificationOutput = $verificationSuffix -match '^\d+-\d+$'
if (-not ($isCacheOutput -or $isVerificationOutput)) {
    throw 'OutputRoot precisa ser artifacts/cache/installer-smoke ou artifacts/verification/install-smoke-<run>-<tentativa>.'
}
$outputParent = if ($isCacheOutput) { Join-Path $artifactsRoot 'cache' } else { $verificationRoot }
foreach ($path in @($projectRoot, $artifactsRoot, $outputParent, $OutputRoot)) {
    Assert-PlainDirectoryPath -Path $path
}

function Assert-Condition {
    param([bool]$Condition, [string]$Message)
    if (-not $Condition) { throw $Message }
}

function Get-PreflightIssueCodes {
    param([string]$FallbackPath = '')
    $candidates = @(
        (Join-Path $env:TEMP 'Majucau\installer-preflight.zip'),
        (Join-Path ([System.IO.Path]::GetTempPath()) 'Majucau\installer-preflight.zip'),
        (Join-Path $env:WINDIR 'Temp\Majucau\installer-preflight.zip')
    ) | Select-Object -Unique
    foreach ($candidate in $candidates) {
        if (-not (Test-Path -LiteralPath $candidate -PathType Leaf)) { continue }
        $extractRoot = Join-Path $OutputRoot 'installer-preflight'
        Assert-PlainDirectoryPath -Path $extractRoot
        if (Test-Path -LiteralPath $extractRoot) {
            Assert-NoReparseDescendants -Path $extractRoot
            Remove-Item -LiteralPath $extractRoot -Recurse -Force
        }
        Expand-Archive -LiteralPath $candidate -DestinationPath $extractRoot -Force
        $jsonPath = Get-ChildItem -LiteralPath $extractRoot -Filter 'preflight.json' -File -Recurse | Select-Object -First 1
        if (-not $jsonPath) { continue }
        $json = (Read-BoundedTextFile -Path $jsonPath.FullName -MaxBytes 262144) | ConvertFrom-Json
        return @($json.issues | ForEach-Object { $_.code })
    }
    if (-not [string]::IsNullOrWhiteSpace($FallbackPath) -and (Test-Path -LiteralPath $FallbackPath -PathType Leaf)) {
        $fallback = (Read-BoundedTextFile -Path $FallbackPath -MaxBytes 262144) | ConvertFrom-Json
        return @($fallback.issues | ForEach-Object { $_.code })
    }
    return @('PREFLIGHT_DIAGNOSTIC_NOT_FOUND')
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
    $taskRegistered = $false
    $exitCode = $null
    $operationError = $null
    $cleanupFailures = New-Object 'System.Collections.Generic.List[string]'
    try {
        Set-Content -LiteralPath $scriptPath -Value $taskScript -Encoding utf8

        $powerShell = (Get-Command powershell.exe -ErrorAction Stop).Source
        $action = New-ScheduledTaskAction -Execute $powerShell -Argument "-NoProfile -ExecutionPolicy Bypass -File `"$scriptPath`""
        $identity = if ($env:USERDOMAIN) { "$env:USERDOMAIN\$env:USERNAME" } else { $env:USERNAME }
        $principal = New-ScheduledTaskPrincipal -UserId $identity -LogonType Interactive -RunLevel Highest
        Register-ScheduledTask -TaskName $taskName -Action $action -Principal $principal -Force | Out-Null
        $taskRegistered = $true
        Start-ScheduledTask -TaskName $taskName
        $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
        do {
            Start-Sleep -Seconds 2
            if (Test-Path -LiteralPath $resultPath -PathType Leaf) { break }
            if (Test-Path -LiteralPath $errorPath -PathType Leaf) { break }
        } while ((Get-Date) -lt $deadline)

        if (Test-Path -LiteralPath $errorPath -PathType Leaf) {
            throw "Processo elevado falhou: $(Read-BoundedTextFile -Path $errorPath -MaxBytes 65536)"
        }
        if (-not (Test-Path -LiteralPath $resultPath -PathType Leaf)) {
            throw "Processo elevado excedeu $TimeoutSeconds segundos: $FilePath"
        }
        $exitCode = [int](Read-BoundedTextFile -Path $resultPath -MaxBytes 4096).Trim()
    } catch {
        $operationError = $_
    }

    if ($taskRegistered) {
        try {
            Unregister-ScheduledTask -TaskName $taskName -Confirm:$false -ErrorAction Stop
        } catch {
            $cleanupHResult = '0x{0:X8}' -f $_.Exception.HResult
            $cleanupFailures.Add("tarefa agendada '$taskName' (HRESULT=$cleanupHResult)")
        }
    }
    foreach ($path in @($scriptPath, $resultPath, $errorPath)) {
        try {
            [System.IO.File]::Delete($path)
        } catch {
            $cleanupHResult = '0x{0:X8}' -f $_.Exception.HResult
            $cleanupName = [System.IO.Path]::GetFileName($path)
            $cleanupFailures.Add("arquivo auxiliar '$cleanupName' (HRESULT=$cleanupHResult)")
        }
    }
    if ($cleanupFailures.Count -gt 0) {
        $cleanupSummary = "limpeza incompleta; verifique os arquivos auxiliares do smoke e a tarefa '$taskName': $($cleanupFailures -join '; ')"
        if ($null -ne $operationError) {
            throw "$($operationError.Exception.Message) Além disso, $cleanupSummary."
        }
        throw $cleanupSummary
    }
    if ($null -ne $operationError) {
        throw $operationError
    }
    return [int]$exitCode
}

if (Test-Path -LiteralPath $OutputRoot) {
    Assert-NoReparseDescendants -Path $OutputRoot
    Remove-Item -LiteralPath $OutputRoot -Recurse -Force
}
New-Item -ItemType Directory -Force -Path $OutputRoot | Out-Null
$installDir = Join-Path $OutputRoot 'app'
$dataDir = Join-Path $OutputRoot 'data'
$helper = Join-Path $installDir 'installer-helper.exe'
$sourceHelper = Join-Path $projectRoot 'build\bin\installer-helper.exe'
$preflightBeforePath = Join-Path $OutputRoot 'preflight-before-install.json'
$preflightPath = Join-Path $OutputRoot 'preflight.json'
$resultPath = Join-Path $OutputRoot 'INSTALL-SMOKE-RESULT.json'

$sourcePreflightExit = $null
if (Test-Path -LiteralPath $sourceHelper -PathType Leaf) {
    & $sourceHelper preflight --install-dir $installDir --data-dir $dataDir --free-space-path $env:ProgramData --min-free-bytes 1 --port 54329 2> (Join-Path $OutputRoot 'preflight-before-install.error') |
        Set-Content -LiteralPath $preflightBeforePath -Encoding utf8
    $sourcePreflightExit = $LASTEXITCODE
}

$installExit = Invoke-ElevatedExecutable -FilePath $InstallerPath -Arguments @('/S', "/D=$installDir") -TaskRoot $OutputRoot
if ($installExit -ne 0) {
    $issueCodes = Get-PreflightIssueCodes -FallbackPath $preflightBeforePath
    $issueSummary = ($issueCodes -join ', ')
    throw "Instalação elevada retornou código $installExit; preflight: $issueSummary."
}
foreach ($name in @('Majucau Financial Intelligence.exe', 'majucau-worker.exe', 'installer-helper.exe', 'uninstall.exe')) {
    Assert-Condition (Test-Path -LiteralPath (Join-Path $installDir $name) -PathType Leaf) "Arquivo ausente após instalação: $name"
}

& $helper preflight --install-dir $installDir --data-dir $dataDir --free-space-path $OutputRoot --min-free-bytes 1 --port 55439 |
    Set-Content -LiteralPath $preflightPath -Encoding utf8
$preflightExit = $LASTEXITCODE
Assert-Condition ($preflightExit -in @(0, 2)) "Preflight pós-instalação retornou código inesperado: $preflightExit"
$preflight = (Read-BoundedTextFile -Path $preflightPath -MaxBytes 262144) | ConvertFrom-Json
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
    source_preflight_exit_code = $sourcePreflightExit
    uninstall_exit_code = $uninstallExit
    installer = [System.IO.Path]::GetFileName($InstallerPath)
    mode = 'scheduled-task-runlevel-highest'
}
$result | ConvertTo-Json -Depth 6 | Set-Content -LiteralPath $resultPath -Encoding utf8
Write-Output "Installer smoke PASS: $resultPath"

