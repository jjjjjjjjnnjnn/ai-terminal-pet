# package.ps1 - build the portable Windows zip that Release assets ship.
#
#   powershell -NoProfile -ExecutionPolicy Bypass -File scripts\package.ps1
#   powershell -NoProfile -ExecutionPolicy Bypass -File scripts\package.ps1 -Build
#
# The zip is a *subset of the repository with the repository's own layout*:
#
#   bin\AI-Terminal-Pet.exe      bin\pet-statusline.exe    bin\pet-hook.exe
#   reporters\protocol.js
#   integrations\deepseek-harness\ai-terminal-pet-bridge\  (7 files)
#   scripts\install.ps1          scripts\install.cmd
#   install.cmd                  (one-click forwarder at the zip root)
#   README.install.txt           SHA256SUMS.txt           LICENSE
#
# That layout is the whole design. scripts\install.ps1 derives everything it
# reads from `Split-Path -Parent $PSScriptRoot`, so a tree that is byte-for-byte
# the repository's shape needs no second install implementation and no path
# rewriting: the same script the repository uses installs the zip. Anything that
# ships a different shape would have to be verified separately, and a verifier
# that checks a different set than the installer writes is exactly how a stale
# protocol.js survived an upgrade once and disconnected every reporter.
#
# Exit codes: 0 success, 1 build or packaging failed, 2 bad usage.
#
# ASCII only, same reason install.ps1 is: Windows PowerShell 5.1 reads a
# BOM-less script in the ANSI code page.

param(
  [switch]$Build,
  [string]$OutputDirectory = (Join-Path (Split-Path -Parent $PSScriptRoot) 'dist'),
  [string]$Version = ''
)

$ErrorActionPreference = 'Stop'
$NewLine = [Environment]::NewLine
$RepoRoot = Split-Path -Parent $PSScriptRoot

# The zip's version is the *overlay's* version - the number `--version` prints -
# not the DSH plugin package's. They are two independent lines (the overlay is at
# 0.9.0 while the plugin package is at 1.2.x) and naming a zip after the plugin
# version would put two numbers on one product. Reading the constant out of the
# source instead of taking a literal here means a future version bump moves the
# filename in the same commit, with no second place to forget.
if (-not $Version) {
  $appFile = Join-Path $RepoRoot 'internal\app\app_windows.go'
  if (-not (Test-Path -LiteralPath $appFile -PathType Leaf)) {
    Write-Output '[fail] package: internal\app\app_windows.go is missing, so the overlay version cannot be read'
    exit 1
  }
  $match = Select-String -LiteralPath $appFile -Pattern 'const\s+Version\s*=\s*"([^"]+)"' | Select-Object -First 1
  if (-not $match) {
    Write-Output '[fail] package: no `const Version = "..."` line in internal\app\app_windows.go'
    exit 1
  }
  $Version = $match.Matches[0].Groups[1].Value
}
Write-Output ('package: overlay version ' + $Version)

function Get-FileSha256 {
  param([Parameter(Mandatory)][string]$Path)
  $stream = [System.IO.File]::OpenRead($Path)
  try {
    $sha = [System.Security.Cryptography.SHA256]::Create()
    try {
      $bytes = $sha.ComputeHash($stream)
    } finally {
      $sha.Dispose()
    }
  } finally {
    $stream.Dispose()
  }
  return ([System.BitConverter]::ToString($bytes) -replace '-', '').ToUpperInvariant()
}

# The three binaries are built with the exact flags AGENTS.md pins, because
# scripts\install.ps1 -VerifyOnly compares an installed copy against this tree.
# A build that carried the VCS stamp would differ between machines and dates and
# make that check report a false failure.
function Invoke-Build {
  $targets = @(
    @{ Out = 'bin\AI-Terminal-Pet.exe'; Package = './cmd/ai-terminal-pet' },
    @{ Out = 'bin\pet-statusline.exe'; Package = './cmd/pet-statusline' },
    @{ Out = 'bin\pet-hook.exe'; Package = './cmd/pet-hook' }
  )
  foreach ($target in $targets) {
    $full = Join-Path $RepoRoot $target.Out
    $parent = Split-Path -Parent $full
    New-Item -ItemType Directory -Path $parent -Force | Out-Null
    Write-Output ('build: ' + $target.Out)
    & go build -trimpath -buildvcs=false -ldflags '-s -w' -o $full $target.Package
    if ($LASTEXITCODE -ne 0) {
      throw ('go build failed for ' + $target.Package + ' (exit ' + $LASTEXITCODE + ')')
    }
  }
}

$manifestEntries = @(
  @{ Source = 'bin\AI-Terminal-Pet.exe'; Target = 'bin\AI-Terminal-Pet.exe' },
  @{ Source = 'bin\pet-statusline.exe'; Target = 'bin\pet-statusline.exe' },
  @{ Source = 'bin\pet-hook.exe'; Target = 'bin\pet-hook.exe' },
  @{ Source = 'reporters\protocol.js'; Target = 'reporters\protocol.js' },
  @{ Source = 'scripts\install.ps1'; Target = 'scripts\install.ps1' },
  @{ Source = 'scripts\install.cmd'; Target = 'scripts\install.cmd' },
  @{ Source = 'LICENSE'; Target = 'LICENSE' }
)

$bundleFiles = @(
  'package.json', 'cordis.patch.yml', 'index.js', 'icon.svg', 'README.md',
  'locale\en.json', 'locale\zh.json'
)
$bundleSource = 'integrations\deepseek-harness\ai-terminal-pet-bridge'
foreach ($relative in $bundleFiles) {
  $manifestEntries += @{
    Source = ($bundleSource + '\' + $relative)
    Target = ($bundleSource + '\' + $relative)
  }
}

# install.cmd at the zip root is generated, not committed: it exists so a user
# who just unzipped can double-click one file, and it must not become a third
# place where install logic can drift. It forwards, nothing else.
$rootForwarder = @'
@echo off
rem Double-click me. This is a forwarder to scripts\install.cmd, which is the
rem real wrapper around scripts\install.ps1 (see that file for why cmd.exe is
rem the entry point and not powershell.exe directly).
call "%~dp0scripts\install.cmd" %*
set "_STATUS=%ERRORLEVEL%"
if "%~1"=="" (
  echo.
  echo Press any key to close this window.
  pause > nul
)
exit /b %_STATUS%
'@

$installNotes = @'
ai-terminal-pet - portable Windows package
=========================================

1. Double-click install.cmd.
   It copies the overlay, the statusline shim, the native hook and the
   DeepSeek Harness bundle into %LOCALAPPDATA%\AI-Terminal-Pet\ and re-hashes
   every file it copied.

2. Optional: run install.cmd -AddToPath so "ai-terminal-pet" works as a bare
   command. Open a new terminal afterwards.

3. Connect your terminal. Installing the files is not enough on its own: each
   terminal has to be told to report. See "Connect your terminal" in README.md.

   Claude Code:  & "%LOCALAPPDATA%\AI-Terminal-Pet\pet-hook.exe" --mode=claude
                 on SessionStart / SessionEnd / UserPromptSubmit / PreToolUse /
                 PostToolUse / Stop / Notification in settings.json
   Antigravity:  statusLine.command -> pet-statusline.exe
                 plus pet-hook.exe --mode=antigravity for events
   OpenCode:     add file:///.../reporters/opencode.js to "plugin"
   Any CLI:      node reporters\generic.js --source=... --state=...

4. Check everything at any time:
   %LOCALAPPDATA%\AI-Terminal-Pet\AI-Terminal-Pet.exe --doctor
   install.cmd -VerifyOnly

5. Uninstall:
   install.cmd -Uninstall

This release is NOT code-signed. The first run shows SmartScreen's
"unknown publisher" prompt: click "More info", then "Run anyway". That prompt
is a code-signing certificate this project does not have yet, not a problem
with the file. SHA256SUMS.txt lists the hash of every file in this package, so
you can verify what you extracted:

   powershell -NoProfile -Command "Get-Content SHA256SUMS.txt"

Nothing in this project talks to a network, requires an account, or sends
telemetry. Everything it writes stays under %LOCALAPPDATA%\AI-Terminal-Pet\.
'@

$stage = Join-Path $OutputDirectory 'ai-terminal-pet'
$archiveName = 'ai-terminal-pet-' + $Version + '-win-x64.zip'
$archivePath = Join-Path $OutputDirectory $archiveName

try {
  if ($Build) { Invoke-Build }

  # ZipArchive is not a PowerShell built-in: Windows PowerShell 5.1 resolves it
  # only after the two assemblies are loaded, and PowerShell 7 ships it already.
  # Loading them by name is harmless on both.
  Add-Type -AssemblyName System.IO.Compression
  Add-Type -AssemblyName System.IO.Compression.FileSystem

  # Every source is read before anything is written, so a missing report,
  # protocol file or binary stops the run while the output directory is still
  # untouched instead of producing a package that is quietly incomplete.
  $missing = @()
  foreach ($entry in $manifestEntries) {
    $full = Join-Path $RepoRoot $entry.Source
    if (-not (Test-Path -LiteralPath $full -PathType Leaf)) { $missing += $entry.Source }
  }
  if ($missing.Count -gt 0) {
    Write-Output '[fail] package: these sources are missing from the checkout:'
    foreach ($item in $missing) { Write-Output ('       ' + $item) }
    Write-Output '       Build the binaries first: scripts\package.ps1 -Build'
    exit 1
  }

  # The zip's overlay and the plugin package's vendored overlay are the same
  # product. If they ever diverge, a Harness user and a zip user run different
  # builds under one version number, so packaging refuses instead of guessing
  # which side is stale.
  #
  # PET_PACKAGE_SKIP_VENDOR_CHECK=1 turns this into a warning. The one legitimate
  # caller is CI, where it always diverges: go.mod's `go 1.23.0` is a floor, so a
  # runner compiles with go1.23.0 while the committed vendor binary came from a
  # newer local Go, and different Go releases emit different bytes for identical
  # sources. Refusing there would fail every run for a non-problem; the check
  # still bites on the machine that cuts the release, which is the only machine
  # whose build can be compared to the vendor copy in the first place.
  $vendored = Join-Path $RepoRoot 'packages\dsh-ai-terminal-pet\vendor\AI-Terminal-Pet.exe'
  if (Test-Path -LiteralPath $vendored -PathType Leaf) {
    $zipOverlay = Get-FileSha256 -Path (Join-Path $RepoRoot 'bin\AI-Terminal-Pet.exe')
    $pluginOverlay = Get-FileSha256 -Path $vendored
    if ($zipOverlay -ne $pluginOverlay) {
      if ($env:PET_PACKAGE_SKIP_VENDOR_CHECK -eq '1') {
        Write-Output '[warn] overlay differs from the vendored copy, cross-check skipped (PET_PACKAGE_SKIP_VENDOR_CHECK=1)'
        Write-Output ('       bin:    ' + $zipOverlay)
        Write-Output ('       vendor: ' + $pluginOverlay)
      } else {
        Write-Output '[fail] package: bin\AI-Terminal-Pet.exe and packages\dsh-ai-terminal-pet\vendor\AI-Terminal-Pet.exe differ.'
        Write-Output ('       bin:    ' + $zipOverlay)
        Write-Output ('       vendor: ' + $pluginOverlay)
        Write-Output '       One of them is stale. Rebuild, then regenerate vendor/manifest.json before packaging.'
        Write-Output '       (CI sets PET_PACKAGE_SKIP_VENDOR_CHECK=1 because its Go release differs.)'
        exit 1
      }
    } else {
      Write-Output ('[ok] overlay matches the vendored copy (' + $zipOverlay.Substring(0, 12) + '...)')
    }
  } else {
    Write-Output '[warn] packages\dsh-ai-terminal-pet\vendor\AI-Terminal-Pet.exe is absent - cross-check skipped'
  }

  if (Test-Path -LiteralPath $stage) { Remove-Item -LiteralPath $stage -Recurse -Force }
  New-Item -ItemType Directory -Path $stage -Force | Out-Null

  $sums = @()
  foreach ($entry in $manifestEntries) {
    $from = Join-Path $RepoRoot $entry.Source
    $to = Join-Path $stage $entry.Target
    $parent = Split-Path -Parent $to
    New-Item -ItemType Directory -Path $parent -Force | Out-Null
    # A binary that was just built can be held for a moment by antivirus
    # scanning it, which surfaces as "the file is in use by another process"
    # on a plain copy. That is not a packaging bug and it is not something
    # the caller can retry by hand in an unattended run, so retry here.
    $copied = $false
    for ($attempt = 1; $attempt -le 10; $attempt++) {
      try {
        Copy-Item -LiteralPath $from -Destination $to -Force -ErrorAction Stop
        $copied = $true
        break
      } catch {
        if ($attempt -eq 10) { throw }
        Start-Sleep -Milliseconds 400
      }
    }
    if (-not $copied) { throw ('could not copy ' + $entry.Source) }
    $sums += ((Get-FileSha256 -Path $to) + '  ' + $entry.Target)
  }

  Set-Content -LiteralPath (Join-Path $stage 'README.install.txt') -Value $installNotes -Encoding ASCII -NoNewline:$false
  Set-Content -LiteralPath (Join-Path $stage 'install.cmd') -Value $rootForwarder -Encoding ASCII -NoNewline:$false
  Set-Content -LiteralPath (Join-Path $stage 'SHA256SUMS.txt') -Value $sums -Encoding ASCII -NoNewline:$false

  if (Test-Path -LiteralPath $archivePath) { Remove-Item -LiteralPath $archivePath -Force }
  # The archive is written entry by entry, not with Compress-Archive or even
  # ZipFile.CreateFromDirectory, for two reasons that both matter to the user:
  #
  #   1. .NET Framework (which is what Windows PowerShell 5.1 loads) writes
  #      entry names with backslashes - "bin\pet-hook.exe". Explorer and 7-Zip
  #      cope; `tar -xf` and `unzip` on a Linux or macOS box create one file
  #      whose name literally contains a backslash. The zip format wants '/'.
  #      Passing explicit names is the only way to get that on both frameworks.
  #   2. A fixed timestamp per entry makes the byte-for-byte output reproducible,
  #      which is the same reason the three builds pass -buildvcs=false and the
  #      same reason npm tarballs are reproducible: two builds of one source
  #      must not differ by the date they ran.
  #
  # $EntryEpoch is that fixed timestamp. It is deliberately not "now".
  $EntryEpoch = [DateTimeOffset]::new(2026, 11, 11, 0, 0, 0, [TimeSpan]::Zero)
  $writeEntries = @()
  foreach ($entry in $manifestEntries) {
    $writeEntries += @{ Name = ($entry.Target -replace '\\', '/'); Path = (Join-Path $stage $entry.Target) }
  }
  foreach ($generated in @('README.install.txt', 'install.cmd', 'SHA256SUMS.txt')) {
    $writeEntries += @{ Name = $generated; Path = (Join-Path $stage $generated) }
  }
  $directoryEntries = @('bin/', 'integrations/', 'integrations/deepseek-harness/',
    'integrations/deepseek-harness/ai-terminal-pet-bridge/',
    'integrations/deepseek-harness/ai-terminal-pet-bridge/locale/', 'reporters/', 'scripts/')

  $stream = [System.IO.File]::Create($archivePath)
  try {
    $archive = [System.IO.Compression.ZipArchive]::new($stream, [System.IO.Compression.ZipArchiveMode]::Create, $true)
    try {
      foreach ($directory in $directoryEntries) {
        $item = $archive.CreateEntry($directory, [System.IO.Compression.CompressionLevel]::NoCompression)
        $item.LastWriteTime = $EntryEpoch
      }
      foreach ($entry in $writeEntries) {
        # LastWriteTime has to be set *before* the entry receives its bytes:
        # once an entry is written, ZipArchive is in create mode and refuses to
        # modify it ("cannot modify in create mode"). So each file is created as
        # an entry, stamped, and then filled from its source stream instead of
        # using the ZipFileExtensions.CreateEntryFromFile convenience call.
        $item = $archive.CreateEntry($entry.Name, [System.IO.Compression.CompressionLevel]::Optimal)
        $item.LastWriteTime = $EntryEpoch
        $target = $item.Open()
        try {
          $source = [System.IO.File]::OpenRead($entry.Path)
          try {
            $source.CopyTo($target)
          } finally {
            $source.Dispose()
          }
        } finally {
          $target.Dispose()
        }
      }
    } finally {
      $archive.Dispose()
    }
  } finally {
    $stream.Dispose()
  }

  # A zip that cannot be read back is not a release asset. Reopening it also
  # proves the archive is not mid-write when the caller prints its hash.
  #
  # The assertion counts *files*, not entries: a directory record per folder is
  # normal, and counting it turns a correct archive into a failing check. What
  # has to hold is that every artifact the installer will look for is inside,
  # under a forward-slash name.
  Add-Type -AssemblyName System.IO.Compression.FileSystem
  $zip = [System.IO.Compression.ZipFile]::OpenRead($archivePath)
  try {
    $entries = @($zip.Entries)
  } finally {
    $zip.Dispose()
  }
  $files = @($entries | Where-Object { -not $_.FullName.EndsWith('/') -and -not $_.FullName.EndsWith('\') })
  $expectedNames = @($manifestEntries | ForEach-Object { ($_.Target -replace '\\', '/') })
  $expectedNames += 'README.install.txt', 'install.cmd', 'SHA256SUMS.txt'
  if ($files.Count -ne $expectedNames.Count) {
    Write-Output ('[fail] package: the archive has ' + $files.Count + ' files, expected ' + $expectedNames.Count)
    foreach ($f in $files) { Write-Output ('       ' + $f.FullName) }
    exit 1
  }
  foreach ($name in $expectedNames) {
    if (-not ($files.FullName -contains $name)) {
      Write-Output ('[fail] package: the archive is missing ' + $name)
      exit 1
    }
  }
  foreach ($f in $files) {
    if ($f.FullName.Contains('\')) {
      Write-Output ('[fail] package: the archive entry ' + $f.FullName + ' uses a backslash separator')
      exit 1
    }
  }

  $hash = Get-FileSha256 -Path $archivePath
  $bytes = (Get-Item -LiteralPath $archivePath).Length
  Write-Output ''
  Write-Output ('package: ok - ' + $archiveName)
  Write-Output ('         ' + $bytes + ' bytes, ' + $files.Count + ' files, ' + ($entries.Count - $files.Count) + ' directory entries')
  Write-Output ('         sha256 ' + $hash)
  Write-Output ('         staged at ' + $stage)
} catch {
  Write-Output ''
  Write-Output ('[fail] package: ' + $_.Exception.Message)
  exit 1
}
