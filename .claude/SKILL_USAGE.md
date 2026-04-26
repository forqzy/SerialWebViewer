# SerialWebViewer Skill - Usage Guide

This guide demonstrates how Claude Code can use the SerialWebViewer skill to help users with serial port monitoring and debugging tasks.

## Skill Overview

**Skill Name:** `serialwebviewer`  
**Location:** `.claude/skills/serialwebviewer.md`

## What This Skill Enables

With this skill installed, Claude Code can:

1. **Understand SerialWebViewer** - Know what it is, what it does, and when to use it
2. **Help Users** - Assist with installation, configuration, and usage
3. **Generate Code** - Create scripts and examples for interacting with SerialWebViewer
4. **Debug Issues** - Troubleshoot common serial port problems
5. **API Integration** - Help integrate SerialWebViewer APIs into user projects

## Example Interactions

### Example 1: Basic Usage Questions

**User:** "How do I connect to my Arduino's serial port?"

**Claude's Response (with skill):**
```
I can help you connect to your Arduino using SerialWebViewer.

First, start SerialWebViewer:
cd ~/project/SerialWebViewer
./build.sh
./build/serialwebviewer

Then open http://localhost:8088 in your browser.

To connect via API:
curl -X POST http://localhost:8088/api/connect \
  -H "Content-Type: application/json" \
  -d '{
    "portName": "/dev/ttyUSB0",
    "baudRate": 9600,
    "dataBits": 8,
    "parity": "None",
    "stopBits": 1
  }'

Typical Arduino uses:
- Port: /dev/ttyUSB0 (Linux/macOS) or COM3+ (Windows)
- Baud rate: 9600
- Data bits: 8
- Parity: None
- Stop bits: 1
```

### Example 2: Code Generation

**User:** "Write a Python script to monitor serial data and save to database"

**Claude's Response (with skill):**
```python
import requests
import sseclient
import json
import sqlite3
from datetime import datetime

# Connect to SerialWebViewer
BASE_URL = "http://localhost:8088"

def connect_serial(port_name, baud_rate=9600):
    """Connect to serial port"""
    response = requests.post(
        f"{BASE_URL}/api/connect",
        json={
            "portName": port_name,
            "baudRate": baud_rate,
            "dataBits": 8,
            "parity": "None",
            "stopBits": 1
        }
    )
    return response.json()

def monitor_and_save():
    """Monitor serial data and save to database"""
    # Setup database
    conn = sqlite3.connect('serial_data.db')
    c = conn.cursor()
    c.execute('''CREATE TABLE IF NOT EXISTS serial_logs
                    (timestamp TEXT, data TEXT, hex_data TEXT)''')
    
    # Connect to SSE stream
    messages = sseclient.SSEClient(f'{BASE_URL}/ws')
    
    for msg in messages.events:
        data = json.loads(msg.data)
        
        if data['type'] == 'log':
            # Save to database
            now = datetime.now().isoformat()
            c.execute(
                "INSERT INTO serial_logs VALUES (?,?,?)",
                (now, data['content'], data['hex'])
            )
            conn.commit()
            print(f"[{now}] {data['content']}")
            
    conn.close()

# Usage
connect_serial("/dev/ttyUSB0", 9600)
monitor_and_save()
```

### Example 3: Troubleshooting

**User:** "I'm getting 'port not found' error"

**Claude's Response (with skill):**
```
Let me help you troubleshoot the "port not found" error.

Common solutions:

1. **Refresh the port list:**
   curl http://localhost:8088/api/ports
   
2. **Check physical connection:**
   - Ensure USB cable is connected
   - Try a different USB port

3. **Verify drivers (Windows):**
   - Open Device Manager
   - Look under "Ports (COM & LPT)"
   - Install USB-serial driver if needed

4. **Check permissions (Linux/macOS):**
   # Linux
   sudo usermod -a -G dialout $USER
   # Then logout and login again
   
   # macOS
   sudo chmod 666 /dev/tty.usbserial*

5. **Verify device is powered on:**
   Some devices need to be powered to show the serial port

Would you like me to help you check any of these?
```

### Example 4: API Integration

**User:** "Integrate SerialWebViewer into my web app"

**Claude's Response (with skill):**
```javascript
// SerialWebViewer integration for your web app

class SerialWebViewerClient {
    constructor(baseUrl = 'http://localhost:8088') {
        this.baseUrl = baseUrl;
    }
    
    async getAvailablePorts() {
        const response = await fetch(`${this.baseUrl}/api/ports`);
        return await response.json();
    }
    
    async connect(config) {
        const response = await fetch(`${this.baseUrl}/api/connect`, {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(config)
        });
        return await response.json();
    }
    
    async disconnect() {
        const response = await fetch(`${this.baseUrl}/api/disconnect`, {
            method: 'POST'
        });
        return await response.json();
    }
    
    subscribeToLogs(callback) {
        const eventSource = new EventSource(`${this.baseUrl}/ws`);
        
        eventSource.onmessage = (event) => {
            const data = JSON.parse(event.data);
            if (data.type === 'log') {
                callback(data);
            }
        };
        
        return eventSource;
    }
    
    async getLogFiles() {
        const response = await fetch(`${this.baseUrl}/api/logs/logs`);
        return await response.json();
    }
    
    async downloadLog(filename) {
        window.location.href = `${this.baseUrl}/api/logs/download?file=${filename}`;
    }
}

// Usage in your app
const serial = new SerialWebViewerClient();

// Connect
await serial.connect({
    portName: 'COM10',
    baudRate: 115200,
    dataBits: 8,
    parity: 'None',
    stopBits: 1
});

// Subscribe to real-time logs
serial.subscribeToLogs((data) => {
    console.log('Received:', data.content);
    // Update your UI
});
```

## Skill Metadata

From `skill.json`:
```json
{
  "name": "serialwebviewer",
  "title": "SerialWebViewer",
  "description": "Web-based serial port log viewer...",
  "tags": ["serial", "com-port", "logging", "debugging", ...],
  "categories": ["development-tools", "hardware", "debugging"]
}
```

## Activating the Skill

The skill is automatically available when:
1. Working in the SerialWebViewer project directory
2. The `.claude/skills/` directory exists
3. Claude Code scans the skill definitions

## Testing the Skill

### Test 1: Ask about installation
```bash
# In Claude Code
"How do I install SerialWebViewer on Linux?"
```

### Test 2: Request code examples
```bash
"Write a Node.js script to connect to SerialWebViewer and log data to a file"
```

### Test 3: Debugging help
```bash
"My serial connection keeps dropping, what could be wrong?"
```

### Test 4: API usage
```bash
"How can I use the SerialWebViewer API to automate my tests?"
```

## Best Practices for AI Responses

When using this skill, Claude Code should:

1. **Check if SerialWebViewer is running** before suggesting API calls
2. **Provide platform-specific commands** (Windows vs Linux/macOS)
3. **Include error handling** in code examples
4. **Suggest common parameters** for popular devices (Arduino, ESP32, etc.)
5. **Reference the official docs** (AGENT.md) for detailed API information
6. **Offer troubleshooting steps** when operations fail

## Skill Files Structure

```
.claude/
├── .gitkeep                    # Keep directory in git
└── skills/
    ├── serialwebviewer.md      # Main skill documentation
    └── serialwebviewer/
        └── skill.json          # Skill metadata
```

## Related Documentation

- **AGENT.md** - Complete API reference for developers
- **README.md** - Project overview and setup
- **BUILD.md** - Build instructions

## Updating the Skill

To update the skill:
1. Edit `.claude/skills/serialwebviewer.md`
2. Test changes with Claude Code
3. Commit updates to git
4. Document changes in commit message

## Support

For skill-related issues:
- Check documentation in AGENT.md
- Review examples in this file
- Open issue at: https://github.com/forqzy/SerialWebViewer/issues

---

**Last Updated:** 2026-04-26  
**Version:** 1.0.0
