# install.ps1 - install, upgrade, verify or uninstall the AI-Terminal-Pet overlay.
#
#   powershell -NoProfile -File install.ps1                 # install / upgrade
#   powershell -NoProfile -File install.ps1 -VerifyOnly     # read-only self-check
#   powershell -NoProfile -File install.ps1 -Uninstall      # confirm, then delete
#   powershell -NoProfile -File install.ps1 -Uninstall -Force
#   powershell -NoProfile -File install.ps1 -AddToPath      # install + user PATH entry
# -Uninstall also removes the HKCU Path entry that -AddToPath added.
#
# Exit codes: 0 success, 1 install or verification failed, 2 bad usage,
# 3 the user cancelled. Every message is plain ASCII on purpose: Windows
# PowerShell 5.1 reads a script without a byte-order mark in the ANSI code page,
# so non-ASCII text here would be mojibake on half the machines that run it.
param(
  [string]$Source = (Join-Path $PSScriptRoot 'bin\AI-Terminal-Pet.exe'),
  [string]$Destination = (Join-Path $env:LOCALAPPDATA 'AI-Terminal-Pet\AI-Terminal-Pet.exe'),
  [string]$ShimSource = (Join-Path $PSScriptRoot 'bin\pet-statusline.exe'),
  [string]$ShimDestination = (Join-Path $env:LOCALAPPDATA 'AI-Terminal-Pet\pet-statusline.exe'),
  [string]$ProtocolSource = (Join-Path $PSScriptRoot 'reporters\protocol.js'),
  [string]$ProtocolDestination = (Join-Path $env:LOCALAPPDATA 'AI-Terminal-Pet\protocol.js'),
  [string]$BundleSource = (Join-Path $PSScriptRoot 'integrations\deepseek-harness\ai-terminal-pet-bridge'),
  [string]$BundleDestination = (Join-Path $env:LOCALAPPDATA 'AI-Terminal-Pet\bundles\ai-terminal-pet-bridge'),
  [switch]$VerifyOnly,
  [switch]$Uninstall,
  [switch]$AddToPath,
  [switch]$Force
)

$ErrorActionPreference = 'Stop'
$NewLine = [Environment]::NewLine
$RunKeyPath = 'Software\Microsoft\Windows\CurrentVersion\Run'
$RunValueName = 'AITerminalPet'

# Everything this script creates, verifies or deletes lives under exactly one
# directory. Deriving it from -Destination (instead of hard-coding
# %LOCALAPPDATA%) is what lets the self-test rehearse a full install, verify and
# uninstall against a scratch directory without touching the real one.
$InstallDirectory = Split-Path -Parent $Destination
if ($InstallDirectory) { $InstallDirectory = [System.IO.Path]::GetFullPath($InstallDirectory) }

# -Destination is the documented escape hatch for rehearsing against a scratch
# directory, so the three sibling destinations must follow it. Without this an
# explicit -Destination would copy one exe into the scratch directory and the
# other nine files into the live install - which is worse than not offering the
# override at all.
if ($PSBoundParameters.ContainsKey('Destination')) {
  if (-not $PSBoundParameters.ContainsKey('ShimDestination')) {
    $ShimDestination = Join-Path $InstallDirectory 'pet-statusline.exe'
  }
  if (-not $PSBoundParameters.ContainsKey('ProtocolDestination')) {
    $ProtocolDestination = Join-Path $InstallDirectory 'protocol.js'
  }
  if (-not $PSBoundParameters.ContainsKey('BundleDestination')) {
    $BundleDestination = Join-Path $InstallDirectory 'bundles\ai-terminal-pet-bridge'
  }
}

# A read-only check and a destructive delete cannot mean the same invocation,
# and -AddToPath is a write to HKCU. Reject the combinations up front with a
# message that says which flag to drop, instead of half-running one of them.
if ($VerifyOnly -and $Uninstall) {
  Write-Output '[fail] -VerifyOnly and -Uninstall are mutually exclusive: the first only reads, the second deletes the install directory.'
  Write-Output '       Run them as two separate commands.'
  exit 2
}
if ($AddToPath -and ($VerifyOnly -or $Uninstall)) {
  Write-Output '[fail] -AddToPath writes to the user PATH and cannot be combined with -VerifyOnly or -Uninstall.'
  exit 2
}

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

# The install is five classes of files. Keeping them in one list means the
# installer, the verifier and the uninstaller can never disagree about what
# "installed" means - and a verifier that checks a different set than the
# installer writes is exactly how a stale protocol.js survived an upgrade in
# the past and disconnected every reporter.
function Get-InstallArtifacts {
  $bundleFiles = @(
    'package.json', 'cordis.patch.yml', 'index.js', 'icon.svg', 'README.md',
    'locale\en.json', 'locale\zh.json'
  )
  $items = @(
    @{ Group = 'overlay'; Label = 'AI-Terminal-Pet.exe'; From = $Source; To = $Destination; Binary = $true },
    @{ Group = 'statusline'; Label = 'pet-statusline.exe'; From = $ShimSource; To = $ShimDestination; Binary = $true },
    @{ Group = 'protocol'; Label = 'protocol.js'; From = $ProtocolSource; To = $ProtocolDestination; Binary = $false }
  )
  foreach ($relative in $bundleFiles) {
    if ($relative -eq 'package.json' -or $relative -eq 'cordis.patch.yml') {
      $group = 'bundle manifest'
    } else {
      $group = 'bundle runtime'
    }
    $items += @{
      Group = $group
      Label = 'bundles\ai-terminal-pet-bridge\' + $relative
      From = (Join-Path $BundleSource $relative)
      To = (Join-Path $BundleDestination $relative)
      Binary = $false
    }
  }
  return ,$items
}

# Get-Sha256OrNull answers "what does this file hash to" without throwing for a
# file that is simply absent, because "absent" is a normal, reportable state in
# -VerifyOnly and must not abort the rest of the report.
function Get-Sha256OrNull {
  param([Parameter(Mandatory)][string]$Path)
  if (-not (Test-Path -LiteralPath $Path -PathType Leaf)) { return $null }
  return (Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash
}

# Test-PathInsideDirectory is the guard that keeps -Uninstall from acting on a
# path it was not asked about. It is used for the HKCU Run entry: deleting an
# entry that points at a *different* copy of the overlay would silently disable
# somebody else's install, so the entry is only removed when it names a file
# under the directory being deleted.
function Test-PathInsideDirectory {
  param([string]$Path, [string]$Directory)
  if (-not $Path) { return $false }
  try {
    $fullDirectory = [System.IO.Path]::GetFullPath($Directory).TrimEnd('\')
    $fullPath = [System.IO.Path]::GetFullPath($Path)
  } catch {
    return $false
  }
  return $fullPath.StartsWith($fullDirectory + '\', [System.StringComparison]::OrdinalIgnoreCase)
}

# Assert-SafeInstallDirectory refuses the two ways a typo turns -Uninstall into
# a data-loss story: a drive root (or any path with only one component), and a
# directory whose name is not the one this project owns.
function Assert-SafeInstallDirectory {
  $leaf = Split-Path -Leaf $InstallDirectory
  $parent = Split-Path -Parent $InstallDirectory
  # Split-Path -Leaf of "C:" returns "C:" and -Parent returns "", so an empty
  # parent (not just an empty leaf) is the shape that means "drive root".
  if (-not $parent -or -not $leaf -or ($leaf -match '^[A-Za-z]:$')) {
    throw ('refusing to touch "' + $InstallDirectory + '": that is a drive root, not an install directory')
  }
  # -Force means "do not ask me", it does NOT mean "delete anything I point at".
  # The name check therefore survives -Force: an uninstaller whose only safety
  # rail is a prompt is one mistyped -Destination away from deleting a
  # directory that was never ours.
  if ($leaf -ne 'AI-Terminal-Pet') {
    throw ('refusing to delete "' + $InstallDirectory + '": its directory name is not AI-Terminal-Pet. Point -Destination at <somewhere>\AI-Terminal-Pet\AI-Terminal-Pet.exe.')
  }
  if ($parent -eq $InstallDirectory) {
    throw ('refusing to touch "' + $InstallDirectory + '": it has no parent directory')
  }
}

# -VerifyOnly never writes a file. It re-reads both sides of every artifact and
# reports [ok] / [warn] / [fail] per item, so a user can prove an install is
# intact instead of taking the installer's word for it.
function Invoke-VerifyOnly {
  $items = Get-InstallArtifacts
  $groupOrder = @('overlay', 'statusline', 'protocol', 'bundle manifest', 'bundle runtime')
  $stats = @{}
  foreach ($group in $groupOrder) { $stats[$group] = @{ Files = 0; Failed = 0; Warned = 0 } }
  $failures = 0
  $warnings = 0

  Write-Output ('verify: ' + $items.Count + ' artifacts against ' + $InstallDirectory)
  foreach ($item in $items) {
    $group = $item.Group
    if (-not $stats.ContainsKey($group)) {
      $stats[$group] = @{ Files = 0; Failed = 0; Warned = 0 }
    }
    $stats[$group].Files = $stats[$group].Files + 1
    # An empty panel shows an empty cell, not the word "null", and PadRight on
    # an empty string still reserves the column.
    $lean = $item.Label -replace 'bundles\\ai-terminal-pet-bridge\\', 'bundle:'
    $want = Get-Sha256OrNull -Path $item.From
    $have = Get-Sha256OrNull -Path $item.To
    $tag = 'ok'
    if ($null -eq $have) {
      if ($null -eq $want) {
        $tag = 'warn'
        $note = 'missing here and in the repository - nothing to compare'
        $warnings = $warnings + 1
        $stats[$group].Warned = $stats[$group].Warned + 1
      } else {
        $tag = 'fail'
        $note = 'not installed'
        $failures = $failures + 1
        $stats[$group].Failed = $stats[$group].Failed + 1
      }
    } elseif ($null -eq $want) {
      $tag = 'warn'
      $note = 'repository copy is missing - cannot compare'
      $warnings = $warnings + 1
      $stats[$group].Warned = $stats[$group].Warned + 1
    } elseif ($want -ne $have) {
      $tag = 'fail'
      $note = 'sha256 ' + $have + ' != repository ' + $want
      $failures = $failures + 1
      $stats[$group].Failed = $stats[$group].Failed + 1
    } else {
      $note = 'sha256 matches the repository'
    }
    $name = '[' + $tag + '] ' + $group + '/' + $lean
    # Two extra spaces keep the columns readable when the label is longer than
    # the pad width (PadRight never truncates). Labels are ASCII, so the pad
    # width and the printed width are the same.
    Write-Output ($name.PadRight(62) + $note)
  }

  Write-Output ''
  foreach ($group in $groupOrder) {
    if (-not $stats.ContainsKey($group)) { continue }
    $entry = $stats[$group]
    if ($entry.Files -eq 0) { continue }
    Write-Output ('  ' + $group.PadRight(17) + $entry.Files.ToString() + ' file(s), ' + $entry.Failed.ToString() + ' failed, ' + $entry.Warned.ToString() + ' cannot be compared')
  }
  Write-Output ''
  if ($failures -gt 0) {
    Write-Output ('install: not verified - ' + $failures + ' failure(s), ' + $warnings + ' warning(s)')
    # Do not echo $PSCommandPath: on a Chinese-locale console PowerShell 5.1
    # renders this directory name as mojibake, which makes a copy-pasteable
    # command look broken. Re-running from the same directory needs no path.
    Write-Output '         Rerun install.ps1 (without -VerifyOnly) from this directory to repair.'
    exit 1
  }
  Write-Output ('install: verified - ' + $items.Count + ' file(s) match the repository, ' + $warnings + ' warning(s)')
  exit 0
}

# Stop-InstalledProcesses uses the same 20-attempt reasoning as
# Copy-InstalledBinary, but here the goal is the opposite: leave no process
# holding the directory so Remove-Item can actually succeed. Reporters relaunch
# the overlay, so a single Stop-Process can lose the race.
function Stop-InstalledProcesses {
  $names = @('AI-Terminal-Pet', 'pet-statusline')
  for ($attempt = 1; $attempt -le 20; $attempt++) {
    $running = Get-Process -Name $names -ErrorAction SilentlyContinue
    if (-not $running) { return $true }
    $running | Stop-Process -Force -ErrorAction SilentlyContinue
    $running | Wait-Process -Timeout 10 -ErrorAction SilentlyContinue
    Start-Sleep -Milliseconds 500
  }
  return $false
}

function Invoke-Uninstall {
  Assert-SafeInstallDirectory
  $exists = Test-Path -LiteralPath $InstallDirectory -PathType Container
  if (-not $exists) {
    Write-Output ('[warn] uninstall: ' + $InstallDirectory + ' does not exist - only the Run entry is checked')
  }
  if (-not $Force) {
    Write-Output ('[warn] uninstall: this deletes ' + $InstallDirectory + ' and every file under it.')
    Write-Output '[warn] uninstall: the overlay and pet-statusline are stopped first; no other directory is touched.'
    $answer = ''
    try {
      $answer = Read-Host 'Type yes to continue'
    } catch {
      Write-Output '[fail] uninstall: this needs confirmation but no console is attached; pass -Force to run unattended'
      exit 2
    }
    if ($answer -ne 'yes') {
      Write-Output 'uninstall: cancelled by the user - nothing was deleted'
      exit 3
    }
  }

  [void](Stop-InstalledProcesses)
  if ($exists) {
    $removed = $false
    for ($attempt = 1; $attempt -le 5; $attempt++) {
      try {
        Remove-Item -LiteralPath $InstallDirectory -Recurse -Force -ErrorAction Stop
        $removed = $true
        break
      } catch {
        if ($attempt -eq 5) {
          Write-Output ('[fail] uninstall: could not delete ' + $InstallDirectory)
          Write-Output ('       ' + $_.Exception.Message)
          Write-Output '       A reporter most likely relaunched the overlay mid-delete; close every host CLI hook and rerun.'
          exit 1
        }
        Start-Sleep -Milliseconds 700
      }
    }
    if (-not $removed -or (Test-Path -LiteralPath $InstallDirectory)) {
      Write-Output ('[fail] uninstall: ' + $InstallDirectory + ' is still there')
      exit 1
    }
    Write-Output ('[ok] uninstall: removed ' + $InstallDirectory)
  }

  # The Run entry is only removed when it points into the directory that was
  # just deleted. An entry aimed at a different copy belongs to that copy.
  $entry = $null
  $runKey = [Microsoft.Win32.Registry]::CurrentUser.OpenSubKey($RunKeyPath, $false)
  if ($null -ne $runKey) {
    $entry = $runKey.GetValue($RunValueName)
    $runKey.Close()
  }
  if ($null -eq $entry) {
    Write-Output ('[ok] uninstall: no HKCU Run entry named ' + $RunValueName)
  } else {
    $target = ([string]$entry).Trim().Trim('"')
    if (Test-PathInsideDirectory -Path $target -Directory $InstallDirectory) {
      try {
        $writable = [Microsoft.Win32.Registry]::CurrentUser.OpenSubKey($RunKeyPath, $true)
        $writable.DeleteValue($RunValueName, $false)
        $writable.Close()
        Write-Output ('[ok] uninstall: removed HKCU Run entry ' + $RunValueName + ' (' + $target + ')')
      } catch {
        Write-Output ('[fail] uninstall: could not remove the HKCU Run entry: ' + $_.Exception.Message)
        exit 1
      }
    } else {
      Write-Output ('[warn] uninstall: HKCU Run entry ' + $RunValueName + ' points at ' + $target)
      Write-Output '       That is not inside the directory just deleted, so it was left alone.'
      Write-Output ('       Remove it by hand if it is stale: Remove-ItemProperty -Path HKCU:\Software\Microsoft\Windows\CurrentVersion\Run -Name ' + $RunValueName)
    }
  }
  # A PATH entry that points at the directory just deleted is a dead entry, so
  # it goes too. A PATH cleanup failure is reported as a warning rather than a
  # failure: the files are already gone, and the remaining entry is cosmetic.
  try {
    $pathResult = Remove-InstallDirectoryFromUserPath
    Write-Output ('[ok] uninstall: PATH: ' + $pathResult.Note)
  } catch {
    Write-Output ('[warn] uninstall: PATH: could not read or rewrite the user PATH: ' + $_.Exception.Message)
    Write-Output '       Remove the stale entry by hand if it is still there: open HKCU:\Environment and edit Path.'
  }

  Write-Output 'uninstall: done'
  return
}

# Remove-InstallDirectoryFromUserPath is the uninstall counterpart of -AddToPath.
# Leaving a deleted directory on PATH is not harmless: every process the user
# starts afterwards inherits a dead entry, and 'where AI-Terminal-Pet.exe'
# keeps reporting a path that no longer exists. Same safety rules as the add:
# HKCU only, exact path match only, and the previous value is written to a
# backup file before the edit (into %TEMP%, because the install directory is
# about to be deleted).
function Remove-InstallDirectoryFromUserPath {
  $environmentKey = [Microsoft.Win32.Registry]::CurrentUser.OpenSubKey('Environment', $true)
  if ($null -eq $environmentKey) {
    return @{ Changed = $false; Note = 'HKCU\Environment has no Path value, so there is no entry to remove' }
  }
  try {
    $kind = [Microsoft.Win32.RegistryValueKind]::String
    $current = ''
    try {
      $kind = $environmentKey.GetValueKind('Path')
      $current = [string]$environmentKey.GetValue('Path', '', [Microsoft.Win32.RegistryValueOptions]::DoNotExpandEnvironmentNames)
    } catch {
      return @{ Changed = $false; Note = 'HKCU\Environment has no Path value, so there is no entry to remove' }
    }
    $candidate = $InstallDirectory.TrimEnd('\')
    $entries = @()
    $matched = @()
    foreach ($part in ($current -split ';')) {
      $trimmed = $part.Trim()
      if (-not $trimmed) { continue }
      if ([string]::Equals($trimmed.TrimEnd('\'), $candidate, [System.StringComparison]::OrdinalIgnoreCase)) {
        $matched = $matched + $trimmed
      } else {
        $entries = $entries + $trimmed
      }
    }
    if ($matched.Count -eq 0) {
      return @{ Changed = $false; Note = 'the install directory is not on the user PATH - nothing to remove' }
    }
    $backupPath = Join-Path $env:TEMP ('ai-terminal-pet-path-backup-' + (Get-Date).ToString('yyyyMMdd-HHmmss') + '.txt')
    $backupText = '# ai-terminal-pet install.ps1 -Uninstall PATH backup' + $NewLine + '# taken ' + (Get-Date).ToString('o') + $NewLine + '# HKCU\Environment Path, registry kind ' + $kind.ToString() + $NewLine + $current + $NewLine
    [System.IO.File]::WriteAllText($backupPath, $backupText, (New-Object System.Text.UTF8Encoding($false)))
    $environmentKey.SetValue('Path', (@($entries) -join ';'), $kind)
    return @{ Changed = $true; Note = ('removed ' + $matched.Count + ' user PATH entr(y|ies) for ' + $candidate + '; previous value saved to ' + $backupPath); Backup = $backupPath }
  } finally {
    $environmentKey.Close()
  }
}
# Add-InstallDirectoryToUserPath appends the install directory to HKCU
# Environment Path only. HKLM is never touched: the overlay is a per-user tool,
# and a machine-wide PATH edit needs administrator rights and changes every
# account on the box. The original value is backed up before the write because
# PATH is the one setting where a bad edit breaks every tool the user has.
function Add-InstallDirectoryToUserPath {
  $backupPath = Join-Path $InstallDirectory 'path-backup.txt'
  $environmentKey = [Microsoft.Win32.Registry]::CurrentUser.OpenSubKey('Environment', $true)
  if ($null -eq $environmentKey) {
    $environmentKey = [Microsoft.Win32.Registry]::CurrentUser.CreateSubKey('Environment')
  }
  if ($null -eq $environmentKey) {
    throw 'cannot open HKCU\Environment for writing'
  }
  try {
    $kind = [Microsoft.Win32.RegistryValueKind]::String
    $current = ''
    try {
      $kind = $environmentKey.GetValueKind('Path')
      $current = [string]$environmentKey.GetValue('Path', '', [Microsoft.Win32.RegistryValueOptions]::DoNotExpandEnvironmentNames)
    } catch {
      $current = ''
      $kind = [Microsoft.Win32.RegistryValueKind]::String
    }
    $entries = @()
    foreach ($part in ($current -split ';')) {
      $trimmed = $part.Trim()
      if ($trimmed) { $entries = $entries + $trimmed }
    }
    $candidate = $InstallDirectory.TrimEnd('\')
    $found = $false
    foreach ($entry in $entries) {
      if ([string]::Equals($entry.TrimEnd('\'), $candidate, [System.StringComparison]::OrdinalIgnoreCase)) {
        $found = $true
        break
      }
    }
    if ($found) {
      return @{ Changed = $false; Note = ('the install directory is already on the user PATH (' + $candidate + ')'); Backup = $null }
    }
    $backupText = '# ai-terminal-pet install.ps1 -AddToPath backup' + $NewLine + '# taken ' + (Get-Date).ToString('o') + $NewLine + '# HKCU\Environment Path, registry kind ' + $kind.ToString() + $NewLine + $current + $NewLine
    New-Item -ItemType Directory -Path (Split-Path -Parent $backupPath) -Force | Out-Null
    [System.IO.File]::WriteAllText($backupPath, $backupText, (New-Object System.Text.UTF8Encoding($false)))
    $newValue = (@($entries + $candidate) -join ';')
    $environmentKey.SetValue('Path', $newValue, $kind)
    return @{ Changed = $true; Note = ('appended ' + $candidate + ' to the user PATH (' + $kind.ToString() + '); previous value saved to ' + $backupPath); Backup = $backupPath }
  } finally {
    $environmentKey.Close()
  }
}

if ($Uninstall) {
  Invoke-Uninstall
  exit 0
}

if ($VerifyOnly) {
  Invoke-VerifyOnly
}

$Step = 'check the repository sources'
$Items = @()
try {
  # Preflight names every missing file and the command that produces it. The
  # old script said only "Build output not found", which is useless to somebody
  # who cloned the repository and has never built it.
  $Items = Get-InstallArtifacts
  $missingBins = @()
  $missingOther = @()
  $binRoot = Join-Path $PSScriptRoot 'bin'
  foreach ($item in $Items) {
    if (Test-Path -LiteralPath $item.From -PathType Leaf) { continue }
    # A caller-supplied -Source outside bin is still a binary; the label on the
    # item knows that, while the path no longer does.
    if ($item.Binary -or $item.From.StartsWith($binRoot, [System.StringComparison]::OrdinalIgnoreCase)) {
      $missingBins = $missingBins + $item.From
    } else {
      $missingOther = $missingOther + $item.From
    }
  }
  if ($missingBins.Count -gt 0 -or $missingOther.Count -gt 0) {
    $message = 'Required repository files are missing:' + $NewLine
    foreach ($missing in $missingBins) { $message = $message + '  ' + $missing + $NewLine }
    foreach ($missing in $missingOther) { $message = $message + '  ' + $missing + $NewLine }
    if ($missingBins.Count -gt 0) {
      $message = $message + $NewLine + 'Build the binaries first, from the repository root (copy-paste):' + $NewLine
      $message = $message + '  go build -trimpath -buildvcs=false -ldflags "-s -w" -o .\bin\AI-Terminal-Pet.exe .\cmd\ai-terminal-pet' + $NewLine
      $message = $message + '  go build -trimpath -buildvcs=false -ldflags "-s -w" -o .\bin\pet-statusline.exe .\cmd\pet-statusline' + $NewLine
    }
    if ($missingOther.Count -gt 0) {
      $message = $message + $NewLine + 'The reporter protocol or the Harness bundle is missing from the checkout; restore it with:' + $NewLine
      $message = $message + '  git checkout -- reporters/protocol.js integrations/deepseek-harness' + $NewLine
    }
    throw $message
  }

  # Hash every source before copying anything. Doing it first is what makes the
  # post-copy check meaningful: if the repository file is already unreadable the
  # run must stop before the installed copy has been damaged.
  $Step = 'hash the repository sources'
  $sourceHashes = @{}
  foreach ($item in $Items) {
    $sourceHashes[$item.To] = (Get-FileHash -LiteralPath $item.From -Algorithm SHA256).Hash
  }

  $Step = 'copy the artifacts into the install directory'
  foreach ($item in $Items) {
    Copy-InstalledBinary -From $item.From -To $item.To
  }

  # Verify after copying instead of trusting Copy-Item. This is the same check
  # the old script had, but a mismatch now reports which files disagree rather
  # than dying on the first one with no state description.
  $Step = 'verify the copied artifacts'
  $mismatched = @()
  foreach ($item in $Items) {
    $installedHash = (Get-FileHash -LiteralPath $item.To -Algorithm SHA256).Hash
    if ($installedHash -ne $sourceHashes[$item.To]) {
      $mismatched = $mismatched + ('  ' + $item.To + $NewLine + '    installed  ' + $installedHash + $NewLine + '    repository ' + $sourceHashes[$item.To])
    }
  }
  if ($mismatched.Count -gt 0) {
    throw ('Installed hash does not match the repository for:' + $NewLine + ($mismatched -join $NewLine))
  }
} catch {
  Write-Output ''
  Write-Output ('[fail] install: failed while trying to ' + $Step)
  Write-Output ('       ' + $_.Exception.Message)
  if ($Step -eq 'check the repository sources' -or $Step -eq 'hash the repository sources') {
    # Nothing has been written yet at these two steps, so saying "half installed"
    # here would be a lie and would send the user chasing a directory that is
    # still exactly as it was.
    Write-Output '       Nothing was copied yet: the install directory is still exactly as it was.'
  } else {
    Write-Output '       The install directory is in an UNKNOWN, possibly half-installed state.'
    Write-Output '       Nothing was rolled back on purpose: the previous binaries may already be'
    Write-Output '       overwritten, so an automatic rollback would be a lie. Re-running install.ps1'
    Write-Output '       is the only way back to a known state.'
  }
  Write-Output '       Rerun install.ps1 from this directory after fixing the sources above.'
  exit 1
}

# PATH is updated only after the files are in place, and a PATH failure never
# pretends the files failed: the binaries are installed either way.
$PathFailed = $false
if ($AddToPath) {
  try {
    $pathResult = Add-InstallDirectoryToUserPath
    Write-Output ('[ok] PATH: ' + $pathResult.Note)
    if ($pathResult.Changed) {
      Write-Output '[warn] PATH: open a new terminal (or sign out and back in) before the install directory is visible to "where AI-Terminal-Pet.exe"'
    }
  } catch {
    $PathFailed = $true
    Write-Output ('[fail] PATH: could not update the user PATH: ' + $_.Exception.Message)
    Write-Output '       The binaries, the protocol and the bundle ARE installed; only the PATH entry is missing.'
  }
}

Write-Output ''
Write-Output ('install: ok - ' + $Items.Count + ' file(s) copied and hash-verified into ' + $InstallDirectory)

Get-Item -LiteralPath $Destination, $ShimDestination, $ProtocolDestination | Select-Object FullName, Length, LastWriteTime
Get-Item -LiteralPath $BundleDestination | Select-Object FullName, LastWriteTime

if ($PathFailed) { exit 1 }
