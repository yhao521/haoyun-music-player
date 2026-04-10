//go:build darwin
// +build darwin

package backend

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework Carbon
#import <Cocoa/Cocoa.h>
#import <Carbon/Carbon.h>

// 媒体键键码定义 (macOS 虚拟键码)
#define kVK_PlayPause 0xB7
#define kVK_NextTrack 0xB9
#define kVK_PreviousTrack 0xB8

// 全局事件处理器引用
static EventHandlerRef g_mediaKeyHandler = NULL;

// Go 函数声明 (由 CGO 导出)
extern void handleMediaPlayPause();
extern void handleMediaNext();
extern void handleMediaPrevious();

// 媒体键回调函数
OSStatus mediaKeyCallback(EventHandlerCallRef nextHandler, EventRef event, void* userData) {
	UInt32 keyCode;
	GetEventParameter(event, kEventParamKeyCode, typeUInt32, NULL, sizeof(keyCode), NULL, &keyCode);
	
	UInt32 modifiers;
	GetEventParameter(event, kEventParamKeyModifiers, typeUInt32, NULL, sizeof(modifiers), NULL, &modifiers);
	
	// 忽略带修饰键的组合 (Cmd, Ctrl, Alt, Shift)
	if (modifiers != 0) {
		return noErr;
	}
	
	// 处理媒体键事件
	switch (keyCode) {
		case kVK_PlayPause:
			handleMediaPlayPause();
			break;
		case kVK_NextTrack:
			handleMediaNext();
			break;
		case kVK_PreviousTrack:
			handleMediaPrevious();
			break;
	}
	
	return noErr;
}

// 注册媒体键
int register_media_keys() {
	if (g_mediaKeyHandler != NULL) {
		return 0; // 已注册
	}
	
	// 获取应用程序事件目标
	EventTargetRef target = GetApplicationEventTarget();
	
	// 定义要监听的事件类型 (原始键盘按下事件)
	EventTypeSpec eventTypes[1];
	eventTypes[0].eventClass = kEventClassKeyboard;
	eventTypes[0].eventKind = kEventRawKeyDown;
	
	// 创建事件处理器
	EventHandlerUPP upp = NewEventHandlerUPP(mediaKeyCallback);
	OSStatus status = InstallEventHandler(
		target,
		upp,
		1,           // 事件类型数量
		eventTypes,  // 事件类型数组
		NULL,        // 用户数据
		&g_mediaKeyHandler
	);
	
	if (status != noErr) {
		return -1;
	}
	
	return 0;
}

// 注销媒体键
void unregister_media_keys() {
	if (g_mediaKeyHandler != NULL) {
		RemoveEventHandler(g_mediaKeyHandler);
		g_mediaKeyHandler = NULL;
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
	globalMediaKeyService = mks
	
	result := C.register_media_keys()
	if result != 0 {
		log.Println("⚠️ macOS 媒体键注册失败,可能需要辅助功能权限")
		log.Println("💡 请前往: 系统偏好设置 > 安全性与隐私 > 隐私 > 辅助功能")
		return nil // 不返回错误,允许应用继续运行
	}
	
	log.Println("✅ macOS 媒体键注册成功")
	log.Println("📝 支持的按键: 播放/暂停(F8), 下一曲(F9), 上一曲(F7)")
	return nil
}

// platformUnregisterMediaKeys macOS 平台注销媒体键
func (mks *MediaKeyService) platformUnregisterMediaKeys() {
	C.unregister_media_keys()
	globalMediaKeyService = nil
	log.Println("🔓 macOS 媒体键已注销")
}
