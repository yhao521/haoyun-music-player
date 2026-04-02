package main

import (
	"changeme/backend"
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
	Path     string `json:"path"`
	Filename string `json:"filename"`
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Album    string `json:"album"`
	Duration int64  `json:"duration"` // 秒
	Size     int64  `json:"size"`     // 字节
}

// LibraryManager 音乐库管理器
type LibraryManager struct {
	ctx context.Context
	app *application.App
	backend.Com
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
	lm.Com.SetApp(app)
}

// SetContext 设置上下文
func (lm *LibraryManager) SetContext(ctx context.Context) {
	lm.ctx = ctx
}

// Init 初始化音乐库管理器
func (lm *LibraryManager) Init() error {
	// 获取用户数据目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("获取用户目录失败：%w", err)
	}

	// 创建音乐库目录
	lm.libDir = filepath.Join(homeDir, ".haoyun-music", "libraries")
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
				fmt.Printf("加载音乐库 %s 失败：%v\n", libName, err)
				continue
			}
			lm.libraries[libName] = lib
		}
	}

	// 如果有库，设置第一个为当前库
	if len(lm.libraries) > 0 {
		for name := range lm.libraries {
			lm.currentLib = name
			break
		}
	}

	return nil
}

// loadLibrary 加载单个音乐库
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

// SaveLibrary 保存音乐库到文件
func (lm *LibraryManager) SaveLibrary(lib *MusicLibrary) error {
	filePath := filepath.Join(lm.libDir, lib.Name+".json")
	data, err := json.MarshalIndent(lib, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// AddLibrary 添加新音乐库
func (lm *LibraryManager) AddLibrary() error {
	selectedDir := lm.Com.SelectPathDownload()
	// 使用目录名作为库名
	libName := filepath.Base(selectedDir)
	log.Printf("选择目录：%s, 库名：%s", selectedDir, libName)
	name := libName
	path := selectedDir

	lm.mu.Lock()
	defer lm.mu.Unlock()

	lib := &MusicLibrary{
		Name:      name,
		Path:      path,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Tracks:    make([]TrackInfo, 0),
	}

	lm.libraries[name] = lib
	lm.currentLib = name

	// 保存到文件
	if err := lm.SaveLibrary(lib); err != nil {
		return err
	}

	// 后台扫描音乐文件
	go lm.ScanLibrary(name)

	return nil
}

// ScanLibrary 扫描音乐库中的音乐文件
func (lm *LibraryManager) ScanLibrary(name string) error {
	lm.mu.RLock()
	lib, ok := lm.libraries[name]
	if !ok {
		lm.mu.RUnlock()
		return fmt.Errorf("音乐库 %s 不存在", name)
	}
	libPath := lib.Path
	lm.mu.RUnlock()

	fmt.Printf("开始扫描音乐库：%s, 路径：%s\n", name, libPath)

	// 音乐文件扩展名
	audioExts := []string{".mp3", ".wav", ".flac", ".aac", ".ogg", ".wma", ".m4a"}

	var tracks []TrackInfo

	// 遍历目录
	err := filepath.Walk(libPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过隐藏文件和目录
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 检查是否是音乐文件
		ext := strings.ToLower(filepath.Ext(path))
		for _, audioExt := range audioExts {
			if ext == audioExt {
				// 获取文件信息
				track := TrackInfo{
					Path:     path,
					Filename: info.Name(),
					Title:    strings.TrimSuffix(info.Name(), ext),
					Size:     info.Size(),
				}

				// TODO: 读取 ID3 标签获取更多信息
				// 目前只使用文件名作为标题

				tracks = append(tracks, track)
				break
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("扫描音乐文件失败：%w", err)
	}

	// 更新音乐库
	lm.mu.Lock()
	if lib, ok := lm.libraries[name]; ok {
		lib.Tracks = tracks
		lib.UpdatedAt = time.Now()
		lm.mu.Unlock()

		// 保存更新
		if err := lm.SaveLibrary(lib); err != nil {
			return fmt.Errorf("保存音乐库失败：%w", err)
		}

		// 发送事件通知前端更新
		if lm.app != nil {
			lm.app.Event.Emit("libraryUpdated", name)
			lm.app.Event.Emit("playlistUpdated", getTrackPaths(tracks))
		}

		fmt.Printf("音乐库 %s 扫描完成，找到 %d 首歌曲\n", name, len(tracks))
	} else {
		lm.mu.Unlock()
	}

	return nil
}

// getTrackPaths 获取所有音轨路径
func getTrackPaths(tracks []TrackInfo) []string {
	paths := make([]string, len(tracks))
	for i, track := range tracks {
		paths[i] = track.Path
	}
	return paths
}

// GetCurrentLibrary 获取当前音乐库
func (lm *LibraryManager) GetCurrentLibrary() *MusicLibrary {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	if lib, ok := lm.libraries[lm.currentLib]; ok {
		return lib
	}
	return nil
}

// GetLibraries 获取所有音乐库
func (lm *LibraryManager) GetLibraries() []string {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	names := make([]string, 0, len(lm.libraries))
	for name := range lm.libraries {
		names = append(names, name)
	}
	return names
}

// RenameLibrary 重命名音乐库
func (lm *LibraryManager) RenameLibrary(oldName, newName string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lib, ok := lm.libraries[oldName]
	if !ok {
		return fmt.Errorf("音乐库 %s 不存在", oldName)
	}

	// 删除旧文件
	oldFile := filepath.Join(lm.libDir, oldName+".json")
	os.Remove(oldFile)

	// 更新库名
	lib.Name = newName
	lib.UpdatedAt = time.Now()

	// 删除旧条目
	delete(lm.libraries, oldName)

	// 添加新条目
	lm.libraries[newName] = lib

	// 更新当前库
	if lm.currentLib == oldName {
		lm.currentLib = newName
	}

	// 保存新文件
	return lm.SaveLibrary(lib)
}

// RefreshLibrary 刷新音乐库
func (lm *LibraryManager) RefreshLibrary(name string) error {
	return lm.ScanLibrary(name)
}

// Shutdown 关闭管理器
func (lm *LibraryManager) Shutdown() error {
	return nil
}
