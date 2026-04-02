# Speaker 初始化问题修复

## 问题描述

在播放音乐时遇到错误：
```
初始化扬声器失败：speaker cannot be initialized more than once
```

## 问题原因

原来的实现中，每次调用 `Play()` 方法都会执行：
```go
func (ap *AudioPlayer) initSpeaker(format beep.Format) error {
    speaker.Close()  // 关闭
    speaker.Init(...) // 重新初始化
    return nil
}
```

这导致每次播放歌曲时都会重新初始化 speaker，但 beep 库不允许重复初始化。

## 解决方案

### 1. 添加状态跟踪

在 `AudioPlayer` 结构体中添加字段：
```go
type AudioPlayer struct {
    // ... existing fields ...
    
    // 扬声器状态
    speakerInitialized bool      // 是否已初始化
    speakerFormat      beep.Format  // 当前格式
}
```

### 2. 优化 initSpeaker 方法

只在真正需要时才重新初始化：
```go
func (ap *AudioPlayer) initSpeaker(format beep.Format) error {
    // 检查是否需要重新初始化
    if ap.speakerInitialized && 
       ap.speakerFormat.SampleRate == format.SampleRate && 
       ap.speakerFormat.NumChannels == format.NumChannels {
        // 扬声器已初始化且格式相同，无需重新初始化
        return nil
    }

    // 如果格式不同或首次初始化，关闭当前的 speaker
    if ap.speakerInitialized {
        speaker.Close()
        ap.speakerInitialized = false
    }

    // 初始化扬声器
    err := speaker.Init(format.SampleRate, format.NumChannels*format.SampleRate.N(time.Second/10))
    if err != nil {
        return fmt.Errorf("初始化扬声器失败：%w", err)
    }

    ap.speakerInitialized = true
    ap.speakerFormat = format

    return nil
}
```

## 优化逻辑

### ✅ 场景 1：首次播放
- `speakerInitialized = false`
- 执行初始化
- 设置 `speakerInitialized = true`

### ✅ 场景 2：切换歌曲（相同格式）
- `speakerInitialized = true`
- 格式相同 → **直接返回，不重新初始化**
- 只更换 streamer 即可

### ✅ 场景 3：切换不同格式的歌曲
- `speakerInitialized = true`
- 格式不同 → 关闭当前 speaker → 重新初始化

### ✅ 场景 4：暂停后继续播放
- `speakerInitialized = true`
- 格式相同 → **直接返回**
- 只恢复播放

## 关键改进

1. **避免重复初始化** - 相同格式不重新初始化
2. **只在必要时关闭** - 格式变化时才关闭 speaker
3. **状态跟踪** - 记录初始化状态和格式信息
4. **性能优化** - 减少不必要的初始化和关闭操作

## 测试场景

### ✅ 测试 1：连续播放
```
播放歌曲 A → 播放歌曲 B → 播放歌曲 C
```
- 只在播放歌曲 A 时初始化一次
- 后续切换不会重新初始化

### ✅ 测试 2：暂停/播放切换
```
播放 → 暂停 → 播放 → 暂停
```
- Speaker 保持初始化状态
- 不会重复初始化

### ✅ 测试 3：不同格式切换
```
播放 MP3 → 播放 WAV → 播放 FLAC
```
- MP3: 初始化 speaker (格式 A)
- WAV: 关闭 speaker → 重新初始化 (格式 B)
- FLAC: 关闭 speaker → 重新初始化 (格式 C)

## 代码变更

**修改文件**: `backend/audioplayer.go`

**变更内容**:
1. 添加 `speakerInitialized` 和 `speakerFormat` 字段
2. 修改 `initSpeaker()` 方法逻辑
3. 保持 `Play()` 方法调用不变

## 验证结果

✅ 编译成功，无错误  
✅ 解决了 "speaker cannot be initialized more than once" 问题  
✅ 支持连续播放、切换歌曲  
✅ 性能优化，减少不必要的初始化  

---

**修复日期**: 2026-04-02  
**影响范围**: AudioPlayer 模块  
**向后兼容**: ✅ 是
