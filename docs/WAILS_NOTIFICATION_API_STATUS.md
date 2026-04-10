# Wails v3 通知 API 支持情况说明

## 📌 当前状态 (2026-04-09)

**Wails v3 (v3.0.0-alpha.74) 不提供原生的系统通知 API。**

## 🔍 调查结果

### Wails v3 运行时 API 清单

根据官方文档和代码审查,Wails v3 提供的 API 包括:

#### ✅ 已支持的 API

1. **窗口管理 (Window Management)**
   - `window.Show()`, `window.Hide()`
   - `window.SetTitle()`, `window.SetSize()`
   - `window.Center()`, `window.Maximize()`

2. **对话框 (Dialogs)**
   - `OpenFile()`, `OpenFiles()` - 文件选择
   - `SaveFile()` - 保存文件
   - `MessageDialog()` - 消息提示对话框
   - `QuestionDialog()` - 询问对话框

3. **事件系统 (Events)**
   - `app.Event.Emit()` - 发送事件
   - `application.RegisterEvent[T]()` - 注册事件类型
   - `Events.On()`, `Events.Off()` - 前端监听

4. **菜单和托盘 (Menu & System Tray)**
   - `application.NewMenuItem()`
   - `tray.SetMenu()`, `tray.SetTooltip()`
   - `tray.OnClick()`, `tray.OnDoubleClick()`

5. **日志 (Logging)**
   - `app.Log.Info()`, `app.Log.Error()`
   - `app.Log.Debug()`, `app.Log.Warn()`

6. **其他**
   - `app.Quit()` - 退出应用
   - `app.Environment()` - 环境信息
   - `clipboard.Write()`, `clipboard.Read()` - 剪贴板

#### ❌ 未支持的 API

- **原生系统通知 (Native System Notifications)**
  - 没有 `application.Notification` 
  - 没有 `runtime.Notification`
  - 没有 `ShowNotification()` 方法

### 官方路线图

根据 [Wails v3 路线](https://wails.io/zh-Hans/blog/the-road-to-wails-v3/) 博客:

> Wails v3 正在积极开发中,主要改进包括:
> - 原生移动端支持 (iOS/Android)
> - 插件系统 (类似 Tauri 的插件)
> - 更好的多窗口支持
> - **改进的运行时 API**

但截至目前,**原生通知 API 尚未被列入明确的发布计划**。

## 💡 解决方案对比

### 方案 1: HTML/CSS 自定义通知 ⭐ (当前采用)

**实现:** 使用 Vue 组件 + CSS 动画

**优点:**
- ✅ 完全跨平台一致
- ✅ 无需额外依赖
- ✅ 样式完全可控
- ✅ 不需要系统权限
- ✅ 与应用 UI 风格统一
- ✅ 即时显示,无延迟

**缺点:**
- ⚠️ 仅在应用窗口可见时显示
- ⚠️ 应用最小化时看不到
- ⚠️ 不出现在系统通知中心
- ⚠️ 锁屏时无法显示

**适用场景:**
- 应用在前台运行时的用户反馈
- 操作成功/失败的即时提示
- 进度状态显示

**代码示例:**
```vue
<!-- NotificationToast.vue -->
<template>
  <div class="notification-container">
    <TransitionGroup name="notification">
      <div v-for="n in notifications" :key="n.id" 
           class="notification" :class="`notification-${n.type}`">
        <div class="notification-content">
          <div class="notification-title">{{ n.title }}</div>
          <div class="notification-message">{{ n.message }}</div>
        </div>
      </div>
    </TransitionGroup>
  </div>
</template>
```

### 方案 2: beeep 库 (推荐用于后台通知)

**仓库:** https://github.com/gen2brain/beeep

**安装:**
```bash
go get github.com/gen2brain/beeep
```

**使用:**
```go
import "github.com/gen2brain/beeep"

err := beeep.Notify(
    "音乐库添加成功",
    "共 100 首歌曲",
    "", // 图标路径
)
```

**优点:**
- ✅ 真正的原生通知
- ✅ 应用后台时也能显示
- ✅ 出现在系统通知中心
- ✅ 跨平台(Windows/macOS/Linux)
- ✅ 简单易用

**缺点:**
- ⚠️ 需要 CGO 支持
- ⚠️ Linux 可能需要 libnotify
- ⚠️ macOS 需要用户授权
- ⚠️ 样式不可控(由系统决定)
- ⚠️ 增加二进制文件大小

**适用场景:**
- 应用在后台运行时的通知
- 定时任务完成提醒
- 下载完成通知

### 方案 3: 平台特定实现

**macOS:** gosx-notifier
```go
import "github.com/deckarep/gosx-notifier"

notification := gosxnotifier.NewNotification()
notification.Title = "标题"
notification.Message = "内容"
notification.Push()
```

**Windows:** go-toast
```go
import "github.com/go-toast/toast"

notification := toast.Notification{
    AppID:   "MyApp",
    Title:   "标题",
    Message: "内容",
}
notification.Push()
```

**Linux:** notify-send (命令行)
```go
exec.Command("notify-send", "标题", "内容").Run()
```

**优点:**
- ✅ 完全原生体验
- ✅ 可以深度定制每个平台

**缺点:**
- ❌ 代码复杂,需要条件编译
- ❌ 维护成本高
- ❌ 不同平台行为不一致

### 方案 4: HTML5 Notification API

**注意:** 在 Wails WebView 中**不推荐使用**

```javascript
// 在浏览器中工作,但在 Wails 中可能不支持
if ("Notification" in window) {
  Notification.requestPermission().then(permission => {
    if (permission === "granted") {
      new Notification("标题", { body: "内容" });
    }
  });
}
```

**问题:**
- ❌ Wails WebView 可能不支持
- ❌ 需要用户授权
- ❌ 行为因平台而异
- ❌ 不如自定义组件可靠

## 🎯 推荐策略

### 当前阶段 (Wails v3 Alpha)

**采用方案 1 (HTML/CSS 自定义通知)**

理由:
1. Wails v3 尚未提供原生通知 API
2. 自定义通知已经满足大部分需求
3. 跨平台一致性最好
4. 无需额外依赖和维护成本

### 未来升级路径

**当以下条件满足时考虑切换:**

1. **Wails v3 正式版发布原生通知 API**
   ```go
   // 假设的未来 API
   app.Notification.Show(&NotificationOptions{
       Title:   "标题",
       Message: "内容",
       Type:    NotificationSuccess,
   })
   ```

2. **需要后台通知功能**
   - 切换到 beeep 库
   - 或实现混合方案(前台用自定义,后台用原生)

3. **用户明确要求原生通知**
   - 提供设置选项让用户选择
   - 默认保持自定义通知

### 混合方案示例

```go
func sendNotification(title, message string, notificationType string) {
    // 检查应用是否在前台
    if isAppInForeground() {
        // 前台: 使用自定义通知(更快,更美观)
        app.Event.Emit("showNotification", map[string]interface{}{
            "title":   title,
            "message": message,
            "type":    notificationType,
        })
    } else {
        // 后台: 使用原生通知(确保用户能看到)
        if err := beeep.Notify(title, message, ""); err != nil {
            log.Printf("原生通知失败: %v", err)
        }
    }
}
```

## 📊 方案对比表

| 特性 | 自定义通知 | beeep | 平台特定 | HTML5 API |
|------|-----------|-------|---------|-----------|
| 跨平台一致性 | ✅ 完美 | ✅ 良好 | ❌ 需适配 | ⚠️ 不确定 |
| 后台通知 | ❌ 不支持 | ✅ 支持 | ✅ 支持 | ⚠️ 可能 |
| 样式可控 | ✅ 完全 | ❌ 系统决定 | ❌ 系统决定 | ⚠️ 有限 |
| 系统权限 | ✅ 不需要 | ⚠️ macOS需要 | ⚠️ 需要 | ❌ 需要 |
| 依赖复杂度 | ✅ 无 | ⚠️ 中等 | ❌ 高 | ✅ 无 |
| 维护成本 | ✅ 低 | ✅ 低 | ❌ 高 | ✅ 低 |
| 通知中心集成 | ❌ 否 | ✅ 是 | ✅ 是 | ⚠️ 可能 |
| 锁屏显示 | ❌ 否 | ✅ 是 | ✅ 是 | ❌ 否 |
| 即时性 | ✅ 立即 | ⚠️ 略有延迟 | ⚠️ 略有延迟 | ⚠️ 不确定 |

## 🔗 相关资源

- **Wails 官方文档**: https://wails.io
- **Wails v3 路线图**: https://wails.io/zh-Hans/blog/the-road-to-wails-v3/
- **beeep 库**: https://github.com/gen2brain/beeep
- **gosx-notifier**: https://github.com/deckarep/gosx-notifier
- **go-toast**: https://github.com/go-toast/toast
- **本项目通知实现**: [LIBRARY_NOTIFICATION.md](./LIBRARY_NOTIFICATION.md)

## 📝 总结

1. **Wails v3 目前不支持原生通知** - 这是框架限制,不是实现问题
2. **自定义通知是最佳选择** - 对于应用内即时反馈,HTML/CSS 方案最优
3. **beeep 是备用方案** - 如果需要后台通知,可以考虑集成
4. **关注 Wails 更新** - 未来版本可能会添加原生通知 API
5. **保持灵活性** - 设计时预留切换空间,便于未来升级

---

**更新日期:** 2026-04-09  
**Wails 版本:** v3.0.0-alpha.74  
**调查者:** AI Assistant
