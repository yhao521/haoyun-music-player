# 系统托盘修复总结

## ✅ 已完成的修复

### 1. **修正变量声明顺序**
```go
// ❌ 错误：先使用未定义的变量
mainWindow = app.Window.NewWithOptions(...)

// ✅ 正确：先声明变量
var mainWindow *application.WebviewWindow
// ... 创建菜单项时使用 mainWindow
mainWindow = app.Window.NewWithOptions(...)
```

### 2. **移除不存在的 API**
```go
// ❌ 错误：Visible 字段不存在
WebviewWindowOptions{
    Visible: false,
}

// ✅ 正确：使用 Hide() 方法
mainWindow = app.Window.NewWithOptions(...)
mainWindow.Hide()
```

### 3. **优化初始化流程**
```go
func main() {
    // 1. 创建服务
    musicService := NewMusicService()
    
    // 2. 创建应用
    app := application.New(...)
    
    // 3. 声明窗口变量（关键！）
    var mainWindow *application.WebviewWindow
    
    // 4. 创建系统托盘（在窗口之前）
    tray := app.SystemTray.New()
    
    // 5. 创建菜单项（引用 mainWindow）
    showItem.OnClick(func() {
        if mainWindow != nil {
            mainWindow.Show()
        }
    })
    
    // 6. 设置托盘菜单
    tray.SetMenu(menu)
    
    // 7. 最后创建窗口并隐藏
    mainWindow = app.Window.NewWithOptions(...)
    mainWindow.Hide()
    
    // 8. 运行应用
    app.Run()
}
```

## 📋 代码变更清单

### 修改的文件
- ✅ [`main.go`](main.go) - 重构主函数，修复托盘初始化逻辑

### 新增的文件
- ✅ [`TROUBLESHOOTING.md`](TROUBLESHOOTING.md) - 详细的问题排查指南

## 🔍 关键技术点

### 1. Go 的零值初始化
```go
var mainWindow *application.WebviewWindow
// mainWindow 初始值为 nil
// 可以在闭包中安全引用
```

### 2. 闭包捕获外部变量
```go
showItem.OnClick(func(ctx *application.Context) {
    // 这里引用的 mainWindow 是外部声明的变量
    // 即使当时是 nil，点击时已经赋值了
    if mainWindow != nil {
        mainWindow.Show()
    }
})
```

### 3. Wails v3 的系统托盘机制
- `app.SystemTray.New()` 创建托盘实例
- `tray.SetMenu(menu)` 设置菜单
- `tray.OnClick()` 绑定单击事件
- 托盘图标在 `app.Run()` 后显示

## 🚀 测试方法

### 快速测试
```bash
cd /Users/yanghao/storage/code_projects/goProjects/haoyun-music-player

# 方式 1：开发模式（推荐）
wails3 dev -config ./build/config.yml

# 方式 2：直接运行编译后的二进制
./haoyun-music-player
```

### 预期结果

#### 控制台输出
```
✓ System tray initialized
✓ System tray menu created
✓ Main window created (hidden)
```

#### 菜单栏显示
- 右上角应该能看到图标或文字提示
- 右键点击显示完整菜单
- 菜单项包括：
  - 播放/暂停
  - 上一首
  - 下一首
  - 分隔线
  - 显示主窗口
  - 分隔线
  - 退出

#### 交互测试
1. **单击图标** - 切换播放状态
2. **双击图标** - 显示主窗口
3. **右键菜单** - 各菜单项可点击

## ⚠️ 注意事项

### macOS 版本兼容性
编译警告显示：
```
object file was built for newer 'macOS' version (26.0) than being linked (11.0)
```

这是因为：
- 你的 macOS 版本较新（可能是 macOS 26 Sequoia）
- Go 工具链使用了新的 SDK
- 链接时指定了较低的部署目标（11.0）

**解决方案**：忽略警告，这不会影响功能。

### Wails v3 Alpha 限制
当前使用版本：`v3.0.0-alpha.74`

可能存在的问题：
- 系统托盘 API 尚未完善
- macOS 特定功能可能需要特殊配置
- 文档和示例较少

## 🎯 如果托盘仍然不显示

### 检查清单

1. **确认应用运行**
   ```bash
   ps aux | grep haoyun-music-player
   ```

2. **查看活动监视器**
   - 打开"活动监视器"
   - 查找 "Haoyun Music Player" 进程

3. **检查菜单栏**
   - 点击右上角所有区域
   - 有些图标可能被隐藏在控制中心里

4. **查看详细日志**
   ```bash
   # 运行应用并保存日志
   wails3 dev 2>&1 | tee app.log
   ```

### 可能的原因

1. **Wails v3 Bug** - Alpha 版本的已知问题
2. **Info.plist 配置** - 需要设置 LSUIElement
3. **沙盒权限** - macOS 安全限制
4. **图标资源缺失** - 默认图标可能不显示

### 临时解决方案

如果托盘确实无法工作，可以：

1. **显示小窗口**
   ```go
   // 不隐藏窗口，作为替代方案
   mainWindow.SetAlwaysOnTop(true)
   ```

2. **使用全局快捷键**
   ```go
   // 注册媒体键
   app.RegisterGlobalHotkey(...)
   ```

3. **等待 Wails v3 稳定版**
   - 关注官方更新
   - 升级到最新版本

## 📞 获取帮助

如果问题持续，可以：

1. **查看官方文档**
   - [Wails v3 Docs](https://wails.io/)
   - [系统托盘参考](https://wails.io/docs/reference/system-tray)

2. **加入社区**
   - [Wails Discord](https://discord.gg/JDdSxwjhGf)
   - [GitHub Discussions](https://github.com/wailsapp/wails/discussions)

3. **提交 Issue**
   - 提供详细的日志
   - 说明 macOS 版本
   - 附上复现步骤

## 🎉 总结

本次修复主要解决了：
- ✅ 变量声明和使用顺序问题
- ✅ 移除不存在的 API 调用
- ✅ 优化系统托盘初始化流程
- ✅ 添加调试日志便于排查

**下一步**：运行 `wails3 dev` 测试托盘是否正常显示！

---

**修复时间**: 2026-04-02  
**状态**: 🟢 代码已修复，等待测试验证
