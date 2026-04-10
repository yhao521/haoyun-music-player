# 🚀 新功能快速参考指南

本文档提供播放历史、歌词显示和专辑封面三个新功能的快速使用参考。

## 📋 目录

- [后端 API 速查](#后端-api-速查)
- [前端调用示例](#前端调用示例)
- [事件监听](#事件监听)
- [常见问题](#常见问题)

---

## 后端 API 速查

### 播放历史

```go
// 获取最近 N 条历史记录
GetPlayHistory(limit int) []HistoryRecord

// 清空历史
ClearPlayHistory() error

// 删除指定记录
RemoveFromPlayHistory(index int) error

// 获取记录数量
GetPlayHistoryCount() int
```

### 歌词管理

```go
// 加载歌词
LoadLyric(trackPath string) (*LyricInfo, error)

// 获取当前歌词行索引
GetCurrentLyricLine(trackPath string, position float64) (int, error)

// 获取所有歌词
GetAllLyrics(trackPath string) ([]LyricLine, error)

// 检查是否有歌词
HasLyric(trackPath string) bool
```

### 专辑封面

```go
// 获取封面 Data URL（推荐）
GetAlbumArtDataURL(trackPath string) (string, error)

// 获取缓存的封面对象
GetCachedCover(trackPath string) *AlbumArt

// 清除封面缓存
ClearCoverCache()
```

---

## 前端调用示例

### TypeScript Bindings 导入路径

```typescript
// 注意：根据实际项目结构调整相对路径
import { 
  GetPlayHistory, 
  LoadLyric, 
  GetAlbumArtDataURL 
} from "../../bindings/github.com/yhao521/wailsMusicPlay/backend/musicservice"

import type { 
  HistoryRecord, 
  LyricInfo, 
  LyricLine 
} from "../../bindings/github.com/yhao521/wailsMusicPlay/backend/models"
```

### 1. 获取播放历史

```typescript
// 获取最近 20 条记录
const history = await GetPlayHistory(20)

// 遍历历史记录
history.forEach(record => {
  console.log(`${record.title} - ${record.artist}`)
  console.log(`播放时间: ${new Date(record.played_at).toLocaleString()}`)
})
```

### 2. 加载并显示歌词

```typescript
// 加载歌词
const lyricInfo = await LoadLyric("/path/to/song.mp3")

if (lyricInfo.has_lyric) {
  // 显示歌词元数据
  console.log(`歌曲: ${lyricInfo.title}`)
  console.log(`艺术家: ${lyricInfo.artist}`)
  console.log(`专辑: ${lyricInfo.album}`)
  console.log(`共 ${lyricInfo.lines.length} 行歌词`)
  
  // 获取当前播放位置的歌词
  const currentIndex = await GetCurrentLyricLine("/path/to/song.mp3", 45.5)
  console.log(`当前歌词: ${lyricInfo.lines[currentIndex].content}`)
  
  // 显示所有歌词
  lyricInfo.lines.forEach(line => {
    const timeStr = formatTime(line.time)
    console.log(`[${timeStr}] ${line.content}`)
  })
} else {
  console.log("暂无歌词")
}

// 辅助函数：格式化时间
function formatTime(seconds: number): string {
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}
```

### 3. 显示专辑封面

```typescript
try {
  // 获取 Data URL
  const dataURL = await GetAlbumArtDataURL("/path/to/song.mp3")
  
  // 在 Vue 模板中使用
  // <img :src="dataURL" alt="Album Art" />
  
  console.log("封面加载成功")
} catch (error) {
  console.log("无专辑封面或提取失败")
  // 显示默认图标
}
```

---

## 事件监听

### 监听播放历史更新

```typescript
import { Events } from "@wailsio/runtime"
import type { HistoryRecord } from "../../bindings/github.com/yhao521/wailsMusicPlay/backend/models"

const unsubscribe = Events.On("historyUpdated", (records: HistoryRecord[]) => {
  console.log(`播放历史已更新，共 ${records.length} 条记录`)
  // 更新 UI
})

// 取消监听
// unsubscribe()
```

### 监听歌词加载完成

```typescript
import type { LyricInfo } from "../../bindings/github.com/yhao521/wailsMusicPlay/backend/models"

Events.On("lyricLoaded", (lyricInfo: LyricInfo) => {
  if (lyricInfo.has_lyric) {
    console.log(`歌词加载成功，共 ${lyricInfo.lines.length} 行`)
  } else {
    console.log("该歌曲没有歌词")
  }
})
```

### 监听当前歌词行变化

```typescript
Events.On("currentLyricLineChanged", (lineIndex: number) => {
  console.log(`当前播放到第 ${lineIndex + 1} 行歌词`)
  // 高亮显示该行
})
```

---

## 常见问题

### Q1: 为什么歌词文件没有被加载？

**A:** 检查以下几点：
1. 歌词文件格式是否为 `.lrc`
2. 文件名是否与音频文件一致（不含扩展名）
3. 歌词文件是否在以下位置之一：
   - 与音频文件同目录
   - `~/.haoyun-music/lyrics/` 目录
4. LRC 格式是否正确（使用文本编辑器检查）

### Q2: 如何手动添加歌词文件？

**A:** 将 `.lrc` 文件放到以下任一位置：
```bash
# 方式 1: 与音频文件同目录
cp song.lrc /path/to/music/song.lrc

# 方式 2: 统一歌词目录
cp song.lrc ~/.haoyun-music/lyrics/song.lrc
```

### Q3: 为什么专辑封面显示不出来？

**A:** 目前 MP3 和 FLAC 的封面提取功能标记为 TODO，需要集成第三方库：
- MP3: 需要 `github.com/bogem/id3v2`
- FLAC: 需要完善 `github.com/mewkiz/flac` 的使用
- 图片缩放: 需要 `github.com/disintegration/imaging`

暂时可以：
1. 手动准备封面图片
2. 使用默认音乐图标作为占位符

### Q4: 播放历史保存在哪里？

**A:** `~/.haoyun-music/history.json`

查看内容：
```bash
cat ~/.haoyun-music/history.json | jq .
```

### Q5: 如何清空播放历史？

**A:** 
```typescript
await ClearPlayHistory()
```

或直接删除文件：
```bash
rm ~/.haoyun-music/history.json
```

### Q6: 如何调整历史记录的最大数量？

**A:** 修改 `backend/historymanager.go` 中的 `maxSize` 字段：
```go
func NewHistoryManager() *HistoryManager {
    return &HistoryManager{
        records: make([]HistoryRecord, 0),
        maxSize: 200, // 改为 200 条
        // ...
    }
}
```

### Q7: 歌词时间不同步怎么办？

**A:** LRC 文件支持 offset 标签来校正时间：
```lrc
[offset:+500]  // 延迟 500 毫秒
[offset:-300]  // 提前 300 毫秒
```

或在歌词文件开头添加：
```lrc
[ti:歌曲名]
[ar:艺术家]
[offset:200]  // 整体偏移 200 毫秒
```

### Q8: 如何清理封面缓存？

**A:** 
```typescript
await ClearCoverCache()
```

或手动删除：
```bash
rm -rf ~/.haoyun-music/covers/*
```

---

## 🔗 相关文档

- [详细实现文档](./NEW_FEATURES.md)
- [项目主文档](./README.md)
- [API 指南](./API_GUIDE.md)

---

<div align="center">

**快速上手，享受音乐！** 🎵

</div>
