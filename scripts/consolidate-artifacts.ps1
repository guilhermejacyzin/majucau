[CmdletBinding()]
param(
    [switch]$KeepSourceOutputs
)

$ErrorActionPreference = 'Stop'
Import-Module (Join-Path $PSScriptRoot 'SafeCleanup.psm1') -Force
Import-Module (Join-Path $PSScriptRoot 'ArtifactManifest.psm1') -Force
$projectRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
$artifactRoot = Join-Path $projectRoot 'artifacts'
$windowsRoot = Join-Path $artifactRoot 'windows'

Assert-MajucauProjectPath -ProjectRoot $projectRoot -Path $artifactRoot -ExpectedType Directory
Assert-MajucauProjectPath -ProjectRoot $projectRoot -Path $windowsRoot -ExpectedType Directory
New-Item -ItemType Directory -Force -Path $windowsRoot | Out-Null
Assert-MajucauProjectPath -ProjectRoot $projectRoot -Path $windowsRoot -ExpectedType Directory

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
        Assert-MajucauProjectPath -ProjectRoot $projectRoot -Path $sourceDirectory -ExpectedType Directory
        $sourceFilter = Split-Path $item.Source -Leaf
        $files = Get-ChildItem -LiteralPath $sourceDirectory -Filter $sourceFilter -File -ErrorAction SilentlyContinue
        foreach ($file in $files) {
            Assert-MajucauProjectPath -ProjectRoot $projectRoot -Path $file.FullName -ExpectedType File
            Copy-Item -LiteralPath $file.FullName -Destination $windowsRoot -Force
        }
        continue
    }

    if (-not (Test-Path -LiteralPath $source)) {
        continue
    }
    Assert-MajucauProjectPath -ProjectRoot $projectRoot -Path $source -ExpectedType File

    $destination = Join-Path $artifactRoot $item.Destination
    if (Test-Path -Path $destination -PathType Container) {
        Copy-Item -LiteralPath $source -Destination $destination -Force
    } else {
        New-Item -ItemType Directory -Force -Path (Split-Path $destination) | Out-Null
        Copy-Item -LiteralPath $source -Destination $destination -Force
    }
}

$manifest = Join-Path $artifactRoot 'MANIFEST.txt'
Write-MajucauArtifactManifest -ProjectRoot $projectRoot -ManifestPath $manifest

if (-not $KeepSourceOutputs) {
    foreach ($relativePath in @('build\bin', 'frontend\dist', 'tmp')) {
        $sourcePath = Join-Path $projectRoot $relativePath
        Remove-MajucauProjectDirectory -ProjectRoot $projectRoot -Path $sourcePath -ExpectedRelativePath $relativePath
    }

    $webview2StagingFile = Join-Path $projectRoot 'build\windows\installer\tmp\MicrosoftEdgeWebview2Setup.exe'
    Remove-MajucauProjectFile -ProjectRoot $projectRoot -Path $webview2StagingFile -ExpectedRelativePath 'build\windows\installer\tmp\MicrosoftEdgeWebview2Setup.exe'
    $webview2StagingRoot = Split-Path -Parent $webview2StagingFile
    Assert-MajucauProjectPath -ProjectRoot $projectRoot -Path $webview2StagingRoot -ExpectedType Directory
    if ((Test-Path -LiteralPath $webview2StagingRoot -PathType Container) -and
        -not (Get-ChildItem -LiteralPath $webview2StagingRoot -Force | Select-Object -First 1)) {
        Remove-MajucauProjectDirectory -ProjectRoot $projectRoot -Path $webview2StagingRoot -ExpectedRelativePath 'build\windows\installer\tmp'
    }
}

Write-Host "Artefatos consolidados em: $artifactRoot"
