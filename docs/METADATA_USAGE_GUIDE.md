# 元数据管理器使用指南

## 概述

[`MetadataManager`](backend/metadatamanager.go) 是一个专门用于读取音频文件元数据的组件，支持 MP3（ID3v1/v2）和 FLAC（Vorbis Comment）格式。

## 快速开始

### 基本用法

```go
package main

import (
    "fmt"
    "log"
    "github.com/yhao521/haoyun-music-player/backend"
)

func main() {
    // 创建元数据管理器
    metadataManager := backend.NewMetadataManager()

    // 获取歌曲元数据
    metadata, err := metadataManager.GetMetadata("/path/to/song.mp3")
    if err != nil {
        log.Printf("警告：获取元数据失败：%v", err)
    }

    // 访问元数据字段
    fmt.Println("标题:", metadata["title"])
    fmt.Println("艺术家:", metadata["artist"])
    fmt.Println("专辑:", metadata["album"])
    fmt.Println("年份:", metadata["year"])
    fmt.Println("流派:", metadata["genre"])
    fmt.Println("音轨号:", metadata["track"])
    fmt.Println("注释:", metadata["comment"])
    fmt.Println("文件路径:", metadata["path"])
}
```

### 在 MusicService 中使用

```go
// 在 MusicService 中已经集成了 MetadataManager
musicService := backend.NewMusicService()

// 直接调用 GetSongMetadata
metadata, err := musicService.GetSongMetadata("/path/to/song.flac")
if err == nil {
    fmt.Printf("歌曲信息：%s - %s\n", metadata["artist"], metadata["title"])
}

// 清除元数据缓存（例如在刷新音乐库后）
musicService.ClearMetadataCache()
```

## 支持的格式

### MP3 文件

支持读取 ID3v1 和 ID3v2 标签：

| 字段    | ID3v2 帧  | ID3v1 位置 | 说明     |
| ------- | --------- | ---------- | -------- |
| title   | TIT2      | 3-32       | 歌曲标题 |
| artist  | TPE1      | 33-62      | 艺术家   |
| album   | TALB      | 63-92      | 专辑名称 |
| year    | TYER/TDRC | 93-96      | 年份     |
| genre   | TCON      | -          | 流派     |
| track   | TRCK      | -          | 音轨号   |
| comment | COMM      | 97-126     | 注释     |

**示例输出：**

```
map[
  album:Abbey Road
  artist:The Beatles
  comment:Remastered 2009
  genre:Rock
  path:/music/beatles/come_together.mp3
  title:Come Together
  track:1
  year:1969
]
```

### FLAC 文件

支持读取 Vorbis Comment：

| 字段    | Vorbis Key          | 说明     |
| ------- | ------------------- | -------- |
| title   | TITLE               | 歌曲标题 |
| artist  | ARTIST              | 艺术家   |
| album   | ALBUM               | 专辑名称 |
| year    | DATE                | 年份     |
| genre   | GENRE               | 流派     |
| track   | TRACKNUMBER         | 音轨号   |
| comment | COMMENT/DESCRIPTION | 注释     |

**示例输出：**

```
map[
  album:Random Access Memories
  artist:Daft Punk
  comment:
  genre:Electronic
  path:/music/daft_punk/get_lucky.flac
  title:Get Lucky
  track:8
  year:2013
]
```

### 其他格式

对于不支持的格式（如 WAV、OGG 等），会自动降级为基于文件名的基本信息：

```go
metadata, _ := metadataManager.GetMetadata("/music/unknown.wav")
// 返回：
// map[
//   album:未知专辑
//   artist:未知艺术家
//   comment:
//   genre:
//   path:/music/unknown.wav
//   title:unknown  (从文件名提取，不含扩展名)
//   track:
//   year:
// ]
```

## 高级功能

### 缓存机制

`MetadataManager` 内置了内存缓存，自动缓存已读取的元数据：

```go
// 第一次调用：从文件读取
metadata1, _ := metadataManager.GetMetadata("/path/to/song.mp3")

// 第二次调用：从缓存读取（更快）
metadata2, _ := metadataManager.GetMetadata("/path/to/song.mp3")

// 两次返回的是相同的数据
```

### 清除缓存

在以下情况下可能需要清除缓存：

- 用户编辑了音频文件的元数据
- 需要释放内存
- 刷新音乐库后

```go
// 清除所有缓存
metadataManager.ClearCache()

// 或者通过 MusicService
musicService.ClearMetadataCache()
```

### 并发安全

`MetadataManager` 使用 `sync.RWMutex` 保证并发安全，可以在多个 goroutine 中安全使用：

```go
go func() {
    metadata, _ := metadataManager.GetMetadata("/path/to/song1.mp3")
    // 处理元数据...
}()

go func() {
    metadata, _ := metadataManager.GetMetadata("/path/to/song2.mp3")
    // 处理元数据...
}()
```

## 错误处理

当元数据读取失败时，会自动降级为基本信息，不会中断程序：

```go
metadata, err := metadataManager.GetMetadata("/corrupted/file.mp3")
if err != nil {
    log.Printf("警告：无法读取完整元数据：%v", err)
    log.Printf("使用基本信息：%v", metadata["title"])
}
// 仍然可以使用 metadata，至少包含 path 和 title
```

常见错误情况：

- 文件不存在或无法访问
- 文件格式损坏
- 没有嵌入元数据标签
- 不支持的文件格式

## 性能优化建议

1. **复用 MetadataManager 实例**

   ```go
   // ✅ 推荐：创建一次，多次使用
   mm := backend.NewMetadataManager()
   for _, song := range playlist {
       metadata, _ := mm.GetMetadata(song.Path)
   }

   // ❌ 不推荐：每次都创建新实例
   for _, song := range playlist {
       mm := backend.NewMetadataManager()
       metadata, _ := mm.GetMetadata(song.Path)
   }
   ```

2. **批量处理时利用缓存**

   ```go
   // 首次扫描会填充缓存
   for _, song := range library {
       metadata, _ := mm.GetMetadata(song.Path)
   }

   // 后续访问直接从缓存读取
   for _, song := range favorites {
       metadata, _ := mm.GetMetadata(song.Path) // 快速！
   }
   ```

3. **适时清除缓存**
   ```go
   // 定期清理或内存紧张时
   if memoryUsage > threshold {
       mm.ClearCache()
   }
   ```

## 调试技巧

### 查看原始元数据

```go
metadata, err := metadataManager.GetMetadata("/path/to/song.mp3")
if err != nil {
    log.Printf("错误：%v", err)
}

// 打印所有元数据字段
for key, value := range metadata {
    log.Printf("%s: %v", key, value)
}
```

### 测试特定文件

```go
// 测试 MP3 文件
mp3Metadata, _ := metadataManager.GetMetadata("test.mp3")
log.Printf("MP3 元数据： %+v", mp3Metadata)

// 测试 FLAC 文件
flacMetadata, _ := metadataManager.GetMetadata("test.flac")
log.Printf("FLAC 元数据： %+v", flacMetadata)
```

## 单元测试

运行单元测试验证功能：

```bash
go test -v ./backend -run TestMetadataManager
```

测试覆盖：

- ✅ 基本元数据获取
- ✅ 缓存功能
- ✅ 清除缓存
- ✅ MP3 文件（ID3v1）
- ✅ FLAC 文件（Vorbis Comment）

## 未来改进方向

1. **扩展格式支持**
   - WAV 文件（INFO chunk）
   - OGG Vorbis 文件
   - AAC/M4A 文件（iTunes 标签）

2. **增强功能**
   - 元数据编辑支持
   - 从在线数据库自动获取（MusicBrainz、Discogs）
   - 专辑封面提取集成

3. **性能优化**
   - 异步预加载常用元数据
   - 磁盘缓存持久化
   - 更智能的缓存淘汰策略

## 技术细节

### ID3v2 解析流程

1. 读取 10 字节头部
   - 标识符："ID3"（3 字节）
   - 版本号（2 字节）
   - 标志位（1 字节）
   - 标签大小（4 字节，同步安全整数）

2. 遍历帧列表
   - 帧 ID（4 字节）
   - 帧大小（4 字节）
   - 标志位（2 字节）
   - 帧数据（可变长度）

3. 解码文本帧
   - 检测编码类型（ISO-8859-1、UTF-16、UTF-8）
   - 转换为 UTF-8 字符串
   - 去除空字符和空白

### Vorbis Comment 解析流程

1. 验证 FLAC 头部（"fLaC"）

2. 遍历元数据块
   - 读取块头部（4 字节）
     - 最后一个块标志（1 bit）
     - 块类型（7 bits）
     - 块大小（24 bits）
   - 读取块数据

3. 解析 Vorbis Comment（类型 4）
   - Vendor string 长度 + 内容
   - 评论数量
   - 遍历评论列表
     - 评论长度
     - 评论内容（KEY=VALUE 格式）

## 常见问题

### Q: 为什么有些 MP3 文件读不到元数据？

A: 可能的原因：

- 文件没有嵌入 ID3 标签
- 使用的是 ID3v2.4 但解析器只支持 v2.3
- 标签使用了非标准编码
- 文件损坏

解决方案：使用工具（如 Mp3tag）重新写入标签。

### Q: FLAC 文件的元数据在哪里查看？

A: 可以使用以下工具：

```bash
# 使用 metaflac 命令行工具
metaflac --export-tags-to=- song.flac

# 或使用 ffprobe
ffprobe -show_format song.flac
```

### Q: 缓存占用多少内存？

A: 每个元数据条目约 200-500 字节。1000 首歌曲约占用 0.2-0.5 MB。可以通过 `ClearCache()` 手动释放。

### Q: 如何贡献代码？

A: 欢迎提交 PR 来：

- 添加新的音频格式支持
- 改进解析准确性
- 优化性能
- 修复 bug

## 相关资源

- [ID3v2 规范](https://id3.org/id3v2.3.0)
- [Vorbis Comment 规范](https://wiki.xiph.org/VorbisComment)
- [FLAC 格式规范](https://xiph.org/flac/format.html)
- [Go 音频处理库](https://pkg.go.dev/github.com/dhowden/tag)

## 许可证

本项目采用 MIT 许可证。
