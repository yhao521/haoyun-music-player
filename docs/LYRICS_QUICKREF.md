# 歌词功能快速参考

## 🚀 快速开始

### 查看实时歌词

```bash
# 1. 运行应用
./haoyun-music-player

# 2. 播放歌曲
# 3. 右键托盘图标 → 查看"正在播放"下方的滚动歌词
```

### 调整歌词同步

```typescript
// TypeScript/JavaScript 示例
const trackPath = "/path/to/song.mp3";

// 延后 2.5 秒 (适用于前奏过长)
await musicService.SetLyricOffset(trackPath, 2.5);

// 提前 1.0 秒 (适用于 LRC 时间戳偏晚)
await musicService.SetLyricOffset(trackPath, -1.0);

// 获取当前偏移量
const offset = await musicService.GetLyricOffset(trackPath);
console.log("当前偏移:", offset);
```

---

## 🔧 常见问题

### Q1: 歌词不显示?

**检查清单**:
- [ ] 确认 `.lrc` 文件存在
- [ ] 查看日志是否有 "✓ 加载歌词" 输出
- [ ] 确认歌词文件格式正确 (UTF-8 编码)

**调试命令**:
```bash
./haoyun-music-player 2>&1 | grep "🎤\|🔍"
```

### Q2: 歌词与播放不同步?

**解决方案**:
```typescript
// 逐步调整偏移量,直到对齐
await musicService.SetLyricOffset(trackPath, 1.0);  // 尝试 +1s
await musicService.SetLyricOffset(trackPath, 1.5);  // 尝试 +1.5s
await musicService.SetLyricOffset(trackPath, 1.3);  // 微调至 +1.3s
```

### Q3: 分离目录结构歌词无法加载?

**确保目录命名规范**:
```
Music/
├── LIB_MUSIC/      ← 音频文件
│   └── song.mp3
└── LIB_LYRIC/      ← 歌词文件 (必须命名为 LIB_LYRIC)
    └── song.lrc
```

---

## 📊 API 速查

### MusicService 方法

| 方法 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| `LoadLyric(trackPath)` | `string` | `*LyricInfo, error` | 加载歌词到缓存 |
| `GetCurrentLyricLine(trackPath, position)` | `string, float64` | `int, error` | 获取当前歌词行索引 |
| `GetAllLyrics(trackPath)` | `string` | `[]LyricLine, error` | 获取所有歌词 |
| `HasLyric(trackPath)` | `string` | `bool` | 检查是否有歌词 |
| `SetLyricOffset(trackPath, offset)` | `string, float64` | `void` | 设置偏移量(秒) |
| `GetLyricOffset(trackPath)` | `string` | `float64` | 获取偏移量 |
| `GetPosition()` | - | `float64, error` | 获取播放位置(秒) |
| `GetDuration()` | - | `float64, error` | 获取总时长(秒) |

---

## 🎯 偏移量规则

| 值 | 效果 | 适用场景 |
|----|------|----------|
| `+2.5` | 歌词延后 2.5 秒 | 前奏过长 |
| `-1.0` | 歌词提前 1.0 秒 | LRC 时间戳偏晚 |
| `0.0` | 无偏移(默认) | 标准对齐 |

**公式**: `最终时间 = 原始时间 + LRC_Offset + Custom_Offset`

---

## 🐛 调试技巧

### 查看详细日志

```bash
# 过滤歌词相关日志
./haoyun-music-player 2>&1 | grep "🎤"

# 查看歌词查找过程
./haoyun-music-player 2>&1 | grep "🔍"
```

### 预期日志输出

```
🔍 查找歌词 - 歌曲: /path/to/song.mp3, 基础名: song
✓ 策略3成功(分离目录): /path/to/LIB_LYRIC/song.lrc
✓ 加载歌词:120 行 (偏移量: 2.50秒)
🎤 检查歌词 - 路径: /path/to/song.mp3, 有歌词: true
🎤 更新歌词显示: 🎤 算命的说我...
```

---

## 📁 文件位置

### 配置文件

- **歌词缓存**: `~/.haoyun-music/lib/lyrics/`
- **音乐库索引**: `~/.haoyun-music/libraries/Music.json`

### 源代码

- **核心逻辑**: `backend/lyricmanager.go`
- **接口暴露**: `backend/music_service.go`
- **托盘菜单**: `tray_menu.go`
- **国际化**: `backend/pkg/i18n/{zh-CN,en-US}.json`

---

## ⚡ 性能提示

1. **更新频率**: 500ms (平衡流畅度和 CPU 占用)
2. **智能去重**: 仅在歌词变化时更新 UI
3. **状态感知**: 暂停时自动停止定时器
4. **缓存复用**: 重复调用零开销

---

## 🔗 相关文档

- **完整实现**: `docs/LYRICS_ENHANCEMENT_IMPLEMENTATION.md`
- **歌词下载**: `docs/LYRICS_DOWNLOAD_FEATURE.md`
- **多源增强**: `docs/MULTI_SOURCE_LYRICS_ENHANCEMENT.md`

---

**最后更新**: 2026-04-12
