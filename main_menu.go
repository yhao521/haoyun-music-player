package main

import (
	"fmt"
	"log"
	"runtime"

	"github.com/yhao521/haoyun-music-player/backend/pkg/file"
	"github.com/yhao521/haoyun-music-player/backend/pkg/i18n"
	"github.com/yhao521/haoyun-music-player/backend/pkg/utils"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// createMenu 创建应用程序主菜单
func createMenu(app *application.App) (*application.Menu, *application.MenuItem, *application.MenuItem, *application.MenuItem) {
	translator := i18n.GetTranslator()
	t := func(key string) string {
		return translator.T(key)
	}

	menu := app.NewMenu()

	if runtime.GOOS == "darwin" {
		menu.AddRole(application.AppMenu)
	}

	// File menu
	fileMenu := menu.AddSubmenu(t("menu.file"))
	fileMenu.Add(t("menu.openRuntimeDir")).
		SetAccelerator("Ctrl+O").OnClick(func(ctx *application.Context) {
		OpenDir()
	})
	fileMenu.Add(t("menu.new")).SetAccelerator("Ctrl+N")
	fileMenu.Add(t("menu.open")).SetAccelerator("Ctrl+O")
	fileMenu.Add(t("menu.save")).SetAccelerator("Ctrl+S")
	fileMenu.AddSeparator()
	fileMenu.Add(t("menu.quit")).OnClick(func(ctx *application.Context) {
		app.Quit()
	})

	// Music menu
	musicMenu := menu.AddSubmenu(t("menu.music"))
	nowPlayingMenuItem := musicMenu.Add(t("status.notPlaying"))
	nowPlayingMenuItem.SetEnabled(false)
	musicMenu.AddSeparator()

	// 播放控制
	menuPlayPauseItem := musicMenu.Add(t("menu.playPause"))
	menuPlayPauseItem.SetAccelerator("CmdOrCtrl+Space")
	menuPlayPauseItem.OnClick(func(ctx *application.Context) {
		log.Println("主菜单: 播放/暂停")
		app.Event.Emit("menu:playPause", nil)
	})

	menuPrevItem := musicMenu.Add(t("menu.previousTrack"))
	menuPrevItem.SetAccelerator("CmdOrCtrl+Shift+[")
	menuPrevItem.OnClick(func(ctx *application.Context) {
		log.Println("主菜单: 上一曲")
		app.Event.Emit("menu:prevTrack", nil)
	})

	menuNextItem := musicMenu.Add(t("menu.nextTrack"))
	menuNextItem.SetAccelerator("CmdOrCtrl+Shift+]")
	menuNextItem.OnClick(func(ctx *application.Context) {
		log.Println("主菜单: 下一曲")
		app.Event.Emit("menu:nextTrack", nil)
	})

	musicMenu.AddSeparator()

	// 窗口管理
	menuBrowseItem := musicMenu.Add(t("menu.browseSongs"))
	menuBrowseItem.SetAccelerator("CmdOrCtrl+Shift+F")

	menuFavoriteItem := musicMenu.Add(t("menu.favoriteSongs"))
	menuFavoriteItem.SetAccelerator("CmdOrCtrl+Shift+H")

	menuMainWindowItem := musicMenu.Add(t("menu.showMainWindow"))
	menuMainWindowItem.OnClick(func(ctx *application.Context) {
		log.Println(t("menu.showMainWindow"))
		app.Event.Emit("openWindow", map[string]interface{}{"type": "main"})
	})

	menuSettingItem := musicMenu.Add(t("menu.settings"))
	menuSettingItem.SetAccelerator("CmdOrCtrl+Shift+S")
	menuSettingItem.OnClick(func(ctx *application.Context) {
		log.Println(t("menu.settings"))
		app.Event.Emit("openWindow", map[string]interface{}{"type": "settings"})
	})

	musicMenu.AddSeparator()

	// 播放模式子菜单
	playModeSubMenu := musicMenu.AddSubmenu(t("menu.playMode"))
	menuPlayModeOrder := playModeSubMenu.Add("  " + t("playMode.order"))
	menuPlayModeOrder.OnClick(func(ctx *application.Context) {
		log.Println("切换到顺序播放")
		app.Event.Emit("setPlayMode", "order")
	})

	menuPlayModeLoop := playModeSubMenu.Add("✓ " + t("playMode.loop"))
	menuPlayModeLoop.OnClick(func(ctx *application.Context) {
		log.Println("切换到循环播放")
		app.Event.Emit("setPlayMode", "loop")
	})

	menuPlayModeRandom := playModeSubMenu.Add("  " + t("playMode.random"))
	menuPlayModeRandom.OnClick(func(ctx *application.Context) {
		log.Println("切换到随机播放")
		app.Event.Emit("setPlayMode", "random")
	})

	menuPlayModeSingle := playModeSubMenu.Add("  " + t("playMode.single"))
	menuPlayModeSingle.OnClick(func(ctx *application.Context) {
		log.Println("切换到单曲循环")
		app.Event.Emit("setPlayMode", "single")
	})

	// 音乐库子菜单
	musicLibSubMenu := musicMenu.AddSubmenu(t("menu.musicLibrary"))
	musicLibSubMenu.Add(t("library.refreshCurrent")).SetAccelerator("CmdOrCtrl+Shift+R")
	musicLibSubMenu.Add(t("library.addNew"))

	musicMenu.AddSeparator()

	// 其他功能
	menuDownloadItem := musicMenu.Add(t("menu.downloadMusic"))
	menuDownloadItem.SetAccelerator("CmdOrCtrl+Shift+D")
	menuDownloadItem.SetEnabled(false)

	_ = musicMenu.AddCheckbox(t("menu.keepAwake"), true)
	_ = musicMenu.AddCheckbox(t("menu.autoLaunch"), true)

	menuVersionItem := musicMenu.Add(fmt.Sprintf(t("menu.version"), AppVersion))
	menuVersionItem.SetEnabled(false)

	// Development menu
	devMenu := menu.AddSubmenu(t("menu.development"))
	devMenu.Add(t("dev.reloadApp")).OnClick(func(ctx *application.Context) {
		window := app.Window.Current()
		if window != nil {
			window.Reload()
		}
	})
	devMenu.Add(t("dev.openDevTools")).OnClick(func(ctx *application.Context) {
		window := app.Window.Current()
		if window != nil {
			window.OpenDevTools()
		}
	})
	devMenu.Add(t("dev.showEnvironment"))

	// Edit menu
	editMenu := menu.AddSubmenu(t("menu.edit"))
	editMenu.Add(t("menu.undo")).SetAccelerator("Ctrl+Z")
	editMenu.Add(t("menu.redo")).SetAccelerator("Ctrl+Y")
	editMenu.AddSeparator()
	editMenu.Add(t("menu.cut")).SetAccelerator("Ctrl+X")
	editMenu.Add(t("menu.copy")).SetAccelerator("Ctrl+C")
	editMenu.Add(t("menu.paste")).SetAccelerator("Ctrl+V")

	// Playback menu
	playbackMenu := menu.AddSubmenu(t("menu.playback"))
	playPauseMenuItem := playbackMenu.Add(t("menu.playPause"))
	playPauseMenuItem.SetAccelerator("Space")

	prevMenuItem := playbackMenu.Add(t("menu.previousTrack"))
	prevMenuItem.SetAccelerator("CmdOrCtrl+[")

	nextMenuItem := playbackMenu.Add(t("menu.nextTrack"))
	nextMenuItem.SetAccelerator("CmdOrCtrl+]")

	// View menu
	viewMenu := menu.AddSubmenu(t("menu.view"))
	darkMode := viewMenu.AddCheckbox(t("view.darkMode"), false)
	darkMode.OnClick(func(ctx *application.Context) {
		isChecked := darkMode.Checked()
		app.Logger.Info("Dark mode", "enabled", isChecked)
	})
	viewMenu.AddSeparator()
	_ = viewMenu.AddRadio(t("view.listView"), true)
	_ = viewMenu.AddRadio(t("view.gridView"), false)
	_ = viewMenu.AddRadio(t("view.detailView"), false)

	// Help menu
	helpMenu := menu.AddSubmenu(t("menu.help"))
	helpMenu.Add(t("help.documentation"))
	helpMenu.Add(t("help.about"))

	return menu, playPauseMenuItem, prevMenuItem, nextMenuItem
}

// OpenDir 打开应用运行目录
func OpenDir() {
	appPath := file.GetAppPath()
	if runtime.GOOS == "darwin" {
		utils.OpenMac(appPath)
	} else {
		utils.OpenWin(appPath)
	}
}
