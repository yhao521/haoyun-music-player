# 音乐库通知功能 - 快速参考

## 🚀 快速开始

### 1. 运行应用
```bash
wails3 dev -config ./build/config.yml
```

### 2. 测试通知
- **添加音乐库**: 托盘菜单 → 音乐 → 添加新音乐库
- **刷新音乐库**: 托盘菜单 → 音乐 → 刷新当前音乐库 (或 Ctrl+R / Cmd+R)

## 📝 发送通知

### 后端 Go 代码
```go
// 信息通知
app.Event.Emit("showNotification", map[string]interface{}{
    "title":   t("notification.info"),
    "message": "操作进行中...",
    "type":    "info",
})

// 成功通知
app.Event.Emit("showNotification", map[string]interface{}{
    "title":   t("notification.success"),
    "message": "操作成功!",
    "type":    "success",
})

// 错误通知
app.Event.Emit("showNotification", map[string]interface{}{
    "title":   t("notification.error"),
    "message": fmt.Sprintf("操作失败: %v", err),
    "type":    "error",
})
```

### 前端 TypeScript 代码
```typescript
import { Events } from "@wailsio/runtime";

// 发送通知到后端(如果需要)
Events.Emit("showNotification", {
  title: "通知标题",
  message: "通知内容",
  type: "success" // "success" | "info" | "error"
});
```

## 🎨 通知类型

| 类型 | 颜色 | 图标 | 用途 |
|------|------|------|------|
| `success` | 绿色 (#52c41a) | ✓ | 操作成功 |
| `info` | 蓝色 (#1890ff) | ℹ | 提示信息 |
| `error` | 红色 (#ff4d4f) | ✕ | 错误信息 |

## 🔑 常用翻译键

### 后端 (`t()` 函数)
```go
t("library.addSuccess")              // 音乐库添加成功
t("library.refreshSuccess")          // 音乐库刷新成功
t("library.scanning")                // 正在扫描音乐库...
t("library.loadingToPlaylist")       // 正在加载到播放列表...
t("notification.success")            // 成功
t("notification.info")               // 提示
t("notification.error")              // 错误
```

### 前端 (`t()` 函数)
```typescript
import { t } from "../i18n";

t("notification.libraryAddSuccess")           // 音乐库添加成功
t("notification.libraryRefreshSuccess")       // 音乐库刷新成功
t("notification.scanning")                    // 正在扫描音乐库...
t("notification.loadingToPlaylist")           // 正在加载到播放列表...
```

## ⚙️ 自定义样式

编辑 `frontend/src/components/NotificationToast.vue`:

```css
/* 修改位置 */
.notification-container {
  top: 20px;      /* 距离顶部 */
  right: 20px;    /* 距离右侧 */
  left: auto;     /* 改为左侧: left: 20px; right: auto; */
}

/* 修改尺寸 */
.notification {
  min-width: 300px;  /* 最小宽度 */
  max-width: 400px;  /* 最大宽度 */
}

/* 修改持续时间 */
setTimeout(() => {
  removeNotification(id);
}, 3000);  // 改为 5000 = 5秒
```

## 🐛 常见问题

### Q: 通知不显示?
**A**: 检查以下几点:
1. 确认事件已注册: `application.RegisterEvent[map[string]interface{}]("showNotification")`
2. 确认前端组件已集成: `<NotificationToast />` 在 `App.vue` 中
3. 查看浏览器控制台是否有错误信息
4. 确认数据类型正确: `{title, message, type}`

### Q: 通知立即消失?
**A**: 调整超时时间:
```javascript
setTimeout(() => {
  removeNotification(id);
}, 5000);  // 增加到 5 秒
```

### Q: 通知重叠在一起?
**A**: 这是正常行为,多个通知会垂直堆叠。如需限制数量:
```javascript
if (notifications.value.length >= 5) {
  notifications.value.shift();  // 移除最旧的通知
}
```

### Q: 如何禁用通知?
**A**: 注释掉 `App.vue` 中的 `<NotificationToast />`:
```vue
<!-- <NotificationToast /> -->
```

## 📚 相关文件

- **后端逻辑**: `main.go` (第 346-500 行)
- **通知组件**: `frontend/src/components/NotificationToast.vue`
- **应用入口**: `frontend/src/App.vue`
- **国际化**: 
  - 后端: `backend/pkg/i18n/{locale}.json`
  - 前端: `frontend/src/i18n/locales/{locale}.json`
- **详细文档**: `LIBRARY_NOTIFICATION.md`
- **实现总结**: `LIBRARY_NOTIFICATION_SUMMARY.md`

## 🔗 相关记忆

已创建项目规范记忆:
- **系统通知栏通知实施规范**
  - 后端事件发送规范
  - 前端通知组件实现
  - 国际化要求
  - 跨平台兼容性
  - 最佳实践
