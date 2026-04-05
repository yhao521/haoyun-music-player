# 🐛 Bug 修复记录

## 2026-04-05: 菜单空指针解引用错误修复

### 问题描述

应用程序启动时出现致命错误：
```
panic error: runtime error: invalid memory address or nil pointer dereference
github.com/wailsapp/wails/v3/pkg/application.(*Menu).processRadioGroups
```

错误发生在 Wails v3 的菜单单选按钮组处理过程中。

### 根本原因

Wails v3 (alpha.74) 在处理 `NewMenuItemCheckbox` 创建的复选框菜单项时，会尝试将其作为单选按钮组（Radio Group）进行处理。在某些情况下，这会导致空指针解引用错误。

受影响的菜单项：
1. **播放模式菜单** - 使用了 3 个复选框（顺序/循环/随机）
2. **音乐库列表菜单** - 动态生成的复选框列表

### 解决方案

将所有可能导致问题的复选框菜单项改为普通菜单项，使用 Unicode 字符 `✓` 来表示选中状态。

#### 修改前（有问题）

```go
// 播放模式 - 使用复选框
playModeOrder = application.NewMenuItemCheckbox("顺序播放", true)
playModeLoop = application.NewMenuItemCheckbox("循环播放", false)
playModeRandom = application.NewMenuItemCheckbox("随机播放", false)

// 音乐库 - 使用复选框
libItem := application.NewMenuItemCheckbox(libName, true)
```

#### 修改后（已修复）

```go
// 播放模式 - 使用普通菜单项 + ✓ 符号
playModeOrder = application.NewMenuItem("✓ 顺序播放")
playModeLoop = application.NewMenuItem("  循环播放")
playModeRandom = application.NewMenuItem("  随机播放")

// 点击时更新标签
playModeOrder.OnClick(func(ctx *application.Context) {
    musicService.SetPlayMode("order")
    playModeOrder.SetLabel("✓ 顺序播放")
    playModeLoop.SetLabel("  循环播放")
    playModeRandom.SetLabel("  随机播放")
})

// 音乐库 - 使用普通菜单项 + ✓ 符号
currentLib := musicService.GetCurrentLibrary()
currentLibName := ""
if currentLib != nil {
    currentLibName = currentLib.Name
}

for _, libName := range libraries {
    label := libName
    if libName == currentLibName {
        label = "✓ " + libName
    } else {
        label = "  " + libName
    }
    
    libItem := application.NewMenuItem(label)
    // ... 点击处理逻辑
}
```

### 技术细节

#### 1. 播放模式菜单

**实现策略**:
- 初始状态：顺序播放显示 `✓`，其他两项显示空格
- 点击任意项时：更新所有三项的标签，将 `✓` 移到被点击的项
- 视觉效果与复选框相同，但避免了 Wails 的内部 bug

**代码位置**: [main.go](file:///Users/yanghao/storage/code_projects/goProjects/haoyun-music-player/main.go#L218-L258)

#### 2. 音乐库菜单

**实现策略**:
- 获取当前选中的音乐库名称
- 遍历所有库，为当前库添加 `✓` 前缀，其他库添加空格前缀
- 点击时更新所有菜单项的标签

**依赖包**: 需要导入 `strings` 包用于字符串操作

**代码位置**: [main.go](file:///Users/yanghao/storage/code_projects/goProjects/haoyun-music-player/main.go#L413-L480)

### 测试验证

✅ 编译通过，无语法错误  
✅ 应用程序正常启动，无 panic  
✅ 托盘菜单正常显示  
✅ 播放模式切换功能正常  
✅ 音乐库切换功能正常  

### 影响范围

**修改的文件**:
- [main.go](file:///Users/yanghao/storage/code_projects/goProjects/haoyun-music-player/main.go)
  - 添加 `strings` 包导入
  - 重构播放模式菜单（~40 行）
  - 重构音乐库菜单（~70 行）

**未受影响的功能**:
- ✅ 喜爱音乐菜单（使用普通菜单项，无问题）
- ✅ 保持系统唤醒（复选框，但在主菜单中，未触发 bug）
- ✅ 开机启动（复选框，但在主菜单中，未触发 bug）

### 注意事项

1. **保留的复选框**: 
   - "保持系统唤醒"和"开机启动"仍使用复选框，因为它们在主菜单层级，不在子菜单中，不会触发 `processRadioGroups` 的处理逻辑
   
2. **未来升级**: 
   - 当 Wails v3 正式版发布后，可以重新评估是否恢复使用复选框
   - 建议关注 Wails 的 issue tracker，查看此 bug 是否已修复

3. **用户体验**: 
   - 使用 `✓` 符号的视觉效果与复选框几乎相同
   - 用户感知不到任何差异

### 相关资源

- Wails v3 Issue: 建议在 GitHub 上报告此 bug
- 替代方案: 如果未来需要真正的单选行为，可以考虑使用 `AddRadio` API（但本次修复中未使用，因为也有潜在问题）

---

<div align="center">

**Bug 已修复，应用可正常运行！** ✅

</div>
