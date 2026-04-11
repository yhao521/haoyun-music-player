//go:build darwin

package backend

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework AppKit -framework Foundation
#import <AppKit/AppKit.h>
#import <Foundation/Foundation.h>

// Forward declarations of Go exported functions
extern void goPlayPauseHandler(void);
extern void goNextHandler(void);
extern void goPreviousHandler(void);

static id eventMonitor = nil;

void setupMediaKeys() {
    // Register for system-defined events (media keys)
    eventMonitor = [NSEvent addGlobalMonitorForEventsMatchingMask:NSSystemDefinedMask handler:^(NSEvent *event) {
        if ([event type] == NSSystemDefined && [event subtype] == 8) {
            int keyCode = (([event data1] & 0xFFFF0000) >> 16);
            int keyFlags = ([event data1] & 0x0000FFFF);
            int keyState = (((keyFlags & 0xFF00) >> 8)) == 0xA;
            
            if (keyState) {
                switch (keyCode) {
                    case NX_KEYTYPE_PLAY:
                        goPlayPauseHandler();
                        break;
                    case NX_KEYTYPE_NEXT:
                        goNextHandler();
                        break;
                    case NX_KEYTYPE_PREVIOUS:
                        goPreviousHandler();
                        break;
                }
            }
        }
    }];
    
    NSLog(@"✅ macOS system media keys registered successfully");
}

void cleanupMediaKeys() {
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
		C.setupMediaKeys()
		mediaKeysInitialized = true
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
	mediaKeyCallbacks.RLock()
	defer mediaKeyCallbacks.RUnlock()
	if mediaKeyCallbacks.playPause != nil {
		log.Println("▶️⏸️  macOS media key: Play/Pause")
		mediaKeyCallbacks.playPause()
	}
}

//export goNextHandler
func goNextHandler() {
	mediaKeyCallbacks.RLock()
	defer mediaKeyCallbacks.RUnlock()
	if mediaKeyCallbacks.next != nil {
		log.Println("⏭️  macOS media key: Next Track")
		mediaKeyCallbacks.next()
	}
}

//export goPreviousHandler
func goPreviousHandler() {
	mediaKeyCallbacks.RLock()
	defer mediaKeyCallbacks.RUnlock()
	if mediaKeyCallbacks.previous != nil {
		log.Println("⏮️  macOS media key: Previous Track")
		mediaKeyCallbacks.previous()
	}
}
