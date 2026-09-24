param(
  [string]$Source = (Join-Path $PSScriptRoot 'bin\AI-Terminal-Pet.exe'),
  [string]$Destination = (Join-Path $env:LOCALAPPDATA 'AI-Terminal-Pet\AI-Terminal-Pet.exe'),
  [string]$ShimSource = (Join-Path $PSScriptRoot 'bin\pet-statusline.exe'),
  [string]$ShimDestination = (Join-Path $env:LOCALAPPDATA 'AI-Terminal-Pet\pet-statusline.exe')
)

$ErrorActionPreference = 'Stop'
foreach ($pair in @(@($Source, $Destination), @($ShimSource, $ShimDestination))) {
  if (-not (Test-Path -LiteralPath $pair[0] -PathType Leaf)) {
    throw "Build output not found: $($pair[0])"
  }
}
$parent = Split-Path -Parent $Destination
New-Item -ItemType Directory -Path $parent -Force | Out-Null
Get-Process -Name 'AI-Terminal-Pet' -ErrorAction SilentlyContinue | Stop-Process -Force
Start-Sleep -Milliseconds 500
foreach ($pair in @(@($Source, $Destination), @($ShimSource, $ShimDestination))) {
  Copy-Item -LiteralPath $pair[0] -Destination $pair[1] -Force
  $sourceHash = (Get-FileHash -LiteralPath $pair[0] -Algorithm SHA256).Hash
  $destinationHash = (Get-FileHash -LiteralPath $pair[1] -Algorithm SHA256).Hash
  if ($sourceHash -ne $destinationHash) {
    throw "Installed executable hash does not match build output: $($pair[1])"
  }
}
Get-Item -LiteralPath $Destination, $ShimDestination | Select-Object FullName, Length, LastWriteTime
