Set-StrictMode -Version Latest

Import-Module (Join-Path $PSScriptRoot 'SafeCleanup.psm1') -Force

function Write-MajucauManifestDirectory {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)] [string]$Directory,
        [Parameter(Mandatory)] [string]$ArtifactRoot,
        [Parameter(Mandatory)] [System.IO.StreamWriter]$Writer,
        [int]$Depth = 0
    )

    if ($Depth -gt 128) {
        throw 'A estrutura de pastas dos artefatos excede o limite de profundidade do inventário.'
    }

    foreach ($entry in [System.IO.Directory]::EnumerateFileSystemEntries($Directory)) {
        $attributes = [System.IO.File]::GetAttributes($entry)
        if (($attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
            throw "O inventário encontrou um link inesperado em artifacts: $entry"
        }

        if (($attributes -band [System.IO.FileAttributes]::Directory) -ne 0) {
            Write-MajucauManifestDirectory -Directory $entry -ArtifactRoot $ArtifactRoot -Writer $Writer -Depth ($Depth + 1)
            continue
        }

        $relativePath = $entry.Substring($ArtifactRoot.Length + 1).Replace('\', '/')
        $relativePath = $relativePath.Replace("`r", '\r').Replace("`n", '\n').Replace("`t", '\t')
        $file = [System.IO.FileInfo]::new($entry)
        $Writer.WriteLine("$relativePath  $($file.Length) bytes")
    }
}

function Write-MajucauArtifactManifest {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)] [string]$ProjectRoot,
        [Parameter(Mandatory)] [string]$ManifestPath
    )

    $projectRootFull = [System.IO.Path]::GetFullPath($ProjectRoot).TrimEnd([System.IO.Path]::DirectorySeparatorChar)
    $artifactRoot = [System.IO.Path]::GetFullPath((Join-Path $projectRootFull 'artifacts'))
    $manifestFull = [System.IO.Path]::GetFullPath($ManifestPath)
    $expectedManifest = [System.IO.Path]::GetFullPath((Join-Path $artifactRoot 'MANIFEST.txt'))

    if (-not [string]::Equals($manifestFull, $expectedManifest, [System.StringComparison]::OrdinalIgnoreCase)) {
        throw 'O inventário só pode ser gravado em artifacts/MANIFEST.txt.'
    }

    Assert-MajucauProjectPath -ProjectRoot $projectRootFull -Path $artifactRoot -ExpectedType Directory
    Assert-MajucauProjectPath -ProjectRoot $projectRootFull -Path $manifestFull -ExpectedType File

    $encoding = [System.Text.UTF8Encoding]::new($false)
    $writer = [System.IO.StreamWriter]::new($manifestFull, $false, $encoding)
    try {
        $writer.WriteLine('Majucau Financial Intelligence artifacts')
        $writer.WriteLine("Gerado em: $(Get-Date -Format o)")
        $writer.WriteLine("Origem do código: $projectRootFull")
        $writer.WriteLine("Arquivos finais: $(Join-Path $artifactRoot 'windows')")
        $writer.WriteLine('')

        foreach ($relativeRoot in @('windows', 'verification')) {
            $directory = Join-Path $artifactRoot $relativeRoot
            if (-not (Test-Path -LiteralPath $directory -PathType Container)) {
                continue
            }
            Assert-MajucauProjectPath -ProjectRoot $projectRootFull -Path $directory -ExpectedType Directory
            Write-MajucauManifestDirectory -Directory $directory -ArtifactRoot $artifactRoot -Writer $writer
        }
    }
    finally {
        $writer.Dispose()
    }
}

Export-ModuleMember -Function Write-MajucauArtifactManifest
