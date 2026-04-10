# 国际化功能测试指南

## 快速测试步骤

### 1. 编译并运行应用

```bash
# 后端编译
go build -o haoyun-music-player .

# 前端构建（如果需要）
cd frontend && npm run build

# 运行应用
./haoyun-music-player
```

### 2. 测试托盘菜单（默认中文）

启动应用后，点击系统托盘图标，检查以下菜单项是否显示为中文：

- ✅ 正在播放：无 / 🎵 [歌曲名]
- ✅ 播放/暂停
- ✅ 上一曲 / 下一曲
- ✅ 浏览歌曲
- ✅ ❤️ 喜爱音乐
- ✅ 播放模式 → 顺序播放 / ✓ 循环播放 / 随机播放 / 单曲循环
- ✅ 音乐库 → [库名称] / 添加新音乐库 / 刷新当前音乐库等
- ✅ 下载音乐
- ✅ 保持系统唤醒
- ✅ 开机启动
- ✅ 设置
- ✅ 版本 0.5.0
- ✅ 退出

### 3. 测试主菜单栏

检查顶部菜单栏的 "Music" 菜单：

- ✅ 所有菜单项应显示为中文
- ✅ 快捷键提示正确

### 4. 测试设置页面语言切换

1. **打开设置窗口**
   - 方法 1: 托盘菜单 → 设置
   - 方法 2: 主菜单 → Music → 设置
   - 方法 3: 快捷键 `Cmd+S` (macOS) / `Ctrl+S` (Windows/Linux)

2. **切换到英文**
   - 在"通用设置"部分找到"语言"下拉框
   - 选择 "English"
   - 观察界面文本立即变为英文

3. **验证托盘菜单更新**
   - 关闭设置窗口
   - 点击托盘图标
   - 所有菜单项应变为英文：
     - Now Playing: None
     - Play/Pause
     - Previous Track / Next Track
     - Browse Songs
     - ❤️ Favorite Songs
     - Play Mode → Order Play / ✓ Loop Play / Random Play / Single Loop
     - Music Library
     - Download Music
     - Keep System Awake
     - Auto Launch
     - Settings
     - Version 0.5.0
     - Quit

4. **切换回中文**
   - 重新打开设置窗口
   - 选择 "简体中文"
   - 验证所有文本恢复为中文

### 5. 测试持久化

1. 将语言切换为英文
2. 完全退出应用（托盘菜单 → 退出）
3. 重新启动应用
4. 检查：
   - ✅ 设置页面仍显示为英文
   - ✅ 托盘菜单仍显示为英文
   - （前端 localStorage 持久化生效）

### 6. 测试正在播放状态

1. 加载音乐库并开始播放
2. 切换语言（中文 ↔ 英文）
3. 检查托盘菜单的"正在播放"项：
   - ✅ 歌曲名称保持不变
   - ✅ 前缀文本根据语言变化（"🎵" 或保持原样）

### 7. 测试播放模式菜单

1. 打开托盘菜单
2. 查看"播放模式"子菜单
3. 切换语言
4. 再次打开托盘菜单
5. 检查：
   - ⚠️ 父菜单标签已更新（"播放模式" → "Play Mode"）
   - ⚠️ 子菜单项可能需要重新打开才能看到更新
   - （这是 Wails v3 的限制，子菜单项的动态更新较复杂）

---

## 常见问题排查

### Q1: 切换语言后托盘菜单没有更新

**可能原因**:
- 事件监听器未正确注册
- `rebuildTrayMenu()` 函数执行失败

**排查步骤**:
1. 查看控制台日志，确认是否有以下输出：
   ```
   ✓ 语言已切换为: en-US
   🔄 开始重建托盘菜单...
   ✅ 托盘菜单重建完成
   ```
2. 如果没有看到日志，检查 `main.go` 中的事件监听器是否正确注册

### Q2: 设置页面文本没有切换

**可能原因**:
- 前端 i18n 模块未正确初始化
- 语言文件路径错误

**排查步骤**:
1. 打开浏览器开发者工具（DevTools）
2. 查看控制台是否有错误
3. 确认 `frontend/src/i18n/locales/` 目录下有对应的 JSON 文件
4. 检查 `localStorage` 中是否有 `preferred-language` 键

### Q3: 编译错误

**常见错误**:
- `undefined: translator` - 确保在 `main()` 函数开头初始化了翻译器
- `undefined: t` - 确保创建了翻译辅助函数
- `cannot find module` (前端) - 检查导入路径是否正确

**解决方案**:
- 参考 `I18N_IMPLEMENTATION.md` 文档中的代码示例
- 确保所有文件都已保存
- 清理缓存后重新编译

---

## 开发者调试技巧

### 1. 查看当前语言状态

**后端**:
```go
log.Printf("当前语言: %s", translator.GetLocale())
```

**前端**:
```typescript
import { getLocale } from '../i18n'
console.log('Current locale:', getLocale())
```

### 2. 手动触发语言切换

**前端控制台**:
```javascript
window.runtime.EventsEmit("changeLanguage", "en-US")
```

### 3. 检查翻译键是否存在

**后端**:
```go
text := translator.T("menu.playPause")
log.Printf("翻译结果: %s", text)
// 如果返回键名本身，说明翻译不存在
```

**前端**:
```typescript
import { t } from '../i18n'
console.log(t('menu.playPause'))
```

### 4. 查看所有支持的语言

**后端**:
```go
locales := translator.GetSupportedLocales()
log.Printf("支持的语言: %v", locales)
```

---

## 性能测试

### 菜单重建耗时

语言切换时，`rebuildTrayMenu()` 的执行时间应在 **10-50ms** 以内。

**测试方法**:
```go
start := time.Now()
rebuildTrayMenu()
elapsed := time.Since(start)
log.Printf("菜单重建耗时: %v", elapsed)
```

如果耗时过长，考虑优化：
- 减少不必要的菜单项更新
- 使用增量更新而非完全重建
- 异步更新非关键菜单项

---

## 回归测试清单

每次修改国际化代码后，请执行以下测试：

- [ ] 应用能正常启动
- [ ] 默认语言为中文
- [ ] 托盘菜单显示正确
- [ ] 主菜单栏显示正确
- [ ] 设置页面能正常打开
- [ ] 语言切换功能正常
- [ ] 切换后托盘菜单更新
- [ ] 切换后设置界面更新
- [ ] 语言偏好持久化生效
- [ ] 正在播放状态正常更新
- [ ] 播放模式切换正常
- [ ] 音乐库菜单正常
- [ ] 无内存泄漏（长时间运行测试）
- [ ] 无崩溃或 panic

---

## 贡献指南

如需添加新语言或改进国际化功能：

1. **添加新语言文件**
   - 后端: `backend/pkg/i18n/{locale}.json`
   - 前端: `frontend/src/i18n/locales/{locale}.json`

2. **注册新语言**
   - 前端: 在 `index.ts` 中导入并添加到 `translations`

3. **翻译所有键**
   - 参考现有语言文件的结构
   - 确保所有键都存在，避免运行时回退到键名

4. **测试**
   - 按照本指南进行完整测试
   - 特别关注特殊字符和长文本的显示

5. **提交 PR**
   - 包含语言文件
   - 更新文档
   - 提供截图证明功能正常

---

**祝测试顺利！如有问题，请参考 `I18N_IMPLEMENTATION.md` 文档。**
