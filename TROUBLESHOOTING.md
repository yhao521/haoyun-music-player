# 系统托盘问题排查指南

## 🔍 问题现象

系统托盘图标未显示在 macOS 菜单栏中。

## ✅ 已修复的问题

### 1. **变量声明顺序**
- ✅ 先声明 `mainWindow` 变量为 `nil`
- ✅ 再创建系统托盘和菜单项
- ✅ 最后初始化窗口并隐藏

### 2. **API 使用修正**
- ✅ 移除不存在的 `Visible` 字段
- ✅ 使用 `mainWindow.Hide()` 方法隐藏窗口
- ✅ 确保托盘在应用启动前完成初始化

### 3. **调试日志**
添加了详细的日志输出，便于追踪托盘创建过程：
```go
log.Println("✓ System tray initialized")
log.Println("✓ System tray menu created")
log.Println("✓ Main window created (hidden)")
```

## 🚀 测试步骤

### 1. 清理并重新编译
```bash
cd /Users/yanghao/storage/code_projects/goProjects/haoyun-music-player

# 清理缓存
go clean -cache

# 整理依赖
go mod tidy
```

### 2. 开发模式运行
```bash
wails3 dev -config ./build/config.yml
```

观察控制台输出，应该看到：
```
✓ System tray initialized
✓ System tray menu created
✓ Main window created (hidden)
```

### 3. 检查托盘位置

在 macOS 上，托盘图标应该显示在：
- **右上角菜单栏**（靠近时间和 WiFi 图标）
- 可能需要点击菜单栏空白处查看是否有隐藏图标

### 4. 验证功能

右键点击托盘图标应该显示菜单：
- 播放/暂停
- 上一首
- 下一首
- 显示主窗口
- 退出

单击托盘图标应该切换播放状态。

## ⚠️ 可能的原因

### Wails v3 Alpha 限制

由于 Wails v3 处于 Alpha 阶段，系统托盘功能可能：
1. **API 不完整** - 部分功能尚未实现
2. **平台差异** - macOS 和 Windows/Linux 行为不同
3. **需要特殊配置** - 可能需要 Info.plist 或其他配置

### macOS 系统限制

macOS 对菜单栏应用有以下要求：
1. **LSUIElement** - 需要在 Info.plist 中设置
2. **激活策略** - 可能需要设置为 accessory application
3. **沙盒限制** - 某些权限可能影响托盘显示

## 🔧 解决方案

### 方案 1：纯菜单栏模式（当前实现）

```go
// 隐藏主窗口
mainWindow.Hide()

// 只显示托盘图标
tray := app.SystemTray.New()
tray.SetMenu(menu)
```

### 方案 2：添加自定义图标

取消注释 main.go 中的图标嵌入代码：

```go
//go:embed frontend/public/tray-icon.png
var trayIcon []byte

// 设置图标
tray.SetIcon(trayIcon)
```

需要准备图标文件：
- 尺寸：22x22 或 44x44（@2x）
- 格式：PNG（透明背景）
- 风格：单色模板（推荐）

### 方案 3：修改 Info.plist

创建 `build/darwin/info.plist`：

```xml
<key>LSUIElement</key>
<string>true</string>
```

这会让应用作为 UI Element 运行，不显示 Dock 图标。

### 方案 4：检查 Wails 版本

```bash
# 查看当前版本
go list -m github.com/wailsapp/wails/v3

# 更新到最新版本
go get -u github.com/wailsapp/wails/v3@latest
```

## 📊 调试技巧

### 1. 查看详细日志
运行应用时注意观察日志输出。

### 2. 使用 Activity Monitor
在 macOS 的"活动监视器"中查找进程，确认应用是否运行。

### 3. 强制显示窗口
如果托盘未显示但应用已运行：
```bash
# 通过命令行发送事件（需要实现相应功能）
```

### 4. 检查系统权限
macOS 可能需要授权：
- 系统偏好设置 → 安全性与隐私 → 辅助功能
- 允许应用控制计算机

## 🎯 预期行为

成功启动后应该：

1. ✅ 菜单栏右上角显示图标（或文字提示）
2. ✅ 右键点击显示完整菜单
3. ✅ 单击/双击有响应
4. ✅ 主窗口默认隐藏
5. ✅ 可通过"显示主窗口"菜单打开窗口

## 📝 已知问题

### Wails v3 Alpha 74 的限制

根据当前使用的版本（v3.0.0-alpha.74）：

1. **系统托盘管理器** - 可能还在开发中
2. **macOS 特定 API** - 调用方式可能不正确
3. **文档不完善** - 缺少官方示例

### 替代方案

如果系统托盘确实无法工作，可以考虑：

1. **使用小窗口** - 创建一个小尺寸的常驻窗口
2. **全局快捷键** - 实现键盘快捷键控制
3. **通知中心** - 集成 macOS 通知中心

## 🔗 相关资源

- [Wails v3 文档](https://wails.io/)
- [系统托盘 API](https://wails.io/docs/reference/system-tray)
- [GitHub Issues](https://github.com/wailsapp/wails/issues)
- [Wails Discord 社区](https://discord.gg/JDdSxwjhGf)

## 💡 下一步

1. **运行测试** - 使用 `wails3 dev` 启动应用
2. **观察日志** - 查看是否有错误信息
3. **报告问题** - 如果托盘仍不显示，收集日志并提交 Issue
4. **等待更新** - Wails v3 稳定后可能会修复此问题

---

**最后更新**: 2026-04-02  
**状态**: 🟡 代码已修复，待测试验证
