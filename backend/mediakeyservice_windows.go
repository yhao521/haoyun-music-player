//go:build windows
// +build windows

package backend

/*
#include <windows.h>

// 热键 ID 定义
#define HK_PLAY_PAUSE 1
#define HK_NEXT 2
#define HK_PREV 3

// 全局变量
static HWND g_hotkeyWindow = NULL;
static int g_registered = 0;

// Go 函数声明 (由 CGO 导出)
extern void handleMediaPlayPause();
extern void handleMediaNext();
extern void handleMediaPrevious();

// 窗口过程函数
LRESULT CALLBACK HotkeyWndProc(HWND hWnd, UINT msg, WPARAM wParam, LPARAM lParam) {
	if (msg == WM_HOTKEY) {
		int hotkeyID = (int)wParam;
		switch (hotkeyID) {
			case HK_PLAY_PAUSE:
				handleMediaPlayPause();
				break;
			case HK_NEXT:
				handleMediaNext();
				break;
			case HK_PREV:
				handleMediaPrevious();
				break;
		}
	}
	return DefWindowProc(hWnd, msg, wParam, lParam);
}

// 注册媒体热键
int register_media_hotkeys() {
	if (g_registered) {
		return 0; // 已注册
	}
	
	// 注册窗口类
	WNDCLASSEX wc = {0};
	wc.cbSize = sizeof(WNDCLASSEX);
	wc.lpfnWndProc = HotkeyWndProc;
	wc.hInstance = GetModuleHandle(NULL);
	wc.lpszClassName = "MediaKeyHotkeyWindow";
	
	if (!RegisterClassEx(&wc)) {
		return -1;
	}
	
	// 创建隐藏窗口用于接收热键消息
	g_hotkeyWindow = CreateWindowEx(
		0,
		"MediaKeyHotkeyWindow",
		"Media Key Hotkey Window",
		WS_OVERLAPPEDWINDOW,
		CW_USEDEFAULT, CW_USEDEFAULT,
		CW_USEDEFAULT, CW_USEDEFAULT,
		NULL, NULL,
		GetModuleHandle(NULL),
		NULL
	);
	
	if (!g_hotkeyWindow) {
		return -1;
	}
	
	// 注册全局热键 (无修饰键)
	BOOL success1 = RegisterHotKey(
		g_hotkeyWindow,
		HK_PLAY_PAUSE,
		0,              // 无修饰键
		VK_MEDIA_PLAY_PAUSE
	);
	
	BOOL success2 = RegisterHotKey(
		g_hotkeyWindow,
		HK_NEXT,
		0,
		VK_MEDIA_NEXT_TRACK
	);
	
	BOOL success3 = RegisterHotKey(
		g_hotkeyWindow,
		HK_PREV,
		0,
		VK_MEDIA_PREV_TRACK
	);
	
	if (success1 || success2 || success3) {
		g_registered = 1;
		return 0;
	}
	
	return -1;
}

// 注销媒体热键
void unregister_media_hotkeys() {
	if (!g_registered) {
		return;
	}
	
	if (g_hotkeyWindow) {
		UnregisterHotKey(g_hotkeyWindow, HK_PLAY_PAUSE);
		UnregisterHotKey(g_hotkeyWindow, HK_NEXT);
		UnregisterHotKey(g_hotkeyWindow, HK_PREV);
		DestroyWindow(g_hotkeyWindow);
		g_hotkeyWindow = NULL;
	}
	
	g_registered = 0;
}

// 轮询消息队列 (需要在主线程调用)
void poll_media_hotkeys() {
	MSG msg;
	while (PeekMessage(&msg, NULL, 0, 0, PM_REMOVE)) {
		if (msg.message == WM_HOTKEY) {
			HotkeyWndProc(msg.hwnd, msg.message, msg.wParam, msg.lParam);
		}
	}
}
*/
import "C"

import (
	"log"
	"runtime"
	"time"
)

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

// platformRegisterMediaKeys Windows 平台注册媒体键
func (mks *MediaKeyService) platformRegisterMediaKeys() error {
	log.Println("🪟 正在注册 Windows 媒体键...")
	globalMediaKeyService = mks
	
	result := C.register_media_hotkeys()
	if result != 0 {
		log.Println("⚠️ Windows 媒体键注册失败")
		return nil // 不返回错误,允许应用继续运行
	}
	
	// 启动 goroutine 轮询热键消息
	go func() {
		for {
			runtime.LockOSThread()
			C.poll_media_hotkeys()
			runtime.UnlockOSThread()
			time.Sleep(50 * time.Millisecond)
		}
	}()
	
	log.Println("✅ Windows 媒体键注册成功")
	log.Println("📝 支持多媒体键盘的播放/暂停、上一曲、下一曲键")
	return nil
}

// platformUnregisterMediaKeys Windows 平台注销媒体键
func (mks *MediaKeyService) platformUnregisterMediaKeys() {
	C.unregister_media_hotkeys()
	globalMediaKeyService = nil
	log.Println("🔓 Windows 媒体键已注销")
}
