# macOS 媒体键快速参考

## 🎹 支持的按键

| 功能 | macOS 键码 | MacBook 键盘 | 外接键盘 |
|------|-----------|-------------|---------|
| **播放/暂停** | `0xB7` | F8 或 Fn+F8 | ▶️⏸ 专用键 |
| **下一曲** | `0xB9` | F9 或 Fn+F9 | ⏭ 专用键 |
| **上一曲** | `0xB8` | F7 或 Fn+F7 | ⏮ 专用键 |

---

## ⚡ 快速测试

```bash
# 1. 启动应用
wails3 dev

# 2. 观察日志
# 应该看到: "✅ macOS 媒体键注册成功"

# 3. 按下媒体键
# F7/F8/F9 或外接键盘的媒体键

# 4. 验证响应
# 日志显示: "▶️⏸️  收到媒体键:播放/暂停"
```

---

## 🔧 故障排除

### 媒体键无响应?

```bash
# 1. 检查辅助功能权限
open "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility"

# 2. 确保应用已勾选

# 3. 重启应用
killall "Haoyun Music Player"
wails3 dev
```

### 编译错误?

```bash
# 清理缓存
go clean -cache

# 重新构建
wails3 build
```

---

## 📁 相关文件

- `backend/mediakeyservice.go` - 跨平台接口
- `backend/mediakeyservice_darwin.go` - macOS 实现
- `backend/mediakeyservice_windows.go` - Windows 实现
- `app_init.go` - 服务初始化

---

## 💡 提示

- ✅ 无需应用获得焦点即可响应
- ✅ 支持后台运行时的媒体控制
- ⚠️ 首次使用可能需要辅助功能权限
- 🎯 与托盘菜单、快捷键协同工作
