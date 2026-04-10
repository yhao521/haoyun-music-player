package backend

import (
	"log"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// MediaKeyService 处理系统媒体键事件
type MediaKeyService struct {
	app          *application.App
	musicService *MusicService
	mu           sync.Mutex
	isRegistered bool
}

// NewMediaKeyService 创建媒体键服务
func NewMediaKeyService() *MediaKeyService {
	return &MediaKeyService{
		isRegistered: false,
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
	err := mks.platformRegisterMediaKeys()
	if err != nil {
		log.Printf("⚠️ 注册媒体键失败:%v", err)
		return err
	}

	mks.isRegistered = true
	log.Println("✅ 系统媒体键注册成功")
	return nil
}

// UnregisterMediaKeys 注销媒体键
func (mks *MediaKeyService) UnregisterMediaKeys() {
	mks.mu.Lock()
	defer mks.mu.Unlock()

	if !mks.isRegistered {
		return
	}

	log.Println("🔓 注销系统媒体键...")
	mks.platformUnregisterMediaKeys()
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
