# macOS 媒体键调试指南

## 🎯 问题描述

用户反馈：
1. ✅ 已授予辅助功能权限
2. ❌ 重启后仍然提示需要权限
3. ❌ F7/F8/F9 媒体键无法工作

---

## 🔧 已实施的修复

### 1. 改进权限检测逻辑

**之前的问题**：
- 仅使用 AppleScript 检测，可能不准确
- 没有区分"未授权"和"检测失败"

**现在的实现**：
```objective-c
// 方法 1: 使用官方 API AXIsProcessTrusted()（最可靠）
Boolean trusted = AXIsProcessTrusted();
if (trusted) {
    return 1; // 已信任
}

// 方法 2: 降级方案 - AppleScript
// 仅在方法 1 失败时使用
```

### 2. 添加详细调试日志

所有关键步骤都会输出到 stderr，格式为 `[MediaKey-C]` 前缀：

```
[MediaKey-C] 🔧 开始注册媒体键...
[MediaKey-C] 🔍 检查辅助功能权限...
[MediaKey-C] ✅ AXIsProcessTrusted: 已信任
[MediaKey-C] 🎯 创建 NSEvent 全局监视器...
[MediaKey-C] ✅ 媒体键监视器创建成功
```

### 3. 媒体键事件追踪

当按下媒体键时，会输出详细信息：

```
[MediaKey-C] 📨 收到系统事件
[MediaKey-C] 🎵 检测到媒体键事件
[MediaKey-C] 📊 按键信息: keyCode=16, keyState=10, isKeyDown=1
[MediaKey-C] ▶️ 触发: 播放/暂停
🎵 [回调] 执行播放/暂停
```

---

## 📋 测试步骤

### 步骤 1：完全清理并重新构建

```bash
# 1. 停止所有相关进程
pkill -f "Haoyun Music Player"
pkill -f "wails3"

# 2. 清理构建缓存
go clean -cache

# 3. 重新构建
cd /Users/yanghao/storage/code_projects/goProjects/haoyun-music-player
wails3 dev
```

### 步骤 2：查看启动日志

应用启动时，应该看到以下日志序列：

#### ✅ 正常情况（已有权限）
```
========================================
🍎 [MediaKey] 开始注册 macOS 媒体键...
💡 [MediaKey] 使用 NSEvent addGlobalMonitorForEventsMatchingMask
========================================
[MediaKey-C] 🔧 开始注册媒体键...
[MediaKey-C] 🔍 检查辅助功能权限...
[MediaKey-C] ✅ AXIsProcessTrusted: 已信任
[MediaKey-C] 🎯 创建 NSEvent 全局监视器...
[MediaKey-C] ✅ 媒体键监视器创建成功
========================================
✅ [MediaKey] macOS 媒体键注册成功！
📝 [MediaKey] 支持的按键:
   - F7 (上一曲)
   - F8 (播放/暂停)
   - F9 (下一曲)
💡 [MediaKey] 现在可以测试媒体键功能了
========================================
```

#### ⚠️ 没有权限的情况
```
[MediaKey-C] 🔧 开始注册媒体键...
[MediaKey-C] 🔍 检查辅助功能权限...
[MediaKey-C] ⚠️ AXIsProcessTrusted 返回 false，尝试 AppleScript...
[MediaKey-C] ❌ AppleScript: UI elements disabled
[MediaKey-C] ❌ 没有辅助功能权限，显示提示对话框
[MediaKey-C] 📢 显示权限提示对话框
```

此时会弹出对话框，提供三个选项。

### 步骤 3：验证权限状态

#### 方法 A：通过系统设置检查
1. 打开"系统偏好设置" > "安全性与隐私" > "隐私" > "辅助功能"
2. 确认您的应用已被勾选
3. **如果已勾选但仍然失败**：
   - 取消勾选
   - 等待 2 秒
   - 重新勾选
   - **完全退出应用**（不是关闭窗口）
   - 重新启动

#### 方法 B：通过终端命令检查
```bash
# 检查应用是否在辅助功能列表中
osascript -e 'tell application "System Events" to get UI elements enabled'
# 应该返回 true
```

### 步骤 4：测试媒体键

#### 4.1 按下 F8（播放/暂停）

**预期日志**：
```
[MediaKey-C] 📨 收到系统事件
[MediaKey-C] 🎵 检测到媒体键事件
[MediaKey-C] 📊 按键信息: keyCode=16, keyState=10, isKeyDown=1
[MediaKey-C] ▶️ 触发: 播放/暂停
🎵 [回调] 执行播放/暂停
```

**如果没有日志**：
- 说明 NSEvent 监视器没有捕获到事件
- 可能是权限问题或 API 调用失败

#### 4.2 按下 F7（上一曲）

**预期日志**：
```
[MediaKey-C] 📨 收到系统事件
[MediaKey-C] 🎵 检测到媒体键事件
[MediaKey-C] 📊 按键信息: keyCode=18, keyState=10, isKeyDown=1
[MediaKey-C] ⏮️ 触发: 上一曲
🎵 [回调] 执行上一曲
```

#### 4.3 按下 F9（下一曲）

**预期日志**：
```
[MediaKey-C] 📨 收到系统事件
[MediaKey-C] 🎵 检测到媒体键事件
[MediaKey-C] 📊 按键信息: keyCode=17, keyState=10, isKeyDown=1
[MediaKey-C] ⏭️ 触发: 下一曲
🎵 [回调] 执行下一曲
```

---

## 🐛 故障排查

### 问题 1：始终提示需要权限

**症状**：
- 已在系统设置中勾选应用
- 重启应用后仍然弹出权限提示

**可能原因**：
1. **AXIsProcessTrusted() 返回 false**
   - macOS 缓存了旧的权限状态
   
2. **应用 Bundle ID 变化**
   - 开发版本和发布版本的 Bundle ID 不同

**解决方案**：

#### 方案 A：重置辅助功能权限
```bash
# 1. 打开终端
# 2. 执行以下命令重置 TCC 数据库（需要管理员权限）
sudo tccutil reset Accessibility

# 3. 重新启动应用
# 4. 系统会再次请求权限，点击"允许"
```

⚠️ **注意**：这会重置所有应用的辅助功能权限，需要重新授权。

#### 方案 B：手动删除权限记录
```bash
# 1. 退出应用
# 2. 打开终端
# 3. 查找应用的权限记录
sqlite3 ~/Library/Application\ Support/com.apple.TCC/TCC.db "SELECT * FROM access WHERE service='kTCCServiceAccessibility';"

# 4. 删除特定应用的记录（替换 YOUR_BUNDLE_ID）
sqlite3 ~/Library/Application\ Support/com.apple.TCC/TCC.db "DELETE FROM access WHERE service='kTCCServiceAccessibility' AND client='YOUR_BUNDLE_ID';"

# 5. 重启应用
```

#### 方案 C：检查 Bundle ID
```bash
# 查看当前应用的 Bundle ID
/usr/libexec/PlistBuddy -c "Print :CFBundleIdentifier" /path/to/your/app.app/Contents/Info.plist

# 确保系统设置中的名称与此一致
```

### 问题 2：有权限但媒体键无响应

**症状**：
- 日志显示"✅ 媒体键监视器创建成功"
- 按下 F7/F8/F9 没有任何日志输出

**可能原因**：
1. **NSSystemDefined 事件未被触发**
   - macOS 版本兼容性问题
   - 其他应用占用了媒体键

2. **事件 subtype 不匹配**
   - 代码期望 `subtype == 8`，但实际值不同

**诊断步骤**：

#### 步骤 1：启用详细日志
编辑 `mediakeyservice_darwin.go`，取消注释以下行：

```objective-c
// 在 handler 中添加
} else {
    // 取消注释这行以查看所有系统事件
    char buffer[256];
    snprintf(buffer, sizeof(buffer), "🔕 忽略非媒体键事件: type=%ld, subtype=%ld", 
             (long)[event type], (long)[event subtype]);
    logToStderr(buffer);
}
```

重新编译并运行，按下 F8，查看是否有类似日志：
```
[MediaKey-C] 🔕 忽略非媒体键事件: type=14, subtype=8
```

如果看到这样的日志，说明事件被捕获了，但 type 或 subtype 不匹配。

#### 步骤 2：检查 macOS 版本
```bash
# 查看 macOS 版本
sw_vers -productVersion

# 某些旧版本可能需要不同的处理方式
```

#### 步骤 3：检查是否有其他应用占用
```bash
# 列出可能占用媒体键的进程
ps aux | grep -E "(Spotify|Apple Music|iTunes|Music)"
```

临时退出这些应用，然后重试。

#### 步骤 4：测试其他事件类型
如果 `NSSystemDefined` 不起作用，可以尝试监听原始键盘事件作为备选方案。

### 问题 3：日志显示 keyCode 不正确

**症状**：
```
[MediaKey-C] 📊 按键信息: keyCode=99, keyState=10, isKeyDown=1
[MediaKey-C] ❓ 未知媒体键: keyCode=99
```

**原因**：
- 不同的键盘布局或 macOS 版本可能导致 keyCode 不同

**解决方案**：
记录下实际的 keyCode 值，然后更新代码中的映射：

```objective-c
// 根据实际测试调整
switch (keyCode) {
    case 16:  // NX_KEYTYPE_PLAY
    case 99:  // 某些系统上的替代值
        handleMediaPlayPause();
        break;
    // ...
}
```

---

## 📊 媒体键键码参考

### 标准 macOS 媒体键键码

| 按键 | 常量 | keyCode | 说明 |
|------|------|---------|------|
| F7 | `NX_KEYTYPE_PREVIOUS` | 18 | 上一曲 |
| F8 | `NX_KEYTYPE_PLAY` | 16 | 播放/暂停 |
| F9 | `NX_KEYTYPE_NEXT` | 17 | 下一曲 |
| F10 | `NX_KEYTYPE_MUTE` | 7 | 静音 |
| F11 | `NX_KEYTYPE_SOUND_DOWN` | 1 | 音量减 |
| F12 | `NX_KEYTYPE_SOUND_UP` | 0 | 音量加 |

### 触摸栏 Mac

对于带 Touch Bar 的 MacBook：
- 物理 F7/F8/F9 键可能不存在
- 需要在 Touch Bar 上显示功能键
- 或者直接使用 Touch Bar 上的媒体控制按钮

**启用方法**：
```
系统偏好设置 > 键盘 > 键盘快捷键 > 功能键
→ 勾选"将 F1、F2 等键用作标准功能键"
```

---

## 🔍 高级调试

### 1. 监控所有系统事件

创建一个测试程序来查看所有 `NSSystemDefined` 事件：

```objective-c
// test_events.m
#import <Cocoa/Cocoa.h>

int main() {
    NSRunLoop *runLoop = [NSRunLoop currentRunLoop];
    
    id monitor = [NSEvent addGlobalMonitorForEventsMatchingMask:NSSystemDefined
        handler:^(NSEvent *event) {
            NSLog(@"Event: type=%ld, subtype=%ld, data1=0x%lx, data2=0x%lx",
                  (long)[event type],
                  (long)[event subtype],
                  (unsigned long)[event data1],
                  (unsigned long)[event data2]);
        }];
    
    [runLoop run];
    return 0;
}
```

编译并运行：
```bash
clang -framework Cocoa test_events.m -o test_events
./test_events
```

按下 F8，观察输出。

### 2. 检查应用权限状态

```bash
# 查询 TCC 数据库
sqlite3 ~/Library/Application\ Support/com.apple.TCC/TCC.db \
  "SELECT client, allowed, last_modified FROM access WHERE service='kTCCServiceAccessibility';"
```

输出示例：
```
com.apple.Terminal|1|1678886400
com.yourcompany.HaoyunMusic|1|1678886400
```

- `allowed=1` 表示已授权
- `allowed=0` 表示未授权

### 3. 强制刷新权限缓存

```bash
# 1. 退出应用
# 2. 清除权限缓存
killall cfprefsd

# 3. 重启应用
```

---

## 💡 最佳实践

### 1. 首次运行时引导用户

在应用首次启动时，主动检查并引导用户授予权限：

```go
func (app *App) OnStartup() {
    if runtime.GOOS == "darwin" {
        checkAndRequestAccessibilityPermission()
    }
}
```

### 2. 提供友好的错误提示

不要只显示"权限不足"，而是提供具体的操作步骤：

```
❌ 媒体键功能不可用

原因：应用缺少辅助功能权限

解决步骤：
1. 打开"系统偏好设置"
2. 进入"安全性与隐私" > "隐私" > "辅助功能"
3. 点击左下角锁图标解锁
4. 勾选 "Haoyun Music Player"
5. 完全退出并重启应用

[打开系统偏好设置] [复制操作指南] [取消]
```

### 3. 定期验证权限状态

在应用运行期间，定期检查权限是否被撤销：

```go
ticker := time.NewTicker(5 * time.Minute)
go func() {
    for range ticker.C {
        if !checkAccessibilityPermission() {
            showPermissionLostNotification()
        }
    }
}()
```

---

## 📞 获取帮助

如果以上方法都无法解决问题，请收集以下信息：

1. **macOS 版本**
   ```bash
   sw_vers
   ```

2. **应用启动日志**
   - 包含所有 `[MediaKey-C]` 和 `[MediaKey]` 前缀的日志

3. **权限状态**
   ```bash
   sqlite3 ~/Library/Application\ Support/com.apple.TCC/TCC.db \
     "SELECT * FROM access WHERE service='kTCCServiceAccessibility' AND client LIKE '%Haoyun%';"
   ```

4. **按键测试结果**
   - 按下 F7/F8/F9 时的完整日志输出

5. **是否有其他应用正在使用媒体键**
   ```bash
   ps aux | grep -iE "(music|spotify|itunes)"
   ```

---

## 🎓 总结

### 关键要点

1. **权限检测**：使用 `AXIsProcessTrusted()` 而非 AppleScript
2. **详细日志**：每个步骤都输出调试信息
3. **完全重启**：权限更改后必须完全退出应用
4. **事件追踪**：记录所有接收到的系统事件
5. **多方案备选**：主方案失败时有降级方案

### 常见问题速查

| 问题 | 可能原因 | 解决方案 |
|------|---------|---------|
| 始终提示需要权限 | 权限缓存未刷新 | `sudo tccutil reset Accessibility` |
| 有权限但无响应 | 事件类型不匹配 | 启用详细日志查看实际事件 |
| keyCode 不正确 | 键盘布局差异 | 记录实际值并更新映射 |
| 偶尔失效 | 其他应用占用 | 关闭冲突应用 |

---

**最后更新**: 2026-04-10  
**调试工具版本**: v2.0（含详细日志）  
**作者**: AI Assistant
