@echo off
setlocal EnableDelayedExpansion

set BIN_NAME=klyntar

echo [INFO] Fetching dependencies ...
go mod download

echo [INFO] Core building process started

REM Build the core
go build -o %BIN_NAME% .

IF %ERRORLEVEL% EQU 0 (
    echo [SUCCESS] Core was successfully built

    for %%I in (%BIN_NAME%) do set BIN_PATH=%%~dpfnI

    echo [INFO] Binary path: !BIN_PATH!

    echo.
    echo To add KLY to PATH permanently, you can run:
    echo setx PATH "%%PATH%%;!BIN_PATH!"
    echo.

    type ..\images\success_build.txt

) ELSE (
    type ..\images\fail_build.txt
)

endlocal
pause
