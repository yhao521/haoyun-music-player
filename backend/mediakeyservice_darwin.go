//go:build darwin

package backend

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework AppKit -framework Foundation -framework ApplicationServices
#import <AppKit/AppKit.h>
#import <Foundation/Foundation.h>
#import <ApplicationServices/ApplicationServices.h>
// Media key codes (from IOKit/hidsystem/ev_keymap.h)
#define NX_KEYTYPE_PLAY      16
#define NX_KEYTYPE_NEXT      17
#define NX_KEYTYPE_PREVIOUS  18
// Forward declarations of Go exported functions
extern void goPlayPauseHandler(void);
extern void goNextHandler(void);
extern void goPreviousHandler(void);
static id eventMonitor = nil;
// Check accessibility permission and prompt user if needed
static BOOL checkAccessibilityPermission() {
    NSDictionary *options = @{(__bridge id)kAXTrustedCheckOptionPrompt: @YES};
    BOOL trusted = AXIsProcessTrustedWithOptions((__bridge CFDictionaryRef)options);
    
    if (!trusted) {
        fprintf(stderr, "[CGO] ⚠️ Accessibility permission not granted\n");
        fflush(stderr);
        
        // Show alert to guide user
        NSAlert *alert = [[NSAlert alloc] init];
        [alert setMessageText:@"需要辅助功能权限"];
        [alert setInformativeText:@"Haoyun Music Player 需要辅助功能权限才能监听媒体键（F7/F8/F9）。\n\n请在系统偏好设置中授予权限：\n1. 打开\"系统偏好设置 > 安全性与隐私 > 隐私 > 辅助功能\"\n2. 点击锁图标解锁\n3. 勾选 Haoyun Music Player\n4. 重启应用"];
        [alert addButtonWithTitle:@"打开系统偏好设置"];
        [alert addButtonWithTitle:@"稍后"];
        
        NSInteger response = [alert runModal];
        
        if (response == NSAlertFirstButtonReturn) {
            // Open System Preferences directly to Accessibility panel
            [[NSWorkspace sharedWorkspace] openURL:[NSURL URLWithString:@"x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility"]];
        }
        
        return NO;
    }
    
    fprintf(stderr, "[CGO] ✅ Accessibility permission granted\n");
    fflush(stderr);
    return YES;
}
static void setupMediaKeys() {
    // Check accessibility permission first
    if (!checkAccessibilityPermission()) {
        fprintf(stderr, "[CGO] ⚠️ Skipping media key registration due to missing permission\n");
        fflush(stderr);
        return;
    }
    
    fprintf(stderr, "[CGO] 🔧 setupMediaKeys called\n");
    fflush(stderr);
    
    // Register for system-defined events (media keys)
    fprintf(stderr, "[CGO] 📡 Setting up NSEvent monitor...\n");
    fflush(stderr);
    
    eventMonitor = [NSEvent addGlobalMonitorForEventsMatchingMask:NSSystemDefinedMask handler:^(NSEvent *event) {
        fprintf(stderr, "[CGO] 📨 Received system event: type=%ld, subtype=%ld\n", (long)[event type], (long)[event subtype]);
        fflush(stderr);
        
        if ([event type] == NSSystemDefined && [event subtype] == 8) {
            int keyCode = (([event data1] & 0xFFFF0000) >> 16);
            int keyFlags = ([event data1] & 0x0000FFFF);
            int keyState = (((keyFlags & 0xFF00) >> 8)) == 0xA;
            
            fprintf(stderr, "[CGO] 🎹 Media key detected: keyCode=%d, keyState=%d\n", keyCode, keyState);
            fflush(stderr);
            
            if (keyState) {
                switch (keyCode) {
                    case NX_KEYTYPE_PLAY:
                        fprintf(stderr, "[CGO] ▶️ Play/Pause key pressed\n");
                        fflush(stderr);
                        goPlayPauseHandler();
                        break;
                    case NX_KEYTYPE_NEXT:
                        fprintf(stderr, "[CGO] ⏭️ Next Track key pressed\n");
                        fflush(stderr);
                        goNextHandler();
                        break;
                    case NX_KEYTYPE_PREVIOUS:
                        fprintf(stderr, "[CGO] ⏮️ Previous Track key pressed\n");
                        fflush(stderr);
                        goPreviousHandler();
                        break;
                }
            }
        }
    }];
    
    if (eventMonitor != nil) {
        fprintf(stderr, "[CGO] ✅ macOS system media keys registered successfully\n");
    } else {
        fprintf(stderr, "[CGO] ❌ Failed to register media keys monitor\n");
    }
    fflush(stderr);
}
static void cleanupMediaKeys() {
    if (eventMonitor != nil) {
        [NSEvent removeMonitor:eventMonitor];
        eventMonitor = nil;
    }
    NSLog(@"🔓 macOS system media keys cleaned up");
}
*/
import "C"

import (
"log"
"sync"
)

var (
mediaKeyCallbacks struct {
sync.RWMutex
playPause func()
		next      func()
		previous  func()
	}
	mediaKeysInitialized bool
)

// SetMediaKeyCallbacks sets media key callback functions
func SetMediaKeyCallbacks(playPause, next, previous func()) {
	mediaKeyCallbacks.Lock()
	defer mediaKeyCallbacks.Unlock()
	mediaKeyCallbacks.playPause = playPause
	mediaKeyCallbacks.next = next
	mediaKeyCallbacks.previous = previous
}

// platformRegisterMediaKeys registers system media keys on macOS
func (mks *MediaKeyService) platformRegisterMediaKeys() error {
	log.Println("🍎 Registering macOS system media keys...")
	
	SetMediaKeyCallbacks(
mks.handlePlayPause,
mks.handleNext,
mks.handlePrevious,
)
	
	if !mediaKeysInitialized {
		log.Println("🔧 Calling C.setupMediaKeys()...")
		C.setupMediaKeys()
		mediaKeysInitialized = true
		log.Println("✅ C.setupMediaKeys() returned successfully")
	} else {
		log.Println("⚠️ Media keys already initialized, skipping")
	}
	
	return nil
}

// platformUnregisterMediaKeys unregisters system media keys on macOS
func (mks *MediaKeyService) platformUnregisterMediaKeys() {
	log.Println("🍎 Unregistering macOS system media keys...")
	if mediaKeysInitialized {
		C.cleanupMediaKeys()
		mediaKeysInitialized = false
	}
	
	mediaKeyCallbacks.Lock()
	mediaKeyCallbacks.playPause = nil
	mediaKeyCallbacks.next = nil
	mediaKeyCallbacks.previous = nil
	mediaKeyCallbacks.Unlock()
}

//export goPlayPauseHandler
func goPlayPauseHandler() {
	log.Println("📨 CGO callback: goPlayPauseHandler called")
	mediaKeyCallbacks.RLock()
	defer mediaKeyCallbacks.RUnlock()
	if mediaKeyCallbacks.playPause != nil {
		log.Println("▶️⏸️  macOS media key: Play/Pause")
		mediaKeyCallbacks.playPause()
	} else {
		log.Println("⚠️ playPause callback is nil")
	}
}

//export goNextHandler
func goNextHandler() {
	log.Println("📨 CGO callback: goNextHandler called")
	mediaKeyCallbacks.RLock()
	defer mediaKeyCallbacks.RUnlock()
	if mediaKeyCallbacks.next != nil {
		log.Println("⏭️  macOS media key: Next Track")
		mediaKeyCallbacks.next()
	} else {
		log.Println("⚠️ next callback is nil")
	}
}

//export goPreviousHandler
func goPreviousHandler() {
	log.Println("📨 CGO callback: goPreviousHandler called")
	mediaKeyCallbacks.RLock()
	defer mediaKeyCallbacks.RUnlock()
	if mediaKeyCallbacks.previous != nil {
		log.Println("⏮️  macOS media key: Previous Track")
		mediaKeyCallbacks.previous()
	} else {
		log.Println("⚠️ previous callback is nil")
	}
}
