# SerialWebViewer

<div align="center">

![SerialWebViewer](https://img.shields.io/badge/SerialWebViewer-v1.0-blue)
![Go](https://img.shields.io/badge/Go-1.19+-00ADD8?logo=go)
![License](https://img.shields.io/badge/License-MIT-green)
![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey)

**一个现代化的串口日志查看工具**

[功能特性](#功能特性) • [快速开始](#快速开始) • [使用说明](#使用说明) • [技术栈](#技术栈)

</div>

---

## 📖 简介

SerialWebViewer 是一个基于 Web 的串口日志记录和查看工具，支持实时监控串口数据、多主题切换、中英文界面，适用于嵌入式开发、硬件调试等场景。

## ✨ 功能特性

### 🔌 串口管理
- 支持多种波特率（9600-115200）
- 可配置数据位、校验位、停止位
- 支持 RTS/DTR 流控信号
- 自动扫描可用串口
- 配置自动保存

### 📊 日志功能
- 实时日志显示（SSE推送）
- 文本/HEX双模式显示
- 可选时间戳显示
- 日志自动保存到文件
- 历史日志管理（查看、下载、删除）

### 🎨 界面设计
- **现代化 UI**：Apple 风格的白色主题
- **多主题支持**：9种预设主题 + 自定义主题
- **自定义外观**：背景色、文字色、时间戳色、字体大小均可调节
- **响应式布局**：适配不同屏幕尺寸
- **国际化**：完整的中英文界面

### 💾 数据管理
- 自动记录日志到文件
- 文件按时间命名
- 支持在线查看历史日志
- 一键下载日志文件
- 最新文件优先显示

## 🚀 快速开始

### 安装

#### Windows
```bash
# 编译
GOOS=windows GOARCH=amd64 go build -o SerialWebViewer.exe main.go

# 运行
./SerialWebViewer.exe
```

#### Linux
```bash
# 编译
go build -o serialwebviewer main.go

# 运行
./serialwebviewer
```

#### macOS
```bash
# 编译
GOOS=darwin GOARCH=amd64 go build -o SerialWebViewer.mac main.go

# 运行
./SerialWebViewer.mac
```

### 使用

1. 启动程序后，在浏览器中打开 `http://localhost:8088`
2. 点击右上角"配置"按钮打开配置面板
3. 选择串口和波特率，点击"连接"
4. 实时查看串口数据
5. 可切换显示模式、主题和语言

## 📖 使用说明

### 串口配置
| 参数 | 说明 | 可选值 |
|------|------|--------|
| COM口 | 串口设备 | 自动扫描 |
| 波特率 | 数据传输速率 | 9600/19200/38400/57600/115200 |
| 数据位 | 数据位数 | 5/6/7/8 |
| 校验位 | 校验方式 | None/Odd/Even |
| 停止位 | 停止位数 | 1/2 |
| RTS | 请求发送 | 开/关 |
| DTR | 数据终端就绪 | 开/关 |

### 主题列表
- **浅色**（默认）：白底黑字，适合日间使用
- **默认**：深灰底，护眼模式
- **Dracula**：紫色调深色主题
- **Nord**：冷色调北欧风格
- **Monokai**：经典深色主题
- **Solarized Dark**：护眼色调
- **GitHub Dark**：GitHub官方暗色
- **One Dark**：Atom编辑器风格
- **Material**：Material Design风格
- **自定义**：完全自定义外观

### 快捷键
- 点击"配置"：打开/关闭配置面板
- 点击"文件"：展开/收起历史文件列表
- 选择主题：实时切换日志显示主题

## 🛠️ 技术栈

- **后端**：Go 1.19+
- **串口通信**：go.bug.st/serial
- **前端**：原生 HTML/CSS/JavaScript
- **UI框架**：Bootstrap 5
- **图标**：Bootstrap Icons
- **实时通信**：Server-Sent Events (SSE)

## 📁 项目结构

```
SerialWebViewer/
├── main.go           # 主程序文件
├── go.mod            # Go模块依赖
├── go.sum            # 依赖校验文件
├── logs/             # 日志文件目录
├── LICENSE           # MIT许可证
├── README.md         # 项目说明
└── .gitignore        # Git忽略文件
```

## 🔧 开发

### 依赖安装
```bash
go get go.bug.st/serial
```

### 编译
```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o SerialWebViewer.exe main.go

# Linux
go build -o serialwebviewer main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o SerialWebViewer.mac main.go
```

## 📝 待办事项

- [ ] 添加多串口同时监控
- [ ] 支持数据过滤和搜索
- [ ] 添加数据导出功能（CSV、JSON）
- [ ] 支持命令行参数配置
- [ ] 添加用户配置导入/导出
- [ ] 支持数据统计和图表

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

本项目采用 [MIT](LICENSE) 许可证。

## 🙏 致谢

- [go.bug.st/serial](https://github.com/bugst/go-serial) - 串口通信库
- [Bootstrap](https://getbootstrap.com/) - UI框架
- [Bootstrap Icons](https://icons.getbootstrap.com/) - 图标库

---

<div align="center">

**Made with ❤️ by SerialWebViewer Team**

</div>
