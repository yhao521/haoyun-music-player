package backend

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// MusicService 音乐服务统一接口（MVC Model 层）
type MusicService struct {
	ctx             context.Context
	app             *application.App
	audioPlayer     *AudioPlayer     // beep 音频播放器
	playlistManager *PlaylistManager // 播放列表管理
	libraryManager  *LibraryManager  // 音乐库管理
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

	currentIndex, err := m.playlistManager.GetCurrentIndex()
	if err != nil {
		return err
	}

	if currentIndex < 0 || currentIndex >= len(playlist) {
		// 如果当前索引无效，播放第一首
		currentIndex = 0
		m.playlistManager.PlayIndex(0)
	}

	currentPath := playlist[currentIndex]
	return m.audioPlayer.Play(currentPath)
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
	// 检查是否正在播放
	_, err := m.audioPlayer.IsPlaying()
	if err != nil {
		// 如果没有正在播放的音乐，尝试播放
		playlist, _ := m.playlistManager.GetPlaylist()
		if len(playlist) > 0 {
			currentIndex, _ := m.playlistManager.GetCurrentIndex()
			if currentIndex < 0 {
				m.playlistManager.PlayIndex(0)
			}
			err := m.Play()
			return true, err
		}
		return false, err
	}

	return m.audioPlayer.TogglePlayPause()
}

// Next 播放下一首
func (m *MusicService) Next() error {
	err := m.playlistManager.Next()
	if err != nil {
		return err
	}
	return m.Play()
}

// Previous 播放上一首
func (m *MusicService) Previous() error {
	err := m.playlistManager.Previous()
	if err != nil {
		return err
	}
	return m.Play()
}

// PlayIndex 播放指定索引的歌曲
func (m *MusicService) PlayIndex(index int) error {
	err := m.playlistManager.PlayIndex(index)
	if err != nil {
		return err
	}
	return m.Play()
}

// SetVolume 设置音量
func (m *MusicService) SetVolume(volume float64) error {
	return m.audioPlayer.SetVolume(volume)
}

// GetVolume 获取音量
func (m *MusicService) GetVolume() (float64, error) {
	return m.audioPlayer.GetVolume()
}

// SetPlayMode 设置播放模式
func (m *MusicService) SetPlayMode(mode string) error {
	return m.playlistManager.SetPlayMode(mode)
}

// GetPlayMode 获取播放模式
func (m *MusicService) GetPlayMode() (string, error) {
	return m.playlistManager.GetPlayMode()
}

// IsPlaying 检查是否正在播放
func (m *MusicService) IsPlaying() (bool, error) {
	return m.audioPlayer.IsPlaying()
}

// ===== 播放列表方法 =====

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

// ===== 音乐库管理方法 =====

// AddLibrary 添加目录到音乐库（带对话框）
func (m *MusicService) AddLibrary() error {
	// 打开目录选择对话框
	if m.app == nil {
		return fmt.Errorf("app not initialized")
	}

	options := &application.OpenFileDialogOptions{
		CanChooseFiles:       false,
		CanChooseDirectories: true,
		Title:                "选择音乐文件夹",
	}

	dialog := m.app.Dialog.OpenFileWithOptions(options)
	dirPath, err := dialog.PromptForSingleSelection()
	if err != nil {
		return fmt.Errorf("选择目录失败：%w", err)
	}

	if dirPath == "" {
		return nil // 用户取消选择
	}

	// 使用目录名称作为库名称
	libName := filepath.Base(dirPath)
	return m.libraryManager.AddLibrary(libName, dirPath)
}

// GetCurrentLibrary 获取当前音乐库
func (m *MusicService) GetCurrentLibrary() *MusicLibrary {
	return m.libraryManager.GetCurrentLibrary()
}

// SwitchLibrary 切换音乐库
func (m *MusicService) SwitchLibrary(name string) error {
	return m.libraryManager.SwitchLibrary(name)
}

// RefreshLibrary 刷新当前音乐库
func (m *MusicService) RefreshLibrary() error {
	return m.libraryManager.RefreshLibrary()
}

// RenameLibrary 重命名音乐库
func (m *MusicService) RenameLibrary(newName string) error {
	return m.libraryManager.RenameLibrary(newName)
}

// GetLibraries 获取所有音乐库名称列表
func (m *MusicService) GetLibraries() []string {
	libraries := m.libraryManager.GetAllLibraries()
	names := make([]string, len(libraries))
	for i, lib := range libraries {
		names[i] = lib.Name
	}
	return names
}

// SetCurrentLibrary 设置当前音乐库
func (m *MusicService) SetCurrentLibrary(name string) error {
	return m.libraryManager.SwitchLibrary(name)
}

// GetCurrentLibraryTracks 获取当前音乐库的所有音轨路径
func (m *MusicService) GetCurrentLibraryTracks() ([]string, error) {
	return m.libraryManager.GetCurrentLibraryTracks()
}

// AddToLibrary 添加目录到音乐库（指定路径）
func (m *MusicService) AddToLibrary(dirPath string) error {
	libName := filepath.Base(dirPath)
	return m.libraryManager.AddLibrary(libName, dirPath)
}

// LoadCurrentLibrary 加载当前音乐库到播放列表并播放
func (m *MusicService) LoadCurrentLibrary() error {
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

	currentLib := m.libraryManager.GetCurrentLibrary()
	if currentLib != nil {
		log.Printf("已加载音乐库 %s 到播放列表，共 %d 首歌曲", currentLib.Name, len(tracks))
	}

	return nil
}

// ===== 辅助方法 =====

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
	m.audioPlayer.Stop()
	return nil
}
