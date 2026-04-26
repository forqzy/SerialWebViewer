# Changelog

All notable changes to SerialWebViewer will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release preparation

## [1.0.0] - 2026-04-26

### Added
- 🎉 First stable release of SerialWebViewer
- 📊 Real-time serial port data monitoring with SSE
- 🌐 Modern web-based interface with Apple-style design
- 🎨 Multi-theme support (9 preset themes + custom theme)
  - Light, Default, Dracula, Nord, Monokai, Solarized Dark, GitHub Dark, One Dark, Material
- 🌍 Complete bilingual interface (English/Chinese)
- 💾 Automatic log file creation and management
- 📝 Text/HEX dual display modes
- ⚙️ Configurable serial parameters (baud rate, data bits, parity, stop bits)
- 🔌 RTS/DTR flow control support
- 📱 Responsive design for all screen sizes
- 🔧 Custom display settings (background color, text color, timestamp, font size)
- 📁 Historical log file management (view, download, delete)
- 🌐 RESTful API for all operations
- 🔌 Server-Sent Events (SSE) for real-time updates
- 📦 Cross-platform support (Windows, Linux, macOS Intel/ARM64)

### Documentation
- Comprehensive README (English & Chinese)
- Quick Start Guide (English & Chinese)
- Build Instructions for all platforms
- API Documentation for developers (AGENT.md)
- Claude Code skill integration
- MIT License

### Technical
- Built with Go 1.19+
- Serial communication via go.bug.st/serial
- Frontend: Bootstrap 5 + Bootstrap Icons
- Real-time: Server-Sent Events
- No database required (file-based logging)

---

## Format

**[Unreleased]**
- Changes that haven't been released yet

**[1.0.0] - 2026-04-26**
- First stable release

**[1.0.1] - YYYY-MM-DD**
- Bug fixes

**[1.1.0] - YYYY-MM-DD**
- New features
- Minor improvements

**[2.0.0] - YYYY-MM-DD**
- Major version
- Breaking changes
