//go:build windows

package backend

import (
	"log"
)

// platformRegisterMediaKeys 在 Windows 上注册系统媒体键
func (mks *MediaKeyService) platformRegisterMediaKeys() error {
	log.Println("🪟 Windows 系统媒体键支持待实现")
	// TODO: 使用 RegisterHotKey API 或 WM_APPCOMMAND 消息
	return nil
}

// platformUnregisterMediaKeys 在 Windows 上注销系统媒体键
func (mks *MediaKeyService) platformUnregisterMediaKeys() {
	log.Println("🪟 Windows 系统媒体键清理待实现")
	// TODO: 使用 UnregisterHotKey API
}
