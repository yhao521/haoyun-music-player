//go:build darwin
// +build darwin

package backend

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework Carbon
#import <Cocoa/Cocoa.h>
#import <Carbon/Carbon.h>
#import <IOKit/hidsystem/ev_keymap.h>  // 包含 NX_KEYTYPE_* 定义

// 全局变量
static id g_mediaKeyMonitor = NULL;

// Go 函数声明 (由 CGO 导出)
extern void handleMediaPlayPause();
extern void handleMediaNext();
extern void handleMediaPrevious();

// 注册媒体键 - 使用 NSEvent addGlobalMonitorForEventsMatchingMask
static int register_media_keys() {
	if (g_mediaKeyMonitor != NULL) {
		return 0; // 已注册
	}
	
	// 监听系统定义的媒体键事件 (NSSystemDefined)
	NSUInteger eventMask = NSSystemDefined;
	
	// 创建全局事件监视器
	g_mediaKeyMonitor = [[NSClassFromString(@"NSEvent") 
		addGlobalMonitorForEventsMatchingMask:eventMask
		handler:^(NSEvent *event) {
			// 检查是否为媒体键事件
			if ([event type] == NSSystemDefined && [event subtype] == 8) {
				// 解析媒体键数据
				int keyCode = (([event data1] & 0xFFFF0000) >> 16);
				int keyFlags = ([event data1] & 0x0000FFFF);
				int keyState = ((keyFlags & 0xFF00) >> 8);
				int isKeyDown = (keyState == 0xA);
				
				if (isKeyDown) {
					switch (keyCode) {
						case NX_KEYTYPE_PLAY:      // 播放/暂停 (F8)
							handleMediaPlayPause();
							break;
						case NX_KEYTYPE_NEXT:      // 下一曲 (F9)
							handleMediaNext();
							break;
						case NX_KEYTYPE_PREVIOUS:  // 上一曲 (F7)
							handleMediaPrevious();
							break;
					}
				}
			}
		}] retain];
	
	if (g_mediaKeyMonitor == nil) {
		return -1; // 注册失败
	}
	
	return 0;
}

// 注销媒体键
static void unregister_media_keys() {
	if (g_mediaKeyMonitor != NULL) {
		[NSClassFromString(@"NSEvent") removeMonitor:g_mediaKeyMonitor];
		[g_mediaKeyMonitor release];
		g_mediaKeyMonitor = NULL;
	}
}
*/
import "C"

import "log"

//export handleMediaPlayPause
func handleMediaPlayPause() {
	if globalMediaKeyService != nil {
		globalMediaKeyService.handlePlayPause()
	}
}

//export handleMediaNext
func handleMediaNext() {
	if globalMediaKeyService != nil {
		globalMediaKeyService.handleNext()
	}
}

//export handleMediaPrevious
func handleMediaPrevious() {
	if globalMediaKeyService != nil {
		globalMediaKeyService.handlePrevious()
	}
}

var globalMediaKeyService *MediaKeyService

// platformRegisterMediaKeys macOS 平台注册媒体键
func (mks *MediaKeyService) platformRegisterMediaKeys() error {
	log.Println("🍎 正在注册 macOS 媒体键...")
	log.Println("💡 使用 NSEvent addGlobalMonitorForEventsMatchingMask 监听系统级媒体键")
	globalMediaKeyService = mks
	
	result := C.register_media_keys()
	if result != 0 {
		log.Println("❌ macOS 媒体键注册失败")
		log.Println("⚠️ 请检查：系统偏好设置 > 安全性与隐私 > 隐私 > 辅助功能")
		log.Println("⚠️ 确保应用已获得辅助功能权限后重试")
		return nil // 不返回错误，允许应用继续运行
	}
	
	log.Println("✅ macOS 媒体键注册成功")
	log.Println("📝 支持的按键:")
	log.Println("   - F7 (上一曲)")
	log.Println("   - F8 (播放/暂停)")
	log.Println("   - F9 (下一曲)")
	log.Println("💡 提示: 如果按键无响应，请重启应用并确保已授予辅助功能权限")
	return nil
}

// platformUnregisterMediaKeys macOS 平台注销媒体键
func (mks *MediaKeyService) platformUnregisterMediaKeys() {
	C.unregister_media_keys()
	globalMediaKeyService = nil
	log.Println("🔓 macOS 媒体键已注销")
}
