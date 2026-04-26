# 快速开始指南

## 运行程序

### Windows
```bash
SerialWebViewer.exe
```

### Linux/macOS
```bash
chmod +x serialwebviewer  # 或 SerialWebViewer.mac
./serialwebviewer
```

## 访问界面

启动程序后，在浏览器中打开：
```
http://localhost:8088
```

## 基本使用

### 1. 连接串口
1. 点击右上角"配置"按钮
2. 选择COM口和波特率
3. 点击"连接"按钮

### 2. 查看日志
- 实时日志会自动显示在主界面
- 可以切换HEX/文本模式
- 可以显示/隐藏时间戳

### 3. 管理日志文件
- 点击"文件"按钮展开历史文件列表
- 可以查看、下载或删除历史日志

### 4. 自定义主题
1. 在主题选择器中选择"自定义"
2. 调整背景色、文字色、时间戳色和字体大小
3. 点击"完成"保存设置

### 5. 切换语言
- 点击状态栏的"EN"或"中文"按钮切换界面语言

## 常见问题

**Q: 找不到串口？**
A: 点击配置面板中的"刷新端口"按钮重新扫描。

**Q: 日志文件保存在哪里？**
A: 日志文件保存在程序目录下的 `logs/` 文件夹中。

**Q: 如何修改Web端口？**
A: 启动时使用 `-port` 参数，例如：`SerialWebViewer.exe -port=9000`

**Q: 支持多串口同时监控吗？**
A: 当前版本仅支持单串口，多串口支持正在开发中。

## 技术支持

如有问题，请访问：https://github.com/yourusername/SerialWebViewer/issues
