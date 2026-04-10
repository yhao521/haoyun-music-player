package main

import (
	"log"
	"time"

	"github.com/yhao521/wailsMusicPlay/backend"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// registerEventHandlers 注册所有事件处理器
func registerEventHandlers() {
	registerPlaybackEvents()
	registerWindowEvents()
	registerSettingEvents()
	registerLibraryEvents()
}

// registerPlaybackEvents 注册播放控制相关事件
func registerPlaybackEvents() {
	// 监听播放模式设置事件
	app.Event.On("setPlayMode", func(event *application.CustomEvent) {
		if mode, ok := event.Data.(string); ok {
			musicService.SetPlayMode(mode)
			log.Printf("✓ 切换到%s播放", mode)
		}
	})

	// 监听主菜单播放控制事件
	app.Event.On("menu:playPause", func(event *application.CustomEvent) {
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

	app.Event.On("menu:prevTrack", func(event *application.CustomEvent) {
		if err := musicService.Previous(); err != nil {
			log.Printf("切换上一曲失败：%v", err)
		}
	})

	app.Event.On("menu:nextTrack", func(event *application.CustomEvent) {
		if err := musicService.Next(); err != nil {
			log.Printf("切换下一曲失败：%v", err)
		}
	})
}

// registerWindowEvents 注册窗口管理事件
func registerWindowEvents() {
	// 监听窗口打开事件（从主菜单触发）
	app.Event.On("openWindow", func(event *application.CustomEvent) {
		if data, ok := event.Data.(map[string]interface{}); ok {
			if windowType, exists := data["type"].(string); exists {
				switch windowType {
				case "main":
					if mainWindow != nil {
						if mainWindow.IsVisible() {
							mainWindow.Hide()
						} else {
							mainWindow.Show()
							mainWindow.Focus()
						}
					}
				case "browse":
					if browseWindow != nil {
						if browseWindow.IsVisible() {
							browseWindow.Hide()
						} else {
							browseWindow.Show()
							browseWindow.Focus()
						}
					}
				case "favorites":
					if favoritesWindow != nil {
						if favoritesWindow.IsVisible() {
							favoritesWindow.Hide()
						} else {
							favoritesWindow.Show()
							favoritesWindow.Focus()
						}
					}
				case "settings":
					if settingsWindow != nil {
						if settingsWindow.IsVisible() {
							settingsWindow.Hide()
						} else {
							settingsWindow.Show()
							settingsWindow.Focus()
						}
					}
				}
			}
		}
	})
}

// registerSettingEvents 注册设置相关事件
func registerSettingEvents() {
	// 监听语言切换事件
	app.Event.On("changeLanguage", func(event *application.CustomEvent) {
		if locale, ok := event.Data.(string); ok {
			// 保存语言设置到配置文件
			if err := configManager.SetLanguage(locale); err != nil {
				log.Printf("保存语言设置失败: %v", err)
				return
			}

			log.Printf("✓ 语言已切换为: %s (已保存到配置文件)", locale)

			// 重建托盘菜单以应用新语言
			rebuildTrayMenu()

			// 通知前端需要重启才能完全生效
			app.Event.Emit("languageChanged", map[string]interface{}{
				"locale":      locale,
				"needRestart": true,
				"message":     "语言已切换，部分界面需要重启应用后才能完全生效",
			})
		}
	})

	// 监听其他设置更新事件
	app.Event.On("updateSetting", func(event *application.CustomEvent) {
		if data, ok := event.Data.(map[string]interface{}); ok {
			for key, value := range data {
				switch key {
				case "theme":
					if theme, ok := value.(string); ok {
						configManager.SetTheme(theme)
					}
				case "autoLaunch":
					if enabled, ok := value.(bool); ok {
						configManager.SetAutoLaunch(enabled)
					}
				case "keepAwake":
					if enabled, ok := value.(bool); ok {
						configManager.SetKeepAwake(enabled)
					}
				case "defaultVolume":
					if volume, ok := value.(float64); ok {
						configManager.SetDefaultVolume(int(volume))
					}
				case "showLyrics":
					if show, ok := value.(bool); ok {
						configManager.SetShowLyrics(show)
					}
				case "enableMediaKeys":
					if enabled, ok := value.(bool); ok {
						configManager.SetEnableMediaKeys(enabled)
					}
				case "defaultPlayMode":
					if mode, ok := value.(string); ok {
						configManager.SetDefaultPlayMode(mode)
					}
				}
			}
			log.Println("✓ 设置已保存")
		}
	})

	// 监听获取设置事件
	app.Event.On("getSettings", func(event *application.CustomEvent) {
		cfg := configManager.Get()

		app.Event.Emit("getSettings:response", map[string]interface{}{
			"language":        cfg.Language,
			"theme":           cfg.Theme,
			"autoLaunch":      cfg.AutoLaunch,
			"keepAwake":       cfg.KeepAwake,
			"defaultVolume":   cfg.DefaultVolume,
			"showLyrics":      cfg.ShowLyrics,
			"enableMediaKeys": cfg.EnableMediaKeys,
			"defaultPlayMode": cfg.DefaultPlayMode,
		})

		log.Println("📤 已发送配置到前端")
	})

	// 监听应用重启请求
	app.Event.On("restartApp", func(event *application.CustomEvent) {
		log.Println("🔄 收到重启请求...")

		if err := configManager.Save(); err != nil {
			log.Printf("⚠️ 保存配置失败: %v", err)
		}

		musicService.Shutdown()
		log.Println("正在退出应用...")
		app.Quit()
	})
}

// registerLibraryEvents 注册音乐库相关事件
func registerLibraryEvents() {
	// 监听压缩音乐库请求
	app.Event.On("compactLibraries", func(event *application.CustomEvent) {
		log.Println("🗜️ 收到压缩音乐库请求...")

		count, err := musicService.CompactLibraries()
		if err != nil {
			log.Printf("⚠️ 压缩音乐库失败: %v", err)
			app.Event.Emit("compactLibraries:response", map[string]interface{}{
				"error": err.Error(),
			})
			return
		}

		log.Printf("✓ 压缩完成：共处理 %d 个音乐库", count)
		app.Event.Emit("compactLibraries:response", map[string]interface{}{
			"count": count,
		})
	})
}

// setupDependencyManagerCallback 设置依赖管理器回调
func setupDependencyManagerCallback() {
	depManager.SetCallback(func(toolName string, status backend.ToolStatus, message string) {
		log.Printf("🔧 工具状态变化: %s - %d (%s)", toolName, status, message)

		app.Event.Emit("dependencyStatusChanged", map[string]interface{}{
			"tool":    toolName,
			"status":  status,
			"message": message,
		})

		if status == backend.ToolInstalled || status == backend.ToolInstallFailed {
			go func() {
				time.Sleep(500 * time.Millisecond)
				rebuildTrayMenu()
			}()
		}
	})
}
