package backend

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf16"
)

// MetadataManager 元数据管理器
type MetadataManager struct {
	mu             sync.RWMutex
	cache          map[string]map[string]interface{} // 缓存：文件路径 -> 元数据
	durationReader *AudioDurationReader              // 时长读取器
}

// NewMetadataManager 创建元数据管理器
func NewMetadataManager() *MetadataManager {
	return &MetadataManager{
		cache:          make(map[string]map[string]interface{}),
		durationReader: NewAudioDurationReader(),
	}
}

// GetMetadata 获取音频文件元数据（包含时长）
func (mm *MetadataManager) GetMetadata(filePath string) (map[string]interface{}, error) {
	mm.mu.RLock()
	// 检查缓存
	if metadata, ok := mm.cache[filePath]; ok {
		mm.mu.RUnlock()
		return metadata, nil
	}
	mm.mu.RUnlock()

	// 读取元数据
	metadata, err := mm.readMetadata(filePath)
	if err != nil {
		log.Printf("⚠️ 读取元数据失败 %s：%v", filePath, err)
		// 返回基本元数据
		metadata = mm.getBasicMetadata(filePath)
	}

	// 读取时长信息
	duration, durationErr := mm.durationReader.GetDuration(filePath)
	if durationErr == nil {
		metadata["duration"] = duration
	} else {
		metadata["duration"] = int64(0)
		// log.Printf("⚠️ 读取时长失败 %s：%v", filePath, durationErr)
	}

	// 缓存结果
	mm.mu.Lock()
	mm.cache[filePath] = metadata
	mm.mu.Unlock()

	return metadata, nil
}

// readMetadata 从音频文件中读取元数据
func (mm *MetadataManager) readMetadata(filePath string) (map[string]interface{}, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".mp3":
		return mm.readMP3Metadata(filePath)
	case ".flac":
		return mm.readFLACMetadata(filePath)
	default:
		return nil, fmt.Errorf("不支持的音频格式：%s", ext)
	}
}

// readMP3Metadata 读取 MP3 文件元数据（ID3 标签）
func (mm *MetadataManager) readMP3Metadata(filePath string) (map[string]interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败：%w", err)
	}
	defer file.Close()

	// 首先尝试读取 ID3v2 标签
	id3v2Data, err := mm.readID3v2(file)
	if err == nil && id3v2Data != nil {
		return id3v2Data, nil
	}

	// 如果 ID3v2 读取失败，尝试 ID3v1
	file.Seek(0, io.SeekStart)
	id3v1Data, err := mm.readID3v1(file)
	if err == nil && id3v1Data != nil {
		return id3v1Data, nil
	}

	// 都失败了，返回基本信息
	return mm.getBasicMetadata(filePath), nil
}

// readID3v2 读取 ID3v2 标签
func (mm *MetadataManager) readID3v2(file *os.File) (map[string]interface{}, error) {
	// 读取前 10 字节（ID3v2 头部）
	header := make([]byte, 10)
	if _, err := file.Read(header); err != nil {
		return nil, err
	}

	// 检查是否为 ID3v2
	if string(header[:3]) != "ID3" {
		return nil, fmt.Errorf("不是 ID3v2 标签")
	}

	// 获取版本
	version := header[3]
	_ = version // 暂时不使用

	// 获取标签大小（同步安全整数）
	tagSize := int(header[6])<<21 | int(header[7])<<14 | int(header[8])<<7 | int(header[9])

	// 读取标签数据
	tagData := make([]byte, tagSize)
	if _, err := file.Read(tagData); err != nil {
		return nil, err
	}

	// 解析帧
	metadata := mm.parseID3v2Frames(tagData)

	// 添加基本信息
	basicInfo := mm.getBasicMetadata(file.Name())
	for k, v := range basicInfo {
		if _, exists := metadata[k]; !exists {
			metadata[k] = v
		}
	}

	return metadata, nil
}

// parseID3v2Frames 解析 ID3v2 帧
func (mm *MetadataManager) parseID3v2Frames(data []byte) map[string]interface{} {
	metadata := make(map[string]interface{})
	offset := 0

	for offset < len(data)-10 {
		// 读取帧头（10 字节）
		if offset+10 > len(data) {
			break
		}

		frameID := string(data[offset : offset+4])
		frameSize := int(binary.BigEndian.Uint32(data[offset+4 : offset+8]))
		_ = data[offset+8] // 标志位
		_ = data[offset+9] // 标志位

		offset += 10

		// 检查帧大小是否有效
		if frameSize <= 0 || offset+frameSize > len(data) {
			break
		}

		frameData := data[offset : offset+frameSize]
		offset += frameSize

		// 解析常见帧
		switch frameID {
		case "TIT2": // 标题
			metadata["title"] = mm.decodeTextFrame(frameData)
		case "TPE1": // 艺术家
			metadata["artist"] = mm.decodeTextFrame(frameData)
		case "TALB": // 专辑
			metadata["album"] = mm.decodeTextFrame(frameData)
		case "TYER", "TDRC": // 年份
			metadata["year"] = mm.decodeTextFrame(frameData)
		case "TCON": // 流派
			metadata["genre"] = mm.decodeTextFrame(frameData)
		case "TRCK": // 音轨号
			metadata["track"] = mm.decodeTextFrame(frameData)
		case "COMM": // 注释
			metadata["comment"] = mm.decodeTextFrame(frameData)
		}
	}

	return metadata
}

// decodeTextFrame 解码文本帧
func (mm *MetadataManager) decodeTextFrame(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	// 第一个字节是编码
	encoding := data[0]
	textData := data[1:]

	var text string
	switch encoding {
	case 0: // ISO-8859-1
		text = string(textData)
	case 1: // UTF-16 with BOM
		text = mm.decodeUTF16(textData)
	case 2: // UTF-16BE
		text = mm.decodeUTF16BE(textData)
	case 3: // UTF-8
		text = string(textData)
	default:
		text = string(textData)
	}

	// 去除空字符
	text = strings.TrimRight(text, "\x00")
	return strings.TrimSpace(text)
}

// decodeUTF16 解码 UTF-16 编码（带 BOM）
func (mm *MetadataManager) decodeUTF16(data []byte) string {
	if len(data) < 2 {
		return ""
	}

	// 检查 BOM
	bom := binary.LittleEndian.Uint16(data[:2])
	var utf16Bytes []uint16

	if bom == 0xFEFF {
		// Little Endian
		utf16Bytes = mm.decodeUTF16LE(data[2:])
	} else if bom == 0xFFFE {
		// Big Endian
		utf16Bytes = mm.decodeUTF16BEBytes(data[2:])
	} else {
		// 没有 BOM，默认 Little Endian
		utf16Bytes = mm.decodeUTF16LE(data)
	}

	if len(utf16Bytes) == 0 {
		return ""
	}

	// 转换 UTF-16 到 UTF-8
	runes := utf16.Decode(utf16Bytes)
	return string(runes)
}

// decodeUTF16LE 解码 Little Endian UTF-16
func (mm *MetadataManager) decodeUTF16LE(data []byte) []uint16 {
	if len(data)%2 != 0 {
		data = data[:len(data)-1] // 确保偶数长度
	}

	var utf16Bytes []uint16
	for i := 0; i < len(data); i += 2 {
		if i+1 >= len(data) {
			break
		}
		char := binary.LittleEndian.Uint16(data[i : i+2])
		if char == 0 {
			break // 遇到空字符停止
		}
		utf16Bytes = append(utf16Bytes, char)
	}
	return utf16Bytes
}

// decodeUTF16BE 解码 Big Endian UTF-16
func (mm *MetadataManager) decodeUTF16BE(data []byte) string {
	utf16Bytes := mm.decodeUTF16BEBytes(data)
	if len(utf16Bytes) == 0 {
		return ""
	}
	runes := utf16.Decode(utf16Bytes)
	return string(runes)
}

// decodeUTF16BEBytes 解码 Big Endian UTF-16 字节数组
func (mm *MetadataManager) decodeUTF16BEBytes(data []byte) []uint16 {
	if len(data)%2 != 0 {
		data = data[:len(data)-1] // 确保偶数长度
	}

	var utf16Bytes []uint16
	for i := 0; i < len(data); i += 2 {
		if i+1 >= len(data) {
			break
		}
		char := binary.BigEndian.Uint16(data[i : i+2])
		if char == 0 {
			break // 遇到空字符停止
		}
		utf16Bytes = append(utf16Bytes, char)
	}
	return utf16Bytes
}

// readID3v1 读取 ID3v1 标签（文件末尾 128 字节）
func (mm *MetadataManager) readID3v1(file *os.File) (map[string]interface{}, error) {
	// 定位到文件末尾前 128 字节
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if fileInfo.Size() < 128 {
		return nil, fmt.Errorf("文件太小，无法包含 ID3v1 标签")
	}

	_, err = file.Seek(-128, io.SeekEnd)
	if err != nil {
		return nil, err
	}

	// 读取 128 字节
	data := make([]byte, 128)
	if _, err := file.Read(data); err != nil {
		return nil, err
	}

	// 检查 TAG 标识
	if string(data[:3]) != "TAG" {
		return nil, fmt.Errorf("不是 ID3v1 标签")
	}

	metadata := make(map[string]interface{})

	// 解析字段（固定长度，需要去除空字符）
	title := strings.TrimRight(string(data[3:33]), "\x00")
	artist := strings.TrimRight(string(data[33:63]), "\x00")
	album := strings.TrimRight(string(data[63:93]), "\x00")
	year := strings.TrimRight(string(data[93:97]), "\x00")
	comment := strings.TrimRight(string(data[97:127]), "\x00")

	if title != "" {
		metadata["title"] = title
	}
	if artist != "" {
		metadata["artist"] = artist
	}
	if album != "" {
		metadata["album"] = album
	}
	if year != "" {
		metadata["year"] = year
	}
	if comment != "" {
		metadata["comment"] = comment
	}

	// 添加基本信息
	basicInfo := mm.getBasicMetadata(file.Name())
	for k, v := range basicInfo {
		if _, exists := metadata[k]; !exists {
			metadata[k] = v
		}
	}

	return metadata, nil
}

// readFLACMetadata 读取 FLAC 文件元数据
func (mm *MetadataManager) readFLACMetadata(filePath string) (map[string]interface{}, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败：%w", err)
	}
	defer file.Close()

	metadata := make(map[string]interface{})

	// 读取 FLAC 文件的元数据块
	// FLAC 文件格式：前 4 字节是 "fLaC"，然后是元数据块
	header := make([]byte, 4)
	if _, err := file.Read(header); err != nil {
		return nil, fmt.Errorf("读取 FLAC 头部失败：%w", err)
	}

	if string(header) != "fLaC" {
		return nil, fmt.Errorf("不是有效的 FLAC 文件")
	}

	// 读取元数据块
	for {
		// 读取元数据块头部（4 字节）
		blockHeader := make([]byte, 4)
		if _, err := file.Read(blockHeader); err != nil {
			break
		}

		// 解析块头部
		isLastBlock := (blockHeader[0] & 0x80) != 0
		blockType := blockHeader[0] & 0x7F
		blockSize := uint32(blockHeader[1])<<16 | uint32(blockHeader[2])<<8 | uint32(blockHeader[3])

		// 读取块数据
		blockData := make([]byte, blockSize)
		if _, err := file.Read(blockData); err != nil {
			break
		}

		// 处理 Vorbis Comment 块（类型 4）
		if blockType == 4 {
			vorbisMetadata := mm.parseVorbisComment(blockData)
			for k, v := range vorbisMetadata {
				metadata[k] = v
			}
		}

		// 如果是最后一个块，退出循环
		if isLastBlock {
			break
		}
	}

	// 添加基本信息
	basicInfo := mm.getBasicMetadata(filePath)
	for k, v := range basicInfo {
		if _, exists := metadata[k]; !exists {
			metadata[k] = v
		}
	}

	return metadata, nil
}

// parseVorbisComment 解析 Vorbis Comment
func (mm *MetadataManager) parseVorbisComment(data []byte) map[string]interface{} {
	metadata := make(map[string]interface{})

	if len(data) < 8 {
		return metadata
	}

	offset := 0

	// 读取 vendor string 长度
	if offset+4 > len(data) {
		return metadata
	}
	vendorLen := binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4

	// 跳过 vendor string
	if offset+int(vendorLen) > len(data) {
		return metadata
	}
	offset += int(vendorLen)

	// 读取用户评论数量
	if offset+4 > len(data) {
		return metadata
	}
	commentCount := binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4

	// 读取每个评论
	for i := uint32(0); i < commentCount && offset < len(data); i++ {
		// 读取评论长度
		if offset+4 > len(data) {
			break
		}
		commentLen := binary.LittleEndian.Uint32(data[offset : offset+4])
		offset += 4

		// 读取评论内容
		if offset+int(commentLen) > len(data) {
			break
		}
		comment := string(data[offset : offset+int(commentLen)])
		offset += int(commentLen)

		// 解析键值对
		parts := strings.SplitN(comment, "=", 2)
		if len(parts) == 2 {
			key := strings.ToUpper(parts[0])
			value := parts[1]

			switch key {
			case "TITLE":
				metadata["title"] = value
			case "ARTIST":
				metadata["artist"] = value
			case "ALBUM":
				metadata["album"] = value
			case "DATE":
				metadata["year"] = value
			case "GENRE":
				metadata["genre"] = value
			case "TRACKNUMBER":
				metadata["track"] = value
			case "DESCRIPTION", "COMMENT":
				metadata["comment"] = value
			}
		}
	}

	return metadata
}

// getBasicMetadata 获取基本元数据（从文件名和路径）
func (mm *MetadataManager) getBasicMetadata(filePath string) map[string]interface{} {
	filename := filepath.Base(filePath)
	// 去除扩展名
	nameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))

	return map[string]interface{}{
		"title":    nameWithoutExt,
		"artist":   "未知艺术家",
		"album":    "未知专辑",
		"path":     filePath,
		"year":     "",
		"genre":    "",
		"track":    "",
		"comment":  "",
		"duration": int64(0), // 默认时长为 0
	}
}

// ClearCache 清除元数据缓存
func (mm *MetadataManager) ClearCache() {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.cache = make(map[string]map[string]interface{})
	if mm.durationReader != nil {
		mm.durationReader.ClearCache()
	}
	log.Println("✓ 元数据缓存已清除")
}
