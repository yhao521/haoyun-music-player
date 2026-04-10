# 📚 Haoyun Music Player 文档中心

欢迎来到 Haoyun Music Player 的完整文档中心！这里包含了项目的所有技术文档、使用指南和开发资料。

---

## 🚀 快速导航

### 新用户必读
- [快速开始指南](./QUICKSTART.md) - 5 分钟上手
- [功能总览](./FEATURES.md) - 了解所有功能
- [键盘快捷键](./KEYBOARD_SHORTCUTS.md) - 提高效率
- [故障排除](./TROUBLESHOOTING.md) - 常见问题解决

### 开发者必读
- [实现文档](./IMPLEMENTATION.md) - 功能实现细节
- [后端设计](./BACKEND_DESIGN.md) - 架构设计详解
- [代码结构](./CODE_STRUCTURE.md) - 项目组织方式
- [API 指南](./API_GUIDE.md) - 后端 API 使用说明

---

## 📖 文档分类

### 🎵 核心功能

#### 音乐播放
- [音乐信息显示](./MUSIC_INFO_DISPLAY.md)
- [正在播放功能](./NOW_PLAYING_FEATURE.md)
- [媒体键支持](./MEDIA_KEYS_GUIDE.md)
- [音频时长显示](./AUDIO_DURATION_FEATURE.md)
- [扬声器修复](./SPEAKER_FIX.md)

#### 音乐库管理
- [音乐库元数据扫描](./LIBRARY_METADATA_SCAN.md)
- [元数据缓存优化](./METADATA_CACHE_OPTIMIZATION.md)
- [元数据实现总结](./METADATA_IMPLEMENTATION_SUMMARY.md)
- [元数据使用指南](./METADATA_USAGE_GUIDE.md)
- [播放列表元数据集成](./PLAYLIST_METADATA_INTEGRATION.md)
- [库通知功能](./LIBRARY_NOTIFICATION.md)
- [库通知总结](./LIBRARY_NOTIFICATION_SUMMARY.md)

#### 收藏与设置
- [收藏功能](./FAVORITES_FEATURE.md)
- [收藏窗口总结](./FAVORITES_WINDOW_SUMMARY.md)
- [设置功能](./SETTINGS_FEATURE.md)

---

### 📝 歌词功能

#### 基础功能
- [歌词下载功能](./LYRICS_DOWNLOAD_FEATURE.md)
- [歌词下载快速参考](./LYRICS_DOWNLOAD_QUICKREF.md)
- [多源歌词增强](./MULTI_SOURCE_LYRICS_ENHANCEMENT.md)

#### 高级优化
- [lrclib.net 增强完成](./LRCLIB_ENHANCEMENT_COMPLETE.md) - 智能搜索和缓存
- [模糊匹配优化完成](./FUZZY_MATCHING_OPTIMIZATION_COMPLETE.md) - Levenshtein Distance 算法
- [歌词 API 评估](./LYRICS_API_EVALUATION.md)

#### 许可证相关
- [music-lib 替代方案](./MUSICLIB_ALTERNATIVES.md)
- [移除 music-lib 可行性分析](./REMOVE_MUSICLIB_FEASIBILITY.md)
- [许可证变更完成](./LICENSE_CHANGE_COMPLETE.md)

---

### 🔊 音频解码

- [FFmpeg 音频解码](./FFMPEG_GUIDE.md) - 广泛的音频格式支持 ⭐
- [FFmpeg 集成总结](./FFMPEG_INTEGRATION_SUMMARY.md)
- [FFmpeg 快速参考](./FFMPEG_QUICKREF.md)
- [MP3 时长修复](./MP3_DURATION_FIX.md)
- [测试 FFmpeg](./TEST_FFMPEG_README.md)

---

### 🌍 国际化 (i18n)

- [国际化实施](./I18N_IMPLEMENTATION.md)
- [国际化快速开始](./I18N_QUICKSTART.md)
- [国际化完成报告](./I18N_COMPLETION_REPORT.md)
- [国际化测试指南](./I18N_TESTING_GUIDE.md)

---

### 🛠️ 开发与架构

#### 架构设计
- [后端设计](./BACKEND_DESIGN.md)
- [代码结构](./CODE_STRUCTURE.md)
- [Wails 绑定](./WAILS_BINDINGS.md) - 前后端通信机制
- [重构总结](./REFACTORING_SUMMARY.md)

#### 配置管理
- [配置持久化](./CONFIG_PERSISTENCE.md)
- [配置加载修复](./CONFIG_LOADING_FIX.md)

#### 性能优化
- [运行时内存优化](./RUNTIME_MEMORY_OPTIMIZATION.md)
- [运行时内存优化快速参考](./RUNTIME_MEMORY_OPTIMIZATION_QUICKREF.md)
- [元数据缓存优化](./METADATA_CACHE_OPTIMIZATION.md)

---

### 🚀 CI/CD 与部署

#### GitHub Actions
- [GitHub Actions 快速开始](./GITHUB_ACTIONS_QUICKSTART.md) ⭐
- [GitHub Actions 详细指南](./GITHUB_ACTIONS_RELEASE.md)
- [GitHub Actions 工作流程](./GITHUB_ACTIONS_WORKFLOW.md)
- [GitHub Actions 总结](./GITHUB_ACTIONS_SUMMARY.md)
- [GitHub Actions 索引](./GITHUB_ACTIONS_INDEX.md)
- [部署检查清单](./GITHUB_ACTIONS_CHECKLIST.md)

#### GitHub Pages
- [GitHub Pages 指南](./GITHUB_PAGES_GUIDE.md)
- [GitHub Pages 总结](./GITHUB_PAGES_SUMMARY.md)
- [网站部署](./WEBSITE.md)

#### 平台特定
- [Linux CI 修复](./LINUX_CI_PKGCONFIG_FIX.md)

---

### 🔧 Bug 修复与维护

- [Bug 修复记录](./BUGFIX_MENU_CRASH.md) - 菜单空指针错误修复
- [托盘修复说明](./TRAY_FIX.md)
- [托盘菜单时序修复](./TRAY_MENU_TIMING_FIX.md)
- [托盘菜单更新修复](./TRAY_MENU_UPDATE_FIX.md)
- [托盘更新](./TRAY_UPDATE.md)
- [通知调试指南](./NOTIFICATION_DEBUG_GUIDE.md)
- [Wails 通知 API 状态](./WAILS_NOTIFICATION_API_STATUS.md)

---

### 📋 其他资源

- [新功能实现](./NEW_FEATURES.md) - 播放历史、歌词、专辑封面
- [新功能快速参考](./QUICK_REFERENCE_NEW_FEATURES.md)
- [快速参考](./QUICK_REFERENCE.md)
- [依赖安装验证](./DEPENDENCY_INSTALL_VERIFICATION.md)
- [依赖安装状态](./DEPENDENCY_INSTALL_STATUS.md)
- [依赖安装快速参考](./QUICKREF_DEPENDENCY_INSTALL.md)
- [自动依赖安装](./DEPENDENCY_AUTO_INSTALL.md)
- [测试收藏功能](./TEST_FAVORITES.md)

---

## 📊 文档统计

- **总文档数**: 81 个
- **主要分类**: 10+ 个
- **最近更新**: 2026-04-10

---

## 💡 使用建议

### 对于用户
1. 从 [快速开始](./QUICKSTART.md) 了解基本用法
2. 查看 [功能总览](./FEATURES.md) 探索所有功能
3. 记住 [键盘快捷键](./KEYBOARD_SHORTCUTS.md) 提高效率
4. 遇到问题时查阅 [故障排除](./TROUBLESHOOTING.md)

### 对于开发者
1. 阅读 [实现文档](./IMPLEMENTATION.md) 了解架构
2. 查看 [后端设计](./BACKEND_DESIGN.md) 理解设计理念
3. 参考 [API 指南](./API_GUIDE.md) 学习接口使用
4. 遵循 [代码结构](./CODE_STRUCTURE.md) 保持代码规范

### 对于贡献者
1. 了解 [CI/CD 流程](./GITHUB_ACTIONS_QUICKSTART.md)
2. 查看 [Bug 修复记录](./BUGFIX_MENU_CRASH.md) 学习经验
3. 遵循项目规范和最佳实践

---

## 🔗 相关链接

- [🏠 项目主页](../README.md)
- [🌐 官方网站](https://yhao521.github.io/haoyun-music-player/)
- [💬 Issues](https://github.com/yhao521/haoyun-music-player/issues)
- [📦 Releases](https://github.com/yhao521/haoyun-music-player/releases)

---

<div align="center">

**文档维护者**: YHao521  
**最后更新**: 2026-04-10  
**许可证**: Apache 2.0

📚 Happy Reading!

</div>
