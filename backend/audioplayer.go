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
	streamer beep.Streamer
	format   beep.Format
	
	// 音效处理
	gain    *effects.Gain // 增益
	
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
func (ap *AudioPlayer) loadAudioFile(path string) (beep.Streamer, beep.Format, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, beep.Format{}, fmt.Errorf("打开文件失败：%w", err)
	}

	ext := filepath.Ext(path)
	var streamer beep.Streamer
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
		streamer.(interface{ Close() }).Close()
		return err
	}
	
	// 创建增益控制
	gainEffect := &effects.Gain{
		Streamer: streamer,
		Gain:     ap.volume,
	}
	
	// 创建控制器
	ctrl := &beep.Ctrl{
		Streamer: gainEffect,
		Paused:   false,
	}
	
	ap.streamer = streamer
	ap.ctrl = ctrl
	ap.format = format
	ap.isPlaying = true
	ap.paused = false
	ap.gain = gainEffect
	
	// 开始播放
	speaker.Play(ctrl)
	
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
	
	if !ap.isPlaying || ap.ctrl == nil {
		return nil
	}
	
	err := speaker.Suspend()
	if err != nil {
		return fmt.Errorf("暂停失败：%w", err)
	}
	
	ap.paused = true
	
	if ap.app != nil {
		ap.app.Event.Emit("playbackStateChanged", "paused")
	}
	
	return nil
}

// Resume 恢复播放
func (ap *AudioPlayer) Resume() error {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	
	if !ap.isPlaying || !ap.paused || ap.ctrl == nil {
		return nil
	}
	
	err := speaker.Resume()
	if err != nil {
		return fmt.Errorf("恢复失败：%w", err)
	}
	
	ap.paused = false
	
	if ap.app != nil {
		ap.app.Event.Emit("playbackStateChanged", "playing")
	}
	
	return nil
}

// Stop 停止播放
func (ap *AudioPlayer) Stop() error {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	
	if ap.ctrl == nil {
		return nil
	}
	
	speaker.Clear()
	ap.ctrl.Streamer = nil
	ap.ctrl = nil
	if closer, ok := ap.streamer.(interface{ Close() }); ok {
		closer.Close()
	}
	ap.streamer = nil
	ap.isPlaying = false
	ap.paused = false
	
	if ap.app != nil {
		ap.app.Event.Emit("playbackStateChanged", "stopped")
	}
	
	return nil
}

// TogglePlayPause 切换播放/暂停
func (ap *AudioPlayer) TogglePlayPause() (bool, error) {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	
	if !ap.isPlaying || ap.ctrl == nil {
		return false, fmt.Errorf("当前没有播放任何音频")
	}
	
	if ap.paused {
		// 恢复播放
		err := speaker.Resume()
		if err != nil {
			return false, fmt.Errorf("恢复播放失败：%w", err)
		}
		
		ap.paused = false
		
		if ap.app != nil {
			ap.app.Event.Emit("playbackStateChanged", "playing")
		}
		
		return true, nil
	} else {
		// 暂停播放
		err := speaker.Suspend()
		if err != nil {
			return false, fmt.Errorf("暂停播放失败：%w", err)
		}
		
		ap.paused = true
		
		if ap.app != nil {
			ap.app.Event.Emit("playbackStateChanged", "paused")
		}
		
		return false, nil
	}
}

// IsPlaying 检查是否正在播放
func (ap *AudioPlayer) IsPlaying() (bool, error) {
	ap.mu.RLock()
	defer ap.mu.RUnlock()
	return ap.isPlaying && !ap.paused, nil
}

// SetVolume 设置音量
func (ap *AudioPlayer) SetVolume(volume float64) error {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	
	if volume < 0 || volume > 1 {
		return fmt.Errorf("音量必须在 0-1 之间")
	}
	
	ap.volume = volume
	
	if ap.gain != nil {
		ap.gain.Gain = volume
	}
	
	return nil
}

// GetVolume 获取音量
func (ap *AudioPlayer) GetVolume() (float64, error) {
	ap.mu.RLock()
	defer ap.mu.RUnlock()
	return ap.volume, nil
}

// Seek 跳转到指定位置（简化实现）
func (ap *AudioPlayer) Seek(position time.Duration) error {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	
	if ap.ctrl == nil || ap.streamer == nil {
		return fmt.Errorf("当前没有播放任何音频")
	}
	
	// TODO: 实现跳转功能，需要包装 streamer 支持 Seek
	// 目前 beep 库的 Streamer 接口不直接支持 Seek
	
	return nil
}

// GetPosition 获取当前播放位置（简化实现）
func (ap *AudioPlayer) GetPosition() (time.Duration, error) {
	ap.mu.RLock()
	defer ap.mu.RUnlock()
	
	if ap.streamer == nil {
		return 0, fmt.Errorf("当前没有播放任何音频")
	}
	
	// TODO: 实现获取当前位置
	
	return 0, nil
}

// GetDuration 获取音频总时长（简化实现）
func (ap *AudioPlayer) GetDuration() (time.Duration, error) {
	ap.mu.RLock()
	defer ap.mu.RUnlock()
	
	if ap.streamer == nil {
		return 0, fmt.Errorf("当前没有播放任何音频")
	}
	
	// TODO: 实现获取总时长
	
	return 0, nil
}

// SetEQ 设置均衡器（预留接口）
func (ap *AudioPlayer) SetEQ(lowGain, midGain, highGain float64) error {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	
	if ap.gain == nil {
		return fmt.Errorf("播放器未初始化")
	}
	
	// TODO: 使用 effects.NewEqualizer 实现均衡器
	
	return nil
}

// Shutdown 关闭播放器
func (ap *AudioPlayer) Shutdown() error {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	
	if ap.ctrl != nil {
		speaker.Clear()
	}
	
	speaker.Close()
	
	return nil
}
