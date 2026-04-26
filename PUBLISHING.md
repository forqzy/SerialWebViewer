# 发布指南 - SerialWebViewer

## 📖 理解发布概念

### 源代码 vs 二进制文件

**源代码:**
- 人类可读的代码（如 main.go）
- 需要编译才能运行
- GitHub 上默认显示的就是源代码
- 开发者获取源代码后可以自行编译

**二进制文件:**
- 已经编译好的可执行程序
- 直接运行，不需要编译
- 比如：`SerialWebViewer.exe` (Windows) 或 `serialwebviewer` (Linux)
- 通常通过 GitHub Releases 发布

### 发布流程概述

```
开发 → 源代码 → GitHub → 编译 → GitHub Releases → 用户下载
```

1. **开发阶段**: 在本地编写代码
2. **推送到GitHub**: 源代码上传到 GitHub
3. **创建Release**: 在 GitHub 上创建一个发布版本
4. **上传二进制**: 把编译好的可执行文件上传到 Release
5. **用户下载**: 用户可以直接下载使用，无需编译

---

## 🚀 完整发布流程

### 步骤1：推送到GitHub（首次）

```bash
cd ~/project/SerialWebViewer

# 初始化Git（如果还没做）
git init
git add .
git commit -m "Initial commit"

# 创建GitHub仓库后
git remote add origin https://github.com/forqzy/SerialWebViewer.git
git branch -M main
git push -u origin main
```

### 步骤2：构建所有平台的二进制文件

```bash
cd ~/project/SerialWebViewer

# 使用构建脚本
./build.sh

# 或使用Makefile
make all
```

构建完成后，`build/` 目录包含：
- `SerialWebViewer.exe` - Windows版本
- `serialwebviewer` - Linux版本
- `SerialWebViewer.mac` - macOS Intel版本
- `SerialWebViewer-arm64.mac` - macOS ARM64版本

### 步骤3：创建GitHub Release

#### 方法A：使用GitHub网页（推荐新手）

1. 访问你的GitHub仓库
   ```
   https://github.com/forqzy/SerialWebViewer
   ```

2. 点击右侧的 **"Releases"** 链接

3. 点击 **"Create a new release"** 按钮

4. 填写发布信息：
   ```
   Tag: v1.0.0
   Title: SerialWebViewer v1.0.0 - First Stable Release
   Description: 
   ```
   
   在描述框中添加发布说明（可以复制下面的内容）：

```markdown
## 🎉 SerialWebViewer v1.0.0

这是SerialWebViewer的第一个稳定版本！

### ✨ 新功能
- 📊 实时串口数据监控
- 🎨 现代化Web界面
- 🌐 多主题支持（9种预设 + 自定义）
- 🌍 中英文界面
- 💾 自动日志记录
- 🔧 完整的API支持

### 📦 下载说明
- **Windows**: 下载 `SerialWebViewer.exe`
- **Linux**: 下载 `serialwebviewer`，运行 `chmod +x serialwebviewer`
- **macOS (Intel)**: 下载 `SerialWebViewer.mac`
- **macOS (Apple Silicon)**: 下载 `SerialWebViewer-arm64.mac`

### 🚀 快速开始
1. 下载适合你平台的版本
2. 双击运行（Windows）或 `./serialwebviewer` (Linux/macOS)
3. 浏览器打开 `http://localhost:8088`
4. 配置串口参数并连接

### 📝 更新日志
- 初始发布
- 完整功能实现
- 跨平台支持

### 🙏 致谢
感谢所有测试用户的反馈！

Full changelog: https://github.com/forqzy/SerialWebViewer/compare/v0.0.0...v1.0.0
```

5. **上传二进制文件**：
   - 拖拽 `build/` 目录下的所有文件到发布页面
   - 或点击 "Attach binaries" 按钮

6. 点击 **"Publish release"** 按钮

#### 方法B：使用GitHub CLI（推荐高级用户）

```bash
# 安装GitHub CLI（如果还没安装）
# Ubuntu/Debian
sudo apt install gh

# macOS
brew install gh

# 登录
gh auth login

# 构建所有平台
./build.sh

# 创建发布
gh release create v1.0.0 \
  --title "SerialWebViewer v1.0.0 - First Stable Release" \
  --notes "🎉 First stable release of SerialWebViewer!" \
  build/*
```

### 步骤4：验证发布

1. 访问Releases页面：
   ```
   https://github.com/forqzy/SerialWebViewer/releases
   ```

2. 检查：
   - ✅ 版本标签显示正确（v1.0.0）
   - ✅ 二进制文件都上传了
   - ✅ 描述信息完整

3. 测试下载：
   - 下载自己平台的版本
   - 运行测试
   - 确认功能正常

---

## 📋 版本号规范

使用语义化版本号：

- **v1.0.0** - 第一个稳定版本
- **v1.0.1** - Bug修复版本
- **v1.1.0** - 新功能版本（向后兼容）
- **v2.0.0** - 重大更新（可能不兼容）

### 标签示例：
```
v1.0.0  → 初始发布
v1.0.1  → 修复bug
v1.1.0  → 添加小功能
v2.0.0  → 大版本更新
```

---

## 🔄 后续更新流程

### 修改代码后

```bash
# 1. 修改代码
vim main.go

# 2. 测试
go run main.go

# 3. 提交到GitHub
git add .
git commit -m "Add new feature"
git push

# 4. 构建新版本
./build.sh

# 5. 创建新的Release
gh release create v1.1.0 \
  --title "SerialWebViewer v1.1.0" \
  --notes "新功能说明" \
  build/*
```

### GitHub CLI 发布命令示例

```bash
# 预发布（草稿）
gh release create v1.0.0-rc1 \
  --title "SerialWebViewer v1.0.0-rc1" \
  --draft \
  build/*

# 正式发布
gh release create v1.0.0 \
  --title "SerialWebViewer v1.0.0" \
  --notes "正式发布" \
  --latest \
  build/*

# 删除发布
gh release delete v1.0.0
```

---

## 📦 Release 检查清单

发布前检查：

- [ ] 代码已测试，功能正常
- [ ] 所有平台的二进制都已编译
- [ ] 版本号已更新
- [ ] README.md 已更新
- [ ] CHANGELOG.md 已更新
- [ ] 发布说明已准备好
- [ ] 测试过下载和安装

---

## 🎯 实际操作示例

### 第一次发布 v1.0.0

```bash
# 1. 准备工作
cd ~/project/SerialWebViewer
git status
git log --oneline -5

# 2. 构建
./build.sh
ls -lh build/

# 3. 使用GitHub CLI发布
gh release create v1.0.0 \
  --title "SerialWebViewer v1.0.0" \
  --notes "$(cat <<'EOF'
## 🎉 SerialWebViewer v1.0.0

第一个稳定版本发布！

### ✨ 特性
- 实时串口监控
- 现代化Web界面
- 多主题支持
- 中英文界面
- 自动日志记录

### 📦 下载
选择适合你平台的版本下载即可。

EOF
)" \
  build/*

# 4. 验证
gh release view v1.0.0
```

### 发布Bug修复版本 v1.0.1

```bash
# 1. 修复bug
vim main.go
git commit -am "Fix crash when port disconnected"

# 2. 推送
git push

# 3. 重新构建
./build.sh

# 4. 发布
gh release create v1.0.1 \
  --title "SerialWebViewer v1.0.1" \
  --notes "修复了端口断开时的崩溃问题" \
  build/*
```

---

## 💡 发布最佳实践

### 1. 版本管理
- 每次发布都创建Git标签
- 使用语义化版本号
- 在CHANGELOG.md记录所有更改

### 2. 发布说明
- 清楚说明新功能和改进
- 列出已知问题
- 提供升级指南

### 3. 二进制文件
- 为所有主流平台编译
- 包含版本信息
- 文件名清晰明了

### 4. 安全性
- 不要在Release中包含敏感信息
- 检查日志文件中没有敏感数据
- 代码已审查

---

## 🔗 相关链接

- **创建Release**: https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-project
- **GitHub CLI**: https://cli.github.com/manual/
- **语义化版本**: https://semver.org/

---

## ❓ 常见问题

### Q: 每次都要上传所有平台的二进制吗？
A: 是的，但使用脚本可以一键完成所有平台的编译和上传。

### Q: 可不可以只发布源代码？
A: 可以，但用户需要自己编译，体验不好。建议同时提供二进制。

### Q: 发布的文件可以删除吗？
A: 可以，但不建议。删除后用户就下载不到了。

### Q: 可以更新已发布的二进制吗？
A: 不推荐。应该发布新版本（v1.0.1, v1.0.2等），旧版本保持不变。

### Q: 如何回滚到旧版本？
A: 用户可以从Releases页面下载旧版本，但标记为 "Pre-release" 或直接删除新版本。

---

## 📞 需要帮助？

发布过程中遇到问题？
- 查看: https://github.com/forqzy/SerialWebViewer/issues
- Email: forqzy@gmail.com

---

**Happy Publishing! 🎉**
