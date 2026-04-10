# 音乐库通知功能说明

## 📋 功能概述

在音乐库添加和刷新操作完成后,系统会自动发送通知栏通知,支持 Windows 和 macOS 平台。

## ⚠️ 重要说明

### Wails v3 原生通知支持情况

**Wails v3 (v3.0.0-alpha.74) 目前不提供原生的系统通知 API。**

根据官方文档和 API 参考,Wails v3 的运行时 API 包括:
- ✅ 窗口管理 (Window Management)
- ✅ 对话框 (Dialogs - 文件选择、消息提示等)
- ✅ 事件系统 (Events)
- ✅ 日志 (Logging)
- ✅ 菜单和托盘 (Menu & System Tray)
- ❌ **原生系统通知 (Native System Notifications)** - **暂不支持**

### 当前实现方案

由于 Wails v3 不支持原生通知,我们采用了 **HTML/CSS 自定义通知组件**方案:

**优势:**
- ✅ **跨平台一致**: 在 Windows、macOS、Linux 上表现完全一致
- ✅ **完全可控**: 样式、动画、行为都可以自定义
- ✅ **无需权限**: 不需要请求系统通知权限
- ✅ **即时显示**: 不依赖系统通知中心的状态
- ✅ **应用内集成**: 与应用 UI 风格统一

**局限性:**
- ⚠️ 仅在应用窗口可见时显示
- ⚠️ 应用最小化或隐藏时用户可能看不到通知
- ⚠️ 无法在锁屏界面显示
- ⚠️ 不会出现在系统通知中心的历史记录中

### 未来改进方向

当以下情况发生时,可以考虑切换到原生通知:

1. **Wails v3 正式版发布原生通知 API**
   - 关注 Wails 官方更新: https://wails.io
   - 检查 `application.Notification` 或 `runtime.Notification` API

2. **使用第三方 Go 库**
   - [notify](https://github.com/gen2brain/beeep) - 跨平台通知库
   - [gosx-notifier](https://github.com/deckarep/gosx-notifier) - macOS 专用
   - [toast](https://github.com/go-toast/toast) - Windows 专用

3. **混合方案**
   - 应用在前台时使用自定义组件
   - 应用在后台时使用原生通知(如果可用)

## ✨ 功能特性

### 1. 通知类型

- **信息通知 (Info)**: 蓝色,用于显示操作进行中的状态
- **成功通知 (Success)**: 绿色,用于显示操作成功
- **错误通知 (Error)**: 红色,用于显示操作失败

### 2. 通知场景

#### 添加音乐库时

1. **开始添加**: 显示"正在扫描音乐库..."提示
2. **加载播放列表**: 显示"正在加载到播放列表..."提示
3. **添加成功**: 显示"音乐库添加成功: [音乐库名称] (X 首歌曲)"
4. **添加失败**: 显示具体错误信息

#### 刷新音乐库时

1. **开始刷新**: 显示"正在扫描音乐库..."提示
2. **加载播放列表**: 显示"正在加载到播放列表..."提示
3. **刷新成功**: 显示"音乐库刷新成功: [音乐库名称] (X 首歌曲)"
4. **刷新失败**: 显示具体错误信息

### 3. 通知行为

- **自动消失**: 通知显示 3 秒后自动消失
- **手动关闭**: 点击通知右上角的 × 按钮可立即关闭
- **位置**: 固定在窗口右上角
- **动画**: 滑入/滑出过渡动画效果
- **堆叠**: 多个通知会垂直堆叠显示

## 🔧 技术实现

### 后端实现

1. **事件注册** (`main.go`):
```go
application.RegisterEvent[map[string]interface{}]("showNotification")
```

2. **发送通知**:
```go
app.Event.Emit("showNotification", map[string]interface{}{
    "title":   "通知标题",
    "message": "通知内容",
    "type":    "success", // success | info | error
})
```

### 前端实现

1. **通知组件** (`NotificationToast.vue`):
   - 监听 `showNotification` 事件
   - 根据类型显示不同样式
   - 自动管理通知生命周期

2. **国际化支持**:
   - 后端: `backend/pkg/i18n/{locale}.json`
   - 前端: `frontend/src/i18n/locales/{locale}.json`

## 🌍 国际化

### 新增翻译键

#### 后端 (`library` 模块)
- `library.addSuccess`: 音乐库添加成功
- `library.addSuccessWithCount`: 音乐库添加成功,共 %d 首歌曲
- `library.refreshSuccess`: 音乐库刷新成功
- `library.refreshSuccessWithCount`: 音乐库刷新成功,共 %d 首歌曲
- `library.scanning`: 正在扫描音乐库...
- `library.loadingToPlaylist`: 正在加载到播放列表...

#### 后端 (`notification` 模块)
- `notification.success`: 成功
- `notification.info`: 提示
- `notification.error`: 错误

#### 前端 (`notification` 模块)
- 与后端保持一致的翻译键

## 🧪 测试方法

### 1. 测试添加音乐库

1. 启动应用
2. 右键点击系统托盘图标
3. 选择"音乐" → "添加新音乐库"
4. 选择一个包含音乐的文件夹
5. 观察通知:
   - 应立即显示"正在扫描音乐库..."
   - 约 2 秒后显示"正在加载到播放列表..."
   - 最后显示"音乐库添加成功: [名称] (X 首歌曲)"

### 2. 测试刷新音乐库

1. 确保已添加至少一个音乐库
2. 右键点击系统托盘图标
3. 选择"音乐" → "刷新当前音乐库" (或按 Ctrl+R / Cmd+R)
4. 观察通知流程(同添加)

### 3. 测试错误情况

1. 尝试添加一个空文件夹或不存在的文件夹
2. 应显示错误通知,包含具体错误信息

### 4. 测试多语言

1. 进入设置页面
2. 切换语言(中文 ↔ 英文)
3. 重复上述测试,验证通知文本正确翻译

### 5. 测试通知交互

1. 触发多个通知(快速连续添加/刷新)
2. 验证通知正确堆叠显示
3. 点击 × 按钮,验证通知立即关闭
4. 等待 3 秒,验证通知自动消失

## 🎨 样式定制

如需调整通知样式,编辑 `NotificationToast.vue` 中的 CSS:

```css
/* 修改位置 */
.notification-container {
  top: 20px;      /* 距离顶部 */
  right: 20px;    /* 距离右侧 */
}

/* 修改尺寸 */
.notification {
  min-width: 300px;  /* 最小宽度 */
}

/* 修改颜色 */
.notification-success {
  border-left: 4px solid #52c41a;  /* 成功色 */
  background: #f6ffed;
}

/* 修改持续时间 */
setTimeout(() => {
  removeNotification(id);
}, 3000);  // 改为其他毫秒数
```

## 💡 切换到原生通知的方案

如果未来需要原生通知,可以参考以下方案:

### 方案 1: 使用 beeep 库(推荐)

```go
import "github.com/gen2brain/beeep"

// 发送通知
err := beeep.Notify(
    "音乐库添加成功",
    fmt.Sprintf("%s (%d 首歌曲)", libName, len(tracks)),
    "", // 图标路径,空则使用默认图标
)
if err != nil {
    log.Printf("发送通知失败: %v", err)
}
```

**优点:**
- 跨平台(Windows/macOS/Linux)
- 简单易用
- 维护活跃

**缺点:**
- 需要 CGO 支持
- 在某些 Linux 发行版上可能需要额外依赖

### 方案 2: 平台特定实现

```go
import (
    "runtime"
    "github.com/deckarep/gosx-notifier" // macOS
    "github.com/go-toast/toast"         // Windows
)

func sendNativeNotification(title, message string) {
    switch runtime.GOOS {
    case "darwin":
        // macOS
        notification := gosxnotifier.NewNotification()
        notification.Title = title
        notification.Message = message
        notification.Push()
        
    case "windows":
        // Windows
        notification := toast.Notification{
            AppID:   "Haoyun Music Player",
            Title:   title,
            Message: message,
        }
        notification.Push()
        
    default:
        // 回退到自定义通知
        app.Event.Emit("showNotification", map[string]interface{}{
            "title":   title,
            "message": message,
            "type":    "info",
        })
    }
}
```

## ⚠️ 注意事项

1. **跨平台兼容**: 
   - 当前使用 HTML/CSS 自定义通知组件
   - 在 Windows、macOS、Linux 上表现一致
   - Wails v3 暂不支持原生通知 API

2. **性能考虑**:
   - 避免频繁发送通知(建议间隔 > 1 秒)
   - 同时显示的通知不超过 5 个

3. **用户体验**:
   - 重要操作才发送通知
   - 错误通知必须清晰说明问题
   - 成功通知应包含关键信息(如歌曲数量)
   
4. **应用状态**:
   - 自定义通知仅在应用窗口可见时显示
   - 如果需要在后台通知,考虑使用原生方案

## 📝 扩展建议

未来可以考虑的功能:

1. **通知持久化**: 将重要通知保存到历史记录
2. **通知分组**: 相同类型的通知合并显示
3. **自定义时长**: 根据通知类型设置不同的显示时间
4. **声音提示**: 为不同类型的通知添加提示音
5. **原生集成**: 当 Wails 支持或使用第三方库时,使用系统原生通知 API
6. **混合模式**: 应用在前台用自定义通知,后台用原生通知
