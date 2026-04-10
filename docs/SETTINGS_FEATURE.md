# 设置窗口功能实现

## 概述

本次更新为 Haoyun Music Player 添加了完整的设置窗口功能，用户可以通过托盘菜单或快捷键访问应用程序设置。

## 实现内容

### 1. 后端实现 (main.go)

#### 1.1 创建设置窗口
- **位置**: 在 favoritesWindow 之后创建
- **URL**: `#/settings`
- **尺寸**: 600x500（比浏览/喜爱窗口更紧凑）
- **特性**: 
  - macOS 透明标题栏
  - 深色背景主题
  - 启动时默认隐藏

```go
var settingsWindow *application.WebviewWindow
settingsWindow = app.Window.NewWithOptions(application.WebviewWindowOptions{
    Title: "设置 - Haoyun Music Player",
    Mac: application.MacWindow{
        InvisibleTitleBarHeight: 50,
        Backdrop:                application.MacBackdropTranslucent,
        TitleBar:                application.MacTitleBarHiddenInset,
    },
    BackgroundColour: application.NewRGB(27, 38, 54),
    URL:              "#/settings",
    Width:            600,
    Height:           500,
})
settingsWindow.Hide()
```

#### 1.2 注册关闭拦截钩子
遵循 Wails v3 窗口生命周期规范，为所有窗口（包括设置窗口）注册关闭拦截：

```go
settingsWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
    log.Println("⚠️ [设置窗口] 关闭事件触发")
    
    if hasOtherVisibleWindows("settings") {
        log.Println("ℹ️ [设置窗口] 检测到其他可见窗口，但仍执行隐藏操作")
    } else {
        log.Println("ℹ️ [设置窗口] 这是最后一个可见窗口")
    }
    
    // 统一行为：所有窗口关闭时都隐藏，不真正关闭
    settingsWindow.Hide()
    e.Cancel() // 取消关闭操作
    log.Println("✅ [设置窗口] 已隐藏并取消关闭")
})
```

**核心原则**: 
- ❌ **严禁销毁实例**: 窗口关闭按钮 (X) 仅作为隐藏交互
- ✅ **统一隐藏策略**: 调用 `Hide()` + `e.Cancel()` 阻止默认关闭
- 🔒 **退出机制**: 唯一退出入口为托盘"退出"菜单项

#### 1.3 更新菜单项点击事件
延迟绑定 settingItem.OnClick，确保窗口实例已创建：

```go
settingItem.OnClick(func(ctx *application.Context) {
    log.Println("打开设置窗口")
    
    defer func() {
        if r := recover(); r != nil {
            log.Printf("❌ 打开设置窗口时发生 panic: %v", r)
            debug.PrintStack()
        }
    }()
    
    if settingsWindow == nil {
        log.Println("❌ settingsWindow 为 nil")
        return
    }
    
    isVisible := settingsWindow.IsVisible()
    if isVisible {
        settingsWindow.Hide()
    } else {
        settingsWindow.Show()
        settingsWindow.Focus()
    }
})
```

**快捷键**: Cmd/Ctrl+S

#### 1.4 辅助函数更新
更新 `hasOtherVisibleWindows` 函数，添加 settingsWindow 检查：

```go
hasOtherVisibleWindows := func(currentWindow string) bool {
    switch currentWindow {
    case "settings":
        return (mainWindow != nil && mainWindow.IsVisible()) ||
            (browseWindow != nil && browseWindow.IsVisible()) ||
            (favoritesWindow != nil && favoritesWindow.IsVisible())
    // ... 其他情况
    }
}
```

### 2. 前端实现 (SettingsView.vue)

#### 2.1 组件结构
创建 `/frontend/src/views/SettingsView.vue`，包含以下模块：

**顶部标题栏**:
- 返回按钮（←）：点击返回主界面
- 居中标题："⚙️ 设置"
- 刷新按钮（🔄）：预留功能

**设置内容区域**:
1. **通用设置**
   - 开机自动启动（复选框）
   - 保持系统唤醒（复选框，默认启用）
   - 主题模式选择（下拉框：跟随系统/浅色/深色）

2. **播放设置**
   - 默认播放模式（下拉框：列表循环/顺序/随机/单曲循环）
   - 显示歌词（复选框，默认启用）
   - 音量调节（滑块 0-100%，默认 80%）

3. **媒体键设置**
   - 启用媒体键控制（复选框，默认启用）

4. **关于信息**
   - 应用名称：Haoyun Music Player
   - 版本：0.5.0
   - 技术栈：Wails v3 + Vue 3

#### 2.2 UI 设计规范

**标题居中布局**:
```css
.title-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}
```

**滚动容器优化**:
```css
.settings-content {
  flex: 1;
  overflow-y: auto;
  min-height: 0; /* 关键：允许 Flex 子项滚动 */
}
```

**视觉风格**:
- 渐变背景：`linear-gradient(135deg, #1a1a2e 0%, #16213e 100%)`
- 半透明卡片：`rgba(255, 255, 255, 0.05)`
- 强调色：`#4fc3f7`（天蓝色）
- 悬停效果：边框高亮 + 轻微位移

### 3. 路由配置 (App.vue)

#### 3.1 导入组件
```typescript
import SettingsView from "./views/SettingsView.vue";
```

#### 3.2 路由判断逻辑
在 `checkRoute()` 函数中添加 settings 路由匹配：

```typescript
// 检查 hash 路由（#/settings）
if (hash === "#/settings" || hash.startsWith("#/settings/")) {
  console.log("[路由匹配] 匹配到 settings 视图");
  currentView.value = "settings";
  return;
}

// 检查 path 路由（/settings）
if (pathname === "/settings" || pathname.startsWith("/settings/")) {
  console.log("[路由匹配] 匹配到 settings 视图 (pathname)");
  currentView.value = "settings";
  return;
}
```

#### 3.3 动态组件渲染
```vue
<template>
  <component :is="
    currentView === 'browse' ? BrowseView : 
    currentView === 'favorites' ? FavoritesView : 
    currentView === 'settings' ? SettingsView : 
    AppMain
  " />
</template>
```

## 访问方式

### 1. 托盘菜单
- 点击系统托盘图标
- 选择"设置"菜单项
- 或使用快捷键 **Cmd/Ctrl+S**

### 2. 路由导航
- Hash 模式：`#/settings`
- Path 模式：`/settings`

### 3. 程序化导航
```javascript
window.location.hash = "#/settings";
```

## 关键技术要点

### 1. 窗口生命周期管理
- **创建时序**: 在所有窗口变量声明后、菜单定义前创建
- **初始化流程**: 创建 → Hide() → 注册 Hook → 绑定 OnClick
- **关闭行为**: 拦截 WindowClosing 事件，执行 Hide() + Cancel()
- **退出机制**: 仅通过托盘"退出"菜单调用 `app.Quit()`

### 2. 依赖注入时序
```
1. 声明所有窗口变量（初始化为 nil）
2. 创建基本菜单项（不含 OnClick）
3. 创建所有窗口并立即 Hide()
4. 统一注册所有窗口的关闭拦截钩子
5. 绑定菜单项的 OnClick 回调
6. 启动应用
```

### 3. 前端路由兼容
- 同时支持 hash (`#/settings`) 和 path (`/settings`) 模式
- 优先匹配 hash 路由
- 监听 `hashchange` 事件实现动态切换
- 使用动态组件 `<component :is="...">` 渲染

### 4. Flex 布局最佳实践
- **滚动条失效解决**: 为滚动子项设置 `min-height: 0`
- **标题居中**: 使用 `flex: 1` + `align-items: center`
- **防止压缩**: 按钮设置 `flex-shrink: 0` + `white-space: nowrap`

### 5. 错误处理
- 所有窗口操作添加 `defer/recover` panic 恢复
- nil 检查避免空指针异常
- 详细日志记录便于调试（包含 emoji 标识）

## 测试建议

### 1. 窗口显示/隐藏
- [ ] 点击托盘"设置"菜单，窗口应显示
- [ ] 再次点击，窗口应隐藏
- [ ] 点击窗口关闭按钮 (X)，窗口应隐藏而非关闭
- [ ] 使用快捷键 Cmd/Ctrl+S，窗口应切换显示/隐藏状态

### 2. 多窗口交互
- [ ] 同时打开多个窗口（主窗口、浏览、喜爱、设置）
- [ ] 关闭其中一个窗口，其他窗口应保持可见
- [ ] 关闭所有窗口后，应用不应退出（仅隐藏）

### 3. 路由导航
- [ ] 直接访问 `#/settings`，应显示设置页面
- [ ] 在设置页面点击返回按钮，应返回主界面
- [ ] 从其他页面切换到设置页面，路由应正确更新

### 4. UI/UX
- [ ] 标题居中显示
- [ ] 滚动条正常工作
- [ ] 复选框、下拉框、滑块交互正常
- [ ] 悬停效果流畅
- [ ] 响应式布局适配不同窗口尺寸

## 已知限制

1. **设置持久化**: 当前设置项仅为 UI 展示，未实现后端存储和读取
2. **功能实现**: 部分设置项（如开机启动、媒体键）标记为 TODO，待后续实现
3. **主题切换**: 主题模式选择暂未联动实际主题切换逻辑

## 后续优化方向

1. **设置持久化**: 使用 JSON 文件或本地存储保存用户设置
2. **实时生效**: 部分设置（如音量、播放模式）应立即应用到播放器
3. **国际化**: 支持多语言切换
4. **高级设置**: 添加音频输出设备选择、歌词同步偏移等高级选项
5. **设置导入/导出**: 支持备份和恢复设置

## 相关文件

- `/main.go`: 后端窗口创建和管理
- `/frontend/src/App.vue`: 前端路由配置
- `/frontend/src/views/SettingsView.vue`: 设置页面组件
- `/frontend/public/style.css`: 全局样式（如需调整）

## 参考文档

- [Wails v3 窗口生命周期规范](./BACKEND_DESIGN.md)
- [Flex 布局滚动条失效解决方案](memory:0568361d-febd-4145-a139-f2e356b644b9)
- [视图标题布局规范](memory:ececc50f-f43b-4ae9-8b8e-22a86f7d5101)
