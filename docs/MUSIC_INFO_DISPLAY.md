# 音乐信息展示改进

## 改进内容

本次更新优化了音乐播放器的信息展示功能，从原来仅显示文件名升级为显示完整的音乐元数据。

### 主要变更

#### 1. 后端改进 (`backend/musicsmanager.go`)

- **添加 `createTrackInfo` 辅助函数**: 从文件路径创建包含完整信息的 `TrackInfo` 对象
- **更新事件发送逻辑**: 在 `PlayIndex()`, `Next()`, `Previous()` 方法中，发送完整的 `TrackInfo` 对象而非仅文件名
- **统一数据结构**: 使用 `libraryservice.go` 中定义的 `TrackInfo` 结构体

```go
type TrackInfo struct {
    Path     string `json:"path"`
    Filename string `json:"filename"`
    Title    string `json:"title"`
    Artist   string `json:"artist"`
    Album    string `json:"album"`
    Duration int64  `json:"duration"`
    Size     int64  `json:"size"`
}
```

#### 2. 前端改进 (`frontend/src/components/AppMain.vue`)

- **定义 TypeScript 接口**: 添加 `TrackInfo` 接口定义，与后端保持一致
- **更新状态管理**: `currentTrack` 类型从 `string` 改为 `TrackInfo | null`
- **增强显示逻辑**:
  - `displayTrackTitle()`: 优先显示 Title 元数据，其次显示 Filename
  - `displayArtist()`: 显示艺术家信息，默认为"未知艺术家"
  - `displayAlbum()`: 显示专辑信息，默认为"未知专辑"
- **兼容旧版本**: 事件监听器同时支持字符串和 TrackInfo 对象两种格式

#### 3. UI 布局优化

- **三行信息显示**:
  - 第一行：歌曲标题 (加粗，14px)
  - 第二行：艺术家 (半透明，12px)
  - 第三行：专辑 (更淡，11px)
- **文本溢出处理**: 所有文本行都应用了省略号截断，防止布局溢出

### 技术细节

#### 向后兼容性

前端事件监听器设计为同时兼容两种数据格式:

```typescript
Events.On("currentTrackChanged", (track: any) => {
  if (typeof track === "string") {
    // 兼容旧版本：转换为 TrackInfo 对象
    currentTrack.value = { /* ... */ };
  } else if (track && typeof track === "object") {
    // 新版本：直接使用 TrackInfo 对象
    currentTrack.value = track as TrackInfo;
  }
});
```

#### 未来扩展

当前实现中 `Artist` 和 `Album` 字段为空字符串，后续可以通过以下方式增强:

1. **ID3 标签读取**: 集成 `github.com/dhowden/tag` 库读取音频文件的元数据
2. **文件名解析**: 智能解析文件名格式 (如 "Artist - Title.mp3")
3. **在线元数据**: 调用 MusicBrainz 等 API 获取在线音乐数据库信息

### 测试验证

应用已成功编译并运行:
- ✅ 后端编译无错误
- ✅ 前端 TypeScript 类型检查通过
- ✅ Wails bindings 生成成功
- ✅ 应用正常启动并加载音乐库 (4280 首歌曲)
- ✅ 系统托盘和菜单功能正常

### 相关文件

- `backend/musicsmanager.go` - 播放列表管理和事件发送
- `backend/libraryservice.go` - TrackInfo 结构体定义
- `frontend/src/components/AppMain.vue` - 音乐信息展示组件
