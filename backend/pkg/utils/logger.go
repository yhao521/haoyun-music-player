package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/yhao521/wailsMusicPlay/backend/pkg/file"
)

// InitLogger 初始化日志系统，将日志输出到运行时目录下的 logs 子目录
func InitLogger() error {
	// 获取日志目录路径
	logDir := file.GetRuntimeLogPath()
	if logDir == "" {
		return fmt.Errorf("无法获取日志目录路径")
	}

	// 生成日志文件名（按日期）
	logFileName := fmt.Sprintf("app-%s.log", time.Now().Format("20060102"))
	logFilePath := filepath.Join(logDir, logFileName)

	// 打开或创建日志文件
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("创建日志文件失败: %w", err)
	}

	// 设置日志输出目标：同时输出到文件和控制台
	multiWriter := io.MultiWriter(logFile, os.Stdout)
	log.SetOutput(multiWriter)

	// 设置日志格式
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)

	log.Printf("✅ 日志系统已初始化，日志文件: %s", logFilePath)
	return nil
}
