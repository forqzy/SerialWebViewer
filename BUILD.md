# Building SerialWebViewer

This guide explains how to build SerialWebViewer for different platforms.

## Prerequisites

- Go 1.19 or higher
- Git (optional, for cloning the repository)

## Quick Build

### Using Build Scripts

#### Linux/macOS
```bash
# Make build script executable (first time only)
chmod +x build.sh

# Build for all platforms
./build.sh
```

#### Windows
```cmd
# Double-click build.bat
# or run from command line
build.bat
```

### Using Makefile

```bash
# Build for all platforms
make all

# Build for specific platform
make windows    # Windows
make linux      # Linux
make mac        # macOS (Intel)
make mac-arm64  # macOS (Apple Silicon)

# Clean build directory
make clean

# Show help
make help
```

## Manual Build

### Windows (AMD64)
```bash
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o SerialWebViewer.exe main.go
```

### Linux (AMD64)
```bash
go build -ldflags="-s -w" -o serialwebviewer main.go
```

### macOS (Intel)
```bash
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o SerialWebViewer.mac main.go
```

### macOS (Apple Silicon)
```bash
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o SerialWebViewer-arm64.mac main.go
```

## Build Flags Explanation

- `-ldflags="-s -w"`: Strip debug information to reduce binary size
  - `-s`: Remove symbol table
  - `-w`: Remove DWARF debug info
- Result: Smaller binary files (~50% size reduction)

## Output Files

After building, you'll find the following files:

| Platform | Output File | Description |
|----------|-------------|-------------|
| Windows | `SerialWebViewer.exe` | Windows executable |
| Linux | `serialwebviewer` | Linux binary |
| macOS (Intel) | `SerialWebViewer.mac` | macOS binary (Intel) |
| macOS (ARM64) | `SerialWebViewer-arm64.mac` | macOS binary (Apple Silicon) |

## Running the Built Binaries

### Windows
```cmd
SerialWebViewer.exe
```

### Linux
```bash
chmod +x serialwebviewer
./serialwebviewer
```

### macOS
```bash
# For Intel Macs
chmod +x SerialWebViewer.mac
./SerialWebViewer.mac

# For Apple Silicon Macs
chmod +x SerialWebViewer-arm64.mac
./SerialWebViewer-arm64.mac
```

## Creating a Release

### Using GitHub CLI

```bash
# Build all platforms
make all

# Create a new release
gh release create v1.0 \
  --title "SerialWebViewer v1.0" \
  --notes "Release notes here" \
  build/*
```

### Manual Release

1. Build all platforms using one of the methods above
2. Go to GitHub → Releases → "Create a new release"
3. Tag version: `v1.0`
4. Upload the binaries from the `build/` directory
5. Publish the release

## Cross-Compilation Requirements

### Linux Building Windows Binaries

No additional tools required for AMD64. For other architectures:

```bash
# Install cross-compilation tools
sudo apt-get install gcc-mingw-w64
```

### macOS Building for Other Platforms

No additional tools required for basic cross-compilation.

### Windows Building for Other Platforms

No additional tools required for basic cross-compilation.

## Troubleshooting

### "go: command not found"
- Install Go from https://golang.org/dl/
- Ensure Go is in your PATH

### "permission denied" when running build scripts
```bash
chmod +x build.sh  # Linux/macOS
```

### Build fails on macOS (M1/M2)
- Make sure Xcode command line tools are installed:
```bash
xcode-select --install
```

### Binary won't run on Windows
- Windows may block unsigned binaries
- Right-click → Properties → Unblock
- Or run: `powershell.exe -Command "Unblock-File SerialWebViewer.exe"`

## Advanced Build Options

### Enable Race Detector
```bash
go build -race -o SerialWebViewer.exe main.go
```

### Verbose Build
```bash
go build -v -o SerialWebViewer.exe main.go
```

### Build with Specific Go Version
```bash
go1.19 build -o SerialWebViewer.exe main.go
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Build

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    strategy:
      matrix:
        include:
          - goos: windows
            goarch: amd64
            output: SerialWebViewer.exe
          - goos: linux
            goarch: amd64
            output: serialwebviewer
          - goos: darwin
            goarch: amd64
            output: SerialWebViewer.mac
          - goos: darwin
            goarch: arm64
            output: SerialWebViewer-arm64.mac
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.19'
      - name: Build
        run: GOOS=${{ matrix.goos }} GOARCH=${{ matrix.goarch }} go build -ldflags="-s -w" -o ${{ matrix.output }} main.go
      - name: Upload
        uses: actions/upload-artifact@v3
        with:
          name: ${{ matrix.output }}
          path: ${{ matrix.output }}
```

## Support

For build-related issues, please visit:
https://github.com/forqzy/SerialWebViewer/issues
