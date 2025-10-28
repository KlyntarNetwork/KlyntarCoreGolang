@echo off
setlocal EnableExtensions EnableDelayedExpansion

REM ===========================================
REM   Klyntar build script (native Windows)
REM   - Pretty banners (ANSI, Win10+)
REM   - UTF-8 ASCII art print helper
REM ===========================================

REM ---- ANSI colors (Win10+ with VT support) ----
for /F "delims=" %%A in ('echo prompt $E^| cmd') do set "ESC=%%A"
set "BOLD=%ESC%[1m"
set "RESET=%ESC%[0m"
set "YELLOW_BG=%ESC%[43m"
set "GREEN_BG=%ESC%[42m"
set "RED_BG=%ESC%[41m"

goto :MAIN

REM =========================
REM Helpers (labels)
REM =========================

:banner
REM Usage: call :banner "<bg_color>" "Message..."
set "bg=%~1"
set "msg=%~2"
echo(
echo %bg%%BOLD%%msg%%RESET%
goto :eof

:hr
echo ------------------------------------------------------------
goto :eof

:show_art_utf8
REM Usage: call :show_art_utf8 "path\to\file.txt"
set "ART=%~1"
if not exist "%ART%" goto :eof

REM Remember current active code page (locale independent parsing)
for /f "tokens=2 delims=:" %%C in ('chcp') do set "OLD_CP=%%C"
set "OLD_CP=%OLD_CP: =%"

REM Switch to UTF-8 for proper art rendering
chcp 65001 >nul
type "%ART%"
REM Restore previous code page
if defined OLD_CP chcp %OLD_CP% >nul
goto :eof


REM =========================
REM Main flow
REM =========================
:MAIN
set "BIN_NAME=klyntar.exe"
set "SUCCESS_ART=.\images\success_build.txt"
set "FAIL_ART=.\images\fail_build.txt"

set "TS=%date% %time%"
call :banner "%YELLOW_BG%" "Fetching dependencies  •  %TS%"
call :hr
go mod download || goto FAIL

set "TS=%date% %time%"
call :banner "%GREEN_BG%" "Core building process started  •  %TS%"
call :hr
echo %BOLD%Building the project...%RESET%
REM Native build (no GOOS/GOARCH tweaks)
go build -o "%BIN_NAME%" . || goto FAIL

call :banner "%GREEN_BG%" "Build succeeded"
echo Binary: %BIN_NAME%
echo Path  : %cd%\%BIN_NAME%
if exist "%SUCCESS_ART%" (
  echo(
  call :show_art_utf8 "%SUCCESS_ART%"
)
exit /b 0


REM =========================
REM Fail flow
REM =========================
:FAIL
call :banner "%RED_BG%" "Build failed"
if exist "%FAIL_ART%" (
  echo(
  call :show_art_utf8 "%FAIL_ART%"
)
exit /b 1
