//go:build linux

package backend

import (
	"log"
)

// platformRegisterMediaKeys 在 Linux 上注册系统媒体键
func (mks *MediaKeyService) platformRegisterMediaKeys() error {
	log.Println("🐧 Linux 系统媒体键支持待实现")
	// TODO: 使用 D-Bus MPRIS2 接口或键盘事件监听
	return nil
}

// platformUnregisterMediaKeys 在 Linux 上注销系统媒体键
func (mks *MediaKeyService) platformUnregisterMediaKeys() {
	log.Println("🐧 Linux 系统媒体键清理待实现")
}
