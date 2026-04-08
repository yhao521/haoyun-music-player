# 🎵 Haoyun Music Player

<div align="center">

![Banner](https://img.shields.io/badge/🎵-Haoyun%20Music%20Player-blue?style=for-the-badge)

**简约而不简单的跨平台音乐播放器**

*基于 Wails 3 + Vue 3 打造，为您带来极致的音乐体验*

[![Wails](https://img.shields.io/badge/Wails-v3-blue)](https://wails.io/)
[![Vue](https://img.shields.io/badge/Vue-3-green)](https://vuejs.org/)
[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8)](https://golang.org/)
[![Platform](https://img.shields.io/badge/Platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey)](#-多平台支持)
[![License](https://img.shields.io/badge/License-Apache%202.0-yellow)](./LICENSE)

[✨ 核心特性](#-核心特性) • [🖼️ 界面预览](#️-界面预览) • [🚀 快速开始](#-快速开始) • [📥 下载安装](#-下载安装) • [📚 完整文档](#-完整文档)

</div>

---

## 🌟 为什么选择 Haoyun Music Player？

> 在数字音乐时代，我们需要的不仅仅是一个播放器，而是一个懂您的音乐伙伴。

**Haoyun Music Player** 是一款专为音乐爱好者打造的现代化桌面应用。它融合了极简的设计理念与强大的功能，让您能够：

- 🎯 **专注音乐本身** - 清爽的界面设计，无干扰的播放体验
- ⚡ **极速响应** - Go 语言驱动，毫秒级启动和切换
- 🔊 **全格式支持** - 通过 FFmpeg 支持几乎所有音频格式
- 💾 **智能管理** - 自动扫描、索引和管理您的音乐库
- 🌍 **多语言界面** - 中英文无缝切换，全球用户友好

---

## ✨ 核心特性

### 🎨 现代化设计

- **毛玻璃特效** - macOS 风格的半透明界面，优雅美观
- **渐变主题** - 精心调制的色彩方案，视觉舒适
- **响应式布局** - 自适应不同窗口大小，始终保持最佳显示

### 🎵 专业音频引擎

```
支持的音频格式：
✅ MP3    ✅ WAV    ✅ FLAC   ✅ AAC
✅ M4A    ✅ OGG    ✅ WMA    ✅ APE
✅ Opus   ✅ 以及更多...
```

- **FFmpeg 集成** - 业界领先的音频解码引擎
- **断点续播** - 暂停后精确恢复到原位置
- **多播放模式** - 顺序 / 循环 / 随机 / 单曲循环

### 📂 智能音乐库管理

- **多库支持** - 同时管理多个音乐文件夹
- **后台扫描** - 异步索引，不阻塞主界面
- **实时搜索** - 按标题、艺术家、专辑快速查找
- **分页浏览** - 大型音乐库也能流畅展示（每页 20-200 首）

### ⌨️ 高效操作体验

| 快捷键 | 功能 |
|--------|------|
| `Space` | 播放 / 暂停 |
| `Cmd/Ctrl + [` | 上一曲 |
| `Cmd/Ctrl + ]` | 下一曲 |
| `Cmd/Ctrl + F` | 打开浏览窗口 |
| `Cmd/Ctrl + H` | 打开喜爱音乐 |
| `Cmd/Ctrl + S` | 打开设置 |

### 🌐 系统深度集成

- **系统托盘** - 最小化到托盘，随时控制播放
- **菜单栏控制** - macOS 菜单栏快速访问
- **多窗口管理** - 主播放器、浏览视图、喜爱音乐、设置独立窗口
- **配置持久化** - 所有设置自动保存，重启后恢复

### 📊 数据统计与分析

- **播放历史** - 记录每首歌的播放次数
- **喜爱排行** - 按播放频率智能排序
- **库统计信息** - 歌曲数量、总时长、占用空间一目了然

### 🌍 国际化支持

- 🇨🇳 简体中文
- 🇺🇸 English
- 实时切换，无需重启（部分界面需重启生效）

---

## 🖼️ 界面预览

### 主播放器界面

```
┌──────────────────────────────────────┐
│  🎵 Haoyun Music Player              │
├──────────────────────────────────────┤
│                                      │
│    ┌──────────┐                      │
│    │          │  夜曲.mp3             │
│    │  🎵 封面  │  周杰伦               │
│    │          │  十一月的萧邦          │
│    └──────────┘                      │
│                                      │
│    1:23  ━━━━━━━○━━━━━━━  3:45      │
│                                      │
│         ⏮    ▶️    ⏭                │
│                                      │
│    🔊 ━━━━━━━━━━━━━━  80%            │
│                                      │
│    ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │
│    播放列表 (15)                      │
│    1. ▶ 夜曲.mp3                     │
│    2.   发如雪.mp3                   │
│    3.   青花瓷.mp3                   │
│    ...                               │
└──────────────────────────────────────┘
```

### 浏览音乐库

```
┌─────────────────────────────────────────────────────┐
│  🎵 浏览音乐库                                       │
├─────────────────────────────────────────────────────┤
│  音乐库: [我的收藏 ▼]  🔍 搜索歌曲、艺术家、专辑...  │
├─────────────────────────────────────────────────────┤
│  📁 我的收藏 | 🎵 1,250 首 | ⏱️ 3天 12小时          │
│  💾 8.5 GB | 📂 /Users/yanghao/Music                │
├─────────────────────────────────────────────────────┤
│ #  │ 标题            │ 艺术家   │ 专辑      │ 时长  │
│ ─────────────────────────────────────────────────── │
│ 1  │ 夜曲            │ 周杰伦   │ 十一月的..│ 3:45  │
│ 2  │ 晴天            │ 周杰伦   │ 叶惠美    │ 4:29  │
│ 3  │ 稻香            │ 周杰伦   │ 魔杰座    │ 3:43  │
│ ...│ ...             │ ...     │ ...      │ ...   │
├─────────────────────────────────────────────────────┤
│ 第 1-50 首，共 1,250 首  ⏮ ◀ 1 2 3 ... 25 ▶ ⏭     │
│ 每页显示: [50 ▼]                                     │
└─────────────────────────────────────────────────────┘
```

### 系统托盘菜单

```
🎵 正在播放：夜曲 - 周杰伦
─────────────────────────────
⏯ 播放/暂停          Space
⏮ 上一曲        Cmd+[
⏭ 下一曲        Cmd+]
─────────────────────────────
📂 浏览歌曲       Cmd+F
❤️ 喜爱音乐       Cmd+H
🔄 播放模式
📚 音乐库管理
🛠️ 依赖工具
⬇️ 下载音乐       Cmd+D
─────────────────────────────
⚙️ 设置           Cmd+S
🪟 显示主窗口
─────────────────────────────
v1.0.0
❌ 退出
```

---

## 🚀 快速开始

### 前置要求

确保您的系统已安装以下软件：

- **Go 1.25+** - [下载](https://golang.org/dl/)
- **Node.js 18+** - [下载](https://nodejs.org/)
- **FFmpeg** - 音频解码引擎
  - macOS: `brew install ffmpeg`
  - Ubuntu: `sudo apt-get install ffmpeg`
  - Windows: `choco install ffmpeg`

### 三步启动

#### 1️⃣ 克隆项目

```bash
git clone https://github.com/yhao521/haoyun-music-player.git
cd haoyun-music-player
```

#### 2️⃣ 安装依赖

```bash
# Go 依赖
go mod tidy

# 前端依赖
cd frontend && npm install && cd ..
```

#### 3️⃣ 运行应用

```bash
# 开发模式（热重载）
wails3 dev -config ./build/config.yml

# 或直接运行
go run .
```

🎉 完成！您现在可以看到 Haoyun Music Player 的主界面了。

### 生产构建

```bash
# 构建适用于当前平台的安装包
wails3 build

# 输出目录：build/bin/
```

---

## 📥 下载安装

### 预编译版本

前往 [Releases 页面](https://github.com/yhao521/haoyun-music-player/releases) 下载最新版本的预编译安装包。

#### macOS

```bash
# Homebrew (推荐)
brew install --cask haoyun-music-player

# 或手动下载 DMG 文件
# 支持 Intel (AMD64) 和 Apple Silicon (ARM64)
```

#### Windows

```powershell
# Chocolatey
choco install haoyun-music-player

# 或下载 .exe 安装程序
# 支持 64 位系统 (AMD64/ARM64)
```

#### Linux

```bash
# Ubuntu/Debian
sudo dpkg -i haoyun-music-player_*.deb

# Fedora/RHEL
sudo rpm -i haoyun-music-player-*.rpm

# Arch Linux
yay -S haoyun-music-player
```

### 从源码构建

参见 [快速开始](#-快速开始) 章节。

---

## 📚 完整文档

### 用户指南

- [📖 快速开始](./QUICKSTART.md) - 5 分钟上手教程
- [⌨️ 键盘快捷键](./KEYBOARD_SHORTCUTS.md) - 完整快捷键参考
- [🌍 多语言支持](./I18N_QUICKSTART.md) - 国际化使用指南
- [🔧 FFmpeg 安装](./FFMPEG_GUIDE.md) - 音频解码引擎配置
- [❓ 故障排除](./TROUBLESHOOTING.md) - 常见问题解答

### 开发者文档

- [🏗️ 架构设计](./BACKEND_DESIGN.md) - MVC 架构详解
- [🔌 API 指南](./API_GUIDE.md) - 后端接口文档
- [📦 依赖管理](./DEPENDENCY_AUTO_INSTALL.md) - 自动化工具安装
- [🚀 CI/CD 流程](./GITHUB_ACTIONS_QUICKSTART.md) - 自动化发布指南
- [📝 实现细节](./IMPLEMENTATION.md) - 技术实现说明

### 特性文档

- [🎵 新功能介绍](./NEW_FEATURES.md) - 播放历史、歌词、封面
- [❤️ 喜爱音乐功能](./FAVORITES_FEATURE.md) - 智能排序算法
- [📄 音乐信息显示](./MUSIC_INFO_DISPLAY.md) - 元数据展示
- [🎹 媒体键支持](./MEDIA_KEYS_GUIDE.md) - 系统级控制

---

## 🛠️ 技术栈

### 后端核心技术

```
Go 1.25+          → 高性能后端逻辑
Wails v3          → 跨平台桌面框架
Oto v3            → 原生音频播放引擎
FFmpeg            → 全格式音频解码
SQLite            → 本地数据存储
```

### 前端核心技术

```
Vue 3             → 响应式 UI 框架
TypeScript        → 类型安全开发
Vite              → 极速构建工具
CSS3              → 现代化样式设计
```

### 架构模式

```
MVC + Facade      → 清晰的分层架构
事件驱动          → 前后端松耦合通信
单例模式          → 全局资源管理
观察者模式        → 状态同步机制
```

---

## 🌐 多平台支持

| 平台 | 支持状态 | 最低版本 | 架构支持 |
|------|---------|---------|---------|
| **macOS** | ✅ 完全支持 | macOS 10.15+ | Intel & Apple Silicon |
| **Windows** | ✅ 完全支持 | Windows 10+ | x86_64 & ARM64 |
| **Linux** | ✅ 完全支持 | Ubuntu 20.04+ | x86_64 & ARM64 |

---

## 🤝 参与贡献

我们欢迎任何形式的贡献！

### 贡献方式

- 🐛 **报告 Bug** - [提交 Issue](https://github.com/yhao521/haoyun-music-player/issues)
- 💡 **功能建议** - 分享您的创意想法
- 📝 **改进文档** - 帮助我们完善文档
- 🔧 **提交代码** - Fork → 修改 → Pull Request

### 开发规范

- 遵循 [Go 代码规范](https://go.dev/doc/effective_go)
- 使用 TypeScript 严格模式
- 编写必要的单元测试
- 保持向后兼容性

详见 [贡献指南](./CONTRIBUTING.md)（待创建）。

---

## 📄 开源协议

本项目采用 **Apache License 2.0** 开源协议。

您可以自由地使用、修改和分发本软件，但需要保留版权声明和许可证声明。

详见 [LICENSE](./LICENSE) 文件。

---

## 🙏 致谢

感谢以下优秀的开源项目为本项目提供支持：

- [Wails](https://wails.io/) - 卓越的 Go 桌面应用框架
- [Vue.js](https://vuejs.org/) - 渐进式 JavaScript 框架
- [FFmpeg](https://ffmpeg.org/) - 完整的音视频解决方案
- [Oto](https://github.com/ebitengine/oto) - 跨平台音频播放库
- [Vite](https://vitejs.dev/) - 下一代前端构建工具

---

## 📬 联系我们

- 🌐 **GitHub**: [yhao521/haoyun-music-player](https://github.com/yhao521/haoyun-music-player)
- 📧 **Email**: your.email@example.com
- 💬 **Issues**: [问题反馈](https://github.com/yhao521/haoyun-music-player/issues)
- 📖 **Wiki**: [项目维基](https://github.com/yhao521/haoyun-music-player/wiki)

---

<div align="center">

**Made with ❤️ by YHao521**

🎵 *Enjoy Your Music, Enjoy Your Life*

[⭐ Star this project](https://github.com/yhao521/haoyun-music-player) • [🍴 Fork it](https://github.com/yhao521/haoyun-music-player/fork) • [📥 Download](https://github.com/yhao521/haoyun-music-player/releases)

</div>
