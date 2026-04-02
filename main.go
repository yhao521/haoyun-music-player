package main

import (
	"embed"
	_ "embed"
	"log"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed frontend/public/wails.png
var trayIcon []byte

// //go:embed frontend/public/tray-icon-dark.png
// var trayIconDark []byte

func init() {
	application.RegisterEvent[string]("time")
	application.RegisterEvent[string]("playbackStateChanged")
	application.RegisterEvent[map[string]interface{}]("playbackProgress")
	application.RegisterEvent[[]string]("playlistUpdated")
	application.RegisterEvent[string]("currentTrackChanged")
}

func main() {
	musicService := NewMusicService()

	app := application.New(application.Options{
		Name:        "Haoyun Music Player",
		Description: "A menu bar music player built with Wails 3 + Vue 3",
		Services: []application.Service{
			application.NewService(musicService),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	musicService.SetApp(app)

	// 声明窗口变量（先初始化为 nil）
	var mainWindow *application.WebviewWindow

	// 创建系统托盘（在窗口创建之前）
	tray := app.SystemTray.New()
	log.Println("✓ System tray initialized")

	// 设置托盘图标
	tray.SetIcon(trayIcon)
	tray.SetTooltip("Haoyun Music Player")

	// 创建菜单项
	playPauseItem := application.NewMenuItem("播放/暂停")
	playPauseItem.OnClick(func(ctx *application.Context) {
		musicService.TogglePlayPause()
	})

	prevItem := application.NewMenuItem("上一首")
	prevItem.OnClick(func(ctx *application.Context) {
		musicService.Previous()
	})

	nextItem := application.NewMenuItem("下一首")
	nextItem.OnClick(func(ctx *application.Context) {
		musicService.Next()
	})

	showItem := application.NewMenuItem("显示主窗口")
	showItem.OnClick(func(ctx *application.Context) {
		if mainWindow != nil {
			mainWindow.Show()
			mainWindow.Focus()
		}
	})

	quitItem := application.NewMenuItem("退出")
	quitItem.OnClick(func(ctx *application.Context) {
		app.Quit()
	})

	// 创建菜单
	menu := application.NewMenuFromItems(
		playPauseItem,
		application.NewMenuItemSeparator(),
		prevItem,
		nextItem,
		application.NewMenuItemSeparator(),
		showItem,
		application.NewMenuItemSeparator(),
		quitItem,
	)

	// 设置菜单
	tray.SetMenu(menu)

	// 交互事件
	// 注意：macOS 上单击托盘图标会自动显示菜单
	// 如果需要双击显示窗口，保留 OnDoubleClick
	tray.OnDoubleClick(func() {
		if mainWindow != nil {
			mainWindow.Show()
			mainWindow.Focus()
		}
	})

	log.Println("✓ System tray menu created")

	// 创建主窗口（默认隐藏，通过托盘菜单打开）
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
		// Visible 字段不存在，使用 Hide() 方法
	})

	// 初始隐藏窗口
	mainWindow.Hide()
	log.Println("✓ Main window created (hidden)")

	go func() {
		for {
			now := time.Now().Format(time.RFC1123)
			app.Event.Emit("time", now)
			time.Sleep(time.Second)
		}
	}()

	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}

	musicService.Shutdown()
}
