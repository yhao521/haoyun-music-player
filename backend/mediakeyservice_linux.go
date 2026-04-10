//go:build linux
// +build linux

package backend

import "log"

// platformRegisterMediaKeys Linux 平台注册媒体键 (暂未实现)
func (mks *MediaKeyService) platformRegisterMediaKeys() error {
	log.Println("🐧 Linux 平台媒体键支持暂未实现")
	log.Println("💡 提示: 可以使用 xbindkeys 或系统设置配置快捷键")
	return nil
}

// platformUnregisterMediaKeys Linux 平台注销媒体键
func (mks *MediaKeyService) platformUnregisterMediaKeys() {
	log.Println("🔓 Linux 媒体键已注销 (无操作)")
}
