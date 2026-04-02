package backend

import (
	"context"
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

// TrackInfo 音乐文件信息
type TrackInfo2 struct {
	Path     string `json:"path"`
	Filename string `json:"filename"`
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Album    string `json:"album"`
	Duration int64  `json:"duration"` // 秒
	Size     int64  `json:"size"`     // 字节
}

// PlaybackState 播放状态管理
type PlaybackState struct {
	mu        sync.RWMutex
	isPlaying bool
	volume    float64
	app       *application.App
}

// NewPlaybackState 创建播放状态管理器
func NewPlaybackState() *PlaybackState {
	return &PlaybackState{
		isPlaying: false,
		volume:    0.7,
	}
}

// SetApp 设置应用实例
func (p *PlaybackState) SetApp(app *application.App) {
	p.app = app
}

// TogglePlayPause 切换播放/暂停
func (p *PlaybackState) TogglePlayPause() (bool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.isPlaying = !p.isPlaying

	state := "playing"
	if !p.isPlaying {
		state = "paused"
	}

	if p.app != nil {
		p.app.Event.Emit("playbackStateChanged", state)
	}

	return p.isPlaying, nil
}

// Play 播放
func (p *PlaybackState) Play() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.isPlaying = true
	if p.app != nil {
		p.app.Event.Emit("playbackStateChanged", "playing")
	}
	return nil
}

// Pause 暂停
func (p *PlaybackState) Pause() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.isPlaying = false
	if p.app != nil {
		p.app.Event.Emit("playbackStateChanged", "paused")
	}
	return nil
}

// Stop 停止
func (p *PlaybackState) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.isPlaying = false
	if p.app != nil {
		p.app.Event.Emit("playbackStateChanged", "stopped")
	}
	return nil
}

// SetVolume 设置音量
func (p *PlaybackState) SetVolume(volume float64) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if volume < 0 || volume > 1 {
		return fmt.Errorf("volume must be between 0 and 1")
	}

	p.volume = volume
	return nil
}

// GetVolume 获取音量
func (p *PlaybackState) GetVolume() (float64, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.volume, nil
}

// IsPlaying 检查是否正在播放
func (p *PlaybackState) IsPlaying() (bool, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.isPlaying, nil
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
		return fmt.Errorf("file does not exist: %s", path)
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
		return fmt.Errorf("invalid index: %d", index)
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
		return fmt.Errorf("playlist is empty")
	}

	switch pm.playMode {
	case "random":
		// 随机播放
		pm.current = rand.Intn(len(pm.playlist))
	case "loop":
		// 循环播放（包括当前歌曲）
		// 保持当前索引不变，由外部控制重新播放
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
		return fmt.Errorf("playlist is empty")
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
		return fmt.Errorf("invalid play mode: %s", mode)
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

// MusicController 音乐服务主控制器（组合模式）
type MusicController struct {
	ctx             context.Context
	app             *application.App
	Com
	playbackState   *PlaybackState
	playlistManager *PlaylistManager
	libraryManager  *LibraryManager
}

// NewMusicManager 创建音乐服务实例
func NewMusicManager() *MusicController {
	return &MusicController{
		playbackState:   NewPlaybackState(),
		playlistManager: NewPlaylistManager(),
	}
}

// SetApp 设置应用实例
func (m *MusicController) SetApp(app *application.App) {
	m.app = app
	m.Com.SetApp(app)
	m.playbackState.SetApp(app)
	m.playlistManager.SetApp(app)
}

// SetLibraryManager 设置音乐库管理器引用
func (m *MusicController) SetLibraryManager(lm *LibraryManager) {
	m.libraryManager = lm
}

// SetContext 设置上下文
func (m *MusicController) SetContext(ctx context.Context) {
	m.ctx = ctx
}

// OpenFilePicker 打开文件选择器
func (m *MusicController) OpenFilePicker() ([]string, error) {
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
func (m *MusicController) LoadFile(path string) error {
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
func (m *MusicController) Play() error {
	return m.playbackState.Play()
}

// Pause 暂停音乐
func (m *MusicController) Pause() error {
	return m.playbackState.Pause()
}

// Stop 停止音乐
func (m *MusicController) Stop() error {
	return m.playbackState.Stop()
}

// TogglePlayPause 切换播放/暂停
func (m *MusicController) TogglePlayPause() (bool, error) {
	return m.playbackState.TogglePlayPause()
}

// SetVolume 设置音量
func (m *MusicController) SetVolume(volume float64) error {
	return m.playbackState.SetVolume(volume)
}

// GetVolume 获取当前音量
func (m *MusicController) GetVolume() (float64, error) {
	return m.playbackState.GetVolume()
}

// Seek 跳转到指定位置
func (m *MusicController) Seek(position float64) error {
	// 简化实现
	return nil
}

// GetDuration 获取歌曲时长
func (m *MusicController) GetDuration() (float64, error) {
	return 0, nil
}

// GetPosition 获取当前播放位置
func (m *MusicController) GetPosition() (float64, error) {
	return 0, nil
}

// IsPlaying 检查是否正在播放
func (m *MusicController) IsPlaying() (bool, error) {
	return m.playbackState.IsPlaying()
}

// AddToPlaylist 添加到播放列表
func (m *MusicController) AddToPlaylist(path string) error {
	return m.playlistManager.AddToPlaylist(path)
}

// ClearPlaylist 清空播放列表
func (m *MusicController) ClearPlaylist() error {
	return m.playlistManager.ClearPlaylist()
}

// GetPlaylist 获取播放列表
func (m *MusicController) GetPlaylist() ([]string, error) {
	return m.playlistManager.GetPlaylist()
}

// PlayIndex 播放指定索引的歌曲
func (m *MusicController) PlayIndex(index int) error {
	err := m.playlistManager.PlayIndex(index)
	if err != nil {
		return err
	}
	// 播放状态设为播放中
	return m.playbackState.Play()
}

// Next 播放下一首
func (m *MusicController) Next() error {
	err := m.playlistManager.Next()
	if err != nil {
		return err
	}
	// 播放状态设为播放中
	return m.playbackState.Play()
}

// Previous 播放上一首
func (m *MusicController) Previous() error {
	err := m.playlistManager.Previous()
	if err != nil {
		return err
	}
	// 播放状态设为播放中
	return m.playbackState.Play()
}

// SetPlayMode 设置播放模式
func (m *MusicController) SetPlayMode(mode string) error {
	return m.playlistManager.SetPlayMode(mode)
}

// GetPlayMode 获取当前播放模式
func (m *MusicController) GetPlayMode() (string, error) {
	return m.playlistManager.GetPlayMode()
}

// LoadCurrentLibrary 加载当前音乐库到播放列表并播放
func (m *MusicController) LoadCurrentLibrary() error {
	if m.libraryManager == nil {
		return fmt.Errorf("library manager not initialized")
	}

	tracks, err := m.libraryManager.GetCurrentLibraryTracks()
	if err != nil {
		return fmt.Errorf("获取音轨失败：%w", err)
	}

	if len(tracks) == 0 {
		return fmt.Errorf("音乐库中没有音轨")
	}

	// 清空当前播放列表
	m.ClearPlaylist()

	// 将所有音轨添加到播放列表
	for _, track := range tracks {
		if err := m.AddToPlaylist(track); err != nil {
			log.Printf("添加音轨失败 %s: %v", track, err)
		}
	}

	// 播放第一首
	if len(tracks) > 0 {
		if err := m.PlayIndex(0); err != nil {
			return fmt.Errorf("播放第一首失败：%w", err)
		}
	}

	log.Printf("已加载音乐库 %s 到播放列表，共 %d 首歌曲", m.libraryManager.GetCurrentLibrary().Name, len(tracks))
	return nil
}

// GetSongMetadata 获取歌曲元数据
func (m *MusicController) GetSongMetadata(path string) (map[string]interface{}, error) {
	filename := filepath.Base(path)
	return map[string]interface{}{
		"title":  filename,
		"artist": "未知艺术家",
		"album":  "未知专辑",
		"path":   path,
	}, nil
}

// Shutdown 关闭服务
func (m *MusicController) Shutdown() error {
	return nil
}

// MusicService 音乐服务统一接口（MVC Model 层）
type MusicService struct {
	ctx             context.Context
	app             *application.App
	audioPlayer     *AudioPlayer        // beep 音频播放器
	playlistManager *PlaylistManager    // 播放列表管理
	libraryManager  *LibraryManager     // 音乐库管理
}

// NewMusicService 创建音乐服务实例
func NewMusicService() *MusicService {
	return &MusicService{
		audioPlayer:     NewAudioPlayer(),
		playlistManager: NewPlaylistManager(),
		libraryManager:  NewLibraryManager(),
	}
}

// SetApp 设置应用实例
func (m *MusicService) SetApp(app *application.App) {
	m.app = app
	m.audioPlayer.SetApp(app)
	m.playlistManager.SetApp(app)
	m.libraryManager.SetApp(app)
}

// SetContext 设置上下文
func (m *MusicService) SetContext(ctx context.Context) {
	m.ctx = ctx
}

// Init 初始化服务
func (m *MusicService) Init() error {
	return m.libraryManager.Init()
}

// ===== 播放控制方法 =====

// Play 播放音乐
func (m *MusicService) Play() error {
	// 获取当前播放的歌曲
	playlist, err := m.playlistManager.GetPlaylist()
	if err != nil {
		return err
	}
	
	if len(playlist) == 0 {
		return fmt.Errorf("播放列表为空")
	}
	
	currentIndex, _ := m.playlistManager.GetCurrentIndex()
	if currentIndex < 0 || currentIndex >= len(playlist) {
		// 如果没有当前歌曲，播放第一首
		return m.PlayIndex(0)
	}
	
	return m.audioPlayer.Play(playlist[currentIndex])
}

// Pause 暂停音乐
func (m *MusicService) Pause() error {
	return m.audioPlayer.Pause()
}

// Stop 停止音乐
func (m *MusicService) Stop() error {
	return m.audioPlayer.Stop()
}

// TogglePlayPause 切换播放/暂停
func (m *MusicService) TogglePlayPause() (bool, error) {
	// 如果当前没有在播放任何歌曲，尝试播放
	isPlaying, _ := m.audioPlayer.IsPlaying()
	if !isPlaying {
		err := m.Play()
		return true, err
	}
	
	return m.audioPlayer.TogglePlayPause()
}

// SetVolume 设置音量
func (m *MusicService) SetVolume(volume float64) error {
	return m.audioPlayer.SetVolume(volume)
}

// GetVolume 获取当前音量
func (m *MusicService) GetVolume() (float64, error) {
	return m.audioPlayer.GetVolume()
}

// IsPlaying 检查是否正在播放
func (m *MusicService) IsPlaying() (bool, error) {
	return m.audioPlayer.IsPlaying()
}

// ===== 播放列表管理方法 =====

// AddToPlaylist 添加到播放列表
func (m *MusicService) AddToPlaylist(path string) error {
	return m.playlistManager.AddToPlaylist(path)
}

// ClearPlaylist 清空播放列表
func (m *MusicService) ClearPlaylist() error {
	return m.playlistManager.ClearPlaylist()
}

// GetPlaylist 获取播放列表
func (m *MusicService) GetPlaylist() ([]string, error) {
	return m.playlistManager.GetPlaylist()
}

// PlayIndex 播放指定索引的歌曲
func (m *MusicService) PlayIndex(index int) error {
	err := m.playlistManager.PlayIndex(index)
	if err != nil {
		return err
	}
	
	// 获取播放列表
	playlist, err := m.playlistManager.GetPlaylist()
	if err != nil {
		return err
	}
	
	if index < 0 || index >= len(playlist) {
		return fmt.Errorf("无效的索引：%d", index)
	}
	
	// 使用 beep 播放器播放
	return m.audioPlayer.Play(playlist[index])
}

// Next 播放下一首
func (m *MusicService) Next() error {
	err := m.playlistManager.Next()
	if err != nil {
		return err
	}
	
	// 获取当前索引和播放列表
	currentIndex, err := m.playlistManager.GetCurrentIndex()
	if err != nil {
		return fmt.Errorf("获取当前索引失败：%w", err)
	}
	
	playlist, err := m.playlistManager.GetPlaylist()
	if err != nil {
		return fmt.Errorf("获取播放列表失败：%w", err)
	}
	
	if currentIndex >= 0 && currentIndex < len(playlist) {
		// 播放下一首
		return m.audioPlayer.Play(playlist[currentIndex])
	}
	
	return nil
}

// Previous 播放上一首
func (m *MusicService) Previous() error {
	err := m.playlistManager.Previous()
	if err != nil {
		return err
	}
	
	// 获取当前索引和播放列表
	currentIndex, err := m.playlistManager.GetCurrentIndex()
	if err != nil {
		return fmt.Errorf("获取当前索引失败：%w", err)
	}
	
	playlist, err := m.playlistManager.GetPlaylist()
	if err != nil {
		return fmt.Errorf("获取播放列表失败：%w", err)
	}
	
	if currentIndex >= 0 && currentIndex < len(playlist) {
		// 播放上一首
		return m.audioPlayer.Play(playlist[currentIndex])
	}
	
	return nil
}

// SetPlayMode 设置播放模式
func (m *MusicService) SetPlayMode(mode string) error {
	return m.playlistManager.SetPlayMode(mode)
}

// GetPlayMode 获取当前播放模式
func (m *MusicService) GetPlayMode() (string, error) {
	return m.playlistManager.GetPlayMode()
}

// LoadCurrentLibrary 加载当前音乐库到播放列表并播放
func (m *MusicService) LoadCurrentLibrary() error {
	tracks, err := m.libraryManager.GetCurrentLibraryTracks()
	if err != nil {
		return fmt.Errorf("获取音轨失败：%w", err)
	}

	if len(tracks) == 0 {
		return fmt.Errorf("音乐库中没有音轨")
	}

	// 清空当前播放列表
	m.ClearPlaylist()

	// 将所有音轨添加到播放列表
	for _, track := range tracks {
		if err := m.AddToPlaylist(track); err != nil {
			log.Printf("添加音轨失败 %s: %v", track, err)
		}
	}

	// 播放第一首
	if len(tracks) > 0 {
		if err := m.PlayIndex(0); err != nil {
			return fmt.Errorf("播放第一首失败：%w", err)
		}
	}

	log.Printf("已加载音乐库 %s 到播放列表，共 %d 首歌曲", m.libraryManager.GetCurrentLibrary().Name, len(tracks))
	return nil
}

// ===== 音乐库管理方法 =====

// GetLibraries 获取所有音乐库
func (m *MusicService) GetLibraries() []string {
	return m.libraryManager.GetLibraries()
}

// GetCurrentLibrary 获取当前音乐库
func (m *MusicService) GetCurrentLibrary() *MusicLibrary {
	return m.libraryManager.GetCurrentLibrary()
}

// AddLibrary 添加新音乐库
func (m *MusicService) AddLibrary() error {
	return m.libraryManager.AddLibrary()
}

// SetCurrentLibrary 设置当前音乐库
func (m *MusicService) SetCurrentLibrary(name string) error {
	return m.libraryManager.SetCurrentLibrary(name)
}

// RefreshLibrary 刷新音乐库
func (m *MusicService) RefreshLibrary(name string) error {
	return m.libraryManager.RefreshLibrary(name)
}

// RenameLibrary 重命名音乐库
func (m *MusicService) RenameLibrary(oldName, newName string) error {
	return m.libraryManager.RenameLibrary(oldName, newName)
}

// GetCurrentLibraryTracks 获取当前音乐库的所有音轨
func (m *MusicService) GetCurrentLibraryTracks() ([]string, error) {
	return m.libraryManager.GetCurrentLibraryTracks()
}

// ===== 文件操作相关方法 =====

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

// LoadFile 加载音乐文件
func (m *MusicService) LoadFile(path string) error {
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
	// 关闭音频播放器
	if err := m.audioPlayer.Shutdown(); err != nil {
		return err
	}
	
	return nil
}
