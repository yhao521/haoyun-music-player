//go:build darwin

package backend

import (
	"log"
)

// platformRegisterMediaKeys registers system media keys on macOS
// Note: System media keys require Accessibility permission, so we skip this and use global hotkeys instead
func (mks *MediaKeyService) platformRegisterMediaKeys() error {
	log.Println("⚠️ macOS 系统媒体键已禁用（需要辅助功能权限）")
	log.Println("💡 使用全局快捷键方案 (Ctrl+Shift+P/N/B)")
	// 不注册系统媒体键，直接使用全局快捷键
	return nil
}

// platformUnregisterMediaKeys unregisters system media keys on macOS
func (mks *MediaKeyService) platformUnregisterMediaKeys() {
	log.Println("🔓 macOS 系统媒体键清理完成（无操作）")
	// 无需清理，因为没有注册
}
