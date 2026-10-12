Set-StrictMode -Version Latest

Import-Module (Join-Path $PSScriptRoot 'SafeCleanup.psm1') -Force

function Get-MajucauRelativePath {
    param([Parameter(Mandatory)] [string]$ProjectRoot, [Parameter(Mandatory)] [string]$Path)

    $rootFull = [System.IO.Path]::GetFullPath($ProjectRoot).TrimEnd([char[]]@('\', '/'))
    $pathFull = [System.IO.Path]::GetFullPath($Path)
    $prefix = $rootFull + [System.IO.Path]::DirectorySeparatorChar
    if (-not $pathFull.StartsWith($prefix, [System.StringComparison]::OrdinalIgnoreCase)) {
        throw 'O caminho precisa permanecer dentro do projeto.'
    }
    return $pathFull.Substring($prefix.Length)
}

function Preserve-MajucauFailedBuildOutputs {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)] [string]$ProjectRoot,
        [Parameter(Mandatory)]
        [ValidatePattern('^[A-Za-z0-9-]{8,64}$')]
        [string]$RunId,
        [switch]$IncludePublishedArtifacts
    )

    $projectRootFull = [System.IO.Path]::GetFullPath($ProjectRoot)
    $artifactRoot = Join-Path $projectRootFull 'artifacts'
    $retainedRelative = Join-Path 'artifacts\retained\manual-cleanup' "failed-build-$RunId"
    $retainedRoot = Join-Path $projectRootFull $retainedRelative
    $registryPath = Join-Path $artifactRoot 'retained\cleanup-register.md'
    $sourceRoots = @(
        'build\bin',
        'frontend\dist',
        'tmp',
        'build\windows\installer\tmp'
    )
    if ($IncludePublishedArtifacts) { $sourceRoots += 'artifacts\windows' }
    $sourceFiles = if ($IncludePublishedArtifacts) { @('artifacts\MANIFEST.txt') } else { @() }
    $moved = [System.Collections.Generic.List[string]]::new()
    $notMoved = [System.Collections.Generic.List[string]]::new()

    Assert-MajucauProjectPath -ProjectRoot $projectRootFull -Path $artifactRoot -ExpectedType Directory
    Assert-MajucauProjectPath -ProjectRoot $projectRootFull -Path $registryPath -ExpectedType File
    $registryParent = Split-Path -Parent $registryPath
    Assert-MajucauProjectPath -ProjectRoot $projectRootFull -Path $registryParent -ExpectedType Directory
    New-Item -ItemType Directory -Force -Path $registryParent | Out-Null

    foreach ($relativeSource in $sourceRoots) {
        $source = Join-Path $projectRootFull $relativeSource
        if (-not (Test-Path -LiteralPath $source)) { continue }

        try {
            Assert-MajucauProjectPath -ProjectRoot $projectRootFull -Path $source -ExpectedType Directory
            $destination = Join-Path $retainedRoot $relativeSource
            Assert-MajucauProjectPath -ProjectRoot $projectRootFull -Path $destination -ExpectedType Directory
            $destinationParent = Split-Path -Parent $destination
            Assert-MajucauProjectPath -ProjectRoot $projectRootFull -Path $destinationParent -ExpectedType Directory
            New-Item -ItemType Directory -Force -Path $destinationParent | Out-Null
            if (Test-Path -LiteralPath $destination) {
                throw 'O destino de retenção já existe.'
            }
            Move-Item -LiteralPath $source -Destination $destination -ErrorAction Stop
            if (Test-Path -LiteralPath $source) {
                throw 'A pasta de origem permaneceu após a tentativa de transferência.'
            }
            $moved.Add(($destination.Substring($projectRootFull.TrimEnd([char[]]@('\', '/')).Length + 1)).Replace('\', '/'))
        }
        catch {
            $notMoved.Add($relativeSource.Replace('\', '/') + ' (revisão manual necessária)')
        }
    }

    foreach ($relativeSource in $sourceFiles) {
        $source = Join-Path $projectRootFull $relativeSource
        if (-not (Test-Path -LiteralPath $source)) { continue }

        try {
            Assert-MajucauProjectPath -ProjectRoot $projectRootFull -Path $source -ExpectedType File
            $destination = Join-Path $retainedRoot $relativeSource
            Assert-MajucauProjectPath -ProjectRoot $projectRootFull -Path $destination -ExpectedType File
            $destinationParent = Split-Path -Parent $destination
            Assert-MajucauProjectPath -ProjectRoot $projectRootFull -Path $destinationParent -ExpectedType Directory
            New-Item -ItemType Directory -Force -Path $destinationParent | Out-Null
            if (Test-Path -LiteralPath $destination) {
                throw 'O destino de retenção já existe.'
            }
            Move-Item -LiteralPath $source -Destination $destination -ErrorAction Stop
            if (Test-Path -LiteralPath $source) {
                throw 'O arquivo de origem permaneceu após a tentativa de transferência.'
            }
            $moved.Add((Get-MajucauRelativePath -ProjectRoot $projectRootFull -Path $destination).Replace('\', '/'))
        }
        catch {
            $notMoved.Add($relativeSource.Replace('\', '/') + ' (revisão manual necessária)')
        }
    }

    if ($moved.Count -eq 0 -and $notMoved.Count -eq 0) {
        return [pscustomobject]@{ RunId = $RunId; Moved = @(); NotMoved = @() }
    }

    $registryLines = [System.Collections.Generic.List[string]]::new()
    $registryLines.Add('')
    $registryLines.Add("## Saídas encontradas após build interrompido — $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss zzz')")
    $registryLines.Add("- Identificador: $RunId.")
    if ($moved.Count -gt 0) {
        $registryLines.Add('- Guardadas em `artifacts/retained/manual-cleanup/failed-build-' + $RunId + '/`:')
        foreach ($relativePath in $moved) { $registryLines.Add("  - ``$relativePath``") }
    }
    if ($notMoved.Count -gt 0) {
        $registryLines.Add('- Não transferidas automaticamente; permanecem nos caminhos abaixo para revisão manual:')
        foreach ($relativePath in $notMoved) { $registryLines.Add("  - ``$relativePath``") }
    }
    $registryLines.Add('- Motivo: a execução do build foi interrompida. As pastas podem incluir arquivos de uma execução anterior; revise antes da limpeza manual.')

    $utf8NoBom = [System.Text.UTF8Encoding]::new($false)
    if (-not (Test-Path -LiteralPath $registryPath -PathType Leaf)) {
        [System.IO.File]::WriteAllText($registryPath, "# Registro local de limpeza manual`r`n", $utf8NoBom)
    }
    [System.IO.File]::AppendAllText(
        $registryPath,
        ([string]::Join([Environment]::NewLine, $registryLines) + [Environment]::NewLine),
        $utf8NoBom
    )

    return [pscustomobject]@{
        RunId = $RunId
        Moved = $moved.ToArray()
        NotMoved = $notMoved.ToArray()
    }
}

function New-MajucauArtifactBaseline {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)] [string]$ProjectRoot,
        [Parameter(Mandatory)]
        [ValidatePattern('^[A-Za-z0-9-]{8,64}$')]
        [string]$RunId
    )

    $projectRootFull = [System.IO.Path]::GetFullPath($ProjectRoot)
    $artifactRoot = Join-Path $projectRootFull 'artifacts'
    $windowsRoot = Join-Path $artifactRoot 'windows'
    $manifestPath = Join-Path $artifactRoot 'MANIFEST.txt'
    $baselineRelative = Join-Path 'artifacts\cache' "build-baseline-$RunId"
    $baselineRoot = Join-Path $projectRootFull $baselineRelative
    $hasWindows = Test-Path -LiteralPath $windowsRoot -PathType Container
    $hasManifest = Test-Path -LiteralPath $manifestPath -PathType Leaf

    if (-not $hasWindows -and -not $hasManifest) {
        return [pscustomobject]@{ Root = $null; Windows = $false; Manifest = $false }
    }

    Assert-MajucauProjectPath -ProjectRoot $projectRootFull -Path $artifactRoot -ExpectedType Directory
    Assert-MajucauProjectPath -ProjectRoot $projectRootFull -Path $baselineRoot -ExpectedType Directory
    if (Test-Path -LiteralPath $baselineRoot) { throw 'Já existe uma cópia-base para esta execução.' }
    if ($hasWindows) {
        Assert-MajucauProjectPath -ProjectRoot $projectRootFull -Path $windowsRoot -ExpectedType Directory
        Assert-MajucauNoReparseDescendants -Path $windowsRoot
    }
    if ($hasManifest) {
        Assert-MajucauProjectPath -ProjectRoot $projectRootFull -Path $manifestPath -ExpectedType File
    }

    New-Item -ItemType Directory -Force -Path $baselineRoot | Out-Null
    try {
        if ($hasWindows) {
            $windowsBaseline = Join-Path $baselineRoot 'windows'
            New-Item -ItemType Directory -Force -Path $windowsBaseline | Out-Null
            foreach ($entry in [System.IO.Directory]::EnumerateFileSystemEntries($windowsRoot)) {
                Copy-Item -LiteralPath $entry -Destination $windowsBaseline -Recurse -Force
            }
        }
        if ($hasManifest) {
            Copy-Item -LiteralPath $manifestPath -Destination (Join-Path $baselineRoot 'MANIFEST.txt') -Force
        }
        return [pscustomobject]@{
            Root = $baselineRoot
            Windows = $hasWindows
            Manifest = $hasManifest
        }
    }
    catch {
        try {
            Remove-MajucauProjectDirectory -ProjectRoot $projectRootFull -Path $baselineRoot -ExpectedRelativePath $baselineRelative
        }
        catch {
            Add-MajucauCleanupRegistryNote -ProjectRoot $projectRootFull -Lines @(
                '',
                '## Cópia-base incompleta de build',
                "- Verifique e remova manualmente, após conferir, $baselineRelative."
            )
            throw 'Não foi possível copiar a versão anterior dos artefatos nem limpar a cópia incompleta; o caminho foi anotado no registro local.'
        }
        throw
    }
}

function Add-MajucauCleanupRegistryNote {
    param([string]$ProjectRoot, [string[]]$Lines)

    $projectRootFull = [System.IO.Path]::GetFullPath($ProjectRoot)
    $registryPath = Join-Path $projectRootFull 'artifacts\retained\cleanup-register.md'
    $registryParent = Split-Path -Parent $registryPath
    Assert-MajucauProjectPath -ProjectRoot $projectRootFull -Path $registryParent -ExpectedType Directory
    New-Item -ItemType Directory -Force -Path $registryParent | Out-Null
    Assert-MajucauProjectPath -ProjectRoot $projectRootFull -Path $registryPath -ExpectedType File
    $utf8NoBom = [System.Text.UTF8Encoding]::new($false)
    if (-not (Test-Path -LiteralPath $registryPath -PathType Leaf)) {
        [System.IO.File]::WriteAllText($registryPath, "# Registro local de limpeza manual`r`n", $utf8NoBom)
    }
    [System.IO.File]::AppendAllText(
        $registryPath,
        ([string]::Join([Environment]::NewLine, $Lines) + [Environment]::NewLine),
        $utf8NoBom
    )
}

function Complete-MajucauArtifactBaseline {
    [CmdletBinding()]
    param([Parameter(Mandatory)] [string]$ProjectRoot, [Parameter(Mandatory)]$Baseline)

    if ($null -eq $Baseline.Root) { return }
    $relative = Get-MajucauRelativePath -ProjectRoot $ProjectRoot -Path $Baseline.Root
    try {
        Remove-MajucauProjectDirectory -ProjectRoot $ProjectRoot -Path $Baseline.Root -ExpectedRelativePath $relative
    }
    catch {
        Add-MajucauCleanupRegistryNote -ProjectRoot $ProjectRoot -Lines @(
            '',
            '## Cópia-base de build mantida após sucesso',
            "- Verifique e remova manualmente, após conferir, $relative."
        )
        throw 'Não foi possível remover a cópia-base; o caminho foi anotado no registro local.'
    }
}

function Restore-MajucauArtifactBaseline {
    [CmdletBinding()]
    param([Parameter(Mandatory)] [string]$ProjectRoot, [Parameter(Mandatory)]$Baseline)

    if ($null -eq $Baseline.Root) { return [pscustomobject]@{ Restored = $true; Pending = @() } }
    $projectRootFull = [System.IO.Path]::GetFullPath($ProjectRoot)
    $artifactRoot = Join-Path $projectRootFull 'artifacts'
    $windowsRoot = Join-Path $artifactRoot 'windows'
    $manifestPath = Join-Path $artifactRoot 'MANIFEST.txt'
    $pending = [System.Collections.Generic.List[string]]::new()

    foreach ($entry in @(
        @{ Name = 'windows'; Source = (Join-Path $Baseline.Root 'windows'); Destination = $windowsRoot; WasPresent = [bool]$Baseline.Windows; IsDirectory = $true },
        @{ Name = 'MANIFEST.txt'; Source = (Join-Path $Baseline.Root 'MANIFEST.txt'); Destination = $manifestPath; WasPresent = [bool]$Baseline.Manifest; IsDirectory = $false }
    )) {
        $destinationExists = Test-Path -LiteralPath $entry.Destination
        if ($destinationExists) {
            $pending.Add("$($entry.Destination.Substring($projectRootFull.Length + 1).Replace('\', '/')) ainda existe e não será sobrescrito")
            continue
        }
        if (-not $entry.WasPresent) { continue }
        if (-not (Test-Path -LiteralPath $entry.Source)) {
            $pending.Add("a cópia anterior de $($entry.Name) está ausente em $($Baseline.Root)")
            continue
        }
        Assert-MajucauProjectPath -ProjectRoot $projectRootFull -Path $entry.Source -ExpectedType $(if ($entry.IsDirectory) { 'Directory' } else { 'File' })
        Assert-MajucauProjectPath -ProjectRoot $projectRootFull -Path $entry.Destination -ExpectedType $(if ($entry.IsDirectory) { 'Directory' } else { 'File' })
        try {
            Move-Item -LiteralPath $entry.Source -Destination $entry.Destination -ErrorAction Stop
        }
        catch {
            $pending.Add("não foi possível restaurar $($entry.Name); a cópia permanece em $($Baseline.Root)")
        }
    }

    if ($pending.Count -eq 0) {
        $relative = Get-MajucauRelativePath -ProjectRoot $projectRootFull -Path $Baseline.Root
        try {
            Remove-MajucauProjectDirectory -ProjectRoot $projectRootFull -Path $Baseline.Root -ExpectedRelativePath $relative
            return [pscustomobject]@{ Restored = $true; Pending = @() }
        }
        catch {
            $pending.Add("a cópia-base restaurada ainda precisa ser limpa em $relative")
        }
    }

    $registryLines = @('', '## Restauração de artefatos pendente', "- A cópia anterior está preservada em ``$($Baseline.Root.Substring($projectRootFull.Length + 1).Replace('\', '/'))``.")
    foreach ($item in $pending) { $registryLines += "- $item." }
    Add-MajucauCleanupRegistryNote -ProjectRoot $projectRootFull -Lines $registryLines
    return [pscustomobject]@{ Restored = $false; Pending = $pending.ToArray() }
}

Export-ModuleMember -Function Preserve-MajucauFailedBuildOutputs, New-MajucauArtifactBaseline, Complete-MajucauArtifactBaseline, Restore-MajucauArtifactBaseline
