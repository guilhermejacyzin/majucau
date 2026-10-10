Set-StrictMode -Version Latest

function ConvertTo-MajucauFullPath {
    param([Parameter(Mandatory)] [string]$Path)

    $fullPath = [System.IO.Path]::GetFullPath($Path)
    $pathRoot = [System.IO.Path]::GetPathRoot($fullPath)
    if ($fullPath.Length -gt $pathRoot.Length) {
        $fullPath = $fullPath.TrimEnd([char[]]@(
            [System.IO.Path]::DirectorySeparatorChar,
            [System.IO.Path]::AltDirectorySeparatorChar
        ))
    }
    return $fullPath
}

function Get-MajucauPathItem {
    param([Parameter(Mandatory)] [string]$Path)

    try {
        [System.IO.File]::GetAttributes($Path) | Out-Null
    } catch [System.IO.FileNotFoundException] {
        return $null
    } catch [System.IO.DirectoryNotFoundException] {
        return $null
    }
    return Get-Item -LiteralPath $Path -Force -ErrorAction Stop
}

function Assert-MajucauProjectPath {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)] [string]$ProjectRoot,
        [Parameter(Mandatory)] [string]$Path,
        [ValidateSet('Any', 'Directory', 'File')] [string]$ExpectedType = 'Any'
    )

    $rootFull = ConvertTo-MajucauFullPath -Path $ProjectRoot
    $pathFull = ConvertTo-MajucauFullPath -Path $Path
    $isRoot = [string]::Equals($rootFull, $pathFull, [System.StringComparison]::OrdinalIgnoreCase)
    $rootPrefix = $rootFull
    if (-not $rootPrefix.EndsWith([string][System.IO.Path]::DirectorySeparatorChar) -and
        -not $rootPrefix.EndsWith([string][System.IO.Path]::AltDirectorySeparatorChar)) {
        $rootPrefix += [System.IO.Path]::DirectorySeparatorChar
    }
    if (-not $isRoot -and -not $pathFull.StartsWith($rootPrefix, [System.StringComparison]::OrdinalIgnoreCase)) {
        throw "O caminho saiu da raiz permitida do projeto: $Path"
    }
    if ($isRoot -and $ExpectedType -eq 'File') {
        throw 'A raiz do projeto não pode ser tratada como arquivo.'
    }

    $pathsToCheck = New-Object 'System.Collections.Generic.List[string]'
    $pathsToCheck.Add($rootFull)
    if (-not $isRoot) {
        $relativePath = $pathFull.Substring($rootPrefix.Length)
        $components = @($relativePath -split '[\\/]+' | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
        $currentPath = $rootFull
        foreach ($component in $components) {
            $currentPath = Join-Path $currentPath $component
            $pathsToCheck.Add($currentPath)
        }
    }

    for ($index = 0; $index -lt $pathsToCheck.Count; $index++) {
        $currentPath = $pathsToCheck[$index]
        $item = Get-MajucauPathItem -Path $currentPath
        if ($null -eq $item) { continue }
        if (($item.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
            throw "O caminho contém um link para outro local (reparse point): $currentPath"
        }

        $isLeaf = $index -eq ($pathsToCheck.Count - 1)
        if (-not $isLeaf -and -not $item.PSIsContainer) {
            throw "Um componente do caminho não é uma pasta: $currentPath"
        }
        if ($isLeaf -and $ExpectedType -eq 'Directory' -and -not $item.PSIsContainer) {
            throw "O caminho esperado como pasta não é uma pasta: $currentPath"
        }
        if ($isLeaf -and $ExpectedType -eq 'File' -and $item.PSIsContainer) {
            throw "O caminho esperado como arquivo é uma pasta: $currentPath"
        }
    }
}

function Assert-MajucauNoReparseDescendants {
    [CmdletBinding()]
    param([Parameter(Mandatory)] [string]$Path)

    $item = Get-MajucauPathItem -Path $Path
    if ($null -eq $item) { return }
    if (-not $item.PSIsContainer) {
        throw "O destino de limpeza não é uma pasta: $Path"
    }

    $linkedItem = Get-ChildItem -LiteralPath $Path -Force -Recurse -ErrorAction Stop |
        Where-Object { ($_.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0 } |
        Select-Object -First 1
    if ($null -ne $linkedItem) {
        throw "O destino contém um link para outro local e não será apagado: $($linkedItem.FullName)"
    }
}

function Remove-MajucauProjectDirectory {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)] [string]$ProjectRoot,
        [Parameter(Mandatory)] [string]$Path,
        [Parameter(Mandatory)] [string]$ExpectedRelativePath
    )

    $rootFull = ConvertTo-MajucauFullPath -Path $ProjectRoot
    $pathFull = ConvertTo-MajucauFullPath -Path $Path
    $expectedFull = ConvertTo-MajucauFullPath -Path (Join-Path $rootFull $ExpectedRelativePath)
    if ([string]::Equals($rootFull, $pathFull, [System.StringComparison]::OrdinalIgnoreCase) -or
        -not [string]::Equals($pathFull, $expectedFull, [System.StringComparison]::OrdinalIgnoreCase)) {
        throw "A limpeza recusou um destino diferente da pasta prevista: $Path"
    }

    Assert-MajucauProjectPath -ProjectRoot $rootFull -Path $pathFull -ExpectedType Directory
    $item = Get-MajucauPathItem -Path $pathFull
    if ($null -eq $item) { return }
    Assert-MajucauNoReparseDescendants -Path $pathFull
    Remove-Item -LiteralPath $pathFull -Recurse -Force -ErrorAction Stop
}

function Remove-MajucauProjectFile {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)] [string]$ProjectRoot,
        [Parameter(Mandatory)] [string]$Path,
        [Parameter(Mandatory)] [string]$ExpectedRelativePath
    )

    $rootFull = ConvertTo-MajucauFullPath -Path $ProjectRoot
    $pathFull = ConvertTo-MajucauFullPath -Path $Path
    $expectedFull = ConvertTo-MajucauFullPath -Path (Join-Path $rootFull $ExpectedRelativePath)
    if (-not [string]::Equals($pathFull, $expectedFull, [System.StringComparison]::OrdinalIgnoreCase)) {
        throw "A limpeza recusou um arquivo diferente do previsto: $Path"
    }

    Assert-MajucauProjectPath -ProjectRoot $rootFull -Path $pathFull -ExpectedType File
    $item = Get-MajucauPathItem -Path $pathFull
    if ($null -eq $item) { return }
    Remove-Item -LiteralPath $pathFull -Force -ErrorAction Stop
}

Export-ModuleMember -Function Assert-MajucauProjectPath, Remove-MajucauProjectDirectory, Remove-MajucauProjectFile
