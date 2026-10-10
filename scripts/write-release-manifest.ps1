[CmdletBinding()]
param(
    [string]$ArtifactRoot = ''
)

$ErrorActionPreference = 'Stop'
Import-Module (Join-Path $PSScriptRoot 'SafeCleanup.psm1') -Force
$projectRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
if ([string]::IsNullOrWhiteSpace($ArtifactRoot)) {
    $ArtifactRoot = Join-Path $projectRoot 'artifacts'
}
$ArtifactRoot = [System.IO.Path]::GetFullPath($ArtifactRoot)
$expectedArtifactRoot = [System.IO.Path]::GetFullPath((Join-Path $projectRoot 'artifacts'))
$separatorCharacters = [char[]]@([System.IO.Path]::DirectorySeparatorChar, [System.IO.Path]::AltDirectorySeparatorChar)
if (-not [string]::Equals($ArtifactRoot.TrimEnd($separatorCharacters), $expectedArtifactRoot.TrimEnd($separatorCharacters), [System.StringComparison]::OrdinalIgnoreCase)) {
    throw 'ArtifactRoot precisa ser a pasta artifacts do projeto.'
}
$ArtifactRoot = $expectedArtifactRoot
$windowsRoot = Join-Path $ArtifactRoot 'windows'
$migrationSource = Join-Path $projectRoot 'database\migrations'
$migrationRoot = Join-Path $windowsRoot 'migrations'
$manifestPath = Join-Path $windowsRoot 'release-manifest.json'

$requiredBinaries = @('majucau.exe', 'majucau-worker.exe', 'installer-helper.exe')
Assert-MajucauProjectPath -ProjectRoot $projectRoot -Path $ArtifactRoot -ExpectedType Directory
Assert-MajucauProjectPath -ProjectRoot $projectRoot -Path $windowsRoot -ExpectedType Directory
if (-not (Test-Path -LiteralPath $windowsRoot -PathType Container)) {
    New-Item -ItemType Directory -Force -Path $windowsRoot | Out-Null
}
Assert-MajucauProjectPath -ProjectRoot $projectRoot -Path $windowsRoot -ExpectedType Directory
foreach ($fileName in $requiredBinaries) {
    if (-not (Test-Path -LiteralPath (Join-Path $windowsRoot $fileName) -PathType Leaf)) {
        throw "Artefato obrigatório ausente para o manifesto: $fileName"
    }
}

if (-not (Test-Path -LiteralPath $migrationSource -PathType Container)) {
    throw "Diretório de migrations ausente: $migrationSource"
}
Assert-MajucauProjectPath -ProjectRoot $projectRoot -Path $migrationSource -ExpectedType Directory
Remove-MajucauProjectDirectory -ProjectRoot $projectRoot -Path $migrationRoot -ExpectedRelativePath 'artifacts\windows\migrations'
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
    $binaryPath = Join-Path $windowsRoot $fileName
    Assert-MajucauProjectPath -ProjectRoot $projectRoot -Path $binaryPath -ExpectedType File
    $file = Get-FileHash -Algorithm SHA256 -LiteralPath $binaryPath
    $info = Get-Item -LiteralPath $binaryPath
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
Assert-MajucauProjectPath -ProjectRoot $projectRoot -Path $manifestPath -ExpectedType File
[System.IO.File]::WriteAllText($manifestPath, $json, $utf8NoBom)
Write-Host "Manifesto de release criado em: $manifestPath"
