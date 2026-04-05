package backend

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/yhao521/wailsMusicPlay/backend/pkg/file"
)

// LyricLine 歌词行
type LyricLine struct {
	Time    float64 `json:"time"`    // 时间点（秒）
	Content string  `json:"content"` // 歌词内容
}

// LyricInfo 歌词信息
type LyricInfo struct {
	Title   string      `json:"title"`   // 歌曲标题
	Artist  string      `json:"artist"`  // 艺术家
	Album   string      `json:"album"`   // 专辑
	Offset  float64     `json:"offset"`  // 时间偏移量（秒）
	Lines   []LyricLine `json:"lines"`   // 歌词行列表（按时间排序）
	HasLyric bool       `json:"has_lyric"` // 是否有歌词
}

// LyricManager 歌词管理器
type LyricManager struct {
	mu       sync.RWMutex
	cache    map[string]*LyricInfo // 缓存：文件路径 -> 歌词信息
	lyricDir string                // 歌词目录
}

// NewLyricManager 创建歌词管理器
func NewLyricManager() *LyricManager {
	return &LyricManager{
		cache:    make(map[string]*LyricInfo),
		lyricDir: filepath.Join(file.GetLibPath(), "lyrics"),
	}
}

// Init 初始化歌词管理器
func (lm *LyricManager) Init() error {
	// 创建歌词目录
	if err := os.MkdirAll(lm.lyricDir, 0755); err != nil {
		return fmt.Errorf("创建歌词目录失败：%w", err)
	}
	log.Println("✓ 歌词管理器初始化完成")
	return nil
}

// LoadLyric 加载歌词文件
func (lm *LyricManager) LoadLyric(trackPath string) (*LyricInfo, error) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	// 检查缓存
	if lyric, ok := lm.cache[trackPath]; ok {
		return lyric, nil
	}

	// 查找歌词文件
	lrcPath := lm.findLyricFile(trackPath)
	if lrcPath == "" {
		// 无歌词文件，返回空对象
		emptyLyric := &LyricInfo{
			HasLyric: false,
			Lines:    make([]LyricLine, 0),
		}
		lm.cache[trackPath] = emptyLyric
		return emptyLyric, nil
	}

	// 解析歌词文件
	lyric, err := lm.parseLRCFile(lrcPath)
	if err != nil {
		log.Printf("⚠️ 解析歌词文件失败 %s：%v", lrcPath, err)
		emptyLyric := &LyricInfo{
			HasLyric: false,
			Lines:    make([]LyricLine, 0),
		}
		lm.cache[trackPath] = emptyLyric
		return emptyLyric, nil
	}

	// 缓存结果
	lm.cache[trackPath] = lyric
	log.Printf("✓ 加载歌词：%d 行", len(lyric.Lines))

	return lyric, nil
}

// findLyricFile 查找歌词文件
func (lm *LyricManager) findLyricFile(trackPath string) string {
	baseName := strings.TrimSuffix(filepath.Base(trackPath), filepath.Ext(trackPath))
	dirPath := filepath.Dir(trackPath)

	// 策略 1: 同目录下的 .lrc 文件
	lrcPath1 := filepath.Join(dirPath, baseName+".lrc")
	if _, err := os.Stat(lrcPath1); err == nil {
		return lrcPath1
	}

	// 策略 2: 歌词目录下的 .lrc 文件
	lrcPath2 := filepath.Join(lm.lyricDir, baseName+".lrc")
	if _, err := os.Stat(lrcPath2); err == nil {
		return lrcPath2
	}

	return ""
}

// parseLRCFile 解析 LRC 歌词文件
func (lm *LyricManager) parseLRCFile(filePath string) (*LyricInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开歌词文件失败：%w", err)
	}
	defer file.Close()

	lyric := &LyricInfo{
		Lines: make([]LyricLine, 0),
	}

	// 正则表达式匹配时间标签 [mm:ss.xx] 或 [mm:ss:xx]
	timePattern := regexp.MustCompile(`\[(\d{2}):(\d{2})[.:](\d{2,3})\]`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		// 解析元数据标签
		if strings.HasPrefix(line, "[ti:") {
			lyric.Title = strings.TrimSuffix(strings.TrimPrefix(line, "[ti:"), "]")
			continue
		}
		if strings.HasPrefix(line, "[ar:") {
			lyric.Artist = strings.TrimSuffix(strings.TrimPrefix(line, "[ar:"), "]")
			continue
		}
		if strings.HasPrefix(line, "[al:") {
			lyric.Album = strings.TrimSuffix(strings.TrimPrefix(line, "[al:"), "]")
			continue
		}
		if strings.HasPrefix(line, "[offset:") {
			offsetStr := strings.TrimSuffix(strings.TrimPrefix(line, "[offset:"), "]")
			if offset, err := strconv.ParseFloat(offsetStr, 64); err == nil {
				lyric.Offset = offset / 1000.0 // 转换为秒
			}
			continue
		}

		// 解析歌词行
		matches := timePattern.FindAllStringSubmatch(line, -1)
		if len(matches) > 0 {
			// 提取歌词内容（去除所有时间标签）
			content := timePattern.ReplaceAllString(line, "")
			content = strings.TrimSpace(content)

			// 一个歌词行可能有多个时间标签
			for _, match := range matches {
				minutes, _ := strconv.Atoi(match[1])
				seconds, _ := strconv.Atoi(match[2])
				hundredths, _ := strconv.Atoi(match[3])

				// 处理百分秒或毫秒
				var timeSeconds float64
				if len(match[3]) == 3 {
					// 毫秒格式 [mm:ss:xxx]
					timeSeconds = float64(minutes)*60 + float64(seconds) + float64(hundredths)/1000.0
				} else {
					// 百分秒格式 [mm:ss.xx]
					timeSeconds = float64(minutes)*60 + float64(seconds) + float64(hundredths)/100.0
				}

				// 应用偏移量
				timeSeconds += lyric.Offset

				lyricLine := LyricLine{
					Time:    timeSeconds,
					Content: content,
				}
				lyric.Lines = append(lyric.Lines, lyricLine)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取歌词文件失败：%w", err)
	}

	// 按时间排序
	sort.Slice(lyric.Lines, func(i, j int) bool {
		return lyric.Lines[i].Time < lyric.Lines[j].Time
	})

	lyric.HasLyric = len(lyric.Lines) > 0

	return lyric, nil
}

// GetCurrentLyricLine 获取当前时间点的歌词行索引
func (lm *LyricManager) GetCurrentLyricLine(trackPath string, position float64) (int, error) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	lyric, ok := lm.cache[trackPath]
	if !ok || !lyric.HasLyric {
		return -1, fmt.Errorf("没有可用的歌词")
	}

	if len(lyric.Lines) == 0 {
		return -1, fmt.Errorf("歌词行为空")
	}

	// 二分查找最接近的歌词行
	index := sort.Search(len(lyric.Lines), func(i int) bool {
		return lyric.Lines[i].Time > position
	})

	// index 是第一个大于 position 的行，所以当前行是 index-1
	if index > 0 {
		return index - 1, nil
	}

	return 0, nil
}

// GetAllLyrics 获取所有歌词行
func (lm *LyricManager) GetAllLyrics(trackPath string) ([]LyricLine, error) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	lyric, ok := lm.cache[trackPath]
	if !ok {
		return nil, fmt.Errorf("歌词未加载")
	}

	if !lyric.HasLyric {
		return make([]LyricLine, 0), nil
	}

	// 返回副本
	result := make([]LyricLine, len(lyric.Lines))
	copy(result, lyric.Lines)
	return result, nil
}

// HasLyric 检查是否有歌词
func (lm *LyricManager) HasLyric(trackPath string) bool {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	lyric, ok := lm.cache[trackPath]
	if !ok {
		return false
	}

	return lyric.HasLyric
}

// ClearCache 清除歌词缓存
func (lm *LyricManager) ClearCache() {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.cache = make(map[string]*LyricInfo)
	log.Println("✓ 清除歌词缓存")
}
