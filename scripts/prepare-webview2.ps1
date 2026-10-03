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

# Wails' generated NSIS macro embeds this official Evergreen Bootstrapper. Keep
# the known hash pinned, but authenticate Microsoft's rolling bootstrapper with
# Authenticode so legitimate Evergreen updates do not permanently break builds.
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
$signature = Get-AuthenticodeSignature -FilePath $Destination
$publisher = $null
$hasCodeSigningEku = $false
if ($null -ne $signature.SignerCertificate) {
    $publisher = $signature.SignerCertificate.GetNameInfo(
        [System.Security.Cryptography.X509Certificates.X509NameType]::SimpleName,
        $false
    )
    foreach ($extension in $signature.SignerCertificate.Extensions) {
        if ($extension -is [System.Security.Cryptography.X509Certificates.X509EnhancedKeyUsageExtension]) {
            $hasCodeSigningEku = @(
                $extension.EnhancedKeyUsages |
                    Where-Object { $_.Value -eq '1.3.6.1.5.5.7.3.3' }
            ).Count -gt 0
        }
    }
}

if ($signature.Status -ne [System.Management.Automation.SignatureStatus]::Valid -or
    $publisher -cne 'Microsoft Corporation' -or
    -not $hasCodeSigningEku) {
    Remove-Item -LiteralPath $Destination -Force -ErrorAction SilentlyContinue
    $signatureStatus = [string]$signature.Status
    if ([string]::IsNullOrWhiteSpace($publisher)) { $publisher = '<ausente>' }
    throw "WebView2 Bootstrapper rejeitado: Authenticode=$signatureStatus; signatário=$publisher; EKU de assinatura de código=$hasCodeSigningEku."
}

if ($actualHash -ne $expectedSha256) {
    Write-Warning "O SHA-256 do Bootstrapper Evergreen mudou; aceito somente após Authenticode válido de Microsoft Corporation. SHA-256 atual: $actualHash; certificado: $($signature.SignerCertificate.Thumbprint)."
}

$destinationFull = [System.IO.Path]::GetFullPath($Destination)
$installerSourceFull = [System.IO.Path]::GetFullPath($installerSource)
if (-not [string]::Equals($destinationFull, $installerSourceFull, [System.StringComparison]::OrdinalIgnoreCase)) {
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $installerSourceFull) | Out-Null
    Copy-Item -LiteralPath $destinationFull -Destination $installerSourceFull -Force
}

Write-Output "WebView2 Bootstrapper validado e centralizado em artifacts/cache: $destinationFull"
