[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
Import-Module (Join-Path $PSScriptRoot 'SafeCleanup.psm1') -Force
Import-Module (Join-Path $PSScriptRoot 'ArtifactManifest.psm1') -Force
$projectRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
$artifactRoot = Join-Path $projectRoot 'artifacts'
$windowsRoot = Join-Path $artifactRoot 'windows'
$stageRoot = Join-Path $artifactRoot 'cache\portable-stage'
$bundleName = 'majucau-windows-x64-portable.zip'
$bundlePath = Join-Path $windowsRoot $bundleName

$required = @(
    'majucau.exe',
    'majucau-worker.exe',
    'installer-helper.exe'
)
$releaseManifest = Join-Path $windowsRoot 'release-manifest.json'
$releaseMigrations = Join-Path $windowsRoot 'migrations'

if (-not (Test-Path -LiteralPath $windowsRoot -PathType Container)) {
    throw "Diretório de artefatos Windows ausente: $windowsRoot"
}
SafeCleanup\Assert-MajucauProjectPath -ProjectRoot $projectRoot -Path $windowsRoot -ExpectedType Directory
SafeCleanup\Assert-MajucauProjectPath -ProjectRoot $projectRoot -Path $stageRoot -ExpectedType Directory

foreach ($fileName in $required) {
    $filePath = Join-Path $windowsRoot $fileName
    if (-not (Test-Path -LiteralPath $filePath -PathType Leaf)) {
        throw "Artefato obrigatório ausente: $fileName"
    }
}
if (-not (Test-Path -LiteralPath $releaseManifest -PathType Leaf)) {
    throw "Manifesto de release ausente: $releaseManifest"
}
if (-not (Test-Path -LiteralPath $releaseMigrations -PathType Container)) {
    throw "Migrations do manifesto ausentes: $releaseMigrations"
}

SafeCleanup\Remove-MajucauProjectDirectory -ProjectRoot $projectRoot -Path $stageRoot -ExpectedRelativePath 'artifacts\cache\portable-stage'
try {
New-Item -ItemType Directory -Force -Path $stageRoot | Out-Null

foreach ($fileName in $required) {
    Copy-Item -LiteralPath (Join-Path $windowsRoot $fileName) -Destination $stageRoot
}
Copy-Item -LiteralPath $releaseManifest -Destination (Join-Path $stageRoot 'release-manifest.json')
New-Item -ItemType Directory -Force -Path (Join-Path $stageRoot 'migrations') | Out-Null
Copy-Item -Path (Join-Path $releaseMigrations '*') -Destination (Join-Path $stageRoot 'migrations') -Force

$readme = @'
Majucau Financial Intelligence — bundle Windows x64

Este arquivo é um bundle portátil para validação local e diagnóstico.
Ele NÃO substitui o instalador NSIS assinado do gate G7 e não deve ser tratado
como release de produção. O worker e o helper estão incluídos para inspeção.

Para uma validação rápida, execute installer-helper.exe preflight em um
PowerShell elevado e abra majucau.exe. O primeiro uso ainda depende dos
requisitos descritos no README do repositório, incluindo WebView2 e o
PostgreSQL local dedicado quando o fluxo correspondente estiver habilitado.

O pacote também contém release-manifest.json e as migrations usadas para
verificação determinística do conteúdo. Não mova credenciais, tokens ou dados reais para este diretório. Segredos são
configurados pela tela de integrações e protegidos pelo worker.
'@
Set-Content -LiteralPath (Join-Path $stageRoot 'README-PORTABLE.txt') -Value $readme -Encoding utf8

Get-ChildItem -LiteralPath $stageRoot -File |
    Get-FileHash -Algorithm SHA256 |
    ForEach-Object { '{0}  {1}' -f $_.Hash.ToLowerInvariant(), $_.Path.Substring($stageRoot.Length + 1) } |
    Set-Content -LiteralPath (Join-Path $stageRoot 'SHA256SUMS.txt') -Encoding ascii

SafeCleanup\Assert-MajucauProjectPath -ProjectRoot $projectRoot -Path $bundlePath -ExpectedType File
SafeCleanup\Remove-MajucauProjectFile -ProjectRoot $projectRoot -Path $bundlePath -ExpectedRelativePath 'artifacts\windows\majucau-windows-x64-portable.zip'
Compress-Archive -Path (Join-Path $stageRoot '*') -DestinationPath $bundlePath -CompressionLevel Optimal

$manifest = Join-Path $artifactRoot 'MANIFEST.txt'
Write-MajucauArtifactManifest -ProjectRoot $projectRoot -ManifestPath $manifest

Write-Host "Bundle portátil criado em: $bundlePath"
} finally {
    SafeCleanup\Remove-MajucauProjectDirectory -ProjectRoot $projectRoot -Path $stageRoot -ExpectedRelativePath 'artifacts\cache\portable-stage'
}
