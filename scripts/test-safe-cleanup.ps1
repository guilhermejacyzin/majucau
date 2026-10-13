[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
Import-Module (Join-Path $PSScriptRoot 'SafeCleanup.psm1') -Force
$projectRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
$testDirectoryName = "safe-cleanup-contract-$([guid]::NewGuid().ToString('N'))"
$relativeTestRoot = Join-Path 'artifacts\cache' $testDirectoryName
$testRoot = Join-Path $projectRoot $relativeTestRoot
$targetRoot = Join-Path $testRoot 'expected'
$sentinelRoot = Join-Path $testRoot 'sentinel'
$symbolicLinkTarget = Join-Path $testRoot 'symbolic-link-target.txt'
$symbolicLinkPath = Join-Path $testRoot 'symbolic-link.txt'
$linkedRoot = Join-Path $testRoot 'with-link'
$junctionPath = Join-Path $linkedRoot 'external'
$relativeLinkedRoot = Join-Path $relativeTestRoot 'with-link'
$externalTargetName = "safe-cleanup-target-$([guid]::NewGuid().ToString('N'))"
$relativeExternalTargetRoot = Join-Path 'artifacts\cache' $externalTargetName
$externalTargetRoot = Join-Path $projectRoot $relativeExternalTargetRoot
$relativeTargetRoot = Join-Path $relativeTestRoot 'expected'
$mismatchRejected = $false
$escapeRejected = $false
$reparseRejected = $false

Assert-MajucauProjectPath -ProjectRoot $projectRoot -Path $testRoot -ExpectedType Directory
try {
    New-Item -ItemType Directory -Force -Path $targetRoot | Out-Null
    New-Item -ItemType Directory -Force -Path $sentinelRoot | Out-Null
    Set-Content -LiteralPath (Join-Path $sentinelRoot 'keep.txt') -Value 'preserve' -Encoding ascii
    Set-Content -LiteralPath $symbolicLinkTarget -Value 'preserve' -Encoding ascii

    try {
        Assert-MajucauProjectPath -ProjectRoot $projectRoot -Path (Join-Path $projectRoot '..\outside-project')
    } catch {
        $escapeRejected = $true
    }
    if (-not $escapeRejected) {
        throw 'SafeCleanup aceitou um caminho fora da raiz do projeto.'
    }

    try {
        Remove-MajucauProjectDirectory -ProjectRoot $projectRoot -Path $sentinelRoot -ExpectedRelativePath $relativeTargetRoot
    } catch {
        $mismatchRejected = $true
    }
    if (-not $mismatchRejected -or -not (Test-Path -LiteralPath (Join-Path $sentinelRoot 'keep.txt') -PathType Leaf)) {
        throw 'SafeCleanup não rejeitou o destino diferente ou removeu o arquivo sentinela.'
    }

    Remove-MajucauProjectDirectory -ProjectRoot $projectRoot -Path $targetRoot -ExpectedRelativePath $relativeTargetRoot
    if (Test-Path -LiteralPath $targetRoot) {
        throw 'SafeCleanup não removeu a pasta de teste autorizada.'
    }

    New-Item -ItemType Directory -Force -Path $linkedRoot, $externalTargetRoot | Out-Null
    Set-Content -LiteralPath (Join-Path $externalTargetRoot 'keep.txt') -Value 'preserve' -Encoding ascii
    New-Item -ItemType Junction -Path $junctionPath -Target $externalTargetRoot | Out-Null
    try {
        Remove-MajucauProjectDirectory -ProjectRoot $projectRoot -Path $linkedRoot -ExpectedRelativePath $relativeLinkedRoot
    } catch {
        $reparseRejected = $_.Exception.Message -like '*link para outro local*'
        if (-not $reparseRejected) { throw }
    }
    if (-not $reparseRejected -or -not (Test-Path -LiteralPath (Join-Path $externalTargetRoot 'keep.txt') -PathType Leaf)) {
        throw 'SafeCleanup não rejeitou o reparse point ou removeu o arquivo externo sentinela.'
    }
    [System.IO.Directory]::Delete($junctionPath, $false)
    Remove-MajucauProjectDirectory -ProjectRoot $projectRoot -Path $linkedRoot -ExpectedRelativePath $relativeLinkedRoot

    New-Item -ItemType SymbolicLink -Path $symbolicLinkPath -Target $symbolicLinkTarget | Out-Null
    $wrongTargetRejected = $false
    try {
        Remove-MajucauProjectSymbolicLink -ProjectRoot $projectRoot -Path $symbolicLinkPath -ExpectedRelativePath (Join-Path $relativeTestRoot 'symbolic-link.txt') -ExpectedTarget (Join-Path $testRoot 'sentinel/keep.txt')
    } catch {
        $wrongTargetRejected = $_.Exception.Message -like '*outro destino*'
        if (-not $wrongTargetRejected) { throw }
    }
    if (-not $wrongTargetRejected -or -not (Test-Path -LiteralPath $symbolicLinkPath) -or -not (Test-Path -LiteralPath $symbolicLinkTarget)) {
        throw 'SafeCleanup não rejeitou o destino incorreto ou removeu o arquivo de destino.'
    }
    Remove-MajucauProjectSymbolicLink -ProjectRoot $projectRoot -Path $symbolicLinkPath -ExpectedRelativePath (Join-Path $relativeTestRoot 'symbolic-link.txt') -ExpectedTarget $symbolicLinkTarget
    if ((Test-Path -LiteralPath $symbolicLinkPath) -or -not (Test-Path -LiteralPath $symbolicLinkTarget)) {
        throw 'SafeCleanup não removeu somente o link simbólico autorizado.'
    }

    Write-Output 'Safe cleanup contract: PASS'
} finally {
    if (Test-Path -LiteralPath $symbolicLinkPath) {
        Remove-MajucauProjectSymbolicLink -ProjectRoot $projectRoot -Path $symbolicLinkPath -ExpectedRelativePath (Join-Path $relativeTestRoot 'symbolic-link.txt') -ExpectedTarget $symbolicLinkTarget
    }
    if (Test-Path -LiteralPath $junctionPath) {
        [System.IO.Directory]::Delete($junctionPath, $false)
    }
    Remove-MajucauProjectDirectory -ProjectRoot $projectRoot -Path $testRoot -ExpectedRelativePath $relativeTestRoot
    Remove-MajucauProjectDirectory -ProjectRoot $projectRoot -Path $externalTargetRoot -ExpectedRelativePath $relativeExternalTargetRoot
}
