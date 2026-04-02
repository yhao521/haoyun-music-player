package main

import (
	"embed"
	_ "embed"
	"log"
	"runtime/debug"
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

	// 创建基本播放控制菜单项
	playPauseItem := application.NewMenuItem("播放")
	playPauseItem.OnClick(func(ctx *application.Context) {
		musicService.TogglePlayPause()
	})

	prevItem := application.NewMenuItem("上一曲")
	prevItem.OnClick(func(ctx *application.Context) {
		musicService.Previous()
	})

	nextItem := application.NewMenuItem("下一曲")
	nextItem.OnClick(func(ctx *application.Context) {
		musicService.Next()
	})

	showItem := application.NewMenuItem("显示主窗口")
	showItem.OnClick(func(ctx *application.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("❌ 显示窗口时发生 panic: %v", r)
				debug.PrintStack()
			}
		}()

		log.Println("=== 显示主窗口菜单项被点击 ===")

		if mainWindow == nil {
			log.Println("❌ mainWindow 为 nil")
			return
		}

		isvisible := mainWindow.IsVisible()
		log.Println("✓ IsVisible() = ", isvisible)
		log.Println("准备调用 Maximise()...")
		mainWindow.Maximise()
		log.Println("✓ Maximise() 完成")

		mainWindow.Focus()
		log.Println("✓ Focus() 完成")

		log.Println("=== 操作完成 ===")
	})

	// 创建浏览歌曲菜单项（带快捷键 Cmd+F）
	browseItem := application.NewMenuItem("浏览歌曲")
	browseItem.SetAccelerator("CmdOrCtrl+F")
	browseItem.OnClick(func(ctx *application.Context) {
		// TODO: 实现浏览歌曲功能
		log.Println("浏览歌曲")
	})

	// 创建播放模式子菜单
	playModeMenu := application.NewMenuFromItems(
		application.NewMenuItem("顺序播放"),
		application.NewMenuItem("循环播放"),
		application.NewMenuItem("随机播放"),
	)
	playModeItem := application.NewSubmenu("播放模式", playModeMenu)

	// 创建音乐库子菜单
	musicLibMenu := application.NewMenuFromItems(
		application.NewMenuItem("✓ music"),
		application.NewMenuItemSeparator(),
		application.NewMenuItem("刷新当前音乐库"),
		application.NewMenuItem("添加新音乐库"),
		application.NewMenuItem("重命名当前音乐库"),
	)
	musicLibItem := application.NewSubmenu("音乐库", musicLibMenu)

	// 创建下载音乐菜单项（带快捷键 Cmd+D）
	downloadItem := application.NewMenuItem("下载音乐")
	downloadItem.SetAccelerator("CmdOrCtrl+D")
	downloadItem.OnClick(func(ctx *application.Context) {
		// TODO: 实现下载音乐功能
		log.Println("下载音乐")
	})

	// 创建保持系统唤醒菜单项（带复选框）
	wakeItem := application.NewMenuItemCheckbox("保持系统唤醒", true)
	wakeItem.OnClick(func(ctx *application.Context) {
		// TODO: 实现保持唤醒功能
		log.Println("保持系统唤醒")
	})

	// 创建开机启动菜单项（带复选框）
	launchItem := application.NewMenuItemCheckbox("开机启动", true)
	launchItem.OnClick(func(ctx *application.Context) {
		// TODO: 实现开机启动功能
		log.Println("开机启动")
	})

	// 创建设置菜单项（带快捷键 Cmd+S）
	settingItem := application.NewMenuItem("设置")
	settingItem.SetAccelerator("CmdOrCtrl+S")
	settingItem.OnClick(func(ctx *application.Context) {
		// TODO: 实现设置功能
		log.Println("设置")
	})

	// 创建版本信息（禁用状态）
	versionItem := application.NewMenuItem("Version 0.5.0")
	versionItem.SetEnabled(false)

	// 创建退出菜单项
	quitItem := application.NewMenuItem("退出")
	quitItem.OnClick(func(ctx *application.Context) {
		app.Quit()
	})

	// 创建菜单
	menu := application.NewMenuFromItems(
		playPauseItem,
		prevItem,
		nextItem,
		application.NewMenuItemSeparator(),
		browseItem,
		playModeItem,
		musicLibItem,
		downloadItem,
		wakeItem,
		launchItem,
		settingItem,
		showItem,
		application.NewMenuItemSeparator(),
		versionItem,
		quitItem,
	)

	// 设置菜单
	tray.SetMenu(menu)

	// 交互事件
	// 注意：macOS 上单击托盘图标会自动显示菜单
	// 如果需要双击显示窗口，保留 OnDoubleClick
	tray.OnDoubleClick(func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("❌ 双击显示窗口时发生 panic: %v", r)
				debug.PrintStack()
			}
		}()

		log.Println("=== 托盘双击事件 ===")

		if mainWindow == nil {
			log.Println("❌ mainWindow 为 nil")
			return
		}
		isvisible := mainWindow.IsVisible()
		log.Println("✓ IsVisible() = ", isvisible)

		mainWindow.Maximise()
		log.Println("✓ Maximise() 完成")

		mainWindow.Focus()
		log.Println("✓ Focus() 完成")

		log.Println("=== 操作完成 ===")
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
	mainWindow.Minimise()
	log.Println("✓ Main window created (Minimise)")

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
