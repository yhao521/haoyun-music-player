# 系统媒体键功能 - 故障排查指南

## ⚠️ 重要提示

如果你已经授予了辅助功能权限，但系统媒体键（F7/F8/F9）仍然不工作，请按照以下步骤操作：

## 🔧 完整重置步骤

### 1. 完全退出应用
```bash
# 使用 Cmd+Q 完全退出，不是关闭窗口
# 或者在终端中：
pkill -f haoyun-music-player
```

### 2. 重置辅助功能权限

**方法 A: 通过系统偏好设置**
1. 打开 **系统偏好设置** > **安全性与隐私** > **隐私**
2. 选择左侧的 **"辅助功能"**
3. 点击左下角的锁图标解锁
4. 找到 `haoyun-music-player`，选中后点击 **"-"** 移除
5. 点击 **"+"** 重新添加应用
6. 确保勾选了该应用
7. 关闭系统偏好设置

**方法 B: 通过命令行（需要管理员权限）**
```bash
# 重置所有应用的辅助功能权限（谨慎使用）
sudo tccutil reset Accessibility
```

### 3. 重新启动应用并监控日志

**在一个终端窗口中启动实时监控：**
```bash
cd /Users/yanghao/storage/code_projects/goProjects/haoyun-music-player
./test_media_keys.sh
```

**在另一个终端窗口中启动应用：**
```bash
cd /Users/yanghao/storage/code_projects/goProjects/haoyun-music-player
./haoyun-music-player
```

### 4. 验证注册成功

查看日志输出，应该看到：
```
🍎 Registering macOS system media keys...
🔧 setupMediaKeys called
📡 Setting up NSEvent monitor for media keys...
✅ macOS system media keys registered successfully
💡 Note: You may need to grant Accessibility permissions in System Preferences
```

### 5. 测试媒体键

按下 F7/F8/F9，实时日志应该显示：
```
📨 Received system event: type=14, subtype=8
🎹 Media key detected: keyCode=16, keyState=1
▶️ Play/Pause key pressed
▶️⏸️  macOS media key: Play/Pause
```

## ❓ 常见问题

### Q1: 完全没有看到 "setupMediaKeys called" 日志

**原因**：应用可能使用了旧的编译版本

**解决**：
```bash
# 清理并重新编译
cd /Users/yanghao/storage/code_projects/goProjects/haoyun-music-player
go clean -cache
go build
./haoyun-music-player
```

### Q2: 看到注册成功日志，但按键无响应

**可能原因**：
- 权限未真正生效
- 需要重启系统（极端情况）

**解决**：
1. 确认系统偏好设置中已勾选应用
2. 尝试注销并重新登录 macOS
3. 作为最后手段，重启 Mac

### Q3: 日志显示 "Failed to create event monitor"

**原因**：权限被拒绝或系统限制

**解决**：
1. 检查是否有其他应用正在监听全局事件
2. 尝试以管理员身份运行（不推荐）
3. 使用备选方案：自定义快捷键

## 💡 备选方案：自定义全局快捷键

如果系统媒体键始终无法工作，可以使用自定义快捷键（无需特殊权限）：

| 快捷键 | 功能 |
|--------|------|
| `Ctrl + Shift + P` | 播放/暂停 |
| `Ctrl + Shift + N` | 下一曲 |
| `Ctrl + Shift + B` | 上一曲 |
| `Ctrl + Shift + ↑` | 音量增加 10% |
| `Ctrl + Shift + ↓` | 音量减少 10% |

这些快捷键会自动注册，无论系统媒体键是否工作。

## 📊 诊断脚本

运行完整的诊断：
```bash
./test_media_keys.sh
```

这会检查：
- ✅ C 函数是否被调用
- ✅ NSEvent 监听器是否设置
- ✅ 是否捕获到系统事件
- ✅ 媒体键按下次数统计

## 🆘 仍然无法解决？

如果以上步骤都无法解决问题，请提供以下信息：

1. **macOS 版本**：
   ```bash
   sw_vers
   ```

2. **应用日志**（最近100行）：
   ```bash
   tail -100 ~/.haoyun-music/runtime/logs/app-*.log
   ```

3. **辅助功能权限状态**：
   ```bash
   ls -la ~/Library/Application\ Support/com.apple.TCC/
   ```

4. **调试输出**：
   ```bash
   ./test_media_keys.sh 2>&1 | tee debug_output.txt
   ```
