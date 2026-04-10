# macOS 媒体键支持实现指南

## 📋 概述

本项目现已完整支持 **macOS 原生系统媒体键**,包括 MacBook 键盘和外部多媒体键盘的物理按键。

---

## ✅ 已实现功能

### 支持的媒体键

| 按键 | macOS 键码 | 功能 | 对应键盘 |
|------|-----------|------|---------|
| **播放/暂停** | `0xB7` (kVK_PlayPause) | 切换播放状态 | F8 或专用媒体键 |
| **下一曲** | `0xB9` (kVK_NextTrack) | 切换到下一首歌曲 | F9 或专用媒体键 |
| **上一曲** | `0xB8` (kVK_PreviousTrack) | 切换到上一首歌曲 | F7 或专用媒体键 |

### 平台支持

- ✅ **macOS**: 使用 Carbon 框架监听系统级媒体键事件
- ✅ **Windows**: 使用 Win32 API 注册全局热键
- ⚠️ **Linux**: 暂未实现(可使用 xbindkeys 配置)

---

## 🔧 技术实现

### 架构设计

```
┌──────────────────────┐
│  MediaKeyService      │  ← 跨平台统一接口
│  (mediakeyservice.go) │
└──────────┬───────────┘
           │
     ┌─────┴──────┬──────────┐
     │            │          │
┌────▼────┐ ┌────▼────┐ ┌───▼────┐
│ macOS   │ │Windows  │ │ Linux  │
│ Darwin  │ │ Win32   │ │ (TODO) │
└─────────┘ └─────────┘ └────────┘
```

### 核心文件

1. **`backend/mediakeyservice.go`** - 跨平台抽象层
   - 定义 `MediaKeyService` 结构体
   - 提供统一的 `RegisterMediaKeys()` / `UnregisterMediaKeys()` 方法
   - 处理播放控制逻辑

2. **`backend/mediakeyservice_darwin.go`** - macOS 实现
   - 使用 CGO 调用 Carbon 框架
   - 监听 `kEventRawKeyDown` 事件
   - 捕获媒体键虚拟键码

3. **`backend/mediakeyservice_windows.go`** - Windows 实现
   - 使用 Win32 API 注册全局热键
   - 创建隐藏窗口接收 `WM_HOTKEY` 消息
   - 后台轮询消息队列

4. **`backend/mediakeyservice_linux.go`** - Linux 占位实现
   - 预留接口,暂未实现

### 关键技术点

#### macOS 实现细节

```go
// CGO 调用 Carbon 框架
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework Carbon

// 注册事件处理器
EventTypeSpec eventTypes[1];
eventTypes[0].eventClass = kEventClassKeyboard;
eventTypes[0].eventKind = kEventRawKeyDown;

InstallEventHandler(target, upp, 1, eventTypes, NULL, &g_mediaKeyHandler);
```

**优势:**
- ✅ 真正的系统级监听,无需应用获得焦点
- ✅ 低功耗,基于事件驱动
- ✅ 支持所有带媒体键的 Mac 键盘

**注意事项:**
- ⚠️ 可能需要辅助功能权限(系统偏好设置 > 安全性与隐私 > 隐私 > 辅助功能)
- ⚠️ 仅监听无修饰键的媒体键(忽略 Cmd/Ctrl/Alt/Shift 组合)

---

## 🚀 使用方法

### 1. 启动应用

```bash
# 开发模式
wails3 dev -config ./build/config.yml

# 或构建后运行
wails3 build -config ./build/config.yml
./build/bin/Haoyun\ Music\ Player
```

### 2. 测试媒体键

在应用运行时,按下以下按键:

- **MacBook 键盘**: `Fn + F7/F8/F9` 或直接按 `F7/F8/F9`(取决于键盘设置)
- **外接键盘**: 专用的媒体控制键(▶️⏸、⏭、⏮)
- **Touch Bar Mac**: Touch Bar 上的媒体控制按钮

### 3. 验证功能

查看终端日志输出:

```
🍎 正在注册 macOS 媒体键...
✅ macOS 媒体键注册成功
📝 支持的按键: 播放/暂停(F8), 下一曲(F9), 上一曲(F7)

▶️⏸️  收到媒体键:播放/暂停
⏭️  收到媒体键:下一曲
⏮️  收到媒体键:上一曲
```

---

## ⚙️ 故障排除

### 问题 1: 媒体键无响应

**可能原因:**
- 应用未获得辅助功能权限
- 键盘不支持媒体键
- 其他应用占用了媒体键

**解决方案:**

1. **检查辅助功能权限**:
   ```
   系统偏好设置 > 安全性与隐私 > 隐私 > 辅助功能
   → 确保 "Haoyun Music Player" 已勾选
   ```

2. **重启应用**:
   ```bash
   # 完全退出应用
   killall "Haoyun Music Player"
   
   # 重新启动
   wails3 dev
   ```

3. **检查键盘设置**:
   ```
   系统偏好设置 > 键盘 > 快捷键 > 媒体
   → 确保媒体键功能已启用
   ```

### 问题 2: 编译错误

**错误信息:**
```
undefined: C.register_media_keys
```

**原因:** 在非 macOS 平台编译 macOS 特定代码

**解决方案:**
- 确保在 macOS 系统上编译
- 检查构建标签是否正确(`//go:build darwin`)
- 清理缓存后重新编译:
  ```bash
  go clean -cache
  wails3 build
  ```

### 问题 3: 与其他应用冲突

**症状:** 媒体键被其他音乐播放器拦截

**解决方案:**
- 关闭其他音乐播放器(Spotify、Apple Music 等)
- 或在系统设置中调整优先级

---

## 📊 与其他控制方式的对比

| 控制方式 | 需要焦点 | 后台可用 | 实现复杂度 | 用户体验 |
|---------|---------|---------|-----------|---------|
| **原生媒体键** ✅ | ❌ | ✅ | 中等(CGO) | ⭐⭐⭐⭐⭐ |
| 应用内快捷键 | ✅ | ❌ | 简单 | ⭐⭐⭐ |
| 托盘菜单 | ✅ | ❌ | 简单 | ⭐⭐⭐⭐ |
| 主菜单栏 | ✅ | ❌ | 简单 | ⭐⭐⭐⭐ |

---

## 🔮 未来优化方向

1. **增强兼容性**:
   - 支持自定义快捷键映射
   - 添加媒体键冲突检测

2. **扩展功能**:
   - 支持音量调节媒体键
   - 支持静音键

3. **Linux 支持**:
   - 集成 D-Bus MPRIS 接口
   - 或使用 xbindkeys + xdotool

4. **用户反馈**:
   - 媒体键触发时显示系统通知
   - 在 UI 上显示当前按键状态

---

## 📚 相关文档

- [MEDIA_KEYS_GUIDE.md](./MEDIA_KEYS_GUIDE.md) - 媒体键使用指南
- [BACKEND_DESIGN.md](./BACKEND_DESIGN.md) - 后端架构设计
- [API_GUIDE.md](./API_GUIDE.md) - API 使用指南

---

## 🎯 总结

✅ **已完成:**
- macOS 原生媒体键支持(Carbon 框架)
- Windows 全局热键支持(Win32 API)
- 跨平台统一接口设计
- 与应用播放控制无缝集成

🎉 **现在你可以:**
- 使用 MacBook 键盘的 F7/F8/F9 控制播放
- 使用外接多媒体键盘的物理按键
- 在应用后台运行时仍能响应媒体键
- 享受原生的系统级媒体控制体验
