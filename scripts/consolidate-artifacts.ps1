[CmdletBinding()]
param(
    [switch]$KeepSourceOutputs
)

$ErrorActionPreference = 'Stop'
$projectRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
$artifactRoot = Join-Path $projectRoot 'artifacts'
$windowsRoot = Join-Path $artifactRoot 'windows'

New-Item -ItemType Directory -Force -Path $windowsRoot | Out-Null

$copyIfPresent = @(
    @{ Source = 'build\bin\majucau.exe'; Destination = 'windows\majucau.exe' },
    @{ Source = 'build\bin\majucau-worker.exe'; Destination = 'windows\majucau-worker.exe' },
    @{ Source = 'build\bin\installer-helper.exe'; Destination = 'windows\installer-helper.exe' },
    @{ Source = 'build\bin\*-installer.exe'; Destination = 'windows' }
)

foreach ($item in $copyIfPresent) {
    $source = Join-Path $projectRoot $item.Source
    if ($item.Source.Contains('*')) {
        $sourceDirectory = Join-Path $projectRoot (Split-Path $item.Source)
        $sourceFilter = Split-Path $item.Source -Leaf
        $files = Get-ChildItem -LiteralPath $sourceDirectory -Filter $sourceFilter -File -ErrorAction SilentlyContinue
        foreach ($file in $files) {
            Copy-Item -LiteralPath $file.FullName -Destination $windowsRoot -Force
        }
        continue
    }

    if (-not (Test-Path -LiteralPath $source)) {
        continue
    }

    $destination = Join-Path $artifactRoot $item.Destination
    if (Test-Path -Path $destination -PathType Container) {
        Copy-Item -LiteralPath $source -Destination $destination -Force
    } else {
        New-Item -ItemType Directory -Force -Path (Split-Path $destination) | Out-Null
        Copy-Item -LiteralPath $source -Destination $destination -Force
    }
}

$manifest = Join-Path $artifactRoot 'MANIFEST.txt'
$lines = @(
    "Majucau Financial Intelligence artifacts",
    "Gerado em: $(Get-Date -Format o)",
    "Origem do código: $projectRoot",
    "Arquivos finais: $windowsRoot",
    ""
)
Get-ChildItem -LiteralPath $artifactRoot -Recurse -File |
    Where-Object { $_.FullName -ne $manifest } |
    Sort-Object FullName |
    ForEach-Object { $lines += "$($_.FullName.Substring($artifactRoot.Length + 1))  $($_.Length) bytes" }
Set-Content -LiteralPath $manifest -Value $lines -Encoding utf8

if (-not $KeepSourceOutputs) {
    foreach ($relativePath in @('build\bin', 'frontend\dist', 'tmp')) {
        $sourcePath = Join-Path $projectRoot $relativePath
        if (Test-Path -LiteralPath $sourcePath) {
            Remove-Item -LiteralPath $sourcePath -Recurse -Force
        }
    }
}

Write-Host "Artefatos consolidados em: $artifactRoot"
