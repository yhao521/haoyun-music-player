package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// MusicService 音乐播放服务（简化版本）
type MusicService struct {
	ctx      context.Context
	app      *application.App
	mu       sync.RWMutex
	playlist []string
	current  int
	isPlaying bool
	volume   float64
}

// NewMusicService 创建音乐服务实例
func NewMusicService() *MusicService {
	return &MusicService{
		playlist: make([]string, 0),
		current:  -1,
		volume:   0.7,
	}
}

// SetApp 设置应用实例
func (m *MusicService) SetApp(app *application.App) {
	m.app = app
}

// SetContext 设置上下文
func (m *MusicService) SetContext(ctx context.Context) {
	m.ctx = ctx
}

// OpenFilePicker 打开文件选择器
func (m *MusicService) OpenFilePicker() ([]string, error) {
	if m.app == nil {
		return nil, fmt.Errorf("app not initialized")
	}
	
	// 创建文件对话框选项
	options := &application.OpenFileDialogOptions{
		CanChooseFiles:          true,
		CanChooseDirectories:    false,
		AllowsMultipleSelection: true,
		Title:                   "Select Music Files",
		Filters: []application.FileFilter{
			{DisplayName: "Audio Files", Pattern: "*.mp3,*.wav,*.flac,*.aac,*.ogg,*.wma"},
			{DisplayName: "All Files", Pattern: "*.*"},
		},
	}
	
	// 打开文件选择对话框
	dialog := m.app.Dialog.OpenFileWithOptions(options)
	selectedFiles, err := dialog.PromptForMultipleSelection()
	if err != nil {
		return nil, fmt.Errorf("file dialog error: %w", err)
	}
	
	// 如果没有选择文件，返回空数组
	if selectedFiles == nil {
		return []string{}, nil
	}
	
	return selectedFiles, nil
}

// LoadFile 加载音乐文件（模拟实现）
func (m *MusicService) LoadFile(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 验证文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", path)
	}

	// 发送当前歌曲变化事件
	if m.app != nil {
		m.app.Event.Emit("currentTrackChanged", filepath.Base(path))
	}

	return nil
}

// Play 播放音乐
func (m *MusicService) Play() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.isPlaying = true
	if m.app != nil {
		m.app.Event.Emit("playbackStateChanged", "playing")
	}

	return nil
}

// Pause 暂停音乐
func (m *MusicService) Pause() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.isPlaying = false
	if m.app != nil {
		m.app.Event.Emit("playbackStateChanged", "paused")
	}

	return nil
}

// Stop 停止音乐
func (m *MusicService) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.isPlaying = false
	if m.app != nil {
		m.app.Event.Emit("playbackStateChanged", "stopped")
	}

	return nil
}

// TogglePlayPause 切换播放/暂停
func (m *MusicService) TogglePlayPause() (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.isPlaying = !m.isPlaying
	
	state := "playing"
	if !m.isPlaying {
		state = "paused"
	}
	
	if m.app != nil {
		m.app.Event.Emit("playbackStateChanged", state)
	}

	return m.isPlaying, nil
}

// SetVolume 设置音量
func (m *MusicService) SetVolume(volume float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if volume < 0 || volume > 1 {
		return fmt.Errorf("volume must be between 0 and 1")
	}

	m.volume = volume
	return nil
}

// GetVolume 获取当前音量
func (m *MusicService) GetVolume() (float64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.volume, nil
}

// Seek 跳转到指定位置
func (m *MusicService) Seek(position float64) error {
	// 简化实现
	return nil
}

// GetDuration 获取歌曲时长
func (m *MusicService) GetDuration() (float64, error) {
	return 0, nil
}

// GetPosition 获取当前播放位置
func (m *MusicService) GetPosition() (float64, error) {
	return 0, nil
}

// IsPlaying 检查是否正在播放
func (m *MusicService) IsPlaying() (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isPlaying, nil
}

// AddToPlaylist 添加到播放列表
func (m *MusicService) AddToPlaylist(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", path)
	}

	m.playlist = append(m.playlist, path)
	if m.app != nil {
		m.app.Event.Emit("playlistUpdated", m.playlist)
	}
	return nil
}

// GetPlaylist 获取播放列表
func (m *MusicService) GetPlaylist() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.playlist, nil
}

// PlayIndex 播放指定索引的歌曲
func (m *MusicService) PlayIndex(index int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if index < 0 || index >= len(m.playlist) {
		return fmt.Errorf("invalid index: %d", index)
	}

	m.current = index
	path := m.playlist[index]

	if m.app != nil {
		m.app.Event.Emit("currentTrackChanged", filepath.Base(path))
		m.isPlaying = true
		m.app.Event.Emit("playbackStateChanged", "playing")
	}

	return nil
}

// Next 播放下一首
func (m *MusicService) Next() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.playlist) == 0 {
		return fmt.Errorf("playlist is empty")
	}

	m.current = (m.current + 1) % len(m.playlist)
	path := m.playlist[m.current]

	if m.app != nil {
		m.app.Event.Emit("currentTrackChanged", filepath.Base(path))
		m.isPlaying = true
		m.app.Event.Emit("playbackStateChanged", "playing")
	}

	return nil
}

// Previous 播放上一首
func (m *MusicService) Previous() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.playlist) == 0 {
		return fmt.Errorf("playlist is empty")
	}

	m.current = (m.current - 1 + len(m.playlist)) % len(m.playlist)
	path := m.playlist[m.current]

	if m.app != nil {
		m.app.Event.Emit("currentTrackChanged", filepath.Base(path))
		m.isPlaying = true
		m.app.Event.Emit("playbackStateChanged", "playing")
	}

	return nil
}

// GetSongMetadata 获取歌曲元数据
func (m *MusicService) GetSongMetadata(path string) (map[string]interface{}, error) {
	filename := filepath.Base(path)
	return map[string]interface{}{
		"title":  filename,
		"artist": "未知艺术家",
		"album":  "未知专辑",
		"path":   path,
	}, nil
}

// Shutdown 关闭服务
func (m *MusicService) Shutdown() error {
	return nil
}
