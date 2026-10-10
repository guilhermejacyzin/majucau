Set-StrictMode -Version Latest

function Read-BoundedTextFile {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [ValidateNotNullOrEmpty()]
        [string]$Path,

        [Parameter(Mandatory = $true)]
        [ValidateRange(1, 67108864)]
        [long]$MaxBytes
    )

    $file = $null
    $buffered = $null
    $textReader = $null
    try {
        $file = [System.IO.File]::Open(
            $Path,
            [System.IO.FileMode]::Open,
            [System.IO.FileAccess]::Read,
            [System.IO.FileShare]::ReadWrite
        )
        if ($file.Length -gt $MaxBytes) {
            throw [System.IO.InvalidDataException]::new("Arquivo excede o limite de $MaxBytes bytes.")
        }

        $buffered = [System.IO.MemoryStream]::new()
        [byte[]]$buffer = New-Object byte[] 8192
        while ($true) {
            $remaining = $MaxBytes + 1 - $buffered.Length
            if ($remaining -le 0) {
                throw [System.IO.InvalidDataException]::new("Arquivo excede o limite de $MaxBytes bytes.")
            }

            $read = $file.Read($buffer, 0, [int][Math]::Min($buffer.Length, $remaining))
            if ($read -eq 0) {
                break
            }

            $buffered.Write($buffer, 0, $read)
            if ($buffered.Length -gt $MaxBytes) {
                throw [System.IO.InvalidDataException]::new("Arquivo excede o limite de $MaxBytes bytes.")
            }
        }

        $buffered.Position = 0
        $utf8 = [System.Text.UTF8Encoding]::new($false, $true)
        $textReader = [System.IO.StreamReader]::new($buffered, $utf8, $true, 4096, $true)
        return $textReader.ReadToEnd()
    }
    finally {
        if ($null -ne $textReader) { $textReader.Dispose() }
        if ($null -ne $buffered) { $buffered.Dispose() }
        if ($null -ne $file) { $file.Dispose() }
    }
}

Export-ModuleMember -Function Read-BoundedTextFile
