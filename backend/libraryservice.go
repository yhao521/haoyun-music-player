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
	"github.com/yhao521/wailsMusicPlay/backend/pkg/file"
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
	Path      string `json:"path"`       // 歌曲路径
	Filename  string `json:"filename"`   // 文件名
	Title     string `json:"title"`      // 标题
	Artist    string `json:"artist"`     // 艺术家
	Album     string `json:"album"`      // 专辑
	Duration  int64  `json:"duration"`   // 秒
	Size      int64  `json:"size"`       // 字节
	LyricPath string `json:"lyric_path"` // 歌词文件路径（如果有）
}

// LibraryManager 音乐库管理器
type LibraryManager struct {
	ctx        context.Context
	app        *application.App
	libraries  map[string]*MusicLibrary
	currentLib string
	mu         sync.RWMutex
	libDir     string
}

// NewLibraryManager 创建音乐库管理器
func NewLibraryManager() *LibraryManager {
	return &LibraryManager{
		libraries: make(map[string]*MusicLibrary),
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

	// 加载已有的音乐库
	return lm.LoadAllLibraries()
}

// LoadAllLibraries 加载所有音乐库
func (lm *LibraryManager) LoadAllLibraries() error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

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
			log.Printf("✓ 加载音乐库：%s (%d 首歌曲)", libName, len(lib.Tracks))
		}
	}

	// 如果没有音乐库，创建一个默认的
	if len(lm.libraries) == 0 {
		defaultLib := &MusicLibrary{
			Name:      "music",
			Path:      "",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Tracks:    make([]TrackInfo, 0),
		}
		lm.libraries["music"] = defaultLib
		lm.currentLib = "music"
		if err := lm.saveLibrary(defaultLib); err != nil {
			log.Printf("保存默认音乐库失败：%v", err)
		}
		log.Println("✓ 创建默认音乐库：music")
	} else {
		// 设置第一个库为当前库
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

	data, err := json.MarshalIndent(lib, "", "  ")
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

	// 设置为当前库
	lm.currentLib = name

	log.Printf("✓ 添加音乐库：%s (路径：%s, 歌曲数：%d)", name, path, len(tracks))

	// 发送事件通知
	if lm.app != nil {
		lm.app.Event.Emit("libraryUpdated", lib)
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

	// 重新扫描目录
	tracks, err := lm.scanDirectory(lib.Path)
	if err != nil {
		return fmt.Errorf("扫描目录失败：%w", err)
	}

	lib.Tracks = tracks
	lib.UpdatedAt = time.Now()

	// 保存到 JSON 文件
	if err := lm.saveLibrary(lib); err != nil {
		return fmt.Errorf("保存音乐库失败：%w", err)
	}

	log.Printf("✓ 刷新音乐库：%s (歌曲数：%d)", lib.Name, len(tracks))

	// 发送事件通知
	if lm.app != nil {
		lm.app.Event.Emit("libraryUpdated", lib)
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
	if lm.app != nil {
		lm.app.Event.Emit("libraryUpdated", lib)
	}

	return nil
}

// scanDirectory 扫描目录中的音乐文件和歌词文件
func (lm *LibraryManager) scanDirectory(dirPath string) ([]TrackInfo, error) {
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
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".lrc" {
			baseName := strings.TrimSuffix(info.Name(), ext)
			lyricMap[baseName] = path
		}
		return nil
	})

	if err != nil {
		log.Printf("⚠️ 扫描歌词文件失败：%v", err)
	}

	// 扫描音乐文件
	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 跳过无法访问的文件
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if supportedFormats[ext] {
			baseName := strings.TrimSuffix(info.Name(), ext)
			
			// 查找对应的歌词文件
			lyricPath := ""
			if lrcPath, ok := lyricMap[baseName]; ok {
				lyricPath = lrcPath
			}

			track := TrackInfo{
				Path:      path,
				Filename:  info.Name(),
				Title:     baseName,
				Artist:    "未知艺术家",
				Album:     "未知专辑",
				Duration:  0, // TODO: 从音频文件中读取
				Size:      info.Size(),
				LyricPath: lyricPath, // 保存歌词路径
			}
			tracks = append(tracks, track)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	log.Printf("✓ 扫描完成：找到 %d 首歌曲，%d 个歌词文件", len(tracks), len(lyricMap))
	return tracks, nil
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
