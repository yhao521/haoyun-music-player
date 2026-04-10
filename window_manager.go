package main

import (
	"log"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// 其他窗口变量
var (
	browseWindow    *application.WebviewWindow
	favoritesWindow *application.WebviewWindow
	settingsWindow  *application.WebviewWindow
)

// createAllWindows 创建所有应用窗口
func createAllWindows() {
	createMainWindow()
	createBrowseWindow()
	createFavoritesWindow()
	createSettingsWindow()
	registerWindowCloseHooks()
}

// createMainWindow 创建主窗口（默认隐藏）
func createMainWindow() {
	mainWindow = app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "Haoyun Music Player",
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
		Width:            400,
		Height:           600,
	})
	mainWindow.Hide()
	log.Println("✓ Main window created and hidden")
}

// createBrowseWindow 创建浏览歌曲窗口
func createBrowseWindow() {
	browseWindow = app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "浏览歌曲 - Haoyun Music Player",
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "#/browse",
		Width:            900,
		Height:           700,
	})
	browseWindow.Hide()
	log.Println("✓ Browse window created and hidden")
}

// createFavoritesWindow 创建喜爱音乐窗口
func createFavoritesWindow() {
	favoritesWindow = app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "喜爱音乐 - Haoyun Music Player",
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "#/favorites",
		Width:            900,
		Height:           700,
	})
	favoritesWindow.Hide()
	log.Println("✓ Favorites window created and hidden")
}

// createSettingsWindow 创建设置窗口
func createSettingsWindow() {
	settingsWindow = app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "设置 - Haoyun Music Player",
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "#/settings",
		Width:            600,
		Height:           500,
	})
	settingsWindow.Hide()
	log.Println("✓ Settings window created and hidden")
}

// hasOtherVisibleWindows 检查是否还有其他可见窗口
func hasOtherVisibleWindows(currentWindow string) bool {
	switch currentWindow {
	case "main":
		return (browseWindow != nil && browseWindow.IsVisible()) ||
			(favoritesWindow != nil && favoritesWindow.IsVisible()) ||
			(settingsWindow != nil && settingsWindow.IsVisible())
	case "browse":
		return (mainWindow != nil && mainWindow.IsVisible()) ||
			(favoritesWindow != nil && favoritesWindow.IsVisible()) ||
			(settingsWindow != nil && settingsWindow.IsVisible())
	case "favorites":
		return (mainWindow != nil && mainWindow.IsVisible()) ||
			(browseWindow != nil && browseWindow.IsVisible()) ||
			(settingsWindow != nil && settingsWindow.IsVisible())
	case "settings":
		return (mainWindow != nil && mainWindow.IsVisible()) ||
			(browseWindow != nil && browseWindow.IsVisible()) ||
			(favoritesWindow != nil && favoritesWindow.IsVisible())
	default:
		return false
	}
}

// registerWindowCloseHooks 注册所有窗口的关闭拦截钩子
func registerWindowCloseHooks() {
	// 拦截主窗口关闭事件
	log.Println("🔧 正在为主窗口注册关闭拦截钩子...")
	mainWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		log.Println("⚠️ [主窗口] 关闭事件触发")

		if hasOtherVisibleWindows("main") {
			log.Println("ℹ️ [主窗口] 检测到其他可见窗口，但仍执行隐藏操作")
		} else {
			log.Println("ℹ️ [主窗口] 这是最后一个可见窗口")
		}

		mainWindow.Hide()
		e.Cancel()
		log.Println("✅ [主窗口] 已隐藏并取消关闭")
	})
	log.Println("✅ 主窗口关闭拦截钩子注册成功")

	// 拦截浏览窗口关闭事件
	log.Println("🔧 正在为浏览窗口注册关闭拦截钩子...")
	browseWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		log.Println("⚠️ [浏览窗口] 关闭事件触发")

		if hasOtherVisibleWindows("browse") {
			log.Println("ℹ️ [浏览窗口] 检测到其他可见窗口，但仍执行隐藏操作")
		} else {
			log.Println("ℹ️ [浏览窗口] 这是最后一个可见窗口")
		}

		browseWindow.Hide()
		e.Cancel()
		log.Println("✅ [浏览窗口] 已隐藏并取消关闭")
	})
	log.Println("✅ 浏览窗口关闭拦截钩子注册成功")

	// 拦截喜爱音乐窗口关闭事件
	log.Println("🔧 正在为喜爱音乐窗口注册关闭拦截钩子...")
	favoritesWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		log.Println("⚠️ [喜爱音乐窗口] 关闭事件触发")

		if hasOtherVisibleWindows("favorites") {
			log.Println("ℹ️ [喜爱音乐窗口] 检测到其他可见窗口，但仍执行隐藏操作")
		} else {
			log.Println("ℹ️ [喜爱音乐窗口] 这是最后一个可见窗口")
		}

		favoritesWindow.Hide()
		e.Cancel()
		log.Println("✅ [喜爱音乐窗口] 已隐藏并取消关闭")
	})
	log.Println("✅ 喜爱音乐窗口关闭拦截钩子注册成功")

	// 拦截设置窗口关闭事件
	log.Println("🔧 正在为设置窗口注册关闭拦截钩子...")
	settingsWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		log.Println("⚠️ [设置窗口] 关闭事件触发")

		if hasOtherVisibleWindows("settings") {
			log.Println("ℹ️ [设置窗口] 检测到其他可见窗口，但仍执行隐藏操作")
		} else {
			log.Println("ℹ️ [设置窗口] 这是最后一个可见窗口")
		}

		settingsWindow.Hide()
		e.Cancel()
		log.Println("✅ [设置窗口] 已隐藏并取消关闭")
	})
	log.Println("✅ 设置窗口关闭拦截钩子注册成功")
}

// toggleWindowVisibility 切换窗口显示/隐藏状态
func toggleWindowVisibility(window *application.WebviewWindow, windowName string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("❌ 切换 %s 窗口时发生 panic: %v", windowName, r)
		}
	}()

	if window == nil {
		log.Printf("❌ %s 窗口为 nil", windowName)
		return
	}

	isVisible := window.IsVisible()
	log.Printf("✓ %s IsVisible() = %v", windowName, isVisible)

	if isVisible {
		log.Println("准备调用 Hide()...")
		window.Hide()
	} else {
		log.Println("准备调用 Show()...")
		window.Show()
		log.Println("准备调用 Focus()...")
		window.Focus()
		log.Println("✓ Focus() 完成")
	}

	log.Printf("=== %s 窗口操作完成 ===", windowName)
}

// sendTestMessageToWindow 向指定窗口发送测试消息
func sendTestMessageToWindow(window *application.WebviewWindow, url string) {
	go func() {
		if window != nil && window.IsVisible() {
			time.Sleep(100 * time.Millisecond)
			app.Event.Emit("windowUrl", map[string]interface{}{
				"type":       "test_message",
				"message":    "后端定时测试消息",
				"serverTime": time.Now().Format(time.RFC1123),
				"url":        url,
			})
			log.Printf("📤 [测试消息] 已发送到 %s", url)
		}
	}()
}
