# 元数据获取优化

## 概述

本次优化实现了从音频文件中读取真实元数据的功能，替代了之前仅从文件名提取信息的简单实现。

## 主要改进

### 1. 新增 MetadataManager

创建了专门的元数据管理器 [`metadatamanager.go`](backend/metadatamanager.go)，提供以下功能：

- **MP3 支持**：读取 ID3v1 和 ID3v2 标签
  - 标题（TIT2）
  - 艺术家（TPE1）
  - 专辑（TALB）
  - 年份（TYER/TDRC）
  - 流派（TCON）
  - 音轨号（TRCK）
  - 注释（COMM）

- **FLAC 支持**：读取 Vorbis Comment
  - TITLE
  - ARTIST
  - ALBUM
  - DATE
  - GENRE
  - TRACKNUMBER
  - COMMENT/DESCRIPTION

- **缓存机制**：自动缓存已读取的元数据，提高重复访问性能
- **降级处理**：如果无法读取元数据，自动回退到基于文件名的基本信息

### 2. 集成到 MusicService

更新了 [`music_service.go`](backend/music_service.go) 中的 `GetSongMetadata` 方法：

```go
// 之前：仅返回文件名
func (m *MusicService) GetSongMetadata(path string) (map[string]interface{}, error) {
    filename := filepath.Base(path)
    return map[string]interface{}{
        "title":  filename,
        "artist": "未知艺术家",
        "album":  "未知专辑",
        "path":   path,
    }, nil
}

// 现在：读取真实元数据
func (m *MusicService) GetSongMetadata(path string) (map[string]interface{}, error) {
    if m.metadataManager == nil {
        // 降级处理
        return basicMetadata, nil
    }
    
    // 使用元数据管理器获取详细的元数据
    return m.metadataManager.GetMetadata(path)
}
```

### 3. 新增 API

添加了清除元数据缓存的方法：

```go
// ClearMetadataCache 清除元数据缓存
func (m *MusicService) ClearMetadataCache() {
    if m.metadataManager != nil {
        m.metadataManager.ClearCache()
    }
}
```

## 技术实现

### MP3 元数据读取

#### ID3v2 解析
- 读取 10 字节头部（包含 "ID3" 标识、版本、标签大小）
- 解析帧结构（10 字节帧头 + 帧数据）
- 支持多种文本编码（ISO-8859-1、UTF-16、UTF-8）

#### ID3v1 解析
- 读取文件末尾 128 字节
- 检查 "TAG" 标识
- 解析固定长度字段（标题 30 字节、艺术家 30 字节等）

### FLAC 元数据读取

- 验证 "fLaC" 文件头部
- 遍历元数据块（STREAMINFO、VORBIS_COMMENT、PICTURE 等）
- 解析 Vorbis Comment 结构：
  - Vendor string
  - 评论数量
  - 键值对列表（UTF-8 编码）

## 使用示例

```go
// 获取歌曲元数据
metadata, err := musicService.GetSongMetadata("/path/to/song.mp3")
if err != nil {
    log.Printf("获取元数据失败：%v", err)
}

// 访问元数据字段
fmt.Println("标题:", metadata["title"])
fmt.Println("艺术家:", metadata["artist"])
fmt.Println("专辑:", metadata["album"])
fmt.Println("年份:", metadata["year"])
fmt.Println("流派:", metadata["genre"])
fmt.Println("音轨:", metadata["track"])
```

## 性能优化

- **内存缓存**：使用 `sync.RWMutex` 保护并发访问
- **懒加载**：仅在首次访问时读取元数据
- **智能降级**：读取失败时自动使用基本信息，不影响用户体验

## 兼容性

- ✅ MP3 文件（ID3v1/v2）
- ✅ FLAC 文件（Vorbis Comment）
- ⚠️ 其他格式（WAV、OGG 等）：返回基于文件名的基本信息

## 未来改进

1. 添加 WAV 文件的 INFO chunk 支持
2. 添加 OGG Vorbis 文件的元数据读取
3. 考虑集成 `github.com/dhowden/tag` 库以获得更广泛的格式支持
4. 添加元数据编辑功能
5. 支持从在线数据库（如 MusicBrainz）获取元数据

## 测试建议

使用包含完整 ID3/Vorbis 标签的测试文件验证：
- MP3 文件测试 ID3v1 和 ID3v2
- FLAC 文件测试 Vorbis Comment
- 损坏或不完整的标签文件测试降级处理
- 大文件库的性能测试
