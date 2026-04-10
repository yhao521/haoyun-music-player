# 快速开始指南

## 项目说明

这是一个基于 Wails 3 + Vue 3 + TypeScript 的菜单栏音乐播放器。

**注意**: Wails v3 目前处于 alpha 阶段，API 可能不稳定。本项目实现了核心的音乐播放器界面和基础功能。

## 当前实现的功能

✅ **前端界面**
- 精美的渐变 UI 设计
- 播放/暂停控制
- 进度条显示和拖拽
- 音量调节
- 播放列表管理
- 文件选择器调用

✅ **后端服务**
- 音乐播放状态管理
- 播放列表管理
- 事件系统（前端与后端通信）
- 文件对话框集成

⚠️ **音频播放核心**
- 由于 Go 音频库的复杂性和平台依赖性，当前版本实现了模拟播放功能
- 真实的音频播放需要集成以下库之一：
  - `oto` + `mp3`/`wav` 解码器（需要 CGO）
  - `portaudio`（跨平台音频 I/O）
  - 系统原生音频 API

## 运行步骤

### 1. 安装依赖

```bash
# 安装 Go 依赖
go mod tidy

# 安装前端依赖
cd frontend
npm install
cd ..
```

### 2. 开发模式运行

```bash
# 使用 wails dev 命令
wails3 dev -config ./build/config.yml
```

或者使用 Task（如果已安装）:

```bash
task dev
```

### 3. 生产构建

```bash
# macOS 构建
wails3 build -platform darwin

# Windows 构建
wails3 build -platform windows

# Linux 构建
wails3 build -platform linux
```

## 技术架构

### 后端 (Go)
- **框架**: Wails v3
- **语言**: Go 1.25+
- **主要功能**:
  - 音乐服务 (`MusicService`)
  - 事件注册和发送
  - 文件对话框集成

### 前端 (Vue 3)
- **框架**: Vue 3 + TypeScript
- **构建工具**: Vite
- **运行时**: @wailsio/runtime
- **主要组件**:
  - `App.vue` - 主播放器界面
  - 事件监听和状态管理

## 如何添加真实音频播放

要添加真实的音频播放功能，你需要：

1. **选择音频库**:
   ```bash
   # 选项 1: Oto (推荐，但需要 CGO)
   go get github.com/hajimehoshi/oto/v2
   
   # 选项 2: PortAudio (跨平台)
   go get github.com/go-portaudio/portaudio
   ```

2. **在 `greetservice.go` 中实现**:
   - 初始化音频上下文
   - 加载音频文件并解码
   - 创建播放器并控制播放

3. **参考实现**:
   - [Oto 示例](https://github.com/hajimehoshi/oto)
   - [Go 音频处理教程](https://github.com/golang/example/audio)

## 项目结构

```
haoyun-music-player/
├── main.go                 # Go 主入口
├── greetservice.go         # 音乐服务实现
├── go.mod                  # Go 模块配置
├── Taskfile.yml           # Task 任务配置
├── frontend/              # 前端代码
│   ├── src/
│   │   ├── App.vue       # 主应用组件
│   │   ├── main.ts       # Vue 入口
│   │   └── vite-env.d.ts # TypeScript 类型定义
│   ├── public/
│   │   └── style.css     # 全局样式
│   └── package.json      # 前端依赖配置
└── build/                # 构建配置
    └── config.yml        # Wails 配置文件
```

## 常见问题

### Q: 为什么播放功能是模拟的？
A: 真实的音频播放需要平台相关的音频库和编解码器，这增加了复杂性和构建难度。当前版本专注于展示 Wails + Vue 的集成架构。

### Q: 如何测试播放功能？
A: 你可以：
1. 点击"打开文件"选择音乐文件
2. 点击播放按钮查看状态变化
3. 调节音量和拖动进度条
4. 管理播放列表

### Q: Wails v3 稳定吗？
A: Wails v3 目前处于 alpha 阶段，API 可能会变化。建议关注官方文档更新。

## 下一步计划

1. ✅ 完成基础 UI 界面
2. ✅ 实现前后端通信
3. ✅ 添加播放列表功能
4. ⏳ 集成真实音频播放库
5. ⏳ 添加全局快捷键支持
6. ⏳ 实现歌词显示
7. ⏳ 添加专辑封面显示

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License

---

**Happy Coding! 🎵**
