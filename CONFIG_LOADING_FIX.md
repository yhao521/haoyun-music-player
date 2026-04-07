# 配置加载问题修复说明

## 🐛 问题描述

用户反馈：通过设置界面将语言切换为英文后，界面确实显示为英文，但 `config.json` 文件中的 `language` 字段仍然是 `"zh-CN"`。

## 🔍 根本原因

1. **前端未加载后端配置**：设置页面初始化时，加载配置的代码被注释掉了，导致前端始终使用默认值
2. **缺少配置同步机制**：前端没有从后端获取当前配置的事件处理
3. **语言切换逻辑问题**：使用了 `watch` 监听 `currentLanguage`，可能导致循环触发

## ✅ 解决方案

### 1. 前端修改

#### 添加配置加载逻辑

```typescript
onMounted(() => {
  // 从后端加载配置
  if (window.runtime && window.runtime.EventsEmit && window.runtime.EventsOn) {
    // 设置监听器接收配置响应
    window.runtime.EventsOn("getSettings:response", (response: any) => {
      if (response) {
        // 应用加载的配置到前端状态
        settings.value = {
          autoLaunch: response.autoLaunch ?? settings.value.autoLaunch,
          keepAwake: response.keepAwake ?? settings.value.keepAwake,
          theme: response.theme ?? settings.value.theme,
          // ... 其他配置项
        };
        
        // 应用语言设置
        if (response.language) {
          const locale = response.language as Locale;
          setLocale(locale);
          currentLanguage.value = locale;
        }
      }
    });
    
    // 请求配置
    window.runtime.EventsEmit("getSettings", {});
  }
});
```

#### 修改语言切换函数

```typescript
// 从 watch 改为显式函数调用
const changeLanguage = (locale: Locale) => {
  // 更新前端语言
  setLocale(locale);
  currentLanguage.value = locale;
  
  // 通知后端切换语言并保存配置
  if (window.runtime && window.runtime.EventsEmit) {
    window.runtime.EventsEmit("changeLanguage", locale);
  }
  
  // 显示重启提示
  showRestartTip.value = true;
  restartMessage.value = t('settings.languageChangedTip');
  
  // 5秒后自动隐藏
  setTimeout(() => {
    showRestartTip.value = false;
  }, 5000);
};
```

### 2. 后端修改

#### 添加获取配置事件处理

```go
// 监听获取设置事件
app.Event.On("getSettings", func(event *application.CustomEvent) {
    // 获取当前配置
    cfg := configManager.Get()
    
    // 发送配置给前端
    app.Event.Emit("getSettings:response", map[string]interface{}{
        "language":        cfg.Language,
        "theme":           cfg.Theme,
        "autoLaunch":      cfg.AutoLaunch,
        "keepAwake":       cfg.KeepAwake,
        "defaultVolume":   cfg.DefaultVolume,
        "showLyrics":      cfg.ShowLyrics,
        "enableMediaKeys": cfg.EnableMediaKeys,
        "defaultPlayMode": cfg.DefaultPlayMode,
    })
    
    log.Println("📤 已发送配置到前端")
})
```

## 📊 工作流程

### 启动时配置加载流程

```
前端 SettingsView 挂载
    ↓
onMounted 执行
    ↓
设置 getSettings:response 监听器
    ↓
发送 getSettings 请求事件
    ↓
后端接收 getSettings 事件
    ↓
configManager.Get() 读取配置
    ↓
发送 getSettings:response 响应事件
    ↓
前端接收响应
    ↓
应用配置到 settings.value
    ↓
应用语言设置 (setLocale)
    ↓
界面显示正确的语言 ✅
```

### 语言切换流程

```
用户选择新语言
    ↓
changeLanguage(locale) 被调用
    ↓
setLocale(locale) - 前端立即更新
    ↓
currentLanguage.value = locale
    ↓
EventsEmit("changeLanguage", locale)
    ↓
后端接收事件
    ↓
configManager.SetLanguage(locale)
    ├─ translator.SetLocale(locale)
    ├─ 保存到 config.json
    └─ rebuildTrayMenu()
    ↓
发送 languageChanged 事件给前端
    ↓
前端显示重启提示
```

## 🧪 测试验证

### 1. 手动修改配置文件测试

```bash
# 修改配置文件为英文
cat > ~/Library/Application\ Support/haoyun-music-player/config.json << 'EOF'
{
  "language": "en-US",
  "theme": "auto",
  "autoLaunch": false,
  "keepAwake": true,
  "defaultVolume": 80,
  "showLyrics": true,
  "enableMediaKeys": true,
  "defaultPlayMode": "loop"
}
EOF

# 启动应用
./haoyun-music-player

# 打开设置页面
# 预期：界面显示英文，Language 下拉框显示 "English"
```

### 2. 界面切换语言测试

```bash
# 1. 启动应用（配置文件为 en-US）
./haoyun-music-player

# 2. 打开设置页面
# 预期：界面显示英文

# 3. 切换语言为中文
# 选择 "简体中文"

# 4. 检查配置文件
cat ~/Library/Application\ Support/haoyun-music-player/config.json
# 预期：language 字段变为 "zh-CN"

# 5. 重启应用
# 预期：界面显示中文
```

### 3. 控制台日志验证

启动应用并打开设置页面时，应该看到：

```
[SettingsView] 设置页面已加载
从后端加载配置: {language: "en-US", theme: "auto", ...}
✓ 已应用语言设置: en-US
```

切换语言时，应该看到：

```
✓ Language changed to: zh-CN
📤 已发送配置到前端（后端日志）
✓ 语言已切换为: zh-CN (已保存到配置文件)
🔄 开始重建托盘菜单...
✅ 托盘菜单重建完成
```

## 📝 修改的文件

| 文件 | 修改内容 |
|------|---------|
| [frontend/src/views/SettingsView.vue](file:///Users/yanghao/storage/code_projects/goProjects/haoyun-music-player/frontend/src/views/SettingsView.vue) | - 移除 `watch(currentLanguage)` 改为显式函数<br>- 添加 `onMounted` 加载配置逻辑<br>- 使用事件请求-响应模式 |
| [main.go](file:///Users/yanghao/storage/code_projects/goProjects/haoyun-music-player/main.go) | - 添加 `getSettings` 事件监听器<br>- 返回当前配置给前端 |

## ⚠️ 注意事项

1. **事件通信模式**：Wails v3 使用事件系统进行前后端通信，采用请求-响应模式需要：
   - 先设置监听器（[EventsOn](file:///Users/yanghao/go/pkg/mod/github.com/wailsapp/wails/v2@v2.12.0/pkg/runtime/events.go#L7-L10)）
   - 再发送请求（[EventsEmit](file:///Users/yanghao/go/pkg/mod/github.com/wailsapp/wails/v2@v2.12.0/pkg/runtime/events.go#L3-L5)）
   - 监听器会接收响应事件

2. **配置加载时机**：配置在设置页面 [onMounted](file:///Users/yanghao/storage/code_projects/goProjects/haoyun-music-player/frontend/node_modules/@vue/runtime-core/dist/runtime-core.d.ts#L749-L749) 时加载，如果其他页面也需要配置，需要：
   - 在 [AppMain.vue](file:///Users/yanghao/storage/code_projects/goProjects/haoyun-music-player/frontend/src/AppMain.vue) 或根组件加载
   - 或使用全局状态管理（如 Pinia）

3. **类型安全**：所有配置项都使用了 TypeScript 类型断言，确保类型正确

## 🎯 后续优化建议

1. **全局配置状态**：使用 Pinia 或 Vuex 管理配置状态，避免每个页面重复加载
2. **配置缓存**：前端缓存配置，减少后端请求
3. **实时同步**：配置文件被外部修改时，自动重新加载
4. **错误处理**：添加更完善的错误处理和用户提示
5. **配置验证**：后端验证配置值的合法性

## ✅ 验证清单

- [x] 配置文件 `config.json` 正确读取
- [x] 前端启动时加载配置
- [x] 语言设置正确应用到界面
- [x] 语言切换后保存到配置文件
- [x] 重启后配置保持
- [x] 控制台无错误日志
- [x] 前端编译通过
- [x] 后端编译通过

---

**问题已修复！现在配置文件的语言设置会和界面保持一致。** 🎉
