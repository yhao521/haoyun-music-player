package backend

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// PlaylistManager 播放列表管理
type PlaylistManager struct {
	mu       sync.RWMutex
	playlist []string
	current  int
	app      *application.App
	playMode string // 播放模式：order(顺序), loop(循环), random(随机)
}

// NewPlaylistManager 创建播放列表管理器
func NewPlaylistManager() *PlaylistManager {
	return &PlaylistManager{
		playlist: make([]string, 0),
		current:  -1,
		playMode: "order", // 默认为顺序播放
	}
}

// SetApp 设置应用实例
func (pm *PlaylistManager) SetApp(app *application.App) {
	pm.app = app
}

// AddToPlaylist 添加到播放列表
func (pm *PlaylistManager) AddToPlaylist(path string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在：%s", path)
	}

	pm.playlist = append(pm.playlist, path)
	if pm.app != nil {
		pm.app.Event.Emit("playlistUpdated", pm.playlist)
	}
	return nil
}

// ClearPlaylist 清空播放列表
func (pm *PlaylistManager) ClearPlaylist() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.playlist = make([]string, 0)
	pm.current = -1
	if pm.app != nil {
		pm.app.Event.Emit("playlistUpdated", pm.playlist)
	}
	return nil
}

// GetPlaylist 获取播放列表
func (pm *PlaylistManager) GetPlaylist() ([]string, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.playlist, nil
}

// GetCurrentIndex 获取当前播放索引
func (pm *PlaylistManager) GetCurrentIndex() (int, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.current, nil
}

// PlayIndex 播放指定索引的歌曲
func (pm *PlaylistManager) PlayIndex(index int) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if index < 0 || index >= len(pm.playlist) {
		return fmt.Errorf("索引越界：%d", index)
	}

	pm.current = index
	path := pm.playlist[index]

	if pm.app != nil {
		pm.app.Event.Emit("currentTrackChanged", filepath.Base(path))
	}

	return nil
}

// Next 播放下一首
func (pm *PlaylistManager) Next() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if len(pm.playlist) == 0 {
		return fmt.Errorf("播放列表为空")
	}

	switch pm.playMode {
	case "random":
		// 随机播放
		pm.current = rand.Intn(len(pm.playlist))
	case "loop":
		// 循环播放（包括当前歌曲）
		pm.current = pm.current % len(pm.playlist)
	case "order":
		fallthrough
	default:
		// 顺序播放
		pm.current = (pm.current + 1) % len(pm.playlist)
	}

	path := pm.playlist[pm.current]

	if pm.app != nil {
		pm.app.Event.Emit("currentTrackChanged", filepath.Base(path))
	}

	return nil
}

// Previous 播放上一首
func (pm *PlaylistManager) Previous() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if len(pm.playlist) == 0 {
		return fmt.Errorf("播放列表为空")
	}

	switch pm.playMode {
	case "random":
		// 随机播放
		pm.current = rand.Intn(len(pm.playlist))
	case "loop":
		// 循环播放
		pm.current = (pm.current - 1 + len(pm.playlist)) % len(pm.playlist)
	case "order":
		fallthrough
	default:
		// 顺序播放
		pm.current = (pm.current - 1 + len(pm.playlist)) % len(pm.playlist)
	}

	path := pm.playlist[pm.current]

	if pm.app != nil {
		pm.app.Event.Emit("currentTrackChanged", filepath.Base(path))
	}

	return nil
}

// SetPlayMode 设置播放模式
func (pm *PlaylistManager) SetPlayMode(mode string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	validModes := map[string]bool{
		"order":  true,
		"loop":   true,
		"random": true,
	}

	if !validModes[mode] {
		return fmt.Errorf("无效的播放模式：%s", mode)
	}

	pm.playMode = mode
	log.Printf("播放模式设置为：%s", mode)
	return nil
}

// GetPlayMode 获取当前播放模式
func (pm *PlaylistManager) GetPlayMode() (string, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.playMode, nil
}
