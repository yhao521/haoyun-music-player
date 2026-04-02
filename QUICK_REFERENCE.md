# 音乐播放器 - 快速参考卡片

## 🚀 快速开始

### 开发模式
```bash
# 安装依赖
go mod tidy
cd frontend && npm install

# 运行
wails3 dev
```

### 生产构建
```bash
wails3 build
```

---

## 📦 核心模块

| 模块 | 文件 | 功能 |
|------|------|------|
| **AudioPlayer** | `backend/audioplayer.go` | 音频播放 (MP3/WAV/FLAC) |
| **PlaylistManager** | `backend/musicsmanager.go` | 播放列表管理 |
| **LibraryManager** | `backend/libraryservice.go` | 音乐库管理 |
| **MusicService** | `backend/music_service.go` | 统一服务接口 |

---

## 🎵 播放控制 API

```typescript
// 基本控制
MusicService.Play()
MusicService.Pause()
MusicService.Stop()
MusicService.TogglePlayPause()

// 切歌
MusicService.Next()
MusicService.Previous()
MusicService.PlayIndex(0)

// 音量
MusicService.SetVolume(0.8)
MusicService.GetVolume()

// 播放模式
MusicService.SetPlayMode('loop') // order/loop/random
```

---

## 📚 音乐库 API

```typescript
// 添加音乐库
MusicService.AddLibrary()

// 切换音乐库
MusicService.SwitchLibrary('music')

// 刷新音乐库
MusicService.RefreshLibrary()

// 重命名音乐库
MusicService.RenameLibrary('new-name')

// 获取所有音乐库
MusicService.GetLibraries()

// 加载当前音乐库到播放列表
MusicService.LoadCurrentLibrary()
```

---

## 📡 事件监听

```typescript
// 播放状态变化
EventsOn('playbackStateChanged', (state: string) => {
  // state: 'playing' | 'paused' | 'stopped'
})

// 当前歌曲变化
EventsOn('currentTrackChanged', (filename: string) => {
  // filename: 歌曲文件名
})

// 播放列表更新
EventsOn('playlistUpdated', (playlist: string[]) => {
  // playlist: 文件路径数组
})

// 音乐库更新
EventsOn('libraryUpdated', (library: MusicLibrary) => {
  // library: 音乐库对象
})
```

---

## 🏗️ 架构模式

```
┌─────────────────────────────────────┐
│     MusicService (Facade)           │
│  - 统一对外接口                     │
│  - 委托调用子服务                   │
└─────────────────────────────────────┘
         ↓        ↓        ↓
    ┌────┴───┐ ┌─┴──────┐ ┌┴──────
    │Audio   │ │Playlist│ │Library│
    │Player  │ │Manager │ │Manager│
    └────────┘ └────────┘ └───────┘
```

---

## 🔒 线程安全

所有服务使用独立的 `sync.RWMutex`:

```go
// 读操作 (不阻塞其他读)
pm.mu.RLock()
defer pm.mu.RUnlock()

// 写操作 (独占锁)
pm.mu.Lock()
defer pm.mu.Unlock()
```

---

## 📁 数据存储

```
~/.haoyun-music/
└── libraries/
    ├── music.json      // 默认音乐库
    ├── work.json       // 工作音乐库
    └── ...
```

**音乐库 JSON 格式**:
```json
{
  "name": "music",
  "path": "/Users/username/Music",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-02T00:00:00Z",
  "tracks": [
    {
      "path": "/path/to/song.mp3",
      "filename": "song.mp3",
      "title": "Song Title",
      "artist": "Artist",
      "album": "Album",
      "duration": 240,
      "size": 5242880
    }
  ]
}
```

---

## 🎯 支持的音频格式

- ✅ MP3 (.mp3)
- ✅ WAV (.wav)
- ✅ FLAC (.flac)
- ✅ AAC (.aac)
- ✅ OGG (.ogg)
- ✅ WMA (.wma)

---

## 🛠️ 技术栈

| 层级 | 技术 | 版本 |
|------|------|------|
| 后端框架 | Wails | v3.0.0-alpha.74 |
| 后端语言 | Go | 1.25 |
| 音频处理 | beep | v2 |
| 前端框架 | Vue | 3 |
| 前端语言 | TypeScript | - |
| 构建工具 | Vite | - |

---

## 📖 文档索引

| 文档 | 说明 |
|------|------|
| [BACKEND_DESIGN.md](BACKEND_DESIGN.md) | 后台方案设计 (24KB) |
| [README_IMPLEMENTATION.md](README_IMPLEMENTATION.md) | 实现总结 (9KB) |
| [IMPLEMENTATION_SUMMARY.md](IMPLEMENTATION_SUMMARY.md) | 详细实现文档 (11KB) |
| [API_GUIDE.md](API_GUIDE.md) | API 使用指南 (15KB) |

---

## 🔍 常见问题

### Q: 播放列表为空怎么办？
```typescript
try {
  await MusicService.Play()
} catch (error) {
  // 先加载音乐库
  await MusicService.LoadCurrentLibrary()
}
```

### Q: 如何获取当前播放的歌曲？
```typescript
const playlist = await MusicService.GetPlaylist()
const index = await MusicService.GetCurrentIndex()
const current = playlist[index]
```

### Q: 如何添加音乐库？
```typescript
// 打开目录选择对话框
await MusicService.AddLibrary()

// 或指定路径
await MusicService.AddToLibrary('/path/to/music')
```

---

## 📊 代码统计

| 模块 | 大小 | 功能 |
|------|------|------|
| audioplayer.go | 5.4KB | 音频播放 |
| musicsmanager.go | 3.8KB | 播放列表 |
| libraryservice.go | 9.6KB | 音乐库 |
| music_service.go | 7.4KB | 统一接口 |
| **总计** | **26.2KB** | **4 个核心模块** |

---

## ✅ 功能清单

### 播放控制
- [x] 播放/暂停/停止
- [x] 上一首/下一首
- [x] 指定索引播放
- [x] 音量调节
- [x] 播放模式切换

### 播放列表
- [x] 添加/清空
- [x] 获取列表
- [x] 索引跟踪
- [x] 自动更新

### 音乐库
- [x] 添加库
- [x] 切换库
- [x] 刷新库
- [x] 重命名库
- [x] 多库支持
- [x] JSON 存储

---

## 🎓 最佳实践

### 1. 使用 Facade 模式
```go
// ✅ 推荐：通过 MusicService 统一调用
musicService.Play()

// ❌ 不推荐：直接调用底层服务
audioPlayer.Play()
```

### 2. 事件驱动更新
```typescript
// ✅ 推荐：监听事件自动更新
EventsOn('playlistUpdated', () => {
  loadPlaylist()
})

// ❌ 不推荐：轮询检查
setInterval(async () => {
  const playlist = await MusicService.GetPlaylist()
}, 1000)
```

### 3. 错误处理
```typescript
// ✅ 推荐：完整的错误处理
try {
  await MusicService.Play()
} catch (error) {
  console.error('播放失败:', error)
  alert('播放失败，请检查播放列表')
}
```

---

## 🚀 下一步

- [ ] 前端 Vue3 组件开发
- [ ] ID3 标签读取
- [ ] 播放进度显示
- [ ] 歌词显示
- [ ] 音效均衡器

---

**版本**: v1.0  
**更新日期**: 2026-04-02  
**状态**: 后台完成 ✅，前端开发中 ⏳
