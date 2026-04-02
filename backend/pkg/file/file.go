package file

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

func ExistDir(path string) {
	// 判断路径是否存在
	_, err := os.ReadDir(path)
	if err != nil {
		// 不存在就创建
		err = os.MkdirAll(path, fs.ModePerm)
		if err != nil {
			fmt.Println(err)
		}
	}
}

// getAppPath 获取应用主目录
func GetAppPath() string {
	//获取系统我的文档目录
	// homeDir := userdir.GetDataHome()

	// 获取用户数据目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Println("获取用户目录失败：%w", err)
		return ""
	}
	//获取我的文档目录
	return PathExist(fmt.Sprintf("%s/.haoyun-music", homeDir))
}

func GetLibPath() string {

	// 获取用户数据目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Println("获取用户目录失败：%w", err)
		return ""
	}

	// 创建音乐库目录
	return PathExist(filepath.Join(homeDir, ".haoyun-music", "libraries"))
}

func GetResourcePath() string {
	return fmt.Sprintf(GetAppPath()+"%s", "/resource")
}

func GetRuntimePath() string {
	return PathExist(fmt.Sprintf(GetAppPath()+"%s", "/runtime"))
}

func GetRuntimeLogPath() string {
	return PathExist(fmt.Sprintf(GetRuntimePath()+"%s", "/logs"))
}
func GetRuntimeDataPath() string {
	return PathExist(fmt.Sprintf(GetRuntimePath()+"%s", "/data"))
}

// pathExist 判断文件目录是否存在，不存在创建
func PathExist(path string) string {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		_ = os.MkdirAll(path, os.ModePerm)
	}
	return path
}

func GetDownloadPath() string {
	return filepath.Join(os.Getenv("HOME"), "Downloads")
}
