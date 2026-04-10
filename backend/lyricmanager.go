package backend

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

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
	mu            sync.RWMutex
	cache         map[string]*LyricInfo    // 缓存：文件路径 -> 歌词信息
	searchCache   map[string]string        // 搜索缓存：key -> 歌词内容 (避免重复API调用)
	lyricDir      string                   // 歌词目录
}

// NewLyricManager 创建歌词管理器
func NewLyricManager() *LyricManager {
	return &LyricManager{
		cache:       make(map[string]*LyricInfo),
		searchCache: make(map[string]string),
		lyricDir:    filepath.Join(file.GetLibPath(), "lyrics"),
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

// LRCLibResponse lrclib.net API 响应结构
type LRCLibResponse struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	TrackName    string  `json:"trackName"`
	ArtistName   string  `json:"artistName"`
	AlbumName    string  `json:"albumName"`
	Duration     float64 `json:"duration"`
	Instrumental bool    `json:"instrumental"`
	PlainLyrics  string  `json:"plainLyrics"`
	SyncedLyrics string  `json:"syncedLyrics"`
}

// LyricSource 歌词源配置
type LyricSource struct {
	Name        string
	Description string
	Priority    int // 优先级,数字越小优先级越高
	DownloadFn  func(title, artist, album string) (string, error)
}

// DownloadLyricFromLRCLib 从 lrclib.net 下载歌词(增强版 - 支持多种搜索策略)
func (lm *LyricManager) DownloadLyricFromLRCLib(trackPath string, title, artist, album string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	log.Printf("🎵 尝试从 lrclib.net 下载歌词: %s - %s", artist, title)

	// 定义多种搜索变体,按优先级尝试
	searchVariants := []struct {
		name   string
		title  string
		artist string
		album  string
	}{
		{"标准格式", title, artist, album},
		{"仅标题+艺术家", title, artist, ""},
		{"清理特殊字符", cleanTitle(title), cleanArtist(artist), album},
		{"艺术家-标题组合格式", fmt.Sprintf("%s - %s", artist, title), "", ""},
		{"仅标题", title, "", ""},
	}

	var lastErr error
	for i, variant := range searchVariants {
		if variant.title == "" && variant.artist == "" {
			continue
		}

		log.Printf("  🔄 尝试搜索变体 %d/%d: %s", i+1, len(searchVariants), variant.name)

		err := lm.tryLRCLibSearch(trackPath, variant.title, variant.artist, variant.album)
		if err == nil {
			log.Printf("  ✓ 成功: %s", variant.name)
			return nil
		}

		log.Printf("  ❌ 失败: %v", err)
		lastErr = err
	}

	return fmt.Errorf("lrclib 所有搜索变体均失败: %w", lastErr)
}

// tryLRCLibSearch 尝试单次 lrclib 搜索(不锁定)
func (lm *LyricManager) tryLRCLibSearch(trackPath, title, artist, album string) error {
	// 构建缓存键
	cacheKey := fmt.Sprintf("%s|%s|%s", title, artist, album)
	
	// 检查搜索缓存
	lm.mu.RLock()
	if cachedLyrics, ok := lm.searchCache[cacheKey]; ok {
		lm.mu.RUnlock()
		log.Printf("  ⚡ 使用缓存的搜索结果")
		
		// 直接使用缓存的歌词保存
		return lm.saveLyricsToFile(trackPath, cachedLyrics)
	}
	lm.mu.RUnlock()

	// 构建搜索参数
	params := make([]string, 0)
	if title != "" {
		params = append(params, fmt.Sprintf("track_name=%s", urlEncode(title)))
	}
	if artist != "" {
		params = append(params, fmt.Sprintf("artist_name=%s", urlEncode(artist)))
	}
	if album != "" {
		params = append(params, fmt.Sprintf("album_name=%s", urlEncode(album)))
	}

	if len(params) == 0 {
		return fmt.Errorf("没有足够的信息来搜索歌词")
	}

	searchURL := fmt.Sprintf("https://lrclib.net/api/get?%s", strings.Join(params, "&"))

	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 发送请求
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置 User-Agent
	req.Header.Set("User-Agent", "HaoyunMusicPlayer/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("未找到歌词")
		}
		return fmt.Errorf("API 返回错误状态码: %d", resp.StatusCode)
	}

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析 JSON
	var result LRCLibResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查是否有歌词
	lyricsContent := result.SyncedLyrics
	if lyricsContent == "" {
		lyricsContent = result.PlainLyrics
	}

	if lyricsContent == "" {
		return fmt.Errorf("未找到可用的歌词")
	}

	// 保存到缓存
	lm.mu.Lock()
	lm.searchCache[cacheKey] = lyricsContent
	lm.mu.Unlock()

	// 保存歌词文件
	return lm.saveLyricsToFile(trackPath, lyricsContent)
}

// saveLyricsToFile 保存歌词到文件(通用方法)
func (lm *LyricManager) saveLyricsToFile(trackPath, lyricsContent string) error {
	baseName := strings.TrimSuffix(filepath.Base(trackPath), filepath.Ext(trackPath))
	dirPath := filepath.Dir(trackPath)
	lrcPath := filepath.Join(dirPath, baseName+".lrc")

	// 如果同目录下已有歌词文件,备份
	if _, err := os.Stat(lrcPath); err == nil {
		backupPath := lrcPath + ".bak"
		if err := os.Rename(lrcPath, backupPath); err != nil {
			log.Printf("⚠️ 备份旧歌词失败: %v", err)
		} else {
			log.Printf("✓ 已备份旧歌词到: %s", backupPath)
		}
	}

	// 写入新歌词文件
	if err := os.WriteFile(lrcPath, []byte(lyricsContent), 0644); err != nil {
		return fmt.Errorf("保存歌词文件失败: %w", err)
	}

	log.Printf("✓ 歌词下载成功并保存到: %s", lrcPath)

	// 清除解析缓存,确保下次加载时使用新歌词
	delete(lm.cache, trackPath)

	return nil
}

// cleanTitle 清理歌曲标题中的特殊字符和后缀
func cleanTitle(title string) string {
	// 移除括号内容: "歌曲名 (Live)" -> "歌曲名"
	re := regexp.MustCompile(`\s*\(.*?\)\s*`)
	title = re.ReplaceAllString(title, "")

	// 移除方括号内容
	re2 := regexp.MustCompile(`\s*\[.*?\]\s*`)
	title = re2.ReplaceAllString(title, "")

	// 移除常见后缀
	suffixes := []string{
		"Official", "MV", "HD", "HQ", "Audio", "Video",
		"official", "mv", "hd", "hq", "audio", "video",
		"官方版", "现场版", "伴奏", "翻唱",
	}
	for _, suffix := range suffixes {
		title = strings.ReplaceAll(title, suffix, "")
		title = strings.ReplaceAll(title, strings.ToLower(suffix), "")
	}

	// 移除多余空格
	title = strings.TrimSpace(title)
	// 合并多个空格为一个
	re3 := regexp.MustCompile(`\s+`)
	title = re3.ReplaceAllString(title, " ")

	return title
}

// cleanArtist 清理艺术家名称
func cleanArtist(artist string) string {
	// 移除 feat./ft. 后面的内容
	re := regexp.MustCompile(`\s*(feat\.?|ft\.?|featuring)\s+.*$`)
	artist = re.ReplaceAllString(artist, "")

	// 移除多余空格
	artist = strings.TrimSpace(artist)
	re2 := regexp.MustCompile(`\s+`)
	artist = re2.ReplaceAllString(artist, " ")

	return artist
}

// downloadFromNetease 从网易云音乐 API 下载歌词
func (lm *LyricManager) downloadFromNetease(title, artist string) (string, error) {
	log.Printf("  🎵 尝试从网易云音乐搜索: %s - %s", artist, title)

	// 第一步: 搜索歌曲
	searchURL := fmt.Sprintf("https://music.163.com/api/search/get/web?csrf_token=&s=%s&type=1&limit=5", 
		urlEncode(fmt.Sprintf("%s %s", title, artist)))

	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", searchURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", "https://music.163.com/")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("网易云搜索失败: %w", err)
	}
	defer resp.Body.Close()

	var searchResult map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &searchResult)

	// 解析搜索结果
	result, ok := searchResult["result"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("网易云搜索结果格式错误")
	}

	songs, ok := result["songs"].([]interface{})
	if !ok || len(songs) == 0 {
		return "", fmt.Errorf("网易云未找到歌曲")
	}

	// 获取第一首歌曲的 ID
	firstSong := songs[0].(map[string]interface{})
	songID := int(firstSong["id"].(float64))

	// 第二步: 获取歌词
	lyricURL := fmt.Sprintf("https://music.163.com/api/song/lyric?id=%d&lv=1&kv=1&tv=-1", songID)
	req2, _ := http.NewRequest("GET", lyricURL, nil)
	req2.Header.Set("User-Agent", "Mozilla/5.0")
	req2.Header.Set("Referer", "https://music.163.com/")

	resp2, err := client.Do(req2)
	if err != nil {
		return "", fmt.Errorf("获取歌词失败: %w", err)
	}
	defer resp2.Body.Close()

	var lyricResult map[string]interface{}
	body2, _ := io.ReadAll(resp2.Body)
	json.Unmarshal(body2, &lyricResult)

	// 提取歌词
	if lrc, ok := lyricResult["lrc"].(map[string]interface{}); ok {
		if lyric, ok := lrc["lyric"].(string); ok && lyric != "" {
			log.Printf("  ✓ 网易云歌词下载成功")
			return lyric, nil
		}
	}

	return "", fmt.Errorf("网易云歌词为空")
}

// downloadFromQQMusic 从 QQ 音乐 API 下载歌词
func (lm *LyricManager) downloadFromQQMusic(title, artist string) (string, error) {
	log.Printf("  🎵 尝试从 QQ 音乐搜索: %s - %s", artist, title)

	// QQ 音乐搜索 API
	searchURL := fmt.Sprintf("https://c.y.qq.com/soso/fcgi-bin/client_search_cp?format=json&p=1&n=5&w=%s",
		urlEncode(fmt.Sprintf("%s %s", title, artist)))

	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", searchURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", "https://y.qq.com/")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("QQ 音乐搜索失败: %w", err)
	}
	defer resp.Body.Close()

	var searchResult map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &searchResult)

	// 解析搜索结果
	data, ok := searchResult["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("QQ 音乐搜索结果格式错误")
	}

	song, ok := data["song"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("QQ 音乐歌曲数据缺失")
	}

	list, ok := song["list"].([]interface{})
	if !ok || len(list) == 0 {
		return "", fmt.Errorf("QQ 音乐未找到歌曲")
	}

	// 获取歌曲信息
	firstSong := list[0].(map[string]interface{})
	songMid := firstSong["songmid"].(string)
	_ = int(firstSong["songid"].(float64)) // songID 暂未使用,保留以备将来扩展

	// 获取歌词
	lyricURL := fmt.Sprintf("https://c.y.qq.com/lyric/fcgi-bin/fcg_query_lyric_new.fcg?songmid=%s&g_tk=5381&format=json&inCharset=utf-8&outCharset=utf-8",
		songMid)

	req2, _ := http.NewRequest("GET", lyricURL, nil)
	req2.Header.Set("User-Agent", "Mozilla/5.0")
	req2.Header.Set("Referer", "https://y.qq.com/")

	resp2, err := client.Do(req2)
	if err != nil {
		return "", fmt.Errorf("获取 QQ 音乐歌词失败: %w", err)
	}
	defer resp2.Body.Close()

	var lyricResult map[string]interface{}
	body2, _ := io.ReadAll(resp2.Body)
	json.Unmarshal(body2, &lyricResult)

	// 提取并解码歌词
	if lyricStr, ok := lyricResult["lyric"].(string); ok && lyricStr != "" {
		// QQ 音乐歌词通常是 Base64 编码的
		decoded, err := decodeBase64Gzip(lyricStr)
		if err == nil {
			log.Printf("  ✓ QQ 音乐歌词下载成功")
			return string(decoded), nil
		}
		// 如果不是压缩的,直接返回
		log.Printf("  ✓ QQ 音乐歌词下载成功(未压缩)")
		return lyricStr, nil
	}

	return "", fmt.Errorf("QQ 音乐歌词为空")
}

// downloadFromAuralive 从 Auralive Lyrics API 下载歌词
func (lm *LyricManager) downloadFromAuralive(title, artist string) (string, error) {
	log.Printf("  🎵 尝试从 Auralive Lyrics API 搜索: %s - %s", artist, title)

	// Auralive Lyrics API 端点 (使用公共实例)
	// API 文档: https://github.com/auralive/lyrics-api
	apiBase := "https://api.auralive.net"
	
	// 构建搜索 URL
	searchURL := fmt.Sprintf("%s/api/v1/lyrics/search?title=%s&artist=%s",
		apiBase,
		urlEncode(title),
		urlEncode(artist))

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return "", fmt.Errorf("创建 Auralive 请求失败: %w", err)
	}

	req.Header.Set("User-Agent", "HaoyunMusicPlayer/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Auralive 请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return "", fmt.Errorf("Auralive 未找到歌词")
		}
		return "", fmt.Errorf("Auralive API 返回错误状态码: %d", resp.StatusCode)
	}

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取 Auralive 响应失败: %w", err)
	}

	// 解析 JSON 响应
	var result AuraliveResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析 Auralive 响应失败: %w", err)
	}

	// 检查是否有结果
	if len(result.Data) == 0 {
		return "", fmt.Errorf("Auralive 未找到匹配的歌曲")
	}

	// 获取第一首匹配歌曲的歌词
	bestMatch := result.Data[0]
	
	// 优先使用同步歌词,否则使用普通歌词
	lyricsContent := bestMatch.SyncedLyrics
	if lyricsContent == "" {
		lyricsContent = bestMatch.PlainLyrics
	}

	if lyricsContent == "" {
		return "", fmt.Errorf("Auralive 歌词为空")
	}

	log.Printf("  ✓ Auralive Lyrics API 歌词下载成功")
	return lyricsContent, nil
}

// AuraliveResponse Auralive Lyrics API 响应结构
type AuraliveResponse struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    []AuraliveResult `json:"data"`
}

// AuraliveResult Auralive 搜索结果
type AuraliveResult struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Artist        string `json:"artist"`
	Album         string `json:"album"`
	Duration      int    `json:"duration"`
	SyncedLyrics  string `json:"synced_lyrics"`  // 同步歌词(LRC格式)
	PlainLyrics   string `json:"plain_lyrics"`   // 普通歌词
	MatchScore    float64 `json:"match_score"`   // 匹配度分数
}

// decodeBase64Gzip 解码 Base64 编码的 Gzip 压缩数据
func decodeBase64Gzip(encoded string) ([]byte, error) {
	// 这里简化处理,实际可能需要更复杂的解码逻辑
	// QQ 音乐的歌词有时是 Base64 编码的 gzip 压缩数据
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	// 尝试解压缩
	reader, err := gzip.NewReader(bytes.NewReader(decoded))
	if err != nil {
		// 如果不是 gzip,直接返回解码后的数据
		return decoded, nil
	}
	defer reader.Close()

	var result bytes.Buffer
	_, err = io.Copy(&result, reader)
	if err != nil {
		return decoded, nil
	}

	return result.Bytes(), nil
}

// DownloadLyricWithFallback 使用多个源下载歌词(带降级策略)
func (lm *LyricManager) DownloadLyricWithFallback(trackPath string, title, artist, album string) error {
	log.Printf("🎵 开始多源歌词下载: %s - %s", artist, title)

	// 定义歌词源列表(按优先级排序)
	type lyricSource struct {
		name       string
		downloadFn func(string, string) (string, error)
	}

	sources := []lyricSource{
		{
			name: "lrclib.net (增强版)",
			downloadFn: func(t, a string) (string, error) {
				// 使用增强版方法(包含智能搜索+模糊匹配)
				err := lm.DownloadLyricFromLRCLibEnhanced(trackPath, t, artist, album)
				if err != nil {
					return "", err
				}
				// 读取刚保存的文件内容
				baseName := strings.TrimSuffix(filepath.Base(trackPath), filepath.Ext(trackPath))
				dirPath := filepath.Dir(trackPath)
				lrcPath := filepath.Join(dirPath, baseName+".lrc")
				content, err := os.ReadFile(lrcPath)
				if err != nil {
					return "", err
				}
				return string(content), nil
			},
		},
		{"网易云音乐", lm.downloadFromNetease},
		{"QQ 音乐", lm.downloadFromQQMusic},
		{"Auralive Lyrics", lm.downloadFromAuralive}, // Auralive Lyrics API
	}

	var lastErr error
	for _, source := range sources {
		log.Printf("  📡 尝试从 %s 下载...", source.name)

		lyrics, err := source.downloadFn(title, artist)
		if err != nil {
			log.Printf("  ❌ %s 失败: %v", source.name, err)
			lastErr = err
			continue
		}

		// 保存歌词
		baseName := strings.TrimSuffix(filepath.Base(trackPath), filepath.Ext(trackPath))
		dirPath := filepath.Dir(trackPath)
		lrcPath := filepath.Join(dirPath, baseName+".lrc")

		// 备份旧文件
		if _, err := os.Stat(lrcPath); err == nil {
			backupPath := lrcPath + ".bak"
			os.Rename(lrcPath, backupPath)
		}

		// 写入新文件
		if err := os.WriteFile(lrcPath, []byte(lyrics), 0644); err != nil {
			log.Printf("  ❌ 保存歌词失败: %v", err)
			lastErr = err
			continue
		}

		log.Printf("  ✓ 从 %s 成功下载并保存歌词", source.name)
		delete(lm.cache, trackPath)
		return nil
	}

	return fmt.Errorf("所有歌词源均失败: %w", lastErr)
}

// DownloadLyricWithFallbackToDir 使用多个源下载歌词并保存到指定目录(带降级策略)
func (lm *LyricManager) DownloadLyricWithFallbackToDir(trackPath, lyricsDir string, title, artist, album string) error {
	log.Printf("🎵 开始多源歌词下载: %s - %s", artist, title)

	// 定义歌词源列表(按优先级排序)
	type lyricSource struct {
		name       string
		downloadFn func(string, string) (string, error)
	}

	sources := []lyricSource{
		{
			name: "lrclib.net (增强版)",
			downloadFn: func(t, a string) (string, error) {
				// 使用增强版方法(包含智能搜索+模糊匹配)
				err := lm.DownloadLyricFromLRCLibEnhancedToDir(trackPath, lyricsDir, t, artist, album)
				if err != nil {
					return "", err
				}
				// 读取刚保存的文件内容
				baseName := strings.TrimSuffix(filepath.Base(trackPath), filepath.Ext(trackPath))
				lrcPath := filepath.Join(lyricsDir, baseName+".lrc")
				content, err := os.ReadFile(lrcPath)
				if err != nil {
					return "", err
				}
				return string(content), nil
			},
		},
		{"网易云音乐", lm.downloadFromNetease},
		{"QQ 音乐", lm.downloadFromQQMusic},
		{"Auralive Lyrics", lm.downloadFromAuralive}, // Auralive Lyrics API
	}

	var lastErr error
	for _, source := range sources {
		log.Printf("  📡 尝试从 %s 下载...", source.name)

		lyrics, err := source.downloadFn(title, artist)
		if err != nil {
			log.Printf("  ❌ %s 失败: %v", source.name, err)
			lastErr = err
			continue
		}

		// 保存歌词到指定目录
		baseName := strings.TrimSuffix(filepath.Base(trackPath), filepath.Ext(trackPath))
		lrcPath := filepath.Join(lyricsDir, baseName+".lrc")

		// 备份旧文件
		if _, err := os.Stat(lrcPath); err == nil {
			backupPath := lrcPath + ".bak"
			os.Rename(lrcPath, backupPath)
		}

		// 写入新文件
		if err := os.WriteFile(lrcPath, []byte(lyrics), 0644); err != nil {
			log.Printf("  ❌ 保存歌词失败: %v", err)
			lastErr = err
			continue
		}

		log.Printf("  ✓ 从 %s 成功下载并保存歌词", source.name)
		delete(lm.cache, trackPath)
		return nil
	}

	return fmt.Errorf("所有歌词源均失败: %w", lastErr)
}

// downloadAndSaveFromLRCLib 从 lrclib 下载并保存歌词(不锁定)
func (lm *LyricManager) downloadAndSaveFromLRCLib(trackPath, title, artist, album string) error {
	// 构建搜索参数
	params := make([]string, 0)
	if title != "" {
		params = append(params, fmt.Sprintf("track_name=%s", urlEncode(title)))
	}
	if artist != "" {
		params = append(params, fmt.Sprintf("artist_name=%s", urlEncode(artist)))
	}
	if album != "" {
		params = append(params, fmt.Sprintf("album_name=%s", urlEncode(album)))
	}

	if len(params) == 0 {
		return fmt.Errorf("没有足够的信息来搜索歌词")
	}

	searchURL := fmt.Sprintf("https://lrclib.net/api/get?%s", strings.Join(params, "&"))

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "HaoyunMusicPlayer/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("未找到歌词")
		}
		return fmt.Errorf("API 返回错误状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result LRCLibResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	lyricsContent := result.SyncedLyrics
	if lyricsContent == "" {
		lyricsContent = result.PlainLyrics
	}

	if lyricsContent == "" {
		return fmt.Errorf("未找到可用的歌词")
	}

	// 保存歌词文件
	baseName := strings.TrimSuffix(filepath.Base(trackPath), filepath.Ext(trackPath))
	dirPath := filepath.Dir(trackPath)
	lrcPath := filepath.Join(dirPath, baseName+".lrc")

	if _, err := os.Stat(lrcPath); err == nil {
		backupPath := lrcPath + ".bak"
		os.Rename(lrcPath, backupPath)
	}

	return os.WriteFile(lrcPath, []byte(lyricsContent), 0644)
}

// urlEncode 简单的 URL 编码
func urlEncode(s string) string {
	var result strings.Builder
	for _, r := range s {
		if isUnreserved(r) {
			result.WriteRune(r)
		} else {
			for _, b := range []byte(string(r)) {
				result.WriteString(fmt.Sprintf("%%%02X", b))
			}
		}
	}
	return result.String()
}

// isUnreserved 判断字符是否为 RFC 3986 中的非保留字符
func isUnreserved(r rune) bool {
	return (r >= 'A' && r <= 'Z') ||
		(r >= 'a' && r <= 'z') ||
		(r >= '0' && r <= '9') ||
		r == '-' || r == '_' || r == '.' || r == '~'
}

// DownloadLyricsForLibrary 为音乐库中的所有歌曲下载歌词
func (lm *LyricManager) DownloadLyricsForLibrary(libraryPath string, metadataManager *MetadataManager) (successCount, failCount, skipCount int, errors []string) {
	log.Printf("🎵 开始为音乐库下载歌词: %s", libraryPath)

	// 创建 LIB_LYRIC 目录
	lyricsDir := filepath.Join(libraryPath, "LIB_LYRIC")
	if err := os.MkdirAll(lyricsDir, 0755); err != nil {
		errors = append(errors, fmt.Sprintf("创建歌词目录失败: %v", err))
		return
	}
	log.Printf("📁 歌词目录: %s", lyricsDir)

	// 定义音乐文件扩展名
	musicExtensions := map[string]bool{
		".mp3":  true,
		".flac": true,
		".wav":  true,
		".ogg":  true,
		".aac":  true,
		".m4a":  true,
		".wma":  true,
		".ape":  true,
	}

	// 收集所有音乐文件
	var musicFiles []string
	err := filepath.Walk(libraryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("⚠️ 访问路径失败 %s: %v", path, err)
			return nil
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if musicExtensions[ext] {
			musicFiles = append(musicFiles, path)
		}

		return nil
	})

	if err != nil {
		errors = append(errors, fmt.Sprintf("遍历目录失败: %v", err))
		return
	}

	log.Printf("📊 找到 %d 个音乐文件", len(musicFiles))

	// 逐个下载歌词
	for i, trackPath := range musicFiles {
		log.Printf("[%d/%d] 处理: %s", i+1, len(musicFiles), filepath.Base(trackPath))

		// 检查是否已有歌词文件（在 LIB_LYRIC 目录下）
		baseName := strings.TrimSuffix(filepath.Base(trackPath), filepath.Ext(trackPath))
		lrcPath := filepath.Join(lyricsDir, baseName+".lrc")

		if _, err := os.Stat(lrcPath); err == nil {
			log.Printf("  ⏭️  跳过: 歌词文件已存在")
			skipCount++
			continue
		}

		// 获取元数据
		title := ""
		artist := ""
		album := ""

		if metadataManager != nil {
			meta, err := metadataManager.GetMetadata(trackPath)
			if err == nil {
				if t, ok := meta["title"].(string); ok {
					title = t
				}
				if a, ok := meta["artist"].(string); ok {
					artist = a
				}
				if al, ok := meta["album"].(string); ok {
					album = al
				}
			}
		}

		// 如果元数据中没有标题和艺术家,使用文件名
		if title == "" || artist == "" {
			fileName := strings.TrimSuffix(filepath.Base(trackPath), filepath.Ext(trackPath))
			parts := strings.Split(fileName, " - ")
			if len(parts) >= 2 {
				if artist == "" {
					artist = strings.TrimSpace(parts[0])
				}
				if title == "" {
					title = strings.TrimSpace(parts[1])
				}
			} else {
				if title == "" {
					title = fileName
				}
			}
		}

		// 使用多源降级策略下载歌词，直接保存到 LIB_LYRIC 目录
		err := lm.DownloadLyricWithFallbackToDir(trackPath, lyricsDir, title, artist, album)
		if err != nil {
			log.Printf("  ❌ 失败: %v", err)
			failCount++
			errors = append(errors, fmt.Sprintf("%s: %v", filepath.Base(trackPath), err))
		} else {
			log.Printf("  ✓ 成功")
			successCount++
		}

		// 避免频繁请求,添加延迟
		time.Sleep(500 * time.Millisecond)
	}

	log.Printf("✓ 歌词下载完成: 成功 %d, 失败 %d, 跳过 %d", successCount, failCount, skipCount)
	return
}

// calculateSimilarity 计算两个字符串的相似度 (0.0 - 1.0)
// 使用简化的 Levenshtein Distance 算法
func calculateSimilarity(s1, s2 string) float64 {
	s1 = strings.ToLower(strings.TrimSpace(s1))
	s2 = strings.ToLower(strings.TrimSpace(s2))

	if s1 == s2 {
		return 1.0
	}

	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	// 如果一个是另一个的子串,给予较高分数
	if strings.Contains(s1, s2) || strings.Contains(s2, s1) {
		shorter := float64(len(s1))
		longer := float64(len(s2))
		if len(s1) > len(s2) {
			shorter = float64(len(s2))
			longer = float64(len(s1))
		}
		return shorter / longer * 0.9
	}

	// 计算编辑距离
	distance := levenshteinDistance(s1, s2)
	maxLen := float64(max(len(s1), len(s2)))

	// 转换为相似度 (距离越小,相似度越高)
	similarity := 1.0 - float64(distance)/maxLen

	if similarity < 0.0 {
		return 0.0
	}
	return similarity
}

// levenshteinDistance 计算两个字符串的 Levenshtein 编辑距离
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// 创建矩阵
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	// 填充矩阵
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}

			matrix[i][j] = min(
				min(matrix[i-1][j]+1, matrix[i][j-1]+1), // 删除或插入
				matrix[i-1][j-1]+cost,                    // 替换
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max 返回两个整数中的较大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// searchLRCLibWithFallback 使用 lrclib 搜索 API 进行模糊搜索
func (lm *LyricManager) searchLRCLibWithFallback(title, artist string) (*LRCLibResponse, error) {
	log.Printf("  🔍 使用 lrclib 搜索 API 进行模糊搜索: %s - %s", artist, title)

	// 构建搜索查询
	query := fmt.Sprintf("%s %s", title, artist)
	searchURL := fmt.Sprintf("https://lrclib.net/api/search?q=%s", urlEncode(query))

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建搜索请求失败: %w", err)
	}

	req.Header.Set("User-Agent", "HaoyunMusicPlayer/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("搜索请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("搜索 API 返回错误状态码: %d", resp.StatusCode)
	}

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取搜索响应失败: %w", err)
	}

	// 解析搜索结果数组
	var results []LRCLibResponse
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, fmt.Errorf("解析搜索结果失败: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("未找到任何匹配结果")
	}

	// 计算每个结果的相似度评分
	type scoredResult struct {
		result     LRCLibResponse
		score      float64
		titleScore float64
		artistScore float64
	}

	var scoredResults []scoredResult
	for _, result := range results {
		// 计算标题相似度
		titleScore := calculateSimilarity(title, result.TrackName)
		// 计算艺术家相似度
		artistScore := calculateSimilarity(artist, result.ArtistName)
		
		// 综合评分 (标题权重 60%, 艺术家权重 40%)
		overallScore := titleScore*0.6 + artistScore*0.4

		scoredResults = append(scoredResults, scoredResult{
			result:      result,
			score:       overallScore,
			titleScore:  titleScore,
			artistScore: artistScore,
		})

		log.Printf("    📊 候选: \"%s\" by %s (相似度: %.2f, 标题: %.2f, 艺术家: %.2f)",
			result.TrackName, result.ArtistName, overallScore, titleScore, artistScore)
	}

	// 按评分排序
	sort.Slice(scoredResults, func(i, j int) bool {
		return scoredResults[i].score > scoredResults[j].score
	})

	// 选择最佳匹配 (评分 >= 0.7)
	bestMatch := scoredResults[0]
	if bestMatch.score < 0.7 {
		log.Printf("  ⚠️ 最佳匹配评分过低 (%.2f),可能不准确", bestMatch.score)
		// 仍然返回,但记录警告
	}

	log.Printf("  ✓ 选择最佳匹配: \"%s\" by %s (评分: %.2f)",
		bestMatch.result.TrackName, bestMatch.result.ArtistName, bestMatch.score)

	return &bestMatch.result, nil
}

// DownloadLyricFromLRCLibEnhanced 增强版 lrclib 下载(包含模糊搜索)
func (lm *LyricManager) DownloadLyricFromLRCLibEnhanced(trackPath string, title, artist, album string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	log.Printf("🎵 [增强模式] 尝试从 lrclib.net 下载歌词: %s - %s", artist, title)

	// 第一步: 尝试标准精确搜索(5种变体)
	log.Printf("  📍 阶段 1: 精确搜索")
	err := lm.DownloadLyricFromLRCLib(trackPath, title, artist, album)
	if err == nil {
		log.Printf("  ✓ 精确搜索成功")
		return nil
	}

	log.Printf("  ❌ 精确搜索失败: %v", err)
	log.Printf("  📍 阶段 2: 模糊搜索")

	// 第二步: 使用搜索 API 进行模糊匹配
	result, err := lm.searchLRCLibWithFallback(title, artist)
	if err != nil {
		return fmt.Errorf("模糊搜索也失败: %w", err)
	}

	// 检查是否有歌词
	lyricsContent := result.SyncedLyrics
	if lyricsContent == "" {
		lyricsContent = result.PlainLyrics
	}

	if lyricsContent == "" {
		return fmt.Errorf("模糊搜索结果中无可用歌词")
	}

	// 保存歌词文件
	baseName := strings.TrimSuffix(filepath.Base(trackPath), filepath.Ext(trackPath))
	dirPath := filepath.Dir(trackPath)
	lrcPath := filepath.Join(dirPath, baseName+".lrc")

	// 备份旧文件
	if _, err := os.Stat(lrcPath); err == nil {
		backupPath := lrcPath + ".bak"
		if err := os.Rename(lrcPath, backupPath); err != nil {
			log.Printf("⚠️ 备份旧歌词失败: %v", err)
		} else {
			log.Printf("✓ 已备份旧歌词到: %s", backupPath)
		}
	}

	// 写入新歌词文件
	if err := os.WriteFile(lrcPath, []byte(lyricsContent), 0644); err != nil {
		return fmt.Errorf("保存歌词文件失败: %w", err)
	}

	log.Printf("✓ [模糊匹配] 歌词下载成功并保存到: %s", lrcPath)
	log.Printf("  📊 匹配信息: 标题=\"%s\", 艺术家=\"%s\"", result.TrackName, result.ArtistName)

	// 清除缓存
	delete(lm.cache, trackPath)

	return nil
}

// DownloadLyricFromLRCLibEnhancedToDir 增强版 lrclib 下载(包含模糊搜索)，保存到指定目录
func (lm *LyricManager) DownloadLyricFromLRCLibEnhancedToDir(trackPath, lyricsDir string, title, artist, album string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	log.Printf("🎵 [增强模式] 尝试从 lrclib.net 下载歌词: %s - %s", artist, title)

	// 第一步: 尝试标准精确搜索(5种变体)
	log.Printf("  📍 阶段 1: 精确搜索")
	err := lm.downloadAndSaveFromLRCLibToDir(trackPath, lyricsDir, title, artist, album)
	if err == nil {
		log.Printf("  ✓ 精确搜索成功")
		return nil
	}

	log.Printf("  ❌ 精确搜索失败: %v", err)
	log.Printf("  📍 阶段 2: 模糊搜索")

	// 第二步: 使用搜索 API 进行模糊匹配
	result, err := lm.searchLRCLibWithFallback(title, artist)
	if err != nil {
		return fmt.Errorf("模糊搜索也失败: %w", err)
	}

	// 检查是否有歌词
	lyricsContent := result.SyncedLyrics
	if lyricsContent == "" {
		lyricsContent = result.PlainLyrics
	}

	if lyricsContent == "" {
		return fmt.Errorf("模糊搜索结果中无可用歌词")
	}

	// 保存歌词文件到指定目录
	baseName := strings.TrimSuffix(filepath.Base(trackPath), filepath.Ext(trackPath))
	lrcPath := filepath.Join(lyricsDir, baseName+".lrc")

	// 备份旧文件
	if _, err := os.Stat(lrcPath); err == nil {
		backupPath := lrcPath + ".bak"
		if err := os.Rename(lrcPath, backupPath); err != nil {
			log.Printf("⚠️ 备份旧歌词失败: %v", err)
		} else {
			log.Printf("✓ 已备份旧歌词到: %s", backupPath)
		}
	}

	// 写入新歌词文件
	if err := os.WriteFile(lrcPath, []byte(lyricsContent), 0644); err != nil {
		return fmt.Errorf("保存歌词文件失败: %w", err)
	}

	log.Printf("✓ [模糊匹配] 歌词下载成功并保存到: %s", lrcPath)
	log.Printf("  📊 匹配信息: 标题=\"%s\", 艺术家=\"%s\"", result.TrackName, result.ArtistName)

	// 清除缓存
	delete(lm.cache, trackPath)

	return nil
}

// downloadAndSaveFromLRCLibToDir 从 lrclib 下载并保存歌词到指定目录(不锁定)
func (lm *LyricManager) downloadAndSaveFromLRCLibToDir(trackPath, lyricsDir string, title, artist, album string) error {
	// 构建搜索参数
	params := make([]string, 0)
	if title != "" {
		params = append(params, fmt.Sprintf("track_name=%s", urlEncode(title)))
	}
	if artist != "" {
		params = append(params, fmt.Sprintf("artist_name=%s", urlEncode(artist)))
	}
	if album != "" {
		params = append(params, fmt.Sprintf("album_name=%s", urlEncode(album)))
	}

	if len(params) == 0 {
		return fmt.Errorf("没有足够的信息来搜索歌词")
	}

	searchURL := fmt.Sprintf("https://lrclib.net/api/get?%s", strings.Join(params, "&"))

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "HaoyunMusicPlayer/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("未找到歌词")
		}
		return fmt.Errorf("API 返回错误状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result LRCLibResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	lyricsContent := result.SyncedLyrics
	if lyricsContent == "" {
		lyricsContent = result.PlainLyrics
	}

	if lyricsContent == "" {
		return fmt.Errorf("未找到可用的歌词")
	}

	// 保存歌词文件到指定目录
	baseName := strings.TrimSuffix(filepath.Base(trackPath), filepath.Ext(trackPath))
	lrcPath := filepath.Join(lyricsDir, baseName+".lrc")

	if _, err := os.Stat(lrcPath); err == nil {
		backupPath := lrcPath + ".bak"
		os.Rename(lrcPath, backupPath)
	}

	return os.WriteFile(lrcPath, []byte(lyricsContent), 0644)
}
