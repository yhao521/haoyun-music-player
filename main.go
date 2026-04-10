package main

import (
	"log"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func main() {
	// 初始化应用核心组件
	if err := initializeApp(); err != nil {
		log.Fatal(err)
	}

	// 检查依赖工具状态
	checkDependencies()

	// 创建主菜单
	menus, playPauseMenuItem, prevMenuItem, nextMenuItem := createMenu(app)
	app.Menu.Set(menus)

	// 设置播放控制菜单项的回调
	setupMainMenuCallbacks(playPauseMenuItem, prevMenuItem, nextMenuItem)

	// 创建系统托盘和菜单
	createTrayAndMenu()

	// 设置依赖管理器回调
	setupDependencyManagerCallback()

	// 创建所有窗口
	createAllWindows()

	// 注册事件处理器
	registerEventHandlers()

	// 初始化正在播放显示
	go func() {
		time.Sleep(500 * time.Millisecond)
		updateNowPlayingItem()

		// 初始化播放模式显示
		if currentMode, err := musicService.GetPlayMode(); err == nil {
			log.Printf("✓ 初始化托盘菜单播放模式：%s", currentMode)
			updatePlayModeMenuLabels(currentMode)
		}
	}()

	log.Println("✅ 应用启动完成")

	// 运行应用
	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}

	// 优雅关闭
	musicService.Shutdown()
}

// setupMainMenuCallbacks 设置主菜单播放控制的回调
func setupMainMenuCallbacks(playPauseMenuItem, prevMenuItem, nextMenuItem *application.MenuItem) {
	playPauseMenuItem.OnClick(func(ctx *application.Context) {
		playlist, _ := musicService.GetPlaylist()
		if len(playlist) == 0 {
			currentLib := musicService.GetCurrentLibrary()
			if currentLib != nil {
				if err := musicService.LoadCurrentLibrary(); err != nil {
					log.Printf("加载音乐库失败：%v", err)
				}
			}
		} else {
			musicService.TogglePlayPause()
		}
	})

	prevMenuItem.OnClick(func(ctx *application.Context) {
		if err := musicService.Previous(); err != nil {
			log.Printf("切换上一曲失败：%v", err)
		}
	})

	nextMenuItem.OnClick(func(ctx *application.Context) {
		if err := musicService.Next(); err != nil {
			log.Printf("切换下一曲失败：%v", err)
		}
	})
}

