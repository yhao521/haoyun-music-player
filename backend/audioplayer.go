package backend

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
	"github.com/gopxl/beep/v2/flac"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/gopxl/beep/v2/wav"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// AudioPlayer 基于 beep 的音频播放器
type AudioPlayer struct {
	mu        sync.RWMutex
	isPlaying bool
	paused    bool
	volume    float64

	// 当前播放控制
	ctrl     *beep.Ctrl
	streamer beep.StreamSeekCloser
	format   beep.Format

	// 音效处理
	gain *effects.Gain // 增益

	app *application.App
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

// initSpeaker 初始化扬声器
func (ap *AudioPlayer) initSpeaker(format beep.Format) error {
	// 如果已经初始化，先关闭
	speaker.Close()

	err := speaker.Init(format.SampleRate, format.NumChannels*format.SampleRate.N(time.Second/10))
	if err != nil {
		return fmt.Errorf("初始化扬声器失败：%w", err)
	}

	return nil
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
		streamer, format, err = mp3.Decode(file)
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

// Play 播放音频文件
func (ap *AudioPlayer) Play(path string) error {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	// 停止当前播放
	if ap.ctrl != nil {
		speaker.Clear()
	}

	// 加载音频文件
	streamer, format, err := ap.loadAudioFile(path)
	if err != nil {
		return err
	}

	// 初始化扬声器
	if err := ap.initSpeaker(format); err != nil {
		streamer.Close()
		return err
	}

	// 创建增益控制器
	gain := &effects.Gain{
		Streamer: streamer,
		Gain:     ap.volume - 1, // Gain is multiplied by 1+Gain, so volume-1 gives 0 to 1 range
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

	// 发送播放状态事件
	if ap.app != nil {
		ap.app.Event.Emit("playbackStateChanged", "playing")
	}

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
		speaker.Clear()
		ap.ctrl = nil
	}

	// streamer 不需要手动 Close，beep 库会自动管理
	ap.streamer = nil

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
	ap.mu.RLock()
	defer ap.mu.RUnlock()
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
	ap.mu.RLock()
	defer ap.mu.RUnlock()

	if ap.streamer == nil {
		return 0, fmt.Errorf("当前没有加载音乐")
	}

	// TODO: 实现时长获取
	return 0, nil
}

// GetPosition 获取当前播放位置 (秒)
func (ap *AudioPlayer) GetPosition() (float64, error) {
	ap.mu.RLock()
	defer ap.mu.RUnlock()

	if ap.ctrl == nil {
		return 0, fmt.Errorf("当前没有播放的音乐")
	}

	// TODO: 实现位置获取
	return 0, nil
}

// IsPlaying 检查是否正在播放
func (ap *AudioPlayer) IsPlaying() (bool, error) {
	ap.mu.RLock()
	defer ap.mu.RUnlock()
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
