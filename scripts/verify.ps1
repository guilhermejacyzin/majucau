[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSScriptRoot
$goCommand = if ($env:MAJUCAU_GO) { $env:MAJUCAU_GO } else { (Get-Command go -ErrorAction Stop).Source }
$env:Path = "$(Split-Path -Parent $goCommand);$env:Path"
$sqlcCommand = if ($env:MAJUCAU_SQLC) { $env:MAJUCAU_SQLC } else { (Get-Command sqlc -ErrorAction Stop).Source }
$vulncheckCommand = if ($env:MAJUCAU_GOVULNCHECK) { $env:MAJUCAU_GOVULNCHECK } else { (Get-Command govulncheck -ErrorAction Stop).Source }
$goPackages = @('.', './cmd/...', './database/...', './internal/...')

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

Push-Location $projectRoot
try {
    Write-Host 'go test . ./cmd/... ./database/... ./internal/...'
    Invoke-NativeChecked -Command $goCommand -Arguments (@('test') + $goPackages)

    Write-Host 'go vet . ./cmd/... ./database/... ./internal/...'
    Invoke-NativeChecked -Command $goCommand -Arguments (@('vet') + $goPackages)

    Write-Host 'govulncheck . ./cmd/... ./database/... ./internal/...'
    Invoke-NativeChecked -Command $vulncheckCommand -Arguments $goPackages

    & (Join-Path $PSScriptRoot 'verify-migrations.ps1')

    $beforeSqlc = Get-ChildItem -LiteralPath (Join-Path $projectRoot 'database\gen') -File |
        Sort-Object Name |
        ForEach-Object { "$($_.Name):$((Get-FileHash -Algorithm SHA256 -LiteralPath $_.FullName).Hash)" }
    Invoke-NativeChecked -Command $sqlcCommand -Arguments @('generate', '-f', (Join-Path $projectRoot 'database\sqlc.yaml'))
    $afterSqlc = Get-ChildItem -LiteralPath (Join-Path $projectRoot 'database\gen') -File |
        Sort-Object Name |
        ForEach-Object { "$($_.Name):$((Get-FileHash -Algorithm SHA256 -LiteralPath $_.FullName).Hash)" }
    if (Compare-Object $beforeSqlc $afterSqlc) {
        throw 'sqlc generate alterou o código gerado.'
    }
    Write-Host 'sqlc generate determinístico: PASS'

    & (Join-Path $PSScriptRoot 'scan-secrets.ps1')

    Push-Location (Join-Path $projectRoot 'frontend')
    try {
        Write-Host 'npm run lint'
        Invoke-NativeChecked -Command 'npm' -Arguments @('run', 'lint')

        Write-Host 'npm run typecheck'
        Invoke-NativeChecked -Command 'npm' -Arguments @('run', 'typecheck')

        Write-Host 'npm run test:run'
        Invoke-NativeChecked -Command 'npm' -Arguments @('run', 'test:run')

        Write-Host 'npm run build'
        Invoke-NativeChecked -Command 'npm' -Arguments @('run', 'build')

        Write-Host 'npm audit'
        Invoke-NativeChecked -Command 'npm' -Arguments @('audit')
    }
    finally {
        Pop-Location
    }
}
finally {
    Pop-Location
}

Write-Host 'Verificações locais concluídas.'
