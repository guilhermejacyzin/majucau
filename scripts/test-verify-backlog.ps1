[CmdletBinding()]
param()

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$repoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$artifactRoot = [System.IO.Path]::GetFullPath((Join-Path $repoRoot 'artifacts'))
$tempRoot = [System.IO.Path]::GetFullPath((Join-Path $artifactRoot 'tmp'))
$tempPrefix = $tempRoot.TrimEnd([System.IO.Path]::DirectorySeparatorChar) + [System.IO.Path]::DirectorySeparatorChar
$cleanupModule = Join-Path $PSScriptRoot 'SafeCleanup.psm1'
Import-Module $cleanupModule -Force

if (-not $tempRoot.StartsWith(
    $artifactRoot.TrimEnd([System.IO.Path]::DirectorySeparatorChar) + [System.IO.Path]::DirectorySeparatorChar,
    [System.StringComparison]::OrdinalIgnoreCase
)) {
    throw 'A pasta temporária do teste precisa permanecer dentro de artifacts/tmp.'
}
Assert-MajucauProjectPath -ProjectRoot $repoRoot -Path $tempRoot -ExpectedType Directory
[void][System.IO.Directory]::CreateDirectory($tempRoot)
Assert-MajucauProjectPath -ProjectRoot $repoRoot -Path $tempRoot -ExpectedType Directory

$fixturePath = [System.IO.Path]::GetFullPath((Join-Path $tempRoot ('backlog-check-' + [guid]::NewGuid().ToString('N') + '.md')))
$fixtureRelativePath = Join-Path 'artifacts\tmp' ([System.IO.Path]::GetFileName($fixturePath))
if (-not $fixturePath.StartsWith($tempPrefix, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw 'O arquivo temporário do teste saiu de artifacts/tmp.'
}
Assert-MajucauProjectPath -ProjectRoot $repoRoot -Path $fixturePath -ExpectedType File

$validatorPath = Join-Path $PSScriptRoot 'verify-backlog.ps1'
$utf8 = [System.Text.UTF8Encoding]::new($false)
$validFixture = @(
    '## Recontagem atual'
    '- O backlog tem 1 tarefas: 1 concluídas, 0 em andamento e 0 não iniciadas.'
    '- Pela média simples, `(1 × 100%) ÷ 1 = 100%`; restam 0%.'
    '## 00 Tarefas detalhadas'
    '### BK-000 — Exemplo'
    '**Situação:** Concluída. **Evolução:** 100%.'
) -join "`n"

try {
    [System.IO.File]::WriteAllText($fixturePath, $validFixture, $utf8)
    $validOutput = @(& $validatorPath -Path $fixturePath) -join "`n"
    if ($validOutput -notmatch '1 tarefas; 1 concluídas, 0 em andamento e 0 não iniciadas; 100% concluído, 0% restante') {
        throw "O verificador não reconheceu o fixture válido: $validOutput"
    }

    $invalidFixture = $validFixture.Replace('= 100%`; restam 0%.', '= 90%`; restam 10%.')
    if ($invalidFixture -eq $validFixture) {
        throw 'O fixture de teste não encontrou a frase do percentual esperado.'
    }
    [System.IO.File]::WriteAllText($fixturePath, $invalidFixture, $utf8)

    $rejected = $false
    try {
        [void](& $validatorPath -Path $fixturePath)
    }
    catch {
        if ($_.Exception.Message -like '*Percentual divergente*') {
            $rejected = $true
        }
        else {
            throw
        }
    }
    if (-not $rejected) {
        throw 'O verificador aceitou um resumo com percentual incorreto.'
    }

    Write-Output 'Teste do verificador aprovado: aceita contagens corretas e rejeita percentual divergente.'
}
finally {
    Remove-MajucauProjectFile -ProjectRoot $repoRoot -Path $fixturePath -ExpectedRelativePath $fixtureRelativePath
}
