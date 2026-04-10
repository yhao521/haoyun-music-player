# FFmpeg 测试程序

这是一个独立的测试程序，用于验证 FFmpeg 音频解码功能。

## 📁 文件位置

测试程序位于 `tests/` 目录：
```
haoyun-music-player/
└── tests/
    └── test_ffmpeg.go    # 测试程序源代码
```

## 🚀 使用方法

### macOS/Linux（推荐）
```bash
# 运行测试脚本（自动编译和清理）
./test_ffmpeg.sh
```

### Windows（推荐）
```powershell
# 运行测试脚本（自动编译和清理）
.\test_ffmpeg.bat
```

### 手动运行
```bash
# 进入 tests 目录
cd tests

# 编译并运行
go run test_ffmpeg.go
```

## ⚠️ 注意事项

### 编译错误说明

如果你在项目根目录看到 `main redeclared in this block` 错误，这是**正常现象**。

**原因**：
- [test_ffmpeg.go](file:///Users/yanghao/storage/code_projects/goProjects/haoyun-music-player/tests/test_ffmpeg.go) 是一个独立的测试程序
- 它与 [main.go](file:///Users/yanghao/storage/code_projects/goProjects/haoyun-music-player/main.go) 都有 `main` 函数
- Go 不允许同一个包中有多个 `main` 函数

**解决方案**：
1. ✅ 使用测试脚本（推荐）：`./test_ffmpeg.sh` 或 `.\test_ffmpeg.bat`
2. ✅ 进入 tests 目录运行：`cd tests && go run test_ffmpeg.go`
3. ❌ 不要在项目根目录直接运行 `go run .`

## 📋 测试内容

测试程序会执行以下检查：

1. ✅ **FFmpeg 可用性检查**
   - 检测 FFmpeg 是否已安装
   - 显示 FFmpeg 路径
   - 提供安装指引

2. ✅ **音频文件扫描**
   - 扫描当前目录下的音频文件
   - 支持格式：MP3, WAV, FLAC, AAC, M4A, OGG, WMA, APE, Opus, AIFF

3. ✅ **解码功能测试**
   - 尝试加载每个音频文件
   - 验证解码是否成功
   - 显示详细信息：
     - 采样率 (Hz)
     - 声道数
     - 时长 (秒)
     - 数据大小 (KB)

## 📊 示例输出

```
=== FFmpeg 音频解码器测试 ===

📋 测试 1: 检查 FFmpeg 可用性
✅ FFmpeg 路径: /usr/local/bin/ffmpeg

📋 测试 2: 扫描音频文件
✅ 找到 3 个音频文件:
   1. song1.mp3
   2. song2.flac
   3. song3.aac

📋 测试 3: 测试音频解码

🎵 测试文件: song1.mp3
   ✅ 解码成功
   - 采样率: 44100 Hz
   - 声道数: 2
   - 时长: 245 秒
   - 数据大小: 9800 KB

🎵 测试文件: song2.flac
   ✅ 解码成功
   - 采样率: 44100 Hz
   - 声道数: 2
   - 时长: 180 秒
   - 数据大小: 7200 KB

...

=== 测试完成 ===
```

## 🔧 故障排除

### 问题 1: 找不到 FFmpeg
```
❌ FFmpeg 未找到: 未找到 FFmpeg，请安装 FFmpeg 或设置 FFMPEG_PATH 环境变量
```

**解决**：
```bash
# macOS
brew install ffmpeg

# Ubuntu
sudo apt-get install ffmpeg

# Windows
choco install ffmpeg
```

### 问题 2: 没有音频文件
```
⚠️  当前目录没有找到音频文件
请将测试音频文件放在当前目录后重新运行
```

**解决**：将一些音频文件复制到 `tests/` 目录，然后重新运行测试

### 问题 3: 解码失败
```
❌ 解码失败: FFmpeg 转换失败：exit status 1
```

**可能原因**：
- 音频文件损坏
- FFmpeg 版本过旧
- 文件格式不受支持

**解决**：
1. 尝试用其他播放器播放该文件
2. 更新 FFmpeg 到最新版本
3. 查看详细错误日志

## 💡 提示

- 测试程序会将音频文件转换为 PCM 数据进行验证
- 大文件可能需要较长时间加载
- 建议先用小文件测试（< 10MB）
- 测试完成后会自动清理临时文件

## 📖 相关文档

- [FFmpeg 安装指南](../FFMPEG_GUIDE.md)
- [集成总结](../FFMPEG_INTEGRATION_SUMMARY.md)
- [快速参考](../FFMPEG_QUICKREF.md)

---

**最后更新**: 2026-04-07