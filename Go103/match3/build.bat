@echo off
REM Build script for Crystal Cascade - Windows

echo Building Crystal Cascade for Windows...

go build -o crystal-cascade.exe .\cmd

if %errorlevel% equ 0 (
    echo Build successful! Run with: crystal-cascade.exe
) else (
    echo Build failed!
    exit /b 1
)
