[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSScriptRoot
$extensions = '\.(go|ts|tsx|js|json|yml|yaml|md|ps1|sql|nsi|xml|toml|env|txt)$'
$patterns = @(
    '-----BEGIN (RSA |EC |OPENSSH )?PRIVATE KEY-----',
    '\bgh[pousr]_[A-Za-z0-9_]{30,}\b',
    '\bAKIA[0-9A-Z]{16}\b',
    '(?i)\b(api[_-]?key|client[_-]?secret|access[_-]?token|refresh[_-]?token|password)\b\s*[:=]\s*[''"][^''"]{12,}[''"]'
)

Push-Location $projectRoot
try {
    $files = @(git ls-files --cached --others --exclude-standard) |
        Where-Object { $_ -match $extensions -and $_ -notmatch '^docs/source/' }
    $findings = @()
    foreach ($file in $files) {
        $lineNumber = 0
        foreach ($line in Get-Content -LiteralPath $file -ErrorAction Stop) {
            $lineNumber++
            foreach ($pattern in $patterns) {
                if ($line -match $pattern -and $line -notmatch '(?i)(example|placeholder|redacted|dummy|test-only)') {
                    $findings += "$($file):$lineNumber"
                    break
                }
            }
        }
    }
    if ($findings.Count -gt 0) {
        Write-Error ("Possível secret detectado em: " + ($findings -join ', '))
    }
}
finally {
    Pop-Location
}

Write-Host 'Scan local de secrets: PASS'
