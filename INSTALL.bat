@echo off
echo ========================================
echo   P4CHIMARI - Building Executable
echo ========================================
echo.

cd src
go build -o ../bin/p4chimari.exe

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ========================================
    echo   BUILD SUCCESS!
    echo ========================================
    echo.
    echo Executable created in bin/ folder
    echo.
    echo To run: Double-click RUN.bat
    echo.
) else (
    echo.
    echo ========================================
    echo   BUILD FAILED!
    echo ========================================
    echo.
    echo Make sure Go is installed and in PATH
    echo.
)

pause
