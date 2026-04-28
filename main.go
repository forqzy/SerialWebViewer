// SerialWebViewer - A modern web-based serial port log viewer
// Copyright (c) 2025 forqzy <forqzy@gmail.com>
//
// MIT License
// See LICENSE file for details

package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.bug.st/serial"
)

// 全局变量
var (
	currentPort     serial.Port
	webPort        = flag.String("port", "8088", "Web服务器端口")
	portConfig     = PortConfig{
		PortName:     "COM1",
		BaudRate:     115200,
		DataBits:     8,
		Parity:       "None",
		StopBits:     1,
		RTS:          false,
		DTR:          false,
	}
	isConnected    = false
	logFile       *os.File
	logWriter     *bufio.Writer
	mu            sync.Mutex
	logFiles      []LogFile
	clients       = make(map[*http.ResponseWriter]bool)
	clientsMu     sync.Mutex
	hexMode       = false // 后端HEX模式状态
	stopLogging   = make(chan struct{}) // 用于停止日志记录协程
)

// PortConfig 串口配置结构
type PortConfig struct {
	PortName string `json:"portName"`
	BaudRate int    `json:"baudRate"`
	DataBits int    `json:"dataBits"`
	Parity   string `json:"parity"`
	StopBits int    `json:"stopBits"`
	RTS      bool   `json:"rts"` // Request to Send
	DTR      bool   `json:"dtr"` // Data Terminal Ready
}

// LogFile 日志文件信息
type LogFile struct {
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	ModTime string `json:"modTime"`
}

// 状态响应
type StatusResponse struct {
	Connected  bool   `json:"connected"`
	PortName   string `json:"portName"`
	LogFile    string `json:"logFile"`
	LogSize    int64  `json:"logSize"`
	StartTime  string `json:"startTime"`
}

func main() {
	// 解析命令行参数
	flag.Parse()

	// 确保日志目录存在
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatal("无法创建日志目录:", err)
	}

	// 启动时扫描日志文件
	scanLogFiles()

	// 静态文件服务
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// API路由
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/api/status", statusHandler)
	http.HandleFunc("/api/ports", listPortsHandler)
	http.HandleFunc("/api/connect", connectHandler)
	http.HandleFunc("/api/disconnect", disconnectHandler)
	http.HandleFunc("/api/config", configHandler)
	http.HandleFunc("/api/hexmode", setHexModeHandler) // 新增：设置HEX模式
	http.HandleFunc("/api/logs/logs", listLogsHandler)
	http.HandleFunc("/api/logs/view", viewLogHandler)
	http.HandleFunc("/api/logs/download", downloadLogHandler)
	http.HandleFunc("/api/logs/delete", deleteLogHandler)
	http.HandleFunc("/api/logs/current", currentLogHandler)

	// WebSocket用于实时日志显示
	http.HandleFunc("/ws", wsHandler)

	// 构建监听地址
	listenAddr := ":" + *webPort

	fmt.Println("🚀 SerialWebViewer 启动")
	fmt.Printf("🌐 Web界面: http://localhost:%s\n", *webPort)
	fmt.Println("📁 日志目录: ./logs/")
	fmt.Printf("🔧 使用 -port 参数可修改端口，如: -port=9000\n")

	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}

// 首页处理器
func indexHandler(w http.ResponseWriter, r *http.Request) {
	htmlContent := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SerialWebViewer</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.8.0/font/bootstrap-icons.css">
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            background-color: #f5f5f7;
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Microsoft YaHei', sans-serif;
            overflow: hidden;
            height: 100vh;
            color: #1d1d1f;
        }
        .app-container {
            display: flex;
            height: 100vh;
        }
        .main-content {
            flex: 1;
            display: flex;
            flex-direction: column;
            padding: 0;
            overflow: hidden;
        }
        .status-bar {
            background: rgba(255, 255, 255, 0.6);
            backdrop-filter: blur(10px);
            -webkit-backdrop-filter: blur(10px);
            padding: 8px 20px;
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: 16px;
            border-bottom: 1px solid rgba(0, 0, 0, 0.08);
            flex-shrink: 0;
        }
        .status-buttons {
            display: flex;
            align-items: center;
            gap: 6px;
        }
        .status-buttons .btn {
            height: 28px;
            font-size: 13px;
            padding: 0 12px;
        }
        .status-buttons .btn.btn-outline-primary {
            border: 1px solid rgba(0, 122, 255, 0.4);
            color: #007aff;
            background: transparent;
        }
        .status-buttons .btn.btn-outline-primary:hover {
            background: rgba(0, 122, 255, 0.08);
            border-color: #007aff;
        }
        .status-buttons .btn .btn-icon {
            font-size: 12px;
            margin-right: 4px;
        }
        .status-indicator {
            width: 8px;
            height: 8px;
            border-radius: 50%;
            display: inline-block;
            margin-right: 8px;
        }
        .status-connected {
            background-color: #34c759;
            box-shadow: 0 0 0 3px rgba(52, 199, 89, 0.2);
        }
        .status-disconnected {
            background-color: #ff3b30;
            box-shadow: 0 0 0 3px rgba(255, 59, 48, 0.2);
        }
        .status-info {
            color: #86868b;
            font-size: 13px;
            display: flex;
            align-items: center;
            gap: 15px;
            font-weight: 500;
        }
        .log-section {
            flex: 1;
            display: flex;
            flex-direction: column;
            overflow: hidden;
            background: #ffffff;
        }
        .log-header {
            background: rgba(255, 255, 255, 0.8);
            backdrop-filter: blur(10px);
            -webkit-backdrop-filter: blur(10px);
            padding: 10px 20px;
            display: flex;
            align-items: center;
            justify-content: space-between;
            border-bottom: 1px solid rgba(0, 0, 0, 0.08);
            flex-shrink: 0;
        }
        .log-title {
            color: #1d1d1f;
            font-size: 13px;
            font-weight: 600;
            display: flex;
            align-items: center;
            gap: 6px;
            letter-spacing: -0.2px;
        }
        .log-controls {
            display: flex;
            gap: 6px;
        }
        .log-display {
            background-color: #ffffff;
            color: #1d1d1f;
            padding: 16px;
            flex: 1;
            overflow-y: auto;
            overflow-x: auto;
            font-family: 'SF Mono', 'Monaco', 'Consolas', 'Courier New', monospace;
            font-size: 13px;
            line-height: 1.6;
            border: 1px solid rgba(0, 0, 0, 0.12);
            border-radius: 8px;
            margin: 0 16px 16px 16px;
            white-space: pre-wrap;
            word-break: keep-all;
        }
        .log-entry {
            margin-bottom: 2px;
            display: block;
            white-space: pre;
        }
        .log-timestamp {
            color: #0a84ff;
        }
        .sidebar {
            width: 300px;
            background: rgba(255, 255, 255, 0.95);
            backdrop-filter: blur(20px);
            -webkit-backdrop-filter: blur(20px);
            border-left: 1px solid rgba(0, 0, 0, 0.08);
            display: flex;
            flex-direction: column;
            transform: translateX(100%);
            transition: transform 0.25s cubic-bezier(0.4, 0, 0.2, 1);
            position: fixed;
            right: 0;
            top: 0;
            height: 100vh;
            z-index: 1000;
            box-shadow: -3px 0 20px rgba(0, 0, 0, 0.08);
        }
        .sidebar.show {
            transform: translateX(0);
        }
        .sidebar-header {
            padding: 16px 20px;
            border-bottom: 1px solid rgba(0, 0, 0, 0.08);
            display: flex;
            align-items: center;
            justify-content: space-between;
        }
        .sidebar-title {
            color: #1d1d1f;
            font-size: 16px;
            font-weight: 600;
            letter-spacing: -0.4px;
        }
        .sidebar-content {
            flex: 1;
            overflow-y: auto;
            padding: 20px;
        }
        .config-group {
            margin-bottom: 20px;
        }
        .config-label {
            color: #86868b;
            font-size: 11px;
            margin-bottom: 6px;
            text-transform: uppercase;
            letter-spacing: 0.4px;
            font-weight: 600;
        }
        .form-select, .form-control {
            background: #ffffff;
            border: 1px solid rgba(0, 0, 0, 0.12);
            color: #1d1d1f;
            border-radius: 6px;
            padding: 8px 12px;
            font-size: 13px;
            transition: all 0.15s ease;
        }
        .form-select:focus, .form-control:focus {
            background: #ffffff;
            border-color: #007aff;
            color: #1d1d1f;
            box-shadow: 0 0 0 3px rgba(0, 122, 255, 0.1);
            outline: none;
        }
        .form-select option {
            background: #ffffff;
            color: #1d1d1f;
        }
        .log-controls .form-select {
            padding: 4px 8px;
            font-size: 12px;
            width: auto;
            min-width: 130px;
            height: 28px;
            border-radius: 5px;
        }
        .form-check {
            margin-bottom: 10px;
        }
        .form-check-input {
            background: #ffffff;
            border: 1px solid rgba(0, 0, 0, 0.2);
            width: 16px;
            height: 16px;
            border-radius: 4px;
            cursor: pointer;
            transition: all 0.15s ease;
        }
        .form-check-input:checked {
            background: #007aff;
            border-color: #007aff;
        }
        .form-check-label {
            color: #1d1d1f;
            font-size: 13px;
            margin-left: 8px;
            cursor: pointer;
            font-weight: 500;
        }
        .btn {
            border-radius: 6px;
            font-size: 13px;
            padding: 6px 12px;
            font-weight: 500;
            transition: all 0.15s ease;
            border: none;
            cursor: pointer;
            letter-spacing: -0.2px;
            display: inline-flex;
            align-items: center;
            gap: 4px;
            white-space: nowrap;
        }
        .btn:hover {
            transform: translateY(-1px);
        }
        .btn:active {
            transform: translateY(0);
        }
        .btn-primary {
            background: #007aff;
            color: white;
        }
        .btn-primary:hover {
            background: #0051d5;
            box-shadow: 0 2px 8px rgba(0, 122, 255, 0.25);
        }
        .btn-success {
            background: #34c759;
            color: white;
        }
        .btn-success:hover {
            background: #248a3d;
            box-shadow: 0 2px 8px rgba(52, 199, 89, 0.25);
        }
        .btn-danger {
            background: #ff3b30;
            color: white;
        }
        .btn-danger:hover {
            background: #d70015;
            box-shadow: 0 2px 8px rgba(255, 59, 48, 0.25);
        }
        .btn-light {
            background: #ffffff;
            border: 1px solid rgba(0, 0, 0, 0.1);
            color: #1d1d1f;
        }
        .btn-light:hover {
            background: #f5f5f7;
        }
        .btn-dark {
            background: rgba(0, 0, 0, 0.04);
            border: 1px solid rgba(0, 0, 0, 0.08);
            color: #1d1d1f;
        }
        .btn-dark:hover {
            background: rgba(0, 0, 0, 0.08);
        }
        .btn-outline-secondary {
            border: 1px solid rgba(0, 0, 0, 0.12);
            color: #86868b;
            background: transparent;
        }
        .btn-outline-secondary:hover {
            background: rgba(0, 0, 0, 0.04);
            color: #1d1d1f;
            border-color: rgba(0, 0, 0, 0.2);
        }
        .btn-outline-primary {
            border: 1px solid rgba(0, 122, 255, 0.5);
            color: #007aff;
            background: transparent;
        }
        .btn-outline-primary:hover {
            background: rgba(0, 122, 255, 0.08);
            border-color: #007aff;
        }
        .btn-outline-success {
            border: 1px solid rgba(52, 199, 89, 0.5);
            color: #34c759;
            background: transparent;
        }
        .btn-outline-success:hover {
            background: rgba(52, 199, 89, 0.08);
            border-color: #34c759;
        }
        .btn-outline-danger {
            border: 1px solid rgba(255, 59, 48, 0.5);
            color: #ff3b30;
            background: transparent;
        }
        .btn-outline-danger:hover {
            background: rgba(255, 59, 48, 0.08);
            border-color: #ff3b30;
        }
        .btn-sm {
            padding: 4px 10px;
            font-size: 12px;
            border-radius: 5px;
        }
        .btn-lg {
            padding: 8px 20px;
            font-size: 14px;
        }
        .btn-icon {
            margin-right: 3px;
            font-size: 12px;
        }
        .log-controls .btn {
            height: 28px;
        }
        .badge {
            padding: 3px 8px;
            font-size: 10px;
            font-weight: 600;
            border-radius: 4px;
            letter-spacing: 0.3px;
        }
        .files-section {
            background: rgba(255, 255, 255, 0.8);
            backdrop-filter: blur(10px);
            -webkit-backdrop-filter: blur(10px);
            max-height: 250px;
            overflow-y: auto;
            transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
            margin: 0 16px 16px 16px;
            border-radius: 8px;
            border: 1px solid rgba(0, 0, 0, 0.12);
        }
        .files-section.hidden {
            max-height: 0;
            overflow: hidden;
        }
        .files-header {
            padding: 12px 16px;
            border-bottom: 1px solid rgba(0, 0, 0, 0.08);
            display: flex;
            align-items: center;
            justify-content: space-between;
            border-radius: 8px 8px 0 0;
        }
        .files-title {
            color: #1d1d1f;
            font-size: 14px;
            font-weight: 600;
            letter-spacing: -0.3px;
        }
        .files-table {
            width: 100%;
            border-collapse: collapse;
            border-radius: 0 0 8px 8px;
            overflow: hidden;
        }
        .files-table th {
            background: rgba(0, 0, 0, 0.02);
            color: #86868b;
            font-size: 11px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            padding: 10px 20px;
            text-align: left;
            font-weight: 600;
        }
        .files-table td {
            padding: 12px 20px;
            border-bottom: 1px solid rgba(0, 0, 0, 0.05);
            color: #1d1d1f;
            font-size: 13px;
        }
        .files-table tr:hover {
            background: rgba(0, 0, 0, 0.02);
        }
        .file-size {
            color: #86868b;
            font-weight: 500;
        }
        .sidebar-footer {
            padding: 16px 20px;
            border-top: 1px solid rgba(0, 0, 0, 0.08);
            background: rgba(0, 0, 0, 0.02);
        }
        .custom-theme-panel {
            background: rgba(255, 255, 255, 0.9);
            backdrop-filter: blur(10px);
            -webkit-backdrop-filter: blur(10px);
            border: 1px solid rgba(0, 0, 0, 0.12);
            border-radius: 8px;
            margin: 0 16px 16px 16px;
            padding: 16px;
            display: flex;
            flex-wrap: wrap;
            gap: 20px;
            align-items: flex-end;
        }
        .custom-theme-section {
            display: flex;
            flex-direction: column;
            gap: 8px;
            flex: 1;
            min-width: 150px;
        }
        .custom-theme-label {
            font-size: 12px;
            font-weight: 600;
            color: #86868b;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
        .custom-theme-section input[type="color"] {
            width: 100%;
            height: 40px;
            border: 1px solid rgba(0, 0, 0, 0.12);
            border-radius: 6px;
            cursor: pointer;
            padding: 4px;
        }
        .custom-theme-section input[type="range"] {
            width: 100%;
            cursor: pointer;
        }
        .modal-overlay {
            display: none;
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0, 0, 0, 0.5);
            backdrop-filter: blur(4px);
            -webkit-backdrop-filter: blur(4px);
            z-index: 2000;
            align-items: center;
            justify-content: center;
        }
        .modal-overlay.show {
            display: flex;
        }
        .modal-dialog {
            background: #ffffff;
            border-radius: 12px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            max-width: 400px;
            width: 90%;
            overflow: hidden;
            animation: modalSlideIn 0.3s cubic-bezier(0.4, 0, 0.2, 1);
        }
        @keyframes modalSlideIn {
            from {
                opacity: 0;
                transform: scale(0.9) translateY(-20px);
            }
            to {
                opacity: 1;
                transform: scale(1) translateY(0);
            }
        }
        .modal-header {
            padding: 20px 24px 16px;
            border-bottom: 1px solid rgba(0, 0, 0, 0.08);
        }
        .modal-title {
            font-size: 18px;
            font-weight: 600;
            color: #1d1d1f;
            margin: 0;
        }
        .modal-body {
            padding: 20px 24px;
            color: #86868b;
            font-size: 15px;
            line-height: 1.5;
        }
        .modal-footer {
            padding: 16px 24px 20px;
            display: flex;
            gap: 8px;
            justify-content: flex-end;
        }
        .modal-btn {
            padding: 8px 16px;
            border-radius: 8px;
            font-size: 14px;
            font-weight: 600;
            cursor: pointer;
            border: none;
            transition: all 0.15s ease;
        }
        .modal-btn-primary {
            background: #007aff;
            color: white;
        }
        .modal-btn-primary:hover {
            background: #0051d5;
        }
        .modal-btn-secondary {
            background: rgba(0, 0, 0, 0.05);
            color: #1d1d1f;
        }
        .modal-btn-secondary:hover {
            background: rgba(0, 0, 0, 0.1);
        }
        .modal-btn-danger {
            background: #ff3b30;
            color: white;
        }
        .modal-btn-danger:hover {
            background: #d70015;
        }
        .toast-container {
            position: fixed;
            top: 20px;
            right: 20px;
            z-index: 3000;
        }
        .toast {
            background: #ffffff;
            border-radius: 10px;
            box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
            padding: 16px 20px;
            margin-bottom: 10px;
            display: flex;
            align-items: center;
            gap: 12px;
            min-width: 280px;
            animation: toastSlideIn 0.3s cubic-bezier(0.4, 0, 0.2, 1);
        }
        @keyframes toastSlideIn {
            from {
                opacity: 0;
                transform: translateX(400px);
            }
            to {
                opacity: 1;
                transform: translateX(0);
            }
        }
        .toast-icon {
            width: 24px;
            height: 24px;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 14px;
            flex-shrink: 0;
        }
        .toast-success .toast-icon {
            background: #34c759;
            color: white;
        }
        .toast-error .toast-icon {
            background: #ff3b30;
            color: white;
        }
        .toast-info .toast-icon {
            background: #007aff;
            color: white;
        }
        .toast-message {
            color: #1d1d1f;
            font-size: 14px;
            font-weight: 500;
        }
    </style>
</head>
<body>
    <div class="app-container">
        <div class="main-content">
            <!-- 日志显示区域 -->
            <div class="log-section">
                <div class="log-header">
                    <div class="log-title">
                        <i class="bi bi-terminal"></i>
                        实时日志
                        <span class="badge bg-secondary ms-2" id="displayModeBadge">文本模式</span>
                        <span class="badge bg-success ms-2" id="timestampBadge" style="display:none;">时间戳已显示</span>
                    </div>
                    <div class="log-controls">
                        <select id="themeSelect" class="form-select" style="width: 150px;" onchange="setLogTheme(this.value)">
                            <option value="light" selected>浅色</option>
                            <option value="default">默认</option>
                            <option value="dracula">Dracula</option>
                            <option value="nord">Nord</option>
                            <option value="monokai">Monokai</option>
                            <option value="solarized-dark">Solarized Dark</option>
                            <option value="github-dark">GitHub Dark</option>
                            <option value="one-dark">One Dark</option>
                            <option value="material">Material</option>
                            <option value="custom">自定义</option>
                        </select>
                        <button class="btn btn-sm btn-outline-success" onclick="toggleTimestamp()" id="timestampToggleBtn">
                            <i class="bi bi-clock"></i> 时间
                        </button>
                        <button class="btn btn-sm btn-outline-primary" onclick="toggleDisplayMode()" id="modeToggleBtn">
                            <i class="bi bi-code-square"></i> HEX
                        </button>
                        <button class="btn btn-sm btn-outline-secondary" onclick="clearLogDisplay()">
                            <i class="bi bi-x-circle"></i> 清空
                        </button>
                        <button class="btn btn-sm btn-outline-secondary" onclick="scrollToBottom()">
                            <i class="bi bi-arrow-down"></i> 底部
                        </button>
                    </div>
                </div>
                <div id="logDisplay" class="log-display"></div>
            </div>

            <!-- 自定义主题面板 -->
            <div class="custom-theme-panel" id="customThemePanel" style="display: none;">
                <div class="custom-theme-section">
                    <label class="custom-theme-label">背景颜色</label>
                    <input type="color" id="customBgColor" value="#ffffff" onchange="updateCustomTheme()">
                </div>
                <div class="custom-theme-section">
                    <label class="custom-theme-label">文字颜色</label>
                    <input type="color" id="customTextColor" value="#1d1d1f" onchange="updateCustomTheme()">
                </div>
                <div class="custom-theme-section">
                    <label class="custom-theme-label">时间戳颜色</label>
                    <input type="color" id="customTimestampColor" value="#007aff" onchange="updateCustomTheme()">
                </div>
                <div class="custom-theme-section">
                    <label class="custom-theme-label"><span id="fontSizeLabel">字体大小: </span><span id="fontSizeValue">13</span>px</label>
                    <input type="range" id="customFontSize" min="10" max="24" value="13" oninput="updateCustomTheme(); document.getElementById('fontSizeValue').textContent = this.value;">
                </div>
                <div class="custom-theme-section" style="flex: 0;">
                    <label class="custom-theme-label">&nbsp;</label>
                    <button class="btn btn-sm btn-primary" onclick="hideCustomThemePanel()">
                        <i class="bi bi-check"></i> 完成
                    </button>
                </div>
            </div>

            <!-- 状态栏 -->
            <div class="status-bar">
                <div class="status-buttons">
                    <button id="disconnectBtn" class="btn btn-danger d-none" onclick="disconnect()">
                        <i class="bi bi-plug-fill"></i> 断开
                    </button>
                    <button id="connectBtn" class="btn btn-success" onclick="connect()">
                        <i class="bi bi-plug"></i> 连接
                    </button>
                    <button class="btn btn-dark" onclick="toggleSidebar()">
                        <i class="bi bi-gear"></i> 配置
                    </button>
                    <button class="btn btn-outline-primary" onclick="toggleFilesSection()" id="filesToggleBtn">
                        <i class="bi bi-folder" id="filesToggleIcon"></i> 文件
                    </button>
                    <button class="btn btn-outline-secondary" onclick="toggleLanguage()" id="langToggleBtn">
                        EN
                    </button>
                </div>
                <div>
                    <span id="statusIndicator" class="status-indicator status-disconnected"></span>
                    <span id="statusText" style="color: #1d1d1f; font-size: 13px;">未连接</span>
                </div>
                <div class="status-info">
                    <span id="connectionDetails">端口: - | 波特率: -</span>
                </div>
            </div>

            <!-- 文件列表 -->
            <div class="files-section" id="filesSection">
                <div class="files-header">
                    <div class="files-title">
                        <i class="bi bi-folder"></i> 历史日志文件
                    </div>
                </div>
                <table class="files-table">
                    <thead>
                        <tr>
                            <th>文件名</th>
                            <th>大小</th>
                            <th>修改时间</th>
                            <th>操作</th>
                        </tr>
                    </thead>
                    <tbody id="logsTableBody">
                        <!-- 动态生成 -->
                    </tbody>
                </table>
            </div>
        </div>

        <!-- 侧边栏配置面板 -->
        <div class="sidebar" id="sidebar">
            <div class="sidebar-header">
                <div class="sidebar-title">
                    <i class="bi bi-gear"></i> 串口配置
                </div>
                <button class="btn btn-sm btn-dark" onclick="toggleSidebar()">
                    <i class="bi bi-x"></i>
                </button>
            </div>
            <div class="sidebar-content">
                <div class="config-group">
                    <div class="config-label">COM口</div>
                    <select id="portName" class="form-select w-100">
                        <option value="">选择端口...</option>
                    </select>
                </div>
                <div class="config-group">
                    <div class="config-label">波特率</div>
                    <select id="baudRate" class="form-select w-100">
                        <option value="9600">9600</option>
                        <option value="19200">19200</option>
                        <option value="38400">38400</option>
                        <option value="57600">57600</option>
                        <option value="115200" selected>115200</option>
                    </select>
                </div>
                <div class="config-group">
                    <div class="config-label">数据位</div>
                    <select id="dataBits" class="form-select w-100">
                        <option value="5">5</option>
                        <option value="6">6</option>
                        <option value="7">7</option>
                        <option value="8" selected>8</option>
                    </select>
                </div>
                <div class="config-group">
                    <div class="config-label">校验位</div>
                    <select id="parity" class="form-select w-100">
                        <option value="None" selected>None</option>
                        <option value="Odd">Odd</option>
                        <option value="Even">Even</option>
                    </select>
                </div>
                <div class="config-group">
                    <div class="config-label">停止位</div>
                    <select id="stopBits" class="form-select w-100">
                        <option value="1" selected>1</option>
                        <option value="2">2</option>
                    </select>
                </div>
                <div class="config-group">
                    <div class="config-label">流控信号</div>
                    <div class="form-check">
                        <input class="form-check-input" type="checkbox" id="rts">
                        <label class="form-check-label" for="rts">RTS (Request to Send)</label>
                    </div>
                    <div class="form-check">
                        <input class="form-check-input" type="checkbox" id="dtr">
                        <label class="form-check-label" for="dtr">DTR (Data Terminal Ready)</label>
                    </div>
                </div>
            </div>
            <div class="sidebar-footer">
                <button class="btn btn-primary w-100 mb-2" onclick="refreshPorts()">
                    <i class="bi bi-arrow-clockwise"></i> 刷新端口
                </button>
            </div>
        </div>
    </div>

    <!-- 自定义模态对话框 -->
    <div class="modal-overlay" id="modalOverlay">
        <div class="modal-dialog">
            <div class="modal-header">
                <h3 class="modal-title" id="modalTitle">标题</h3>
            </div>
            <div class="modal-body" id="modalBody">
                内容
            </div>
            <div class="modal-footer" id="modalFooter">
                <button class="modal-btn modal-btn-secondary" onclick="closeModal()">取消</button>
                <button class="modal-btn modal-btn-primary" id="modalConfirmBtn">确定</button>
            </div>
        </div>
    </div>

    <!-- Toast通知容器 -->
    <div class="toast-container" id="toastContainer"></div>

    <script>
        // SSE连接
        let eventSource;
        let logDisplay = document.getElementById('logDisplay');
        let autoScroll = true;
        let hexMode = false; // HEX显示模式
        let showTimestamp = false; // 是否显示时间戳
        let receiveBuffer = ''; // 接收缓冲区，用于合并分片数据

        // 侧边栏切换
        function toggleSidebar() {
            const sidebar = document.getElementById('sidebar');
            sidebar.classList.toggle('show');
        }

        // 自定义对话框和通知
        function showModal(title, message, onConfirm, showCancel = true, isDanger = false) {
            return new Promise((resolve) => {
                const overlay = document.getElementById('modalOverlay');
                const titleEl = document.getElementById('modalTitle');
                const bodyEl = document.getElementById('modalBody');
                const footerEl = document.getElementById('modalFooter');
                const confirmBtn = document.getElementById('modalConfirmBtn');

                titleEl.textContent = title;
                bodyEl.textContent = message;

                // 设置按钮样式
                confirmBtn.className = 'modal-btn';
                if (isDanger) {
                    confirmBtn.classList.add('modal-btn-danger');
                } else {
                    confirmBtn.classList.add('modal-btn-primary');
                }

                // 设置是否显示取消按钮
                if (showCancel) {
                    const btnClass = isDanger ? 'modal-btn-danger' : 'modal-btn-primary';
                    footerEl.innerHTML = '<button class="modal-btn modal-btn-secondary" id="modalCancelBtn">取消</button>' +
                        '<button class="modal-btn ' + btnClass + '" id="modalConfirmBtn">确定</button>';
                } else {
                    const btnClass = isDanger ? 'modal-btn-danger' : 'modal-btn-primary';
                    footerEl.innerHTML = '<button class="modal-btn ' + btnClass + '" id="modalConfirmBtn">确定</button>';
                }

                overlay.classList.add('show');

                const handleConfirm = () => {
                    closeModal();
                    if (onConfirm) onConfirm();
                    resolve(true);
                };

                const handleCancel = () => {
                    closeModal();
                    resolve(false);
                };

                document.getElementById('modalConfirmBtn').onclick = handleConfirm;

                const cancelBtn = document.getElementById('modalCancelBtn');
                if (cancelBtn) {
                    cancelBtn.onclick = handleCancel;
                }

                // 点击背景关闭
                overlay.onclick = (e) => {
                    if (e.target === overlay) {
                        handleCancel();
                    }
                };
            });
        }

        function closeModal() {
            const overlay = document.getElementById('modalOverlay');
            overlay.classList.remove('show');
        }

        function showToast(message, type = 'info') {
            const container = document.getElementById('toastContainer');
            const toast = document.createElement('div');
            toast.className = 'toast toast-' + type;

            let icon = 'info';
            if (type === 'success') icon = 'check';
            if (type === 'error') icon = 'x';

            toast.innerHTML = '<div class="toast-icon"><i class="bi bi-' + icon + '"></i></div>' +
                '<div class="toast-message">' + message + '</div>';

            container.appendChild(toast);

            // 3秒后自动移除
            setTimeout(function() {
                toast.style.opacity = '0';
                toast.style.transform = 'translateX(400px)';
                setTimeout(function() {
                    if (container.contains(toast)) {
                        container.removeChild(toast);
                    }
                }, 300);
            }, 3000);
        }

        // 文件列表切换
        function toggleFilesSection() {
            const filesSection = document.getElementById('filesSection');
            const icon = document.getElementById('filesToggleIcon');
            const btn = document.getElementById('filesToggleBtn');

            if (filesSection.classList.contains('hidden')) {
                // 展开文件列表
                filesSection.classList.remove('hidden');
                filesSection.style.maxHeight = '250px';
                icon.className = 'bi bi-folder';
                const filesText = currentLanguage === 'zh' ? '文件' : 'Files';
                btn.innerHTML = '<i class="bi bi-folder-open" id="filesToggleIcon"></i> ' + filesText;
            } else {
                // 隐藏文件列表
                filesSection.classList.add('hidden');
                filesSection.style.maxHeight = '0';
                icon.className = 'bi bi-folder';
                const filesText = currentLanguage === 'zh' ? '文件' : 'Files';
                btn.innerHTML = '<i class="bi bi-folder" id="filesToggleIcon"></i> ' + filesText;
            }
        }

        // 语言切换
        let currentLanguage = 'zh'; // 默认中文

        const translations = {
            zh: {
                connect: '连接',
                disconnect: '断开',
                config: '配置',
                files: '文件',
                notConnected: '未连接',
                connected: '已连接',
                realTimeLog: '实时日志',
                textMode: '文本模式',
                hexMode: 'HEX模式',
                timestampShown: '时间戳已显示',
                time: '时间',
                hex: 'HEX',
                clear: '清空',
                bottom: '底部',
                lightTheme: '浅色',
                defaultTheme: '默认',
                customTheme: '自定义',
                port: '端口',
                baudRate: '波特率',
                logFile: '日志',
                theme: '主题',
                done: '完成',
                bgColor: '背景颜色',
                textColor: '文字颜色',
                timestampColor: '时间戳颜色',
                fontSize: '字体大小',
                historyLogFiles: '历史日志文件',
                fileName: '文件名',
                size: '大小',
                modifyTime: '修改时间',
                operation: '操作',
                view: '查看',
                download: '下载',
                delete: '删除',
                refreshPorts: '刷新端口',
                close: '关闭',
                serialConfig: '串口配置',
                comPort: 'COM口',
                dataBits: '数据位',
                parity: '校验位',
                stopBits: '停止位',
                flowControl: '流控信号'
            },
            en: {
                connect: 'Connect',
                disconnect: 'Disconnect',
                config: 'Config',
                files: 'Files',
                notConnected: 'Disconnected',
                connected: 'Connected',
                realTimeLog: 'Real-time Log',
                textMode: 'Text Mode',
                hexMode: 'HEX Mode',
                timestampShown: 'Timestamp Shown',
                time: 'Time',
                hex: 'HEX',
                clear: 'Clear',
                bottom: 'Bottom',
                lightTheme: 'Light',
                defaultTheme: 'Default',
                customTheme: 'Custom',
                port: 'Port',
                baudRate: 'Baud Rate',
                logFile: 'Log',
                theme: 'Theme',
                done: 'Done',
                bgColor: 'Background',
                textColor: 'Text Color',
                timestampColor: 'Timestamp Color',
                fontSize: 'Font Size',
                historyLogFiles: 'History Log Files',
                fileName: 'File Name',
                size: 'Size',
                modifyTime: 'Modified',
                operation: 'Actions',
                view: 'View',
                download: 'Download',
                delete: 'Delete',
                refreshPorts: 'Refresh Ports',
                close: 'Close',
                serialConfig: 'Serial Config',
                comPort: 'COM Port',
                dataBits: 'Data Bits',
                parity: 'Parity',
                stopBits: 'Stop Bits',
                flowControl: 'Flow Control'
            }
        };

        function t(key) {
            return translations[currentLanguage][key] || key;
        }

        function setLanguage(lang) {
            currentLanguage = lang;
            setCookie('language', lang, 365);
            updateLanguage();
        }

        function toggleLanguage() {
            const newLang = currentLanguage === 'zh' ? 'en' : 'zh';
            setLanguage(newLang);
        }

        function updateLanguage() {
            // 更新按钮文本
            document.getElementById('connectBtn').innerHTML = '<i class="bi bi-plug"></i> ' + t('connect');
            document.getElementById('disconnectBtn').innerHTML = '<i class="bi bi-plug-fill"></i> ' + t('disconnect');

            // 更新配置按钮
            const configBtn = document.querySelector('.status-buttons .btn-dark');
            if (configBtn) {
                configBtn.innerHTML = '<i class="bi bi-gear"></i> ' + t('config');
            }

            // 更新文件按钮
            const filesBtn = document.getElementById('filesToggleBtn');
            if (filesBtn) {
                const isFolderOpen = filesBtn.querySelector('.bi-folder-open');
                const icon = isFolderOpen ? 'bi-folder-open' : 'bi-folder';
                filesBtn.innerHTML = '<i class="bi ' + icon + '" id="filesToggleIcon"></i> ' + t('files');
            }

            // 更新状态文本
            const statusText = document.getElementById('statusText');
            if (statusText) {
                statusText.textContent = t('notConnected');
            }

            // 更新日志标题
            const logTitle = document.querySelector('.log-title');
            if (logTitle) {
                logTitle.innerHTML = '<i class="bi bi-terminal"></i> ' + t('realTimeLog') +
                    ' <span class="badge bg-secondary ms-2" id="displayModeBadge">' + t('textMode') + '</span>' +
                    ' <span class="badge bg-success ms-2" id="timestampBadge" style="display:none;">' + t('timestampShown') + '</span>';
            }

            // 更新控制按钮
            const timestampBtn = document.getElementById('timestampToggleBtn');
            if (timestampBtn) {
                timestampBtn.innerHTML = '<i class="bi bi-clock"></i> ' + t('time');
            }

            const modeBtn = document.getElementById('modeToggleBtn');
            if (modeBtn) {
                modeBtn.innerHTML = '<i class="bi bi-code-square"></i> ' + t('hex');
            }

            const clearBtn = document.querySelector('button[onclick="clearLogDisplay()"]');
            if (clearBtn) {
                clearBtn.innerHTML = '<i class="bi bi-x-circle"></i> ' + t('clear');
            }

            const scrollBtn = document.querySelector('button[onclick="scrollToBottom()"]');
            if (scrollBtn) {
                scrollBtn.innerHTML = '<i class="bi bi-arrow-down"></i> ' + t('bottom');
            }

            // 更新历史文件标题
            const filesTitle = document.querySelector('.files-title');
            if (filesTitle) {
                filesTitle.innerHTML = '<i class="bi bi-folder"></i> ' + t('historyLogFiles');
            }

            // 更新侧边栏标题
            const sidebarTitle = document.querySelector('.sidebar-title');
            if (sidebarTitle) {
                sidebarTitle.innerHTML = '<i class="bi bi-gear"></i> ' + t('serialConfig');
            }

            // 更新自定义面板标签
            const labels = document.querySelectorAll('.custom-theme-label');
            if (labels.length > 0) {
                labels[0].textContent = t('bgColor');
                labels[1].textContent = t('textColor');
                labels[2].textContent = t('timestampColor');
                const fontSizeLabel = document.getElementById('fontSizeLabel');
                if (fontSizeLabel) {
                    fontSizeLabel.textContent = t('fontSize') + ': ';
                }
                if (labels[4]) labels[4].innerHTML = '&nbsp;';
            }

            const doneBtn = document.querySelector('.custom-theme-panel .btn-primary');
            if (doneBtn) {
                doneBtn.innerHTML = '<i class="bi bi-check"></i> ' + t('done');
            }

            // 更新表格标题
            const tableHeaders = document.querySelectorAll('.files-table th');
            if (tableHeaders.length >= 4) {
                tableHeaders[0].textContent = t('fileName');
                tableHeaders[1].textContent = t('size');
                tableHeaders[2].textContent = t('modifyTime');
                tableHeaders[3].textContent = t('operation');
            }

            // 更新刷新按钮
            const refreshBtn = document.querySelector('.sidebar-footer .btn-primary');
            if (refreshBtn) {
                refreshBtn.innerHTML = '<i class="bi bi-arrow-clockwise"></i> ' + t('refreshPorts');
            }

            // 更新侧边栏关闭按钮
            const closeBtn = document.querySelector('.sidebar-header .btn-dark');
            if (closeBtn) {
                closeBtn.innerHTML = '<i class="bi bi-x"></i> ' + t('close');
            }

            // 更新语言切换按钮
            const langBtn = document.getElementById('langToggleBtn');
            if (langBtn) {
                langBtn.textContent = currentLanguage === 'zh' ? 'EN' : '中文';
            }

            // 重新更新文件按钮的图标
            setTimeout(() => {
                const filesSection = document.getElementById('filesSection');
                const icon = document.getElementById('filesToggleIcon');
                const btn = document.getElementById('filesToggleBtn');
                if (filesSection && btn && !filesSection.classList.contains('hidden')) {
                    btn.innerHTML = '<i class="bi bi-folder-open" id="filesToggleIcon"></i> ' + t('files');
                } else if (btn) {
                    btn.innerHTML = '<i class="bi bi-folder" id="filesToggleIcon"></i> ' + t('files');
                }
            }, 50);
        }

        function loadLanguage() {
            const savedLang = getCookie('language');
            if (savedLang && (savedLang === 'zh' || savedLang === 'en')) {
                currentLanguage = savedLang;
            }
        }

        // Log主题定义（类似iTerm2）
        const logThemes = {
            'light': {
                name: '浅色',
                background: '#ffffff',
                color: '#1d1d1f',
                timestamp: '#007aff',
                fontSize: 13
            },
            'default': {
                name: '默认',
                background: '#1d1d1f',
                color: '#f5f5f7',
                timestamp: '#0a84ff',
                fontSize: 13
            },
            'dracula': {
                name: 'Dracula',
                background: '#282a36',
                color: '#f8f8f2',
                timestamp: '#bd93f9',
                fontSize: 13
            },
            'nord': {
                name: 'Nord',
                background: '#2e3440',
                color: '#d8dee9',
                timestamp: '#88c0d0',
                fontSize: 13
            },
            'monokai': {
                name: 'Monokai',
                background: '#272822',
                color: '#f8f8f2',
                timestamp: '#a6e22e',
                fontSize: 13
            },
            'solarized-dark': {
                name: 'Solarized Dark',
                background: '#002b36',
                color: '#839496',
                timestamp: '#268bd2',
                fontSize: 13
            },
            'github-dark': {
                name: 'GitHub Dark',
                background: '#0d1117',
                color: '#c9d1d9',
                timestamp: '#58a6ff',
                fontSize: 13
            },
            'one-dark': {
                name: 'One Dark',
                background: '#282c34',
                color: '#abb2bf',
                timestamp: '#61afef',
                fontSize: 13
            },
            'material': {
                name: 'Material',
                background: '#263238',
                color: '#eeffff',
                timestamp: '#80cbc4',
                fontSize: 13
            },
            'custom': {
                name: '自定义',
                background: '#ffffff',
                color: '#1d1d1f',
                timestamp: '#007aff',
                fontSize: 13
            }
        };

        let currentTheme = 'light';
        let customTheme = {
            background: '#ffffff',
            color: '#1d1d1f',
            timestamp: '#007aff',
            fontSize: 13
        };

        function setLogTheme(themeName) {
            const theme = logThemes[themeName];
            if (!theme) return;

            currentTheme = themeName;
            const logDisplay = document.getElementById('logDisplay');
            logDisplay.style.backgroundColor = theme.background;
            logDisplay.style.color = theme.color;
            logDisplay.style.fontSize = theme.fontSize + 'px';

            // 更新时间戳颜色
            const style = document.getElementById('theme-override') || document.createElement('style');
            style.id = 'theme-override';
            style.textContent = '.log-timestamp { color: ' + theme.timestamp + ' !important; }';
            if (!document.getElementById('theme-override')) {
                document.head.appendChild(style);
            }

            // 如果是自定义主题，加载自定义设置
            if (themeName === 'custom') {
                loadCustomTheme();
            }

            // 保存到cookie
            setCookie('logTheme', themeName, 365);

            // 更新主题选择器显示
            const themeSelect = document.getElementById('themeSelect');
            if (themeSelect) {
                themeSelect.value = themeName;
            }

            // 显示/隐藏自定义面板 - 立即执行
            const customPanel = document.getElementById('customThemePanel');
            if (customPanel) {
                if (themeName === 'custom') {
                    customPanel.style.display = 'flex';
                } else {
                    customPanel.style.display = 'none';
                }
            }
        }

        function updateCustomTheme() {
            customTheme.background = document.getElementById('customBgColor').value;
            customTheme.color = document.getElementById('customTextColor').value;
            customTheme.timestamp = document.getElementById('customTimestampColor').value;
            customTheme.fontSize = parseInt(document.getElementById('customFontSize').value);

            // 应用自定义主题
            const logDisplay = document.getElementById('logDisplay');
            logDisplay.style.backgroundColor = customTheme.background;
            logDisplay.style.color = customTheme.color;
            logDisplay.style.fontSize = customTheme.fontSize + 'px';

            // 更新字体大小显示
            const fontSizeValue = document.getElementById('fontSizeValue');
            if (fontSizeValue) {
                fontSizeValue.textContent = customTheme.fontSize;
            }

            // 更新时间戳颜色
            const style = document.getElementById('theme-override');
            style.textContent = '.log-timestamp { color: ' + customTheme.timestamp + ' !important; }';

            // 保存到cookie
            setCookie('customTheme', JSON.stringify(customTheme), 365);
        }

        function loadCustomTheme() {
            const saved = getCookie('customTheme');
            if (saved) {
                try {
                    customTheme = JSON.parse(saved);
                    document.getElementById('customBgColor').value = customTheme.background;
                    document.getElementById('customTextColor').value = customTheme.color;
                    document.getElementById('customTimestampColor').value = customTheme.timestamp;
                    document.getElementById('customFontSize').value = customTheme.fontSize;
                    const fontSizeValue = document.getElementById('fontSizeValue');
                    if (fontSizeValue) {
                        fontSizeValue.textContent = customTheme.fontSize;
                    }
                } catch (e) {
                    console.error('加载自定义主题失败:', e);
                }
            }
            updateCustomTheme();
        }

        function hideCustomThemePanel() {
            const customPanel = document.getElementById('customThemePanel');
            if (customPanel) {
                customPanel.style.display = 'none';
            }
        }

        function loadLogTheme() {
            const savedTheme = getCookie('logTheme');
            if (savedTheme && logThemes[savedTheme]) {
                setLogTheme(savedTheme);
            } else {
                // 默认使用浅色主题
                setLogTheme('light');
            }

            // 确保自定义面板的显示状态正确
            setTimeout(() => {
                const customPanel = document.getElementById('customThemePanel');
                const themeSelect = document.getElementById('themeSelect');
                if (customPanel && themeSelect) {
                    if (themeSelect.value !== 'custom') {
                        customPanel.style.display = 'none';
                    }
                }
            }, 100);
        }

        // Cookie管理函数
        function setCookie(name, value, days) {
            const expires = new Date();
            expires.setTime(expires.getTime() + (days * 24 * 60 * 60 * 1000));
            document.cookie = name + '=' + value + ';expires=' + expires.toUTCString() + ';path=/';
        }

        function getCookie(name) {
            const nameEQ = name + '=';
            const cookies = document.cookie.split(';');
            for (let i = 0; i < cookies.length; i++) {
                let cookie = cookies[i];
                while (cookie.charAt(0) === ' ') {
                    cookie = cookie.substring(1, cookie.length);
                }
                if (cookie.indexOf(nameEQ) === 0) {
                    return cookie.substring(nameEQ.length, cookie.length);
                }
            }
            return null;
        }

        function loadConfig() {
            // 从cookie加载配置
            const savedHexMode = getCookie('hexMode');
            const savedTimestamp = getCookie('showTimestamp');

            if (savedHexMode === 'true') {
                hexMode = true;
                const badge = document.getElementById('displayModeBadge');
                const btn = document.getElementById('modeToggleBtn');
                badge.textContent = 'HEX模式';
                badge.className = 'badge bg-primary ms-2';
                btn.innerHTML = '<i class="bi bi-type"></i> 文本模式';

                // 通知后端更新HEX模式
                updateBackendHexMode(true);
            }

            if (savedTimestamp === 'true') {
                showTimestamp = true;
                const btn = document.getElementById('timestampToggleBtn');
                const badge = document.getElementById('timestampBadge');
                btn.innerHTML = '<i class="bi bi-clock-history"></i> 隐藏时间';
                btn.className = 'btn btn-sm btn-warning';
                badge.textContent = '时间戳已显示';
                badge.style.display = 'inline-block';
            }

            // 加载COM口配置
            loadComConfig();
        }

        function saveComConfig(config) {
            // 保存COM口配置到cookie
            setCookie('portName', config.portName, 365);
            setCookie('baudRate', config.baudRate, 365);
            setCookie('dataBits', config.dataBits, 365);
            setCookie('parity', config.parity, 365);
            setCookie('stopBits', config.stopBits, 365);
            setCookie('rts', config.rts, 365);
            setCookie('dtr', config.dtr, 365);
        }

        function loadComConfig() {
            // 从cookie加载COM口配置
            const savedPortName = getCookie('portName');
            const savedBaudRate = getCookie('baudRate');
            const savedDataBits = getCookie('dataBits');
            const savedParity = getCookie('parity');
            const savedStopBits = getCookie('stopBits');
            const savedRts = getCookie('rts');
            const savedDtr = getCookie('dtr');

            if (savedPortName) {
                document.getElementById('portName').value = savedPortName;
            }
            if (savedBaudRate) {
                document.getElementById('baudRate').value = savedBaudRate;
            }
            if (savedDataBits) {
                document.getElementById('dataBits').value = savedDataBits;
            }
            if (savedParity) {
                document.getElementById('parity').value = savedParity;
            }
            if (savedStopBits) {
                document.getElementById('stopBits').value = savedStopBits;
            }
            if (savedRts === 'true') {
                document.getElementById('rts').checked = true;
            }
            if (savedDtr === 'true') {
                document.getElementById('dtr').checked = true;
            }
        }

        function toggleDisplayMode() {
            hexMode = !hexMode;
            const badge = document.getElementById('displayModeBadge');
            const btn = document.getElementById('modeToggleBtn');

            if (hexMode) {
                badge.textContent = 'HEX模式';
                badge.className = 'badge bg-primary ms-2';
                btn.innerHTML = '<i class="bi bi-type"></i> 文本模式';
            } else {
                badge.textContent = '文本模式';
                badge.className = 'badge bg-secondary ms-2';
                btn.innerHTML = '<i class="bi bi-code-square"></i> HEX模式';
            }

            // 保存到cookie (365天)
            setCookie('hexMode', hexMode, 365);

            // 通知后端更新HEX模式
            updateBackendHexMode(hexMode);
        }

        async function updateBackendHexMode(mode) {
            try {
                const response = await fetch('/api/hexmode', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ hexMode: mode })
                });
                const result = await response.json();
                if (result.success) {
                    console.log('后端HEX模式已更新:', result.hexMode);
                }
            } catch (error) {
                console.error('更新后端HEX模式失败:', error);
            }
        }

        function formatHex(message) {
            let hexResult = '';
            for (let i = 0; i < message.length; i++) {
                const charCode = message.charCodeAt(i);
                const hex = charCode.toString(16).toUpperCase().padStart(2, '0');
                hexResult += hex + ' ';
                if ((i + 1) % 16 === 0) {
                    hexResult += '\n';
                }
            }
            return hexResult.trim();
        }

        function toggleTimestamp() {
            showTimestamp = !showTimestamp;
            const btn = document.getElementById('timestampToggleBtn');
            const badge = document.getElementById('timestampBadge');

            if (showTimestamp) {
                btn.innerHTML = '<i class="bi bi-clock-history"></i> 隐藏时间';
                btn.className = 'btn btn-sm btn-warning';
                badge.textContent = '时间戳已显示';
                badge.style.display = 'inline-block';
            } else {
                btn.innerHTML = '<i class="bi bi-clock"></i> 显示时间';
                btn.className = 'btn btn-sm btn-success';
                badge.style.display = 'none';
            }

            // 保存到cookie (365天)
            setCookie('showTimestamp', showTimestamp, 365);
        }

        function connectEventSource() {
            eventSource = new EventSource('/ws');

            eventSource.onmessage = function(event) {
                try {
                    const data = JSON.parse(event.data);
                    if (data.type === 'log') {
                        // 解码Base64数据
                        const decodedContent = atob(data.content);
                        const decodedHex = atob(data.hex);

                        // 调试：显示接收到的数据长度
                        if (decodedContent.length > 50) {
                            console.log('📥 收到SSE消息，长度:', decodedContent.length, '预览:', decodedContent.substring(0, 50) + '...');
                        }

                        // 后端可能发送不完整的行（被SSE分片），需要缓冲
                        appendLog(decodedContent, decodedHex);
                    } else if (data.type === 'info') {
                        console.log('Info:', data.content);
                    }
                } catch (e) {
                    console.error('解析消息失败:', e, event.data);
                }
            };

            eventSource.onopen = function() {
                console.log('SSE连接已建立');
            };

            eventSource.onerror = function(error) {
                console.error('SSE连接错误:', error);
                // SSE会自动重连，不需要手动处理
            };
        }

        function disconnectEventSource() {
            if (eventSource) {
                eventSource.close();
                eventSource = null;
            }
        }

        function appendLog(message, hexMessage) {
            const now = new Date();
            const timestamp = now.toLocaleTimeString('zh-CN', { hour12: false });

            // 将SSE消息添加到缓冲区（SSE可能会把长行分片发送）
            receiveBuffer += message;

            // 统一换行符
            receiveBuffer = receiveBuffer.replace(/\r\n/g, '\n').replace(/\r/g, '\n');

            // 按换行符分割，保留最后一个不完整的行
            const lines = receiveBuffer.split('\n');
            receiveBuffer = lines.pop() || ''; // 保留最后可能不完整的行

            // 显示所有完整的行
            for (let i = 0; i < lines.length; i++) {
                let line = lines[i];
                if (line === '') continue; // 跳过空行

                // 移除ANSI转义序列
                line = removeAnsiCodes(line);

                const logEntry = document.createElement('div');
                logEntry.className = 'log-entry';

                let displayContent = line;
                if (hexMode && hexMessage) {
                    // HEX模式下，对整行进行处理
                    displayContent = formatHexLine(line);
                }

                // 根据showTimestamp决定是否显示时间戳
                if (showTimestamp) {
                    const timestampSpan = document.createElement('span');
                    timestampSpan.className = 'log-timestamp';
                    timestampSpan.textContent = '[' + timestamp + '] ';
                    logEntry.appendChild(timestampSpan);

                    const contentSpan = document.createElement('span');
                    contentSpan.textContent = displayContent;
                    logEntry.appendChild(contentSpan);
                } else {
                    logEntry.textContent = displayContent;
                }

                logDisplay.appendChild(logEntry);
            }

            if (autoScroll) {
                logDisplay.scrollTop = logDisplay.scrollHeight;
            }
        }

        // HEX模式下的单行格式化
        function formatHexLine(line) {
            let hexResult = '';
            for (let i = 0; i < line.length; i++) {
                const charCode = line.charCodeAt(i);
                const hex = charCode.toString(16).toUpperCase().padStart(2, '0');
                hexResult += hex + ' ';
                if ((i + 1) % 16 === 0) {
                    hexResult += '\n';
                }
            }
            return hexResult.trim();
        }

        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }

        // 移除ANSI转义序列（ESP32日志中的颜色代码）
        function removeAnsiCodes(text) {
            // 移除ANSI转义序列，包括ESC[...m格式的颜色代码
            // ESC字符的ASCII码是27，用\x1b表示
            // 模式匹配：ESC后跟[，然后是数字和分号，最后以m结尾
            return text.replace(/\x1b\[[0-9;]*[a-zA-Z]/g, '')
                       .replace(/\x1b\[[0-9;]*m/g, '');  // 额外的保险，处理单独的m
        }

        function clearLogDisplay() {
            logDisplay.innerHTML = '';
            receiveBuffer = ''; // 清空缓冲区
        }

        function scrollToBottom() {
            logDisplay.scrollTop = logDisplay.scrollHeight;
            autoScroll = true;
        }

        logDisplay.addEventListener('scroll', function() {
            autoScroll = logDisplay.scrollHeight - logDisplay.scrollTop <= logDisplay.clientHeight + 50;
        });

        // 初始化
        window.onload = function() {
            // 首先加载保存的配置
            loadConfig();

            // 加载语言设置
            loadLanguage();
            updateLanguage();

            // 加载log主题
            loadLogTheme();

            refreshPorts();
            updateStatus();
            loadLogFiles();
            connectEventSource();

            // 定时更新状态
            setInterval(updateStatus, 2000);
            setInterval(loadLogFiles, 10000);
        };

        async function refreshPorts() {
            try {
                const response = await fetch('/api/ports');
                const ports = await response.json();
                const select = document.getElementById('portName');
                const currentValue = select.value || getCookie('portName');

                select.innerHTML = '<option value="">选择端口...</option>';
                ports.forEach(port => {
                    const option = document.createElement('option');
                    option.value = port;
                    option.textContent = port;
                    select.appendChild(option);
                });

                if (currentValue) {
                    select.value = currentValue;
                }
            } catch (error) {
                console.error('刷新端口失败:', error);
            }
        }

        async function updateStatus() {
            try {
                const response = await fetch('/api/status');
                const status = await response.json();

                const indicator = document.getElementById('statusIndicator');
                const statusText = document.getElementById('statusText');
                const connectBtn = document.getElementById('connectBtn');
                const disconnectBtn = document.getElementById('disconnectBtn');
                const details = document.getElementById('connectionDetails');

                if (status.connected) {
                    indicator.className = 'status-indicator status-connected';
                    statusText.textContent = t('connected');
                    connectBtn.classList.add('d-none');
                    disconnectBtn.classList.remove('d-none');
                    details.innerHTML = '<small>' + t('port') + ': ' + status.portName + ' | ' + t('baudRate') + ': ' + document.getElementById('baudRate').value + ' | ' + t('logFile') + ': ' + status.logFile + '</small>';
                } else {
                    indicator.className = 'status-indicator status-disconnected';
                    statusText.textContent = t('notConnected');
                    connectBtn.classList.remove('d-none');
                    disconnectBtn.classList.add('d-none');
                    details.innerHTML = '<small>' + t('port') + ': - | ' + t('baudRate') + ': - | ' + t('logFile') + ': -</small>';
                }
            } catch (error) {
                console.error('更新状态失败:', error);
            }
        }

        async function connect() {
            const config = {
                portName: document.getElementById('portName').value,
                baudRate: parseInt(document.getElementById('baudRate').value),
                dataBits: parseInt(document.getElementById('dataBits').value),
                parity: document.getElementById('parity').value,
                stopBits: parseInt(document.getElementById('stopBits').value),
                rts: document.getElementById('rts').checked,
                dtr: document.getElementById('dtr').checked
            };

            if (!config.portName) {
                const title = currentLanguage === 'zh' ? '提示' : 'Notice';
                const msg = currentLanguage === 'zh' ? '请选择COM口' : 'Please select a COM port';
                await showModal(title, msg);
                return;
            }

            try {
                const response = await fetch('/api/connect', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(config)
                });

                const result = await response.json();
                if (result.success) {
                    const title = currentLanguage === 'zh' ? '连接成功' : 'Connected';
                    const msg = currentLanguage === 'zh' ? '串口连接成功！' : 'Serial port connected successfully!';
                    await showModal(title, msg, null, false);
                    showToast(msg, 'success');

                    // 保存COM口配置到cookie
                    saveComConfig(config);

                    updateStatus();
                    loadLogFiles();
                } else {
                    const title = currentLanguage === 'zh' ? '连接失败' : 'Connection Failed';
                    const msg = currentLanguage === 'zh' ? '连接失败: ' : 'Connection failed: ' + result.error;
                    await showModal(title, msg, null, false);
                    showToast(result.error, 'error');
                }
            } catch (error) {
                const title = currentLanguage === 'zh' ? '连接失败' : 'Connection Failed';
                const msg = currentLanguage === 'zh' ? '连接失败: ' : 'Connection failed: ' + error.message;
                await showModal(title, msg, null, false);
                showToast(error.message, 'error');
            }
        }

        async function disconnect() {
            const title = currentLanguage === 'zh' ? '确认断开' : 'Confirm Disconnect';
            const msg = currentLanguage === 'zh' ? '确定要断开串口连接吗？' : 'Are you sure you want to disconnect?';

            const confirmed = await showModal(title, msg, null, true, true);
            if (!confirmed) return;

            try {
                const response = await fetch('/api/disconnect', { method: 'POST' });
                const result = await response.json();
                if (result.success) {
                    const successMsg = currentLanguage === 'zh' ? '已断开连接' : 'Disconnected';
                    showToast(successMsg, 'success');
                    updateStatus();
                }
            } catch (error) {
                const errorMsg = currentLanguage === 'zh' ? '断开连接失败: ' : 'Disconnect failed: ' + error.message;
                showToast(errorMsg, 'error');
            }
        }

        async function loadLogFiles() {
            try {
                const response = await fetch('/api/logs/logs');
                const files = await response.json();

                const tbody = document.getElementById('logsTableBody');
                tbody.innerHTML = '';

                files.forEach(file => {
                    const row = document.createElement('tr');
                    row.innerHTML = '<td>' + file.name + '</td>' +
                        '<td class="file-size">' + formatFileSize(file.size) + '</td>' +
                        '<td>' + file.modTime + '</td>' +
                        '<td>' +
                        '<button class="btn btn-sm btn-outline-primary" onclick="viewLog(\'' + file.name + '\')">' +
                        '<i class="bi bi-eye"></i> 查看</button> ' +
                        '<button class="btn btn-sm btn-outline-success" onclick="downloadLog(\'' + file.name + '\')">' +
                        '<i class="bi bi-download"></i> 下载</button> ' +
                        '<button class="btn btn-sm btn-outline-danger" onclick="deleteLog(\'' + file.name + '\')">' +
                        '<i class="bi bi-trash"></i> 删除</button>' +
                        '</td>';
                    tbody.appendChild(row);
                });
            } catch (error) {
                console.error('加载日志文件失败:', error);
            }
        }

        function formatFileSize(bytes) {
            if (bytes < 1024) return bytes + ' B';
            if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB';
            return (bytes / (1024 * 1024)).toFixed(2) + ' MB';
        }

        async function viewLog(filename) {
            try {
                // 根据当前模式决定是否使用hex参数
                const url = hexMode
                    ? '/api/logs/view?file=' + encodeURIComponent(filename) + '&hex=true'
                    : '/api/logs/view?file=' + encodeURIComponent(filename);

                const response = await fetch(url);
                const content = await response.text();

                let title = hexMode ? '日志查看器 - HEX模式' : '日志查看器 - 文本模式';

                // 在新窗口中打开日志查看器
                const win = window.open('', '_blank');
                win.document.write('<!DOCTYPE html><html><head><title>' + title + '</title>' +
                    '<style>body{margin:0;padding:20px;background:#1e1e1e;color:#d4d4d4;font-family:monospace;font-size:13px;}' +
                    'pre{white-space:pre-wrap;word-wrap:break-word;}</style></head><body>' +
                    '<h3>' + title + '</h3><pre>' + escapeHtml(content) + '</pre></body></html>');
                win.document.close();
            } catch (error) {
                alert('查看日志失败: ' + error.message);
            }
        }

        function downloadLog(filename) {
            window.location.href = '/api/logs/download?file=' + encodeURIComponent(filename);
        }

        async function deleteLog(filename) {
            const title = currentLanguage === 'zh' ? '确认删除' : 'Confirm Delete';
            const msg = currentLanguage === 'zh' ? '确定要删除日志文件 "' + filename + '" 吗？' : 'Are you sure you want to delete log file "' + filename + '"?';

            const confirmed = await showModal(title, msg, null, true, true);
            if (!confirmed) return;

            try {
                const response = await fetch('/api/logs/delete?file=' + encodeURIComponent(filename), {
                    method: 'DELETE'
                });
                const result = await response.json();
                if (result.success) {
                    const successMsg = currentLanguage === 'zh' ? '删除成功' : 'Deleted successfully';
                    showToast(successMsg, 'success');
                    loadLogFiles();
                } else {
                    const errorMsg = currentLanguage === 'zh' ? '删除失败: ' : 'Delete failed: ' + result.error;
                    showToast(errorMsg, 'error');
                }
            } catch (error) {
                const errorMsg = currentLanguage === 'zh' ? '删除失败: ' : 'Delete failed: ' + error.message;
                showToast(errorMsg, 'error');
            }
        }
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(htmlContent))
}

// 状态处理器
func statusHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	response := StatusResponse{
		Connected: isConnected,
		PortName:  portConfig.PortName,
	}

	if isConnected {
		response.LogFile = filepath.Base(logFile.Name())
		response.StartTime = time.Now().Format("2006-01-02 15:04:05")

		if stat, err := logFile.Stat(); err == nil {
			response.LogSize = stat.Size()
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 列出可用端口
func listPortsHandler(w http.ResponseWriter, r *http.Request) {
	ports, err := serial.GetPortsList()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ports)
}

// 连接串口
func connectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var config PortConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if isConnected {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "已经连接到串口",
		})
		return
	}

	// 配置串口
	mode := &serial.Mode{
		BaudRate: config.BaudRate,
	}

	switch config.DataBits {
	case 5:
		mode.DataBits = 5
	case 6:
		mode.DataBits = 6
	case 7:
		mode.DataBits = 7
	case 8:
		mode.DataBits = 8
	}

	switch config.Parity {
	case "None":
		mode.Parity = serial.NoParity
	case "Odd":
		mode.Parity = serial.OddParity
	case "Even":
		mode.Parity = serial.EvenParity
	}

	switch config.StopBits {
	case 1:
		mode.StopBits = serial.OneStopBit
	case 2:
		mode.StopBits = serial.TwoStopBits
	}

	// 打开串口
	port, err := serial.Open(config.PortName, mode)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// 设置RTS和DTR信号
	if config.RTS {
		if err := port.SetRTS(true); err != nil {
			log.Printf("⚠️  设置RTS失败: %v", err)
			// 不中断连接，只记录警告
		}
	}
	if config.DTR {
		if err := port.SetDTR(true); err != nil {
			log.Printf("⚠️  设置DTR失败: %v", err)
			// 不中断连接，只记录警告
		}
	}

	// 创建日志文件
	logFileName := fmt.Sprintf("com_%s_%s.log",
		strings.ReplaceAll(config.PortName, ":", "_"),
		time.Now().Format("20060102_150405"))
	logFilePath := filepath.Join("logs", logFileName)

	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		port.Close()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "无法创建日志文件: " + err.Error(),
		})
		return
	}

	// 更新全局状态
	currentPort = port
	portConfig = config
	isConnected = true
	logFile = file
	logWriter = bufio.NewWriter(file)

	// 重新创建停止信号通道
	stopLogging = make(chan struct{})

	// 启动日志记录协程
	go startLogging()

	// 重新扫描日志文件
	scanLogFiles()

	log.Printf("✅ 连接成功: %s @ %d baud", config.PortName, config.BaudRate)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "连接成功",
	})
}

// 断开串口
func disconnectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	mu.Lock()

	if !isConnected {
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "未连接到串口",
		})
		return
	}

	// 先设置isConnected为false，停止日志记录
	isConnected = false
	mu.Unlock()

	// 发送停止信号给日志记录协程
	select {
	case stopLogging <- struct{}{}:
		log.Println("📤 已发送停止信号给日志记录协程")
	default:
		log.Println("⚠️  停止信号通道已满，协程可能已退出")
	}

	// 等待一小段时间让协程退出
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	// 关闭连接
	if logWriter != nil {
		logWriter.Flush()
		logWriter = nil
	}
	if logFile != nil {
		logFile.Close()
		logFile = nil
	}
	if currentPort != nil {
		currentPort.Close()
		currentPort = nil
	}

	log.Printf("🔌 已断开串口连接")

	// 通知所有客户端
	msg := "🔌 串口连接已断开"
	hexMsg := bytesToHex([]byte(msg))
	broadcastToClients(msg, hexMsg)

	// 重新扫描日志文件
	scanLogFiles()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "已断开连接",
	})
}

// 配置处理器（保存/加载配置）
func configHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		mu.Lock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(portConfig)
		mu.Unlock()

	case http.MethodPost:
		var config PortConfig
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		mu.Lock()
		portConfig = config
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "配置已更新",
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// 设置HEX模式处理器
func setHexModeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		HexMode bool `json:"hexMode"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	hexMode = request.HexMode
	mu.Unlock()

	log.Printf("🔄 后端HEX模式已更新: %v", hexMode)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"hexMode": hexMode,
		"message": "HEX模式已更新",
	})
}

// 列出日志文件
func listLogsHandler(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir("logs")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var logFiles []LogFile
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".log") {
			info, _ := file.Info()
			logFile := LogFile{
				Name:    file.Name(),
				Size:    info.Size(),
				ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
			}
			logFiles = append(logFiles, logFile)
		}
	}

	// 按修改时间排序（最新的在前）
	for i := 0; i < len(logFiles); i++ {
		for j := i + 1; j < len(logFiles); j++ {
			timeI, _ := time.Parse("2006-01-02 15:04:05", logFiles[i].ModTime)
			timeJ, _ := time.Parse("2006-01-02 15:04:05", logFiles[j].ModTime)
			if timeI.Before(timeJ) {
				logFiles[i], logFiles[j] = logFiles[j], logFiles[i]
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logFiles)
}

// 查看日志内容
func viewLogHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("file")
	if filename == "" {
		http.Error(w, "文件名不能为空", http.StatusBadRequest)
		return
	}

	// 检查是否需要HEX格式
	hexFormat := r.URL.Query().Get("hex") == "true"

	// 安全检查：确保文件在logs目录下
	filename = filepath.Base(filename)
	filePath := filepath.Join("logs", filename)

	content, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, "无法读取文件", http.StatusNotFound)
		return
	}

	var outputContent string
	if hexFormat {
		// 转换为HEX格式
		outputContent = formatToHex(content)
	} else {
		outputContent = string(content)
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(outputContent))
}

// 格式化字节数组为HEX字符串
func formatToHex(data []byte) string {
	var result string
	for i, b := range data {
		hex := fmt.Sprintf("%02X", b)
		result += hex + " "
		if (i+1)%16 == 0 {
			result += "\n"
		}
	}
	return strings.TrimSpace(result)
}

// 下载日志文件
func downloadLogHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("file")
	if filename == "" {
		http.Error(w, "文件名不能为空", http.StatusBadRequest)
		return
	}

	// 安全检查
	filename = filepath.Base(filename)
	filePath := filepath.Join("logs", filename)

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "无法打开文件", http.StatusNotFound)
		return
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "无法获取文件信息", http.StatusInternalServerError)
		return
	}

	// 设置响应头，强制浏览器下载而不是预览
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// 手动写入文件内容
	http.ServeContent(w, r, filename, fileInfo.ModTime(), file)
}

// 删除日志文件
func deleteLogHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filename := r.URL.Query().Get("file")
	if filename == "" {
		http.Error(w, "文件名不能为空", http.StatusBadRequest)
		return
	}

	// 安全检查
	filename = filepath.Base(filename)
	filePath := filepath.Join("logs", filename)

	if err := os.Remove(filePath); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	scanLogFiles()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "文件已删除",
	})
}

// 当前日志处理器（用于实时更新）
func currentLogHandler(w http.ResponseWriter, r *http.Request) {
	if !isConnected {
		http.Error(w, "未连接", http.StatusNotFound)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	http.ServeFile(w, r, logFile.Name())
}

// WebSocket处理器
func wsHandler(w http.ResponseWriter, r *http.Request) {
	// 简单的WebSocket实现用于实时日志推送
	// 这里使用Server-Sent Events (SSE)更简单
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// 注册客户端
	clientsMu.Lock()
	clients[&w] = true
	clientsMu.Unlock()

	// 确保在函数退出时删除客户端
	defer func() {
		clientsMu.Lock()
		delete(clients, &w)
		clientsMu.Unlock()
	}()

	// 发送连接确认
	fmt.Fprintf(w, "data: {\"type\":\"connected\"}\n\n")
	flusher.Flush()

	// 保持连接并定期发送心跳
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// 通知客户端有新连接
	fmt.Fprintf(w, "data: {\"type\":\"info\",\"content\":\"已连接到日志服务器\"}\n\n")
	flusher.Flush()

	for {
		select {
		case <-ticker.C:
			fmt.Fprintf(w, "data: {\"type\":\"heartbeat\"}\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

// 启动日志记录
func startLogging() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("日志记录协程panic: %v", r)
		}
		log.Printf("📝 日志记录协程已退出")
	}()

	buffer := make([]byte, 4096)

	log.Printf("📝 开始记录日志到: %s", logFile.Name())

	for {
		select {
		case <-stopLogging:
			log.Printf("🛑 收到停止信号，退出日志记录")
			return
		default:
			// 检查连接状态
			mu.Lock()
			connected := isConnected
			port := currentPort
			mu.Unlock()

			if !connected || port == nil {
				log.Printf("⚠️  连接已断开，退出日志记录")
				return
			}

			// 设置读取超时，避免永久阻塞
			n, err := port.Read(buffer)
			if err != nil {
				mu.Lock()
				if isConnected {
					log.Printf("❌ 串口读取错误: %v", err)
				}
				mu.Unlock()
				return
			}

			if n > 0 {
				// 直接读取原始字节并立即发送，不做任何处理
				dataBytes := buffer[:n]
				data := string(dataBytes)

				// 生成HEX
				hexStr := bytesToHex(dataBytes)

				// 根据HEX模式决定写入文件的格式
				mu.Lock()
				currentHexMode := hexMode
				mu.Unlock()

				var logEntry string
				if currentHexMode {
					logEntry = hexStr
				} else {
					timestamp := time.Now().Format("2006-01-02 15:04:05.000")
					logEntry = fmt.Sprintf("[%s] %s", timestamp, data)
				}

				// 写入文件
				mu.Lock()
				if logWriter != nil {
					logWriter.WriteString(logEntry)
					logWriter.Flush()
				}
				mu.Unlock()

				// 打印到控制台
				fmt.Print(data)

				// 直接发送原始数据到前端，让前端处理
				broadcastToClients(data, hexStr)
			}
		}
	}
}

// 广播数据到所有WebSocket客户端
func broadcastToClients(data string, hexData string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	// 检查数据长度
	if len(data) > 100 {
		log.Printf("📤 发送原始数据: %d 字符, 预览: %q...", len(data), data[:min(50, len(data))])
	}

	// 直接发送原始文本，不用JSON（避免JSON转义问题）
	// 使用Base64编码确保二进制安全
	encoded := base64.StdEncoding.EncodeToString([]byte(data))
	encodedHex := base64.StdEncoding.EncodeToString([]byte(hexData))

	message := fmt.Sprintf("data: {\"type\":\"log\",\"content\":\"%s\",\"hex\":\"%s\"}\n\n", encoded, encodedHex)

	for client := range clients {
		fmt.Fprintf(*client, message)
		if flusher, ok := (*client).(http.Flusher); ok {
			flusher.Flush()
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 字节数组转HEX字符串（避免UTF-8编码问题）
func bytesToHex(data []byte) string {
	var result string
	for i, b := range data {
		hex := fmt.Sprintf("%02X", b)
		result += hex + " "
		if (i+1)%16 == 0 {
			result += "\n"
		}
	}
	return strings.TrimSpace(result)
}

// JSON转义函数
func jsonEscape(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

// 扫描日志文件
func scanLogFiles() {
	files, err := os.ReadDir("logs")
	if err != nil {
		return
	}

	logFiles = []LogFile{}
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".log") {
			info, _ := file.Info()
			logFile := LogFile{
				Name:    file.Name(),
				Size:    info.Size(),
				ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
			}
			logFiles = append(logFiles, logFile)
		}
	}

	// 按修改时间排序（最新的在前）
	for i := 0; i < len(logFiles); i++ {
		for j := i + 1; j < len(logFiles); j++ {
			timeI, _ := time.Parse("2006-01-02 15:04:05", logFiles[i].ModTime)
			timeJ, _ := time.Parse("2006-01-02 15:04:05", logFiles[j].ModTime)
			if timeI.Before(timeJ) {
				logFiles[i], logFiles[j] = logFiles[j], logFiles[i]
			}
		}
	}
}