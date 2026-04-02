# 本地音乐播放器后台方案设计文档

## 一、系统架构设计

### 1.1 整体架构

```
─────────────────────────────────────────────────────────────┐
│                      前端层 (Vue3 + TS)                      │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────────┐        │
│  │  播放控制   │  │  音乐库管理  │  │  播放列表 UI  │        │
│  └─────────────┘  └─────────────┘  └──────────────┘        │
└─────────────────────────────────────────────────────────────┘
                            ↓ Wails RPC Bridge
┌─────────────────────────────────────────────────────────────┐
│                    后端服务层 (Go)                           │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              MusicService (统一服务接口)               │  │
│  └──────────────────────────────────────────────────────┘  │
│           ↓                  ↓                  ↓           │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────────┐       │
│  │ AudioPlayer │  │LibraryManager│ │PlaylistManager│       │
│  │ (音频播放)  │  │ (音乐库管理) │  │ (播放列表)   │       │
│  └─────────────┘  └─────────────┘  └──────────────┘       │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                      数据存储层                              │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────────┐        │
│  │  beep v2    │  │  JSON 文件   │  │  系统文件系统 │        │
│  │ (音频解码)  │  │  (配置存储)  │  │  (音乐文件)  │        │
│  └─────────────┘  └─────────────┘  └──────────────┘        │
─────────────────────────────────────────────────────────────┘
```

### 1.2 设计模式

#### 1.2.1 MVC + Facade 模式
- **Model 层**: `MusicService` 作为统一数据模型和服务外观
- **View 层**: Vue3 组件负责 UI 渲染
- **Controller 层**: `main.go` 作为应用入口，协调各服务

#### 1.2.2 组合模式
- `MusicService` 通过组合持有所有子服务引用
- 各子服务独立封装，内部维护自己的状态和锁

#### 1.2.3 依赖注入
- 所有子服务在 `main.go` 中实例化
- 通过 `SetApp()` 方法显式注入应用实例
- 服务间依赖通过 Setter 方法注入

### 1.3 模块划分

| 模块 | 职责 | 关键文件 |
|------|------|----------|
| 音频播放模块 | 音频文件解码、播放控制、音量调节 | `audioplayer.go` |
| 音乐库管理模块 | 音乐库增删改查、目录扫描、元数据提取 | `libraryservice.go` |
| 播放列表模块 | 播放列表管理、播放模式控制、索引管理 | `musicsmanager.go` |
| 系统服务模块 | 系统托盘、菜单、快捷键、事件通知 | `com.go` |

---

## 二、数据结构设计

### 2.1 核心数据结构

#### 2.1.1 音乐文件信息 (TrackInfo)
```go
type TrackInfo struct {
    Path     string    `json:"path"`      // 文件绝对路径
    Filename string    `json:"filename"`  // 文件名
    Title    string    `json:"title"`     // 歌曲标题
    Artist   string    `json:"artist"`    // 艺术家
    Album    string    `json:"album"`     // 专辑
    Duration int64     `json:"duration"`  // 时长 (秒)
    Size     int64     `json:"size"`      // 文件大小 (字节)
}
```

#### 2.1.2 音乐库结构 (MusicLibrary)
```go
type MusicLibrary struct {
    Name      string      `json:"name"`       // 库名称
    Path      string      `json:"path"`       // 目录路径
    CreatedAt time.Time   `json:"created_at"` // 创建时间
    UpdatedAt time.Time   `json:"updated_at"` // 更新时间
    Tracks    []TrackInfo `json:"tracks"`     // 音轨列表
}
```

#### 2.1.3 播放状态 (PlaybackState)
```go
type PlaybackState struct {
    mu        sync.RWMutex
    isPlaying bool        // 是否正在播放
    volume    float64     // 音量 (0.0-1.0)
    app       *application.App
}
```

#### 2.1.4 播放列表管理器 (PlaylistManager)
```go
type PlaylistManager struct {
    mu       sync.RWMutex
    playlist []string    // 播放列表 (文件路径)
    current  int         // 当前播放索引
    app      *application.App
    playMode string     // 播放模式：order/loop/random
}
```

#### 2.1.5 音频播放器 (AudioPlayer)
```go
type AudioPlayer struct {
    mu        sync.RWMutex
    isPlaying bool
    paused    bool
    volume    float64
    ctrl      *beep.Ctrl      // 播放控制器
    streamer  beep.Streamer   // 音频流
    format    beep.Format     // 音频格式
    gain      *effects.Gain   // 增益效果器
    app       *application.App
}
```

### 2.2 数据存储结构

#### 2.2.1 存储目录
```
~/.haoyun-music/
├── libraries/           # 音乐库数据
│   ├── music.json      # 默认音乐库
│   ├── work.json       # 工作音乐库
│   └── ...
└── config.json         # 应用配置 (可选)
```

#### 2.2.2 音乐库 JSON 文件格式
```json
{
  "name": "music",
  "path": "/Users/username/Music",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-02T00:00:00Z",
  "tracks": [
    {
      "path": "/Users/username/Music/song1.mp3",
      "filename": "song1.mp3",
      "title": "Song Title",
      "artist": "Artist Name",
      "album": "Album Name",
      "duration": 240,
      "size": 5242880
    }
  ]
}
```

---

## 三、服务层设计

### 3.1 MusicService (统一服务接口)

#### 3.1.1 结构定义
```go
type MusicService struct {
    ctx             context.Context
    app             *application.App
    audioPlayer     *AudioPlayer        // beep 音频播放器
    playlistManager *PlaylistManager    // 播放列表管理
    libraryManager  *LibraryManager     // 音乐库管理
}
```

#### 3.1.2 核心方法

| 方法 | 功能 | 参数 | 返回值 |
|------|------|------|--------|
| `Play()` | 播放音乐 | - | `error` |
| `Pause()` | 暂停音乐 | - | `error` |
| `Stop()` | 停止播放 | - | `error` |
| `TogglePlayPause()` | 切换播放/暂停 | - | `(bool, error)` |
| `SetVolume(float64)` | 设置音量 | volume | `error` |
| `GetVolume()` | 获取音量 | - | `(float64, error)` |
| `Next()` | 下一首 | - | `error` |
| `Previous()` | 上一首 | - | `error` |
| `PlayIndex(int)` | 播放指定索引 | index | `error` |
| `SetPlayMode(string)` | 设置播放模式 | mode | `error` |
| `GetPlayMode()` | 获取播放模式 | - | `(string, error)` |
| `AddToLibrary(string)` | 添加目录到音乐库 | dirPath | `error` |
| `GetCurrentLibrary()` | 获取当前音乐库 | - | `*MusicLibrary` |
| `SwitchLibrary(string)` | 切换音乐库 | libName | `error` |
| `RefreshLibrary()` | 刷新当前音乐库 | - | `error` |
| `RenameLibrary(string)` | 重命名音乐库 | newName | `error` |
| `LoadCurrentLibrary()` | 加载当前库到播放列表 | - | `error` |

### 3.2 AudioPlayer (音频播放器)

#### 3.2.1 核心功能
1. **音频解码**: 支持 MP3, WAV, FLAC 格式
2. **播放控制**: Play, Pause, Stop, Seek
3. **音量控制**: Gain 效果器调节音量
4. **扬声器管理**: 初始化和管理 speaker

#### 3.2.2 方法列表
```go
func (ap *AudioPlayer) Play(path string) error
func (ap *AudioPlayer) Pause() error
func (ap *AudioPlayer) Stop() error
func (ap *AudioPlayer) SetVolume(volume float64) error
func (ap *AudioPlayer) Seek(position float64) error
func (ap *AudioPlayer) GetDuration() float64
func (ap *AudioPlayer) GetPosition() float64
func (ap *AudioPlayer) IsPlaying() bool
```

### 3.3 LibraryManager (音乐库管理器)

#### 3.3.1 核心功能
1. **库管理**: 创建、加载、保存、删除音乐库
2. **目录扫描**: 后台 goroutine 扫描音乐文件
3. **元数据提取**: 从音频文件中提取 ID3 标签
4. **持久化**: JSON 文件存储

#### 3.3.2 方法列表
```go
func (lm *LibraryManager) Init() error
func (lm *LibraryManager) AddLibrary(name, path string) error
func (lm *LibraryManager) RemoveLibrary(name string) error
func (lm *LibraryManager) SwitchLibrary(name string) error
func (lm *LibraryManager) GetCurrentLibrary() *MusicLibrary
func (lm *LibraryManager) RefreshLibrary() error
func (lm *LibraryManager) RenameLibrary(newName string) error
func (lm *LibraryManager) GetAllLibraries() []*MusicLibrary
func (lm *LibraryManager) scanDirectory(dirPath string) ([]TrackInfo, error)
func (lm *LibraryManager) saveLibrary(lib *MusicLibrary) error
func (lm *LibraryManager) loadLibrary(name string) (*MusicLibrary, error)
```

### 3.4 PlaylistManager (播放列表管理器)

#### 3.4.1 核心功能
1. **播放列表管理**: 添加、删除、清空播放列表
2. **播放控制**: 下一首、上一首、指定播放
3. **播放模式**: 顺序、循环、随机

#### 3.4.2 方法列表
```go
func (pm *PlaylistManager) AddToPlaylist(path string) error
func (pm *PlaylistManager) ClearPlaylist() error
func (pm *PlaylistManager) GetPlaylist() ([]string, error)
func (pm *PlaylistManager) PlayIndex(index int) error
func (pm *PlaylistManager) Next() error
func (pm *PlaylistManager) Previous() error
func (pm *PlaylistManager) SetPlayMode(mode string) error
func (pm *PlaylistManager) GetPlayMode() (string, error)
```

---

## 四、关键业务流程

### 4.1 音乐库管理流程

#### 4.1.1 添加音乐库
```
1. 用户选择"添加新音乐库"菜单项
2. 打开目录选择对话框
3. 用户选择音乐文件夹
4. 后台创建 MusicLibrary 对象
5. 启动 goroutine 扫描目录
6. 提取音频文件元数据
7. 保存到 JSON 文件
8. 更新 UI 显示新库
```

#### 4.1.2 切换音乐库
```
1. 用户点击菜单中的库名称 (如"✓ music")
2. 调用 SwitchLibrary(name)
3. 加载目标库的 JSON 文件
4. 清空当前播放列表
5. 将所有音轨添加到播放列表
6. 播放第一首歌曲
7. 更新 UI 显示当前库
```

#### 4.1.3 刷新音乐库
```
1. 用户选择"刷新当前音乐库"菜单项
2. 获取当前库的目录路径
3. 启动 goroutine 重新扫描目录
4. 对比文件变化 (新增/删除)
5. 更新 Tracks 列表
6. 保存到 JSON 文件
7. 如果当前正在播放，保持播放状态
8. 发送事件通知 UI 更新
```

#### 4.1.4 重命名音乐库
```
1. 用户选择"重命名当前音乐库"菜单项
2. 弹出输入框获取新名称
3. 检查名称是否已存在
4. 修改 MusicLibrary.Name
5. 删除旧 JSON 文件
6. 以新名称保存 JSON 文件
7. 更新 libraries map
8. 更新 UI 菜单显示
```

### 4.2 播放控制流程

#### 4.2.1 播放音乐
```
1. 从播放列表获取当前歌曲路径
2. 调用 AudioPlayer.Play(path)
3. 打开音频文件
4. 根据格式选择解码器 (mp3/wav/flac)
5. 初始化 speaker (如果未初始化)
6. 创建 Ctrl 控制器
7. 应用 Gain 效果器
8. 开始播放
9. 发送"playbackStateChanged"事件
```

#### 4.2.2 切换歌曲
```
1. 根据播放模式计算下一首索引
   - order: (current + 1) % len
   - loop: current % len
   - random: rand.Intn(len)
2. 更新 current 索引
3. 发送"currentTrackChanged"事件
4. 调用 AudioPlayer.Play(newPath)
```

### 4.3 事件通知机制

#### 4.3.1 后端事件
```go
// 播放状态变化
app.Event.Emit("playbackStateChanged", "playing" | "paused" | "stopped")

// 当前歌曲变化
app.Event.Emit("currentTrackChanged", filename)

// 播放列表更新
app.Event.Emit("playlistUpdated", playlist)

// 音乐库更新
app.Event.Emit("libraryUpdated", library)
```

#### 4.3.2 前端监听
```typescript
EventsOn('playbackStateChanged', (state: string) => {
  // 更新播放按钮图标
})

EventsOn('currentTrackChanged', (filename: string) => {
  // 更新当前歌曲显示
})

EventsOn('playlistUpdated', (playlist: string[]) => {
  // 刷新播放列表 UI
})
```

---

## 五、并发控制设计

### 5.1 锁机制

#### 5.1.1 各服务独立锁
```go
// PlaybackState
type PlaybackState struct {
    mu sync.RWMutex
    // ...
}

// PlaylistManager
type PlaylistManager struct {
    mu sync.RWMutex
    // ...
}

// AudioPlayer
type AudioPlayer struct {
    mu sync.RWMutex
    // ...
}
```

#### 5.1.2 读写锁使用规范
- **读操作**: `mu.RLock()` / `mu.RUnlock()`
  - GetPlaylist(), IsPlaying(), GetVolume()
- **写操作**: `mu.Lock()` / `mu.Unlock()`
  - Play(), Pause(), SetVolume(), AddToPlaylist()

### 5.2 异步任务处理

#### 5.2.1 后台扫描
```go
func (lm *LibraryManager) scanDirectoryAsync(dirPath string) {
    go func() {
        tracks, err := lm.scanDirectory(dirPath)
        if err != nil {
            log.Printf("扫描失败：%v", err)
            return
        }
        // 更新库数据
        lm.updateTracks(tracks)
        // 发送完成事件
        lm.app.Event.Emit("libraryScanComplete", len(tracks))
    }()
}
```

#### 5.2.2 上下文控制
```go
// 设置上下文
func (lm *LibraryManager) SetContext(ctx context.Context) {
    lm.ctx = ctx
}

// 扫描中检查取消
func (lm *LibraryManager) scanDirectory(dirPath string) ([]TrackInfo, error) {
    for _, file := range files {
        select {
        case <-lm.ctx.Done():
            return nil, lm.ctx.Err()
        default:
            // 继续处理
        }
    }
}
```

---

## 六、API 接口设计

### 6.1 前端可调用的后端方法

#### 6.1.1 播放控制 API
```typescript
// 播放
MusicService.Play(): Promise<void>

// 暂停
MusicService.Pause(): Promise<void>

// 停止
MusicService.Stop(): Promise<void>

// 切换播放/暂停
MusicService.TogglePlayPause(): Promise<boolean>

// 下一首
MusicService.Next(): Promise<void>

// 上一首
MusicService.Previous(): Promise<void>

// 播放指定索引
MusicService.PlayIndex(index: number): Promise<void>

// 设置音量
MusicService.SetVolume(volume: number): Promise<void>

// 获取音量
MusicService.GetVolume(): Promise<number>

// 设置播放模式
MusicService.SetPlayMode(mode: 'order' | 'loop' | 'random'): Promise<void>

// 获取播放模式
MusicService.GetPlayMode(): Promise<string>
```

#### 6.1.2 音乐库管理 API
```typescript
// 添加音乐库
MusicService.AddToLibrary(dirPath: string): Promise<void>

// 获取当前音乐库
MusicService.GetCurrentLibrary(): Promise<MusicLibrary>

// 切换音乐库
MusicService.SwitchLibrary(name: string): Promise<void>

// 刷新音乐库
MusicService.RefreshLibrary(): Promise<void>

// 重命名音乐库
MusicService.RenameLibrary(newName: string): Promise<void>

// 获取所有音乐库
MusicService.GetAllLibraries(): Promise<MusicLibrary[]>

// 加载当前音乐库到播放列表
MusicService.LoadCurrentLibrary(): Promise<void>
```

#### 6.1.3 播放列表 API
```typescript
// 获取播放列表
MusicService.GetPlaylist(): Promise<string[]>

// 添加到播放列表
MusicService.AddToPlaylist(path: string): Promise<void>

// 清空播放列表
MusicService.ClearPlaylist(): Promise<void>
```

### 6.2 TypeScript 类型定义

```typescript
// 音乐文件信息
export interface TrackInfo {
  path: string;
  filename: string;
  title: string;
  artist: string;
  album: string;
  duration: number;
  size: number;
}

// 音乐库
export interface MusicLibrary {
  name: string;
  path: string;
  created_at: string;
  updated_at: string;
  tracks: TrackInfo[];
}

// 播放模式
export type PlayMode = 'order' | 'loop' | 'random';
```

---

## 七、错误处理设计

### 7.1 错误类型

#### 7.1.1 文件操作错误
```go
// 文件不存在
if _, err := os.Stat(path); os.IsNotExist(err) {
    return fmt.Errorf("文件不存在：%s", path)
}

// 权限错误
if err != nil && os.IsPermission(err) {
    return fmt.Errorf("无权限访问：%s", path)
}
```

#### 7.1.2 音频解码错误
```go
// 不支持的格式
return fmt.Errorf("不支持的音频格式：%s", ext)

// 解码失败
return fmt.Errorf("解码音频文件失败：%w", err)
```

#### 7.1.3 业务逻辑错误
```go
// 播放列表为空
if len(playlist) == 0 {
    return fmt.Errorf("播放列表为空")
}

// 索引越界
if index < 0 || index >= len(playlist) {
    return fmt.Errorf("索引越界：%d", index)
}
```

### 7.2 错误处理策略

1. **包装错误**: 使用 `fmt.Errorf("%w", err)` 包装底层错误
2. **日志记录**: 使用 `log.Printf` 记录详细错误信息
3. **用户友好**: 返回给前端的错误信息应简洁明了
4. **资源清理**: 使用 `defer` 确保资源正确释放

---

## 八、性能优化设计

### 8.1 扫描优化

#### 8.1.1 并发扫描
```go
func (lm *LibraryManager) scanDirectory(dirPath string) ([]TrackInfo, error) {
    var tracks []TrackInfo
    var mu sync.Mutex
    var wg sync.WaitGroup
    
    // 并发处理文件
    for _, file := range files {
        wg.Add(1)
        go func(f os.FileInfo) {
            defer wg.Done()
            // 处理文件
        }(file)
    }
    wg.Wait()
    
    return tracks, nil
}
```

#### 8.1.2 增量扫描
- 对比文件修改时间，只扫描新增或变更的文件
- 使用文件 hash 值判断文件是否变化

### 8.2 内存优化

#### 8.2.1 延迟加载
- 音乐库列表只加载元数据，不加载 Tracks
- 点击切换时再加载完整的 Tracks 数据

#### 8.2.2 流式处理
- 使用 beep.Streamer 流式解码，避免一次性加载整个文件
- 播放完成后及时关闭文件句柄

### 8.3 UI 响应优化

#### 8.3.1 异步操作
- 所有耗时操作 (扫描、文件读取) 使用 goroutine 后台执行
- 操作完成后通过事件通知 UI 更新

#### 8.3.2 防抖处理
- 前端对频繁调用 (如音量调节) 进行防抖
- 避免短时间内多次调用后端方法

---

## 九、扩展性设计

### 9.1 支持的音频格式扩展

当前支持：MP3, WAV, FLAC

扩展方案：
```go
// 添加新格式解码器
import "github.com/gopxl/beep/v2/ogg"

case ".ogg":
    streamer, format, err = ogg.Decode(file)
```

### 9.2 元数据读取扩展

当前方案：从文件名提取基础信息

扩展方案：集成 ID3 标签读取库
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

### 9.3 播放效果扩展

当前支持：音量控制 (Gain)

扩展方案：
- 均衡器 (EQ)
- 混响 (Reverb)
- 变速不变调

### 9.4 数据持久化扩展

当前方案：JSON 文件

扩展方案：
- SQLite 数据库 (适合大规模音乐库)
- 云同步 (备份音乐库配置)

---

## 十、安全设计

### 10.1 文件访问安全

1. **路径验证**: 确保访问的文件在允许的目录内
2. **符号链接处理**: 防止通过符号链接访问系统文件
3. **权限检查**: 检查文件读取权限

### 10.2 数据安全

1. **JSON 文件校验**: 加载时验证 JSON 格式
2. **备份机制**: 修改前先备份，失败后恢复
3. **并发写入控制**: 使用锁避免同时写入导致数据损坏

---

## 十一、测试策略

### 11.1 单元测试

```go
func TestPlaylistManager_AddToPlaylist(t *testing.T) {
    pm := NewPlaylistManager()
    err := pm.AddToPlaylist("/path/to/song.mp3")
    if err != nil {
        t.Errorf("添加失败：%v", err)
    }
    
    playlist, _ := pm.GetPlaylist()
    if len(playlist) != 1 {
        t.Errorf("期望播放列表长度为 1, 实际为%d", len(playlist))
    }
}
```

### 11.2 集成测试

```go
func TestMusicService_FullPlayback(t *testing.T) {
    ms := NewMusicService()
    ms.Init()
    
    // 添加音乐库
    ms.AddToLibrary("/test/music")
    
    // 加载播放列表
    ms.LoadCurrentLibrary()
    
    // 播放
    ms.Play()
    
    // 验证播放状态
    if !ms.IsPlaying() {
        t.Error("期望正在播放")
    }
}
```

### 11.3 性能测试

```go
func BenchmarkScanDirectory(b *testing.B) {
    lm := NewLibraryManager()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        lm.scanDirectory("/test/music")
    }
}
```

---

## 十二、部署方案

### 12.1 开发环境

```bash
# 安装依赖
go mod tidy
cd frontend && npm install

# 开发模式运行
wails3 dev
```

### 12.2 生产构建

```bash
# 构建
wails3 build

# 输出目录
build/bin/
```

### 12.3 跨平台编译

```bash
# Windows
GOOS=windows GOARCH=amd64 wails3 build

# macOS
GOOS=darwin GOARCH=amd64 wails3 build

# Linux
GOOS=linux GOARCH=amd64 wails3 build
```

---

## 十三、项目目录结构

```
haoyun-music-player/
── backend/
│   ├── pkg/
│   │   └── file/
│   │       └── file.go          # 文件工具类
│   ├── audioplayer.go           # 音频播放模块
│   ├── com.go                   # 系统服务模块
│   ├── libraryservice.go        # 音乐库管理模块
│   ├── musicsmanager.go         # 播放列表模块
│   ── music_service.go         # 统一服务接口 (新建)
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── Player.vue       # 播放器组件
│   │   │   ├── Playlist.vue     # 播放列表组件
│   │   │   └── LibraryMenu.vue  # 音乐库菜单组件
│   │   ├── App.vue              # 主应用组件
│   │   ├── main.ts              # 入口文件
│   │   └── types/
│   │       └── music.ts         # TypeScript 类型定义
│   ├── public/
│   │   └── style.css
│   └── index.html
├── main.go                      # 应用入口
── go.mod
├── package.json
└── README.md
```

---

## 十四、总结

### 14.1 技术选型优势

1. **Wails v3**: 统一的 Go+Web 技术栈，避免 Electron 的臃肿
2. **Vue3 + TS**: 现代化的前端开发体验，类型安全
3. **beep v2**: 纯 Go 音频处理库，跨平台支持良好
4. **JSON 存储**: 简单灵活，易于调试和备份

### 14.2 架构设计亮点

1. **MVC + Facade**: 清晰的分层架构，易于维护和扩展
2. **依赖注入**: 松耦合设计，便于单元测试
3. **并发安全**: 细粒度锁控制，高性能并发
4. **事件驱动**: 前后端实时通信，响应式 UI

### 14.3 后续优化方向

1. 集成 ID3 标签读取，完善元数据信息
2. 添加歌词显示功能
3. 实现音乐库封面展示
4. 支持播放历史统计
5. 添加音效均衡器
6. 实现云同步功能

---

**文档版本**: v1.0  
**创建时间**: 2026-04-02  
**技术栈**: Wails3 + Vue3 + TypeScript + beep v2 + JSON  
**目标**: 构建一个现代化、高性能的本地音乐播放器
