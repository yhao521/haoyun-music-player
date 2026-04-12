# 歌词功能增强实现文档

## 概述

本文档记录了音乐播放器歌词功能的完整增强实现,包括系统托盘实时歌词显示、分离目录结构支持、以及动态歌词同步调整功能。

**实现日期**: 2026-04-12  
**版本**: v1.0  
**涉及模块**: 后端 (Go)、国际化 (i18n)

---

## 问题描述

### 问题 1: 缺少实时歌词显示

用户希望在系统托盘菜单中实时查看当前播放歌曲的歌词,类似 macOS 上的 LyricsX 或 DynamicLyrics 应用体验。

### 问题 2: 分离目录结构歌词加载失败

用户使用整理音乐功能后,音频文件和歌词文件分别存储在不同目录:
- 音频文件: `LIB_MUSIC/`
- 歌词文件: `LIB_LYRIC/`

但歌词无法被正确加载,导致前端界面和托盘菜单均无法显示歌词。

### 问题 3: 歌词与播放不同步

当音乐文件有前奏(intro)或 LRC 歌词时间戳不准确时,歌词显示与实际播放位置不同步,影响用户体验。

---

## 解决方案

### 方案架构

采用三层解决方案:

1. **系统托盘实时歌词显示** - 在菜单栏中实时滚动显示当前歌词
2. **智能歌词文件查找** - 支持同目录、全局目录、分离目录三种查找策略
3. **动态偏移量调整** - 运行时实时调整歌词同步,无需修改原始 LRC 文件

---

## 实现细节

### 一、系统托盘实时歌词显示

#### 1.1 后端接口扩展

**文件**: `backend/music_service.go`

新增方法暴露音频播放器的位置和时长信息:

```go
// GetPosition 获取当前播放位置(秒)
func (m *MusicService) GetPosition() (float64, error) {
    return m.audioPlayer.GetPosition()
}

// GetDuration 获取歌曲总时长(秒)
func (m *MusicService) GetDuration() (float64, error) {
    return m.audioPlayer.GetDuration()
}
```

#### 1.2 托盘菜单增强

**文件**: `tray_menu.go`

**核心变量**:
```go
var lyricDisplayItem *application.MenuItem  // 歌词显示菜单项
var lyricUpdateTicker *time.Ticker          // 定时更新器
var lyricUpdateStop chan struct{}           // 停止信号通道
```

**关键函数**:

1. **启动歌词定时更新** (500ms 间隔):
```go
func startLyricUpdateTicker() {
    if lyricUpdateTicker != nil {
        return
    }
    
    lyricUpdateStop = make(chan struct{})
    lyricUpdateTicker = time.NewTicker(500 * time.Millisecond)
    
    go func() {
        for {
            select {
            case <-lyricUpdateTicker.C:
                updateLyricDisplay()
            case <-lyricUpdateStop:
                return
            }
        }
    }()
    
    log.Println("🎤 启动歌词定时更新")
}
```

2. **智能更新歌词显示** (带去重优化):
```go
func updateLyricDisplay() {
    if musicService == nil || lyricDisplayItem == nil {
        return
    }

    // 获取当前歌曲路径
    trackPath, err := musicService.GetCurrentTrack()
    if err != nil || trackPath == "" {
        setLyricLabel(t("lyric.noLyric"))
        return
    }

    // 获取当前播放位置
    position, err := musicService.GetPosition()
    if err != nil {
        log.Printf("⚠️ 获取播放位置失败: %v", err)
        return
    }

    // 先尝试加载歌词(触发 findLyricFile 并缓存结果)
    _, loadErr := musicService.LoadLyric(trackPath)
    if loadErr != nil {
        log.Printf("⚠️ 加载歌词失败: %v", loadErr)
    }

    // 检查是否有歌词
    hasLyric := musicService.HasLyric(trackPath)
    
    // 获取当前歌词行索引
    lineIndex, err := musicService.GetCurrentLyricLine(trackPath, position)
    if err != nil || lineIndex < 0 {
        if hasLyric {
            setLyricLabel(t("lyric.instrumental"))
        } else {
            setLyricLabel(t("lyric.noLyric"))
        }
        return
    }

    // 获取所有歌词
    lyrics, err := musicService.GetAllLyrics(trackPath)
    if err != nil || lineIndex >= len(lyrics) {
        setLyricLabel(t("lyric.loading"))
        return
    }

    // 获取当前歌词文本并截断(最多25字符)
    lyricText := lyrics[lineIndex].Content
    runes := []rune(lyricText)
    if len(runes) > 25 {
        lyricText = string(runes[:22]) + "..."
    }

    // 智能更新:只有当歌词变化时才更新UI
    newLabel := "🎤 " + lyricText
    setLyricLabel(newLabel)
}
```

3. **避免重复渲染**:
```go
var lastLyricLabel string

func setLyricLabel(newLabel string) {
    if lyricDisplayItem == nil {
        return
    }
    
    // 仅在标签变化时才更新
    if newLabel != lastLyricLabel {
        lyricDisplayItem.SetLabel(newLabel)
        lastLyricLabel = newLabel
    }
}
```

#### 1.3 事件驱动集成

**监听歌曲切换**:
```go
app.Event.On("currentTrackChanged", func(event *application.CustomEvent) {
    updateNowPlayingItem()
    updateLyricDisplay()  // 立即更新歌词
    startLyricUpdateTicker()  // 启动定时更新
})
```

**监听播放状态**:
```go
app.Event.On("playbackStateChanged", func(event *application.CustomEvent) {
    state, ok := event.Data.(string)
    if !ok {
        return
    }
    
    if state == "playing" {
        startLyricUpdateTicker()
    } else if state == "paused" || state == "stopped" {
        stopLyricUpdateTicker()
    }
})
```

#### 1.4 UI 效果

```
🎵 夜空中最亮的星
🎤 能否听清那仰望的人...  ← 实时滚动歌词
───────────────
▶️ 暂停
⏮️ 上一曲
⏭️ 下一曲
───────────────
📂 浏览歌曲
❤️ 喜爱音乐
...
```

---

### 二、分离目录结构歌词加载修复

#### 2.1 问题分析

原 `LyricManager.findLyricFile()` 仅支持两种查找策略:
1. 同目录下查找 (`.lrc` 与 `.mp3` 在同一文件夹)
2. 全局歌词目录 (`~/.haoyun-music/lib/lyrics/`)

**未考虑**音乐库内部的分离目录结构 (`LIB_MUSIC` / `LIB_LYRIC`)。

#### 2.2 解决方案

**文件**: `backend/lyricmanager.go`

新增**策略3: 智能路径替换**:

```go
// findLyricFile 查找歌词文件
func (lm *LyricManager) findLyricFile(trackPath string) string {
    baseName := strings.TrimSuffix(filepath.Base(trackPath), filepath.Ext(trackPath))
    dirPath := filepath.Dir(trackPath)

    log.Printf("🔍 查找歌词 - 歌曲: %s, 基础名: %s", trackPath, baseName)

    // 策略 1: 同目录下的 .lrc 文件
    lrcPath1 := filepath.Join(dirPath, baseName+".lrc")
    if _, err := os.Stat(lrcPath1); err == nil {
        log.Printf("✓ 策略1成功(同目录): %s", lrcPath1)
        return lrcPath1
    }

    // 策略 2: 歌词目录下的 .lrc 文件(全局歌词目录)
    lrcPath2 := filepath.Join(lm.lyricDir, baseName+".lrc")
    if _, err := os.Stat(lrcPath2); err == nil {
        log.Printf("✓ 策略2成功(全局目录): %s", lrcPath2)
        return lrcPath2
    }

    // 策略 3: 检测音乐库分离结构(LIB_MUSIC / LIB_LYRIC)
    if strings.Contains(dirPath, "LIB_MUSIC") {
        lyricDir := strings.Replace(dirPath, "LIB_MUSIC", "LIB_LYRIC", 1)
        lrcPath3 := filepath.Join(lyricDir, baseName+".lrc")
        if _, err := os.Stat(lrcPath3); err == nil {
            log.Printf("✓ 策略3成功(分离目录): %s", lrcPath3)
            return lrcPath3
        }
        log.Printf("⚠️ 策略3失败(分离目录): %s", lrcPath3)
    }

    log.Printf("⚠️ 未找到歌词文件")
    return ""
}
```

#### 2.3 查找优先级

| 优先级 | 策略 | 示例路径 | 适用场景 |
|--------|------|----------|----------|
| 1️⃣ | **同目录** | `LIB_MUSIC/song.lrc` | 传统结构 |
| 2️⃣ | **全局目录** | `~/.haoyun-music/lib/lyrics/song.lrc` | 集中管理 |
| 3️⃣ | **分离目录** ✨ | `LIB_LYRIC/song.lrc` | 整理后的结构 |

---

### 三、歌词缓存加载时序修复

#### 3.1 问题现象

日志显示:
```
🎤 检查歌词 - 路径: xxx.mp3, 有歌词: false
🎤 未找到当前歌词行 - 错误: 没有可用的歌词, 索引: -1
```

#### 3.2 根本原因

1. **缓存依赖**: `HasLyric()` 仅检查内存缓存,如果歌词未被预先加载,直接返回 `false`
2. **调用顺序错误**: `updateLyricDisplay()` 先调用 `HasLyric()`,但此时歌词尚未通过 `LoadLyric()` 加载到缓存
3. **懒加载机制**: `GetCurrentLyricLine()` 内部会调用 `LoadLyric()`,但如果先调用 `HasLyric()` 会得到错误结果

#### 3.3 解决方案

在检查歌词状态前,**先主动调用 `LoadLyric()` 触发歌词加载和缓存**:

```go
// ✅ 正确顺序: 先加载,再检查
_, loadErr := musicService.LoadLyric(trackPath)  // 触发 findLyricFile 并缓存
if loadErr != nil {
    log.Printf("⚠️ 加载歌词失败: %v", loadErr)
}

hasLyric := musicService.HasLyric(trackPath)  // 现在能正确读取缓存
log.Printf("🎤 检查歌词 - 路径: %s, 有歌词: %v", trackPath, hasLyric)
```

---

### 四、动态歌词同步调整系统

#### 4.1 问题背景

当音乐文件有前奏或 LRC 歌词时间戳不准确时,歌词显示与实际播放不同步。

#### 4.2 数据结构扩展

**文件**: `backend/lyricmanager.go`

```go
type LyricManager struct {
    mu             sync.RWMutex
    cache          map[string]*LyricInfo
    searchCache    map[string]string
    lyricDir       string
    customOffsets  map[string]float64  // 新增: 用户自定义偏移量
}
```

#### 4.3 核心 API

**设置自定义偏移量**:
```go
func (lm *LyricManager) SetCustomOffset(trackPath string, offset float64) {
    lm.mu.Lock()
    defer lm.mu.Unlock()
    
    lm.customOffsets[trackPath] = offset
    log.Printf("🎵 设置歌词偏移量: %s -> %.2f秒", trackPath, offset)
    
    // 如果歌词已缓存,重新应用偏移量
    if cachedLyric, ok := lm.cache[trackPath]; ok && cachedLyric.HasLyric {
        lm.reapplyOffset(trackPath, cachedLyric)
    }
}
```

**获取自定义偏移量**:
```go
func (lm *LyricManager) GetCustomOffset(trackPath string) float64 {
    lm.mu.RLock()
    defer lm.mu.RUnlock()
    
    if offset, ok := lm.customOffsets[trackPath]; ok {
        return offset
    }
    return 0.0
}
```

**双重偏移叠加逻辑**:
```go
// parseLRCFileWithCustomOffset 解析 LRC 文件并应用自定义偏移量
func (lm *LyricManager) parseLRCFileWithCustomOffset(filePath string, customOffset float64) (*LyricInfo, error) {
    // ... 解析逻辑 ...
    
    // 应用 LRC 文件中的偏移量 + 用户自定义偏移量
    timeSeconds += lyric.Offset + customOffset
    
    // ... 继续处理 ...
}
```

#### 4.4 MusicService 接口暴露

**文件**: `backend/music_service.go`

```go
// SetLyricOffset 设置歌词偏移量(用于同步调整)
func (m *MusicService) SetLyricOffset(trackPath string, offset float64) {
    m.lyricManager.SetCustomOffset(trackPath, offset)
}

// GetLyricOffset 获取歌词偏移量
func (m *MusicService) GetLyricOffset(trackPath string) float64 {
    return m.lyricManager.GetCustomOffset(trackPath)
}
```

#### 4.5 偏移量规则

- **正值 (+)** → 歌词**延后**显示 (适用于前奏过长)
- **负值 (-)** → 歌词**提前**显示 (适用于 LRC 时间戳偏晚)
- **叠加公式**: `最终时间 = 原始时间 + LRC_Offset + Custom_Offset`

---

### 五、国际化支持

#### 5.1 中文翻译

**文件**: `backend/pkg/i18n/zh-CN.json`

```json
{
  "lyric": {
    "currentLyric": "当前歌词",
    "noLyric": "🎵 无歌词",
    "instrumental": "🎵 纯音乐",
    "loading": "🎤 加载中..."
  }
}
```

#### 5.2 英文翻译

**文件**: `backend/pkg/i18n/en-US.json`

```json
{
  "lyric": {
    "currentLyric": "Current Lyric",
    "noLyric": "🎵 No Lyrics",
    "instrumental": "🎵 Instrumental",
    "loading": "🎤 Loading..."
  }
}
```

---

## 使用方法

### 1. 查看实时歌词

1. 运行应用: `./haoyun-music-player`
2. 播放任意歌曲
3. 右键点击系统托盘图标
4. 在"正在播放"下方即可看到实时滚动的歌词

### 2. 调整歌词同步

通过前端调用 API (TypeScript 示例):

```typescript
const trackPath = "/path/to/song.mp3";

// 获取当前偏移量
const currentOffset = await musicService.GetLyricOffset(trackPath);

// 调整偏移量 (例如: 延后 2.5 秒)
await musicService.SetLyricOffset(trackPath, 2.5);

// 歌词会自动重新加载并应用新偏移量
```

### 3. 调试日志

启用详细日志输出:

```bash
./haoyun-music-player 2>&1 | grep "🎤\|🔍"
```

预期输出:
```
🔍 查找歌词 - 歌曲: /path/to/song.mp3
✓ 策略3成功(分离目录): /path/to/LIB_LYRIC/song.lrc
🎵 设置歌词偏移量: /path/to/song.mp3 -> 2.50秒
✓ 已重新应用偏移量并更新缓存
✓ 加载歌词:120 行 (偏移量: 2.50秒)
🎤 更新歌词显示: 🎤 算命的说我...
```

---

## 测试验证

### 测试用例 1: 系统托盘歌词显示

**步骤**:
1. 播放一首有 LRC 歌词的歌曲
2. 右键托盘图标
3. 观察"正在播放"下方的歌词是否实时滚动

**预期结果**: ✅ 歌词每 500ms 更新一次,与人声同步

### 测试用例 2: 分离目录结构加载

**步骤**:
1. 确保音频文件在 `LIB_MUSIC/`,歌词文件在 `LIB_LYRIC/`
2. 播放该歌曲
3. 查看日志输出

**预期结果**: ✅ 日志显示"策略3成功(分离目录)"

### 测试用例 3: 歌词同步调整

**步骤**:
1. 播放一首歌词不同步的歌曲
2. 调用 `SetLyricOffset(trackPath, 2.0)`
3. 观察歌词是否重新对齐

**预期结果**: ✅ 歌词立即重新加载并应用新偏移量

---

## 性能优化

### 1. 智能去重更新

仅在歌词文本变化时才调用 `SetLabel()`,避免频繁 UI 刷新:

```go
var lastLyricLabel string

func setLyricLabel(newLabel string) {
    if newLabel != lastLyricLabel {
        lyricDisplayItem.SetLabel(newLabel)
        lastLyricLabel = newLabel
    }
}
```

### 2. 状态感知控制

- **播放时**: 启动 500ms 定时更新
- **暂停/停止时**: 自动停止定时器,节省 CPU

### 3. 缓存复用

`LoadLyric()` 内部已有缓存检查,重复调用零开销:

```go
if lyric, ok := lm.cache[trackPath]; ok {
    return lyric, nil  // 直接返回缓存
}
```

---

## 注意事项

### 已知限制

1. **跨平台兼容性**:
   - **macOS**: ✅ 完美支持,用户体验最佳
   - **Windows**: ✅ 支持,任务栏右键菜单更新频率不宜过高
   - **Linux**: ⚠️ 依赖桌面环境(AppIndicator/GTK),部分 DE 可能不支持动态更新

2. **菜单栏空间限制**:
   - 单行显示长度建议控制在 20-30 字符
   - 超长歌词会自动截断并显示省略号

3. **偏移量持久化**:
   - 当前实现中,自定义偏移量仅存储在内存中
   - 应用重启后会重置为零
   - **未来优化**: 可扩展为保存到 JSON 配置文件

### 未来优化方向

1. **持久化存储**: 将 `customOffsets` 保存到用户配置目录
2. **前端 UI 控件**: 添加滑块让用户直观调整 (-5s ~ +5s)
3. **自动检测前奏**: 分析音频波形自动计算最佳 offset
4. **批量调整**: 同一专辑/艺术家的歌曲应用相同偏移量
5. **歌词子菜单**: 可选展开显示多行歌词上下文

---

## 相关文件

### 后端文件

- `backend/lyricmanager.go` - 歌词管理器核心逻辑
- `backend/music_service.go` - 音乐服务接口暴露
- `tray_menu.go` - 系统托盘菜单实现

### 国际化文件

- `backend/pkg/i18n/zh-CN.json` - 中文翻译
- `backend/pkg/i18n/en-US.json` - 英文翻译

### 相关文档

- `docs/LYRICS_DOWNLOAD_FEATURE.md` - 歌词下载功能
- `docs/MULTI_SOURCE_LYRICS_ENHANCEMENT.md` - 多源歌词增强
- `docs/LRCLIB_ENHANCEMENT_COMPLETE.md` - LRCLib 集成

---

## 技术要点总结

### 1. 事件驱动架构

利用 Wails v3 的事件系统实现组件解耦:
- `currentTrackChanged` → 触发歌词加载和定时器启动
- `playbackStateChanged` → 智能启停更新协程

### 2. 缓存一致性保障

- **预加载策略**: 在使用前先调用 `LoadLyric()` 确保数据就绪
- **自动重应用**: 修改偏移量时自动触发缓存更新
- **线程安全**: 使用 `sync.RWMutex` 保护并发访问

### 3. 智能查找算法

三层查找策略兼顾灵活性和性能:
1. 同目录 (最快,最常见)
2. 全局目录 (集中管理)
3. 分离目录 (适配整理后的结构)

### 4. 非破坏性修改

- 不修改原始 LRC 文件
- 偏移量存储在内存映射中
- 可随时调整,无副作用

---

## 版本历史

- **v1.0** (2026-04-12): 初始实现
  - ✅ 系统托盘实时歌词显示
  - ✅ 分离目录结构支持
  - ✅ 动态偏移量调整系统
  - ✅ 国际化支持
  - ✅ 性能优化(去重+状态感知)

---

**文档维护者**: 开发团队  
**最后更新**: 2026-04-12
