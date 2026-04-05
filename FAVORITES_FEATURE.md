# 喜爱音乐功能实现说明（独立窗口版）

## 概述
新增了"喜爱音乐"功能，使用**独立窗口**展示按播放次数排序的歌曲列表，用户可以快速访问最常听的音乐。

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

### 3. 独立窗口实现
**文件**: `main.go`

#### 窗口创建
```go
// 创建喜爱音乐窗口（用于展示按播放次数排序的歌曲列表）
var favoritesWindow *application.WebviewWindow
favoritesWindow = app.Window.NewWithOptions(application.WebviewWindowOptions{
    Title: "喜爱音乐 - Haoyun Music Player",
    Mac: application.MacWindow{
        InvisibleTitleBarHeight: 50,
        Backdrop:                application.MacBackdropTranslucent,
        TitleBar:                application.MacTitleBarHiddenInset,
    },
    BackgroundColour: application.NewRGB(27, 38, 54),
    URL:              "#/favorites",  // 直接导航到喜爱音乐页面
    Width:            900,
    Height:           700,
})

// 拦截喜爱音乐窗口关闭事件，改为隐藏窗口
favoritesWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
    log.Println("喜爱音乐窗口关闭事件被拦截，改为隐藏窗口")
    favoritesWindow.Hide()
    e.Cancel() // 取消关闭操作
})

// 初始隐藏喜爱音乐窗口
favoritesWindow.Hide()
log.Println("✓ Favorites window created (Hidden)")
```

#### 托盘菜单集成
```go
favoriteItem = application.NewMenuItem("❤️ 喜爱音乐")
favoriteItem.SetAccelerator("CmdOrCtrl+H") // Cmd/Ctrl + H (Heart)

// 在 favoritesWindow 创建后绑定点击事件
favoriteItem.OnClick(func(ctx *application.Context) {
    log.Println("打开喜爱音乐窗口")
    
    defer func() {
        if r := recover(); r != nil {
            log.Printf("❌ 打开喜爱音乐窗口时发生 panic: %v", r)
            debug.PrintStack()
        }
    }()

    if favoritesWindow == nil {
        log.Println("❌ favoritesWindow 为 nil")
        return
    }

    isVisible := favoritesWindow.IsVisible()
    log.Printf("✓ favoritesWindow IsVisible() = %v", isVisible)
    
    if isVisible {
        log.Println("准备调用 Hide()...")
        favoritesWindow.Hide()
    } else {
        log.Println("准备调用 Show()...")
        favoritesWindow.Show()
        log.Println("准备调用 Focus()...")
        favoritesWindow.Focus()
        log.Println("✓ Focus() 完成")
    }

    log.Println("=== 喜爱音乐窗口操作完成 ===")
})
```

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
检查 favoritesWindow 是否为 nil
    ↓
如果窗口已显示 → Hide()
如果窗口未显示 → Show() + Focus()
    ↓
窗口自动加载 #/favorites 路由
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

#### main.go - 窗口创建与事件绑定
```go
// 1. 创建窗口（在 browseWindow 之后）
var favoritesWindow *application.WebviewWindow
favoritesWindow = app.Window.NewWithOptions(...)

// 2. 注册关闭拦截
favoritesWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
    favoritesWindow.Hide()
    e.Cancel()
})

// 3. 初始隐藏
favoritesWindow.Hide()

// 4. 绑定菜单点击事件（在窗口创建后）
favoriteItem.OnClick(func(ctx *application.Context) {
    if favoritesWindow.IsVisible() {
        favoritesWindow.Hide()
    } else {
        favoritesWindow.Show()
        favoritesWindow.Focus()
    }
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
4. **多窗口管理**: 可以同时打开主窗口、浏览窗口和喜爱音乐窗口

## 注意事项

1. **数据依赖**: 喜爱音乐基于播放历史，需要先有播放记录才会显示
2. **数量限制**: 最多显示 100 首，避免列表过长影响性能
3. **实时更新**: 每次进入页面都会重新加载最新数据
4. **播放次数**: 同一首歌重复播放会累加次数，而非新增记录
5. **窗口管理**: 
   - 点击关闭按钮会隐藏窗口而非退出程序
   - 可以通过托盘菜单再次显示
   - 支持快捷键 `Cmd/Ctrl+H` 快速切换

## Wails v3 多窗口规范遵循

### 1. 窗口创建时序
- ✅ 在所有菜单初始化之后创建窗口
- ✅ 在 `app.Run()` 之前完成所有窗口创建
- ✅ 窗口变量先声明后初始化

### 2. 菜单回调绑定时序
- ✅ **延迟绑定原则**: favoriteItem 的 OnClick 回调在 favoritesWindow 创建**之后**才绑定
- ✅ 避免引用未就绪的 nil 窗口变量
- ✅ 防止空指针异常

### 3. 窗口关闭拦截
- ✅ 使用 `RegisterHook(events.Common.WindowClosing, ...)` 统一处理
- ✅ 跨平台兼容（Windows/Linux/macOS）
- ✅ 调用 `e.Cancel()` 阻止默认关闭行为

### 4. 窗口显示控制
- ✅ 初始状态隐藏 (`Hide()`)
- ✅ 通过菜单项控制显示/隐藏切换
- ✅ 显示时自动获取焦点 (`Focus()`)

## 未来优化方向

1. **导出功能**: 支持导出喜爱音乐列表为 M3U 或 TXT 格式
2. **筛选功能**: 按艺术家、专辑、时间范围筛选
3. **收藏功能**: 允许用户手动标记喜爱歌曲（独立于播放次数）
4. **统计图表**: 可视化展示听歌习惯和趋势
5. **分享功能**: 生成喜爱音乐清单图片分享到社交媒体
6. **窗口布局记忆**: 记住用户调整的窗口大小和位置

## 相关文件

- `frontend/src/views/FavoritesView.vue` - 喜爱音乐视图组件
- `frontend/src/App.vue` - 路由配置
- `main.go` - 独立窗口创建与托盘菜单集成
- `backend/historymanager.go` - 历史记录管理器（提供 GetFavoriteTracks）
- `backend/music_service.go` - 音乐服务层（暴露 API）

## 与单视图版本的对比

| 特性 | 单视图版本 | 独立窗口版本 |
|------|-----------|------------|
| 实现方式 | 在主窗口内切换视图 | 创建独立的 WebviewWindow |
| 用户体验 | 需要返回导航 | 可直接关闭窗口 |
| 多任务 | 不支持同时查看 | 可同时打开多个窗口 |
| 内存占用 | 较低 | 略高（额外窗口实例） |
| 代码复杂度 | 简单 | 中等（需管理窗口生命周期） |
| 适用场景 | 简单应用 | 专业工具类应用 |

**推荐**: 对于音乐播放器这类工具应用，**独立窗口版本**提供更好的用户体验，用户可以同时管理多个视图。