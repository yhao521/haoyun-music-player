# FFmpeg 快速参考

## 🚀 一分钟上手

### 1. 安装 FFmpeg

```bash
# macOS
brew install ffmpeg

# Ubuntu/Debian  
sudo apt-get install ffmpeg

# Windows (Chocolatey)
choco install ffmpeg
```

### 2. 验证安装

```bash
ffmpeg -version
```

### 3. 运行测试

```bash
# macOS/Linux
./test_ffmpeg.sh

# Windows
.\test_ffmpeg.bat
```

## 📁 支持格式

| 类型 | 格式 | 解码器 |
|------|------|--------|
| **原生** | MP3, WAV, FLAC | Go 库（高性能） |
| **FFmpeg** | AAC, M4A, OGG, WMA, APE, Opus, AIFF + 更多 | FFmpeg |

## 🔧 常见问题

### Q: 提示 "未找到 FFmpeg"？
**A**: 安装 FFmpeg 并确保在系统 PATH 中

### Q: 如何指定自定义 FFmpeg 路径？
**A**: 设置环境变量 `FFMPEG_PATH=/path/to/ffmpeg`

### Q: 为什么有些文件播放慢？
**A**: FFmpeg 需要转换时间，建议使用原生格式（MP3/WAV/FLAC）

### Q: 大文件占用内存多？
**A**: FFmpeg 会加载整个文件到内存，这是正常现象

## 📖 详细文档

- [完整安装指南](./FFMPEG_GUIDE.md)
- [集成总结](./FFMPEG_INTEGRATION_SUMMARY.md)
- [测试说明](./TEST_FFMPEG_README.md)

## 💡 最佳实践

✅ 优先使用 MP3/WAV/FLAC 格式（性能最优）  
✅ 保持 FFmpeg 为最新版本  
✅ 生产环境预装 FFmpeg  
✅ 提供友好的错误提示  

---

**更多信息**: 查看 [README.md](./README.md) 中的 FFmpeg 章节