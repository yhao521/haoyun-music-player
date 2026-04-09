package backend

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// OrganizeService 整理音乐服务
type OrganizeService struct {
	libraryManager *LibraryManager // 音乐库管理
}

// NewOrganizeService 创建整理音乐服务实例
func NewOrganizeService() *OrganizeService {
	return &OrganizeService{
		libraryManager: NewLibraryManager(),
	}
}

// SetLibraryManager 设置音乐库管理器
func (s *OrganizeService) SetLibraryManager(lm *LibraryManager) {
	s.libraryManager = lm
}

// OrganizeLibrary 整理音乐库：将音乐文件和歌词文件分别移动到子目录
func (s *OrganizeService) OrganizeLibrary() error {
	currentLib := s.libraryManager.GetCurrentLibrary()
	if currentLib == nil {
		return fmt.Errorf("当前没有音乐库")
	}

	if currentLib.Path == "" {
		return fmt.Errorf("音乐库路径为空")
	}

	libPath := currentLib.Path
	log.Printf("📁 开始整理音乐库：%s (路径：%s)", currentLib.Name, libPath)

	// 创建音乐子目录和歌词子目录
	musicDir := filepath.Join(libPath, "LIB_MUSIC")
	lyricsDir := filepath.Join(libPath, "LIB_LYRIC")

	// 创建目录（如果不存在）
	if err := os.MkdirAll(musicDir, 0755); err != nil {
		return fmt.Errorf("创建音乐目录失败：%w", err)
	}
	if err := os.MkdirAll(lyricsDir, 0755); err != nil {
		return fmt.Errorf("创建歌词目录失败：%w", err)
	}

	log.Printf("✓ 音乐目录：%s", musicDir)
	log.Printf("✓ 歌词目录：%s", lyricsDir)

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

	// 定义歌词文件扩展名
	lyricsExtensions := map[string]bool{
		".lrc": true,
		".txt": true,
	}

	// 统计移动的文件数量
	musicCount := 0
	lyricsCount := 0
	errorCount := 0

	// 使用递归遍历所有子目录
	err := s.walkAndOrganize(libPath, musicDir, lyricsDir, musicExtensions, lyricsExtensions, &musicCount, &lyricsCount, &errorCount)
	if err != nil {
		return err
	}

	log.Printf("✓ 整理完成：音乐文件 %d 个，歌词文件 %d 个，失败 %d 个", musicCount, lyricsCount, errorCount)

	// 更新音乐库的路径索引（因为文件路径变了）
	if musicCount > 0 || lyricsCount > 0 {
		log.Println("🔄 重新加载音乐库索引...")
		if err := s.libraryManager.ReloadCurrentLibrary(); err != nil {
			log.Printf("⚠️ 重新加载音乐库失败：%v", err)
		}
	}

	if errorCount > 0 {
		return fmt.Errorf("整理完成，但有 %d 个文件移动失败", errorCount)
	}

	return nil
}

// walkAndOrganize 递归遍历目录并整理文件
func (s *OrganizeService) walkAndOrganize(
	dirPath, musicDir, lyricsDir string,
	musicExtensions, lyricsExtensions map[string]bool,
	musicCount, lyricsCount, errorCount *int,
) error {
	// 读取目录中的所有条目
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		log.Printf("⚠️ 读取目录失败 %s：%v", dirPath, err)
		*errorCount++
		return nil // 继续处理其他目录
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())

		// 如果是目录
		if entry.IsDir() {
			// 跳过音乐和歌词分类目录（避免递归到自己）
			if fullPath == musicDir || fullPath == lyricsDir {
				continue
			}

			// 递归处理子目录
			if err := s.walkAndOrganize(fullPath, musicDir, lyricsDir, musicExtensions, lyricsExtensions, musicCount, lyricsCount, errorCount); err != nil {
				return err
			}
			continue
		}

		// 处理文件
		fileName := entry.Name()
		ext := strings.ToLower(filepath.Ext(fileName))

		// 检查是否是音乐文件
		if musicExtensions[ext] {
			// 检查是否已经在目标目录中
			if strings.HasPrefix(dirPath, musicDir) {
				continue // 已经在音乐目录中，跳过
			}

			newPath := filepath.Join(musicDir, fileName)
			if err := os.Rename(fullPath, newPath); err != nil {
				log.Printf("⚠️ 移动音乐文件失败 %s：%v", fileName, err)
				(*errorCount)++
				continue
			}
			(*musicCount)++
			log.Printf("  ✓ 移动音乐文件：%s → LIB_MUSIC/", fileName)
			continue
		}

		// 检查是否是歌词文件
		if lyricsExtensions[ext] {
			// 检查是否已经在目标目录中
			if strings.HasPrefix(dirPath, lyricsDir) {
				continue // 已经在歌词目录中，跳过
			}

			newPath := filepath.Join(lyricsDir, fileName)
			if err := os.Rename(fullPath, newPath); err != nil {
				log.Printf("⚠️ 移动歌词文件失败 %s：%v", fileName, err)
				(*errorCount)++
				continue
			}
			(*lyricsCount)++
			log.Printf("  ✓ 移动歌词文件：%s → LIB_LYRIC/", fileName)
			continue
		}
	}

	return nil
}
