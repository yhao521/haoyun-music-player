# FFmpeg 音频解码支持

本项目已集成 FFmpeg，支持几乎所有常见音频格式的播放。

## 🎵 支持的音频格式

### 原生解码（优先使用）
- **MP3** - 标准 Layer III 格式
- **WAV** - 未压缩 PCM 音频
- **FLAC** - 无损压缩音频

### FFmpeg 解码（自动降级）
- **AAC** - Advanced Audio Coding
- **M4A/ALAC** - Apple Lossless
- **OGG Vorbis** - 开源有损压缩
- **WMA** - Windows Media Audio
- **APE** - Monkey's Audio
- **Opus** - 低延迟音频编解码器
- **AIFF** - Audio Interchange File Format
- **其他 FFmpeg 支持的格式...**

## 📦 安装 FFmpeg

### macOS

```bash
# 使用 Homebrew 安装
brew install ffmpeg

# 验证安装
ffmpeg -version
```

### Linux

#### Ubuntu/Debian
```bash
sudo apt-get update
sudo apt-get install ffmpeg

# 验证安装
ffmpeg -version
```

#### Fedora/RHEL
```bash
sudo dnf install ffmpeg
# 或
sudo yum install ffmpeg
```

#### Arch Linux
```bash
sudo pacman -S ffmpeg
```

### Windows

#### 方法 1：使用 Chocolatey（推荐）
```powershell
# 以管理员身份运行 PowerShell
choco install ffmpeg
```

#### 方法 2：手动安装
1. 访问 [FFmpeg 官网](https://ffmpeg.org/download.html)
2. 下载 Windows 版本（建议选择 gyan.dev 或 BtbN 构建）
3. 解压到 `C:\ffmpeg`
4. 将 `C:\ffmpeg\bin` 添加到系统 PATH 环境变量

#### 方法 3：使用 Scoop
```powershell
scoop install ffmpeg
```

## 🔧 配置 FFmpeg 路径

如果 FFmpeg 不在系统 PATH 中，可以通过环境变量指定：

### macOS/Linux
```bash
export FFMPEG_PATH=/usr/local/bin/ffmpeg
# 或添加到 ~/.bashrc / ~/.zshrc
echo 'export FFMPEG_PATH=/opt/homebrew/bin/ffmpeg' >> ~/.zshrc
source ~/.zshrc
```

### Windows
```powershell
# 临时设置（当前会话）
$env:FFMPEG_PATH = "C:\ffmpeg\bin\ffmpeg.exe"

# 永久设置（系统环境变量）
[System.Environment]::SetEnvironmentVariable('FFMPEG_PATH', 'C:\ffmpeg\bin\ffmpeg.exe', 'Machine')
```

## ✅ 验证安装

运行以下命令检查 FFmpeg 是否正确安装：

```bash
ffmpeg -version
```

应该看到类似输出：
```
ffmpeg version 6.x.x Copyright (c) 2000-2024 the FFmpeg developers
built with ...
configuration: ...
libavutil      58.x.x / 58.x.x
libavcodec     60.x.x / 60.x.x
...
```

测试音频转换：
```bash
# 测试将一个音频文件转换为 PCM
ffmpeg -i test.mp3 -f s16le -acodec pcm_s16le -ar 44100 -ac 2 -vn output.pcm
```

## 🚀 工作原理

### 解码策略
1. **优先使用原生 Go 解码器**（MP3/WAV/FLAC）
   - 零外部依赖
   - 更快的加载速度
   - 更低的内存占用

2. **自动降级到 FFmpeg**
   - 当原生解码失败时
   - 处理不支持的格式时
   - 透明的用户体验

### 技术实现
```go
// 伪代码示例
streamer, err := NewFFmpegStreamer("song.aac")
// FFmpeg 内部执行：
// ffmpeg -i song.aac -f s16le -acodec pcm_s16le -ar 44100 -ac 2 -vn pipe:1
// 输出：16-bit LE PCM 数据流
```

### 音频参数
- **采样率**: 44100 Hz（标准 CD 音质）
- **位深度**: 16-bit
- **声道**: 立体声（2 声道）
- **格式**: Signed Little-Endian PCM

## ⚠️ 注意事项

### 性能考虑
- FFmpeg 会将整个音频文件加载到内存
- 大文件（>100MB）可能占用较多 RAM
- 原生解码器无此限制

### 兼容性
- 确保 FFmpeg 版本 >= 4.0（推荐最新稳定版）
- 某些 DRM 保护的音频文件无法播放
- 网络流媒体需要特殊处理（当前不支持）

### 故障排除

#### 问题 1：找不到 FFmpeg
```
错误：未找到 FFmpeg，请安装 FFmpeg 或设置 FFMPEG_PATH 环境变量
```

**解决方案**：
1. 确认 FFmpeg 已正确安装
2. 检查是否在系统 PATH 中：`which ffmpeg` (macOS/Linux) 或 `where ffmpeg` (Windows)
3. 设置 `FFMPEG_PATH` 环境变量

#### 问题 2：解码失败
```
错误：FFmpeg 转换失败：exit status 1
```

**解决方案**：
1. 检查音频文件是否损坏
2. 尝试用其他播放器播放该文件
3. 查看完整错误日志获取详细信息

#### 问题 3：权限问题（Linux）
```
错误：permission denied
```

**解决方案**：
```bash
sudo chmod +x /usr/bin/ffmpeg
```

## 📝 开发说明

### 添加新格式支持
只需在 `supportedFormats` map 中添加扩展名即可：

```go
supportedFormats := map[string]bool{
    ".mp3":  true,
    ".wav":  true,
    // ... 现有格式
    ".newformat": true,  // 添加新格式
}
```

### 自定义 FFmpeg 参数
修改 `NewFFmpegStreamer` 函数中的 `exec.Command` 调用：

```go
cmd := exec.Command(ffmpegPath,
    "-i", filePath,
    "-f", "s16le",        // 输出格式
    "-acodec", "pcm_s16le",
    "-ar", "48000",       // 改为 48kHz
    "-ac", "1",           // 改为单声道
    // ... 其他参数
)
```

## 🔗 相关资源

- [FFmpeg 官方网站](https://ffmpeg.org/)
- [FFmpeg 文档](https://ffmpeg.org/documentation.html)
- [支持的编解码器列表](https://ffmpeg.org/general.html#Codec-Options)
- [oto 音频库](https://github.com/ebitengine/oto)

## 💡 最佳实践

1. **始终安装最新稳定版 FFmpeg**
2. **优先使用原生格式**（MP3/WAV/FLAC）以获得最佳性能
3. **定期检查 FFmpeg 更新**以获取新的编解码器支持
4. **在生产环境中预装 FFmpeg**，避免运行时错误
5. **提供友好的错误提示**，引导用户安装 FFmpeg

---

**最后更新**: 2026-04-07  
**维护者**: Haoyun Music Player Team