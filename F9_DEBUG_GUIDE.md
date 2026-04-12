# macOS F9 媒体键调试指南

## 🔍 问题描述

在 macOS 上按下 F9 (下一曲) 键,应用没有响应。

## 📋 诊断步骤

### 步骤 1: 运行诊断脚本

```bash
./debug_mediakeys.sh
```

这个脚本会:

- ✅ 检查辅助功能权限
- ✅ 验证 gohook 依赖
- ✅ 编译项目
- ✅ 启动应用进行测试

### 步骤 2: 观察控制台日志

启动应用后,你应该看到:

```
========================================
🎣 [MediaKey] 使用 gohook 注册媒体键...
📱 [MediaKey] 当前平台: darwin
🔒 [MediaKey] macOS: 已锁定到主线程
🎧 [MediaKey] 正在启动键盘监听器...
✅ [MediaKey] 键盘监听器已启动
✅ [MediaKey] 事件处理循环已启动,等待媒体键事件...
✅ [MediaKey] 媒体键注册成功!
💡 [MediaKey] 提示: 按下任意键测试,查看控制台 DEBUG 日志
```

### 步骤 3: 测试按键捕获

**按下任意键**(不仅仅是 F9),观察是否有类似这样的日志:

```
🔍 [DEBUG #1] Kind=3, Rawcode=0x0, Keychar=a, Mask=0x0
🔍 [DEBUG #2] Kind=3, Rawcode=0xB9, Keychar=, Mask=0x0
```

#### 情况分析:

**情况 A: 没有任何 DEBUG 日志**

- ❌ **问题**: gohook 完全没有捕获到按键事件
- **原因**: 缺少辅助功能权限
- **解决**:
  1. 打开: 系统偏好设置 > 安全性与隐私 > 隐私 > 辅助功能
  2. 勾选你的应用
  3. **完全退出并重启应用**

**情况 B: 有 DEBUG 日志,但按下 F9 时 Rawcode 不是 0xB9**

- ⚠️ **问题**: 键码映射不正确
- **示例**: 如果 F9 的 Rawcode 是 `0x6B` 而不是 `0xB9`
- **解决**: 需要调整键码映射(见下文)

**情况 C: 看到 `Rawcode=0xB9`,但没有后续日志**

- ⚠️ **问题**: 键码匹配了,但 handler 没有执行
- **可能原因**: `globalMediaKeyService` 为 nil
- **解决**: 检查服务初始化顺序

**情况 D: 看到完整的日志链**

```
🔍 [DEBUG #X] Kind=3, Rawcode=0xB9, Keychar=, Mask=0x0
🎵 [MediaKey] 检测到媒体键: Next (Rawcode: 0xB9)
⏭️  收到媒体键:下一曲
```

- ✅ **说明**: 媒体键监听正常工作
- **如果音乐没切换**: 问题在 MusicService,不在媒体键监听

---

## 🔧 常见解决方案

### 方案 1: 修复权限问题(最常见)

```bash
# 1. 检查权限状态
osascript -e 'tell application "System Events" to get UI elements enabled'

# 如果返回 false,需要手动添加权限

# 2. 打开系统偏好设置
open "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility"

# 3. 勾选应用后,完全退出并重启
killall haoyun-music-player
./haoyun-music-player
```

### 方案 2: 调整键码映射

如果调试日志显示 F9 的 Rawcode 不是 `0xB9`,需要修改映射:

编辑 `backend/mediakeyservice_gohook.go`:

```go
case "darwin":
    switch rawcode {
    case 0xB7: // <-- 根据实际日志修改
        keyName = "Play/Pause"
        // ...
    case YOUR_ACTUAL_RAWCODE: // <-- 改成日志中显示的值
        keyName = "Next"
        handler = func() {
            if globalMediaKeyService != nil {
                globalMediaKeyService.handleNext()
            }
        }
    // ...
    }
```

然后重新编译:

```bash
go build .
```

### 方案 3: 检查服务初始化顺序

确保在调用 `RegisterMediaKeys()` 之前已经设置了 `MusicService`:

```go
// 正确的初始化顺序
mediaKeyService := backend.NewMediaKeyService()
mediaKeyService.SetApp(app)
mediaKeyService.SetMusicService(musicService) // ← 必须先设置
mediaKeyService.RegisterMediaKeys()           // ← 再注册
```

检查你的 `app_init.go` 或 `main.go` 中的初始化代码。

---

## 🐛 高级调试

### 启用详细日志

如果上述方法都不行,可以临时启用更详细的日志:

在 `handleMediaKeyEvent` 函数开头添加:

```go
func handleMediaKeyEvent(rawcode uint16) {
    // 打印所有事件,包括非 KeyDown
    log.Printf("🔍 [RAW EVENT] Rawcode=0x%X", rawcode)

    // ... 原有代码 ...
}
```

### 检查 gohook 版本兼容性

```bash
# 查看当前版本
go list -m github.com/robotn/gohook

# 尝试更新到最新版本
go get github.com/robotn/gohook@latest
go mod tidy

# 或者回退到稳定版本
go get github.com/robotn/gohook@v0.40.0
go mod tidy
```

### 测试其他按键

创建一个简单的测试程序:

```go
package main

import (
    "fmt"
    hook "github.com/robotn/gohook"
)

func main() {
    fmt.Println("按下任意键查看 rawcode...")
    evChan := hook.Start()
    defer hook.End()

    for ev := range evChan {
        if ev.Kind == hook.KeyDown {
            fmt.Printf("Rawcode: 0x%X\n", ev.Rawcode)
        }
    }
}
```

保存为 `test_keys.go`,运行:

```bash
go run test_keys.go
```

按下 F9,记录输出的 Rawcode 值。

---

## 📊 已知问题

### 问题 1: macOS Sonoma/Ventura 权限问题

**症状**: 即使勾选了权限,仍然无法捕获按键

**解决**:

```bash
# 重置 TCC 数据库(需要重启)
tccutil reset Accessibility

# 或者删除特定应用的权限记录
sqlite3 ~/Library/Application\ Support/com.apple.TCC/TCC.db \
  "DELETE FROM access WHERE client='com.your.bundle.id';"
```

### 问题 2: 外接键盘键码不同

**症状**: MacBook 内置键盘正常,外接键盘不工作

**解决**: 外接键盘可能发送不同的 rawcode,需要分别映射:

```go
case 0xB9, 0x6B, 0x7F: // 多个可能的 F9 键码
    keyName = "Next"
    // ...
```

### 问题 3: 与其他软件冲突

**症状**: Spotify、Apple Music 等占用了媒体键

**解决**:

1. 退出其他音乐播放器
2. 或在系统偏好设置中禁用其他应用的媒体键控制

---

## ✅ 验证清单

- [ ] 辅助功能权限已授予
- [ ] 应用已完全重启
- [ ] 控制台显示 "键盘监听器已启动"
- [ ] 按下普通键能看到 DEBUG 日志
- [ ] 按下 F9 看到 Rawcode=0xB9 (或其他值)
- [ ] 看到 "检测到媒体键: Next" 日志
- [ ] 看到 "收到媒体键:下一曲" 日志
- [ ] MusicService 已正确初始化
- [ ] 播放列表中有至少两首歌曲

---

## 🆘 获取帮助

如果以上方法都无法解决,请提供以下信息:

1. **macOS 版本**:

   ```bash
   sw_vers
   ```

2. **完整日志输出**: 从启动到按下 F9 的所有日志

3. **gohook 版本**:

   ```bash
   go list -m github.com/robotn/gohook
   ```

4. **测试结果**:
   - 是否有 DEBUG 日志?
   - F9 的 Rawcode 是多少?
   - 是否看到 "检测到媒体键" 日志?

5. **权限状态**:
   ```bash
   osascript -e 'tell application "System Events" to get UI elements enabled'
   ```
