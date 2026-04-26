# Quick Start Guide

## Running the Program

### Windows
```bash
SerialWebViewer.exe
```

### Linux/macOS
```bash
chmod +x serialwebviewer  # or SerialWebViewer.mac
./serialwebviewer
```

## Access the Interface

After launching the program, open in your browser:
```
http://localhost:8088
```

## Basic Usage

### 1. Connect to Serial Port
1. Click the "Config" button in the top right
2. Select COM port and baud rate
3. Click the "Connect" button

### 2. View Logs
- Real-time logs are displayed in the main interface
- Switch between HEX/Text modes
- Show/hide timestamps

### 3. Manage Log Files
- Click the "Files" button to expand the historical file list
- View, download, or delete historical logs

### 4. Customize Theme
1. Select "Custom" from the theme selector
2. Adjust background color, text color, timestamp color, and font size
3. Click "Done" to save settings

### 5. Switch Language
- Click the "EN" or "中文" button in the status bar to switch interface language

## FAQ

**Q: Can't find the serial port?**
A: Click the "Refresh Ports" button in the configuration panel to rescan.

**Q: Where are log files saved?**
A: Log files are saved in the `logs/` folder in the program directory.

**Q: How to change the web port?**
A: Use the `-port` parameter when launching: `SerialWebViewer.exe -port=9000`

**Q: Does it support multi-port monitoring?**
A: The current version only supports single port. Multi-port support is under development.

## Technical Support

For issues, please visit: https://github.com/forqzy/SerialWebViewer/issues
