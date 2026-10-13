[CmdletBinding()]
param()

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$projectRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
$artifactRoot = Join-Path $projectRoot 'artifacts'
$env:GOCACHE = Join-Path $artifactRoot 'cache\go-build'
$env:GOMODCACHE = Join-Path $artifactRoot 'cache\go-mod'
$env:GOTMPDIR = Join-Path $artifactRoot 'cache\go-tmp'
New-Item -ItemType Directory -Force -Path $env:GOCACHE, $env:GOMODCACHE, $env:GOTMPDIR | Out-Null

$goCommand = if ($env:MAJUCAU_GO) { $env:MAJUCAU_GO } else { (Get-Command go -ErrorAction Stop).Source }
$violations = [System.Collections.Generic.List[string]]::new()

Push-Location $projectRoot
try {
    $modulePath = (& $goCommand list -m -f '{{.Path}}').Trim()
    if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($modulePath)) {
        throw 'Não foi possível identificar o módulo Go do Majucau.'
    }

    $standardPackages = @(& $goCommand list std)
    if ($LASTEXITCODE -ne 0) {
        throw 'Não foi possível ler a lista de pacotes padrão do Go.'
    }
    $standardImports = [System.Collections.Generic.HashSet[string]]::new([System.StringComparer]::Ordinal)
    foreach ($package in $standardPackages) {
        if (-not [string]::IsNullOrWhiteSpace($package)) {
            [void]$standardImports.Add($package.Trim())
        }
    }

    $packageLines = @(& $goCommand list -f '{{.ImportPath}}|{{join .Imports " "}}' ./...)
    if ($LASTEXITCODE -ne 0) {
        throw 'Não foi possível montar o grafo de imports do projeto Go.'
    }

    $allowedLayerImports = @{
        "$modulePath/internal/domain"      = @()
        "$modulePath/internal/financial"   = @("$modulePath/internal/domain")
        "$modulePath/internal/application" = @("$modulePath/internal/domain")
    }
    $forbiddenLayerPrefixes = @('net', 'database', 'os', 'path/filepath', 'io/fs', 'syscall', 'plugin', 'unsafe')

    foreach ($line in $packageLines) {
        $separator = $line.IndexOf('|')
        if ($separator -lt 1) {
            throw 'A saída do inventário de imports Go está incompleta.'
        }

        $packagePath = $line.Substring(0, $separator)
        $importsText = $line.Substring($separator + 1)
        $imports = if ([string]::IsNullOrWhiteSpace($importsText)) { @() } else { @($importsText -split '\s+') }

        foreach ($importPath in $imports) {
            if ($importPath.StartsWith("$modulePath/cmd/", [System.StringComparison]::Ordinal)) {
                $violations.Add("$packagePath importa o ponto de composição $importPath.")
            }

            if (-not $allowedLayerImports.ContainsKey($packagePath)) {
                continue
            }
            if ($forbiddenLayerPrefixes | Where-Object { $importPath -eq $_ -or $importPath.StartsWith("$_/", [System.StringComparison]::Ordinal) }) {
                $violations.Add("$packagePath importa $importPath, uma dependência de rede, sistema ou infraestrutura proibida nesta camada.")
                continue
            }
            if ($standardImports.Contains($importPath) -or $allowedLayerImports[$packagePath] -contains $importPath) {
                continue
            }
            $violations.Add("$packagePath importa $importPath; esta camada aceita somente biblioteca padrão e os contratos documentados.")
        }
    }
}
finally {
    Pop-Location
}

if ($violations.Count -gt 0) {
    $orderedViolations = @($violations | Sort-Object -Unique)
    throw "Limites de arquitetura Go violados:`n- $($orderedViolations -join "`n- ")"
}

Write-Output "Limites Go conferidos: domínio usa biblioteca padrão sem rede ou infraestrutura; aplicação e cálculos dependem apenas do domínio; nenhum pacote depende de cmd/*."
