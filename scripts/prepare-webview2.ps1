[CmdletBinding()]
param(
    [string]$Destination
)

$ErrorActionPreference = 'Stop'
$projectRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
if ([string]::IsNullOrWhiteSpace($Destination)) {
    $Destination = Join-Path $projectRoot 'artifacts\cache\webview2\MicrosoftEdgeWebview2Setup.exe'
}
$installerSource = Join-Path $projectRoot 'build\windows\installer\tmp\MicrosoftEdgeWebview2Setup.exe'

# Wails' generated NSIS macro embeds this official Evergreen Bootstrapper. The
# hash is pinned so a changed or substituted download fails the build closed.
$uri = 'https://go.microsoft.com/fwlink/p/?LinkId=2124703'
$expectedSha256 = '48A7B31419A8EB4FFFDC7B6A02F6B4DFDA60687FC897116BE15370E10C2B66A7'
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

$destinationFull = [System.IO.Path]::GetFullPath($Destination)
$installerSourceFull = [System.IO.Path]::GetFullPath($installerSource)
if (-not [string]::Equals($destinationFull, $installerSourceFull, [System.StringComparison]::OrdinalIgnoreCase)) {
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $installerSourceFull) | Out-Null
    Copy-Item -LiteralPath $destinationFull -Destination $installerSourceFull -Force
}

Write-Output "WebView2 Bootstrapper validado e centralizado em artifacts/cache: $destinationFull"
