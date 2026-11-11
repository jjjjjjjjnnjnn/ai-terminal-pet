@echo off
rem install.cmd - the wrapper the README tells people to run.
rem
rem Why this file exists: on a stock Windows install the effective PowerShell
rem execution policy is RemoteSigned, and RemoteSigned refuses to run a local
rem .ps1 that carries the "downloaded from the internet" zone marker. Anyone
rem who downloads the release zip and double-clicks the installer therefore
rem gets
rem
rem   scripts\install.ps1 cannot be loaded because running scripts is
rem   disabled on this system
rem
rem before the installer ever runs - a security-policy message in place of an
rem install. cmd.exe does not consult that policy, so this wrapper is the way
rem in. -ExecutionPolicy Bypass applies to the single powershell.exe process
rem started here; it changes no MachinePolicy, UserPolicy or CurrentUser
rem setting, so nothing else on the machine is affected.
rem
rem Every argument is forwarded verbatim, so
rem
rem   install.cmd -VerifyOnly
rem   install.cmd -Uninstall
rem   install.cmd -AddToPath
rem
rem behave exactly like their .ps1 spelling, including the exit code.
rem
rem The .ps1 itself stays the single implementation: this file adds no install
rem logic of its own, and it is deliberately ASCII-only for the same reason
rem install.ps1 is (Windows PowerShell 5.1 reads a BOM-less script in the ANSI
rem code page).

setlocal EnableExtensions

set "_SCRIPT=%~dp0install.ps1"
set "_POWERSHELL=powershell.exe"

if not exist "%_SCRIPT%" (
  echo [fail] install: %_SCRIPT% is missing. Run this wrapper from the scripts directory of a complete checkout.
  exit /b 1
)

"%_POWERSHELL%" -NoProfile -NoLogo -ExecutionPolicy Bypass -File "%_SCRIPT%" %*
set "_STATUS=%ERRORLEVEL%"

rem A double-clicked .cmd owns its console, and that console closes the instant
rem the script ends - which would swallow the last line of the report, including
rem a failure. "No arguments" is the reliable signal for that case: every
rem documented invocation passes at least one flag. `pause` returns immediately
rem when stdin is not a console, so unattended runs are unaffected.
if "%~1"=="" (
  echo.
  echo Press any key to close this window.
  pause > nul
)

endlocal & exit /b %_STATUS%
