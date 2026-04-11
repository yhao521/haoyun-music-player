# 全局快捷键使用指南

## 📋 概述

Haoyun Music Player 支持两种类型的键盘控制：

1. **系统媒体键**（F7-F9）- macOS 原生支持，需要辅助功能权限
2. **自定义全局快捷键** - 跨平台支持，无需特殊权限

## ⌨️ 默认快捷键列表

### 系统媒体键 (macOS)
| 按键 | 功能 | 说明 |
|------|------|------|
| `F7` (⏮️) | 上一曲 | 跳到播放列表的上一首歌曲 |
| `F8` (⏯️) | 播放/暂停 | 切换当前播放状态 |
| `F9` (⏭️) | 下一曲 | 跳到播放列表的下一首歌曲 |

**注意**: 系统媒体键需要授予应用"辅助功能"权限才能正常工作。

### 自定义全局快捷键 (跨平台)
| 快捷键 | 功能 | 说明 |
|--------|------|------|
| `Ctrl + Shift + P` | 播放/暂停 | 切换当前播放状态 |
| `Ctrl + Shift + N` | 下一曲 | 跳到播放列表的下一首歌曲 |
| `Ctrl + Shift + B` | 上一曲 | 跳到播放列表的上一首歌曲 |
| `Ctrl + Shift + ↑` | 音量增加 | 每次增加 10% 音量 |
| `Ctrl + Shift + ↓` | 音量减少 | 每次减少 10% 音量 |

## 🚀 快速开始

### 1. 启动应用
```bash
./haoyun-music-player
```

### 2. 查看日志确认快捷键注册成功
启动后，在终端或日志文件中应该看到类似输出：
```
🎹 尝试注册系统媒体键...
🍎 Registering macOS system media keys...
✅ macOS system media keys registered successfully
✅ 系统媒体键注册成功
⌨️  注册全局快捷键...
✅ 注册快捷键: 播放/暂停 (Ctrl+Shift+P)
✅ 注册快捷键: 下一曲 (Ctrl+Shift+N)
✅ 注册快捷键: 上一曲 (Ctrl+Shift+B)
✅ 注册快捷键: 音量增加 (Ctrl+Shift+↑)
✅ 注册快捷键: 音量减少 (Ctrl+Shift+↓)
✅ 全局快捷键注册完成，共 5 个
```

### 3. 测试快捷键
- 按 `Space` 键测试播放/暂停
- 按 `Ctrl + →` 切换到下一首
- 观察日志输出确认触发

## 🔧 自定义快捷键

如需修改快捷键配置，编辑 `backend/mediakeyservice.go` 文件中的 `registerGlobalHotkeys()` 方法。

### 可用的修饰键（Modifier）
- `hotkey.ModCtrl` - Control 键
- `hotkey.ModAlt` - Alt/Option 键
- `hotkey.ModShift` - Shift 键
- `hotkey.ModCmd` - Command/Windows 键

### 可用的主键（Key）
- 字母键：`hotkey.KeyA` 到 `hotkey.KeyZ`
- 数字键：`hotkey.Key0` 到 `hotkey.Key9`
- 功能键：`hotkey.KeyF1` 到 `hotkey.KeyF12`
- 方向键：`hotkey.KeyLeft`, `hotkey.KeyRight`, `hotkey.KeyUp`, `hotkey.KeyDown`
- 其他常用键：`hotkey.KeySpace`, `hotkey.KeyEnter`, `hotkey.KeyEscape` 等

### 示例：添加新的快捷键

```go
// 在 configs 数组中添加新配置
{
    mods:    []hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift},
    key:     hotkey.KeyP,
    handler: mks.handlePlayPause,
    name:    "播放/暂停 (Ctrl+Shift+P)",
},
```

## ⚠️ 注意事项

### macOS 用户
- **系统媒体键**（F7-F9）需要辅助功能权限
- **自定义快捷键**（Ctrl+方向键等）无需特殊权限
- 如果快捷键不工作，检查系统偏好设置 → 安全性与隐私 → 隐私 → 辅助功能

### Windows/Linux 用户
- 所有快捷键都应该正常工作
- 某些系统可能会占用特定组合键（如 Win+方向键）

### 通用建议
1. 避免与系统或其他应用的快捷键冲突
2. 处理函数应快速返回，耗时操作使用 goroutine
3. 应用退出时会自动注销所有快捷键
4. 查看日志文件排查问题：`~/Library/Application Support/haoyun-music-player/logs/`

## 🐛 故障排除

### 快捷键不响应
1. 检查应用是否在前台运行
2. 查看日志确认快捷键注册成功
3. 检查是否有其他应用占用了相同快捷键
4. macOS 用户检查辅助功能权限

### 音量调整无效
- 确保有歌曲正在播放
- 检查系统音量是否被静音
- 查看日志是否有错误信息

### 编译错误
如果遇到编译错误，确保依赖已正确安装：
```bash
go mod tidy
go build
```

## 📝 技术实现

- **库**: `golang.design/x/hotkey v0.4.1`
- **实现位置**: `backend/mediakeyservice.go`
- **注册时机**: 应用初始化时（`Init()` 方法）
- **注销时机**: 应用关闭时（`UnregisterMediaKeys()` 方法）

## 🔗 相关文档

- [媒体键调试指南](MEDIAKEY_DEBUG_GUIDE.md)
- [媒体键修复记录](MEDIAKEY_FIX.md)
- [媒体键调研报告](MEDIAKEY_RESEARCH.md)

---

**最后更新**: 2026-04-11  
**维护者**: YHao521
