//go:build darwin
// +build darwin

package backend

import (
	"log"

	"golang.design/x/hotkey"
)

// getPrimaryModifier 返回 macOS 平台的主修饰键 (Command)
func (mks *MediaKeyService) getPrimaryModifier() hotkey.Modifier {
	log.Println("🍎 macOS 平台: 使用 Command 键作为主修饰键")
	return hotkey.ModCmd
}
