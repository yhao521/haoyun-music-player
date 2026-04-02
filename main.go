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

	// 创建主窗口
	mainWindow := app.Window.NewWithOptions(application.WebviewWindowOptions{
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

	// 创建系统托盘
	createSystemTray(app, musicService, mainWindow)

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

// createSystemTray 创建系统托盘菜单
func createSystemTray(app *application.App, musicService *MusicService, mainWindow *application.WebviewWindow) {
	tray := app.SystemTray.New()
	
	// 创建托盘菜单
	menu := application.NewMenu()
	
	// 添加菜单项
	playPauseItem := application.NewMenuItem("播放/暂停")
	playPauseItem.OnClick(func(ctx *application.Context) {
		musicService.TogglePlayPause()
	})
	menu.Add(playPauseItem)
	
	menu.AddSeparator()
	
	prevItem := application.NewMenuItem("上一首")
	prevItem.OnClick(func(ctx *application.Context) {
		musicService.Previous()
	})
	menu.Add(prevItem)
	
	nextItem := application.NewMenuItem("下一首")
	nextItem.OnClick(func(ctx *application.Context) {
		musicService.Next()
	})
	menu.Add(nextItem)
	
	menu.AddSeparator()
	
	showItem := application.NewMenuItem("显示主窗口")
	showItem.OnClick(func(ctx *application.Context) {
		mainWindow.Show()
		mainWindow.Focus()
	})
	menu.Add(showItem)
	
	menu.AddSeparator()
	
	quitItem := application.NewMenuItem("退出")
	quitItem.OnClick(func(ctx *application.Context) {
		app.Quit()
	})
	menu.Add(quitItem)
	
	// 设置菜单
	tray.SetMenu(menu)
	
	// 设置工具提示
	tray.SetTooltip("Haoyun Music Player")
	
	// 单击托盘图标时切换播放/暂停
	tray.OnClick(func() {
		musicService.TogglePlayPause()
	})
	
	// 双击托盘图标时显示窗口
	tray.OnDoubleClick(func() {
		mainWindow.Show()
		mainWindow.Focus()
	})
}
