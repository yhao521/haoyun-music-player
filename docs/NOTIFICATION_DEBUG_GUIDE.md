# 通知功能调试指南

## 🐛 问题描述

托盘菜单中"添加音乐库"和"刷新音乐库"操作没有触发通知显示。

## 🔍 根本原因

**Wails v3 事件数据结构问题**:

在 Wails v3 中,通过 `Events.On()` 接收的事件数据被包装在一个对象中,实际数据位于 `eventData.data` 属性中。

### 错误的数据访问方式

```typescript
// ❌ 错误:直接访问 eventData
Events.On("showNotification", (data: any) => {
  if (data.title && data.message) {  // data.title 是 undefined
    showNotification(data.title, data.message, data.type);
  }
});
```

### 正确的数据访问方式

```typescript
// ✅ 正确:从 eventData.data 中提取实际数据
Events.On("showNotification", (eventData: any) => {
  const data = eventData?.data || eventData;  // 兼容两种情况
  if (data.title && data.message) {
    showNotification(data.title, data.message, data.type);
  }
});
```

## ✅ 已修复

### 修改的文件

**`frontend/src/components/NotificationToast.vue`**

```typescript
onMounted(() => {
  Events.On("showNotification", (eventData: any) => {
    console.log("[Notification] 收到通知事件:", eventData);
    
    // Wails v3 事件数据结构: eventData.data 包含实际数据
    const data = eventData?.data || eventData;
    
    console.log("[Notification] 解析后的数据:", data);
    
    if (data && data.title && data.message) {
      showNotification(data.title, data.message, data.type || "info");
    } else {
      console.warn("[Notification] 无效的通知数据:", data);
    }
  });
});
```

## 🧪 测试步骤

### 1. 启动应用

```bash
wails3 dev -config ./build/config.yml
```

### 2. 打开开发者工具

- macOS: `Cmd + Option + I`
- Windows/Linux: `Ctrl + Shift + I`

### 3. 测试添加音乐库

1. 右键点击系统托盘图标
2. 选择"音乐" → "添加新音乐库"
3. 选择一个包含音乐的文件夹
4. **观察控制台输出**:

应该看到类似以下日志:
```
[Notification] 收到通知事件: {data: {title: "提示", message: "正在扫描音乐库...", type: "info"}}
[Notification] 解析后的数据: {title: "提示", message: "正在扫描音乐库...", type: "info"}
```

5. **观察通知显示**:
   - 右上角应显示蓝色通知:"提示 - 正在扫描音乐库..."
   - 约 2 秒后显示:"提示 - 正在加载到播放列表..."
   - 最后显示绿色通知:"成功 - 音乐库添加成功: [名称] (X 首歌曲)"

### 4. 测试刷新音乐库

1. 确保已添加至少一个音乐库
2. 右键点击系统托盘图标
3. 选择"音乐" → "刷新当前音乐库" (或按 `Ctrl+R` / `Cmd+R`)
4. 观察相同的通知流程

### 5. 验证通知交互

- ✅ 通知应在 3 秒后自动消失
- ✅ 点击 × 按钮应立即关闭通知
- ✅ 多个通知应垂直堆叠显示
- ✅ 不同颜色表示不同类型(蓝/绿/红)

## 🔧 如果仍然不工作

### 检查清单

#### 1. 确认事件已注册

查看 `main.go` 的 `init()` 函数:

```go
func init() {
    // ... 其他事件注册
    application.RegisterEvent[map[string]interface{}]("showNotification")
}
```

#### 2. 确认组件已挂载

在浏览器控制台中运行:

```javascript
// 检查 NotificationToast 组件是否存在
document.querySelector('.notification-container')
```

应该返回一个 DOM 元素。

#### 3. 手动测试事件发送

在 Go 代码中添加测试:

```go
// 在某个菜单项的 OnClick 中
app.Event.Emit("showNotification", map[string]interface{}{
    "title":   "测试通知",
    "message": "这是一条测试消息",
    "type":    "success",
})
```

点击该菜单项,观察是否显示通知。

#### 4. 检查前端控制台

打开浏览器开发者工具,查看:
- 是否有 JavaScript 错误
- 是否收到 `[Notification] 收到通知事件:` 日志
- 如果有日志但没显示,检查数据解析是否正确

#### 5. 检查后端日志

查看终端输出:
- 是否有 `showNotification` 事件发送的日志
- 是否有错误信息

### 常见问题

#### Q1: 控制台显示 "无效的通知数据"

**原因**: 数据结构不符合预期

**解决**: 检查后端发送的数据格式:

```go
// ✅ 正确
app.Event.Emit("showNotification", map[string]interface{}{
    "title":   "标题",
    "message": "内容",
    "type":    "info",
})

// ❌ 错误:缺少必要字段
app.Event.Emit("showNotification", map[string]interface{}{
    "message": "内容",  // 缺少 title
})
```

#### Q2: 收到事件但通知不显示

**可能原因**:
1. CSS 样式问题(通知被遮挡或位置不对)
2. Vue 响应式问题

**调试方法**:

在 `showNotification` 函数中添加日志:

```typescript
const showNotification = (title: string, message: string, type: "success" | "info" | "error" = "info") => {
  console.log("[Notification] 准备显示通知:", { title, message, type });
  const id = ++notificationId;
  notifications.value.push({ id, title, message, type });
  console.log("[Notification] 当前通知列表:", notifications.value);
  
  setTimeout(() => {
    removeNotification(id);
  }, 3000);
};
```

#### Q3: 通知显示但立即消失

**原因**: 定时器设置问题

**解决**: 调整超时时间:

```typescript
setTimeout(() => {
  removeNotification(id);
}, 5000);  // 改为 5 秒
```

#### Q4: 只有部分通知显示

**原因**: 快速连续发送时,前一个通知可能被覆盖

**解决**: 确保每个通知有唯一 ID(已实现)

## 📊 事件数据流

```
后端 Go                          前端 TypeScript
─────────                       ───────────────
                                
app.Event.Emit(                 
  "showNotification",           
  map[string]interface{}{       
    "title": "标题",            
    "message": "内容",          
    "type": "info"              
  }                             
)                               
         │                      
         │  Wails 运行时包装   
         ▼                      
    {                           
      data: {                   
        title: "标题",          
        message: "内容",        
        type: "info"            
      }                         
    }                           
         │                      
         │  Events.On 接收     
         ▼                      
(eventData: any) => {           
  const data = eventData?.data || eventData;
  // data = {title, message, type}
  showNotification(...)         
}                               
```

## 🎯 其他事件的类似问题

如果其他事件也遇到类似问题,可以使用相同的修复方法:

```typescript
// 通用模式
Events.On("eventName", (eventData: any) => {
  const data = eventData?.data || eventData;
  // 使用 data 而不是 eventData
});
```

### 需要检查的事件

根据项目中的事件使用情况:

- ✅ `playbackStateChanged` - 已在 AppMain.vue 中正确使用 `state.data`
- ✅ `playbackProgress` - 已在 AppMain.vue 中直接使用 `data.position`
- ✅ `playlistUpdated` - 已在 AppMain.vue 中处理 `tracks.data`
- ✅ `currentTrackChanged` - 已在 AppMain.vue 中处理 `track.data`
- ⚠️ **`showNotification`** - **已修复**

## 📝 最佳实践

### 1. 统一的事件处理模式

```typescript
// 推荐:创建辅助函数
const unwrapEventData = (eventData: any): any => {
  return eventData?.data || eventData;
};

// 使用
Events.On("showNotification", (eventData: any) => {
  const data = unwrapEventData(eventData);
  // 处理 data
});
```

### 2. 类型安全

```typescript
interface NotificationData {
  title: string;
  message: string;
  type?: "success" | "info" | "error";
}

Events.On("showNotification", (eventData: any) => {
  const data = eventData?.data as NotificationData | NotificationData;
  if (data?.title && data?.message) {
    showNotification(data.title, data.message, data.type || "info");
  }
});
```

### 3. 错误处理

```typescript
Events.On("showNotification", (eventData: any) => {
  try {
    const data = eventData?.data || eventData;
    if (!data?.title || !data?.message) {
      console.error("[Notification] 缺少必要字段:", data);
      return;
    }
    showNotification(data.title, data.message, data.type || "info");
  } catch (error) {
    console.error("[Notification] 处理通知失败:", error);
  }
});
```

## 🔗 相关文档

- [LIBRARY_NOTIFICATION.md](./LIBRARY_NOTIFICATION.md) - 通知功能详细说明
- [WAILS_NOTIFICATION_API_STATUS.md](./WAILS_NOTIFICATION_API_STATUS.md) - Wails v3 API 支持情况
- [LIBRARY_NOTIFICATION_QUICKREF.md](./LIBRARY_NOTIFICATION_QUICKREF.md) - 快速参考

---

**修复日期**: 2026-04-09  
**修复者**: AI Assistant  
**状态**: ✅ 已修复并准备测试
