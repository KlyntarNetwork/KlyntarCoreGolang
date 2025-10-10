@echo off
setlocal EnableExtensions EnableDelayedExpansion

REM ---------------------------
REM ANSI colors (cmd.exe)
REM ---------------------------
REM Try to get ESC. Works on Win10+ terminals that support VT sequences.
for /F "delims=" %%A in ('echo prompt $E^| cmd') do set "ESC=%%A"

REM Allow disabling colors with NO_COLOR env
if defined NO_COLOR (
  set "BOLD="
  set "RESET="
  set "YELLOW_BG="
  set "GREEN_BG="
  set "RED_BG="
) else (
  set "BOLD=%ESC%[1m"
  set "RESET=%ESC%[0m"
  set "YELLOW_BG=%ESC%[43m"
  set "GREEN_BG=%ESC%[42m"
  set "RED_BG=%ESC%[41m"
)

REM ---------------------------
REM Helpers
REM ---------------------------
:banner
REM Usage: call :banner "<bg>" "Message..."
set "bg=%~1"
set "msg=%~2"
echo(
echo %bg%%BOLD%%msg%%RESET%
goto :eof

:say
REM Usage: call :say any text...
echo %*
goto :eof

:hr
echo ------------------------------------------------------------
goto :eof

:ts
REM Timestamp using locale date/time
set "TS=%date% %time%"
goto :eof

REM ---------------------------
REM Config
REM ---------------------------
set "BIN_NAME=klyntar"
set "SUCCESS_ART=.\images\success_build.txt"
set "FAIL_ART=.\images\fail_build.txt"

REM Resolve TARGET_OS and TARGET_ARCH from args or go env
set "TARGET_OS=%~1"
set "TARGET_ARCH=%~2"

if "%TARGET_OS%"=="" (
  for /f "delims=" %%a in ('go env GOOS') do set "TARGET_OS=%%a"
)
if "%TARGET_ARCH%"=="" (
  for /f "delims=" %%a in ('go env GOARCH') do set "TARGET_ARCH=%%a"
)

REM Output name (add .exe if targeting Windows)
set "OUT_NAME=%BIN_NAME%"
if /I "%TARGET_OS%"=="windows" set "OUT_NAME=%BIN_NAME%.exe"

REM ---------------------------
REM Error flow label
REM ---------------------------
:guard
REM Start timer in ticks (100ns units)
for /f "delims=" %%a in ('powershell -NoProfile -Command "[DateTime]::UtcNow.Ticks"') do set "START_TICKS=%%a"

call :ts
call :banner "%YELLOW_BG%" "Fetching dependencies  •  %TS%"
call :hr
for /f "tokens=2,3*" %%v in ('go version') do set "GOVER=%%v %%w"
call :say Working dir  : %cd%
call :say Go version   : %GOVER%
call :say Target       : %TARGET_OS%/%TARGET_ARCH%
call :hr

REM ---------------------------
REM go mod download
REM ---------------------------
go mod download
if errorlevel 1 goto fail

REM ---------------------------
REM Build
REM ---------------------------
call :ts
call :banner "%GREEN_BG%" "Core building process started  •  %TS%"
call :say %BOLD%Building the project for %TARGET_OS%/%TARGET_ARCH%...%RESET%
call :hr

set "GOOS=%TARGET_OS%"
set "GOARCH=%TARGET_ARCH%"
REM Ensure GOOS/GOARCH only for this call
cmd /c set GOOS=%GOOS%^& set GOARCH=%GOARCH%^& go build -o "%OUT_NAME%" .
if errorlevel 1 goto fail

REM ---------------------------
REM Success
REM ---------------------------
for /f "delims=" %%a in ('powershell -NoProfile -Command "$elapsed = ([DateTime]::UtcNow.Ticks - %START_TICKS%) / 10000000; [Math]::Round($elapsed,2)"') do set "ELAPSED=%%a"

call :ts
call :banner "%GREEN_BG%" "Build succeeded  •  %TS%"
call :say Binary        : %OUT_NAME%
call :say Target        : %TARGET_OS%/%TARGET_ARCH%
call :say Output path   : %cd%\%OUT_NAME%
call :say Elapsed time  : %ELAPSED%s
call :hr

if exist "%SUCCESS_ART%" (
  echo(
  type "%SUCCESS_ART%"
  echo(
)

exit /b 0

REM ---------------------------
REM Fail
REM ---------------------------
:fail
call :ts
call :banner "%RED_BG%" "Build failed  •  %TS%"
if exist "%FAIL_ART%" (
  echo(
  type "%FAIL_ART%"
  echo(
)
exit /b 1
