// mediakeyservice_darwin.m - macOS Media Keys Implementation

#include <MediaPlayer/MediaPlayer.h>
#include <Foundation/Foundation.h>
#include <stdio.h>

// Forward declarations of Go exported functions
extern void goPlayPauseHandler(void);
extern void goNextHandler(void);
extern void goPreviousHandler(void);

static void handlePlayPauseCommand(MPRemoteCommandEvent *event) {
    goPlayPauseHandler();
}

static void handleNextCommand(MPRemoteCommandEvent *event) {
    goNextHandler();
}

static void handlePreviousCommand(MPRemoteCommandEvent *event) {
    goPreviousHandler();
}

void setupMediaKeys() {
    MPRemoteCommandCenter* center = [MPRemoteCommandCenter sharedCommandCenter];
    
    // Register play/pause command
    [center.playCommand addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent * _Nonnull event) {
        handlePlayPauseCommand(event);
        return MPRemoteCommandHandlerStatusSuccess;
    }];
    [center.playCommand setEnabled:YES];
    
    // Register pause command (shares same handler with play)
    [center.pauseCommand addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent * _Nonnull event) {
        handlePlayPauseCommand(event);
        return MPRemoteCommandHandlerStatusSuccess;
    }];
    [center.pauseCommand setEnabled:YES];
    
    // Register next track command
    [center.nextTrackCommand addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent * _Nonnull event) {
        handleNextCommand(event);
        return MPRemoteCommandHandlerStatusSuccess;
    }];
    [center.nextTrackCommand setEnabled:YES];
    
    // Register previous track command
    [center.previousTrackCommand addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent * _Nonnull event) {
        handlePreviousCommand(event);
        return MPRemoteCommandHandlerStatusSuccess;
    }];
    [center.previousTrackCommand setEnabled:YES];
    
    NSLog(@"✅ macOS system media keys registered successfully");
}

void cleanupMediaKeys() {
    MPRemoteCommandCenter* center = [MPRemoteCommandCenter sharedCommandCenter];
    
    [center.playCommand removeTarget:nil];
    [center.pauseCommand removeTarget:nil];
    [center.nextTrackCommand removeTarget:nil];
    [center.previousTrackCommand removeTarget:nil];
    
    NSLog(@"🔓 macOS system media keys cleaned up");
}
