@echo off
REM SerialWebViewer - Build Script for All Platforms (Windows)
REM This script compiles the project for Windows, Linux, and macOS

set VERSION=v1.0
set BUILD_DIR=build
set PROJECT_NAME=SerialWebViewer

echo 🚀 Building SerialWebViewer %VERSION%...
echo.

REM Create build directory
if not exist "%BUILD_DIR%" mkdir "%BUILD_DIR%"

REM Build for Windows (AMD64)
echo 📦 Building for Windows (AMD64)...
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o %BUILD_DIR%\%PROJECT_NAME%.exe main.go
if %ERRORLEVEL% EQU 0 (
    echo ✅ Windows build completed: %BUILD_DIR%\%PROJECT_NAME%.exe
) else (
    echo ❌ Windows build failed
)
echo.

REM Build for Linux (AMD64)
echo 📦 Building for Linux (AMD64)...
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w" -o %BUILD_DIR%\serialwebviewer main.go
if %ERRORLEVEL% EQU 0 (
    echo ✅ Linux build completed: %BUILD_DIR%\serialwebviewer
) else (
    echo ❌ Linux build failed
)
echo.

REM Build for macOS (AMD64)
echo 📦 Building for macOS (AMD64)...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags="-s -w" -o %BUILD_DIR%\%PROJECT_NAME%.mac main.go
if %ERRORLEVEL% EQU 0 (
    echo ✅ macOS build completed: %BUILD_DIR%\%PROJECT_NAME%.mac
) else (
    echo ❌ macOS build failed
)
echo.

REM Build for macOS (ARM64/Apple Silicon)
echo 📦 Building for macOS (ARM64/Apple Silicon)...
set GOOS=darwin
set GOARCH=arm64
go build -ldflags="-s -w" -o %BUILD_DIR%\%PROJECT_NAME%-arm64.mac main.go
if %ERRORLEVEL% EQU 0 (
    echo ✅ macOS ARM64 build completed: %BUILD_DIR%\%PROJECT_NAME%-arm64.mac
) else (
    echo ❌ macOS ARM64 build failed
)
echo.

echo ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
echo ✨ All builds completed!
echo 📁 Output directory: %BUILD_DIR%\
echo.
echo Build artifacts:
dir %BUILD_DIR%
echo.
echo To create a release, use GitHub CLI or upload manually.
pause
