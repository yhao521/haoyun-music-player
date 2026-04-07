# 依赖工具自动安装功能 - 完整实现指南

## 🎉 功能概述

本项目已成功集成智能依赖工具检测和自动安装功能，用户可以：
- ✅ 在托盘菜单中查看所有依赖工具的状态
- ✅ 一键安装缺失的工具（如 FFmpeg）
- ✅ 后台异步安装，不阻塞应用使用
- ✅ 跨平台支持（macOS、Windows、Linux）
- ✅ 实时状态更新和通知

## 📦 已实现的功能

### 1. DependencyManager 核心类

**文件**: [backend/dependency_manager.go](file:///Users/yanghao/storage/code_projects/goProjects/haoyun-music-player/backend/dependency_manager.go)

#### 主要特性
- **工具检测**: 自动扫描系统 PATH 和环境变量
- **版本获取**: 显示已安装工具的版本信息
- **状态管理**: 4 种状态（未安装/安装中/已安装/安装失败）
- **回调机制**: 状态变化时通知前端
- **跨平台安装**: 
  - macOS: Homebrew
  - Windows: Chocolatey / Scoop
  - Linux: apt-get / dnf / pacman

#### 核心 API
```go
// 创建管理器
dm := NewDependencyManager()

// 检查所有工具
tools := dm.CheckAllTools()

// 安装指定工具（异步）
dm.InstallTool("ffmpeg")

// 获取状态摘要
summary := dm.GetInstallSummary()

// 设置状态变化回调
dm.SetCallback(func(toolName string, status ToolStatus, message string) {
    // 处理状态变化
})
```

### 2. 托盘菜单集成

**文件**: [main.go](file:///Users/yanghao/storage/code_projects/goProjects/haoyun-music-player/main.go)

#### 菜单结构
```
🛠️ 依赖工具
├── ✅ FFmpeg (ffmpeg version 6.x...)  # 已安装
├── ❌ FFprobe                          # 未安装
│   └── 📦 安装 FFprobe
│   └── ℹ️ 随 FFmpeg 一起安装
├── ─────────────
└── 🔄 重新检查所有工具
```

#### 交互流程
1. 用户点击"📦 安装 XXX"
2. 显示通知："正在后台安装 XXX，请稍候..."
3. 后台执行安装命令
4. 安装完成后自动重建菜单
5. 显示结果通知

### 3. 应用启动检测

应用启动时自动执行：
```go
go func() {
    time.Sleep(1 * time.Second)
    
    // 检查所有依赖
    depManager.CheckAllTools()
    
    // 打印状态摘要
    log.Println(depManager.GetInstallSummary())
    
    // 如果有缺失工具，通知前端
    if depManager.NeedInstall() {
        app.Event.Emit("missingDependencies", ...)
    }
}()
```

## 🚀 使用方法

### 查看工具状态

1. 点击系统托盘图标
2. 找到 "🛠️ 依赖工具" 菜单项
3. 展开查看各工具状态

### 安装缺失工具

1. 在未安装的工具子菜单中点击 "📦 安装 XXX"
2. 等待后台安装完成
3. 菜单会自动刷新显示新状态

### 手动刷新状态

点击 "🔄 重新检查所有工具" 手动触发检测

## 🔧 技术实现细节

### 1. 前向声明解决闭包问题

由于 Go 的闭包不能引用后面定义的局部函数，我们使用前向声明：

```go
// 步骤 1: 声明函数变量
var buildToolsMenu func()
var rebuildTrayMenu func()

// 步骤 2: 在回调中使用（此时函数还未定义）
depManager.SetCallback(func(...) {
    rebuildTrayMenu()  // ✅ 合法
})

// 步骤 3: 后面再赋值
buildToolsMenu = func() { ... }
rebuildTrayMenu = func() { ... }
```

### 2. 跨平台安装策略

```go
func performInstall(toolName string) error {
    switch runtime.GOOS {
    case "darwin":
        return installOnMacOS(toolName)    // brew install
    case "windows":
        return installOnWindows(toolName)  // choco/scoop
    case "linux":
        return installOnLinux(toolName)    // apt/dnf/pacman
    }
}
```

### 3. 异步安装与状态同步

```go
// 标记为安装中
tool.Status = ToolInstalling

// 异步执行
go func() {
    err := executeInstall()
    
    // 更新状态
    if err != nil {
        tool.Status = ToolInstallFailed
    } else {
        tool.Status = ToolInstalled
    }
    
    // 触发回调
    emitCallback(...)
}()
```

## 📊 支持的工具

| 工具 | 用途 | 检测方式 | 安装方式 |
|------|------|---------|---------|
| **FFmpeg** | 音频解码 | `exec.LookPath` | Homebrew/Choco/apt |
| **FFprobe** | 元数据提取 | `exec.LookPath` | 随 FFmpeg 安装 |

## 🎨 前端集成（可选）

监听后端事件以显示通知：

```typescript
// Vue 组件中
import { onMounted } from 'vue'
import { EventsOn } from '@wailsio/runtime'

onMounted(() => {
  // 监听依赖状态变化
  EventsOn('dependencyStatusChanged', (data: any) => {
    console.log(`工具 ${data.tool} 状态: ${data.status}`)
    showNotification(data.title, data.message, data.type)
  })
  
  // 监听缺失依赖提示
  EventsOn('missingDependencies', (data: any) => {
    alert(`检测到 ${data.tools.length} 个工具缺失`)
  })
})
```

## ⚙️ 配置选项

### 环境变量

```bash
# 自定义 FFmpeg 路径
export FFMPEG_PATH=/custom/path/to/ffmpeg
```

### 添加新工具

在 `initTools()` 中添加：

```go
dm.tools["newtool"] = &ToolInfo{
    Name:    "New Tool",
    Command: "newtool",
    Status:  ToolNotInstalled,
    InstallHint: "安装提示...",
}
```

然后在 `installOnMacOS/Windows/Linux` 中添加安装逻辑。

## 🐛 故障排除

### 问题 1: 安装失败

**症状**: 显示 "安装失败" 通知

**可能原因**:
- 包管理器未安装（Homebrew/Chocolatey/Scoop）
- 网络连接问题
- 权限不足

**解决方案**:
1. 查看日志获取详细错误信息
2. 手动运行安装命令测试
3. 检查网络连接

### 问题 2: 菜单不刷新

**症状**: 安装完成后菜单状态未更新

**解决方案**:
1. 手动点击 "🔄 重新检查所有工具"
2. 重启应用

### 问题 3: 检测不到已安装的工具

**症状**: 工具已安装但显示 "未安装"

**可能原因**:
- 工具不在系统 PATH 中
- 需要重启终端使 PATH 生效

**解决方案**:
1. 运行 `which ffmpeg` (macOS/Linux) 或 `where ffmpeg` (Windows) 确认路径
2. 设置 `FFMPEG_PATH` 环境变量
3. 重启应用

## 📝 开发笔记

### 关键设计决策

1. **异步安装**: 避免阻塞 UI，提升用户体验
2. **状态持久化**: 每次启动重新检测，确保准确性
3. **友好提示**: 提供清晰的安装指引
4. **可扩展性**: 易于添加新工具检测

### 性能考虑

- 检测操作轻量（仅调用 `exec.LookPath`）
- 安装操作异步执行
- 菜单重建有 500ms 延迟确保状态同步

### 安全考虑

- 使用系统包管理器，避免下载不可信二进制
- 需要用户明确点击才触发安装
- 安装命令公开透明

## 🔮 未来扩展

### 计划添加的工具
- [ ] Git（版本控制）
- [ ] ImageMagick（图像处理）
- [ ] youtube-dl（视频下载）

### 增强功能
- [ ] 安装进度实时显示
- [ ] 批量安装所有缺失工具
- [ ] 自动更新检查
- [ ] 工具卸载功能

## 📚 相关文档

- [FFmpeg 安装指南](./FFMPEG_GUIDE.md)
- [FFmpeg 集成总结](./FFMPEG_INTEGRATION_SUMMARY.md)
- [快速参考](./FFMPEG_QUICKREF.md)

---

**最后更新**: 2026-04-07  
**状态**: ✅ 已完成并测试通过  
**维护者**: Haoyun Music Player Team