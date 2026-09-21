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
$migrationSource = Join-Path $projectRoot 'database\migrations'
$migrationRoot = Join-Path $windowsRoot 'migrations'
$manifestPath = Join-Path $windowsRoot 'release-manifest.json'

$requiredBinaries = @('majucau.exe', 'majucau-worker.exe', 'installer-helper.exe')
if (-not (Test-Path -LiteralPath $windowsRoot -PathType Container)) {
    New-Item -ItemType Directory -Force -Path $windowsRoot | Out-Null
}
foreach ($fileName in $requiredBinaries) {
    if (-not (Test-Path -LiteralPath (Join-Path $windowsRoot $fileName) -PathType Leaf)) {
        throw "Artefato obrigatório ausente para o manifesto: $fileName"
    }
}

if (-not (Test-Path -LiteralPath $migrationSource -PathType Container)) {
    throw "Diretório de migrations ausente: $migrationSource"
}
if (Test-Path -LiteralPath $migrationRoot) {
    Remove-Item -LiteralPath $migrationRoot -Recurse -Force
}
New-Item -ItemType Directory -Force -Path $migrationRoot | Out-Null

$migrationFiles = @(Get-ChildItem -LiteralPath $migrationSource -Filter '*.up.sql' -File | Sort-Object Name)
if ($migrationFiles.Count -eq 0) {
    throw 'Nenhuma migration .up.sql encontrada para o manifesto.'
}
$migrationEntries = @()
foreach ($migration in $migrationFiles) {
    Copy-Item -LiteralPath $migration.FullName -Destination (Join-Path $migrationRoot $migration.Name) -Force
    $copy = Get-FileHash -Algorithm SHA256 -LiteralPath (Join-Path $migrationRoot $migration.Name)
    $migrationEntries += [ordered]@{
        path = "migrations/$($migration.Name)"
        sha256 = $copy.Hash.ToLowerInvariant()
        size_bytes = [int64]$migration.Length
    }
}

$artifactEntries = @()
foreach ($fileName in $requiredBinaries) {
    $file = Get-FileHash -Algorithm SHA256 -LiteralPath (Join-Path $windowsRoot $fileName)
    $info = Get-Item -LiteralPath (Join-Path $windowsRoot $fileName)
    $artifactEntries += [ordered]@{
        path = $fileName
        sha256 = $file.Hash.ToLowerInvariant()
        size_bytes = [int64]$info.Length
    }
}

$manifest = [ordered]@{
    manifest_version = '1'
    app_version = '0.1.0'
    schema_version = '1.0'
    min_schema_version = '1.0'
    max_schema_version = '1.0'
    migrations = @($migrationEntries)
    artifacts = @($artifactEntries)
}
$json = $manifest | ConvertTo-Json -Depth 8
$utf8NoBom = New-Object System.Text.UTF8Encoding($false)
[System.IO.File]::WriteAllText($manifestPath, $json, $utf8NoBom)
Write-Host "Manifesto de release criado em: $manifestPath"
