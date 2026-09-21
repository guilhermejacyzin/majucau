[CmdletBinding()]
param(
    [string]$Destination
)

$ErrorActionPreference = 'Stop'
$projectRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
if ([string]::IsNullOrWhiteSpace($Destination)) {
    $Destination = Join-Path $projectRoot 'build\windows\installer\tmp\MicrosoftEdgeWebview2Setup.exe'
}

# Wails' generated NSIS macro embeds this official Evergreen Bootstrapper. The
# hash is pinned so a changed or substituted download fails the build closed.
$uri = 'https://go.microsoft.com/fwlink/p/?LinkId=2124703'
$expectedSha256 = '81C01751C8CC385A5991ABB104205D42AC70094350EE8FB9E8EA580B51BB9554'
$destinationDirectory = Split-Path -Parent $Destination
New-Item -ItemType Directory -Force -Path $destinationDirectory | Out-Null

$needsDownload = $true
if (Test-Path -LiteralPath $Destination -PathType Leaf) {
    $currentHash = (Get-FileHash -Algorithm SHA256 -LiteralPath $Destination).Hash.ToUpperInvariant()
    $needsDownload = $currentHash -ne $expectedSha256
}

if ($needsDownload) {
    Invoke-WebRequest -Uri $uri -OutFile $Destination -UseBasicParsing
}

$actualHash = (Get-FileHash -Algorithm SHA256 -LiteralPath $Destination).Hash.ToUpperInvariant()
if ($actualHash -ne $expectedSha256) {
    Remove-Item -LiteralPath $Destination -Force -ErrorAction SilentlyContinue
    throw "WebView2 Bootstrapper rejeitado: SHA-256 inesperado ($actualHash)."
}

Write-Output "WebView2 Bootstrapper validado: $Destination"
