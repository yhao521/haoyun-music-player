package backend

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// MusicService 音乐服务统一接口（MVC Model 层）
type MusicService struct {
	ctx             context.Context
	app             *application.App
	audioPlayer     *AudioPlayer      // beep 音频播放器
	playlistManager *PlaylistManager  // 播放列表管理
	libraryManager  *LibraryManager   // 音乐库管理
	organizeService *OrganizeService  // 整理音乐服务
	historyManager  *HistoryManager   // 播放历史管理
	lyricManager    *LyricManager     // 歌词管理
	coverManager    *CoverManager     // 专辑封面管理
	metadataManager *MetadataManager  // 元数据管理器
}

// NewMusicService 创建音乐服务实例
func NewMusicService() *MusicService {
	return &MusicService{
		audioPlayer:     NewAudioPlayer(),
		playlistManager: NewPlaylistManager(),
		libraryManager:  NewLibraryManager(),
		organizeService: NewOrganizeService(),
		historyManager:  NewHistoryManager(),
		lyricManager:    NewLyricManager(),
		coverManager:    NewCoverManager(),
		metadataManager: NewMetadataManager(),
	}
}

// SetApp 设置应用实例
func (m *MusicService) SetApp(app *application.App) {
	m.app = app
	m.audioPlayer.SetApp(app)
	m.playlistManager.SetApp(app)
	m.libraryManager.SetApp(app)
	m.organizeService.SetLibraryManager(m.libraryManager)
	m.historyManager.SetApp(app)
	
	// 设置 PlaylistManager 的 LibraryManager 引用，使其能够获取元数据
	m.playlistManager.SetLibraryManager(m.libraryManager)
	
	// 监听播放结束事件，根据播放模式决定是否自动播放下一首
	app.Event.On("playbackEnded", func(event *application.CustomEvent) {
		log.Println("🎵 收到 playbackEnded 事件，检查是否需要自动播放下一首")
		
		// 获取当前播放模式
		playMode, err := m.playlistManager.GetPlayMode()
		if err != nil {
			log.Printf("⚠️ 获取播放模式失败：%v", err)
			return
		}
		
		log.Printf("当前播放模式：%s", playMode)
		
		// 单曲循环模式下不自动播放下一首（保持当前歌曲）
		if playMode == "single" {
			log.Println("🔂 单曲循环模式，重新播放当前歌曲")
			// 重新播放当前歌曲
			if err := m.Play(); err != nil {
				log.Printf("❌ 重新播放失败：%v", err)
			}
			return
		}
		
		// order 模式：顺序播放完最后一首后停止
		if playMode == "order" {
			currentIndex, _ := m.playlistManager.GetCurrentIndex()
			playlistLen := len(m.getPlaylistUnsafe())
			if currentIndex >= playlistLen-1 {
				log.Println("🔢 顺序播放模式，已到达列表末尾，停止播放")
				return // 不自动播放下一首
			}
		}
		
		// loop 和 random 模式：自动播放下一首
		log.Printf("🔄 %s 模式，自动播放下一首", playMode)
		if err := m.Next(); err != nil {
			log.Printf("❌ 自动播放下一首失败：%v", err)
		}
	})
}

// getPlaylistUnsafe 获取播放列表（不加锁，仅在内部使用）
func (m *MusicService) getPlaylistUnsafe() []string {
	playlist, _ := m.playlistManager.GetPlaylist()
	return playlist
}

// SetContext 设置上下文
func (m *MusicService) SetContext(ctx context.Context) {
	m.ctx = ctx
}

// Init 初始化服务
func (m *MusicService) Init() error {
	// 检查 FFmpeg 可用性
	if ffmpegPath, err := findFFmpegPath(); err != nil {
		log.Printf("⚠️ FFmpeg 未找到：%v", err)
		log.Println("⚠️ 部分音频格式可能无法播放，请安装 FFmpeg")
		log.Println("📖 查看安装指南：FFMPEG_GUIDE.md")
	} else {
		log.Printf("✓ FFmpeg 已就绪: %s", ffmpegPath)
	}

	// 初始化音乐库管理器
	if err := m.libraryManager.Init(); err != nil {
		return fmt.Errorf("初始化音乐库失败：%w", err)
	}

	// 初始化播放历史管理器
	if err := m.historyManager.Init(); err != nil {
		log.Printf("⚠️ 初始化播放历史失败：%v", err)
	}

	// 初始化歌词管理器
	if err := m.lyricManager.Init(); err != nil {
		log.Printf("⚠️ 初始化歌词管理器失败：%v", err)
	}

	// 初始化专辑封面管理器
	if err := m.coverManager.Init(); err != nil {
		log.Printf("⚠️ 初始化专辑封面管理器失败：%v", err)
	}

	log.Println("✓ 所有服务初始化完成")
	return nil
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
	
	// 异步记录播放历史（使用音乐库获取完整元数据）
	go func() {
		trackInfo := createTrackInfoFromLibrary(currentPath, m.libraryManager)
		m.historyManager.AddToHistory(trackInfo)
	}()
	
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
	log.Println("[MusicService.TogglePlayPause] 开始")
	
	// 检查是否正在播放
	isPlaying, err := m.audioPlayer.IsPlaying()
	if err != nil {
		log.Printf("[MusicService.TogglePlayPause] IsPlaying 出错：%v", err)
		// 如果出错，尝试获取播放列表并播放
		playlist, _ := m.playlistManager.GetPlaylist()
		if len(playlist) > 0 {
			currentIndex, _ := m.playlistManager.GetCurrentIndex()
			if currentIndex < 0 {
				m.playlistManager.PlayIndex(0)
			}
			err := m.Play()
			log.Printf("[MusicService.TogglePlayPause] 播放结果：%v", err)
			return true, err
		}
		return false, err
	}
	
	log.Printf("[MusicService.TogglePlayPause] isPlaying: %v", isPlaying)

	// 如果正在播放，则暂停；否则播放
	if isPlaying {
		log.Println("[MusicService.TogglePlayPause] 执行暂停")
		err := m.audioPlayer.Pause()
		if err != nil {
			log.Printf("[MusicService.TogglePlayPause] 暂停失败：%v", err)
			return false, err
		}
		log.Println("[MusicService.TogglePlayPause] 暂停成功")
		return false, nil
	} else {
		log.Println("[MusicService.TogglePlayPause] 尝试恢复播放")
		// 已暂停，尝试恢复播放
		success, err := m.audioPlayer.TogglePlayPause()
		if err != nil {
			log.Printf("[MusicService.TogglePlayPause] TogglePlayPause 失败：%v", err)
			// 如果播放器已停止，重新播放当前歌曲
			playlist, _ := m.playlistManager.GetPlaylist()
			if len(playlist) > 0 {
				currentIndex, _ := m.playlistManager.GetCurrentIndex()
				if currentIndex >= 0 && currentIndex < len(playlist) {
					err := m.Play()
					log.Printf("[MusicService.TogglePlayPause] 重新播放结果：%v", err)
					return true, err
				}
			}
			return false, err
		}
		log.Printf("[MusicService.TogglePlayPause] 恢复播放成功：%v", success)
		return true, nil
	}
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

// GetCurrentTrack 获取当前播放的歌曲路径
func (m *MusicService) GetCurrentTrack() (string, error) {
	playlist, err := m.GetPlaylist()
	if err != nil {
		return "", err
	}

	index, err := m.playlistManager.GetCurrentIndex()
	if err != nil {
		return "", err
	}

	if index < 0 || index >= len(playlist) {
		return "", fmt.Errorf("当前播放索引无效：%d", index)
	}

	return playlist[index], nil
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
	
	// 如果名称已存在，添加时间戳后缀
	if m.libraryManager.LibraryExists(libName) {
		timestamp := time.Now().Format("20060102_150405")
		libName = fmt.Sprintf("%s_%s", libName, timestamp)
		log.Printf("⚠️ 音乐库 '%s' 已存在，使用新名称: %s", filepath.Base(dirPath), libName)
	}
	
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

// DeleteLibrary 删除音乐库
func (m *MusicService) DeleteLibrary(name string) error {
	return m.libraryManager.DeleteLibrary(name)
}

// GetCurrentLibraryTracks 获取当前音乐库的所有音轨
func (m *MusicService) GetCurrentLibraryTracks() ([]string, error) {
	tracks, err := m.libraryManager.GetCurrentLibraryTracks()
	if err != nil {
		return nil, err
	}
	result := make([]string, len(tracks))
	for i, track := range tracks {
		result[i] = track // track 已经是 string 类型（Path）
	}
	return result, nil
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

	// 清空当前播放列表（发送一次事件）
	m.ClearPlaylist()

	// 批量添加所有音轨到播放列表（只发送一次事件）
	if err := m.playlistManager.AddToPlaylistBatch(tracks); err != nil {
		log.Printf("批量添加音轨失败：%v", err)
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

// GetCurrentTrackName 获取当前播放的歌曲名称（仅文件名）
func (m *MusicService) GetCurrentTrackName() (string, error) {
	path, err := m.GetCurrentTrack()
	if err != nil {
		return "", err
	}

	// 从路径中提取文件名
	filename := filepath.Base(path)
	return filename, nil
}

// GetSongMetadata 获取歌曲元数据（优先使用音乐库扫描结果）
func (m *MusicService) GetSongMetadata(path string) (map[string]interface{}, error) {
	// 策略 1: 尝试从当前音乐库中获取已扫描的元数据
	if m.libraryManager != nil {
		currentLib := m.libraryManager.GetCurrentLibrary()
		if currentLib != nil {
			for _, track := range currentLib.Tracks {
				if track.Path == path {
					// 找到匹配的音轨，使用扫描时获取的元数据
					metadata := map[string]interface{}{
						"title":    track.Title,
						"artist":   track.Artist,
						"album":    track.Album,
						"duration": track.Duration,
						"path":     track.Path,
						"filename": track.Filename,
						"size":     track.Size,
						"lyric_path": track.LyricPath,
					}
					
					log.Printf("✓ 从音乐库缓存获取元数据：%s - %s", track.Artist, track.Title)
					return metadata, nil
				}
			}
		}
	}
	
	// 策略 2: 如果音乐库中没有，尝试从元数据管理器缓存中获取
	if m.metadataManager != nil {
		metadata, err := m.metadataManager.GetMetadata(path)
		if err == nil {
			log.Printf("✓ 从元数据管理器缓存获取：%s", path)
			return metadata, nil
		}
		log.Printf("⚠️ 元数据管理器读取失败：%v，使用基本信息", err)
	}
	
	// 策略 3: 降级到基本信息
	filename := filepath.Base(path)
	return map[string]interface{}{
		"title":    filename,
		"artist":   "未知艺术家",
		"album":    "未知专辑",
		"duration": int64(0),
		"path":     path,
		"filename": filename,
		"size":     int64(0),
		"lyric_path": "",
	}, nil
}

// GetTrackInfo 获取完整的音轨信息（优先从音乐库）
func (m *MusicService) GetTrackInfo(trackPath string) (*TrackInfo, error) {
	// 尝试从当前音乐库中获取
	if m.libraryManager != nil {
		currentLib := m.libraryManager.GetCurrentLibrary()
		if currentLib != nil {
			for _, track := range currentLib.Tracks {
				if track.Path == trackPath {
					log.Printf("✓ 从音乐库获取 TrackInfo：%s", track.Title)
					return &track, nil
				}
			}
		}
	}
	
	// 如果音乐库中没有，实时获取元数据
	if m.libraryManager != nil {
		track, err := m.libraryManager.GetTrackMetadata(trackPath)
		if err == nil {
			log.Printf("✓ 实时获取 TrackInfo：%s", trackPath)
			return track, nil
		}
	}
	
	// 降级：返回基本信息
	filename := filepath.Base(trackPath)
	return &TrackInfo{
		Path:      trackPath,
		Filename:  filename,
		Title:     filename,
		Artist:    "未知艺术家",
		Album:     "未知专辑",
		Duration:  0,
		Size:      0,
		LyricPath: "",
	}, nil
}

// ===== 播放历史管理方法 =====

// GetPlayHistory 获取播放历史记录
func (m *MusicService) GetPlayHistory(limit int) []HistoryRecord {
	if limit <= 0 {
		limit = 20 // 默认返回 20 条
	}
	return m.historyManager.GetHistory(limit)
}

// ClearPlayHistory 清空播放历史
func (m *MusicService) ClearPlayHistory() error {
	return m.historyManager.ClearHistory()
}

// RemoveFromPlayHistory 删除指定索引的历史记录
func (m *MusicService) RemoveFromPlayHistory(index int) error {
	return m.historyManager.RemoveFromHistory(index)
}

// GetPlayHistoryCount 获取历史记录数量
func (m *MusicService) GetPlayHistoryCount() int {
	return m.historyManager.GetHistoryCount()
}

// GetFavoriteTracks 获取喜爱音乐（按播放次数排序）
func (m *MusicService) GetFavoriteTracks(limit int) []HistoryRecord {
	if limit <= 0 {
		limit = 20 // 默认返回 20 首
	}
	return m.historyManager.GetFavoriteTracks(limit)
}

// ===== 歌词管理方法 =====

// LoadLyric 加载歌词
func (m *MusicService) LoadLyric(trackPath string) (*LyricInfo, error) {
	return m.lyricManager.LoadLyric(trackPath)
}

// GetCurrentLyricLine 获取当前歌词行
func (m *MusicService) GetCurrentLyricLine(trackPath string, position float64) (int, error) {
	return m.lyricManager.GetCurrentLyricLine(trackPath, position)
}

// GetAllLyrics 获取所有歌词
func (m *MusicService) GetAllLyrics(trackPath string) ([]LyricLine, error) {
	return m.lyricManager.GetAllLyrics(trackPath)
}

// HasLyric 检查是否有歌词
func (m *MusicService) HasLyric(trackPath string) bool {
	return m.lyricManager.HasLyric(trackPath)
}

// ===== 专辑封面管理方法 =====

// GetAlbumArtDataURL 获取专辑封面的 Data URL
func (m *MusicService) GetAlbumArtDataURL(trackPath string) (string, error) {
	return m.coverManager.GetCoverDataURL(trackPath)
}

// GetCachedCover 获取缓存的封面
func (m *MusicService) GetCachedCover(trackPath string) *AlbumArt {
	return m.coverManager.GetCachedCover(trackPath)
}

// ClearCoverCache 清除封面缓存
func (m *MusicService) ClearCoverCache() {
	m.coverManager.ClearCache()
}

// ClearMetadataCache 清除元数据缓存
func (m *MusicService) ClearMetadataCache() {
	if m.metadataManager != nil {
		m.metadataManager.ClearCache()
	}
}

// CompactLibraries 压缩所有音乐库文件（移除空字段和多余空白）
func (m *MusicService) CompactLibraries() (int, error) {
	if m.libraryManager == nil {
		return 0, fmt.Errorf("音乐库管理器未初始化")
	}
	return m.libraryManager.CompactLibraries()
}

// GetPlaylistManager 获取播放列表管理器（用于批量操作）
func (m *MusicService) GetPlaylistManager() *PlaylistManager {
	return m.playlistManager
}

// Shutdown 关闭服务
func (m *MusicService) Shutdown() error {
	m.audioPlayer.Stop()
	return nil
}

// OrganizeLibrary 整理音乐库：将音乐文件和歌词文件分别移动到子目录
func (m *MusicService) OrganizeLibrary() error {
	return m.organizeService.OrganizeLibrary()
}

// GetOrganizeService 获取整理音乐服务
func (m *MusicService) GetOrganizeService() *OrganizeService {
	return m.organizeService
}
