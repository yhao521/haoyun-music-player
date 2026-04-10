# 国际化 (i18n) 实施说明

## 概述

本项目已完成国际化改造，支持中文（默认）和英文两种语言。用户可以在设置页面中切换语言。

## 架构设计

### 后端国际化 (Go)

**位置**: `backend/pkg/i18n/`

**文件结构**:
- `i18n.go` - 翻译器核心模块
- `zh-CN.json` - 中文语言文件
- `en-US.json` - 英文语言文件

**特性**:
- 单例模式，全局共享翻译器实例
- 支持嵌套键访问（如 `menu.playPause`）
- 嵌入式资源文件，编译时打包
- 线程安全的语言切换

**使用示例**:
```go
import "github.com/yhao521/wailsMusicPlay/backend/pkg/i18n"

// 获取翻译器实例
translator := i18n.GetTranslator()

// 翻译文本
text := translator.T("menu.playPause") // 返回 "播放/暂停" 或 "Play/Pause"

// 切换语言
translator.SetLocale("en-US")
```

### 前端国际化 (TypeScript/Vue)

**位置**: `frontend/src/i18n/`

**文件结构**:
- `index.ts` - 国际化工具模块
- `locales/zh-CN.json` - 中文语言文件
- `locales/en-US.json` - 英文语言文件

**特性**:
- 基于 localStorage 持久化用户偏好
- 响应式语言切换
- 支持默认值参数
- 嵌套键访问

**使用示例**:
```typescript
import { t, setLocale, getLocale } from '../i18n'

// 翻译文本
const text = t('settings.title') // 返回 "⚙️ 设置" 或 "⚙️ Settings"

// 带默认值的翻译
const mode = t('playMode.loop', '循环播放')

// 切换语言
setLocale('en-US')

// 获取当前语言
const current = getLocale()
```

## 已国际化的内容

### 1. 托盘菜单 (System Tray Menu)
- ✅ 正在播放
- ✅ 播放/暂停、上一曲、下一曲
- ✅ 浏览歌曲、喜爱音乐
- ✅ 播放模式（顺序/循环/随机/单曲）
- ✅ 音乐库管理（添加/刷新/重命名/删除）
- ✅ 下载音乐、保持唤醒、开机启动
- ✅ 设置、版本信息、退出

### 2. 主菜单栏 (Menu Bar)
- ✅ File 菜单
- ✅ Music 菜单（所有子项）
- ✅ Playback 菜单
- ✅ 其他系统菜单

### 3. 设置界面 (Settings View)
- ✅ 通用设置（开机启动、保持唤醒、语言选择、主题）
- ✅ 播放设置（默认播放模式、显示歌词、音量）
- ✅ 媒体键设置
- ✅ 关于信息

## 语言切换机制

### 前端 → 后端通信

当用户在设置页面切换语言时：

1. **前端操作**:
   ```typescript
   // SettingsView.vue
   const changeLanguage = (locale: Locale) => {
     setLocale(locale); // 更新前端语言
     window.runtime.EventsEmit("changeLanguage", locale); // 通知后端
   }
   ```

2. **后端响应**:
   ```go
   // main.go
   app.Event.On("changeLanguage", func(event *application.CustomEvent) {
     if locale, ok := event.Data.(string); ok {
       translator.SetLocale(locale) // 更新后端语言
       rebuildTrayMenu() // 重建托盘菜单以应用新语言
     }
   })
   ```

3. **菜单动态更新**:
   - `rebuildTrayMenu()` 函数会更新所有托盘菜单项的文本
   - 包括：播放控制、窗口管理、播放模式、音乐库等
   - 正在播放的歌曲名称也会根据新语言更新

### 持久化存储

- **前端**: 使用 `localStorage` 存储 `preferred-language`
- **后端**: 当前会话有效，重启后恢复默认（可扩展为配置文件）

## 添加新语言

### 后端步骤

1. 在 `backend/pkg/i18n/` 创建新语言文件（如 `ja-JP.json`）
2. 复制现有 JSON 结构并翻译所有键值
3. 重新编译即可（嵌入文件自动包含）

### 前端步骤

1. 在 `frontend/src/i18n/locales/` 创建新语言文件
2. 在 `frontend/src/i18n/index.ts` 中导入并注册：
   ```typescript
   import jaJP from './locales/ja-JP.json'
   
   export type Locale = 'zh-CN' | 'en-US' | 'ja-JP'
   
   export const translations: Record<Locale, Translation> = {
     'zh-CN': zhCN as unknown as Translation,
     'en-US': enUS as unknown as Translation,
     'ja-JP': jaJP as unknown as Translation
   }
   ```
3. 在设置页面添加语言选项

## 注意事项

### 1. ✅ 动态菜单更新已实现

语言切换后，托盘菜单会**自动更新**所有文本：
- 通过 `rebuildTrayMenu()` 函数实现
- 更新所有菜单项的 Label
- 重建音乐库子菜单
- 更新正在播放状态

### 2. 部分文本未国际化

以下位置仍有硬编码文本，建议后续国际化：
- 日志消息（`log.Println`）- 保留中文便于调试
- 错误提示
- 其他 Vue 组件（BrowseView, FavoritesView, AppMain 等）

### 3. 翻译键命名规范

- 使用小写字母和点号分隔：`menu.playPause`
- 按模块分组：`settings.*`, `playMode.*`, `common.*`
- 保持一致性：前后端使用相同的键名

## 测试

### 编译测试
```bash
# 后端
go build -o /tmp/test_build .

# 前端
cd frontend && npm run build
```

### 运行时测试
1. 启动应用
2. 打开设置页面
3. 切换语言（中文 ↔ 英文）
4. 观察：
   - 设置界面文本立即变化
   - 后端收到 `changeLanguage` 事件
   - 托盘菜单文本（待实现动态更新）

## 未来改进

1. ✨ ~~实现托盘菜单动态更新~~ **已完成**
2. ✨ 国际化其他 Vue 组件
   - BrowseView.vue
   - FavoritesView.vue
   - AppMain.vue
3. ✨ 添加更多语言（日语、韩语等）
4. ✨ 后端语言偏好持久化到配置文件
5. ✨ 自动检测系统语言作为默认值
6. ✨ 提供翻译缺失的检测和警告
7. ✨ 优化播放模式子菜单的动态更新（当前仅更新父菜单标签）

## 相关文件

- 后端: `backend/pkg/i18n/`
- 前端: `frontend/src/i18n/`
- 主入口: `main.go`, `frontend/src/main.ts`
- 设置页面: `frontend/src/views/SettingsView.vue`
