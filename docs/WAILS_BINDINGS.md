# Wails v3 代码生成说明

## 📦 自动生成机制

Wails v3 会自动扫描 Go 后端中绑定的 Service，并为每个公开方法生成对应的 TypeScript/JavaScript 绑定。

### 生成位置

```
frontend/bindings/
├── github.com/yhao521/wailsMusicPlay/           # 模块名称（来自 go.mod）
│   ├── greetservice.ts # GreeterService 的绑定
│   └── index.ts        # 导出索引
└── github.com/         # 第三方库绑定（如果有）
```

## 🔧 MusicService 的绑定

当你在 `main.go` 中绑定 `MusicService`：

```go
app := application.New(application.Options{
    Services: []application.Service{
        application.NewService(&MusicService{}),
    },
})
```

Wails 会自动生成以下前端绑定：

### 生成的 TypeScript 代码

运行 `wails3 dev` 后，会在 `frontend/bindings/github.com/yhao521/wailsMusicPlay/` 目录下生成：

```typescript
// frontend/bindings/github.com/yhao521/wailsMusicPlay/musicservice.ts (自动生成)
export function TogglePlayPause(): Promise<boolean> {
  return $Call.ByID(1234567890);
}

export function Play(): Promise<void> {
  return $Call.ByID(1234567891);
}

export function Pause(): Promise<void> {
  return $Call.ByID(1234567892);
}

// ... 其他方法
```

## 📝 在前端使用

### 方式一：直接调用（推荐）

```typescript
import { MusicService } from "../bindings/github.com/yhao521/wailsMusicPlay";

// 播放音乐
await MusicService.Play();

// 切换播放状态
const isPlaying = await MusicService.TogglePlayPause();
```

### 方式二：通过 window.go.main（当前使用）

```typescript
// 在 Vue 组件中
const togglePlayPause = async () => {
  const result = await window.go.main.MusicService.TogglePlayPause();
  isPlaying.value = result;
};
```

## 🎯 系统托盘功能

### Go 后端实现

```go
func createSystemTray(app *application.App, musicService *MusicService) {
    tray := app.SystemTray.New()

    // 添加菜单项
    tray.AddText("播放/暂停", func(ctx *application.Context) {
        musicService.TogglePlayPause()
    })

    tray.AddText("下一首", func(ctx *application.Context) {
        musicService.Next()
    })

    // 双击托盘图标
    tray.OnLeftDoubleClick(func(ctx *application.Context) {
        mainWindow.Show()
        mainWindow.Focus()
    })
}
```

### 特点

- ✅ **无需前端代码** - 系统托盘完全由 Go 后端控制
- ✅ **原生体验** - 使用操作系统原生托盘 API
- ✅ **自动绑定** - 托盘菜单项直接调用 Go 方法

## 🔄 事件通信

### Go 端发送事件

```go
// 在 MusicService 中
if m.app != nil {
    m.app.Event.Emit("playbackStateChanged", "playing")
}
```

### 前端接收事件

```typescript
import { EventsOn, EventsOff } from "@wailsio/runtime";

onMounted(() => {
  EventsOn("playbackStateChanged", (state: string) => {
    console.log("状态变化:", state);
  });
});

onUnmounted(() => {
  EventsOff("playbackStateChanged");
});
```

## 📋 完整的调用流程

```
用户点击播放按钮
    ↓
Vue 组件调用 window.go.main.MusicService.Play()
    ↓
Wails 运行时转发到 Go 后端 MusicService.Play()
    ↓
执行播放逻辑
    ↓
发送事件 m.app.Event.Emit("playbackStateChanged", "playing")
    ↓
前端 EventsOn 接收事件并更新 UI
```

## ⚙️ 配置选项

### wails.json

```json
{
  "name": "haoyun-music-player",
  "outputfilename": "Haoyun Music Player",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "frontend:dev:watcher": "npm run dev",
  "frontend:dev:serverUrl": "auto"
}
```

### 绑定生成配置

Wails 3 默认会：

- 扫描所有注册的 Service
- 为所有公开方法（大写开头）生成绑定
- 支持复杂类型和结构体

## 🛠️ 开发技巧

### 1. 查看生成的绑定

```bash
# 运行开发模式，Wails 会自动生成绑定
wails3 dev

# 查看生成的文件
cat frontend/bindings/github.com/yhao521/wailsMusicPlay/*.ts
```

### 2. 自定义模块名称

修改 `go.mod` 中的模块名：

```go
module mymusicplayer  // 改这里

go 1.25
```

生成的绑定会变成：

```
frontend/bindings/mymusicplayer/
```

### 3. 排除不需要的方法

使用小写开头或注释：

```go
type MusicService struct {
    // 私有方法，不会生成绑定
    internalMethod() {}

    // 公开方法，会生成绑定
    PublicMethod() {}
}
```

## 📊 性能优势

- **零配置** - 无需手动编写 API 接口
- **类型安全** - TypeScript 自动推断类型
- **高性能** - 直接的二进制调用，无 HTTP 开销
- **热重载** - 修改 Go 代码后自动重新生成

## 🎨 最佳实践

### 1. 服务设计

```go
type MusicService struct {
    ctx context.Context
    app *application.App
    // 业务逻辑字段
}

// 生命周期方法
func (m *MusicService) Startup(ctx context.Context) {
    m.ctx = ctx
}

// 公开方法 - 会生成前端绑定
func (m *MusicService) Play() error {
    // 实现
}

// 私有方法 - 不会生成绑定
func (m *MusicService) loadFile(path string) error {
    // 内部实现
}
```

### 2. 错误处理

```go
// Go 端
func (m *MusicService) Play() error {
    if err := m.player.Play(); err != nil {
        return fmt.Errorf("播放失败：%w", err)
    }
    return nil
}

// 前端
try {
    await window.go.main.MusicService.Play()
} catch (error) {
    console.error('播放失败:', error)
}
```

### 3. 事件命名规范

```go
// 使用驼峰命名
app.Event.Emit("playbackStateChanged", data)
app.Event.Emit("playlistUpdated", playlist)

// 避免使用横杠
// ❌ app.Event.Emit("playback-state-changed", data)
```

## 🔍 调试技巧

### 查看绑定是否生成

```bash
# 检查绑定文件
ls -la frontend/bindings/github.com/yhao521/wailsMusicPlay/

# 查看内容
cat frontend/bindings/github.com/yhao521/wailsMusicPlay/musicservice.ts
```

### 测试调用

```javascript
// 在浏览器控制台测试
window.go.main.MusicService.IsPlaying()
  .then((result) => console.log("Is playing:", result))
  .catch((err) => console.error(err));
```

## 📚 相关资源

- [Wails v3 官方文档](https://wails.io/)
- [代码生成机制](https://wails.io/docs/reference/binding)
- [系统托盘 API](https://wails.io/docs/reference/system-tray)
- [事件系统](https://wails.io/docs/reference/events)

---

**提示**: Wails v3 处于 Alpha 阶段，绑定生成机制可能会随版本更新而变化。
