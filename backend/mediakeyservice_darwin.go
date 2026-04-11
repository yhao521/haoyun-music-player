//go:build darwin
// +build darwin

package backend

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework Carbon
#import <Cocoa/Cocoa.h>
#import <Carbon/Carbon.h>
#import <IOKit/hidsystem/ev_keymap.h>  // 包含 NX_KEYTYPE_* 定义
#import <ApplicationServices/ApplicationServices.h>  // 用于 AXIsProcessTrusted
#include <stdio.h>  // 用于 fprintf
#include <unistd.h>  // 用于 usleep

// 全局变量
static id g_mediaKeyMonitor = NULL;

// Go 函数声明 (由 CGO 导出)
extern void handleMediaPlayPause();
extern void handleMediaNext();
extern void handleMediaPrevious();

// 辅助函数：直接输出到 stderr（用于调试）
static void logToStderr(const char* message) {
	fprintf(stderr, "[MediaKey-C] %s\n", message);
	fflush(stderr);
}

// 检查辅助功能权限 - 使用官方 API
static int check_accessibility_permission() {
	// 方法 1: 使用 AXIsProcessTrusted (最可靠)
	Boolean trusted = AXIsProcessTrusted();
	if (trusted) {
		logToStderr("✅ AXIsProcessTrusted: 已信任");
		return 1;
	}
	
	// 方法 2: 降级方案 - AppleScript
	logToStderr("⚠️ AXIsProcessTrusted 返回 false，尝试 AppleScript...");
	NSAppleScript *script = [[NSAppleScript alloc] 
		initWithSource:@"tell application \"System Events\" to get UI elements enabled"];
	
	NSDictionary *error = nil;
	NSAppleEventDescriptor *result = [script executeAndReturnError:&error];
	[script release];
	
	if (error != nil) {
		NSString *errorMsg = [[error objectForKey:NSAppleScriptErrorMessage] stringValue];
		const char *errorStr = [errorMsg UTF8String];
		logToStderr("❌ AppleScript 错误: ");
		logToStderr(errorStr);
		return 0;
	}
	
	BOOL isEnabled = [result booleanValue];
	if (isEnabled) {
		logToStderr("✅ AppleScript: UI elements enabled");
	} else {
		logToStderr("❌ AppleScript: UI elements disabled");
	}
	
	return isEnabled ? 1 : 0;
}

// 打开辅助功能权限设置页面
static void open_accessibility_settings() {
	logToStderr("🔗 打开系统偏好设置...");
	[[NSWorkspace sharedWorkspace] 
		openURL:[NSURL URLWithString:@"x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility"]];
}

// 显示权限提示对话框
static void show_permission_alert() {
	logToStderr("📢 显示权限提示对话框");
	NSAlert *alert = [[NSAlert alloc] init];
	[alert setMessageText:@"需要辅助功能权限"];
	[alert setInformativeText:@"为了监听媒体键（F7/F8/F9），需要授予辅助功能权限。\n\n请点击\"打开系统偏好设置\"，然后在\"安全性与隐私 > 隐私 > 辅助功能\"中勾选本应用。"];
	[alert addButtonWithTitle:@"打开系统偏好设置"];
	[alert addButtonWithTitle:@"稍后提醒"];
	[alert addButtonWithTitle:@"取消"];
	[alert setAlertStyle:NSWarningAlertStyle];
	
	NSModalResponse response = [alert runModal];
	[alert release];
	
	if (response == NSAlertFirstButtonReturn) {
		// 用户点击"打开系统偏好设置"
		open_accessibility_settings();
		
		// 等待用户操作（简单延迟）
		usleep(500000); // 0.5秒
		
		// 再次检查权限
		int hasPermission = check_accessibility_permission();
		if (hasPermission) {
			logToStderr("✅ 用户已授予权限");
		} else {
			logToStderr("⚠️ 用户尚未授予权限，请稍后重试");
		}
	}
}

// 注册媒体键 - 使用 NSEvent addGlobalMonitorForEventsMatchingMask
static int register_media_keys() {
	logToStderr("🔧 开始注册媒体键...");
	
	if (g_mediaKeyMonitor != NULL) {
		logToStderr("⚠️ 媒体键监视器已存在，跳过注册");
		return 0; // 已注册
	}
	
	// 首先检查辅助功能权限
	logToStderr("🔍 检查辅助功能权限...");
	int hasPermission = check_accessibility_permission();
	
	if (!hasPermission) {
		logToStderr("❌ 没有辅助功能权限，显示提示对话框");
		// 没有权限，显示提示对话框
		show_permission_alert();
		
		// 再次检查（用户可能刚刚授予了权限）
		hasPermission = check_accessibility_permission();
		if (!hasPermission) {
			logToStderr("❌ 用户仍未授予权限，返回 -2");
			return -2; // 用户未授予权限
		}
		logToStderr("✅ 权限检查通过");
	}
	
	// 监听系统定义的媒体键事件 (NSSystemDefined)
	logToStderr("🎯 创建 NSEvent 全局监视器...");
	NSUInteger eventMask = NSSystemDefined;
	
	// 创建全局事件监视器
	g_mediaKeyMonitor = [[NSClassFromString(@"NSEvent") 
		addGlobalMonitorForEventsMatchingMask:eventMask
		handler:^(NSEvent *event) {
			logToStderr("📨 收到系统事件");
			
			// 检查是否为媒体键事件
			if ([event type] == NSSystemDefined && [event subtype] == 8) {
				logToStderr("🎵 检测到媒体键事件");
				
				// 解析媒体键数据
				int keyCode = (([event data1] & 0xFFFF0000) >> 16);
				int keyFlags = ([event data1] & 0x0000FFFF);
				int keyState = ((keyFlags & 0xFF00) >> 8);
				int isKeyDown = (keyState == 0xA);
				
				char buffer[256];
				snprintf(buffer, sizeof(buffer), "📊 按键信息: keyCode=%d, keyState=%d, isKeyDown=%d", keyCode, keyState, isKeyDown);
				logToStderr(buffer);
				
				if (isKeyDown) {
					switch (keyCode) {
						case NX_KEYTYPE_PLAY:      // 播放/暂停 (F8)
							logToStderr("▶️ 触发: 播放/暂停");
							handleMediaPlayPause();
							break;
						case NX_KEYTYPE_NEXT:      // 下一曲 (F9)
							logToStderr("⏭️ 触发: 下一曲");
							handleMediaNext();
							break;
						case NX_KEYTYPE_PREVIOUS:  // 上一曲 (F7)
							logToStderr("⏮️ 触发: 上一曲");
							handleMediaPrevious();
							break;
						default:
							snprintf(buffer, sizeof(buffer), "❓ 未知媒体键: keyCode=%d", keyCode);
							logToStderr(buffer);
							break;
					}
				}
			} else {
				// 非媒体键事件，忽略（避免日志过多）
				// char buffer[256];
				// snprintf(buffer, sizeof(buffer), "🔕 忽略非媒体键事件: type=%ld, subtype=%ld", (long)[event type], (long)[event subtype]);
				// logToStderr(buffer);
			}
		}] retain];
	
	if (g_mediaKeyMonitor == nil) {
		logToStderr("❌ NSEvent 监视器创建失败");
		return -1; // 注册失败
	}
	
	logToStderr("✅ 媒体键监视器创建成功");
	return 0;
}

// 注销媒体键
static void unregister_media_keys() {
	logToStderr("🔓 注销媒体键监视器...");
	if (g_mediaKeyMonitor != NULL) {
		[NSClassFromString(@"NSEvent") removeMonitor:g_mediaKeyMonitor];
		[g_mediaKeyMonitor release];
		g_mediaKeyMonitor = NULL;
		logToStderr("✅ 媒体键监视器已注销");
	}
}
*/
import "C"
import (
	"log"
)

//export handleMediaPlayPause
func handleMediaPlayPause() {
	if globalMediaKeyService != nil {
		log.Println("🎵 [回调] 执行播放/暂停")
		globalMediaKeyService.handlePlayPause()
	} else {
		log.Println("⚠️ [错误] globalMediaKeyService 为 nil")
	}
}

//export handleMediaNext
func handleMediaNext() {
	if globalMediaKeyService != nil {
		log.Println("🎵 [回调] 执行下一曲")
		globalMediaKeyService.handleNext()
	} else {
		log.Println("⚠️ [错误] globalMediaKeyService 为 nil")
	}
}

//export handleMediaPrevious
func handleMediaPrevious() {
	if globalMediaKeyService != nil {
		log.Println("🎵 [回调] 执行上一曲")
		globalMediaKeyService.handlePrevious()
	} else {
		log.Println("⚠️ [错误] globalMediaKeyService 为 nil")
	}
}

var globalMediaKeyService *MediaKeyService

// platformRegisterMediaKeys macOS 平台注册媒体键
func (mks *MediaKeyService) platformRegisterMediaKeys() error {
	log.Println("========================================")
	log.Println("🍎 [MediaKey] 开始注册 macOS 媒体键...")
	log.Println("💡 [MediaKey] 使用 NSEvent addGlobalMonitorForEventsMatchingMask")
	log.Println("========================================")
	
	globalMediaKeyService = mks
	
	result := C.register_media_keys()
	
	log.Println("========================================")
	switch result {
	case 0:
		// 成功
		log.Println("✅ [MediaKey] macOS 媒体键注册成功！")
		log.Println("📝 [MediaKey] 支持的按键:")
		log.Println("   - F7 (上一曲)")
		log.Println("   - F8 (播放/暂停)")
		log.Println("   - F9 (下一曲)")
		log.Println("💡 [MediaKey] 现在可以测试媒体键功能了")
		log.Println("========================================")
		return nil
		
	case -1:
		// 注册失败（技术错误）
		log.Println("❌ [MediaKey] macOS 媒体键注册失败（技术错误）")
		log.Println("⚠️ [MediaKey] 可能原因:")
		log.Println("   1. NSEvent API 调用失败")
		log.Println("   2. 系统事件循环问题")
		log.Println("   3. 应用未正确初始化")
		log.Println("💡 [MediaKey] 建议: 重启应用并重试")
		log.Println("========================================")
		return nil
		
	case -2:
		// 用户未授予权限
		log.Println("⚠️ [MediaKey] 用户暂未授予辅助功能权限")
		log.Println("💡 [MediaKey] 媒体键功能将不可用，但应用会继续运行")
		log.Println("💡 [MediaKey] 如需启用，请:")
		log.Println("   1. 前往: 系统偏好设置 > 安全性与隐私 > 隐私 > 辅助功能")
		log.Println("   2. 勾选本应用")
		log.Println("   3. 完全退出并重启应用")
		log.Println("========================================")
		return nil
		
	default:
		log.Printf("⚠️ [MediaKey] 未知的注册状态: %d", result)
		log.Println("========================================")
		return nil
	}
}

// platformUnregisterMediaKeys macOS 平台注销媒体键
func (mks *MediaKeyService) platformUnregisterMediaKeys() {
	C.unregister_media_keys()
	globalMediaKeyService = nil
	log.Println("🔓 macOS 媒体键已注销")
}
