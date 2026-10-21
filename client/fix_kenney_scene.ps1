param(
    [Parameter(Mandatory)][string]$Source,
    [Parameter(Mandatory)][string]$Target
)

$sourcePath = (Resolve-Path -LiteralPath $Source).Path
$targetPath = [IO.Path]::GetFullPath($Target)

if ($sourcePath -eq $targetPath) {
    throw "A origem e o destino precisam ser diferentes."
}

$bytes = [IO.File]::ReadAllBytes($sourcePath)
$jsonLength = [BitConverter]::ToInt32($bytes, 12)
$json = [Text.Encoding]::UTF8.GetString($bytes, 20, $jsonLength)

$from = '"scenes":[{"nodes":[1]}]'
$to = '"scenes":[{"nodes":[0]}]'

if (-not $json.Contains($from)) {
    throw "A hierarquia esperada não foi encontrada em $Source"
}

$corrected = [Text.Encoding]::UTF8.GetBytes($json.Replace($from, $to))
if ($corrected.Length -ne $jsonLength) {
    throw "A correção alterou o tamanho do bloco JSON."
}

[Array]::Copy($corrected, 0, $bytes, 20, $jsonLength)
[IO.File]::WriteAllBytes($targetPath, $bytes)