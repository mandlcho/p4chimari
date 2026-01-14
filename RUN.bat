@echo off
cd /d "%~dp0"

if not exist p4chimari.exe (
    echo ========================================
    echo   ERROR: p4chimari.exe not found!
    echo ========================================
    echo.
    echo Please run INSTALL.bat first to build the executable.
    echo.
    pause
    exit /b 1
)

p4chimari.exe
