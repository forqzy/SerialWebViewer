# SerialWebViewer Skill

Help users work with SerialWebViewer, a web-based serial port log viewer tool for monitoring, debugging, and managing serial port communications.

## When to use this skill

Use this skill when the user asks to:
- Monitor or debug serial port communications
- View real-time serial data in a web interface
- Log serial port data to files
- Analyze or manage serial log files
- Configure serial port parameters (baud rate, data bits, parity, etc.)
- Test serial communication protocols
- Debug embedded systems or hardware devices
- Work with COM ports on Windows or /dev/tty* on Linux/macOS

## What the skill can do

- **Connection Management**: Connect/disconnect from serial ports with configurable parameters
- **Real-time Monitoring**: View live serial data in the web interface
- **Log Management**: View, download, or delete historical log files
- **Configuration**: Set baud rate, data bits, parity, stop bits, flow control
- **Display Modes**: Switch between text and HEX display modes
- **Multi-platform Support**: Works on Windows, Linux, and macOS

## How to use

### Starting SerialWebViewer

**From project directory:**
```bash
cd ~/project/SerialWebViewer

# Using build script (recommended)
./build.sh
# Then run the appropriate binary for your platform

# Using make
make all
chmod +x build/serialwebviewer  # or SerialWebViewer.exe on Windows
./build/serialwebviewer

# Or run directly with go
go run main.go
```

**Default web interface:** http://localhost:8088

### Connecting to a Serial Port

**Via Web Interface:**
1. Open http://localhost:8088 in browser
2. Click "Config" button
3. Select COM port and configure parameters
4. Click "Connect"

**Via API (curl):**
```bash
# List available ports
curl http://localhost:8088/api/ports

# Connect to COM10 at 115200 baud
curl -X POST http://localhost:8088/api/connect \
  -H "Content-Type: application/json" \
  -d '{
    "portName": "COM10",
    "baudRate": 115200,
    "dataBits": 8,
    "parity": "None",
    "stopBits": 1,
    "rts": false,
    "dtr": false
  }'
```

**Via API (Python):**
```python
import requests

# Connect to serial port
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
print(response.json())
```

### Monitoring Real-time Data

**Using JavaScript (SSE):**
```javascript
const eventSource = new EventSource('http://localhost:8088/ws');

eventSource.onmessage = function(event) {
    const data = JSON.parse(event.data);
    
    if (data.type === 'log') {
        console.log('Received:', data.content);
        console.log('HEX:', data.hex);
        // Display in your UI
    }
};
```

**Using Python:**
```python
import sseclient
import json

def monitor_serial():
    messages = sseclient.SSEClient('http://localhost:8088/ws')
    
    for msg in messages.events:
        data = json.loads(msg.data)
        if data['type'] == 'log':
            print(f"Data: {data['content']}")
            print(f"HEX: {data['hex']}")
```

### Managing Log Files

**List all logs:**
```bash
curl http://localhost:8088/api/logs/logs
```

**Download a log file:**
```bash
# Browser: http://localhost:8088/api/logs/download?file=com_COM10_20260426_092942.log

# Command line
curl -OJ 'http://localhost:8088/api/logs/download?file=com_COM10_20260426_092942.log'
```

**View log content:**
```bash
# Text mode
curl "http://localhost:8088/api/logs/view?file=com_COM10_20260426_092942.log"

# HEX mode
curl "http://localhost:8088/api/logs/view?file=com_COM10_20260426_092942.log&hex=true"
```

**Delete a log file:**
```bash
curl -X DELETE "http://localhost:8088/api/logs/delete?file=com_COM10_20260426_092942.log"
```

### Checking Connection Status

```bash
curl http://localhost:8088/api/status
```

Returns:
```json
{
  "connected": true,
  "portName": "COM10",
  "logFile": "com_COM10_20260426_092942.log",
  "logSize": 1024000
}
```

### Disconnecting

```bash
curl -X POST http://localhost:8088/api/disconnect
```

## Configuration Parameters

### Serial Port Settings

| Parameter | Description | Options |
|-----------|-------------|---------|
| `portName` | Serial port identifier | COM1-COM256 (Windows), /dev/ttyUSB0 (Linux), /dev/tty.usbserial (macOS) |
| `baudRate` | Data transfer rate | 9600, 19200, 38400, 57600, 115200 |
| `dataBits` | Number of data bits | 5, 6, 7, 8 |
| `parity` | Parity checking | None, Odd, Even |
| `stopBits` | Number of stop bits | 1, 2 |
| `rts` | Request to Send signal | true, false |
| `dtr` | Data Terminal Ready | true, false |

### Command Line Options

```bash
# Use custom port
go run main.go -port=9000

# Build with specific port
GOOS=windows GOARCH=amd64 go build -o SerialWebViewer.exe main.go
```

## Common Use Cases

### 1. Debugging Embedded Systems

Connect to your microcontroller's serial port and monitor debug output in real-time.

### 2. Testing Serial Protocols

Send/receive data with configurable parameters and view in both text and HEX formats.

### 3. Logging Sensor Data

Automatically log all serial data to timestamped files for later analysis.

### 4. Hardware Development

Monitor serial communications during hardware development and debugging.

### 5. Automated Testing

Use the REST API to automate connection, data collection, and verification.

## API Quick Reference

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/status` | GET | Get connection status |
| `/api/ports` | GET | List available ports |
| `/api/connect` | POST | Connect to serial port |
| `/api/disconnect` | POST | Disconnect from serial port |
| `/api/logs/logs` | GET | List log files |
| `/api/logs/view` | GET | View log file content |
| `/api/logs/download` | GET | Download log file |
| `/api/logs/delete` | DELETE | Delete log file |
| `/api/hexmode` | POST | Set HEX display mode |
| `/ws` | SSE | Real-time log stream |

## Troubleshooting

### Port not found
- Refresh port list using the "Refresh Ports" button in the web interface
- Check physical connection
- Verify drivers are installed (especially for USB-serial converters)

### Permission denied (Linux/macOS)
```bash
# Add user to dialout group (Linux)
sudo usermod -a -G dialout $USER

# Or run with sudo (not recommended)
sudo ./serialwebviewer
```

### Port already in use
- Close other applications using the port
- Check if another instance is running
- Use `lsof` (Linux/macOS) or Device Manager (Windows) to find what's using the port

### No data appearing
- Verify baud rate matches the device
- Check cable connections
- Ensure device is powered on
- Try swapping TX/RX wires if using custom cable

## Best Practices

1. **Always disconnect properly** before unplugging the device
2. **Use appropriate baud rate** for your device
3. **Save important logs** before clearing
4. **Check HEX mode** for binary protocols
5. **Monitor disk space** if logging extensively
6. **Use timestamps** for debugging and analysis

## Related Files

- `AGENT.md` - Comprehensive API documentation
- `README.md` - Project overview and setup guide
- `BUILD.md` - Build instructions for all platforms

## Getting Help

- **Documentation**: See `AGENT.md` for detailed API documentation
- **Issues**: https://github.com/forqzy/SerialWebViewer/issues
- **Email**: forqzy@gmail.com
