# 依赖工具自动安装 - 快速参考

## 🚀 30 秒快速开始

```bash
# 1. 编译
go build -o haoyun-music-player .

# 2. 运行
./haoyun-music-player

# 3. 点击托盘图标 → 🛠️ 依赖工具
```

---

## 📍 功能位置

**托盘菜单结构**:
```
🛠️ 依赖工具
├── ✅ FFmpeg (version x.x)    ← 已安装
├── ❌ FFprobe                  ← 未安装
│   └── 📦 安装 FFprobe        ← 点击安装
└── 🔄 重新检查所有工具         ← 手动刷新
```

---

## 🔧 常用操作

### 查看工具状态
- 点击托盘图标
- 展开 "🛠️ 依赖工具"
- 查看各工具前的图标：
  - ✅ = 已安装
  - ❌ = 未安装
  - 🔧 = 安装中
  - ⚠️ = 安装失败

### 安装缺失工具
1. 找到显示 ❌ 的工具
2. 点击展开子菜单
3. 点击 "📦 安装 XXX"
4. 等待后台安装完成
5. 菜单自动刷新

### 手动刷新状态
- 点击 "🔄 重新检查所有工具"

---

## 🐛 常见问题

### Q: 安装失败怎么办？
**A**: 
1. 查看控制台日志获取详细错误
2. 确认包管理器已安装：
   - macOS: `which brew`
   - Windows: `choco --version`
   - Linux: `which apt-get`
3. 检查网络连接
4. 尝试手动安装：`brew install ffmpeg`

### Q: 菜单不刷新？
**A**: 
1. 点击 "🔄 重新检查所有工具"
2. 或重启应用

### Q: 如何卸载工具？
**A**: 当前版本不支持，需手动卸载：
```bash
# macOS
brew uninstall ffmpeg

# Windows
choco uninstall ffmpeg

# Linux
sudo apt-get remove ffmpeg
```

### Q: 支持哪些工具？
**A**: 目前支持：
- FFmpeg（音频解码）
- FFprobe（元数据提取）

未来将支持更多工具。

---

## 💡 提示

- ✅ 安装是异步的，不会阻塞应用使用
- ✅ 状态变化会发送事件到前端
- ✅ 应用启动时自动检测
- ✅ 支持跨平台（macOS/Windows/Linux）

---

## 📚 详细文档

- [完整功能说明](./DEPENDENCY_AUTO_INSTALL.md)
- [测试清单](./DEPENDENCY_INSTALL_VERIFICATION.md)
- [实施总结](./IMPLEMENTATION_SUMMARY.md)
- [FFmpeg 指南](./FFMPEG_GUIDE.md)

---

**快捷键**: 无（通过托盘菜单操作）  
**版本**: v1.0.0  
**最后更新**: 2026-04-07