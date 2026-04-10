# 依赖工具自动安装功能 - 验证清单

## ✅ 编译状态

- [x] **代码编译成功**：无错误，仅有 macOS 版本警告（可忽略）
- [x] **可执行文件生成**：`/tmp/haoyun-music-player-final` (19MB)
- [x] **所有文件已保存**：main.go, dependency_manager.go

## 📋 功能测试清单

### 1. 应用启动检测

```bash
# 运行应用
./haoyun-music-player

# 预期日志输出：
# 🔧 开始检查依赖工具...
# ✅ FFmpeg: 已安装 (版本信息)
# === 依赖工具状态 ===
# ✅ FFmpeg: 已就绪
```

**验证点**：
- [ ] 应用启动时自动检测 FFmpeg
- [ ] 控制台显示检测结果
- [ ] 如果缺失，显示安装提示

### 2. 托盘菜单验证

**步骤**：
1. 启动应用
2. 点击系统托盘图标
3. 查找 "🛠️ 依赖工具" 菜单项

**预期结果**：
- [ ] 菜单项存在且可点击
- [ ] 展开后显示 FFmpeg 状态
- [ ] 已安装显示：✅ FFmpeg (version x.x...)
- [ ] 未安装显示：❌ FFmpeg → 📦 安装 FFmpeg

### 3. 安装功能测试

**前置条件**：卸载 FFmpeg（可选测试）

```bash
# macOS 卸载测试
brew uninstall ffmpeg

# 重新启动应用
./haoyun-music-player
```

**测试步骤**：
1. 打开托盘菜单 → 🛠️ 依赖工具
2. 点击 "📦 安装 FFmpeg"
3. 观察通知和日志

**预期结果**：
- [ ] 显示通知："正在后台安装 FFmpeg，请稍候..."
- [ ] 日志显示：`📦 用户请求安装 FFmpeg`
- [ ] 后台执行 `brew install ffmpeg`
- [ ] 安装完成后显示成功通知
- [ ] 菜单自动刷新，显示 ✅ FFmpeg

### 4. 状态变化回调测试

**测试场景**：
1. 手动触发安装
2. 观察前端是否收到事件

**预期事件**：
```javascript
// 前端应收到以下事件
EventsOn('dependencyStatusChanged', (data) => {
  console.log(data.tool)    // "ffmpeg"
  console.log(data.status)  // "installing" -> "installed"
  console.log(data.message) // 状态描述
})
```

**验证点**：
- [ ] 安装开始时发送 "installing" 状态
- [ ] 安装完成时发送 "installed" 状态
- [ ] 安装失败时发送 "failed" 状态

### 5. 跨平台兼容性测试

#### macOS 测试
```bash
# 确认 Homebrew 可用
which brew

# 测试安装
./test_ffmpeg.sh
```

**预期**：
- [ ] 检测到 Homebrew
- [ ] 使用 `brew install ffmpeg` 安装
- [ ] 安装成功后可正常使用

#### Windows 测试（需 Windows 环境）
```powershell
# 确认 Chocolatey 或 Scoop 可用
choco --version
# 或
scoop --version

# 运行测试脚本
.\test_ffmpeg.bat
```

**预期**：
- [ ] 检测到包管理器
- [ ] 使用相应命令安装
- [ ] 安装成功后可正常使用

#### Linux 测试（需 Linux 环境）
```bash
# Ubuntu/Debian
sudo apt-get update && sudo apt-get install ffmpeg

# 或使用脚本
./test_ffmpeg.sh
```

**预期**：
- [ ] 检测到 apt/dnf/pacman
- [ ] 使用相应命令安装
- [ ] 安装成功后可正常使用

### 6. 错误处理测试

#### 测试场景 1: 包管理器未安装

**操作**：
- macOS: 卸载 Homebrew
- Windows: 未安装 Chocolatey/Scoop
- Linux: 无 sudo 权限

**预期**：
- [ ] 显示友好错误提示
- [ ] 提供手动安装指南链接
- [ ] 不崩溃，优雅降级

#### 测试场景 2: 网络问题

**操作**：
- 断开网络连接
- 尝试安装 FFmpeg

**预期**：
- [ ] 超时后显示错误
- [ ] 状态变为 "安装失败"
- [ ] 允许重试

#### 测试场景 3: FFmpeg 已在 PATH 但版本过旧

**操作**：
- 安装旧版本 FFmpeg
- 运行应用

**预期**：
- [ ] 检测到 FFmpeg
- [ ] 显示版本信息
- [ ] （可选）提示更新

### 7. 性能测试

**测试指标**：
- [ ] 启动检测时间 < 2 秒
- [ ] 菜单构建时间 < 500ms
- [ ] 安装过程不阻塞 UI
- [ ] 内存占用无明显增长

**压力测试**：
```bash
# 快速多次点击安装按钮
# 预期：只执行一次安装，后续请求被忽略或排队
```

### 8. 用户体验测试

**主观评估**：
- [ ] 菜单项图标清晰易懂（✅ ❌ 🔧 ⚠️）
- [ ] 安装提示友好明确
- [ ] 通知及时准确
- [ ] 状态变化流畅自然
- [ ] 无卡顿或闪烁

## 🔍 代码审查清单

### main.go
- [x] 前向声明解决闭包问题
- [x] depManager 回调在 rebuildTrayMenu 定义后设置
- [x] toolsMenuItem 正确添加到托盘菜单
- [x] buildToolsMenu 函数完整实现
- [x] rebuildTrayMenu 包含 buildToolsMenu 调用

### dependency_manager.go
- [x] ToolInfo 结构体定义完整
- [x] CheckAllTools 正确检测工具
- [x] InstallTool 异步执行
- [x] SetCallback 机制正常工作
- [x] GetInstallSummary 返回友好摘要
- [x] 跨平台安装逻辑完整

## 🐛 已知限制

1. **FFprobe 检测**：当前将 FFprobe 作为独立工具检测，实际随 FFmpeg 一起安装
2. **安装进度**：仅显示"安装中"，无实时进度条
3. **批量安装**：不支持一键安装所有缺失工具
4. **卸载功能**：暂不提供工具卸载选项

## 📊 测试结果汇总

| 测试项 | 状态 | 备注 |
|--------|------|------|
| 编译通过 | ✅ | 无错误 |
| 启动检测 | ⏳ | 待运行时验证 |
| 托盘菜单 | ⏳ | 待运行时验证 |
| 安装功能 | ⏳ | 待运行时验证 |
| 状态回调 | ⏳ | 待前端集成验证 |
| macOS 兼容 | ⏳ | 待测试 |
| Windows 兼容 | ⏳ | 需 Windows 环境 |
| Linux 兼容 | ⏳ | 需 Linux 环境 |
| 错误处理 | ⏳ | 待边界测试 |
| 性能表现 | ⏳ | 待压力测试 |

## 🎯 下一步行动

### 立即执行
1. **运行应用测试**：
   ```bash
   cd /Users/yanghao/storage/code_projects/goProjects/haoyun-music-player
   wails3 dev
   # 或
   ./haoyun-music-player
   ```

2. **验证托盘菜单**：
   - 检查 "🛠️ 依赖工具" 菜单项
   - 测试状态显示
   - 测试安装功能

3. **前端集成**（如需要）：
   - 监听 `dependencyStatusChanged` 事件
   - 显示友好的安装通知
   - 提供手动检查入口

### 后续优化
1. 添加安装进度实时显示
2. 支持批量安装所有缺失工具
3. 添加工具版本检查和更新提示
4. 完善错误恢复机制

## 📝 反馈收集

测试完成后，请记录以下信息：

**功能是否正常**：
- [ ] 是
- [ ] 否（请说明：__________）

**遇到的问题**：
_________________________________

**改进建议**：
_________________________________

**整体评分**：⭐⭐⭐⭐⭐ (1-5星)

---

**测试人员**: ______________  
**测试日期**: ______________  
**测试环境**: macOS / Windows / Linux  
**应用版本**: ______________