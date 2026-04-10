# 🎵 新功能实现文档

本文档介绍最近实现的三个新功能：**播放历史记录**、**歌词显示（LRC 支持）**和**专辑封面提取**。

## 📋 目录

- [1. 播放历史记录](#1-播放历史记录)
- [2. 歌词显示（LRC 文件支持）](#2-歌词显示lrc-文件支持)
- [3. 专辑封面提取和显示](#3-专辑封面提取和显示)
- [4. API 使用示例](#4-api-使用示例)
- [5. 前端集成指南](#5-前端集成指南)
- [6. 增强功能（2026-04-05）](#6-增强功能2026-04-05)

---

## 1. 播放历史记录

### 功能概述

自动记录用户播放过的每首歌曲，保存播放时间、歌曲信息等，方便快速回顾和重新播放。

### 数据存储

- **存储位置**: `~/.haoyun-music/history.json`
- **最大记录数**: 默认 100 条（可配置）
- **存储格式**: JSON
- **更新策略**: 同一歌曲重复播放时更新时间戳和播放次数，不新增记录

### 数据结构

```go
type HistoryRecord struct {
    Path       string    `json:"path"`        // 歌曲路径
    Title      string    `json:"title"`       // 标题
    Artist     string    `json:"artist"`      // 艺术家
    Album      string    `json:"album"`       // 专辑
    PlayedAt   time.Time `json:"played_at"`   // 播放时间
    Duration   int64     `json:"duration"`    // 播放时长（秒）
    FileSize   int64     `json:"file_size"`   // 文件大小（字节）
    PlayCount  int       `json:"play_count"`  // 播放次数（新增）
}
```

### 后端 API

```go
// 获取播放历史（最近 N 条）
GetPlayHistory(limit int) []HistoryRecord

// 获取喜爱音乐（按播放次数排序）
GetFavoriteTracks(limit int) []HistoryRecord

// 清空播放历史
ClearPlayHistory() error

// 删除指定索引的历史记录
RemoveFromPlayHistory(index int) error

// 获取历史记录数量
GetPlayHistoryCount() int
```

### 事件通知

当历史记录更新时，会触发 `historyUpdated` 事件：

```go
application.RegisterEvent[[]backend.HistoryRecord]("historyUpdated")
```

### 使用场景

1. **自动记录**: 每次调用 `Play()` 方法时自动添加记录
2. **历史列表**: 在 UI 中展示最近播放的歌曲
3. **快速重播**: 点击历史记录可直接播放该歌曲
4. **统计分析**: 分析用户的听歌习惯和偏好
5. **喜爱音乐**: 按播放次数排序，展示最爱听的歌曲

---

## 2. 歌词显示（LRC 文件支持）

### 功能概述

支持标准 LRC 格式的歌词文件，能够根据当前播放位置实时同步显示歌词，并支持滚动高亮。

### LRC 文件格式

```lrc
[ti:歌曲标题]
[ar:艺术家]
[al:专辑]
[offset:+/-毫秒]

[00:12.34]第一句歌词
[00:15.67]第二句歌词
[00:19.00]第三句歌词
```

### 搜索策略

1. **优先**: 同目录下的 `歌曲名.lrc`
2. **其次**: `~/.haoyun-music/lyrics/歌曲名.lrc`

### 数据结构

```go
type LyricLine struct {
    Time    float64 `json:"time"`    // 时间点（秒）
    Content string  `json:"content"` // 歌词内容
}

type LyricInfo struct {
    Title    string      `json:"title"`     // 歌曲标题
    Artist   string      `json:"artist"`    // 艺术家
    Album    string      `json:"album"`     // 专辑
    Offset   float64     `json:"offset"`    // 时间偏移量（秒）
    Lines    []LyricLine `json:"lines"`     // 歌词行列表
    HasLyric bool        `json:"has_lyric"` // 是否有歌词
}
```

### 后端 API

```go
// 加载歌词文件
LoadLyric(trackPath string) (*LyricInfo, error)

// 获取当前时间点的歌词行索引
GetCurrentLyricLine(trackPath string, position float64) (int, error)

// 获取所有歌词行
GetAllLyrics(trackPath string) ([]LyricLine, error)

// 检查是否有歌词
HasLyric(trackPath string) bool
```

### 事件通知

```go
// 歌词加载完成
application.RegisterEvent[*backend.LyricInfo]("lyricLoaded")

// 当前歌词行变化
application.RegisterEvent[int]("currentLyricLineChanged")
```

### 特性

- ✅ 支持标准 LRC 格式（`[mm:ss.xx]` 和 `[mm:ss:xxx]`）
- ✅ 解析元数据标签（ti, ar, al, offset）
- ✅ 自动时间偏移校正
- ✅ 歌词行按时间排序
- ✅ 缓存机制，避免重复解析
- ✅ 容错处理，无歌词时优雅降级

---

## 3. 专辑封面提取和显示

### 功能概述

从音频文件中提取专辑封面图片，支持 MP3（ID3v2 APIC）和 FLAC（METADATA_BLOCK_PICTURE）格式，并提供缓存机制以提升性能。

### 支持的格式

- **MP3**: ID3v2 APIC 帧（JPEG/PNG）
- **FLAC**: METADATA_BLOCK_PICTURE
- **WAV**: 暂不支持

### 缓存机制

- **内存缓存**: `map[string]*AlbumArt`
- **文件缓存**: `~/.haoyun-music/covers/`
- **缓存键**: 文件路径的 MD5 哈希值
- **缓存文件**: 
  - `.dat` - 图片二进制数据
  - `.json` - 元数据（MIME 类型、尺寸）

### 数据结构

```go
type AlbumArt struct {
    Data     []byte `json:"data"`      // 图片二进制数据
    MimeType string `json:"mime_type"` // MIME 类型 (image/jpeg, image/png)
    Width    int    `json:"width"`     // 宽度
    Height   int    `json:"height"`    // 高度
}
```

### 后端 API

```go
// 获取专辑封面的 Data URL（可直接用于 img src）
GetAlbumArtDataURL(trackPath string) (string, error)

// 获取缓存的封面
GetCachedCover(trackPath string) *AlbumArt

// 清除封面缓存
ClearCoverCache()
```

### 降级策略

1. 无封面时返回 `nil`
2. 提取失败时使用缓存
3. 缓存未命中时显示默认图标

### 待实现功能

⚠️ **注意**: 以下功能标记为 TODO，需要后续完善：

1. **MP3 封面提取**: 需要集成 `github.com/bogem/id3v2` 库
2. **FLAC 封面提取**: 需要完善 `github.com/mewkiz/flac` 库的使用
3. **图片缩放**: 需要集成 `github.com/disintegration/imaging` 库进行真正的缩放

目前这些方法会返回错误，但不影响其他功能的使用。

---

## 4. API 使用示例

### 播放历史记录

```typescript
import { GetPlayHistory, ClearPlayHistory } from "../bindings/github.com/yhao521/wailsMusicPlay/backend/musicservice"

// 获取最近 20 条播放历史
const history = await GetPlayHistory(20)
console.log(history) // Array of HistoryRecord

// 清空历史
await ClearPlayHistory()
```

### 歌词显示

```typescript
import { LoadLyric, GetCurrentLyricLine, GetAllLyrics } from "../bindings/github.com/yhao521/wailsMusicPlay/backend/musicservice"

// 加载歌词
const lyricInfo = await LoadLyric("/path/to/song.mp3")
if (lyricInfo.has_lyric) {
  console.log(`共有 ${lyricInfo.lines.length} 行歌词`)
  
  // 获取当前播放位置的歌词行
  const currentLineIndex = await GetCurrentLyricLine("/path/to/song.mp3", 45.5)
  console.log(`当前歌词: ${lyricInfo.lines[currentLineIndex].content}`)
  
  // 获取所有歌词
  const allLines = await GetAllLyrics("/path/to/song.mp3")
  allLines.forEach(line => {
    console.log(`[${line.time.toFixed(2)}] ${line.content}`)
  })
}
```

### 专辑封面

```typescript
import { GetAlbumArtDataURL } from "../bindings/github.com/yhao521/wailsMusicPlay/backend/musicservice"

// 获取封面 Data URL
try {
  const dataURL = await GetAlbumArtDataURL("/path/to/song.mp3")
  // 直接在 img 标签中使用
  // <img :src="dataURL" alt="Album Art" />
} catch (error) {
  console.log("无专辑封面")
}
```

---

## 5. 前端集成指南

### Vue 组件示例

#### 播放历史列表组件

```vue
<template>
  <div class="history-list">
    <h3>播放历史</h3>
    <ul>
      <li v-for="(record, index) in history" :key="index" @click="playFromHistory(record.path)">
        <div class="track-info">
          <span class="title">{{ record.title }}</span>
          <span class="artist">{{ record.artist }}</span>
        </div>
        <span class="time">{{ formatTime(record.played_at) }}</span>
        <span class="count">播放 {{ record.play_count }} 次</span>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue"
import { GetPlayHistory, PlayIndex, AddToPlaylist } from "../bindings/github.com/yhao521/wailsMusicPlay/backend/musicservice"
import type { HistoryRecord } from "../bindings/github.com/yhao521/wailsMusicPlay/backend/models"

const history = ref<HistoryRecord[]>([])

const loadHistory = async () => {
  history.value = await GetPlayHistory(20)
}

const playFromHistory = async (path: string) => {
  await AddToPlaylist(path)
  // 播放最后一首（刚添加的）
  const playlist = await GetPlaylist()
  await PlayIndex(playlist.length - 1)
}

const formatTime = (timestamp: string) => {
  return new Date(timestamp).toLocaleString('zh-CN')
}

onMounted(() => {
  loadHistory()
})
</script>
```

#### 歌词显示组件

```vue
<template>
  <div class="lyric-container">
    <div v-if="!lyricInfo?.has_lyric" class="no-lyric">
      <p>暂无歌词</p>
    </div>
    <div v-else class="lyric-content">
      <div 
        v-for="(line, index) in lyricInfo.lines" 
        :key="index"
        class="lyric-line"
        :class="{ active: index === currentLineIndex }"
      >
        {{ line.content }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from "vue"
import { LoadLyric, GetCurrentLyricLine, GetAllLyrics } from "../bindings/github.com/yhao521/wailsMusicPlay/backend/musicservice"
import { Events } from "@wailsio/runtime"
import type { LyricInfo } from "../bindings/github.com/yhao521/wailsMusicPlay/backend/models"

const props = defineProps<{
  trackPath: string
  currentPosition: number
}>()

const lyricInfo = ref<LyricInfo | null>(null)
const currentLineIndex = ref(-1)

const loadLyric = async () => {
  lyricInfo.value = await LoadLyric(props.trackPath)
  if (lyricInfo.value?.has_lyric) {
    updateCurrentLine()
  }
}

const updateCurrentLine = async () => {
  if (!lyricInfo.value || !lyricInfo.value.has_lyric) return
  
  try {
    currentLineIndex.value = await GetCurrentLyricLine(
      props.trackPath, 
      props.currentPosition
    )
    
    // 自动滚动到当前歌词行
    setTimeout(() => {
      const activeElement = document.querySelector('.lyric-line.active')
      if (activeElement) {
        activeElement.scrollIntoView({ behavior: 'smooth', block: 'center' })
      }
    }, 100)
  } catch (error) {
    console.error('获取当前歌词行失败:', error)
  }
}

// 监听播放位置变化
watch(() => props.currentPosition, () => {
  updateCurrentLine()
})

// 监听歌曲切换
watch(() => props.trackPath, () => {
  loadLyric()
})

onMounted(() => {
  loadLyric()
})
</script>

<style scoped>
.lyric-container {
  height: 300px;
  overflow-y: auto;
  text-align: center;
  padding: 20px;
}

.lyric-line {
  padding: 8px 0;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.5);
  transition: all 0.3s ease;
}

.lyric-line.active {
  font-size: 18px;
  font-weight: bold;
  color: #fff;
  transform: scale(1.1);
}

.no-lyric {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  opacity: 0.5;
}
</style>
```

#### 专辑封面组件

```vue
<template>
  <div class="album-art">
    <img 
      v-if="coverDataURL" 
      :src="coverDataURL" 
      alt="Album Art"
      @error="handleImageError"
    />
    <div v-else class="default-cover">
      <span>🎵</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from "vue"
import { GetAlbumArtDataURL } from "../bindings/github.com/yhao521/wailsMusicPlay/backend/musicservice"

const props = defineProps<{
  trackPath: string
}>()

const coverDataURL = ref<string>("")

const loadCover = async () => {
  try {
    coverDataURL.value = await GetAlbumArtDataURL(props.trackPath)
  } catch (error) {
    console.log("无专辑封面或提取失败")
    coverDataURL.value = ""
  }
}

const handleImageError = () => {
  coverDataURL.value = ""
}

watch(() => props.trackPath, () => {
  loadCover()
})

loadCover()
</script>

<style scoped>
.album-art {
  width: 200px;
  height: 200px;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
}

.album-art img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.default-cover {
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 64px;
}
</style>
```

---

## 6. 增强功能（2026-04-05）

### 6.1 托盘菜单"喜爱音乐"

#### 功能说明

在系统托盘中新增"❤️ 喜爱音乐"子菜单，按照歌曲播放次数从高到低排序，最多显示前 20 首。

#### 实现细节

- **数据来源**: 从播放历史中提取，按 `PlayCount` 字段降序排列
- **显示格式**: `序号. 歌曲名 - 艺术家 (播放次数)`
- **交互**: 点击直接添加到播放列表并播放
- **管理**: 提供"清空播放历史"选项

#### 代码示例

```go
// 获取前 20 首喜爱音乐
favorites := musicService.GetFavoriteTracks(20)

for i, record := range favorites {
    label := fmt.Sprintf("%d. %s - %s (%d次)", i+1, displayName, record.Artist, record.PlayCount)
    favItem := application.NewMenuItem(label)
    favItem.OnClick(func(ctx *application.Context) {
        musicService.AddToPlaylist(record.Path)
        playlist, _ := musicService.GetPlaylist()
        if len(playlist) > 0 {
            musicService.PlayIndex(len(playlist) - 1)
        }
    })
}
```

### 6.2 音乐库管理增强

#### 删除音乐库功能

在"音乐库"子菜单中新增"删除当前音乐库"选项：

- **作用范围**: 仅删除音乐库配置（JSON 文件），不删除实际的音乐文件
- **安全提示**: 删除前会在日志中记录警告信息
- **自动重建**: 删除后自动刷新托盘菜单
- **当前库处理**: 如果删除的是当前选中的音乐库，会自动切换到另一个库或清空选择

#### 代码示例

```go
deleteLibItem := application.NewMenuItem("删除当前音乐库")
deleteLibItem.OnClick(func(ctx *application.Context) {
    currentLib := musicService.GetCurrentLibrary()
    if currentLib == nil {
        log.Println("当前没有音乐库")
        return
    }
    
    libName := currentLib.Name
    if err := musicService.DeleteLibrary(libName); err != nil {
        log.Printf("删除音乐库失败：%v", err)
        return
    }
    
    log.Printf("✓ 已删除音乐库：%s", libName)
    buildMusicLibMenu() // 重建菜单
})
```

### 6.3 歌词文件扫描优化

#### 功能说明

在新增和刷新音乐库时，不仅扫描音乐文件，同时扫描对应的歌词文件（`.lrc`），并建立关联关系。

#### 扫描策略

1. **两阶段扫描**:
   - 第一阶段：扫描所有 `.lrc` 文件，建立 `歌曲名 -> 歌词路径` 映射表
   - 第二阶段：扫描音乐文件，查找对应的歌词文件

2. **匹配规则**: 通过文件名（不含扩展名）进行精确匹配
   - `song.mp3` ↔ `song.lrc`
   - `周杰伦 - 晴天.flac` ↔ `周杰伦 - 晴天.lrc`

3. **数据存储**: 在 `TrackInfo` 结构体中新增 `LyricPath` 字段

#### 数据结构更新

```go
type TrackInfo struct {
    Path      string `json:"path"`
    Filename  string `json:"filename"`
    Title     string `json:"title"`
    Artist    string `json:"artist"`
    Album     string `json:"album"`
    Duration  int64  `json:"duration"`
    Size      int64  `json:"size"`
    LyricPath string `json:"lyric_path"` // 新增：歌词文件路径
}
```

#### 扫描逻辑

```go
// 第一阶段：扫描歌词文件
lyricMap := make(map[string]string)
filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
    if ext := strings.ToLower(filepath.Ext(path)); ext == ".lrc" {
        baseName := strings.TrimSuffix(info.Name(), ext)
        lyricMap[baseName] = path
    }
    return nil
})

// 第二阶段：扫描音乐文件并关联歌词
filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
    if supportedFormats[ext] {
        baseName := strings.TrimSuffix(info.Name(), ext)
        lyricPath := lyricMap[baseName] // 查找对应歌词
        
        track := TrackInfo{
            Path:      path,
            Title:     baseName,
            LyricPath: lyricPath, // 保存歌词路径
            // ... 其他字段
        }
        tracks = append(tracks, track)
    }
    return nil
})
```

#### 优势

- ✅ **自动化**: 无需手动配置歌词路径
- ✅ **高效**: 一次扫描完成音乐和歌词的关联
- ✅ **灵活**: 支持歌词文件与音乐文件在同一目录或不同目录
- ✅ **容错**: 即使没有歌词文件也不影响音乐播放

---

## 📝 总结

这三个新功能大大增强了音乐播放器的用户体验：

1. **播放历史** - 方便用户回顾和重新播放喜欢的歌曲
2. **歌词显示** - 提供卡拉 OK 式的同步歌词体验
3. **专辑封面** - 让播放器界面更加美观和专业

### 下一步优化建议

1. **完善封面提取**: 集成 id3v2 和 imaging 库实现完整的封面提取和缩放功能
2. **歌词编辑**: 支持在线编辑和保存歌词
3. **历史统计**: 添加听歌时长统计、最爱歌曲等功能
4. **歌词翻译**: 支持双语歌词显示
5. **封面下载**: 从网络 API 自动下载缺失的专辑封面

---

<div align="center">

**Made with ❤️ by Yang Hao**

🎵 Enjoy Your Music!

</div>
