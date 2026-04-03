# 🎵 Haoyun Music Player

<div align="center">

![Wails](https://img.shields.io/badge/Wails-v3-blue)
![Vue](https://img.shields.io/badge/Vue-3-green)
![TypeScript](https://img.shields.io/badge/TypeScript-4.9-blue)
![Go](https://img.shields.io/badge/Go-1.25+-00ADD8)
![Platform](https://img.shields.io/badge/Platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey)

**基于 Wails 3 + Vue 3 + TypeScript 的现代化菜单栏音乐播放器**

[快速开始](./QUICKSTART.md) • [实现文档](./IMPLEMENTATION.md) • [主文档](./README.md)

</div>

## ✨ 特性亮点

- 🎨 **现代 UI** - 精美的渐变设计和毛玻璃特效
- ⚡ **高性能** - Go + Vue 的原生性能体验
- 🔧 **可扩展** - 清晰的架构，易于添加新功能
- 📱 **跨平台** - 支持 macOS、Windows、Linux
- 🎯 **类型安全** - 完整的 TypeScript 类型定义
- 📋 **播放列表** - 完善的播放列表管理
- 🎚️ **音量控制** - 精细的音量调节
- 💻 **菜单集成** - 系统菜单栏快速访问

## 🖼️ 界面预览

```
┌─────────────────────────────────┐
│  🎵 Haoyun Music                │
├─────────────────────────────────┤
│                                 │
│  ┌──────┐                       │
│  │ 🎵   │  未播放音乐            │
│  └──────┘  未知艺术家            │
│                                 │
│  0:00              0:00         │
│  ━━━━━━━━━━━━━━━━━━━━━━━        │
│                                 │
│      ⏮   ▶️   ⏭                │
│                                 │
│  🔊 ━━━━━━━━━━━━                │
│                                 │
│  📂 打开文件                     │
│                                 │
│  播放列表 (0)                    │
│  ─────────────────────────────  │
│                                 │
└─────────────────────────────────┘
```

## 🛠️ 技术栈

### 后端
- **Go** - 核心业务逻辑
- **Wails v3** - 桌面应用框架
- **Oto** (可选) - 音频播放引擎

### 前端
- **Vue 3** - 渐进式 JavaScript 框架
- **TypeScript** - 类型安全的 JavaScript 超集
- **Vite** - 下一代前端构建工具
- **@wailsio/runtime** - Wails 运行时

## 🚀 快速开始

### 前置要求

- Go 1.25+
- Node.js 18+
- Wails v3 CLI（可选）

### 安装依赖

```bash
# 安装 Go 依赖
go mod tidy

# 安装前端依赖
cd frontend
npm install
cd ..
```

### 运行应用

```bash
# 使用 Wails CLI
wails3 dev -config ./build/config.yml

# 或使用启动脚本（Unix/Linux/macOS）
./start.sh

# Windows
start.bat
```

### 生产构建

```bash
# macOS
wails3 build -platform darwin

# Windows
wails3 build -platform windows

# Linux
wails3 build -platform linux
```

## 📁 项目结构

```
haoyun-music-player/
├── main.go                 # Go 主入口
├── greetservice.go         # 音乐服务实现
├── go.mod                  # Go 模块配置
├── Taskfile.yml           # 构建任务配置
├── README.md              # 项目说明
├── QUICKSTART.md          # 快速开始指南
├── IMPLEMENTATION.md      # 实现文档
├── start.sh               # Unix 启动脚本
├── start.bat              # Windows 启动脚本
├── .gitignore             # Git 忽略文件
│
└── frontend/              # 前端代码
    ├── src/
    │   ├── App.vue       # 主应用组件
    │   ├── main.ts       # Vue 入口
    │   └── vite-env.d.ts # TypeScript 定义
    ├── public/
    │   └── style.css     # 全局样式
    ├── package.json      # 前端依赖
    ├── vite.config.ts    # Vite 配置
    └── tsconfig.json     # TypeScript 配置
```

## 📖 功能清单

### ✅ 已实现

- [x] 基础 UI 界面
- [x] 播放/暂停控制
- [x] 上一首/下一首切换
- [x] 进度条显示和拖拽
- [x] 音量调节滑块
- [x] 播放列表管理
- [x] 文件选择器集成
- [x] 前后端事件通信
- [x] 状态同步机制
- [x] 菜单栏基础结构

### ⏳ 计划中

- [ ] 真实音频播放核心
- [ ] 全局快捷键支持
- [ ] 歌词显示
- [ ] 专辑封面提取
- [ ] 播放模式切换
- [ ] 均衡器调节
- [ ] 系统托盘图标
- [ ] 媒体键支持

## 🎯 如何添加真实音频播放

当前版本使用模拟播放功能。要添加真实的音频播放：

### 1. 安装音频库

```bash
go get github.com/hajimehoshi/oto/v2
```

### 2. 实现播放逻辑

在 `greetservice.go` 中：

```go
import "github.com/hajimehoshi/oto/v2"

type MusicService struct {
    player *oto.Player
    // ... 其他字段
}

func (m *MusicService) Initialize() error {
    var err error
    m.player, err = oto.NewPlayer(oto.PlayerOptions{
        SampleRate:   44100,
        ChannelCount: 2,
        Format:       oto.FormatSignedInt16LE,
    })
    return err
}

// 实现 Play(), Pause() 等方法
```

详细示例请参考 [IMPLEMENTATION.md](./IMPLEMENTATION.md)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

### 开发流程

1. Fork 本项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

### 开发建议

- 遵循 Go 代码规范
- 使用 TypeScript 类型检查
- 编写必要的测试
- 更新相关文档

## 📄 许可证

MIT License

## 🙏 致谢

感谢以下优秀的开源项目：

- [Wails](https://wails.io/) - 卓越的 Go 桌面应用框架
- [Vue.js](https://vuejs.org/) - 渐进式 JavaScript 框架
- [TypeScript](https://www.typescriptlang.org/) - JavaScript 的超集
- [Vite](https://vitejs.dev/) - 极速的前端构建工具
- [Oto](https://github.com/hajimehoshi/oto) - 跨平台音频播放库

## 📬 联系方式

- 📧 Email: your.email@example.com
- 💬 Issues: [GitHub Issues](https://github.com/yourusername/haoyun-music-player/issues)
- 📖 Docs: [实现文档](./IMPLEMENTATION.md)

---

<div align="center">

**Made with ❤️ by Yang Hao**

🎵 Enjoy Your Music!

</div>

## ⌨️ 键盘快捷键

### 播放控制
- **播放/暂停**: `Space` (空格键)
- **上一曲**: `Cmd+←` (macOS) / `Ctrl+←` (Windows/Linux)
- **下一曲**: `Cmd+→` (macOS) / `Ctrl+→` (Windows/Linux)

### 其他快捷键
- **浏览歌曲**: `Cmd/Ctrl+F`
- **刷新音乐库**: `Cmd/Ctrl+R`
- **下载**: `Cmd/Ctrl+D`
- **设置**: `Cmd/Ctrl+S`

> 📝 详细说明请查看 [KEYBOARD_SHORTCUTS.md](./KEYBOARD_SHORTCUTS.md)
