#!/bin/bash

# SerialWebViewer - Build Script for All Platforms
# This script compiles the project for Windows, Linux, and macOS

VERSION="v1.0"
BUILD_DIR="build"
PROJECT_NAME="SerialWebViewer"

echo "🚀 Building SerialWebViewer ${VERSION}..."
echo ""

# Create build directory
mkdir -p ${BUILD_DIR}

# Build for Windows
echo "📦 Building for Windows (AMD64)..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ${BUILD_DIR}/${PROJECT_NAME}.exe main.go
if [ $? -eq 0 ]; then
    echo "✅ Windows build completed: ${BUILD_DIR}/${PROJECT_NAME}.exe"
else
    echo "❌ Windows build failed"
fi
echo ""

# Build for Linux
echo "📦 Building for Linux (AMD64)..."
go build -ldflags="-s -w" -o ${BUILD_DIR}/serialwebviewer main.go
if [ $? -eq 0 ]; then
    echo "✅ Linux build completed: ${BUILD_DIR}/serialwebviewer"
else
    echo "❌ Linux build failed"
fi
echo ""

# Build for macOS (AMD64)
echo "📦 Building for macOS (AMD64)..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o ${BUILD_DIR}/${PROJECT_NAME}.mac main.go
if [ $? -eq 0 ]; then
    echo "✅ macOS build completed: ${BUILD_DIR}/${PROJECT_NAME}.mac"
else
    echo "❌ macOS build failed"
fi
echo ""

# Build for macOS (ARM64/Apple Silicon)
echo "📦 Building for macOS (ARM64/Apple Silicon)..."
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o ${BUILD_DIR}/${PROJECT_NAME}-arm64.mac main.go
if [ $? -eq 0 ]; then
    echo "✅ macOS ARM64 build completed: ${BUILD_DIR}/${PROJECT_NAME}-arm64.mac"
else
    echo "❌ macOS ARM64 build failed"
fi
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✨ All builds completed!"
echo "📁 Output directory: ${BUILD_DIR}/"
echo ""
echo "Build artifacts:"
ls -lh ${BUILD_DIR}/
echo ""
echo "To create a release, run:"
echo "  gh release create ${VERSION} ${BUILD_DIR}/*"
