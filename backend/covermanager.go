package backend

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// AlbumArt 专辑封面
type AlbumArt struct {
	Data     []byte `json:"data"`      // 图片二进制数据
	MimeType string `json:"mime_type"` // MIME 类型 (image/jpeg, image/png)
	Width    int    `json:"width"`     // 宽度
	Height   int    `json:"height"`    // 高度
}

// CoverManager 专辑封面管理器
type CoverManager struct {
	mu        sync.RWMutex
	cache     map[string]*AlbumArt // 缓存：文件路径 -> 封面
	coverDir  string               // 封面缓存目录
	maxWidth  int                  // 最大宽度
	maxHeight int                  // 最大高度
}

// NewCoverManager 创建专辑封面管理器
func NewCoverManager() *CoverManager {
	// 由于移除了 file 包，这里使用用户主目录作为替代，或者可以根据实际需求调整
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	return &CoverManager{
		cache:     make(map[string]*AlbumArt),
		coverDir:  filepath.Join(homeDir, ".haoyun-music", "covers"),
		maxWidth:  500,
		maxHeight: 500,
	}
}

// Init 初始化封面管理器
func (cm *CoverManager) Init() error {
	// 创建封面缓存目录
	if err := os.MkdirAll(cm.coverDir, 0755); err != nil {
		return fmt.Errorf("创建封面缓存目录失败：%w", err)
	}
	log.Println("✓ 专辑封面管理器初始化完成")
	return nil
}

// ExtractAlbumArt 提取专辑封面
func (cm *CoverManager) ExtractAlbumArt(trackPath string) (*AlbumArt, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 检查缓存
	if art, ok := cm.cache[trackPath]; ok {
		return art, nil
	}

	// 检查文件缓存
	cachedArt := cm.loadFromCache(trackPath)
	if cachedArt != nil {
		cm.cache[trackPath] = cachedArt
		return cachedArt, nil
	}

	// 从音频文件中提取
	ext := strings.ToLower(filepath.Ext(trackPath))
	var art *AlbumArt
	var err error

	switch ext {
	case ".mp3":
		art, err = cm.extractMP3Cover(trackPath)
	case ".flac":
		art, err = cm.extractFLACCover(trackPath)
	default:
		return nil, fmt.Errorf("不支持的音频格式：%s", ext)
	}

	if err != nil {
		log.Printf("⚠️ 提取封面失败 %s：%v", trackPath, err)
		return nil, err
	}

	if art == nil {
		return nil, fmt.Errorf("未找到专辑封面")
	}

	// 调整图片大小
	resizedArt, err := cm.resizeImage(art)
	if err != nil {
		log.Printf("⚠️ 调整图片大小失败：%v", err)
		resizedArt = art
	}

	// 保存到缓存
	cm.cache[trackPath] = resizedArt
	if err := cm.saveToCache(trackPath, resizedArt); err != nil {
		log.Printf("⚠️ 保存封面缓存失败：%v", err)
	}

	log.Printf("✓ 提取专辑封面：%dx%d", resizedArt.Width, resizedArt.Height)
	return resizedArt, nil
}

// extractMP3Cover 从 MP3 文件提取封面（ID3v2 APIC 帧）
// TODO: 需要集成 github.com/bogem/id3v2 库后实现
func (cm *CoverManager) extractMP3Cover(trackPath string) (*AlbumArt, error) {
	// 暂时返回 nil，等待后续实现
	log.Printf("⚠️ MP3 封面提取功能待实现：%s", trackPath)
	return nil, fmt.Errorf("MP3 封面提取功能待实现")
}

// extractFLACCover 从 FLAC 文件提取封面（METADATA_BLOCK_PICTURE）
// TODO: 需要完善 flac 库的使用后实现
func (cm *CoverManager) extractFLACCover(trackPath string) (*AlbumArt, error) {
	// 暂时返回 nil，等待后续实现
	log.Printf("⚠️ FLAC 封面提取功能待实现：%s", trackPath)
	return nil, fmt.Errorf("FLAC 封面提取功能待实现")
}

// readUint32 从字节数组读取 uint32（大端序）
func readUint32(data []byte) uint32 {
	if len(data) < 4 {
		return 0
	}
	return uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3])
}

// getImageDimensions 获取图片尺寸
// 注意：相关依赖库已移除，此功能暂时禁用
func (cm *CoverManager) getImageDimensions(data []byte, mimeType string) (int, int, error) {
	return 0, 0, fmt.Errorf("图片尺寸获取功能暂不可用（依赖库已移除）")
}

// resizeImage 调整图片大小
func (cm *CoverManager) resizeImage(art *AlbumArt) (*AlbumArt, error) {
	// 如果图片已经足够小，直接返回
	if art.Width <= cm.maxWidth && art.Height <= cm.maxHeight {
		return art, nil
	}

	// TODO: 集成 github.com/disintegration/imaging 进行真正的缩放
	// 目前暂时返回原图
	log.Printf("⚠️ 图片缩放功能暂未实现，返回原图：%dx%d", art.Width, art.Height)

	return art, nil
}

// GetCoverDataURL 获取封面的 Data URL（可直接用于 img src）
func (cm *CoverManager) GetCoverDataURL(trackPath string) (string, error) {
	art, err := cm.ExtractAlbumArt(trackPath)
	if err != nil {
		return "", err
	}

	// Base64 编码
	base64Str := base64.StdEncoding.EncodeToString(art.Data)
	dataURL := fmt.Sprintf("data:%s;base64,%s", art.MimeType, base64Str)

	return dataURL, nil
}

// loadFromCache 从文件缓存加载封面
func (cm *CoverManager) loadFromCache(trackPath string) *AlbumArt {
	cacheKey := cm.getCacheKey(trackPath)
	cacheFile := filepath.Join(cm.coverDir, cacheKey+".json")

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil
	}

	var art AlbumArt
	if err := parseCoverJSON(data, &art); err != nil {
		return nil
	}

	// 同时加载图片数据
	imgFile := filepath.Join(cm.coverDir, cacheKey+".dat")
	imgData, err := os.ReadFile(imgFile)
	if err != nil {
		return nil
	}

	art.Data = imgData
	return &art
}

// saveToCache 保存封面到文件缓存
func (cm *CoverManager) saveToCache(trackPath string, art *AlbumArt) error {
	cacheKey := cm.getCacheKey(trackPath)

	// 保存图片数据
	imgFile := filepath.Join(cm.coverDir, cacheKey+".dat")
	if err := os.WriteFile(imgFile, art.Data, 0644); err != nil {
		return err
	}

	// 保存元数据
	meta := struct {
		MimeType string `json:"mime_type"`
		Width    int    `json:"width"`
		Height   int    `json:"height"`
	}{
		MimeType: art.MimeType,
		Width:    art.Width,
		Height:   art.Height,
	}

	metaData, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	metaFile := filepath.Join(cm.coverDir, cacheKey+".json")
	return os.WriteFile(metaFile, metaData, 0644)
}

// getCacheKey 生成缓存键（MD5 哈希）
func (cm *CoverManager) getCacheKey(trackPath string) string {
	hash := md5.Sum([]byte(trackPath))
	return hex.EncodeToString(hash[:])
}

// ClearCache 清除封面缓存
func (cm *CoverManager) ClearCache() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 清除内存缓存
	cm.cache = make(map[string]*AlbumArt)

	// 清除文件缓存
	if err := os.RemoveAll(cm.coverDir); err != nil {
		log.Printf("⚠️ 清除封面缓存目录失败：%v", err)
		return
	}

	// 重新创建目录
	if err := os.MkdirAll(cm.coverDir, 0755); err != nil {
		log.Printf("⚠️ 创建封面缓存目录失败：%v", err)
	}

	log.Println("✓ 清除专辑封面缓存")
}

// GetCachedCover 获取缓存的封面
func (cm *CoverManager) GetCachedCover(trackPath string) *AlbumArt {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if art, ok := cm.cache[trackPath]; ok {
		return art
	}

	return cm.loadFromCache(trackPath)
}

// parseCoverJSON 解析封面 JSON 数据
func parseCoverJSON(data []byte, art *AlbumArt) error {
	var meta struct {
		MimeType string `json:"mime_type"`
		Width    int    `json:"width"`
		Height   int    `json:"height"`
	}

	if err := json.Unmarshal(data, &meta); err != nil {
		return err
	}

	art.MimeType = meta.MimeType
	art.Width = meta.Width
	art.Height = meta.Height

	return nil
}
