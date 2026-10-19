Set-StrictMode -Version Latest

function Remove-MajucauWorkerServiceSmokeData {
    [CmdletBinding()]
    param([Parameter(Mandatory)] [string]$DataRoot)

    $expectedRoot = [System.IO.Path]::GetFullPath((Join-Path $env:ProgramData 'Majucau'))
    $fullRoot = [System.IO.Path]::GetFullPath($DataRoot)
    if (-not [string]::Equals($fullRoot, $expectedRoot, [System.StringComparison]::OrdinalIgnoreCase)) {
        throw 'A limpeza do smoke recusou uma pasta de dados diferente de ProgramData\Majucau.'
    }
    if (-not (Test-Path -LiteralPath $fullRoot)) { return }

    $allowedDirectories = @{
        'WorkerLogs' = $true
        'WorkerLogs\Logs' = $true
        'WorkerSecrets' = $true
        'WorkerSecrets\Vault' = $true
    }
    $allowedFiles = @{
        'WorkerLogs\Logs\majucau-worker.jsonl' = $true
        'WorkerLogs\Logs\majucau-worker.jsonl.1' = $true
        'WorkerLogs\Logs\majucau-worker.jsonl.2' = $true
        'WorkerLogs\Logs\majucau-worker.jsonl.3' = $true
        'WorkerLogs\Logs\majucau-worker.jsonl.4' = $true
        'WorkerSecrets\Vault\.entropy' = $true
        'WorkerSecrets\Vault\.scope' = $true
    }

    $rootItem = Get-Item -LiteralPath $fullRoot -Force -ErrorAction Stop
    if (-not $rootItem.PSIsContainer -or ($rootItem.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
        throw 'A pasta de dados do smoke não é uma pasta local segura.'
    }
    $items = @(Get-ChildItem -LiteralPath $fullRoot -Force -Recurse -ErrorAction Stop)
    foreach ($item in $items) {
        if (($item.Attributes -band [System.IO.FileAttributes]::ReparsePoint) -ne 0) {
            throw 'A pasta de dados do smoke contém um link e foi preservada.'
        }
        $relative = [System.IO.Path]::GetRelativePath($fullRoot, $item.FullName).Replace('/', '\')
        if ($item.PSIsContainer) {
            if (-not $allowedDirectories.ContainsKey($relative)) { throw "A pasta de dados contém uma pasta inesperada e foi preservada: $relative" }
        } elseif (-not $allowedFiles.ContainsKey($relative)) {
            throw "A pasta de dados contém um arquivo inesperado e foi preservada: $relative"
        }
    }

    foreach ($item in $items | Where-Object { -not $_.PSIsContainer } | Sort-Object { $_.FullName.Length } -Descending) {
        Remove-Item -LiteralPath $item.FullName -Force -ErrorAction Stop
    }
    foreach ($item in $items | Where-Object { $_.PSIsContainer } | Sort-Object { $_.FullName.Length } -Descending) {
        Remove-Item -LiteralPath $item.FullName -Force -ErrorAction Stop
    }
    Remove-Item -LiteralPath $fullRoot -Force -ErrorAction Stop
}

function Register-MajucauWorkerServiceSmokeRetention {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory)] [string]$OutputRoot,
        [Parameter(Mandatory)] [string[]]$Paths,
        [Parameter(Mandatory)] [string]$Reason
    )

    $root = [System.IO.Path]::GetFullPath($OutputRoot)
    New-Item -ItemType Directory -Force -Path $root | Out-Null
    [ordered]@{
        schema_version = '1.0'
        status = 'MANUAL_CLEANUP_REQUIRED'
        retained_paths = $Paths
        reason = $Reason
        created_at_utc = [DateTime]::UtcNow.ToString('o')
    } | ConvertTo-Json -Depth 4 | Set-Content -LiteralPath (Join-Path $root 'RETAINED-CLEANUP.json') -Encoding utf8
}

Export-ModuleMember -Function Remove-MajucauWorkerServiceSmokeData, Register-MajucauWorkerServiceSmokeRetention
