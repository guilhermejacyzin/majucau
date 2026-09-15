[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$migrationRoot = Join-Path (Split-Path -Parent $PSScriptRoot) 'database\migrations'
$checksumPath = Join-Path $migrationRoot 'checksums.sha256'

if (-not (Test-Path -LiteralPath $checksumPath)) {
    throw 'Arquivo de checksums das migrations não encontrado.'
}

$expected = @{}
foreach ($line in Get-Content -LiteralPath $checksumPath) {
    if ($line -match '^([0-9a-f]{64})\s{2}(.+)$') {
        $expected[$Matches[2]] = $Matches[1]
    }
}

$migrations = Get-ChildItem -LiteralPath $migrationRoot -Filter '*.sql' -File
if ($migrations.Count -ne $expected.Count) {
    throw 'Quantidade de migrations diferente do manifesto de checksums.'
}

foreach ($migration in $migrations) {
    if (-not $expected.ContainsKey($migration.Name)) {
        throw "Migration sem checksum: $($migration.Name)"
    }
    $actual = (Get-FileHash -Algorithm SHA256 -LiteralPath $migration.FullName).Hash.ToLowerInvariant()
    if ($actual -ne $expected[$migration.Name]) {
        throw "Migration alterada após registro: $($migration.Name)"
    }
}

Write-Host 'Checksums de migrations: PASS'
