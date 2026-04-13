//go:build !darwin
// +build !darwin

package backend

import (
	"log"
	"runtime"

	"golang.design/x/hotkey"
)

// getPrimaryModifier 返回非 macOS 平台的主修饰键 (Ctrl)
func (mks *MediaKeyService) getPrimaryModifier() hotkey.Modifier {
	log.Printf("💻 %s 平台: 使用 Ctrl 键作为主修饰键", runtime.GOOS)
	return hotkey.ModCtrl
}
