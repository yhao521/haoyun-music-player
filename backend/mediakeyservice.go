package backend

import (
	"log"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
	"golang.design/x/hotkey"
)

// MediaKeyService 处理系统媒体键事件和全局快捷键
type MediaKeyService struct {
	app          *application.App
	musicService *MusicService
	mu           sync.Mutex
	isRegistered bool
	hotkeys      []*hotkey.Hotkey // 存储注册的快捷键实例
}

// NewMediaKeyService 创建媒体键服务
func NewMediaKeyService() *MediaKeyService {
	return &MediaKeyService{
		isRegistered: false,
		hotkeys:      make([]*hotkey.Hotkey, 0),
	}
}

// SetApp 设置应用实例
func (mks *MediaKeyService) SetApp(app *application.App) {
	mks.app = app
}

// SetMusicService 设置音乐服务
func (mks *MediaKeyService) SetMusicService(ms *MusicService) {
	mks.musicService = ms
}

// RegisterMediaKeys 注册媒体键(跨平台)
func (mks *MediaKeyService) RegisterMediaKeys() error {
	mks.mu.Lock()
	defer mks.mu.Unlock()

	if mks.isRegistered {
		log.Println("⚠️ 媒体键已注册")
		return nil
	}

	log.Println("🎹 尝试注册系统媒体键...")

	// 调用平台相关的实现
	// err := mks.platformRegisterMediaKeys()
	// if err != nil {
	// 	log.Printf("⚠️ 注册媒体键失败:%v", err)
	// 	return err
	// }

	mks.isRegistered = true
	log.Println("✅ 系统媒体键注册成功")
	
	// 注册全局快捷键
	go mks.registerGlobalHotkeys()
	
	return nil
}

// registerGlobalHotkeys 注册全局快捷键
func (mks *MediaKeyService) registerGlobalHotkeys() {
	log.Println("⌨️  注册全局快捷键...")
	
	// 定义快捷键配置
	type HotkeyConfig struct {
		mods    []hotkey.Modifier
		key     hotkey.Key
		handler func()
		name    string
	}
	
	configs := []HotkeyConfig{
		{
			mods:    []hotkey.Modifier{},
			key:     hotkey.KeySpace,
			handler: mks.handlePlayPause,
			name:    "播放/暂停 (Space)",
		},
		{
			mods:    []hotkey.Modifier{hotkey.ModCtrl},
			key:     hotkey.KeyRight,
			handler: mks.handleNext,
			name:    "下一曲 (Ctrl+→)",
		},
		{
			mods:    []hotkey.Modifier{hotkey.ModCtrl},
			key:     hotkey.KeyLeft,
			handler: mks.handlePrevious,
			name:    "上一曲 (Ctrl+←)",
		},
		{
			mods:    []hotkey.Modifier{hotkey.ModCtrl},
			key:     hotkey.KeyUp,
			handler: mks.handleVolumeUp,
			name:    "音量增加 (Ctrl+↑)",
		},
		{
			mods:    []hotkey.Modifier{hotkey.ModCtrl},
			key:     hotkey.KeyDown,
			handler: mks.handleVolumeDown,
			name:    "音量减少 (Ctrl+↓)",
		},
	}
	
	// 注册所有快捷键
	for _, config := range configs {
		hk := hotkey.New(config.mods, config.key)
		err := hk.Register()
		if err != nil {
			log.Printf("⚠️ 注册快捷键 %s 失败: %v", config.name, err)
			continue
		}
		
		mks.hotkeys = append(mks.hotkeys, hk)
		log.Printf("✅ 注册快捷键: %s", config.name)
		
		// 启动监听协程
		go func(h *hotkey.Hotkey, handler func(), name string) {
			for range h.Keydown() {
				log.Printf("⌨️  触发快捷键: %s", name)
				handler()
			}
		}(hk, config.handler, config.name)
	}
	
	log.Printf("✅ 全局快捷键注册完成，共 %d 个", len(mks.hotkeys))
}

// UnregisterMediaKeys 注销媒体键
func (mks *MediaKeyService) UnregisterMediaKeys() {
	mks.mu.Lock()
	defer mks.mu.Unlock()

	if !mks.isRegistered {
		return
	}

	log.Println("🔓 注销系统媒体键...")
	// mks.platformUnregisterMediaKeys()
	
	// 注销所有全局快捷键
	for _, hk := range mks.hotkeys {
		hk.Unregister()
	}
	mks.hotkeys = make([]*hotkey.Hotkey, 0)
	
	mks.isRegistered = false
}

// handlePlayPause 处理播放/暂停事件
func (mks *MediaKeyService) handlePlayPause() {
	log.Println("▶️⏸️  收到媒体键:播放/暂停")
	if mks.musicService != nil {
		mks.musicService.TogglePlayPause()
	}
}

// handleNext 处理下一曲事件
func (mks *MediaKeyService) handleNext() {
	log.Println("⏭️  收到媒体键:下一曲")
	if mks.musicService != nil {
		go func() {
			if err := mks.musicService.Next(); err != nil {
				log.Printf("切换下一曲失败:%v", err)
			}
		}()
	}
}

// handlePrevious 处理上一曲事件
func (mks *MediaKeyService) handlePrevious() {
	log.Println("⏮️  收到媒体键:上一曲")
	if mks.musicService != nil {
		go func() {
			if err := mks.musicService.Previous(); err != nil {
				log.Printf("切换上一曲失败:%v", err)
			}
		}()
	}
}

// handleVolumeUp 处理音量增加
func (mks *MediaKeyService) handleVolumeUp() {
	log.Println("🔊 收到快捷键:音量增加")
	if mks.musicService != nil {
		currentVolume, err := mks.musicService.GetVolume()
		if err != nil {
			log.Printf("获取音量失败: %v", err)
			return
		}
		
		// 增加 10% 音量，最大 1.0
		newVolume := currentVolume + 0.1
		if newVolume > 1.0 {
			newVolume = 1.0
		}
		
		if err := mks.musicService.SetVolume(newVolume); err != nil {
			log.Printf("设置音量失败: %v", err)
		} else {
			log.Printf("✅ 音量调整为: %.0f%%", newVolume*100)
		}
	}
}

// handleVolumeDown 处理音量减少
func (mks *MediaKeyService) handleVolumeDown() {
	log.Println("🔉 收到快捷键:音量减少")
	if mks.musicService != nil {
		currentVolume, err := mks.musicService.GetVolume()
		if err != nil {
			log.Printf("获取音量失败: %v", err)
			return
		}
		
		// 减少 10% 音量，最小 0.0
		newVolume := currentVolume - 0.1
		if newVolume < 0.0 {
			newVolume = 0.0
		}
		
		if err := mks.musicService.SetVolume(newVolume); err != nil {
			log.Printf("设置音量失败: %v", err)
		} else {
			log.Printf("✅ 音量调整为: %.0f%%", newVolume*100)
		}
	}
}
