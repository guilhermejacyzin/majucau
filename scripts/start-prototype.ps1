[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$projectRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
$artifactRoot = Join-Path $projectRoot 'artifacts'
$instanceRoot = Join-Path $artifactRoot 'retained\postgresql-dev'
$dataRoot = Join-Path $instanceRoot 'data'
$credentialsRoot = Join-Path $instanceRoot 'credentials'
$credentialPath = Join-Path $credentialsRoot 'runtime.dpapi'
$vaultPath = Join-Path $credentialsRoot 'worker-vault'
$artifactCache = Join-Path $artifactRoot 'cache'
$goToolBinRoot = Join-Path $artifactRoot 'tools\go\bin'
$workerPath = Join-Path $artifactCache 'dev\majucau-worker.exe'
$workerStdoutPath = Join-Path $instanceRoot 'worker.stdout.log'
$workerStderrPath = Join-Path $instanceRoot 'worker.stderr.log'
$databaseLogPath = Join-Path $instanceRoot 'postgresql.log'
$frontendDist = Join-Path $projectRoot 'frontend\dist'
$startedDatabase = $false
$workerProcess = $null
$originalEnvironment = @{}
$cleanupErrors = [System.Collections.Generic.List[string]]::new()

Import-Module (Join-Path $PSScriptRoot 'SafeCleanup.psm1') -Force
Import-Module (Join-Path $PSScriptRoot 'BoundedText.psm1') -Force
Assert-MajucauProjectPath -ProjectRoot $projectRoot -Path $artifactRoot -ExpectedType Directory
Assert-MajucauProjectPath -ProjectRoot $projectRoot -Path $instanceRoot -ExpectedType Directory
Assert-MajucauProjectPath -ProjectRoot $projectRoot -Path $dataRoot -ExpectedType Directory

function Protect-PrototypeSecretParent {
    [CmdletBinding()]
    param([Parameter(Mandatory)] [string]$Path)

    Assert-MajucauProjectPath -ProjectRoot $projectRoot -Path $Path -ExpectedType Directory
    $userSid = [System.Security.Principal.WindowsIdentity]::GetCurrent().User.Value
    if ([string]::IsNullOrWhiteSpace($userSid)) {
        throw 'A identidade Windows do usuário não pôde ser confirmada para proteger o cofre local.'
    }

    $icacls = Join-Path $env:SystemRoot 'System32\icacls.exe'
    if (-not (Test-Path -LiteralPath $icacls -PathType Leaf)) {
        throw 'A ferramenta Windows para proteger a pasta do cofre não está disponível.'
    }
    $grants = @(
        ('*{0}:(OI)(CI)F' -f $userSid),
        '*S-1-5-18:(OI)(CI)F',
        '*S-1-5-32-544:(OI)(CI)F'
    )
    & $icacls $Path '/inheritance:r' '/grant:r' @grants *> $null
    if ($LASTEXITCODE -ne 0) {
        throw 'Não foi possível restringir o acesso à pasta pai do cofre local.'
    }

    $acl = Get-Acl -LiteralPath $Path -ErrorAction Stop
    $rules = @($acl.Access)
    if (-not $acl.AreAccessRulesProtected -or $rules.Count -ne 3) {
        throw 'A proteção de acesso à pasta pai do cofre local não corresponde à configuração segura.'
    }
    $expectedSids = @($userSid, 'S-1-5-18', 'S-1-5-32-544')
    $expectedInheritance = [System.Security.AccessControl.InheritanceFlags]::ContainerInherit -bor [System.Security.AccessControl.InheritanceFlags]::ObjectInherit
    $actualSids = [System.Collections.Generic.HashSet[string]]::new([System.StringComparer]::OrdinalIgnoreCase)
    foreach ($rule in $rules) {
        $ruleSid = if ($rule.IdentityReference -is [System.Security.Principal.SecurityIdentifier]) {
            $rule.IdentityReference.Value
        } else {
            $rule.IdentityReference.Translate([System.Security.Principal.SecurityIdentifier]).Value
        }
        if ($ruleSid -notin $expectedSids -or
            -not $actualSids.Add($ruleSid) -or
            $rule.AccessControlType -ne [System.Security.AccessControl.AccessControlType]::Allow -or
            $rule.FileSystemRights -ne [System.Security.AccessControl.FileSystemRights]::FullControl -or
            $rule.IsInherited -or
            $rule.InheritanceFlags -ne $expectedInheritance -or
            $rule.PropagationFlags -ne [System.Security.AccessControl.PropagationFlags]::None) {
            throw 'A proteção de acesso à pasta pai do cofre local não corresponde à configuração segura.'
        }
    }
    if ($actualSids.Count -ne $expectedSids.Count) {
        throw 'A proteção de acesso à pasta pai do cofre local não corresponde à configuração segura.'
    }
}

foreach ($name in @('MAJUCAU_DATABASE_URL', 'MAJUCAU_SECRET_DIR', 'GOCACHE', 'GOMODCACHE', 'GOTMPDIR', 'GOBIN', 'npm_config_cache', 'Path')) {
    $originalEnvironment[$name] = [Environment]::GetEnvironmentVariable($name, 'Process')
}

try {
    if (-not (Test-Path -LiteralPath (Join-Path $dataRoot 'PG_VERSION') -PathType Leaf) -or
        (Read-BoundedTextFile -Path (Join-Path $dataRoot 'PG_VERSION') -MaxBytes 128).Trim() -ne '18') {
        throw 'O banco de desenvolvimento isolado PostgreSQL 18 não está preparado. Nenhum banco existente foi alterado.'
    }
    if (-not (Test-Path -LiteralPath $credentialPath -PathType Leaf)) {
        throw 'A credencial local protegida por DPAPI não foi encontrada. Ela não é armazenada no Git.'
    }
    if (-not (Test-Path -LiteralPath (Join-Path $dataRoot 'postgresql.conf') -PathType Leaf)) {
        throw 'A configuração do banco local não foi encontrada.'
    }

    $listenSetting = Select-String -LiteralPath (Join-Path $dataRoot 'postgresql.conf') -Pattern "^\s*listen_addresses\s*=\s*'127\.0\.0\.1'\s*$" -Quiet
    $portSetting = Select-String -LiteralPath (Join-Path $dataRoot 'postgresql.conf') -Pattern '^\s*port\s*=\s*55438\s*$' -Quiet
    if (-not $listenSetting -or -not $portSetting) {
        throw 'O banco local não está limitado a 127.0.0.1:55438. A inicialização foi interrompida por segurança.'
    }

    $programFilesRoot = [Environment]::GetEnvironmentVariable('ProgramFiles')
    $pgBin = if ($env:MAJUCAU_PG_BIN) { $env:MAJUCAU_PG_BIN } elseif ($programFilesRoot) { Join-Path $programFilesRoot 'PostgreSQL\18\bin' } else { '' }
    $pgCtl = Join-Path $pgBin 'pg_ctl.exe'
    if (-not (Test-Path -LiteralPath $pgCtl -PathType Leaf)) {
        throw 'PostgreSQL 18 não foi localizado. Instale-o ou configure MAJUCAU_PG_BIN para a pasta bin da versão 18.'
    }
    $pgVersion = (& $pgCtl --version | Out-String).Trim()
    if ($LASTEXITCODE -ne 0 -or $pgVersion -notmatch '\b18\.') {
        throw 'A ferramenta localizada não é PostgreSQL 18; o banco de outro projeto não será usado.'
    }

    $goCommand = if ($env:MAJUCAU_GO) { $env:MAJUCAU_GO } else { (Get-Command go -ErrorAction Stop).Source }
    $wailsLocal = Join-Path $goToolBinRoot 'wails.exe'
    $wailsCache = Join-Path $artifactCache 'bin\wails.exe'
    $wailsCommand = if ($env:MAJUCAU_WAILS) {
        $env:MAJUCAU_WAILS
    } elseif (Test-Path -LiteralPath $wailsLocal -PathType Leaf) {
        $wailsLocal
    } elseif (Test-Path -LiteralPath $wailsCache -PathType Leaf) {
        $wailsCache
    } else {
        (Get-Command wails -ErrorAction Stop).Source
    }

    if (-not (Test-Path -LiteralPath (Join-Path $projectRoot 'frontend\node_modules') -PathType Container)) {
        throw 'As dependências da interface não foram instaladas. Execute scripts/bootstrap.ps1 antes de abrir o protótipo.'
    }

    $env:GOCACHE = Join-Path $artifactCache 'go-build'
    $goModCacheRoot = Join-Path $artifactCache 'go-mod'
    $env:GOMODCACHE = $goModCacheRoot
    $env:GOTMPDIR = Join-Path $artifactCache 'go-tmp'
    $env:GOBIN = $goToolBinRoot
    $env:npm_config_cache = Join-Path $artifactCache 'npm'
    $env:Path = "$(Split-Path -Parent $goCommand);$goToolBinRoot;$(Join-Path $artifactCache 'bin');$env:Path"
    New-Item -ItemType Directory -Force -Path $env:GOCACHE, $goModCacheRoot, $env:GOTMPDIR, $goToolBinRoot, (Split-Path -Parent $workerPath), $vaultPath | Out-Null
    Protect-PrototypeSecretParent -Path $credentialsRoot

    $goModCacheBoundary = Join-Path $goModCacheRoot 'go.mod'
    if (-not (Test-Path -LiteralPath $goModCacheBoundary)) {
        @('module majucau.local/artifact-cache', '', 'go 1.26.0') |
            Set-Content -LiteralPath $goModCacheBoundary -Encoding utf8
    }

    Push-Location $projectRoot
    try {
        Write-Host 'Preparando o worker do protótipo...'
        & $goCommand build -trimpath -o $workerPath ./cmd/worker
        if ($LASTEXITCODE -ne 0) { throw 'Não foi possível preparar o worker. Nenhum teste do produto foi executado.' }
    }
    finally {
        Pop-Location
    }

    & $pgCtl status -D $dataRoot *> $null
    if ($LASTEXITCODE -ne 0) {
        Write-Host 'Iniciando o banco de desenvolvimento isolado em 127.0.0.1:55438...'
        & $pgCtl start -D $dataRoot -l $databaseLogPath -w -t 30
        if ($LASTEXITCODE -ne 0) { throw 'O banco isolado não iniciou. O serviço PostgreSQL da porta 5432 não foi usado nem alterado.' }
        $startedDatabase = $true
    }

    $encryptedPassword = (Read-BoundedTextFile -Path $credentialPath -MaxBytes 65536).Trim()
    $securePassword = ConvertTo-SecureString -String $encryptedPassword
    $runtimePassword = (New-Object System.Net.NetworkCredential('', $securePassword)).Password
    $encodedPassword = [Uri]::EscapeDataString($runtimePassword)
    $env:MAJUCAU_DATABASE_URL = "postgres://majucau_runtime:$encodedPassword@127.0.0.1:55438/majucau_dev?sslmode=disable"
    $env:MAJUCAU_SECRET_DIR = $vaultPath

    $workerProcess = Start-Process -FilePath $workerPath -ArgumentList @('--console') -WorkingDirectory $projectRoot -WindowStyle Hidden -PassThru -RedirectStandardOutput $workerStdoutPath -RedirectStandardError $workerStderrPath
    Start-Sleep -Seconds 1
    $workerProcess.Refresh()
    if ($workerProcess.HasExited) { throw 'O worker encerrou ao iniciar. Consulte os logs locais sanitizados.' }

    Write-Host 'Abrindo o protótipo local com o banco Majucau isolado e sem dados financeiros de exemplo.'
    Write-Host 'Feche a janela Majucau para encerrar esta execução.'
    & $wailsCommand dev
    if ($LASTEXITCODE -ne 0) { throw 'O aplicativo encerrou com erro. Consulte a mensagem acima e os logs locais sanitizados.' }
}
finally {
    if ($workerProcess) {
        try {
            $workerProcess.Refresh()
            if (-not $workerProcess.HasExited) {
            $workerProcess.Kill()
            if (-not $workerProcess.WaitForExit(10000)) {
                throw 'O worker não encerrou no prazo previsto.'
            }
            }
        }
    catch {
        # A verificação seguinte distingue uma falha real de uma saída concorrente do worker.
    }
    }

    $workerStillRunning = $false
    if ($workerProcess) {
        try {
            $workerProcess.Refresh()
            $workerStillRunning = -not $workerProcess.HasExited
        }
        catch {
        try {
            $liveWorker = Get-CimInstance Win32_Process -Filter "ProcessId = $($workerProcess.Id)" -ErrorAction Stop
            $workerStillRunning = $null -ne $liveWorker -and
                [string]::Equals($liveWorker.ExecutablePath, $workerPath, [System.StringComparison]::OrdinalIgnoreCase)
        }
        catch {
            $workerStillRunning = $true
        }
    }
    if ($workerStillRunning) {
        $cleanupErrors.Add("Não foi possível encerrar o worker local (PID $($workerProcess.Id)); os arquivos temporários serão preservados.")
        }
    }

    if ($startedDatabase) {
        if ($workerStillRunning) {
            $cleanupErrors.Add('O banco isolado foi mantido ativo porque o worker não encerrou.')
        }
        else {
            try {
                & $pgCtl stop -D $dataRoot -m fast -w
                if ($LASTEXITCODE -ne 0) { throw 'pg_ctl retornou erro.' }
            }
            catch {
                $cleanupErrors.Add('O banco isolado não encerrou corretamente; confira seu estado antes de limpar arquivos.')
            }
        }
    }

    foreach ($name in $originalEnvironment.Keys) {
        try {
            [Environment]::SetEnvironmentVariable($name, $originalEnvironment[$name], 'Process')
        }
        catch {
            $cleanupErrors.Add('Não foi possível restaurar uma variável temporária do ambiente.')
        }
    }

    $runtimePassword = $null
    $encryptedPassword = $null
    $encodedPassword = $null
    if ($securePassword) {
        $securePassword.Dispose()
        $securePassword = $null
    }

    if (-not $workerStillRunning) {
        try {
            if (Test-Path -LiteralPath $workerPath -PathType Leaf) {
                Remove-MajucauProjectFile -ProjectRoot $projectRoot -Path $workerPath -ExpectedRelativePath 'artifacts\cache\dev\majucau-worker.exe'
            }
            if (Test-Path -LiteralPath $frontendDist -PathType Container) {
                Remove-MajucauProjectDirectory -ProjectRoot $projectRoot -Path $frontendDist -ExpectedRelativePath 'frontend\dist'
            }
        }
        catch {
            $cleanupErrors.Add('A limpeza dos arquivos temporários do protótipo falhou; os caminhos continuam registrados no projeto.')
        }
    }
    else {
        $cleanupErrors.Add('Worker e arquivos de suporte foram preservados porque há um processo ativo.')
    }

    if ($cleanupErrors.Count -gt 0) {
        throw "Limpeza incompleta: $($cleanupErrors -join ' ') Consulte o estado local antes de uma nova execução."
    }
}
