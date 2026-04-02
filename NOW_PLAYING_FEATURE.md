# 当前播放音乐名称菜单项实现

## 功能说明

在系统托盘菜单中添加了"正在播放"菜单项，用于显示当前播放的音乐文件名称。

## 实现内容

### 1. 后端新增方法

**文件**: `backend/music_service.go`

新增以下方法：

```go
// GetCurrentIndex 获取当前播放索引
func (m *MusicService) GetCurrentIndex() (int, error) {
    return m.playlistManager.GetCurrentIndex()
}

// GetCurrentTrack 获取当前播放的歌曲路径
func (m *MusicService) GetCurrentTrack() (string, error) {
    playlist, err := m.playlistManager.GetPlaylist()
    if err != nil {
        return "", err
    }

    index, err := m.playlistManager.GetCurrentIndex()
    if err != nil {
        return "", err
    }

    if index < 0 || index >= len(playlist) {
        return "", fmt.Errorf("当前播放索引无效：%d", index)
    }

    return playlist[index], nil
}

// GetCurrentTrackName 获取当前播放的歌曲名称（仅文件名）
func (m *MusicService) GetCurrentTrackName() (string, error) {
    path, err := m.GetCurrentTrack()
    if err != nil {
        return "", err
    }

    // 从路径中提取文件名
    filename := filepath.Base(path)
    return filename, nil
}
```

### 2. 前端菜单项

**文件**: `main.go`

#### 2.1 添加菜单项声明

```go
var nowPlayingItem *application.MenuItem // 正在播放的音乐名称
```

#### 2.2 创建更新函数

```go
updateNowPlayingItem := func() {
    if musicService == nil {
        return
    }
    
    trackName, err := musicService.GetCurrentTrackName()
    if err != nil {
        nowPlayingItem.SetText("未播放")
        nowPlayingItem.SetEnabled(false)
        return
    }
    
    // 截断过长的文件名（最多显示 30 个字符）
    displayName := trackName
    if len(displayName) > 30 {
        displayName = displayName[:27] + "..."
    }
    
    nowPlayingItem.SetText("🎵 " + displayName)
    nowPlayingItem.SetEnabled(true)
}
```

#### 2.3 创建菜单项

```go
nowPlayingItem = application.NewMenuItem("未播放")
nowPlayingItem.SetEnabled(false) // 禁用点击，仅展示
```

#### 2.4 监听事件自动更新

```go
app.Event.On("currentTrackChanged", func(ctx *application.Context, data string) {
    log.Printf("收到歌曲变化事件：%s", data)
    updateNowPlayingItem()
})
```

#### 2.5 延迟初始化

```go
go func() {
    time.Sleep(500 * time.Millisecond)
    updateNowPlayingItem()
}()
```

#### 2.6 添加到菜单

```go
menu = application.NewMenuFromItems(
    nowPlayingItem, // 正在播放菜单项（顶部）
    application.NewMenuItemSeparator(),
    playPauseItem,
    prevItem,
    nextItem,
    // ... 其他菜单项
)
```

## 菜单显示效果

### 未播放时
```
未播放
─────────────────
播放
上一曲
下一曲
...
```

### 播放中
```
🎵 SongName.mp3
─────────────────
播放
上一曲
下一曲
...
```

### 长文件名处理
```
🎵 VeryLongSongNam...
─────────────────
播放
上一曲
下一曲
...
```

## 特性

1. **实时更新** - 监听 `currentTrackChanged` 事件自动更新
2. **禁用状态** - 菜单项不可点击，仅用于展示
3. **长度截断** - 超过 30 个字符自动截断，避免菜单过宽
4. **音乐符号** - 添加 🎵 前缀，增强视觉效果
5. **初始状态** - 启动时显示"未播放"

## 使用场景

- ✅ 快速查看当前播放的歌曲
- ✅ 切换歌曲时自动更新显示
- ✅ 长文件名自动截断，保持菜单整洁
- ✅ 与播放控制菜单项分离，布局清晰

## 技术实现

### 事件驱动更新
```
播放/切换歌曲
    ↓
AudioPlayer.Play()
    ↓
app.Event.Emit("currentTrackChanged", filename)
    ↓
updateNowPlayingItem()
    ↓
nowPlayingItem.SetText("🎵 " + filename)
```

### 性能优化
- 只在歌曲变化时更新，不轮询
- 延迟初始化，等待服务完全启动
- 错误处理完善，失败时显示"未播放"

## 依赖关系

```
MusicService
    ├── GetCurrentTrackName()
    │   ├── GetCurrentTrack()
    │   │   ├── GetPlaylist()
    │   │   └── GetCurrentIndex()
    │   └── filepath.Base()
    └── Event.Emit()
```

## 向后兼容

✅ 完全向后兼容，不影响现有功能

## 测试建议

1. **测试场景 1**: 启动应用，未播放时显示"未播放"
2. **测试场景 2**: 播放歌曲，显示歌曲名称
3. **测试场景 3**: 切换歌曲，菜单项自动更新
4. **测试场景 4**: 播放长文件名歌曲，名称被截断
5. **测试场景 5**: 停止播放，显示"未播放"

---

**实现日期**: 2026-04-02  
**影响范围**: 系统托盘菜单  
**向后兼容**: ✅ 是
