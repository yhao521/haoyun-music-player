# 国际化功能 - 快速开始

## 🚀 5分钟快速体验

### 1. 编译应用

```bash
# 进入项目目录
cd /Users/yanghao/storage/code_projects/goProjects/haoyun-music-player

# 编译后端（自动嵌入前端资源）
go build -o haoyun-music-player .
```

### 2. 运行应用

```bash
# macOS
./haoyun-music-player

# Windows
.\haoyun-music-player.exe

# Linux
./haoyun-music-player
```

### 3. 测试语言切换

#### 步骤 1: 查看默认中文界面
- 应用启动后，点击系统托盘图标
- 所有菜单项应显示为**中文**

#### 步骤 2: 打开设置页面
- 托盘菜单 → **设置**
- 或快捷键 `Cmd+S` (macOS) / `Ctrl+S` (Windows/Linux)

#### 步骤 3: 切换到英文
- 在"通用设置"部分找到"**语言**"下拉框
- 选择 "**English**"
- 观察界面**立即**变为英文

#### 步骤 4: 验证托盘菜单更新
- 关闭设置窗口
- 再次点击托盘图标
- 所有菜单项应变为**英文**

#### 步骤 5: 切换回中文
- 重新打开设置 (Settings)
- 选择 "简体中文"
- 验证恢复为中文

### 4. 测试持久化

```bash
# 完全退出应用
# 托盘菜单 → 退出 (Quit)

# 重新启动
./haoyun-music-player

# 检查语言是否保持为上次的选择
```

---

## 📱 主要功能演示

### 托盘菜单对比

| 功能 | 中文 | English |
|------|------|---------|
| 正在播放 | 🎵 歌曲名 | 🎵 Song Name |
| 播放控制 | 播放/暂停 | Play/Pause |
| 上一曲/下一曲 | 上一曲 / 下一曲 | Previous / Next |
| 浏览歌曲 | 浏览歌曲 | Browse Songs |
| 喜爱音乐 | ❤️ 喜爱音乐 | ❤️ Favorite Songs |
| 播放模式 | 播放模式 | Play Mode |
| 音乐库 | 音乐库 | Music Library |
| 设置 | 设置 | Settings |
| 退出 | 退出 | Quit |

### 播放模式对比

| 模式 | 中文 | English |
|------|------|---------|
| 顺序 | 顺序播放 | Order Play |
| 循环 | ✓ 循环播放 | ✓ Loop Play |
| 随机 | 随机播放 | Random Play |
| 单曲 | 单曲循环 | Single Loop |

---

## 🔍 调试技巧

### 查看控制台日志

启动应用时，终端会显示：

```
✓ 国际化模块已初始化，当前语言: zh-CN
✓ 已加载语言文件: en-US.json
✓ 已加载语言文件: zh-CN.json
```

切换语言时：

```
✓ 语言已切换为: en-US
🔄 开始重建托盘菜单...
✅ 托盘菜单重建完成
```

### 前端开发者工具

```bash
# 在 main.go 中启用 DevTools
# 找到 Development 菜单 → Open DevTools

# 或在代码中添加
window.OpenDevTools()
```

在控制台测试：

```javascript
// 查看当前语言
console.log(localStorage.getItem('preferred-language'))

// 手动切换语言
window.runtime.EventsEmit("changeLanguage", "en-US")

// 测试翻译函数
import { t } from './i18n'
console.log(t('settings.title'))
```

---

## ❓ 常见问题

### Q: 切换语言后菜单没有更新？

**A**: 检查以下几点：
1. 确认看到日志 "✅ 托盘菜单重建完成"
2. 关闭并重新打开托盘菜单
3. 某些子菜单可能需要重新展开

### Q: 设置页面还是中文？

**A**: 
1. 刷新设置窗口（关闭再打开）
2. 检查浏览器控制台是否有错误
3. 确认 `frontend/src/i18n/locales/` 下有对应文件

### Q: 重启后语言重置了？

**A**: 
- 前端偏好保存在 `localStorage`
- 确保正常退出应用（不要强制杀死进程）
- 检查浏览器存储是否被清除

---

## 📚 更多文档

- **技术细节**: [I18N_IMPLEMENTATION.md](./I18N_IMPLEMENTATION.md)
- **完整测试**: [I18N_TESTING_GUIDE.md](./I18N_TESTING_GUIDE.md)
- **完成报告**: [I18N_COMPLETION_REPORT.md](./I18N_COMPLETION_REPORT.md)

---

## 🎯 下一步

想添加新语言？参考以下步骤：

1. 创建语言文件
   ```bash
   # 后端
   cp backend/pkg/i18n/en-US.json backend/pkg/i18n/ja-JP.json
   
   # 前端
   cp frontend/src/i18n/locales/en-US.json frontend/src/i18n/locales/ja-JP.json
   ```

2. 翻译 JSON 文件中的所有值

3. 前端注册新语言
   ```typescript
   // frontend/src/i18n/index.ts
   import jaJP from './locales/ja-JP.json'
   
   export type Locale = 'zh-CN' | 'en-US' | 'ja-JP'
   export const translations = {
     'zh-CN': zhCN,
     'en-US': enUS,
     'ja-JP': jaJP  // 添加这一行
   }
   ```

4. 重新编译并测试

就是这么简单！✨

---

**享受多语言音乐播放器吧！🎵🌍**
