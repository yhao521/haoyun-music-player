package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/yhao521/haoyun-music-player/backend/pkg/file"
)

// MusicLibrary 音乐库结构
type MusicLibrary struct {
	Name      string      `json:"name"`
	Path      string      `json:"path"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	Tracks    []TrackInfo `json:"tracks"`
}

// TrackInfo 音乐文件信息
type TrackInfo struct {
	Path      string `json:"path"`                 // 歌曲路径
	Filename  string `json:"filename"`             // 文件名
	Title     string `json:"title,omitempty"`      // 标题
	Artist    string `json:"artist,omitempty"`     // 艺术家
	Album     string `json:"album,omitempty"`      // 专辑
	Duration  int64  `json:"duration"`             // 秒
	Size      int64  `json:"size"`                 // 字节
	LyricPath string `json:"lyric_path,omitempty"` // 歌词文件路径（如果有）
}

// LibraryManager 音乐库管理器
type LibraryManager struct {
	ctx             context.Context
	app             *application.App
	libraries       map[string]*MusicLibrary
	currentLib      string
	mu              sync.RWMutex
	libDir          string
	lyricDir        string                // 歌词目录
	metadataManager *MetadataManager      // 元数据管理器
	tracksByPath    map[string]*TrackInfo // 路径索引：path -> TrackInfo，用于O(1)查找
}

// NewLibraryManager 创建音乐库管理器
func NewLibraryManager() *LibraryManager {
	return &LibraryManager{
		libraries:       make(map[string]*MusicLibrary),
		metadataManager: NewMetadataManager(),
	}
}

// SetApp 设置应用实例
func (lm *LibraryManager) SetApp(app *application.App) {
	lm.app = app
}

// SetContext 设置上下文
func (lm *LibraryManager) SetContext(ctx context.Context) {
	lm.ctx = ctx
}

// Init 初始化音乐库管理器
func (lm *LibraryManager) Init() error {
	// 创建音乐库目录
	lm.libDir = file.GetLibPath()
	if err := os.MkdirAll(lm.libDir, 0755); err != nil {
		return fmt.Errorf("创建音乐库目录失败：%w", err)
	}

	// 设置歌词目录
	lm.lyricDir = filepath.Join(lm.libDir, "lyrics")
	if err := os.MkdirAll(lm.lyricDir, 0755); err != nil {
		return fmt.Errorf("创建歌词目录失败：%w", err)
	}

	// 加载已有的音乐库
	return lm.LoadAllLibraries()
}

// LoadAllLibraries 加载所有音乐库
func (lm *LibraryManager) LoadAllLibraries() error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	// 初始化路径索引
	lm.tracksByPath = make(map[string]*TrackInfo)

	files, err := os.ReadDir(lm.libDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			libName := strings.TrimSuffix(file.Name(), ".json")
			lib, err := lm.loadLibrary(libName)
			if err != nil {
				log.Printf("加载音乐库 %s 失败：%v\n", libName, err)
				continue
			}
			lm.libraries[libName] = lib

			// 构建该音乐库的路径索引
			lm.buildTracksIndexForLibrary(lib)

			log.Printf("✓ 加载音乐库：%s (%d 首歌曲)", libName, len(lib.Tracks))
		}
	}

	// 设置第一个库为当前库（如果没有音乐库，currentLib 保持为空字符串）
	if len(lm.libraries) > 0 {
		for name := range lm.libraries {
			lm.currentLib = name
			break
		}
	}

	return nil
}

// loadLibrary 从 JSON 文件加载音乐库
func (lm *LibraryManager) loadLibrary(name string) (*MusicLibrary, error) {
	filePath := filepath.Join(lm.libDir, name+".json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var lib MusicLibrary
	if err := json.Unmarshal(data, &lib); err != nil {
		return nil, err
	}

	return &lib, nil
}

// saveLibrary 保存音乐库到 JSON 文件
func (lm *LibraryManager) saveLibrary(lib *MusicLibrary) error {
	lib.UpdatedAt = time.Now()

	// 使用紧凑格式减少文件大小（移除缩进和换行）
	data, err := json.Marshal(lib)
	if err != nil {
		return err
	}

	filePath := filepath.Join(lm.libDir, lib.Name+".json")
	return os.WriteFile(filePath, data, 0644)
}

// AddLibrary 添加新的音乐库
func (lm *LibraryManager) AddLibrary(name, path string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	// 检查名称是否已存在
	if _, exists := lm.libraries[name]; exists {
		return fmt.Errorf("音乐库 %s 已存在", name)
	}

	// 创建新的音乐库对象
	lib := &MusicLibrary{
		Name:      name,
		Path:      path,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Tracks:    make([]TrackInfo, 0),
	}

	// 扫描目录中的音乐文件
	tracks, err := lm.scanDirectory(path)
	if err != nil {
		return fmt.Errorf("扫描目录失败：%w", err)
	}
	lib.Tracks = tracks

	// 保存到 JSON 文件
	if err := lm.saveLibrary(lib); err != nil {
		return fmt.Errorf("保存音乐库失败：%w", err)
	}

	// 添加到 libraries map
	lm.libraries[name] = lib

	// 构建该音乐库的路径索引
	lm.buildTracksIndexForLibrary(lib)

	// 设置为当前库
	lm.currentLib = name

	log.Printf("✓ 添加音乐库：%s (路径：%s, 歌曲数：%d)", name, path, len(tracks))

	// 发送事件通知
	if lm.app != nil && lm.app.Event != nil {
		lib := lm.libraries[name]
		lm.app.Event.Emit("libraryUpdated", lib)
	} else if lm.app != nil {
		log.Printf("[LibraryManager] 警告: app.Event 为 nil，跳过事件发送")
	}

	return nil
}

// RemoveLibrary 删除音乐库
func (lm *LibraryManager) RemoveLibrary(name string) error {
	return lm.DeleteLibrary(name)
}

// DeleteLibrary 删除音乐库（仅删除配置，不删除文件）
func (lm *LibraryManager) DeleteLibrary(name string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if _, exists := lm.libraries[name]; !exists {
		return fmt.Errorf("音乐库 %s 不存在", name)
	}

	// 删除 JSON 文件
	filePath := filepath.Join(lm.libDir, name+".json")
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("删除音乐库文件失败：%w", err)
	}

	// 从 map 中移除
	delete(lm.libraries, name)

	// 如果删除的是当前库，切换到另一个库或清空
	if lm.currentLib == name {
		lm.currentLib = ""
		for libName := range lm.libraries {
			lm.currentLib = libName
			break
		}
	}

	log.Printf("✓ 已删除音乐库：%s", name)
	return nil
}

// LibraryExists 检查音乐库是否存在
func (lm *LibraryManager) LibraryExists(name string) bool {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	_, exists := lm.libraries[name]
	return exists
}

// SwitchLibrary 切换音乐库
func (lm *LibraryManager) SwitchLibrary(name string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if _, exists := lm.libraries[name]; !exists {
		return fmt.Errorf("音乐库 %s 不存在", name)
	}

	lm.currentLib = name
	log.Printf("✓ 切换到音乐库：%s", name)

	// 发送事件通知
	if lm.app != nil {
		lib := lm.libraries[name]
		lm.app.Event.Emit("libraryUpdated", lib)
	}

	return nil
}

// GetCurrentLibrary 获取当前音乐库
func (lm *LibraryManager) GetCurrentLibrary() *MusicLibrary {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	if lm.currentLib == "" {
		return nil
	}

	return lm.libraries[lm.currentLib]
}

// GetAllLibraries 获取所有音乐库
func (lm *LibraryManager) GetAllLibraries() []*MusicLibrary {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	log.Println(lm.libraries)

	libraries := make([]*MusicLibrary, 0, len(lm.libraries))
	for _, lib := range lm.libraries {
		libraries = append(libraries, lib)
	}
	return libraries
}

// RefreshLibrary 刷新当前音乐库
func (lm *LibraryManager) RefreshLibrary() error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if lm.currentLib == "" {
		return fmt.Errorf("当前没有音乐库")
	}

	lib := lm.libraries[lm.currentLib]

	// 清除元数据缓存，确保重新读取元数据
	if lm.metadataManager != nil {
		lm.metadataManager.ClearCache()
		log.Printf("🗑️ 已清除元数据缓存")
	}

	// 重新扫描目录
	tracks, err := lm.scanDirectory(lib.Path)
	if err != nil {
		return fmt.Errorf("扫描目录失败：%w", err)
	}

	lib.Tracks = tracks
	lib.UpdatedAt = time.Now()

	// 重建该音乐库的路径索引
	lm.buildTracksIndexForLibrary(lib)

	// 保存到 JSON 文件
	if err := lm.saveLibrary(lib); err != nil {
		return fmt.Errorf("保存音乐库失败：%w", err)
	}

	log.Printf("✓ 刷新音乐库：%s (歌曲数：%d)", lib.Name, len(tracks))

	// 发送事件通知
	if lm.app != nil && lm.app.Event != nil {
		lm.app.Event.Emit("libraryUpdated", lib)
	} else if lm.app != nil {
		log.Printf("[LibraryManager] 警告: app.Event 为 nil，跳过事件发送")
	}

	return nil
}

// RenameLibrary 重命名音乐库
func (lm *LibraryManager) RenameLibrary(newName string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if lm.currentLib == "" {
		return fmt.Errorf("当前没有音乐库")
	}

	// 检查新名称是否已存在
	if _, exists := lm.libraries[newName]; exists {
		return fmt.Errorf("音乐库 %s 已存在", newName)
	}

	oldName := lm.currentLib
	lib := lm.libraries[oldName]

	// 删除旧文件
	oldPath := filepath.Join(lm.libDir, oldName+".json")
	if err := os.Remove(oldPath); err != nil {
		return fmt.Errorf("删除旧音乐库文件失败：%w", err)
	}

	// 更新库名称
	lib.Name = newName
	lib.UpdatedAt = time.Now()

	// 保存为新文件
	if err := lm.saveLibrary(lib); err != nil {
		return fmt.Errorf("保存音乐库失败：%w", err)
	}

	// 更新 map
	delete(lm.libraries, oldName)
	lm.libraries[newName] = lib
	lm.currentLib = newName

	log.Printf("✓ 重命名音乐库：%s -> %s", oldName, newName)

	// 发送事件通知
	if lm.app != nil && lm.app.Event != nil {
		lm.app.Event.Emit("libraryUpdated", lib)
	} else if lm.app != nil {
		log.Printf("[LibraryManager] 警告: app.Event 为 nil，跳过事件发送")
	}

	return nil
}

// scanDirectory 扫描目录中的音乐文件和歌词文件（使用元数据）
func (lm *LibraryManager) scanDirectory(dirPath string) ([]TrackInfo, error) {
	// 使用新的带元数据的扫描方法
	return lm.scanDirectoryWithMetadata(dirPath)
}

// buildTracksIndexForLibrary 为指定音乐库构建路径索引（必须在持有锁的情况下调用）
func (lm *LibraryManager) buildTracksIndexForLibrary(lib *MusicLibrary) {
	if lm.tracksByPath == nil {
		lm.tracksByPath = make(map[string]*TrackInfo)
	}

	// 清除该音乐库的旧索引
	for path := range lm.tracksByPath {
		// 简单策略：清空所有索引后重建
		// 更精细的策略可以只删除属于该音乐库的路径
		delete(lm.tracksByPath, path)
	}

	// 重新构建索引
	for i := range lib.Tracks {
		lm.tracksByPath[lib.Tracks[i].Path] = &lib.Tracks[i]
	}
}

// GetTrackByPath 通过路径快速获取 TrackInfo（O(1) 时间复杂度）
func (lm *LibraryManager) GetTrackByPath(path string) *TrackInfo {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	if lm.tracksByPath == nil {
		return nil
	}

	return lm.tracksByPath[path]
}

// GetCurrentLibraryTracks 获取当前音乐库的所有音轨
func (lm *LibraryManager) GetCurrentLibraryTracks() ([]string, error) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	if lm.currentLib == "" {
		return []string{}, nil
	}

	lib := lm.libraries[lm.currentLib]
	trackPaths := make([]string, len(lib.Tracks))
	for i, track := range lib.Tracks {
		trackPaths[i] = track.Path
	}

	return trackPaths, nil
}

// SetCurrentLibrary 设置当前音乐库（原子操作）
func (lm *LibraryManager) SetCurrentLibrary(name string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.currentLib = name
}

// GetTrackMetadata 获取单个音轨的元数据
func (lm *LibraryManager) GetTrackMetadata(trackPath string) (*TrackInfo, error) {
	if lm.metadataManager == nil {
		return nil, fmt.Errorf("元数据管理器未初始化")
	}

	// 获取文件信息
	fileInfo, err := os.Stat(trackPath)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败：%w", err)
	}

	// 从元数据管理器读取元数据
	metadata, err := lm.metadataManager.GetMetadata(trackPath)
	if err != nil {
		log.Printf("⚠️ 读取元数据失败 %s：%v，使用基本信息", trackPath, err)
		// 即使失败也继续，使用基本信息
	}

	// 查找对应的歌词文件
	lyricPath := lm.findLyricFile(trackPath)

	// 构建 TrackInfo
	baseName := strings.TrimSuffix(filepath.Base(trackPath), filepath.Ext(trackPath))

	// 从元数据中获取时长
	duration := int64(0)
	if dur, ok := metadata["duration"].(int64); ok {
		duration = dur
	}

	track := &TrackInfo{
		Path:      trackPath,
		Filename:  fileInfo.Name(),
		Title:     getStringFromMetadata(metadata, "title", baseName),
		Artist:    getStringFromMetadata(metadata, "artist", "未知艺术家"),
		Album:     getStringFromMetadata(metadata, "album", "未知专辑"),
		Duration:  duration, // 从元数据中读取时长
		Size:      fileInfo.Size(),
		LyricPath: lyricPath,
	}

	return track, nil
}

// findLyricFile 查找歌词文件
func (lm *LibraryManager) findLyricFile(trackPath string) string {
	baseName := strings.TrimSuffix(filepath.Base(trackPath), filepath.Ext(trackPath))
	dirPath := filepath.Dir(trackPath)

	// 常见的歌词文件扩展名
	lyricExts := []string{".lrc", ".txt"}

	// 策略 1: 同目录下的歌词文件
	for _, ext := range lyricExts {
		lyricPath := filepath.Join(dirPath, baseName+ext)
		if _, err := os.Stat(lyricPath); err == nil {
			return lyricPath
		}
	}

	// 策略 2: 全局歌词目录 (lm.lyricDir)
	if lm.lyricDir != "" {
		for _, ext := range lyricExts {
			lyricPath := filepath.Join(lm.lyricDir, baseName+ext)
			if _, err := os.Stat(lyricPath); err == nil {
				return lyricPath
			}
		}
	}

	return ""
}

// getStringFromMetadata 从元数据中安全地获取字符串值
func getStringFromMetadata(metadata map[string]interface{}, key string, defaultValue string) string {
	if metadata == nil {
		return defaultValue
	}

	if value, ok := metadata[key].(string); ok && value != "" {
		return value
	}

	return defaultValue
}

// scanDirectoryWithMetadata 扫描目录并获取完整的元数据
func (lm *LibraryManager) scanDirectoryWithMetadata(dirPath string) ([]TrackInfo, error) {
	var tracks []TrackInfo

	if dirPath == "" {
		return tracks, nil
	}

	// 支持的音乐格式
	supportedFormats := map[string]bool{
		".mp3":  true,
		".wav":  true,
		".flac": true,
		".aac":  true,
		".ogg":  true,
		".wma":  true,
	}

	// 首先扫描所有歌词文件，建立映射表
	lyricMap := make(map[string]string) // 歌曲名(不含扩展名) -> 歌词路径
	
	// 扫描音乐库目录中的歌词文件
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".lrc" || ext == ".txt" {
			baseName := strings.TrimSuffix(info.Name(), ext)
			lyricMap[baseName] = path
		}
		return nil
	})

	if err != nil {
		log.Printf("⚠️ 扫描音乐库目录歌词文件失败：%v", err)
	}

	// 扫描全局歌词目录 (lm.lyricDir) 中的歌词文件
	if lm.lyricDir != "" && lm.lyricDir != dirPath {
		err = filepath.Walk(lm.lyricDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".lrc" || ext == ".txt" {
				baseName := strings.TrimSuffix(info.Name(), ext)
				// 如果音乐库目录中已有同名歌词，优先使用音乐库目录的（向后兼容）
				if _, exists := lyricMap[baseName]; !exists {
					lyricMap[baseName] = path
				}
			}
			return nil
		})

		if err != nil {
			log.Printf("⚠️ 扫描全局歌词目录失败：%v", err)
		} else {
			log.Printf("📝 从全局歌词目录加载了 %d 个歌词文件", len(lyricMap))
		}
	}

	// 收集所有音频文件路径
	var audioFiles []string
	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if supportedFormats[ext] {
			audioFiles = append(audioFiles, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	log.Printf("🔍 开始扫描 %d 个音频文件...", len(audioFiles))

	// 逐个获取元数据
	successCount := 0
	for i, audioPath := range audioFiles {
		track, err := lm.GetTrackMetadata(audioPath)
		if err != nil {
			log.Printf("⚠️ 处理文件 %d/%d 失败：%s - %v", i+1, len(audioFiles), audioPath, err)
			continue
		}

		// 如果元数据中没有歌词路径，尝试从映射表中查找
		if track.LyricPath == "" {
			baseName := strings.TrimSuffix(filepath.Base(audioPath), filepath.Ext(audioPath))
			if lrcPath, ok := lyricMap[baseName]; ok {
				track.LyricPath = lrcPath
			}
		}

		tracks = append(tracks, *track)
		successCount++

		// 每处理 50 个文件输出一次进度
		if (i+1)%50 == 0 || i+1 == len(audioFiles) {
			log.Printf("📊 进度：%d/%d (%.1f%%)", i+1, len(audioFiles), float64(i+1)/float64(len(audioFiles))*100)
		}
	}

	log.Printf("✓ 扫描完成：成功处理 %d/%d 首歌曲，找到 %d 个歌词文件",
		successCount, len(audioFiles), len(lyricMap))

	return tracks, nil
}

// CompactLibraries 压缩所有音乐库文件（移除空字段和多余空白）
func (lm *LibraryManager) CompactLibraries() (int, error) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	compactedCount := 0

	for name, lib := range lm.libraries {
		log.Printf("🗜️ 正在压缩音乐库：%s...", name)

		// 重新保存以应用新的紧凑格式
		if err := lm.saveLibrary(lib); err != nil {
			log.Printf("⚠️ 压缩音乐库 %s 失败：%v", name, err)
			continue
		}

		compactedCount++
		log.Printf("✓ 音乐库 %s 压缩完成", name)
	}

	log.Printf("✓ 压缩完成：共处理 %d 个音乐库", compactedCount)
	return compactedCount, nil
}

// ReloadCurrentLibrary 重新加载当前音乐库（扫描目录更新索引）
func (lm *LibraryManager) ReloadCurrentLibrary() error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if lm.currentLib == "" {
		return fmt.Errorf("当前没有音乐库")
	}

	lib, exists := lm.libraries[lm.currentLib]
	if !exists {
		return fmt.Errorf("音乐库 %s 不存在", lm.currentLib)
	}

	if lib.Path == "" {
		return fmt.Errorf("音乐库路径为空")
	}

	log.Printf("🔄 重新扫描音乐库：%s (路径：%s)", lib.Name, lib.Path)

	// 清除旧的路径索引
	lm.clearTracksIndexForLibrary(lib)

	// 重新扫描目录
	tracks, err := lm.scanDirectory(lib.Path)
	if err != nil {
		return fmt.Errorf("扫描目录失败：%w", err)
	}

	// 更新音乐库的音轨列表
	lib.Tracks = tracks
	lib.UpdatedAt = time.Now()

	// 构建新的路径索引
	lm.buildTracksIndexForLibrary(lib)

	// 保存到 JSON 文件
	if err := lm.saveLibrary(lib); err != nil {
		return fmt.Errorf("保存音乐库失败：%w", err)
	}

	log.Printf("✓ 音乐库 %s 重新加载完成，共 %d 首歌曲", lib.Name, len(tracks))
	return nil
}

// clearTracksIndexForLibrary 清除音乐库的路径索引
func (lm *LibraryManager) clearTracksIndexForLibrary(lib *MusicLibrary) {
	for _, track := range lib.Tracks {
		delete(lm.tracksByPath, track.Path)
	}
}

// GetMetadataManager 获取元数据管理器
func (lm *LibraryManager) GetMetadataManager() *MetadataManager {
	return lm.metadataManager
}
