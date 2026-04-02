package utils

import (
	"fmt"
	"log"
	"os/exec"
)

func OpenWin(uri string) {
	exec.Command(`cmd`, `/c`, `start`, uri).Start()
}
func OpenMac(uri string) {
	exec.Command("open", uri).Run()
}

func Command(cmdStr string, paramStr string) {
	exec.Command(cmdStr, paramStr).Start()
}

func OpenMacDir(path string) {
	Command("open", path)
}

// openDir 打开ubuntu目录
func OpenDir(path string) {
	Command("xdg-open", fmt.Sprintf("file://%s", path))
}

// setSystemBackground 设置系统壁纸
func SetSystemBackground(path string) {
	// 设置壁纸
	cmd := exec.Command("gsettings", "set", "org.gnome.desktop.background", "picture-uri", "file://"+path)
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Error setting wallpaper: %s", err)
	}
}
