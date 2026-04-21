package main

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
)

// openSupportPage 打开支持页面
func openSupportPage() {
	url := "https://yhao521.github.io/2026/04/21/SUPPORT-Haoyun-Music/"
	log.Printf("🌐 打开支持页面: %s", url)

	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		log.Printf("❌ 打开浏览器失败: %v", err)
	}
}
