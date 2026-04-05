# 喜爱音乐功能实现说明

## 概述
新增了"喜爱音乐"功能，用户可以查看按播放次数排序的歌曲列表，快速访问最常听的音乐。

## 实现内容

### 1. 前端组件
**文件**: `frontend/src/views/FavoritesView.vue`

**主要功能**:
- 调用后端 API `GetFavoriteTracks(100)` 获取前 100 首喜爱音乐
- 表格化展示歌曲信息（排名、歌名、艺术家、专辑、播放次数、时长、文件大小）
- 点击歌曲行直接添加到播放列表并播放
- 支持刷新列表和返回主界面
- 响应式设计，适配不同窗口大小
- 美观的渐变背景和毛玻璃效果

**UI 特点**:
- 紫色渐变背景 (`#667eea` → `#764ba2`)
- 播放次数使用金色徽章高亮显示
- 悬停效果和平滑过渡动画
- 空状态友好提示
- Loading 加载状态指示器

### 2. 路由配置
**文件**: `frontend/src/App.vue`

**修改内容**:
- 导入 `FavoritesView` 组件
- 在 `checkRoute()` 函数中添加 `#/favorites` 和 `/favorites` 路由判断
- 更新动态组件渲染逻辑，支持 favorites 视图

**访问方式**:
- Hash 模式: `window.location.hash = '#/favorites'`
- Path 模式: `window.location.pathname = '/favorites'`

### 3. 托盘菜单集成
**文件**: `main.go`

**修改内容**:
- 更新 `favoriteItem` 菜单项标签为 "❤️ 喜爱音乐"
- 添加快捷键 `CmdOrCtrl+H` (Heart)
- 点击事件实现：
  1. 显示主窗口（如果未显示）
  2. 发送导航事件到 `#/favorites`
  3. 自动聚焦窗口

### 4. 后端 API
**已有接口**: `GetFavoriteTracks(limit int) []HistoryRecord`

**数据来源**: 
- 从 `HistoryManager` 获取历史记录
- 按 `PlayCount` 字段降序排序
- 返回指定数量的记录

## 技术细节

### 数据流
```
用户点击托盘菜单 "❤️ 喜爱音乐"
    ↓
mainWindow.Show() + Focus()
    ↓
Event.Emit("windowUrl", { url: "#/favorites" })
    ↓
App.vue checkRoute() 检测到 #/favorites
    ↓
currentView.value = "favorites"
    ↓
FavoritesView.vue onMounted()
    ↓
GetFavoriteTracks(100) 调用后端 API
    ↓
渲染喜爱音乐列表
```

### 关键代码

#### FavoritesView.vue - 数据加载
```typescript
const loadFavorites = async () => {
  isLoading.value = true;
  error.value = "";
  try {
    favorites.value = await GetFavoriteTracks(100);
  } catch (err) {
    error.value = "加载失败，请稍后重试";
  } finally {
    isLoading.value = false;
  }
};
```

#### FavoritesView.vue - 播放歌曲
```typescript
const playTrack = async (track: HistoryRecord) => {
  try {
    await ClearPlaylist();
    await AddToPlaylist(track.path);
    await PlayIndex(0);
  } catch (err) {
    error.value = "播放失败，请重试";
  }
};
```

#### main.go - 托盘菜单点击
```go
favoriteItem = application.NewMenuItem("❤️ 喜爱音乐")
favoriteItem.SetAccelerator("CmdOrCtrl+H")
favoriteItem.OnClick(func(ctx *application.Context) {
    if mainWindow != nil && !mainWindow.IsVisible() {
        mainWindow.Show()
        mainWindow.Focus()
    }
    
    go func() {
        time.Sleep(100 * time.Millisecond)
        if mainWindow != nil && mainWindow.IsVisible() {
            app.Event.Emit("windowUrl", map[string]interface{}{
                "type": "navigate",
                "url":  "#/favorites",
            })
        }
    }()
})
```

## 样式设计

### 配色方案
- **背景**: 紫色渐变 (`linear-gradient(135deg, #667eea 0%, #764ba2 100%)`)
- **播放次数徽章**: 金色半透明 (`rgba(255, 215, 0, 0.3)`)
- **文字**: 白色系，不同透明度区分层级
- **表格**: 半透明白色背景 + 毛玻璃效果

### 交互反馈
- 表格行悬停：背景变亮 + 轻微放大
- 按钮悬停：上移 2px + 背景变亮
- 加载状态：旋转 spinner 动画
- 滚动条：自定义样式，与主题协调

## 使用场景

1. **快速访问常听歌曲**: 通过播放次数排序，一目了然看到最爱听的歌
2. **发现遗忘的好歌**: 回顾历史播放，重新发现曾经喜欢的音乐
3. **统计听歌习惯**: 通过播放次数了解自己的音乐偏好

## 注意事项

1. **数据依赖**: 喜爱音乐基于播放历史，需要先有播放记录才会显示
2. **数量限制**: 最多显示 100 首，避免列表过长影响性能
3. **实时更新**: 每次进入页面都会重新加载最新数据
4. **播放次数**: 同一首歌重复播放会累加次数，而非新增记录

## 未来优化方向

1. **导出功能**: 支持导出喜爱音乐列表为 M3U 或 TXT 格式
2. **筛选功能**: 按艺术家、专辑、时间范围筛选
3. **收藏功能**: 允许用户手动标记喜爱歌曲（独立于播放次数）
4. **统计图表**: 可视化展示听歌习惯和趋势
5. **分享功能**: 生成喜爱音乐清单图片分享到社交媒体

## 相关文件

- `frontend/src/views/FavoritesView.vue` - 喜爱音乐视图组件
- `frontend/src/App.vue` - 路由配置
- `main.go` - 托盘菜单集成
- `backend/historymanager.go` - 历史记录管理器（提供 GetFavoriteTracks）
- `backend/music_service.go` - 音乐服务层（暴露 API）