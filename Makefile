.PHONY: all clean windows linux mac mac-arm64 help

# SerialWebViewer Makefile
# Use: make [target]

VERSION=v1.0
BUILD_DIR=build
PROJECT_NAME=SerialWebViewer
LDFLAGS=-ldflags="-s -w"

help:
	@echo "SerialWebViewer - Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  make all         - Build for all platforms"
	@echo "  make windows     - Build for Windows (AMD64)"
	@echo "  make linux       - Build for Linux (AMD64)"
	@echo "  make mac         - Build for macOS (AMD64)"
	@echo "  make mac-arm64   - Build for macOS (ARM64/Apple Silicon)"
	@echo "  make clean       - Clean build directory"
	@echo "  make help        - Show this help message"

all: windows linux mac mac-arm64
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "✨ All builds completed!"
	@echo "📁 Output directory: $(BUILD_DIR)/"
	@ls -lh $(BUILD_DIR)/

windows:
	@echo "📦 Building for Windows (AMD64)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(PROJECT_NAME).exe main.go
	@echo "✅ Windows build completed: $(BUILD_DIR)/$(PROJECT_NAME).exe"

linux:
	@echo "📦 Building for Linux (AMD64)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/serialwebviewer main.go
	@echo "✅ Linux build completed: $(BUILD_DIR)/serialwebviewer"

mac:
	@echo "📦 Building for macOS (AMD64)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(PROJECT_NAME).mac main.go
	@echo "✅ macOS build completed: $(BUILD_DIR)/$(PROJECT_NAME).mac"

mac-arm64:
	@echo "📦 Building for macOS (ARM64/Apple Silicon)..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(PROJECT_NAME)-arm64.mac main.go
	@echo "✅ macOS ARM64 build completed: $(BUILD_DIR)/$(PROJECT_NAME)-arm64.mac"

clean:
	@echo "🧹 Cleaning build directory..."
	@rm -rf $(BUILD_DIR)
	@echo "✅ Clean completed"

install-deps:
	@echo "📦 Installing dependencies..."
	@go get go.bug.st/serial
	@echo "✅ Dependencies installed"

run:
	@echo "🚀 Running SerialWebViewer..."
	@go run main.go

test:
	@echo "🧪 Running tests..."
	@go test -v ./...
