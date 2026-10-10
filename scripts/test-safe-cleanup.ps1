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
$relativeTargetRoot = Join-Path $relativeTestRoot 'expected'
$mismatchRejected = $false
$escapeRejected = $false

Assert-MajucauProjectPath -ProjectRoot $projectRoot -Path $testRoot -ExpectedType Directory
try {
    New-Item -ItemType Directory -Force -Path $targetRoot | Out-Null
    New-Item -ItemType Directory -Force -Path $sentinelRoot | Out-Null
    Set-Content -LiteralPath (Join-Path $sentinelRoot 'keep.txt') -Value 'preserve' -Encoding ascii

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

    Write-Output 'Safe cleanup contract: PASS'
} finally {
    Remove-MajucauProjectDirectory -ProjectRoot $projectRoot -Path $testRoot -ExpectedRelativePath $relativeTestRoot
}
