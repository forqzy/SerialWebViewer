# SerialWebViewer - AI Agent Documentation

This document provides comprehensive information about SerialWebViewer for AI agents to understand and interact with the project.

## 📖 Project Overview

**SerialWebViewer** is a web-based serial port logging and viewing tool built with Go.

### Core Features
- Real-time serial port data monitoring via SSE (Server-Sent Events)
- Web-based interface with modern Apple-style UI
- Multi-theme support (9 preset themes + custom theme)
- Bilingual interface (English/Chinese)
- Automatic log file creation and management
- HEX/Text dual display modes
- Configurable serial port parameters

### Architecture
- **Backend**: Go 1.19+ with HTTP server
- **Frontend**: Vanilla HTML/CSS/JavaScript + Bootstrap 5
- **Real-time Communication**: Server-Sent Events (SSE)
- **Serial Communication**: go.bug.st/serial library
- **State Management**: In-memory with file persistence

### Technical Stack
```
Go 1.19+ → HTTP Server → SSE → Browser
                ↓
          Serial Port → go.bug.st/serial
                ↓
          Log Files (local filesystem)
```

---

## 🌐 API Endpoints

### Base URL
```
http://localhost:8088
```

Default port can be changed with `-port` parameter:
```bash
SerialWebViewer.exe -port=9000
```

---

### 1. Status API

**Endpoint:** `GET /api/status`

**Description:** Get current connection status and statistics

**Response:**
```json
{
  "connected": true,
  "portName": "COM10",
  "logFile": "com_COM10_20260426_092942.log",
  "logSize": 1024000,
  "startTime": "2026-04-26 09:29:42"
}
```

**Fields:**
- `connected` (boolean): Connection status
- `portName` (string): Currently connected port name
- `logFile` (string): Current log filename
- `logSize` (integer): Log file size in bytes
- `startTime` (string): Connection start timestamp

**Use Cases:**
- Check if serial port is connected
- Monitor connection status
- Display current session information

---

### 2. List Ports API

**Endpoint:** `GET /api/ports`

**Description:** Get list of available serial ports

**Response:**
```json
[
  "COM1",
  "COM10",
  "COM12"
]
```

**Use Cases:**
- Populate port selection dropdown
- Detect available serial ports
- Auto-refresh port list

---

### 3. Connect API

**Endpoint:** `POST /api/connect`

**Description:** Connect to a serial port with specified configuration

**Request Headers:**
```
Content-Type: application/json
```

**Request Body:**
```json
{
  "portName": "COM10",
  "baudRate": 115200,
  "dataBits": 8,
  "parity": "None",
  "stopBits": 1,
  "rts": false,
  "dtr": false
}
```

**Parameters:**
- `portName` (string, required): Serial port name (e.g., "COM10", "/dev/ttyUSB0")
- `baudRate` (integer, required): Baud rate (9600, 19200, 38400, 57600, 115200)
- `dataBits` (integer, required): Data bits (5, 6, 7, 8)
- `parity` (string, required): Parity ("None", "Odd", "Even")
- `stopBits` (integer, required): Stop bits (1, 2)
- `rts` (boolean, optional): Request to Send signal (default: false)
- `dtr` (boolean, optional): Data Terminal Ready signal (default: false)

**Success Response:**
```json
{
  "success": true,
  "message": "连接成功"
}
```

**Error Response:**
```json
{
  "success": false,
  "error": "Port already in use"
}
```

**Use Cases:**
- Establish serial connection
- Start logging
- Initialize port with specific parameters

---

### 4. Disconnect API

**Endpoint:** `POST /api/disconnect`

**Description:** Disconnect from currently connected serial port

**Response:**
```json
{
  "success": true,
  "message": "已断开连接"
}
```

**Use Cases:**
- Close serial connection
- Stop logging
- Free port for other applications

---

### 5. Configuration API

**Endpoint:** `GET /api/config`

**Description:** Get current serial port configuration

**Response:**
```json
{
  "portName": "COM10",
  "baudRate": 115200,
  "dataBits": 8,
  "parity": "None",
  "stopBits": 1,
  "rts": false,
  "dtr": false
}
```

**Endpoint:** `POST /api/config`

**Description:** Update serial port configuration

**Request Body:**
```json
{
  "portName": "COM10",
  "baudRate": 115200,
  "dataBits": 8,
  "parity": "None",
  "stopBits": 1,
  "rts": false,
  "dtr": false
}
```

**Response:**
```json
{
  "success": true,
  "message": "配置已更新"
}
```

**Use Cases:**
- Save connection parameters
- Pre-configure before connecting
- Remember user preferences

---

### 6. HEX Mode API

**Endpoint:** `POST /api/hexmode`

**Description:** Set HEX display mode for logging

**Request Headers:**
```
Content-Type: application/json
```

**Request Body:**
```json
{
  "hexMode": true
}
```

**Response:**
```json
{
  "success": true,
  "hexMode": true,
  "message": "HEX模式已更新"
}
```

**Use Cases:**
- Switch between text and HEX display
- Control log file format
- Set default display mode

---

### 7. List Logs API

**Endpoint:** `GET /api/logs/logs`

**Description:** Get list of all log files in the logs directory

**Response:**
```json
[
  {
    "name": "com_COM10_20260426_092942.log",
    "size": 1024000,
    "modTime": "2026-04-26 09:30:15"
  },
  {
    "name": "com_COM10_20260426_085355.log",
    "size": 155000,
    "modTime": "2026-04-26 08:59:48"
  }
]
```

**Note:** Files are sorted by modification time (newest first).

**Fields:**
- `name` (string): Log filename
- `size` (integer): File size in bytes
- `modTime` (string): Last modification timestamp

**Use Cases:**
- Display historical log files
- File browser interface
- File management operations

---

### 8. View Log API

**Endpoint:** `GET /api/logs/view?file={filename}&hex={hex}`

**Description:** View log file content

**Query Parameters:**
- `file` (string, required): Log filename
- `hex` (boolean, optional): Return in HEX format (default: false)

**Example:**
```
GET /api/logs/view?file=com_COM10_20260426_092942.log
GET /api/logs/view?file=com_COM10_20260426_092942.log&hex=true
```

**Response:**
- Content-Type: `text/plain; charset=utf-8`
- Body: Log file content (text or HEX format)

**HEX Format Example:**
```
48 45 4C 4C 4F 0A 57 4F 52 4C 44
```

**Use Cases:**
- Online log viewing
- HEX debugging
- In-browser log inspection

---

### 9. Download Log API

**Endpoint:** `GET /api/logs/download?file={filename}`

**Description:** Download log file

**Query Parameters:**
- `file` (string, required): Log filename

**Response Headers:**
```
Content-Type: application/octet-stream
Content-Disposition: attachment; filename="com_COM10_20260426_092942.log"
Content-Length: 1024000
Cache-Control: no-cache, no-store, must-revalidate
```

**Use Cases:**
- Download log files for offline analysis
- Export data
- Archive logs

---

### 10. Delete Log API

**Endpoint:** `DELETE /api/logs/delete?file={filename}`

**Description:** Delete a log file

**Query Parameters:**
- `file` (string, required): Log filename

**Response:**
```json
{
  "success": true,
  "message": "文件已删除"
}
```

**Use Cases:**
- Clean up old logs
- Manage disk space
- Remove sensitive data

---

### 11. Current Log API

**Endpoint:** `GET /api/logs/current`

**Description:** Stream current active log file

**Response:**
- File stream of currently active log
- Returns 404 if not connected

**Use Cases:**
- Real-time log file access
- Debug current session
- Monitor active logging

---

## 🔌 Real-time Communication (SSE)

### WebSocket/SSE Endpoint

**Endpoint:** `GET /ws`

**Description:** Server-Sent Events endpoint for real-time log streaming

**Connection:**
```javascript
const eventSource = new EventSource('http://localhost:8088/ws');

eventSource.onmessage = function(event) {
    const data = JSON.parse(event.data);
    
    if (data.type === 'log') {
        // Handle log data
        console.log('Content:', data.content);
        console.log('HEX:', data.hex);
    } else if (data.type === 'connected') {
        console.log('SSE connected');
    } else if (data.type === 'heartbeat') {
        // Keep-alive heartbeat
    }
};

eventSource.onerror = function(error) {
    console.error('SSE error:', error);
    // SSE will auto-reconnect
};
```

**Message Formats:**

**1. Connection Established:**
```json
{
  "type": "connected"
}
```

**2. Log Data (Real-time):**
```json
{
  "type": "log",
  "content": "Serial data received",
  "hex": "53 65 72 69 61 6C 0A"
}
```

**3. Info Message:**
```json
{
  "type": "info",
  "content": "已连接到日志服务器"
}
```

**4. Heartbeat (every 30s):**
```json
{
  "type": "heartbeat"
}
```

**Features:**
- Auto-reconnect on connection loss
- Bidirectional communication
- Low latency
- Native browser support

**Use Cases:**
- Real-time log display
- Live data monitoring
- Instant notifications
- Connection status updates

---

## 🗄️ File System Structure

### Directory Layout
```
SerialWebViewer/
├── main.go              # Main application
├── logs/                 # Log files directory
│   ├── com_COM10_20260426_092942.log
│   └── com_COM10_20260426_085355.log
└── build/                # Compiled binaries (gitignored)
```

### Log File Naming Convention
```
com_{PORTNAME}_{YYYYMMDD_HHMMSS}.log
```

Example: `com_COM10_20260426_092942.log`

---

## ⚙️ Configuration Options

### Command Line Parameters

| Parameter | Default | Description |
|-----------|---------|-------------|
| `-port` | 8088 | Web server port |

**Example:**
```bash
SerialWebViewer.exe -port=9000
```

### Serial Port Parameters

| Parameter | Type | Options | Default |
|-----------|------|---------|---------|
| Baud Rate | int | 9600, 19200, 38400, 57600, 115200 | 115200 |
| Data Bits | int | 5, 6, 7, 8 | 8 |
| Parity | string | None, Odd, Even | None |
| Stop Bits | int | 1, 2 | 1 |
| RTS | bool | true, false | false |
| DTR | bool | true, false | false |

### Display Modes

| Mode | Description |
|------|-------------|
| Text | Display data as plain text |
| HEX | Display data in hexadecimal format |

### Themes

Available themes:
- `light` - White background (default)
- `default` - Dark gray
- `dracula` - Purple tones
- `nord` - Nordic style
- `monokai` - Classic dark
- `solarized-dark` - Eye protection
- `github-dark` - GitHub dark
- `one-dark` - Atom editor style
- `material` - Material Design
- `custom` - User customizable

---

## 🔐 Security Considerations

### Port Validation
- Only lists available ports from OS
- Validates port names
- Sanitizes file paths

### File Operations
- Path traversal protection
- File extension validation (.log only)
- Safe file naming

### CORS
- Same-origin policy by default
- Local network access only

### Recommendations for Production
- Add authentication
- Implement HTTPS/TLS
- Add rate limiting
- Input sanitization
- Access control

---

## 🧪 Testing with AI Agents

### Example: Test Connection Flow

```bash
# 1. Get available ports
curl http://localhost:8088/api/ports

# 2. Connect to COM10
curl -X POST http://localhost:8088/api/connect \
  -H "Content-Type: application/json" \
  -d '{
    "portName": "COM10",
    "baudRate": 115200,
    "dataBits": 8,
    "parity": "None",
    "stopBits": 1
  }'

# 3. Check status
curl http://localhost:8088/api/status

# 4. List logs
curl http://localhost:8088/api/logs/logs

# 5. Disconnect
curl -X POST http://localhost:8088/api/disconnect
```

### Example: Monitor Real-time Logs (Python)

```python
import requests
import json
import sseclient

def connect_to_serial():
    # Connect
    response = requests.post(
        'http://localhost:8088/api/connect',
        json={
            'portName': 'COM10',
            'baudRate': 115200,
            'dataBits': 8,
            'parity': 'None',
            'stopBits': 1
        }
    )
    return response.json()

def monitor_logs():
    # Connect to SSE
    messages = sseclient.SSEClient(
        'http://localhost:8088/ws'
    )
    
    for msg in messages.events:
        data = json.loads(msg.data)
        if data['type'] == 'log':
            print(f"Log: {data['content']}")
            print(f"HEX: {data['hex']}")
```

---

## 📊 State Management

### Global State Variables
```go
var (
    currentPort     serial.Port       // Active serial port
    portConfig     PortConfig          // Current configuration
    isConnected    bool                // Connection status
    logFile       *os.File            // Log file handle
    hexMode       bool                // HEX mode flag
)
```

### State Persistence
- Configuration saved to browser cookies
- Log files saved to disk
- No database required

---

## 🚨 Error Handling

### Common Errors

| Error | Cause | Solution |
|-------|--------|----------|
| "Port not found" | Port disconnected | Refresh port list |
| "Port already in use" | Port occupied by other app | Close other application |
| "Invalid baud rate" | Unsupported baud rate | Use standard values |
| "Permission denied" | Insufficient privileges | Run as administrator |

### Error Response Format
```json
{
  "success": false,
  "error": "Error message description"
}
```

---

## 📝 Best Practices for AI Agents

### 1. Always Check Status Before Operations
```javascript
// Check if connected before sending data
const status = await fetch('/api/status').then(r => r.json());
if (!status.connected) {
    // Connect first
}
```

### 2. Handle SSE Reconnection
```javascript
eventSource.onerror = function(error) {
    // SSE will auto-reconnect
    // Log for monitoring
    console.log('SSE connection lost, will retry...');
};
```

### 3. Validate Port Availability
```javascript
const ports = await fetch('/api/ports').then(r => r.json());
if (!ports.includes(desiredPort)) {
    // Refresh or handle error
}
```

### 4. Use HEX Mode for Binary Data
```javascript
// Set HEX mode for binary protocols
await fetch('/api/hexmode', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({hexMode: true})
});
```

### 5. Implement Proper Cleanup
```javascript
window.addEventListener('beforeunload', () => {
    if (isConnected) {
        navigator.sendBeacon('/api/disconnect');
    }
});
```

---

## 🔗 Quick Reference

### Essential APIs for Common Tasks

| Task | API | Method |
|------|-----|--------|
| Check connection | `/api/status` | GET |
| List ports | `/api/ports` | GET |
| Connect | `/api/connect` | POST |
| Disconnect | `/api/disconnect` | POST |
| Get logs list | `/api/logs/logs` | GET |
| View log | `/api/logs/view` | GET |
| Download log | `/api/logs/download` | GET |
| Delete log | `/api/logs/delete` | DELETE |
| Real-time data | `/ws` | SSE |
| Set HEX mode | `/api/hexmode` | POST |

---

## 📞 Support & Resources

- **GitHub**: https://github.com/forqzy/SerialWebViewer
- **Issues**: https://github.com/forqzy/SerialWebViewer/issues
- **Email**: forqzy@gmail.com

---

## 📜 License

MIT License - See LICENSE file for details

---

**Last Updated:** 2026-04-26  
**Version:** 1.0.0  
**Maintained by:** forqzy
