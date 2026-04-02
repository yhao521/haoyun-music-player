package backend

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
	"github.com/gopxl/beep/v2/flac"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/gopxl/beep/v2/wav"
	"github.com/tosone/minimp3"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// MP3Streamer 基于 minimp3 的流式读取器
type MP3Streamer struct {
	pcmData []byte      // 完全解码后的 PCM 数据
	pos     int         // 当前读取位置 (字节)
	format  beep.Format // 音频格式
	mu      sync.Mutex  // 并发保护
	closed  bool        // 是否已关闭
}

// NewMP3Streamer 创建 MP3 流式读取器
func NewMP3Streamer(file *os.File) (*MP3Streamer, error) {
	// 读取整个文件到内存
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败：%w", err)
	}

	// 使用 minimp3 的 DecodeFull 一次性解码所有数据
	decoder, pcmData, err := minimp3.DecodeFull(data)
	if err != nil {
		return nil, fmt.Errorf("MP3 解码失败：%w", err)
	}

	// 确定音频格式
	var sampleRate beep.SampleRate
	switch decoder.SampleRate {
	case 8000, 11025, 12000, 16000, 22050, 24000, 32000, 44100, 48000:
		sampleRate = beep.SampleRate(decoder.SampleRate)
	default:
		sampleRate = beep.SampleRate(44100) // 默认
	}

	numChannels := decoder.Channels
	if numChannels < 1 || numChannels > 2 {
		numChannels = 2
	}

	streamer := &MP3Streamer{
		pcmData: pcmData,
		pos:     0,
		format: beep.Format{
			SampleRate:  sampleRate,
			NumChannels: numChannels,
			Precision:   2, // 16-bit
		},
	}

	return streamer, nil
}

// Stream 实现 beep.Streamer 接口
func (m *MP3Streamer) Stream(samples [][2]float64) (n int, ok bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed || m.pos >= len(m.pcmData) {
		return 0, false
	}

	channels := m.format.NumChannels
	bytesPerSample := channels * 2 // 每个声道 2 字节 (16-bit)
	
	decodedSamples := 0
	
	for decodedSamples < len(samples) && m.pos+bytesPerSample <= len(m.pcmData) {
		sample := [2]float64{}
		
		if channels == 1 {
			// 单声道
			val := int16(binary.LittleEndian.Uint16(m.pcmData[m.pos : m.pos+2]))
			sample[0] = float64(val) / 32768.0
			sample[1] = sample[0]
		} else if channels >= 2 {
			// 立体声
			left := int16(binary.LittleEndian.Uint16(m.pcmData[m.pos : m.pos+2]))
			right := int16(binary.LittleEndian.Uint16(m.pcmData[m.pos+2 : m.pos+4]))
			sample[0] = float64(left) / 32768.0
			sample[1] = float64(right) / 32768.0
		}
		
		samples[decodedSamples] = sample
		decodedSamples++
		m.pos += bytesPerSample
	}

	return decodedSamples, decodedSamples > 0
}

// Err 实现 beep.Streamer 接口
func (m *MP3Streamer) Err() error {
	return nil
}

// Close 实现 beep.StreamSeekCloser 接口
func (m *MP3Streamer) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// 释放 PCM 数据占用
	m.pcmData = nil
	m.closed = true
	return nil
}

// Len 返回总时长 (秒)
func (m *MP3Streamer) Len() int {
	bytesPerSample := int(m.format.NumChannels) * m.format.Precision
	totalSamples := len(m.pcmData) / bytesPerSample
	return totalSamples / int(m.format.SampleRate)
}

// Position 返回当前位置 (秒)
func (m *MP3Streamer) Position() int {
	bytesPerSample := int(m.format.NumChannels) * m.format.Precision
	currentSamples := m.pos / bytesPerSample
	return currentSamples / int(m.format.SampleRate)
}

// Seek 跳转位置
func (m *MP3Streamer) Seek(position int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	bytesPerSample := int(m.format.NumChannels) * m.format.Precision
	m.pos = position * int(m.format.SampleRate) * bytesPerSample

	if m.pos < 0 {
		m.pos = 0
	}
	if m.pos > len(m.pcmData) {
		m.pos = len(m.pcmData)
	}

	return nil
}

// AudioPlayer 基于 beep 的音频播放器
type AudioPlayer struct {
	mu                 sync.Mutex
	isPlaying          bool
	paused             bool
	volume             float64
	ctrl               *beep.Ctrl
	streamer           beep.StreamSeekCloser
	format             beep.Format
	gain               *effects.Gain
	app                *application.App
	speakerInitialized bool // 添加：跟踪 speaker 是否已初始化
}

// NewAudioPlayer 创建音频播放器实例
func NewAudioPlayer() *AudioPlayer {
	return &AudioPlayer{
		volume: 0.7,
	}
}

// SetApp 设置应用实例
func (ap *AudioPlayer) SetApp(app *application.App) {
	ap.app = app
}

// loadAudioFile 加载音频文件
func (ap *AudioPlayer) loadAudioFile(path string) (beep.StreamSeekCloser, beep.Format, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, beep.Format{}, fmt.Errorf("打开文件失败：%w", err)
	}

	ext := filepath.Ext(path)
	var streamer beep.StreamSeekCloser
	var format beep.Format

	switch ext {
	case ".mp3":
		// 使用 minimp3 解码 MP3 文件
		streamer, err = NewMP3Streamer(file)
		if err != nil {
			file.Close()
			return nil, beep.Format{}, err
		}
		format = streamer.(*MP3Streamer).format
		return streamer, format, nil // minimp3 streamer 已经包含了格式信息
	case ".wav":
		streamer, format, err = wav.Decode(file)
	case ".flac":
		streamer, format, err = flac.Decode(file)
	default:
		file.Close()
		return nil, beep.Format{}, fmt.Errorf("不支持的音频格式：%s", ext)
	}

	if err != nil {
		file.Close()
		return nil, beep.Format{}, fmt.Errorf("解码音频文件失败：%w", err)
	}

	return streamer, format, nil
}

// initSpeaker 初始化扬声器 (优化版本，避免重复初始化)
func (ap *AudioPlayer) initSpeaker(format beep.Format) error {
	// 如果 speaker 已经初始化且格式相同，则不需要重新初始化
	if ap.speakerInitialized && ap.format.SampleRate == format.SampleRate && ap.format.NumChannels == format.NumChannels {
		return nil
	}

	// 第一次初始化或格式改变时需要初始化
	speaker.Close()

	// 等待异步关闭完成
	time.Sleep(100 * time.Millisecond)

	err := speaker.Init(format.SampleRate, format.NumChannels*format.SampleRate.N(time.Second/10))
	if err != nil {
		ap.speakerInitialized = false
		return fmt.Errorf("初始化扬声器失败：%w", err)
	}

	ap.speakerInitialized = true
	return nil
}

// Play 播放音频文件
func (ap *AudioPlayer) Play(path string) error {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	// 停止当前播放并清理资源
	if ap.ctrl != nil {
		speaker.Clear()
		if ap.streamer != nil {
			ap.streamer.Close()
		}
		ap.ctrl = nil
		ap.streamer = nil
	}

	// 加载音频文件
	streamer, format, err := ap.loadAudioFile(path)
	if err != nil {
		return err
	}

	// 初始化扬声器 (如果格式相同则跳过)
	if err := ap.initSpeaker(format); err != nil {
		streamer.Close()
		return err
	}

	// 创建增益控制器
	gain := &effects.Gain{
		Streamer: streamer,
		Gain:     ap.volume - 1,
	}
	ap.gain = gain

	// 创建播放控制器
	ctrl := &beep.Ctrl{
		Streamer: gain,
		Paused:   false,
	}
	ap.ctrl = ctrl
	ap.streamer = streamer
	ap.format = format

	// 开始播放
	speaker.Play(ctrl)

	ap.isPlaying = true
	ap.paused = false

	return nil
}

// Pause 暂停播放
func (ap *AudioPlayer) Pause() error {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	if ap.ctrl == nil {
		return fmt.Errorf("当前没有播放的音乐")
	}

	ap.ctrl.Paused = true
	ap.paused = true
	ap.isPlaying = false

	if ap.app != nil {
		ap.app.Event.Emit("playbackStateChanged", "paused")
	}

	return nil
}

// Stop 停止播放
func (ap *AudioPlayer) Stop() error {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	if ap.ctrl != nil {
		speaker.Clear() // 完全清理 speaker 状态
		if ap.streamer != nil {
			ap.streamer.Close()
		}
		ap.ctrl = nil
		ap.streamer = nil
	}

	ap.isPlaying = false
	ap.paused = false

	if ap.app != nil {
		ap.app.Event.Emit("playbackStateChanged", "stopped")
	}

	return nil
}

// SetVolume 设置音量
func (ap *AudioPlayer) SetVolume(volume float64) error {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	if volume < 0 || volume > 1 {
		return fmt.Errorf("音量必须在 0 到 1 之间")
	}

	ap.volume = volume

	if ap.gain != nil {
		ap.gain.Gain = volume - 1 // Gain is multiplied by 1+Gain
	}

	return nil
}

// GetVolume 获取音量
func (ap *AudioPlayer) GetVolume() (float64, error) {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	return ap.volume, nil
}

// Seek 跳转到指定位置 (秒)
func (ap *AudioPlayer) Seek(position float64) error {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	if ap.ctrl == nil {
		return fmt.Errorf("当前没有播放的音乐")
	}

	// TODO: 实现 seek 功能，需要计算 sample position
	return nil
}

// GetDuration 获取歌曲总时长 (秒)
func (ap *AudioPlayer) GetDuration() (float64, error) {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	if ap.streamer == nil {
		return 0, fmt.Errorf("当前没有加载音乐")
	}

	// TODO: 实现时长获取
	return 0, nil
}

// GetPosition 获取当前播放位置 (秒)
func (ap *AudioPlayer) GetPosition() (float64, error) {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	if ap.ctrl == nil {
		return 0, fmt.Errorf("当前没有播放的音乐")
	}

	// TODO: 实现位置获取
	return 0, nil
}

// IsPlaying 检查是否正在播放
func (ap *AudioPlayer) IsPlaying() (bool, error) {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	return ap.isPlaying && !ap.paused, nil
}

// TogglePlayPause 切换播放/暂停
func (ap *AudioPlayer) TogglePlayPause() (bool, error) {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	if ap.ctrl == nil {
		return false, fmt.Errorf("当前没有播放的音乐")
	}

	ap.ctrl.Paused = !ap.ctrl.Paused
	ap.paused = ap.ctrl.Paused
	ap.isPlaying = !ap.ctrl.Paused

	state := "playing"
	if ap.paused {
		state = "paused"
	}

	if ap.app != nil {
		ap.app.Event.Emit("playbackStateChanged", state)
	}

	return ap.isPlaying, nil
}
