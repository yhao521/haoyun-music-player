package backend

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// createTrackInfo 从文件路径创建 TrackInfo（基础版本，不使用音乐库）
func createTrackInfo(path string) TrackInfo {
	filename := filepath.Base(path)
	title := strings.TrimSuffix(filename, filepath.Ext(filename))
	
	return TrackInfo{
		Path:     path,
		Filename: filename,
		Title:    title,
		Artist:   "",
		Album:    "",
		Duration: 0,
		Size:     0,
		LyricPath: "",
	}
}

// createTrackInfoFromLibrary 从音乐库获取完整的 TrackInfo（优先使用扫描结果）
func createTrackInfoFromLibrary(path string, libraryManager *LibraryManager) TrackInfo {
	// 策略 1: 尝试从音乐库中获取已扫描的信息（使用 O(1) 索引查找）
	if libraryManager != nil {
		track := libraryManager.GetTrackByPath(path)
		if track != nil {
			return *track
		}
	}
	
	// 策略 2: 降级到基本信息
	return createTrackInfo(path)
}

// PlaylistManager 播放列表管理
type PlaylistManager struct {
	mu              sync.RWMutex
	playlist        []string
	current         int
	app             *application.App
	playMode        string           // 播放模式：order(顺序), loop(循环), single(单曲循环), random(随机)
	libraryManager  *LibraryManager  // 音乐库管理器（用于获取元数据）
}

// NewPlaylistManager 创建播放列表管理器
func NewPlaylistManager() *PlaylistManager {
	return &PlaylistManager{
		playlist: make([]string, 0),
		current:  -1,
		playMode: "loop", // 默认为循环播放
	}
}

// SetLibraryManager 设置音乐库管理器
func (pm *PlaylistManager) SetLibraryManager(lm *LibraryManager) {
	pm.libraryManager = lm
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

// AddToPlaylistBatch 批量添加到播放列表（只发送一次事件）
func (pm *PlaylistManager) AddToPlaylistBatch(paths []string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	validPaths := make([]string, 0, len(paths))
	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			log.Printf("跳过不存在的文件：%s", path)
			continue
		}
		validPaths = append(validPaths, path)
	}

	pm.playlist = append(pm.playlist, validPaths...)
	
	// 只发送一次事件
	if pm.app != nil && len(validPaths) > 0 {
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
		// 使用音乐库获取完整的 TrackInfo（O(1) 查找）
		trackInfo := createTrackInfoFromLibrary(path, pm.libraryManager)
		pm.app.Event.Emit("currentTrackChanged", trackInfo)
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
	case "single":
		// 单曲循环：保持当前歌曲不变
		// current 索引不变
	case "loop":
		// 列表循环：播完最后一首回到第一首
		pm.current = (pm.current + 1) % len(pm.playlist)
	case "order":
		fallthrough
	default:
		// 顺序播放：播完最后一首停止（不自动下一首）
		// 这里暂时和 loop 一样，实际应该在 audioplayer 中处理
		pm.current = (pm.current + 1) % len(pm.playlist)
	}

	path := pm.playlist[pm.current]

	if pm.app != nil {
		// 使用音乐库获取完整的 TrackInfo
		trackInfo := createTrackInfoFromLibrary(path, pm.libraryManager)
		pm.app.Event.Emit("currentTrackChanged", trackInfo)
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
	case "single":
		// 单曲循环：保持当前歌曲不变
		// current 索引不变
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
		// 使用音乐库获取完整的 TrackInfo
		trackInfo := createTrackInfoFromLibrary(path, pm.libraryManager)
		pm.app.Event.Emit("currentTrackChanged", trackInfo)
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
		"single": true, // 单曲循环
		"random": true,
	}

	if !validModes[mode] {
		return fmt.Errorf("无效的播放模式：%s", mode)
	}

	pm.playMode = mode
	log.Printf("播放模式设置为：%s", mode)
	
	// 发送播放模式变化事件，通知所有监听者（前端和托盘菜单）
	if pm.app != nil {
		go func() {
			pm.app.Event.Emit("playModeChanged", mode)
		}()
	}
	
	return nil
}

// GetPlayMode 获取当前播放模式
func (pm *PlaylistManager) GetPlayMode() (string, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.playMode, nil
}
