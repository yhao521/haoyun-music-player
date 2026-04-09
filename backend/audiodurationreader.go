package backend

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	mp3 "github.com/hajimehoshi/go-mp3"
	"github.com/mewkiz/flac"
)

// AudioDurationReader 音频时长读取器
type AudioDurationReader struct {
	mu    sync.RWMutex
	cache map[string]int64 // 缓存：文件路径 -> 时长（秒）
}

// NewAudioDurationReader 创建音频时长读取器
func NewAudioDurationReader() *AudioDurationReader {
	return &AudioDurationReader{
		cache: make(map[string]int64),
	}
}

// GetDuration 获取音频文件时长（秒）
func (adr *AudioDurationReader) GetDuration(filePath string) (int64, error) {
	adr.mu.RLock()
	// 检查缓存
	if duration, ok := adr.cache[filePath]; ok {
		adr.mu.RUnlock()
		return duration, nil
	}
	adr.mu.RUnlock()

	// 读取时长
	duration, err := adr.readDuration(filePath)
	if err != nil {
		log.Printf("⚠️ 读取时长失败 %s：%v", filePath, err)
		return 0, err
	}

	// 缓存结果
	adr.mu.Lock()
	adr.cache[filePath] = duration
	adr.mu.Unlock()

	return duration, nil
}

// readDuration 从音频文件中读取时长
func (adr *AudioDurationReader) readDuration(filePath string) (int64, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".mp3":
		return adr.readMP3Duration(filePath)
	case ".flac":
		return adr.readFLACDuration(filePath)
	case ".wav":
		return adr.readWAVDuration(filePath)
	default:
		return 0, fmt.Errorf("不支持的音频格式：%s", ext)
	}
}

// readMP3Duration 读取 MP3 文件时长
func (adr *AudioDurationReader) readMP3Duration(filePath string) (int64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("打开文件失败：%w", err)
	}
	defer file.Close()

	// 使用 go-mp3 解码器
	decoder, err := mp3.NewDecoder(file)
	if err != nil {
		return 0, fmt.Errorf("MP3 解码失败：%w", err)
	}

	// 获取总采样数
	totalSamples := decoder.Length()
	sampleRate := decoder.SampleRate()

	if sampleRate == 0 {
		return 0, fmt.Errorf("无效的采样率")
	}

	// 计算时长（秒）
	duration := int64(totalSamples) / int64(sampleRate)

	return duration, nil
}

// readFLACDuration 读取 FLAC 文件时长
func (adr *AudioDurationReader) readFLACDuration(filePath string) (int64, error) {
	// 使用 flac 库解析文件
	stream, err := flac.ParseFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("FLAC 解析失败：%w", err)
	}
	defer stream.Close()

	// 从 STREAMINFO 中获取信息
	if stream.Info == nil {
		return 0, fmt.Errorf("FLAC 文件缺少 STREAMINFO")
	}

	// 获取总采样数和采样率
	totalSamples := stream.Info.NSamples
	sampleRate := stream.Info.SampleRate

	if sampleRate == 0 {
		return 0, fmt.Errorf("无效的采样率")
	}

	// 计算时长（秒）
	duration := int64(totalSamples) / int64(sampleRate)

	return duration, nil
}

// readWAVDuration 读取 WAV 文件时长
func (adr *AudioDurationReader) readWAVDuration(filePath string) (int64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("打开文件失败：%w", err)
	}
	defer file.Close()

	// 获取文件大小
	fileInfo, err := file.Stat()
	if err != nil {
		return 0, fmt.Errorf("获取文件信息失败：%w", err)
	}

	// WAV 文件结构：
	// - 44 字节头部（标准 WAV 头）
	// - 剩余部分是音频数据
	
	// 读取 WAV 头部信息
	header := make([]byte, 44)
	if _, err := file.Read(header); err != nil {
		return 0, fmt.Errorf("读取 WAV 头部失败：%w", err)
	}

	// 验证 RIFF 标识
	if string(header[:4]) != "RIFF" {
		return 0, fmt.Errorf("不是有效的 WAV 文件")
	}

	// 读取采样率（偏移 24-27 字节）
	sampleRate := binary.LittleEndian.Uint32(header[24:28])
	
	// 读取声道数（偏移 22-23 字节）
	channels := binary.LittleEndian.Uint16(header[22:24])
	
	// 读取位深（偏移 34-35 字节）
	bitsPerSample := binary.LittleEndian.Uint16(header[34:36])

	if sampleRate == 0 || channels == 0 || bitsPerSample == 0 {
		return 0, fmt.Errorf("无效的 WAV 参数")
	}

	// 计算音频数据大小（文件大小 - 头部大小）
	audioDataSize := fileInfo.Size() - 44
	
	// 计算每秒的字节数
	bytesPerSecond := int64(sampleRate) * int64(channels) * int64(bitsPerSample/8)
	
	if bytesPerSecond == 0 {
		return 0, fmt.Errorf("无效的字节率")
	}

	// 计算时长（秒）
	duration := audioDataSize / bytesPerSecond

	return duration, nil
}

// ClearCache 清除时长缓存
func (adr *AudioDurationReader) ClearCache() {
	adr.mu.Lock()
	defer adr.mu.Unlock()
	adr.cache = make(map[string]int64)
	log.Println("✓ 时长缓存已清除")
}
