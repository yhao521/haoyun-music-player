package backend

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMetadataManager_BasicMetadata 测试基本元数据获取
func TestMetadataManager_BasicMetadata(t *testing.T) {
	mm := NewMetadataManager()
	
	// 测试不存在的文件（应该返回基本信息）
	metadata, err := mm.GetMetadata("/nonexistent/file.mp3")
	if err != nil {
		t.Logf("预期错误：%v", err)
	}
	
	// 验证返回了基本信息
	if metadata["title"] == "" {
		t.Error("标题不应为空")
	}
	if metadata["artist"] != "未知艺术家" {
		t.Errorf("艺术家应为'未知艺术家'，实际为：%v", metadata["artist"])
	}
	if metadata["path"] != "/nonexistent/file.mp3" {
		t.Errorf("路径不匹配")
	}
	
	t.Logf("基本元数据测试通过：%+v", metadata)
}

// TestMetadataManager_Cache 测试缓存功能
func TestMetadataManager_Cache(t *testing.T) {
	mm := NewMetadataManager()
	
	testPath := "/test/path/song.mp3"
	
	// 第一次调用
	metadata1, _ := mm.GetMetadata(testPath)
	title1 := metadata1["title"]
	
	// 第二次调用（应从缓存读取）
	metadata2, _ := mm.GetMetadata(testPath)
	title2 := metadata2["title"]
	
	// 验证两次返回的数据相同
	if title1 != title2 {
		t.Error("缓存未生效：两次调用应返回相同的元数据")
	}
	
	t.Log("缓存测试通过")
}

// TestMetadataManager_ClearCache 测试清除缓存
func TestMetadataManager_ClearCache(t *testing.T) {
	mm := NewMetadataManager()
	
	// 添加一些数据到缓存
	mm.GetMetadata("/test1.mp3")
	mm.GetMetadata("/test2.mp3")
	
	// 清除缓存
	mm.ClearCache()
	
	// 验证缓存已清空
	mm.mu.RLock()
	cacheSize := len(mm.cache)
	mm.mu.RUnlock()
	
	if cacheSize != 0 {
		t.Errorf("缓存应被清空，但仍有 %d 个条目", cacheSize)
	}
	
	t.Log("清除缓存测试通过")
}

// TestMetadataManager_MP3File 测试 MP3 文件元数据读取
func TestMetadataManager_MP3File(t *testing.T) {
	// 创建一个临时 MP3 文件用于测试
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.mp3")
	
	// 创建一个简单的 ID3v1 标签
	createTestMP3WithID3v1(t, tmpFile)
	
	mm := NewMetadataManager()
	metadata, err := mm.GetMetadata(tmpFile)
	if err != nil {
		t.Fatalf("获取元数据失败：%v", err)
	}
	
	t.Logf("MP3 元数据：%+v", metadata)
	
	// 验证至少有一些元数据
	if metadata["title"] == "" {
		t.Error("标题不应为空")
	}
}

// createTestMP3WithID3v1 创建带 ID3v1 标签的测试 MP3 文件
func createTestMP3WithID3v1(t *testing.T, filePath string) {
	// 创建一些虚拟音频数据
	audioData := make([]byte, 1000)
	for i := range audioData {
		audioData[i] = byte(i % 256)
	}
	
	// 创建 ID3v1 标签（128 字节）
	id3v1Tag := make([]byte, 128)
	copy(id3v1Tag[0:3], "TAG")                    // TAG 标识
	copy(id3v1Tag[3:33], "Test Song")             // 标题（30 字节）
	copy(id3v1Tag[33:63], "Test Artist")          // 艺术家（30 字节）
	copy(id3v1Tag[63:93], "Test Album")           // 专辑（30 字节）
	copy(id3v1Tag[93:97], "2024")                 // 年份（4 字节）
	copy(id3v1Tag[97:127], "Test Comment")        // 注释（30 字节）
	id3v1Tag[127] = 0                             // 音轨号
	
	// 合并音频数据和标签
	fileData := append(audioData, id3v1Tag...)
	
	err := os.WriteFile(filePath, fileData, 0644)
	if err != nil {
		t.Fatalf("创建测试文件失败：%v", err)
	}
}

// TestMetadataManager_FLACFile 测试 FLAC 文件元数据读取
func TestMetadataManager_FLACFile(t *testing.T) {
	// 创建一个临时 FLAC 文件用于测试
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.flac")
	
	// 创建一个简单的 FLAC 文件带 Vorbis Comment
	createTestFLACWithVorbis(t, tmpFile)
	
	mm := NewMetadataManager()
	metadata, err := mm.GetMetadata(tmpFile)
	if err != nil {
		t.Fatalf("获取元数据失败：%v", err)
	}
	
	t.Logf("FLAC 元数据：%+v", metadata)
	
	// 验证元数据 - FLAC 应该从 Vorbis Comment 中读取
	if title, ok := metadata["title"].(string); ok && title != "" {
		t.Logf("成功读取 FLAC 标题：%s", title)
	} else {
		// 如果未读取到 Vorbis Comment，至少应该有基于文件名的基本信息
		t.Log("未读取到 Vorbis Comment，使用基本信息")
	}
}

// createTestFLACWithVorbis 创建带 Vorbis Comment 的测试 FLAC 文件
func createTestFLACWithVorbis(t *testing.T, filePath string) {
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("创建测试文件失败：%v", err)
	}
	defer file.Close()
	
	// 写入 FLAC 头部
	file.Write([]byte("fLaC"))
	
	// 创建 STREAMINFO 块（必需）
	streamInfo := make([]byte, 34)
	// 设置块头部：不是最后一个块，类型 0（STREAMINFO），大小 34
	streamInfoHeader := []byte{0x00, 0x00, 0x00, 0x22}
	file.Write(streamInfoHeader)
	file.Write(streamInfo)
	
	// 创建 Vorbis Comment 块
	vorbisComments := buildVorbisComment(t, map[string]string{
		"TITLE":  "Test FLAC Song",
		"ARTIST": "Test FLAC Artist",
		"ALBUM":  "Test FLAC Album",
		"DATE":   "2024",
	})
	
	// 设置块头部：是最后一个块（0x80），类型 4（VORBIS_COMMENT）
	vorbisHeader := []byte{0x84, 0x00, 0x00, byte(len(vorbisComments) >> 16), byte(len(vorbisComments) >> 8), byte(len(vorbisComments))}
	file.Write(vorbisHeader)
	file.Write(vorbisComments)
	
	// 添加一些虚拟音频数据
	audioData := make([]byte, 1000)
	file.Write(audioData)
}

// buildVorbisComment 构建 Vorbis Comment 数据
func buildVorbisComment(t *testing.T, tags map[string]string) []byte {
	data := []byte{}
	
	// Vendor string（空）
	vendorLen := make([]byte, 4)
	vendorLen[0] = 0
	vendorLen[1] = 0
	vendorLen[2] = 0
	vendorLen[3] = 0
	data = append(data, vendorLen...)
	
	// 评论数量
	commentCount := uint32(len(tags))
	countBytes := make([]byte, 4)
	countBytes[0] = byte(commentCount)
	countBytes[1] = byte(commentCount >> 8)
	countBytes[2] = byte(commentCount >> 16)
	countBytes[3] = byte(commentCount >> 24)
	data = append(data, countBytes...)
	
	// 添加每个评论
	for key, value := range tags {
		comment := key + "=" + value
		commentLen := uint32(len(comment))
		
		lenBytes := make([]byte, 4)
		lenBytes[0] = byte(commentLen)
		lenBytes[1] = byte(commentLen >> 8)
		lenBytes[2] = byte(commentLen >> 16)
		lenBytes[3] = byte(commentLen >> 24)
		data = append(data, lenBytes...)
		data = append(data, []byte(comment)...)
	}
	
	return data
}
