[CmdletBinding()]
param(
    [Parameter(Mandatory)] [string]$OutputRoot
)

$ErrorActionPreference = 'Stop'
Import-Module (Join-Path $PSScriptRoot 'WorkerServiceSmokeCleanup.psm1') -Force
$projectRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
$OutputRoot = [System.IO.Path]::GetFullPath($OutputRoot)
$expectedOutputRoot = [System.IO.Path]::GetFullPath((Join-Path $projectRoot 'artifacts\verification\worker-service-smoke'))
if (-not [string]::Equals($OutputRoot, $expectedOutputRoot, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw 'OutputRoot precisa ser artifacts/verification/worker-service-smoke.'
}
if (-not (Test-Path -LiteralPath $env:ProgramFiles -PathType Container) -or -not (Test-Path -LiteralPath $env:ProgramData -PathType Container)) {
    throw 'Program Files ou ProgramData não estão disponíveis.'
}

$helper = (Resolve-Path -LiteralPath (Join-Path $projectRoot 'build\bin\installer-helper.exe')).Path
$worker = (Resolve-Path -LiteralPath (Join-Path $projectRoot 'build\bin\majucau-worker.exe')).Path
$runId = [guid]::NewGuid().ToString('N')
$installRoot = [System.IO.Path]::GetFullPath((Join-Path $env:ProgramFiles "Majucau-Worker-Smoke-$runId"))
$dataRoot = [System.IO.Path]::GetFullPath((Join-Path $env:ProgramData 'Majucau'))
$workerServiceRoot = Join-Path $installRoot 'worker\test'
$serviceWasPresent = $null -ne (Get-Service -Name 'MajucauWorker' -ErrorAction SilentlyContinue)
$dataRootWasPresent = Test-Path -LiteralPath $dataRoot
if ($serviceWasPresent -or $dataRootWasPresent) {
    throw 'O teste foi interrompido para não tocar em uma instalação ou dado existente do Majucau.'
}

$testWorker = Join-Path $workerServiceRoot 'majucau-worker.exe'
$serviceRemoved = $false
$cleanupFailures = [System.Collections.Generic.List[string]]::new()

function Invoke-HelperJson {
    param([Parameter(Mandatory)] [string[]]$Arguments)
    $lines = @(& $helper @Arguments 2>$null)
    $exitCode = $LASTEXITCODE
    $raw = $lines -join "`n"
    if ([string]::IsNullOrWhiteSpace($raw)) {
        throw "installer-helper não retornou JSON (código $exitCode)."
    }
    $result = $raw | ConvertFrom-Json
    if ($result.exit_code -ne $exitCode) {
        throw 'Código do processo diverge do JSON de resultado.'
    }
    return $result
}

function Get-AclSidSet {
    param([Parameter(Mandatory)] [string]$Path)
    $acl = Get-Acl -LiteralPath $Path
    if (-not $acl.AreAccessRulesProtected) {
        throw "ACL não protegida em $Path."
    }
    $sids = @(
        $acl.Access | ForEach-Object {
            if ($_.IdentityReference -is [System.Security.Principal.SecurityIdentifier]) {
                $_.IdentityReference.Value
            } else {
                $_.IdentityReference.Translate([System.Security.Principal.SecurityIdentifier]).Value
            }
        }
    )
    return @($sids | Sort-Object -Unique)
}

try {
    New-Item -ItemType Directory -Force -Path $workerServiceRoot | Out-Null
    Copy-Item -LiteralPath $worker -Destination $testWorker
    $install = Invoke-HelperJson @(
        'install-service', '--install-dir', $installRoot,
        '--worker-path', $testWorker, '--data-dir', $dataRoot
    )
    if ($install.status -ne 'SERVICE_REGISTERED_AND_RUNNING' -or $install.exit_code -ne 0) {
        $diagnostics = [ordered]@{
            installer = $install
            service = $null
            recent_service_events = @()
        }
        $serviceInfo = Get-CimInstance -ClassName Win32_Service -Filter "Name='MajucauWorker'" -ErrorAction SilentlyContinue
        if ($null -ne $serviceInfo) {
            $diagnostics.service = $serviceInfo | Select-Object State, Status, StartName, ExitCode, ServiceSpecificExitCode
        }
        try {
            $diagnostics.recent_service_events = @(
                Get-WinEvent -FilterHashtable @{
                    LogName = 'System'
                    ProviderName = 'Service Control Manager'
                    StartTime = (Get-Date).AddMinutes(-3)
                } -MaxEvents 30 -ErrorAction SilentlyContinue |
                    Where-Object { $_.Message -match 'MajucauWorker' } |
                    Select-Object -First 3 TimeCreated, Id, LevelDisplayName, Message
            )
        } catch { }
        throw "O serviço não foi registrado e iniciado: $($diagnostics | ConvertTo-Json -Depth 5 -Compress)"
    }
    $service = Get-Service -Name 'MajucauWorker' -ErrorAction Stop
    if ($service.Status -ne 'Running') { throw "Estado do worker inesperado: $($service.Status)." }
    $serviceKey = Get-ItemProperty -LiteralPath 'HKLM:\SYSTEM\CurrentControlSet\Services\MajucauWorker'
    if ($serviceKey.ObjectName -ne 'NT SERVICE\MajucauWorker' -or [int]$serviceKey.Start -ne 2 -or [int]$serviceKey.DelayedAutoStart -ne 1 -or [int]$serviceKey.ServiceSidType -ne 3) {
        throw 'A conta, inicialização ou SID restrito do serviço diverge da configuração aprovada.'
    }
    $serviceSid = ([System.Security.Principal.NTAccount]'NT SERVICE\MajucauWorker').Translate([System.Security.Principal.SecurityIdentifier]).Value
    foreach ($path in @((Join-Path $dataRoot 'WorkerSecrets'), (Join-Path $dataRoot 'WorkerLogs'))) {
        if (-not (Test-Path -LiteralPath $path -PathType Container)) { throw "Diretório protegido ausente: $path" }
        $sids = Get-AclSidSet -Path $path
        $expected = @($serviceSid, 'S-1-5-18', 'S-1-5-32-544')
        $expected = @($expected | Sort-Object -Unique)
        if (Compare-Object -ReferenceObject $expected -DifferenceObject $sids) {
            throw 'A ACL do diretório do worker não contém exatamente o SID do serviço, SYSTEM e Administradores.'
        }
    }
    if (-not (Test-Path -LiteralPath (Join-Path $dataRoot 'WorkerSecrets\Vault') -PathType Container) -or -not (Test-Path -LiteralPath (Join-Path $dataRoot 'WorkerLogs\Logs') -PathType Container)) {
        throw 'O worker não criou o cofre DPAPI e o diretório privado de logs.'
    }

    $parametersPath = 'Registry::HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\MajucauWorker\Parameters'
    $uiSid = (Get-ItemProperty -LiteralPath $parametersPath -Name UIUserSID -ErrorAction Stop).UIUserSID
    try { [void][System.Security.Principal.SecurityIdentifier]::new($uiSid) } catch { throw 'O SID da interface salvo para o serviço é inválido.' }
    $parametersAcl = Get-Acl -LiteralPath $parametersPath
    if (-not $parametersAcl.AreAccessRulesProtected) { throw 'A configuração de identidade do serviço está herdando permissões.' }
    $parameterRules = @($parametersAcl.GetAccessRules($true, $false, [System.Security.Principal.SecurityIdentifier]))
    if ($parameterRules.Count -ne 3) { throw 'As permissões da configuração de identidade do serviço não são exatas.' }
    $expectedRegistryRights = @{}
    $expectedRegistryRights[$serviceSid] = [int][Microsoft.Win32.RegistryRights]::ReadKey
    $expectedRegistryRights['S-1-5-18'] = [int][Microsoft.Win32.RegistryRights]::FullControl
    $expectedRegistryRights['S-1-5-32-544'] = [int][Microsoft.Win32.RegistryRights]::FullControl
    foreach ($rule in $parameterRules) {
        $sid = $rule.IdentityReference.Value
        if (-not $expectedRegistryRights.ContainsKey($sid) -or [int]$rule.RegistryRights -ne $expectedRegistryRights[$sid] -or $rule.AccessControlType -ne [System.Security.AccessControl.AccessControlType]::Allow) {
            throw 'A configuração do serviço concede permissões de registro inesperadas.'
        }
        $expectedRegistryRights.Remove($sid)
    }
    if ($expectedRegistryRights.Count -ne 0) { throw 'Faltam permissões necessárias para o serviço ler sua configuração.' }

    $remove = Invoke-HelperJson @('remove-service', '--install-dir', $installRoot)
    if ($remove.status -ne 'SERVICE_REMOVED' -or $remove.exit_code -ne 0) { throw 'O serviço não foi removido com sucesso.' }
    $serviceRemoved = $null -eq (Get-Service -Name 'MajucauWorker' -ErrorAction SilentlyContinue)
    if (-not $serviceRemoved) { throw 'O serviço permaneceu registrado após a remoção.' }
    if (-not (Test-Path -LiteralPath (Join-Path $dataRoot 'WorkerSecrets') -PathType Container) -or -not (Test-Path -LiteralPath (Join-Path $dataRoot 'WorkerLogs') -PathType Container)) {
        throw 'A remoção do serviço apagou diretórios de dados que deveriam ser preservados.'
    }

    New-Item -ItemType Directory -Force -Path $OutputRoot | Out-Null
    [ordered]@{
        schema_version = '1.0'
        status = 'PASS'
        service_account = 'NT SERVICE\MajucauWorker'
        service_sid_restricted = $true
        delayed_auto_start = $true
        protected_directories = @('WorkerSecrets', 'WorkerLogs')
        uninstall_preserved_data = $true
    } | ConvertTo-Json -Depth 4 | Set-Content -LiteralPath (Join-Path $OutputRoot 'WORKER-SERVICE-SMOKE.json') -Encoding utf8
}
finally {
    if (Get-Service -Name 'MajucauWorker' -ErrorAction SilentlyContinue) {
        try {
            $cleanup = Invoke-HelperJson @('remove-service', '--install-dir', $installRoot)
            $serviceRemoved = $cleanup.status -eq 'SERVICE_REMOVED' -and $cleanup.exit_code -eq 0
        } catch {
            Write-Warning 'Não foi possível retirar o serviço de smoke com segurança; seus arquivos e dados foram preservados.'
        }
    }
    if (-not (Get-Service -Name 'MajucauWorker' -ErrorAction SilentlyContinue)) {
        $programFilesRoot = [System.IO.Path]::GetFullPath($env:ProgramFiles).TrimEnd([char[]]@([System.IO.Path]::DirectorySeparatorChar, [System.IO.Path]::AltDirectorySeparatorChar))
        if ($installRoot.StartsWith($programFilesRoot + [System.IO.Path]::DirectorySeparatorChar, [System.StringComparison]::OrdinalIgnoreCase) -and (Test-Path -LiteralPath $installRoot)) {
            try {
                $installRootItem = Get-Item -LiteralPath $installRoot -Force -ErrorAction Stop
                if (-not $installRootItem.PSIsContainer -or ($installRootItem.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) { throw 'A raiz da instalação smoke não é uma pasta local segura.' }
                $expectedInstallPaths = @('worker', 'worker\test', 'worker\test\majucau-worker.exe')
                $installItems = @(Get-ChildItem -LiteralPath $installRoot -Force -Recurse -ErrorAction Stop)
                foreach ($item in $installItems) {
                    if (($item.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) { throw 'A pasta de instalação do smoke contém um link.' }
                    $relative = [System.IO.Path]::GetRelativePath($installRoot, $item.FullName).Replace('/', '\')
                    if ($relative -notin $expectedInstallPaths) { throw 'A pasta de instalação do smoke contém arquivos inesperados.' }
                }
                $workerFile = Join-Path $workerServiceRoot 'majucau-worker.exe'
                if (Test-Path -LiteralPath $workerFile -PathType Leaf) { Remove-Item -LiteralPath $workerFile -Force -ErrorAction Stop }
                foreach ($directory in @($workerServiceRoot, (Join-Path $installRoot 'worker'), $installRoot)) {
                    if (Test-Path -LiteralPath $directory -PathType Container) { Remove-Item -LiteralPath $directory -Force -ErrorAction Stop }
                }
            } catch {
                $cleanupFailures.Add("$installRoot ($($_.Exception.Message))")
            }
        }
        if (-not $dataRootWasPresent -and (Test-Path -LiteralPath $dataRoot)) {
            try { Remove-MajucauWorkerServiceSmokeData -DataRoot $dataRoot }
            catch { $cleanupFailures.Add("$dataRoot ($($_.Exception.Message))") }
        }
    } elseif (-not $serviceRemoved) {
        $cleanupFailures.Add("$installRoot; $dataRoot (o serviço de smoke continua registrado)")
    }
    if ($cleanupFailures.Count -gt 0) {
        Register-MajucauWorkerServiceSmokeRetention -OutputRoot $OutputRoot -Paths @($cleanupFailures) -Reason 'A limpeza validada recusou remover saída inesperada ou o serviço não pôde ser removido.'
        throw 'A limpeza do smoke falhou; veja RETAINED-CLEANUP.json nos outputs de verificação.'
    }
}
