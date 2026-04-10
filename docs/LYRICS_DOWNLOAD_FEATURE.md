# 歌词下载功能 - 多源集成与智能降级

## 功能概述

为音乐播放器添加了从多个在线源下载歌词的功能,支持**智能降级策略**,提高下载成功率。支持单首歌曲和批量下载整个音乐库的歌词。

### 🎯 支持的 API 源

1. **lrclib.net** (主要源)
   - ✅ 免费开放 API
   - ✅ 支持同步歌词和纯文本歌词
   - ✅ 无需认证
   - ✅ 国际化支持好

2. **网易云音乐** (备用源 1)
   - ✅ 中文歌曲覆盖率高
   - ✅ 歌词质量高
   - ⚠️ 需要模拟浏览器请求

3. **QQ 音乐** (备用源 2)
   - ✅ 流行歌曲资源丰富
   - ✅ 支持 Base64 编码歌词
   - ⚠️ 可能需要解码处理

## 核心特性

### 1. 智能降级策略
```
开始下载
    ↓
尝试 lrclib.net
    ├─ 成功 → 保存并返回 ✓
    └─ 失败 ↓
尝试 网易云音乐
    ├─ 成功 → 保存并返回 ✓
    └─ 失败 ↓
尝试 QQ 音乐
    ├─ 成功 → 保存并返回 ✓
    └─ 失败 → 报告错误 ❌
```

### 2. 批量下载支持
- 一键为整个音乐库的所有歌曲下载歌词
- 自动跳过已存在歌词文件的歌曲
- 实时显示下载进度和统计结果

### 3. 安全备份机制
- 下载新歌词前自动备份旧歌词文件(`.lrc.bak`)
- 下载失败不影响原有歌词文件
- 支持手动恢复备份

### 4. API 友好设计
- 每次请求间隔 500ms,避免触发速率限制
- 设置合理的超时时间(10秒)
- 友好的 User-Agent 标识

## 使用方法

### 通过托盘菜单

1. 右键点击系统托盘图标
2. 选择 **"整理音乐"** → **"从网络下载歌词"**
3. 等待后台下载完成
4. 查看通知消息了解下载结果

### 下载流程

```
开始下载
    ↓
检查当前音乐库
    ↓
遍历所有音乐文件
    ↓
检查是否已有歌词文件
    ├─ 有 → 跳过
    └─ 无 → 获取元数据
              ↓
         多源降级下载
              ↓
         保存 .lrc 文件
              ↓
         清除缓存
    ↓
显示统计结果
```

## 技术实现

### API 端点

#### 1. lrclib.net (主要)
```
GET https://lrclib.net/api/get?track_name={title}&artist_name={artist}&album_name={album}
```

#### 2. 网易云音乐 (备用)
```
# 搜索歌曲
GET https://music.163.com/api/search/get/web?s={keyword}&type=1&limit=5

# 获取歌词
GET https://music.163.com/api/song/lyric?id={songID}&lv=1&kv=1&tv=-1
```

#### 3. QQ 音乐 (备用)
```
# 搜索歌曲
GET https://c.y.qq.com/soso/fcgi-bin/client_search_cp?w={keyword}&format=json

# 获取歌词
GET https://c.y.qq.com/lyric/fcgi-bin/fcg_query_lyric_new.fcg?songmid={songMid}
```

### 响应结构

#### lrclib.net
```json
{
  "id": 12345,
  "trackName": "Song Title",
  "artistName": "Artist Name",
  "albumName": "Album Name",
  "duration": 240.5,
  "plainLyrics": "纯文本歌词内容...",
  "syncedLyrics": "[00:00.00]同步歌词内容..."
}
```

#### 网易云音乐
```json
{
  "result": {
    "songs": [
      {
        "id": 123456789,
        "name": "歌曲名",
        "artists": [{"name": "艺术家"}]
      }
    ]
  },
  "lrc": {
    "lyric": "[00:00.00]歌词内容..."
  }
}
```

### 代码架构

```
tray_menu.go (UI层)
    ↓ handleDownloadLyrics()
MusicService (服务层)
    ↓ DownloadLyricsForLibrary()
OrganizeService (业务层)
    ↓ DownloadLyricsForLibrary()
LyricManager (核心层)
    ├─ DownloadLyricWithFallback() - 多源降级下载
    ├─ downloadFromLRCLib() - lrclib.net
    ├─ downloadFromNetease() - 网易云音乐
    └─ downloadFromQQMusic() - QQ 音乐
```

### 关键函数

```go
// 多源降级下载(主入口)
func (lm *LyricManager) DownloadLyricWithFallback(
    trackPath, title, artist, album string
) error

// 批量下载
func (lm *LyricManager) DownloadLyricsForLibrary(
    libraryPath string, 
    metadataManager *MetadataManager
) (successCount, failCount, skipCount int, errors []string)
```

## 注意事项

### 依赖条件
- ✅ 需要网络连接
- ✅ 需要歌曲元数据(标题、艺术家)
- ⚠️ lrclib.net API 可用性

### 性能考虑
- 批量下载速度取决于歌曲数量和网速
- 每首歌约需 1-3 秒(含多源尝试和延迟保护)
- 100 首歌曲约需 3-5 分钟

### 常见问题

**Q: 为什么有些歌曲下载失败?**
A: 可能原因:
- 所有 API 源都没有该歌曲
- 元数据不准确或不完整
- 网络连接问题
- API 临时不可用

**Q: 多源下载会影响速度吗?**
A: 会有一定影响,但提高了成功率:
- 第一个源成功: ~1 秒
- 需要尝试第二个源: ~2 秒
- 需要尝试所有源: ~3 秒

**Q: 下载的歌词保存在哪里?**
A: 与音乐文件同一目录下,文件名为 `{歌曲名}.lrc`

**Q: 可以自定义下载路径吗?**
A: 当前版本固定在音乐文件同目录,未来版本可能支持自定义

**Q: 如何恢复旧歌词?**
A: 备份文件为 `{歌曲名}.lrc.bak`,重命名即可恢复

## 国际化支持

已添加中英文翻译:

**中文**:
- 菜单项: "从网络下载歌词"
- 提示: "正在从 lrclib.net 下载歌词..."
- 结果: "歌词下载完成：成功 X 首，失败 Y 首，跳过 Z 首"

**English**:
- Menu: "Download Lyrics from Internet"
- Progress: "Downloading lyrics from lrclib.net..."
- Result: "Lyrics download completed: X succeeded, Y failed, Z skipped"

## API 源对比

| 特性 | lrclib.net | 网易云音乐 | QQ 音乐 |
|------|-----------|----------|---------|
| **覆盖率** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
| **速度** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ |
| **稳定性** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ |
| **中文支持** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **同步歌词** | ✅ | ✅ | ✅ |
| **无需认证** | ✅ | ⚠️ | ⚠️ |
| **推荐场景** | 国际歌曲 | 中文流行 | 华语音乐 |

## 未来改进方向

- [ ] 支持更多 API 源(Kugou, Migu 等)
- [ ] 支持手动指定搜索关键词
- [ ] 歌词预览和编辑功能
- [ ] 下载进度条显示
- [ ] 支持中断和恢复下载
- [ ] 歌词质量评分和筛选
- [ ] 用户可配置 API 优先级
- [ ] 缓存搜索结果减少重复请求

## 相关文档

- [lrclib.net 官方文档](https://lrclib.net/docs)
- [LRC 歌词格式规范](https://wiki.xmplay.com/index.php?title=LRC_File_Format)
- [项目歌词管理架构](./LIBRARY_METADATA_SCAN.md)
- [歌词下载快速参考](./LYRICS_DOWNLOAD_QUICKREF.md)
