[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
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

if (Test-Path -LiteralPath $stageRoot) {
    Remove-Item -LiteralPath $stageRoot -Recurse -Force
}
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

if (Test-Path -LiteralPath $bundlePath) {
    Remove-Item -LiteralPath $bundlePath -Force
}
Compress-Archive -Path (Join-Path $stageRoot '*') -DestinationPath $bundlePath -CompressionLevel Optimal

$manifest = Join-Path $artifactRoot 'MANIFEST.txt'
$lines = @(
    'Majucau Financial Intelligence artifacts',
    "Gerado em: $(Get-Date -Format o)",
    "Origem do código: $projectRoot",
    "Arquivos finais: $windowsRoot",
    ''
)
Get-ChildItem -LiteralPath $artifactRoot -Recurse -File |
    Where-Object { $_.FullName -ne $manifest -and $_.FullName -notlike "$stageRoot*" } |
    Sort-Object FullName |
    ForEach-Object { $lines += "$($_.FullName.Substring($artifactRoot.Length + 1))  $($_.Length) bytes" }
Set-Content -LiteralPath $manifest -Value $lines -Encoding utf8

Remove-Item -LiteralPath $stageRoot -Recurse -Force
Write-Host "Bundle portátil criado em: $bundlePath"
