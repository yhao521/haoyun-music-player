# 🎵 Haoyun Music Player

<div align="center">

![Wails](https://img.shields.io/badge/Wails-v3-blue)
![Vue](https://img.shields.io/badge/Vue-3-green)
![TypeScript](https://img.shields.io/badge/TypeScript-4.9-blue)
![Go](https://img.shields.io/badge/Go-1.25+-00ADD8)
![Platform](https://img.shields.io/badge/Platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey)
![License](https://img.shields.io/badge/License-MIT-yellow)

**基于 Wails 3 + Vue 3 + TypeScript 的现代化菜单栏音乐播放器**

[快速开始](./QUICKSTART.md) • [实现文档](./IMPLEMENTATION.md) • [键盘快捷键](./KEYBOARD_SHORTCUTS.md) • [API 指南](./API_GUIDE.md) • [多语言支持](./I18N_QUICKSTART.md)

</div>

## ✨ 特性亮点

- 🎨 **现代 UI** - 精美的渐变设计和毛玻璃特效
- ⚡ **高性能** - Go + Vue 的原生性能体验
- 🔧 **可扩展** - 清晰的 MVC 架构，易于添加新功能
- 📱 **跨平台** - 支持 macOS、Windows、Linux
- 🌍 **多语言支持** - 内置中英文界面，支持实时切换（[详情](./I18N_QUICKSTART.md)）
- 🎯 **类型安全** - 完整的 TypeScript 类型定义
- 📋 **播放列表管理** - 完善的播放列表控制和播放模式（顺序/循环/随机/单曲）
- 🎚️ **音量控制** - 精细的音量调节 (0-100%)
- 💻 **菜单集成** - 系统托盘菜单快速访问
- 🎵 **广泛的格式支持** - 通过 FFmpeg 支持 MP3、WAV、FLAC、AAC、M4A、OGG、WMA、APE、Opus 等几乎所有音频格式（[详情](./FFMPEG_GUIDE.md)）
- 📂 **音乐库管理** - 支持多个音乐库，自动扫描和索引
- ⌨️ **全局快捷键** - 空格键播放/暂停，Cmd/Ctrl+[ 上一曲，Cmd/Ctrl+] 下一曲
- 🔍 **歌曲搜索** - 浏览视图中支持按标题、艺术家、专辑搜索
- 📄 **分页显示** - 大型音乐库支持分页浏览（20/50/100/200 首每页）
- 🔄 **断点续播** - 暂停后恢复播放时自动跳转到原位置

## 🖼️ 界面预览

### 主播放器界面

```
┌─────────────────────────────────┐
│  🎵 Haoyun Music Player         │
├─────────────────────────────────┤
│                                 │
│  ┌──────┐                       │
│  │ 🎵   │  歌曲名称.mp3          │
│  └──────┘  未知艺术家            │
│                                 │
│  1:23              3:45         │
│  ━━━━━━━━━━━━━━○━━━━━━━        │
│                                 │
│      ⏮   ▶️   ⏭                │
│                                 │
│  🔊 ━━━━━━━━━━━━                │
│                                 │
│  📂 打开文件                     │
│                                 │
│  播放列表 (15)                   │
│  ─────────────────────────────  │
│  1. 歌曲1.mp3                    │
│  2. 歌曲2.mp3                    │
│  ...                             │
│                                 │
└─────────────────────────────────┘
```

### 浏览音乐库界面

```
┌──────────────────────────────────────────────────┐
│  🎵 浏览音乐库                                    │
├──────────────────────────────────────────────────┤
│  音乐库: [我的音乐 ▼]  🔍 搜索歌曲、艺术家...     │
├──────────────────────────────────────────────────┤
│  📁 我的音乐 | 🎵 150 首 | ⏱️ 总时长：8:32:15    │
│  💾 总大小：1.2 GB | 📂 /Users/xxx/Music         │
├──────────────────────────────────────────────────┤
│  #  | 标题           | 艺术家  | 专辑   | 时长 | 大小│
│  ──────────────────────────────────────────────── │
│  1  | 歌曲名称1.mp3  | 艺术家1 | 专辑1  | 3:45 | 8MB │
│  2  | 歌曲名称2.mp3  | 艺术家2 | 专辑2  | 4:12 | 9MB │
│  ...                                              │
├──────────────────────────────────────────────────┤
│  显示第 1-50 首，共 150 首                        │
│  ⏮ ◀ 1 2 3 4 5 ▶ ⏭  每页显示: [50 ▼]           │
└──────────────────────────────────────────────────┘
```

## 🛠️ 技术栈

### 后端

- **Go 1.25+** - 核心业务逻辑
- **Wails v3 (Alpha)** - 桌面应用框架
- **Oto v3** - 跨平台音频播放引擎
- **go-mp3** - MP3 流式解码器
- **go-audio/wav** - WAV 音频解码器
- **mewkiz/flac** - FLAC 音频解码器

### 前端

- **Vue 3** - 渐进式 JavaScript 框架
- **TypeScript** - 类型安全的 JavaScript 超集
- **Vite** - 下一代前端构建工具
- **@wailsio/runtime** - Wails 运行时

### 音频架构

- **Oto Context 单例模式** - 整个应用生命周期只创建一次
- **流式解码** - MP3 采用流式读取，内存效率高
- **PCM 缓存** - WAV/FLAC 全量解码后缓存
- **状态标志位控制** - 暂停通过停止数据流实现，而非关闭 Player

## 🚀 快速开始

### 前置要求

- **Go 1.25+** - [下载安装](https://golang.org/dl/)
- **Node.js 18+** - [下载安装](https://nodejs.org/)
- **FFmpeg** - 音频解码引擎（[安装指南](./FFMPEG_GUIDE.md)）
  - macOS: `brew install ffmpeg`
  - Ubuntu: `sudo apt-get install ffmpeg`
  - Windows: `choco install ffmpeg`
- **Wails v3 CLI**（可选）- `go install github.com/wailsapp/wails/v3/cmd/wails3@latest`
- **编译器** - macOS: Xcode Command Line Tools, Windows: GCC/MinGW, Linux: build-essential

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

#### 开发模式（热重载）

```bash
# 使用 Wails CLI
wails3 dev -config ./build/config.yml

# 或使用启动脚本（Unix/Linux/macOS）
./start.sh

# Windows
start.bat
```

#### 生产模式

```bash
# 构建应用
wails3 build

# 或直接运行
go run .
```

### 生产构建

```bash
# macOS (Intel & Apple Silicon)
wails3 build -platform darwin

# Windows
wails3 build -platform windows

# Linux
wails3 build -platform linux

# 指定输出目录
wails3 build -o ./dist
```

## 📁 项目结构

```
haoyun-music-player/
├── main.go                      # Go 主入口，应用初始化
├── backend/                     # 后端代码
│   ├── music_service.go         # 统一音乐服务（MVC Model 层）
│   ├── audioplayer.go           # 音频播放器（支持 FFmpeg 解码）
│   ├── libraryservice.go        # 音乐库管理服务
│   ├── musicsmanager.go         # 播放列表管理器
│   ├── historymanager.go        # 播放历史管理
│   ├── lyricmanager.go          # 歌词管理
│   ├── covermanager.go          # 专辑封面管理
│   └── pkg/                     # 公共工具包
│       ├── config/              # 配置管理
│       ├── i18n/                # 国际化支持
│       ├── file/                # 文件操作工具
│       └── utils/               # 通用工具函数
├── frontend/                    # 前端代码
│   ├── src/
│   │   ├── components/          # Vue 组件
│   │   ├── views/               # 页面视图
│   │   ├── i18n/                # 前端国际化
│   │   └── main.ts              # 前端入口
│   └── public/                  # 静态资源
├── FFMPEG_GUIDE.md              # FFmpeg 安装和使用指南
├── test_ffmpeg.sh               # macOS/Linux FFmpeg 测试脚本
├── test_ffmpeg.bat              # Windows FFmpeg 测试脚本
└── ...                          # 其他文档和配置文件
```

## 📖 功能清单

### ✅ 已实现

#### 核心功能

- [x] 基础 UI 界面（主播放器 + 浏览视图）
- [x] 真实音频播放核心（MP3/WAV/FLAC）
- [x] **FFmpeg 集成** - 支持 AAC、M4A、OGG、WMA、APE、Opus 等格式（[详情](./FFMPEG_GUIDE.md)）
- [x] 播放/暂停控制（含断点续播）
- [x] 上一首/下一首切换
- [x] 进度条显示和拖拽跳转
- [x] 音量调节滑块（0-100%）
- [x] 播放列表管理（添加、清空、切换）
- [x] 播放模式支持（顺序/循环/随机/单曲）
- [x] 文件选择器集成
- [x] 前后端事件通信（Wails RPC）
- [x] 状态同步机制

#### 音乐库管理

- [x] 多音乐库支持
- [x] 后台异步扫描
- [x] 音乐库切换
- [x] 音乐库刷新
- [x] JSON 持久化存储（~/.haoyun-music/）
- [x] 自动加载当前音乐库到播放列表

#### 系统集成

- [x] 系统托盘图标
- [x] 托盘菜单控制
- [x] 正在播放显示（托盘菜单）
- [x] 窗口隐藏/显示（关闭时不退出）
- [x] 多窗口支持（主窗口 + 浏览窗口）
- [x] 键盘快捷键（空格、Cmd/Ctrl+[、Cmd/Ctrl+]）
- [x] 菜单栏播放控制

#### 浏览视图

- [x] 音乐库列表展示
- [x] 歌曲列表表格显示
- [x] 实时搜索过滤（标题/艺术家/专辑）
- [x] 分页显示（20/50/100/200 首每页）
- [x] 双击播放歌曲
- [x] 统计信息（总数、总时长、总大小）
- [x] 响应式布局

#### 增强功能

- [x] 播放历史记录保存（含播放次数统计）
- [x] 歌词显示支持（LRC 文件，自动扫描关联）
- [x] 专辑封面提取框架（待完善 MP3/FLAC 解析）
- [x] 托盘菜单"喜爱音乐"（按播放次数排序）
- [x] 音乐库删除功能

### ⏳ 计划中

- [ ] 均衡器调节
- [ ] 媒体键支持（播放/暂停、上一曲、下一曲）
- [ ] 全局快捷键自定义
- [ ] 睡眠定时器
- [ ] ~~播放历史记录保存~~ ✅ 已实现
- [ ] ~~歌词显示（LRC 文件支持）~~ ✅ 已实现
- [ ] ~~专辑封面提取和显示~~ ✅ 基础框架已实现
- [ ] 收藏夹功能
- [ ] 在线音乐搜索
- [ ] 主题切换（深色/浅色模式）
- [ ] 国际化支持（i18n）
- [ ] 单元测试覆盖

## 🎯 架构设计

### MVC 架构模式

本项目采用 **MVC + Facade** 架构模式：

```
┌─────────────────────────────────────┐
│         Frontend (Vue 3)            │  ← View 层
│  - AppMain.vue (主播放器)            │
│  - BrowseView.vue (浏览视图)         │
└──────────────┬──────────────────────┘
               │ Wails RPC
┌──────────────▼──────────────────────┐
│      MusicService (Facade)          │  ← Controller 层
│  - 统一对外接口                      │
│  - 协调子服务                        │
└──┬────────────┬──────────────┬──────┘
   │            │              │
┌──▼──────┐ ┌──▼────────┐ ┌──▼──────────┐
│Audio    │ │Playlist   │ │Library      │  ← Model 层
│Player   │ │Manager    │ │Manager      │
│         │ │           │ │             │
│- Oto    │ │- 播放列表  │ │- 音乐库扫描  │
│- 解码器  │ │- 播放模式  │ │- JSON 存储  │
│- 状态管理│ │- 索引管理  │ │- 元数据解析  │
└─────────┘ └───────────┘ └─────────────┘
```

### 关键设计决策

1. **Oto Context 全局单例**
   - 整个应用生命周期只创建一次
   - 严禁重复创建或中途关闭
   - Player 动态重建，Context 保持不变

2. **暂停实现策略**
   - 采用状态标志位 + 数据流控制
   - 暂停时关闭 streamer，保存播放位置
   - 恢复时重新加载文件并 Seek 到原位置

3. **MP3 流式解码**
   - 使用 go-mp3 实现真正的流式读取
   - 内存效率高，适合大文件
   - WAV/FLAC 采用全量解码后缓存

4. **服务下沉与依赖注入**
   - 具体功能拆分为独立子服务
   - 每个子服务内部封装 sync.RWMutex
   - 顶层 Facade 组合所有子服务

详细架构说明请查看 [BACKEND_DESIGN.md](./BACKEND_DESIGN.md)

## 📚 文档

### 用户文档

- [快速开始指南](./QUICKSTART.md) - 5 分钟上手
- [键盘快捷键](./KEYBOARD_SHORTCUTS.md) - 完整快捷键列表
- [故障排除](./TROUBLESHOOTING.md) - 常见问题解决

### 开发文档

- [实现文档](./IMPLEMENTATION.md) - 功能实现细节
- [API 指南](./API_GUIDE.md) - 后端 API 使用说明
- [后端设计](./BACKEND_DESIGN.md) - 架构设计详解
- [Wails 绑定](./WAILS_BINDINGS.md) - 前后端通信机制

### 特性文档

- [音乐信息显示](./MUSIC_INFO_DISPLAY.md)
- [正在播放功能](./NOW_PLAYING_FEATURE.md)
- [媒体键支持](./MEDIA_KEYS_GUIDE.md)
- [托盘修复说明](./TRAY_FIX.md)
- [扬声器修复](./SPEAKER_FIX.md)
- [FFmpeg 音频解码](./FFMPEG_GUIDE.md) - 广泛的音频格式支持
- [新功能实现](./NEW_FEATURES.md) - 播放历史、歌词、专辑封面
- [Bug 修复记录](./BUGFIX_MENU_CRASH.md) - 菜单空指针错误修复

## 🧪 FFmpeg 测试

本项目集成了 FFmpeg 以支持更多音频格式。您可以使用以下命令测试 FFmpeg 是否正常工作：

### macOS/Linux

```
# 运行测试脚本
./test_ffmpeg.sh
```

### Windows

```
# 运行测试脚本
.\test_ffmpeg.bat
```

测试脚本会：

1. ✅ 检查 FFmpeg 是否已安装
2. ✅ 扫描当前目录下的音频文件
3. ✅ 尝试解码每个文件并显示信息
4. ✅ 验证采样率、声道数、时长等参数

详细安装和使用说明请查看 [FFMPEG_GUIDE.md](./FFMPEG_GUIDE.md)。

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

### 开发流程

1. Fork 本项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

### 开发建议

- ✅ 遵循 Go 代码规范（gofmt、golint）
- ✅ 使用 TypeScript 严格模式
- ✅ 编写必要的单元测试
- ✅ 更新相关文档
- ✅ 保持向后兼容性
- ✅ 添加适当的日志记录

### 代码规范

#### Go 后端

```
// 使用有意义的变量名
audioPlayer := NewAudioPlayer()

// 错误处理
if err != nil {
    log.Printf("操作失败：%v", err)
    return err
}

// 并发安全
ap.mu.Lock()
defer ap.mu.Unlock()
```

#### Vue 前端

```
// 使用 Composition API
import { ref, computed } from "vue"

// 类型定义
interface TrackInfo {
  title: string
  artist: string
  duration: number
}

// 响应式状态
const tracks = ref<TrackInfo[]>([])
```

## 🐛 已知问题

1. **Wails v3 Alpha 阶段**
   - 框架处于 Alpha 版本，API 可能变动
   - 部分功能可能需要 workaround

2. **音频格式支持**
   - ✅ MP3、WAV、FLAC 使用原生解码器（高性能）
   - ✅ AAC、M4A、OGG、WMA、APE、Opus 等通过 FFmpeg 支持
   - ⚠️ 需要安装 FFmpeg 才能播放非原生格式
   - 📖 查看 [FFMPEG_GUIDE.md](./FFMPEG_GUIDE.md) 了解详细信息

3. **macOS 权限**
   - 首次运行可能需要授权访问音乐文件夹
   - 系统偏好设置 > 安全性与隐私 > 完全磁盘访问权限

4. **内存使用**
   - FFmpeg 解码会将整个音频文件加载到内存
   - 大文件（>100MB）可能占用较多 RAM
   - 建议使用原生支持的格式（MP3/WAV/FLAC）以获得最佳性能

## 📄 许可证

Apache 2.0 License - 详见 [LICENSE](./LICENSE) 文件

## 🙏 致谢

感谢以下优秀的开源项目：

- [Wails](https://wails.io/) - 卓越的 Go 桌面应用框架
- [Vue.js](https://vuejs.org/) - 渐进式 JavaScript 框架
- [TypeScript](https://www.typescriptlang.org/) - JavaScript 的超集
- [Vite](https://vitejs.dev/) - 极速的前端构建工具
- [Oto](https://github.com/ebitengine/oto) - 跨平台音频播放库
- [go-mp3](https://github.com/hajimehoshi/go-mp3) - MP3 解码器
- [go-audio](https://github.com/go-audio/wav) - WAV 解码器
- [flac](https://github.com/mewkiz/flac) - FLAC 解码器

## 📬 联系方式

- 👤 **作者**: \*\* (yhao521)
- 📧 **Email**: your.email@example.com
- 💬 **Issues**: [GitHub Issues](https://github.com/yhao521/haoyun-music-player/issues)
- 📖 **文档**: [完整文档列表](#-文档)

---

<div align="center">

**Made with ❤️ by YHao521**

🎵 Enjoy Your Music!

⭐ 如果这个项目对你有帮助，请给个 Star！

</div>
