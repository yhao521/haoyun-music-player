# 音乐播放器后台实现总结

## 一、已实现的功能模块

### 1.1 核心服务层

#### ✅ AudioPlayer (音频播放器)
**文件**: `backend/audioplayer.go`

**功能**:
- ✅ 支持 MP3, WAV, FLAC 格式解码播放
- ✅ 播放控制：Play, Pause, Stop, TogglePlayPause
- ✅ 音量控制：SetVolume, GetVolume (范围 0.0-1.0)
- ✅ 播放状态查询：IsPlaying
- ✅ 基于 beep v2 库的流式播放
- ✅ 线程安全：使用 sync.RWMutex

**技术实现**:
```go
type AudioPlayer struct {
    mu        sync.RWMutex
    ctrl      *beep.Ctrl
    streamer  beep.StreamSeekCloser
    gain      *effects.Gain
    // ...
}
```

**注意事项**:
- Gain 效果器的计算方式：`Gain = volume - 1` (因为 beep 库使用 1+Gain 的乘法)
- speaker 需要在播放前初始化
- 使用 StreamSeekCloser 接口而非 Streamer，以支持 Close 方法

---

#### ✅ PlaylistManager (播放列表管理)
**文件**: `backend/musicsmanager.go`

**功能**:
- ✅ 播放列表管理：AddToPlaylist, ClearPlaylist, GetPlaylist
- ✅ 播放控制：PlayIndex, Next, Previous
- ✅ 播放模式：顺序 (order), 循环 (loop), 随机 (random)
- ✅ 当前播放索引跟踪
- ✅ 线程安全：使用 sync.RWMutex

**播放模式实现**:
```go
switch pm.playMode {
case "random":
    pm.current = rand.Intn(len(pm.playlist))
case "loop":
    pm.current = pm.current % len(pm.playlist)
case "order":
    pm.current = (pm.current + 1) % len(pm.playlist)
}
```

---

#### ✅ LibraryManager (音乐库管理)
**文件**: `backend/libraryservice.go`

**功能**:
- ✅ 音乐库 CRUD：AddLibrary, RemoveLibrary, SwitchLibrary, RenameLibrary
- ✅ 目录扫描：scanDirectory (支持 mp3, wav, flac, aac, ogg, wma)
- ✅ JSON 持久化：saveLibrary, loadLibrary
- ✅ 多音乐库支持：使用 map[string]*MusicLibrary
- ✅ 当前库跟踪：currentLib
- ✅ 刷新功能：RefreshLibrary
- ✅ 获取所有库：GetAllLibraries
- ✅ 获取音轨路径：GetCurrentLibraryTracks
- ✅ 线程安全：使用 sync.RWMutex

**存储结构**:
```
~/.haoyun-music/
└── libraries/
    ├── music.json
    ├── work.json
    └── ...
```

**数据结构**:
```go
type MusicLibrary struct {
    Name      string
    Path      string
    CreatedAt time.Time
    UpdatedAt time.Time
    Tracks    []TrackInfo
}

type TrackInfo struct {
    Path     string
    Filename string
    Title    string
    Artist   string
    Album    string
    Duration int64
    Size     int64
}
```

---

#### ✅ MusicService (统一服务接口)
**文件**: `backend/music_service.go`

**功能**:
- ✅ 组合模式：持有 AudioPlayer, PlaylistManager, LibraryManager 引用
- ✅ 播放控制：Play, Pause, Stop, TogglePlayPause, Next, Previous, PlayIndex
- ✅ 音量控制：SetVolume, GetVolume
- ✅ 播放模式：SetPlayMode, GetPlayMode
- ✅ 播放列表：AddToPlaylist, ClearPlaylist, GetPlaylist
- ✅ 音乐库：AddLibrary, GetCurrentLibrary, SwitchLibrary, RefreshLibrary, RenameLibrary
- ✅ 加载音乐库：LoadCurrentLibrary (自动加载到播放列表并播放)
- ✅ 辅助方法：GetSongMetadata, IsPlaying, GetLibraries, SetCurrentLibrary

**架构模式**:
- Facade 模式：对外提供统一接口，对内委托给子服务
- 依赖注入：通过 SetApp 显式注入应用实例

---

### 1.2 系统服务层

#### ✅ Com (通用服务)
**文件**: `backend/com.go`

**功能**:
- ✅ 系统信息：IsMacOS
- ✅ 文件对话框：SelectPathDownload
- ✅ 应用实例管理

---

## 二、技术要点与解决方案

### 2.1 Beep API 适配

**问题**: beep v2 API 与预期不同
**解决方案**:
1. 使用 `go doc` 查看实际 API
2. Streamer 没有 Close 方法，需要使用 StreamSeekCloser
3. Decode 函数返回 StreamSeekCloser 而非 Streamer
4. Gain 效果器使用 `Gain = volume - 1` 计算方式

**正确用法**:
```go
// 1. 加载音频文件
streamer, format, err := mp3.Decode(file) // 返回 StreamSeekCloser

// 2. 创建增益控制器
gain := &effects.Gain{
    Streamer: streamer,
    Gain:     volume - 1, // 范围 -1 到 0
}

// 3. 创建播放控制器
ctrl := &beep.Ctrl{
    Streamer: gain,
    Paused:   false,
}

// 4. 开始播放
speaker.Play(ctrl)
```

### 2.2 并发控制

**策略**: 每个子服务独立的 sync.RWMutex

**优点**:
- 细粒度锁，提高并发性能
- 避免全局锁的瓶颈
- 各服务独立管理自己的状态

**示例**:
```go
func (pm *PlaylistManager) PlayIndex(index int) error {
    pm.mu.Lock()
    defer pm.mu.Unlock()
    // ...
}

func (pm *PlaylistManager) GetPlaylist() ([]string, error) {
    pm.mu.RLock()
    defer pm.mu.RUnlock()
    // ...
}
```

### 2.3 依赖注入

**模式**: 在 main.go 中统一创建和注入

```go
// 1. 创建服务实例
musicService := backend.NewMusicService()

// 2. 设置应用实例
musicService.SetApp(app)

// 3. 初始化
musicService.Init()

// 4. 注册到 Wails
app.Services: []application.Service{
    application.NewService(musicService),
}
```

### 2.4 事件通知机制

**后端事件**:
```go
// 播放状态变化
app.Event.Emit("playbackStateChanged", "playing" | "paused" | "stopped")

// 当前歌曲变化
app.Event.Emit("currentTrackChanged", filepath.Base(path))

// 播放列表更新
app.Event.Emit("playlistUpdated", playlist)

// 音乐库更新
app.Event.Emit("libraryUpdated", library)
```

**前端监听** (待实现):
```typescript
EventsOn('playbackStateChanged', (state: string) => {
  // 更新播放按钮图标
})

EventsOn('currentTrackChanged', (filename: string) => {
  // 更新当前歌曲显示
})
```

---

## 三、编译验证

### 3.1 编译命令
```bash
cd /Users/yanghao/storage/code_projects/goProjects/haoyun-music-player
go build -v
```

### 3.2 编译结果
✅ **编译成功** - 无错误，仅有链接警告（正常）

### 3.3 依赖整理
```bash
go mod tidy
```

---

## 四、待实现功能

### 4.1 前端 UI (Vue3 + TypeScript)

**待创建组件**:
1. `Player.vue` - 播放器控制组件
   - 播放/暂停按钮
   - 上一首/下一首按钮
   - 音量调节滑块
   - 播放进度条
   - 播放模式切换

2. `Playlist.vue` - 播放列表组件
   - 显示当前播放列表
   - 点击播放指定歌曲
   - 显示当前播放状态

3. `LibraryMenu.vue` - 音乐库菜单组件
   - 显示所有音乐库
   - 切换音乐库
   - 添加/刷新/重命名音乐库

### 4.2 TypeScript 类型定义

**待创建文件**: `frontend/src/types/music.ts`

```typescript
export interface TrackInfo {
  path: string;
  filename: string;
  title: string;
  artist: string;
  album: string;
  duration: number;
  size: number;
}

export interface MusicLibrary {
  name: string;
  path: string;
  created_at: string;
  updated_at: string;
  tracks: TrackInfo[];
}

export type PlayMode = 'order' | 'loop' | 'random';
```

### 4.3 元数据读取增强

**当前**: 仅从文件名提取基础信息

**扩展方案**: 集成 ID3 标签读取库
```go
import "github.com/dhowden/tag"

func readMetadata(path string) (TrackInfo, error) {
    file, _ := os.Open(path)
    metadata, _ := tag.ReadFrom(file)
    return TrackInfo{
        Title:    metadata.Title(),
        Artist:   metadata.Artist(),
        Album:    metadata.Album(),
        Duration: metadata.Duration().Seconds(),
    }, nil
}
```

### 4.4 播放进度事件

**待实现**: 
```go
// 在 AudioPlayer 中添加进度回调
go func() {
    for ap.isPlaying {
        position := ap.GetPosition()
        duration := ap.GetDuration()
        ap.app.Event.Emit("playbackProgress", map[string]float64{
            "position": position,
            "duration": duration,
        })
        time.Sleep(time.Second)
    }
}()
```

---

## 五、使用说明

### 5.1 开发环境运行

```bash
# 1. 安装依赖
go mod tidy
cd frontend && npm install

# 2. 开发模式运行
wails3 dev
```

### 5.2 生产构建

```bash
# 构建
wails3 build

# 输出目录
build/bin/
```

### 5.3 音乐库管理

1. **添加音乐库**:
   - 点击菜单"音乐库" → "添加新音乐库"
   - 选择音乐文件夹
   - 自动扫描并保存为 JSON 文件

2. **切换音乐库**:
   - 点击菜单"音乐库" → 选择库名称 (如 "✓ music")
   - 自动加载该库所有歌曲到播放列表并播放

3. **刷新音乐库**:
   - 点击菜单"音乐库" → "刷新当前音乐库"
   - 重新扫描目录，更新歌曲列表

4. **重命名音乐库**:
   - 点击菜单"音乐库" → "重命名当前音乐库"
   - 输入新名称

---

## 六、架构优势

### 6.1 MVC + Facade 模式
- **清晰的职责分离**: 每个服务专注单一功能
- **易于测试**: 各服务可独立单元测试
- **易于扩展**: 新增功能不影响现有架构

### 6.2 依赖注入
- **松耦合**: 服务间不直接依赖具体实现
- **可替换**: 可轻松替换子服务实现
- **易维护**: 依赖关系清晰明了

### 6.3 并发安全
- **细粒度锁**: 每个服务独立管理锁
- **高性能**: 避免全局锁的性能瓶颈
- **线程安全**: 所有公开方法都是线程安全的

---

## 七、技术栈总结

| 组件 | 技术 | 版本 | 说明 |
|------|------|------|------|
| 后端框架 | Wails | v3.0.0-alpha.74 | 跨平台桌面应用框架 |
| 后端语言 | Go | 1.25 | 系统编程语言 |
| 音频处理 | beep | v2 | 纯 Go 音频库 |
| 前端框架 | Vue | 3 | 渐进式 JS 框架 |
| 前端语言 | TypeScript | - | 类型安全的 JS 超集 |
| 构建工具 | Vite | - | 下一代前端构建工具 |
| 数据存储 | JSON | - | 轻量级数据交换格式 |

---

## 八、文件清单

### 8.1 后端核心文件

| 文件 | 行数 | 功能 |
|------|------|------|
| `backend/audioplayer.go` | ~250 行 | 音频播放控制 |
| `backend/musicsmanager.go` | ~200 行 | 播放列表管理 |
| `backend/libraryservice.go` | ~400 行 | 音乐库管理 |
| `backend/music_service.go` | ~300 行 | 统一服务接口 |
| `backend/com.go` | ~134 行 | 通用服务 |
| `main.go` | ~521 行 | 应用入口 |

### 8.2 文档文件

| 文件 | 功能 |
|------|------|
| `BACKEND_DESIGN.md` | 后台方案设计文档 |
| `IMPLEMENTATION_SUMMARY.md` | 实现总结 (本文档) |

---

## 九、总结

### 9.1 已完成
✅ 完整的后台服务架构实现  
✅ 音频播放控制 (MP3, WAV, FLAC)  
✅ 播放列表管理 (顺序/循环/随机)  
✅ 音乐库管理 (多库支持，JSON 持久化)  
✅ 统一服务接口 (Facade 模式)  
✅ 线程安全的并发控制  
✅ 事件通知机制  
✅ 编译通过，无错误  

### 9.2 待完成
⏳ 前端 Vue3 组件实现  
⏳ TypeScript 类型定义  
⏳ ID3 标签读取  
⏳ 播放进度显示  
⏳ 歌词显示功能  
⏳ 音效均衡器  

### 9.3 技术亮点
- **架构清晰**: MVC + Facade 模式，职责分离
- **代码质量**: 线程安全，错误处理完善
- **可扩展性**: 模块化设计，易于添加新功能
- **文档完善**: 设计文档 + 实现总结

---

**实现日期**: 2026-04-02  
**技术栈**: Wails3 + Vue3 + TypeScript + beep v2 + JSON  
**状态**: 后台实现完成 ✅，前端开发中 ⏳
