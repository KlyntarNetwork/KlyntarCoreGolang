@echo off
setlocal EnableExtensions EnableDelayedExpansion

REM ===========================================
REM   Klyntar build script (native Windows)
REM ===========================================

REM ANSI colors (Win10+ with VT support)
for /F "delims=" %%A in ('echo prompt $E^| cmd') do set "ESC=%%A"
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
REM Config (native build)
REM ---------------------------
set "BIN_NAME=klyntar"
set "OUT_NAME=%BIN_NAME%.exe"
set "SUCCESS_ART=.\images\success_build.txt"
set "FAIL_ART=.\images\fail_build.txt"

REM Force native build (clear cross-compile env if set from outside)
set "GOOS="
set "GOARCH="
set "CGO_ENABLED="

REM Resolve host target (info only)
for /f "delims=" %%a in ('go env GOOS') do set "HOST_OS=%%a"
for /f "delims=" %%a in ('go env GOARCH') do set "HOST_ARCH=%%a"

REM ---------------------------
REM Timer start
REM ---------------------------
for /f "delims=" %%a in ('powershell -NoProfile -Command "[DateTime]::UtcNow.Ticks"') do set "START_TICKS=%%a"

call :ts
call :banner "%YELLOW_BG%" "Fetching dependencies  •  %TS%"
call :hr
for /f "tokens=3,4" %%v in ('go version') do set "GOVER=%%v %%w"
call :say Working dir  : %cd%
call :say Go version   : %GOVER%
call :say Target       : %HOST_OS%/%HOST_ARCH% (native)
call :hr

go mod download
if errorlevel 1 goto fail

call :ts
call :banner "%GREEN_BG%" "Core building process started  •  %TS%"
call :say %BOLD%Building the project (native %HOST_OS%/%HOST_ARCH%)...%RESET%
call :hr

REM --- Native build ---
go build -o "%OUT_NAME%" .
if errorlevel 1 goto fail

for /f "delims=" %%a in ('powershell -NoProfile -Command "$elapsed = ([DateTime]::UtcNow.Ticks - %START_TICKS%) / 10000000; [Math]::Round($elapsed,2)"') do set "ELAPSED=%%a"

call :ts
call :banner "%GREEN_BG%" "Build succeeded  •  %TS%"
call :say Binary       : %OUT_NAME%
call :say Target       : %HOST_OS%/%HOST_ARCH%
call :say Output path  : %cd%\%OUT_NAME%
call :say Elapsed time : %ELAPSED%s
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
