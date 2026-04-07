package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/yhao521/wailsMusicPlay/backend"
)

func main() {
	fmt.Println("=== FFmpeg 音频解码器测试 ===\n")

	// 测试 1: 检查 FFmpeg 是否可用
	fmt.Println("📋 测试 1: 检查 FFmpeg 可用性")
	ffmpegPath, err := backend.FindFFmpegPath()
	if err != nil {
		log.Printf("❌ FFmpeg 未找到: %v", err)
		log.Println("\n请安装 FFmpeg:")
		log.Println("  macOS: brew install ffmpeg")
		log.Println("  Ubuntu: sudo apt-get install ffmpeg")
		log.Println("  Windows: choco install ffmpeg")
		os.Exit(1)
	}
	fmt.Printf("✅ FFmpeg 路径: %s\n\n", ffmpegPath)

	// 测试 2: 列出当前目录下的音频文件
	fmt.Println("📋 测试 2: 扫描音频文件")
	audioFiles, err := findAudioFiles(".")
	if err != nil {
		log.Printf("❌ 扫描失败: %v", err)
		os.Exit(1)
	}

	if len(audioFiles) == 0 {
		fmt.Println("⚠️  当前目录没有找到音频文件")
		fmt.Println("请将测试音频文件放在当前目录后重新运行")
		os.Exit(0)
	}

	fmt.Printf("✅ 找到 %d 个音频文件:\n", len(audioFiles))
	for i, file := range audioFiles {
		fmt.Printf("   %d. %s\n", i+1, filepath.Base(file))
	}
	fmt.Println()

	// 测试 3: 尝试加载每个音频文件
	fmt.Println("📋 测试 3: 测试音频解码")
	player := backend.NewAudioPlayer()

	for _, audioFile := range audioFiles {
		fmt.Printf("\n🎵 测试文件: %s\n", filepath.Base(audioFile))
		
		reader, sampleRate, channels, err := player.LoadAudioFileForTest(audioFile)
		if err != nil {
			fmt.Printf("   ❌ 解码失败: %v\n", err)
			continue
		}
		
		fmt.Printf("   ✅ 解码成功\n")
		fmt.Printf("   - 采样率: %d Hz\n", sampleRate)
		fmt.Printf("   - 声道数: %d\n", channels)
		fmt.Printf("   - 时长: %d 秒\n", reader.Len())
		fmt.Printf("   - 数据大小: %d KB\n", getDataSize(reader)/1024)
		
		reader.Close()
	}

	fmt.Println("\n=== 测试完成 ===")
}

// findAudioFiles 查找目录下的音频文件
func findAudioFiles(dir string) ([]string, error) {
	supportedExts := map[string]bool{
		".mp3": true, ".wav": true, ".flac": true,
		".aac": true, ".m4a": true, ".ogg": true,
		".wma": true, ".ape": true, ".opus": true,
		".aiff": true, ".alac": true,
	}

	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() {
			ext := filepath.Ext(path)
			if supportedExts[ext] {
				files = append(files, path)
			}
		}
		return nil
	})
	
	return files, err
}

// getDataSize 获取数据大小（仅用于测试）
func getDataSize(reader interface{ Len() int }) int {
	// 这是一个估算值，实际大小取决于实现
	return reader.Len() * 44100 * 2 * 2 // 假设 44100Hz, 16-bit, 立体声
}