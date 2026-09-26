param(
  [string]$Source = (Join-Path $PSScriptRoot 'bin\AI-Terminal-Pet.exe'),
  [string]$Destination = (Join-Path $env:LOCALAPPDATA 'AI-Terminal-Pet\AI-Terminal-Pet.exe'),
  [string]$ShimSource = (Join-Path $PSScriptRoot 'bin\pet-statusline.exe'),
  [string]$ShimDestination = (Join-Path $env:LOCALAPPDATA 'AI-Terminal-Pet\pet-statusline.exe')
)

$ErrorActionPreference = 'Stop'

# Copy-InstalledBinary replaces a binary that may be running right now.
#
# The overlay is launched on demand by the reporters, so it can be alive (or
# come back) at any moment, and the statusline shim is spawned on every
# render. A single Stop-Process plus a fixed sleep is not enough: with AV
# hooks the process can take well over a second to release its image, and an
# active reporter relaunches the overlay within seconds. The old script
# therefore failed at random with "file is being used by another process",
# which is exactly what an upgrade on a machine that is actually using the
# tool looks like.
#
# So: clear whatever is holding the file, wait for it to really exit, and
# retry until the copy lands.
function Copy-InstalledBinary {
  param(
    [Parameter(Mandatory)][string]$From,
    [Parameter(Mandatory)][string]$To
  )
  $parent = Split-Path -Parent $To
  New-Item -ItemType Directory -Path $parent -Force | Out-Null
  $names = @('AI-Terminal-Pet', 'pet-statusline')
  for ($attempt = 1; $attempt -le 20; $attempt++) {
    $running = Get-Process -Name $names -ErrorAction SilentlyContinue
    if ($running) {
      $running | Stop-Process -Force -ErrorAction SilentlyContinue
      $running | Wait-Process -Timeout 10 -ErrorAction SilentlyContinue
    }
    try {
      Copy-Item -LiteralPath $From -Destination $To -Force -ErrorAction Stop
      return
    } catch {
      if ($attempt -eq 20) { throw }
      Start-Sleep -Milliseconds 500
    }
  }
}

foreach ($pair in @(@($Source, $Destination), @($ShimSource, $ShimDestination))) {
  if (-not (Test-Path -LiteralPath $pair[0] -PathType Leaf)) {
    throw "Build output not found: $($pair[0])"
  }
}

foreach ($pair in @(@($Source, $Destination), @($ShimSource, $ShimDestination))) {
  Copy-InstalledBinary -From $pair[0] -To $pair[1]
  $sourceHash = (Get-FileHash -LiteralPath $pair[0] -Algorithm SHA256).Hash
  $destinationHash = (Get-FileHash -LiteralPath $pair[1] -Algorithm SHA256).Hash
  if ($sourceHash -ne $destinationHash) {
    throw "Installed executable hash does not match build output: $($pair[1])"
  }
}
Get-Item -LiteralPath $Destination, $ShimDestination | Select-Object FullName, Length, LastWriteTime
