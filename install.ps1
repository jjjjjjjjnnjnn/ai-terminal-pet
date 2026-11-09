param(
  [string]$Source = (Join-Path $PSScriptRoot 'bin\AI-Terminal-Pet.exe'),
  [string]$Destination = (Join-Path $env:LOCALAPPDATA 'AI-Terminal-Pet\AI-Terminal-Pet.exe'),
  [string]$ShimSource = (Join-Path $PSScriptRoot 'bin\pet-statusline.exe'),
  [string]$ShimDestination = (Join-Path $env:LOCALAPPDATA 'AI-Terminal-Pet\pet-statusline.exe'),
  [string]$ProtocolSource = (Join-Path $PSScriptRoot 'reporters\protocol.js'),
  [string]$ProtocolDestination = (Join-Path $env:LOCALAPPDATA 'AI-Terminal-Pet\protocol.js'),
  [string]$BundleSource = (Join-Path $PSScriptRoot 'integrations\deepseek-harness\ai-terminal-pet-bridge'),
  [string]$BundleDestination = (Join-Path $env:LOCALAPPDATA 'AI-Terminal-Pet\bundles\ai-terminal-pet-bridge')
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
# Keep the installed reporter protocol in sync with the repository.
#
# The host-CLI hooks (Claude Code, Codex, PowerShell) require the installed
# copy of protocol.js. It owns the per-user pipe name, the v1 schema, the
# legal state set, redaction, snapshot naming and on-demand launch. A stale
# copy - or a second, independent protocol implementation - is exactly what
# disconnected every reporter from the overlay on 2026-11-05. Every install
# must therefore refresh this single source of truth.
if (-not (Test-Path -LiteralPath $ProtocolSource -PathType Leaf)) {
  throw "Reporter protocol not found: $ProtocolSource"
}
New-Item -ItemType Directory -Path (Split-Path -Parent $ProtocolDestination) -Force | Out-Null
Copy-Item -LiteralPath $ProtocolSource -Destination $ProtocolDestination -Force -ErrorAction Stop
$protocolSourceHash = (Get-FileHash -LiteralPath $ProtocolSource -Algorithm SHA256).Hash
$protocolDestinationHash = (Get-FileHash -LiteralPath $ProtocolDestination -Algorithm SHA256).Hash
if ($protocolSourceHash -ne $protocolDestinationHash) {
  throw "Installed protocol.js hash does not match repository copy"
}

# Ship the DeepSeek Harness bundle next to the overlay so it has a stable
# absolute path to import with plugin_manager(install_bundle). The bundle is
# only an adapter: it loads the installed protocol.js above and owns no pipe
# name, state set or snapshot format of its own. Syncing it on every install
# prevents a second, drifting copy from disconnecting that reporter too.
$bundleFiles = @(
  'package.json', 'cordis.patch.yml', 'index.js', 'icon.svg', 'README.md',
  'locale\en.json', 'locale\zh.json'
)
foreach ($relative in $bundleFiles) {
  $from = Join-Path $BundleSource $relative
  $to = Join-Path $BundleDestination $relative
  if (-not (Test-Path -LiteralPath $from -PathType Leaf)) {
    throw "Harness bundle file not found: $from"
  }
  New-Item -ItemType Directory -Path (Split-Path -Parent $to) -Force | Out-Null
  Copy-Item -LiteralPath $from -Destination $to -Force -ErrorAction Stop
  if ((Get-FileHash -LiteralPath $from -Algorithm SHA256).Hash -ne
      (Get-FileHash -LiteralPath $to -Algorithm SHA256).Hash) {
    throw "Installed Harness bundle hash does not match repository copy: $relative"
  }
}

Get-Item -LiteralPath $Destination, $ShimDestination, $ProtocolDestination | Select-Object FullName, Length, LastWriteTime
Get-Item -LiteralPath $BundleDestination | Select-Object FullName, LastWriteTime
