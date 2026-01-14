@echo off
echo ========================================
echo   P4CHIMARI - Building Executable
echo ========================================
echo.

cd src
go build -o ../p4chimari.exe

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ========================================
    echo   BUILD SUCCESS!
    echo ========================================
    echo.
    echo Executable created: p4chimari.exe
    echo.
    echo To run: Double-click p4chimari.exe
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
