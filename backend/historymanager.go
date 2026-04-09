package backend

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/yhao521/wailsMusicPlay/backend/pkg/file"
)

// HistoryRecord 播放历史记录
type HistoryRecord struct {
	Path       string    `json:"path"`        // 歌曲路径
	Title      string    `json:"title"`       // 标题
	Artist     string    `json:"artist"`      // 艺术家
	Album      string    `json:"album"`       // 专辑
	PlayedAt   time.Time `json:"played_at"`   // 播放时间
	Duration   int64     `json:"duration"`    // 播放时长（秒）
	FileSize   int64     `json:"file_size"`   // 文件大小（字节）
	PlayCount  int       `json:"play_count"`  // 播放次数
}

// HistoryManager 播放历史管理器
type HistoryManager struct {
	mu       sync.RWMutex
	records  []HistoryRecord
	maxSize  int          // 最大记录数
	histFile string       // 历史文件路径
	app      *application.App
}

// NewHistoryManager 创建播放历史管理器
func NewHistoryManager() *HistoryManager {
	return &HistoryManager{
		records: make([]HistoryRecord, 0),
		maxSize: 100, // 默认保存最近 100 条
		histFile: filepath.Join(file.GetLibPath(), "history.json"),
	}
}

// SetApp 设置应用实例
func (hm *HistoryManager) SetApp(app *application.App) {
	hm.app = app
}

// Init 初始化历史管理器
func (hm *HistoryManager) Init() error {
	return hm.loadHistory()
}

// loadHistory 从文件加载历史记录
func (hm *HistoryManager) loadHistory() error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	data, err := os.ReadFile(hm.histFile)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，创建空的历史记录
			hm.records = make([]HistoryRecord, 0)
			return nil
		}
		return fmt.Errorf("读取历史文件失败：%w", err)
	}

	var records []HistoryRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return fmt.Errorf("解析历史文件失败：%w", err)
	}

	hm.records = records
	log.Printf("✓ 加载播放历史记录：%d 条", len(records))
	return nil
}

// saveHistory 保存历史记录到文件
func (hm *HistoryManager) saveHistory() error {
	data, err := json.Marshal(hm.records)
	if err != nil {
		return fmt.Errorf("序列化历史记录失败：%w", err)
	}

	if err := os.WriteFile(hm.histFile, data, 0644); err != nil {
		return fmt.Errorf("写入历史文件失败：%w", err)
	}

	return nil
}

// AddToHistory 添加播放记录
func (hm *HistoryManager) AddToHistory(track TrackInfo) {
	go func() {
		hm.mu.Lock()
		defer hm.mu.Unlock()

		now := time.Now()

		// 检查是否已存在该歌曲的记录
		existingIndex := -1
		for i, record := range hm.records {
			if record.Path == track.Path {
				existingIndex = i
				break
			}
		}

		newRecord := HistoryRecord{
			Path:      track.Path,
			Title:     track.Title,
			Artist:    track.Artist,
			Album:     track.Album,
			PlayedAt:  now,
			Duration:  track.Duration,
			FileSize:  track.Size,
			PlayCount: 1, // 初始播放次数为 1
		}

		if existingIndex >= 0 {
			// 更新现有记录：增加播放次数，更新时间戳
			existingRecord := hm.records[existingIndex]
			newRecord.PlayCount = existingRecord.PlayCount + 1
			
			// 移到最前面
			copy(hm.records[1:existingIndex+1], hm.records[:existingIndex])
			hm.records[0] = newRecord
			log.Printf("📝 更新播放历史：%s (播放 %d 次)", track.Title, newRecord.PlayCount)
		} else {
			// 添加新记录到开头
			hm.records = append([]HistoryRecord{newRecord}, hm.records...)
			log.Printf("📝 添加播放历史：%s (第 1 次播放)", track.Title)
		}

		// 限制记录数量
		if len(hm.records) > hm.maxSize {
			hm.records = hm.records[:hm.maxSize]
		}

		// 异步保存到文件
		if err := hm.saveHistory(); err != nil {
			log.Printf("⚠️ 保存历史记录失败：%v", err)
		}

		// 发送事件通知
		if hm.app != nil {
			hm.app.Event.Emit("historyUpdated", hm.records)
		}
	}()
}

// GetHistory 获取历史记录列表
func (hm *HistoryManager) GetHistory(limit int) []HistoryRecord {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	if limit <= 0 || limit > len(hm.records) {
		limit = len(hm.records)
	}

	// 返回副本，避免并发问题
	result := make([]HistoryRecord, limit)
	copy(result, hm.records[:limit])
	return result
}

// ClearHistory 清空历史记录
func (hm *HistoryManager) ClearHistory() error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.records = make([]HistoryRecord, 0)

	if err := hm.saveHistory(); err != nil {
		return err
	}

	log.Println("✓ 清空播放历史记录")

	if hm.app != nil {
		hm.app.Event.Emit("historyUpdated", hm.records)
	}

	return nil
}

// RemoveFromHistory 删除指定索引的历史记录
func (hm *HistoryManager) RemoveFromHistory(index int) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if index < 0 || index >= len(hm.records) {
		return fmt.Errorf("索引越界：%d", index)
	}

	// 删除记录
	hm.records = append(hm.records[:index], hm.records[index+1:]...)

	if err := hm.saveHistory(); err != nil {
		return err
	}

	log.Printf("✓ 删除历史记录索引：%d", index)

	if hm.app != nil {
		hm.app.Event.Emit("historyUpdated", hm.records)
	}

	return nil
}

// GetHistoryCount 获取历史记录数量
func (hm *HistoryManager) GetHistoryCount() int {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return len(hm.records)
}

// GetFavoriteTracks 获取喜爱音乐（按播放次数递减排序）
func (hm *HistoryManager) GetFavoriteTracks(limit int) []HistoryRecord {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	// 创建副本用于排序
	sorted := make([]HistoryRecord, len(hm.records))
	copy(sorted, hm.records)

	// 按播放次数递减排序
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].PlayCount > sorted[i].PlayCount {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// 返回前 N 条
	if limit <= 0 || limit > len(sorted) {
		limit = len(sorted)
	}

	result := make([]HistoryRecord, limit)
	copy(result, sorted[:limit])
	return result
}
