# 系统媒体键功能

## ✅ 已实现功能

Haoyun Music Player 现已支持 **macOS 系统媒体键**监听！

### 支持的按键

| 按键 | 图标 | 功能 |
|------|------|------|
| F7 | ⏮️ | 上一曲 |
| F8 | ⏯️ | 播放/暂停 |
| F9 | ⏭️ | 下一曲 |

## 🔧 配置要求

### macOS 辅助功能权限

系统媒体键需要授予应用"辅助功能"权限才能正常工作。

#### 如何授予权限：

1. 打开 **系统偏好设置** > **安全性与隐私** > **隐私**
2. 在左侧列表中选择 **"辅助功能"**
3. 点击左下角的锁图标解锁（需要管理员密码）
4. 点击 "+" 按钮添加应用
5. 找到并选择 `haoyun-music-player`
6. 确保勾选了该应用
7. **重启应用**使权限生效

#### 验证权限状态

运行测试脚本检查权限：
```bash
./test_media_keys.sh
```

## 🎹 备选方案：自定义全局快捷键

如果不想授予辅助功能权限，可以使用自定义全局快捷键（无需特殊权限）：

| 快捷键 | 功能 |
|--------|------|
| `Ctrl + Shift + P` | 播放/暂停 |
| `Ctrl + Shift + N` | 下一曲 |
| `Ctrl + Shift + B` | 上一曲 |
| `Ctrl + Shift + ↑` | 音量增加 10% |
| `Ctrl + Shift + ↓` | 音量减少 10% |

## 📝 技术实现

- **方案**：CGO + NSEvent 全局事件监听
- **框架**：AppKit (NSSystemDefinedMask)
- **文件**：`backend/mediakeyservice_darwin.go`
- **架构**：混合方案（系统媒体键 + 自定义快捷键）

## 🐛 故障排除

### 问题：按下媒体键没有反应

**解决方案**：
1. 确认已授予辅助功能权限
2. 重启应用
3. 查看日志文件确认按键是否被捕获：
   ```bash
   tail -f ~/.haoyun-music/runtime/logs/app-*.log
   ```
4. 应该看到类似输出：
   ```
   ▶️⏸️  macOS media key: Play/Pause
   ⏭️  macOS media key: Next Track
   ⏮️  macOS media key: Previous Track
   ```

### 问题：无法授予权限

**解决方案**：
1. 从"辅助功能"列表中移除应用
2. 重新添加应用
3. 确保使用最新构建的应用版本
4. 重启 Mac（极端情况下可能需要）

## 📚 相关文档

- [HOTKEY_GUIDE.md](HOTKEY_GUIDE.md) - 完整的快捷键使用指南
- [MEDIAKEY_DEBUG_GUIDE.md](MEDIAKEY_DEBUG_GUIDE.md) - 媒体键调试指南
