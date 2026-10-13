[CmdletBinding()]
param(
    [string]$Path,
    [string]$StructuredPath
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

if ([string]::IsNullOrWhiteSpace($Path)) {
    $scriptDirectory = Split-Path -Parent $MyInvocation.MyCommand.Path
    $Path = Join-Path $scriptDirectory '..\artifacts\planejamento\outputs\MAJUCAU_BACKLOG_DETALHADO_v2.md'
    if ([string]::IsNullOrWhiteSpace($StructuredPath)) {
        $StructuredPath = Join-Path $scriptDirectory '..\docs\planejamento\work\backlog_v2.json'
    }
}

$maxFileBytes = 8MB
$maxLineChars = 256KB
$fullPath = [System.IO.Path]::GetFullPath($Path)
if (-not [System.IO.File]::Exists($fullPath)) {
    throw "Arquivo de backlog não encontrado: $fullPath"
}

$file = Get-Item -LiteralPath $fullPath
if ($file.Length -gt $maxFileBytes) {
    throw "Arquivo de backlog excede o limite de 8 MiB: $($file.Length) bytes."
}

$tasks = [System.Collections.Generic.List[object]]::new()
$taskIds = [System.Collections.Generic.HashSet[string]]::new([System.StringComparer]::Ordinal)
$currentTask = $null
$readingSummary = $true
$declaredCounts = $null
$declaredProgress = $null
$lineNumber = 0
$reader = [System.IO.StreamReader]::new(
    $fullPath,
    [System.Text.UTF8Encoding]::new($false, $true),
    $true
)

try {
    while (($line = $reader.ReadLine()) -ne $null) {
        $lineNumber++
        if ($line.Length -gt $maxLineChars) {
            throw "Linha $lineNumber excede o limite de 256 Ki caracteres."
        }

        if ($readingSummary -and $line -match '^## 00\s+') {
            $readingSummary = $false
        }

        if ($readingSummary) {
            $countMatch = [regex]::Match(
                $line,
                '^\s*-\s*O backlog (?:mantém|tem|continua com) (?<total>\d+) tarefas: (?<done>\d+) concluídas, (?<progress>\d+) em andamento e (?<pending>\d+) não iniciadas\.'
            )
            if ($countMatch.Success) {
                $declaredCounts = [pscustomobject]@{
                    Total = [int]$countMatch.Groups['total'].Value
                    Done = [int]$countMatch.Groups['done'].Value
                    InProgress = [int]$countMatch.Groups['progress'].Value
                    NotStarted = [int]$countMatch.Groups['pending'].Value
                }
            }

            $progressMatch = [regex]::Match(
                $line,
                'Pela média simples.*?=\s*(?<complete>\d+(?:,\d+)?)%(?:`)?;\s*restam\s*(?<remaining>\d+(?:,\d+)?)%'
            )
            if ($progressMatch.Success) {
                $declaredProgress = [pscustomobject]@{
                    Complete = [double]::Parse(
                        $progressMatch.Groups['complete'].Value.Replace(',', '.'),
                        [System.Globalization.CultureInfo]::InvariantCulture
                    )
                    Remaining = [double]::Parse(
                        $progressMatch.Groups['remaining'].Value.Replace(',', '.'),
                        [System.Globalization.CultureInfo]::InvariantCulture
                    )
                }
            }
        }

        $taskHeader = [regex]::Match($line, '^### BK-(?<id>\d{3})\s+—')
        if ($taskHeader.Success) {
            if ($null -ne $currentTask) {
                throw "A tarefa BK-$($currentTask.Id) não tem linha de situação/percentual (antes da linha $lineNumber)."
            }

            $taskId = $taskHeader.Groups['id'].Value
            if (-not $taskIds.Add($taskId)) {
                throw "ID de tarefa repetido no backlog: BK-$taskId."
            }
            $currentTask = [pscustomobject]@{ Id = $taskId; HeaderLine = $lineNumber }
            continue
        }

        if ($null -ne $currentTask) {
            $statusMatch = [regex]::Match(
                $line,
                '\*\*Situação:\*\*\s*(?<status>Concluída|Em andamento|Não iniciada)\.'
            )
            $percentMatch = [regex]::Match($line, '\*\*Evolução:\*\*\s*(?<percent>\d{1,3}(?:[.,]\d+)?)%')
            if ($statusMatch.Success -or $percentMatch.Success) {
                if (-not $statusMatch.Success -or -not $percentMatch.Success) {
                    throw "A tarefa BK-$($currentTask.Id) tem situação ou percentual em formato incompleto na linha $lineNumber."
                }

                $percent = [double]::Parse(
                    $percentMatch.Groups['percent'].Value.Replace(',', '.'),
                    [System.Globalization.CultureInfo]::InvariantCulture
                )
                $status = $statusMatch.Groups['status'].Value
                if ($percent -lt 0 -or $percent -gt 100) {
                    throw "A tarefa BK-$($currentTask.Id) tem percentual fora de 0–100: $percent."
                }
                if ($status -eq 'Concluída' -and $percent -ne 100) {
                    throw "A tarefa BK-$($currentTask.Id) está concluída, mas não está em 100%."
                }
                if ($status -eq 'Não iniciada' -and $percent -ne 0) {
                    throw "A tarefa BK-$($currentTask.Id) ainda não iniciada precisa estar em 0%."
                }

                $tasks.Add([pscustomobject]@{
                    Id = $currentTask.Id
                    Status = $status
                    Progress = $percent
                })
                $currentTask = $null
            }
        }
    }
}
finally {
    $reader.Dispose()
}

if ($null -ne $currentTask) {
    throw "A tarefa BK-$($currentTask.Id) não tem linha de situação/percentual."
}
if ($readingSummary) {
    throw 'O início das tarefas detalhadas ("## 00") não foi encontrado.'
}
if ($null -eq $declaredCounts -or $null -eq $declaredProgress) {
    throw 'O resumo atual do backlog não contém contagens e percentuais reconhecíveis.'
}
if ($tasks.Count -eq 0) {
    throw 'Nenhuma tarefa detalhada foi encontrada no backlog.'
}

$completed = @($tasks | Where-Object { $_.Status -eq 'Concluída' }).Count
$inProgress = @($tasks | Where-Object { $_.Status -eq 'Em andamento' }).Count
$notStarted = @($tasks | Where-Object { $_.Status -eq 'Não iniciada' }).Count
$sumProgress = ($tasks | Measure-Object -Property Progress -Sum).Sum
$calculatedPercent = [Math]::Round(
    ([double]$sumProgress / [double]$tasks.Count),
    1,
    [System.MidpointRounding]::AwayFromZero
)
$calculatedRemaining = [Math]::Round(
    (100 - ([double]$sumProgress / [double]$tasks.Count)),
    1,
    [System.MidpointRounding]::AwayFromZero
)

if ($declaredCounts.Total -ne $tasks.Count -or
    $declaredCounts.Done -ne $completed -or
    $declaredCounts.InProgress -ne $inProgress -or
    $declaredCounts.NotStarted -ne $notStarted) {
    throw "Resumo divergente: detalhado=$($tasks.Count) tarefas ($completed concluídas, $inProgress em andamento, $notStarted não iniciadas); resumo=$($declaredCounts.Total) ($($declaredCounts.Done), $($declaredCounts.InProgress), $($declaredCounts.NotStarted))."
}

if ([Math]::Abs($declaredProgress.Complete - $calculatedPercent) -gt 0.0001 -or
    [Math]::Abs($declaredProgress.Remaining - $calculatedRemaining) -gt 0.0001) {
    throw "Percentual divergente: calculado=$calculatedPercent% concluído e $calculatedRemaining% restante; resumo=$($declaredProgress.Complete)% concluído e $($declaredProgress.Remaining)% restante."
}

if (-not [string]::IsNullOrWhiteSpace($StructuredPath)) {
    $structuredFullPath = [System.IO.Path]::GetFullPath($StructuredPath)
    if (-not [System.IO.File]::Exists($structuredFullPath)) {
        throw "Arquivo estruturado do backlog não encontrado: $structuredFullPath"
    }
    $structuredFile = Get-Item -LiteralPath $structuredFullPath
    if ($structuredFile.Length -gt $maxFileBytes) {
        throw "Arquivo estruturado do backlog excede o limite de 8 MiB: $($structuredFile.Length) bytes."
    }

    try {
        $structuredTasks = @(
            ConvertFrom-Json -InputObject ([System.IO.File]::ReadAllText(
                $structuredFullPath,
                [System.Text.UTF8Encoding]::new($false, $true)
            )) -ErrorAction Stop
        )
    }
    catch {
        throw "Arquivo estruturado do backlog não contém JSON válido: $($_.Exception.Message)"
    }

    if ($structuredTasks.Count -ne $tasks.Count) {
        throw "Quantidade divergente entre backlog detalhado ($($tasks.Count)) e estruturado ($($structuredTasks.Count))."
    }

    for ($index = 0; $index -lt $tasks.Count; $index++) {
        $structuredTask = $structuredTasks[$index]
        foreach ($propertyName in @('id', 'status', 'progress')) {
            if ($null -eq $structuredTask.PSObject.Properties[$propertyName]) {
                throw "Tarefa estruturada na posição $($index + 1) não contém o campo obrigatório '$propertyName'."
            }
        }

        $expectedId = "BK-$($tasks[$index].Id)"
        $expectedProgress = [double]$tasks[$index].Progress / 100
        $structuredProgress = 0.0
        if (-not [double]::TryParse(
            [string]$structuredTask.progress,
            [System.Globalization.NumberStyles]::Float,
            [System.Globalization.CultureInfo]::InvariantCulture,
            [ref]$structuredProgress
        )) {
            throw "A tarefa estruturada $($structuredTask.id) tem evolução inválida."
        }

        if ([string]$structuredTask.id -cne $expectedId -or
            [string]$structuredTask.status -cne [string]$tasks[$index].Status -or
            [Math]::Abs($structuredProgress - $expectedProgress) -gt 0.000001) {
            throw "Registro estruturado divergente na tarefa ${expectedId}: Markdown=($($tasks[$index].Status), $($tasks[$index].Progress)%); JSON=($($structuredTask.status), $($structuredProgress * 100)%)."
        }
    }
}

Write-Output "Backlog conferido: $($tasks.Count) tarefas; $completed concluídas, $inProgress em andamento e $notStarted não iniciadas; $calculatedPercent% concluído, $calculatedRemaining% restante."
