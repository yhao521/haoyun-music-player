package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/yhao521/wailsMusicPlay/backend"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// 托盘菜单相关变量
var (
	tray           *application.SystemTray
	menu           *application.Menu
	playPauseItem  *application.MenuItem
	prevItem       *application.MenuItem
	nextItem       *application.MenuItem
	mainWindowItem *application.MenuItem
	browseItem     *application.MenuItem
	favoriteItem   *application.MenuItem
	playModeItem   *application.MenuItem
	musicLibItem   *application.MenuItem
	toolsMenuItem  *application.MenuItem
	nowPlayingItem *application.MenuItem
	downloadItem   *application.MenuItem
	wakeItem       *application.MenuItem
	launchItem     *application.MenuItem
	settingItem    *application.MenuItem
	versionItem    *application.MenuItem
	quitItem       *application.MenuItem

	// 播放模式子菜单项
	playModeOrder  *application.MenuItem
	playModeLoop   *application.MenuItem
	playModeRandom *application.MenuItem
	playModeSingle *application.MenuItem

	// 音乐库菜单
	musicLibMenu *application.Menu
)

// createTrayAndMenu 创建系统托盘和菜单
func createTrayAndMenu() {
	createSystemTray()
	createTrayMenuItems()
	buildInitialTrayMenu()
	setupTrayEventListeners()
}

// createSystemTray 创建系统托盘
func createSystemTray() {
	tray = app.SystemTray.New()
	log.Println("✓ System tray initialized")

	tray.SetIcon(trayIcon)
	tray.SetTooltip("Haoyun Music Player")
}

// createTrayMenuItems 创建所有托盘菜单项
func createTrayMenuItems() {
	// 正在播放菜单项
	nowPlayingItem = application.NewMenuItem(t("status.notPlaying"))
	nowPlayingItem.SetEnabled(false)

	// 播放控制菜单项
	playPauseItem = application.NewMenuItem(t("menu.playPause"))
	playPauseItem.SetAccelerator("Space")
	playPauseItem.OnClick(func(ctx *application.Context) {
		handlePlayPauseClick()
	})

	prevItem = application.NewMenuItem(t("menu.previousTrack"))
	prevItem.SetAccelerator("CmdOrCtrl+[")
	prevItem.OnClick(func(ctx *application.Context) {
		if err := musicService.Previous(); err != nil {
			log.Printf("切换上一曲失败：%v", err)
		}
	})

	nextItem = application.NewMenuItem(t("menu.nextTrack"))
	nextItem.SetAccelerator("CmdOrCtrl+]")
	nextItem.OnClick(func(ctx *application.Context) {
		if err := musicService.Next(); err != nil {
			log.Printf("切换下一曲失败：%v", err)
		}
	})

	// 窗口管理菜单项
	mainWindowItem = application.NewMenuItem(t("menu.showMainWindow"))
	mainWindowItem.OnClick(func(ctx *application.Context) {
		toggleWindowVisibility(mainWindow, "主窗口")
		sendTestMessageToWindow(mainWindow, "#/main")
	})

	browseItem = application.NewMenuItem(t("menu.browseSongs"))
	browseItem.SetAccelerator("CmdOrCtrl+F")
	browseItem.OnClick(func(ctx *application.Context) {
		log.Println("浏览歌曲")
		toggleWindowVisibility(browseWindow, "浏览窗口")
		sendTestMessageToWindow(browseWindow, "#/browse")
	})

	favoriteItem = application.NewMenuItem(t("menu.favoriteSongs"))
	favoriteItem.SetAccelerator("CmdOrCtrl+H")
	favoriteItem.OnClick(func(ctx *application.Context) {
		toggleWindowVisibility(favoritesWindow, "喜爱音乐窗口")
	})

	settingItem = application.NewMenuItem(t("menu.settings"))
	settingItem.SetAccelerator("CmdOrCtrl+S")
	settingItem.OnClick(func(ctx *application.Context) {
		toggleWindowVisibility(settingsWindow, "设置窗口")
	})

	// 其他功能菜单项
	downloadItem = application.NewMenuItem(t("menu.downloadMusic"))
	downloadItem.SetAccelerator("CmdOrCtrl+D")
	downloadItem.SetEnabled(false)

	wakeItem = application.NewMenuItemCheckbox(t("menu.keepAwake"), true)
	wakeItem.SetEnabled(false)

	launchItem = application.NewMenuItemCheckbox(t("menu.autoLaunch"), true)
	launchItem.SetEnabled(false)

	versionLabel := fmt.Sprintf(t("menu.version"), AppVersion)
	versionItem = application.NewMenuItem(versionLabel)
	versionItem.SetEnabled(false)

	quitItem = application.NewMenuItem(t("menu.quit"))
	quitItem.OnClick(func(ctx *application.Context) {
		app.Quit()
	})

	// 创建播放模式子菜单
	createPlayModeMenu()
}

// createPlayModeMenu 创建播放模式子菜单
func createPlayModeMenu() {
	playModeOrder = application.NewMenuItem("  " + t("playMode.order"))
	playModeOrder.OnClick(func(ctx *application.Context) {
		if err := musicService.SetPlayMode("order"); err != nil {
			log.Printf("切换播放模式失败: %v", err)
		}
	})

	playModeLoop = application.NewMenuItem("✓ " + t("playMode.loop"))
	playModeLoop.OnClick(func(ctx *application.Context) {
		if err := musicService.SetPlayMode("loop"); err != nil {
			log.Printf("切换播放模式失败: %v", err)
		}
	})

	playModeRandom = application.NewMenuItem("  " + t("playMode.random"))
	playModeRandom.OnClick(func(ctx *application.Context) {
		if err := musicService.SetPlayMode("random"); err != nil {
			log.Printf("切换播放模式失败: %v", err)
		}
	})

	playModeSingle = application.NewMenuItem("  " + t("playMode.single"))
	playModeSingle.OnClick(func(ctx *application.Context) {
		if err := musicService.SetPlayMode("single"); err != nil {
			log.Printf("切换播放模式失败: %v", err)
		}
	})

	playModeMenu := application.NewMenuFromItems(
		playModeOrder,
		playModeLoop,
		playModeRandom,
		playModeSingle,
	)
	playModeItem = application.NewSubmenu(t("menu.playMode"), playModeMenu)
}

// buildInitialTrayMenu 构建初始托盘菜单
func buildInitialTrayMenu() {
	// 构建音乐库菜单
	buildMusicLibMenu()

	// 构建工具菜单
	buildToolsMenu()

	// 组装完整菜单
	menu = application.NewMenuFromItems(
		nowPlayingItem,
		application.NewMenuItemSeparator(),
		playPauseItem,
		prevItem,
		nextItem,
		application.NewMenuItemSeparator(),
		browseItem,
		favoriteItem,
		playModeItem,
		musicLibItem,
		toolsMenuItem,
		downloadItem,
		wakeItem,
		launchItem,
		settingItem,
		mainWindowItem,
		application.NewMenuItemSeparator(),
		versionItem,
		quitItem,
	)

	tray.SetMenu(menu)
	log.Println("✓ System tray menu created")
}

// handlePlayPauseClick 处理播放/暂停点击
func handlePlayPauseClick() {
	playlist, _ := musicService.GetPlaylist()
	log.Println("GetPlaylist", len(playlist))

	if len(playlist) == 0 {
		currentLib := musicService.GetCurrentLibrary()
		if currentLib == nil {
			log.Println("当前没有音乐库，请先添加音乐库")
			return
		}

		if err := musicService.LoadCurrentLibrary(); err != nil {
			log.Printf("加载音乐库失败：%v", err)
			return
		}

		log.Printf("✓ 已从音乐库 %s 加载并播放", currentLib.Name)
	} else {
		musicService.TogglePlayPause()
	}
}

// updateNowPlayingItem 更新正在播放菜单项
func updateNowPlayingItem() {
	if musicService == nil {
		log.Println("❌ musicService 为 nil")
		return
	}

	trackName, err := musicService.GetCurrentTrackName()
	if err != nil {
		log.Printf("⚠️ 获取歌曲名称失败：%v", err)
		nowPlayingItem.SetLabel(t("status.notPlaying"))
		nowPlayingItem.SetEnabled(false)
		return
	}

	log.Printf("✓ 更新正在播放：%s", trackName)

	displayName := strings.TrimSuffix(trackName, filepath.Ext(trackName))

	runes := []rune(displayName)
	if len(runes) > 30 {
		displayName = string(runes[:27]) + "..."
	}

	newLabel := fmt.Sprintf("🎵 %s", displayName)
	if newLabel == "" || newLabel == "🎵 " {
		newLabel = t("status.notPlaying")
	}

	nowPlayingItem.SetLabel(newLabel)
	nowPlayingItem.SetEnabled(true)
	log.Printf("✓ 菜单项已更新为：%s", newLabel)
}

// rebuildTrayMenu 重建托盘菜单（用于语言切换时）
func rebuildTrayMenu() {
	log.Println("🔄 开始重建托盘菜单...")

	// 更新所有菜单项的标签
	playPauseItem.SetLabel(t("menu.playPause"))
	prevItem.SetLabel(t("menu.previousTrack"))
	nextItem.SetLabel(t("menu.nextTrack"))
	mainWindowItem.SetLabel(t("menu.showMainWindow"))
	browseItem.SetLabel(t("menu.browseSongs"))
	favoriteItem.SetLabel(t("menu.favoriteSongs"))
	downloadItem.SetLabel(t("menu.downloadMusic"))
	wakeItem.SetLabel(t("menu.keepAwake"))
	launchItem.SetLabel(t("menu.autoLaunch"))
	settingItem.SetLabel(t("menu.settings"))
	versionItem.SetLabel(fmt.Sprintf(t("menu.version"), AppVersion))
	quitItem.SetLabel(t("menu.quit"))

	playModeItem.SetLabel(t("menu.playMode"))

	musicLibItem.SetLabel(t("menu.musicLibrary"))
	buildMusicLibMenu()

	buildToolsMenu()

	updateNowPlayingItem()

	menu = application.NewMenuFromItems(
		nowPlayingItem,
		application.NewMenuItemSeparator(),
		playPauseItem,
		prevItem,
		nextItem,
		application.NewMenuItemSeparator(),
		browseItem,
		favoriteItem,
		playModeItem,
		musicLibItem,
		toolsMenuItem,
		downloadItem,
		wakeItem,
		launchItem,
		settingItem,
		mainWindowItem,
		application.NewMenuItemSeparator(),
		versionItem,
		quitItem,
	)

	tray.SetMenu(menu)
	log.Println("✅ 托盘菜单重建完成")
}

// setupTrayEventListeners 设置托盘相关的事件监听器
func setupTrayEventListeners() {
	// 监听当前歌曲变化事件
	app.Event.On("currentTrackChanged", func(event *application.CustomEvent) {
		log.Printf("收到歌曲变化事件：%v", event.Data)
		updateNowPlayingItem()
	})

	// 监听播放模式变化事件
	app.Event.On("playModeChanged", func(event *application.CustomEvent) {
		if mode, ok := event.Data.(string); ok {
			log.Printf("✓ 收到播放模式变化事件：%s", mode)
			updatePlayModeMenuLabels(mode)
		}
	})
}

// updatePlayModeMenuLabels 更新播放模式菜单标签
func updatePlayModeMenuLabels(mode string) {
	playModeOrder.SetLabel(func() string {
		if mode == "order" {
			return "✓ " + t("playMode.order")
		}
		return "  " + t("playMode.order")
	}())

	playModeLoop.SetLabel(func() string {
		if mode == "loop" {
			return "✓ " + t("playMode.loop")
		}
		return "  " + t("playMode.loop")
	}())

	playModeRandom.SetLabel(func() string {
		if mode == "random" {
			return "✓ " + t("playMode.random")
		}
		return "  " + t("playMode.random")
	}())

	playModeSingle.SetLabel(func() string {
		if mode == "single" {
			return "✓ " + t("playMode.single")
		}
		return "  " + t("playMode.single")
	}())

	log.Printf("✓ 托盘菜单播放模式已更新为：%s", mode)
}

// buildMusicLibMenu 构建音乐库菜单
func buildMusicLibMenu() {
	addLibItem := application.NewMenuItem(t("library.addNew"))
	addLibItem.OnClick(func(ctx *application.Context) {
		handleAddLibrary()
	})

	refreshLibItem := application.NewMenuItem(t("library.refreshCurrent"))
	refreshLibItem.SetAccelerator("CmdOrCtrl+R")
	refreshLibItem.OnClick(func(ctx *application.Context) {
		handleRefreshLibrary()
	})

	renameLibItem := application.NewMenuItem(t("library.renameCurrent"))
	renameLibItem.OnClick(func(ctx *application.Context) {
		log.Println(t("library.renameCurrent"))
	})

	deleteLibItem := application.NewMenuItem(t("library.deleteCurrent"))
	deleteLibItem.OnClick(func(ctx *application.Context) {
		handleDeleteLibrary()
	})

	var libItems []*application.MenuItem
	libraries := musicService.GetLibraries()
	currentLib := musicService.GetCurrentLibrary()
	currentLibName := ""
	if currentLib != nil {
		currentLibName = currentLib.Name
	}

	for _, libName := range libraries {
		label := libName
		if libName == currentLibName {
			label = "✓ " + libName
		} else {
			label = "  " + libName
		}

		libItem := application.NewMenuItem(label)
		libItem.OnClick(func(ctx *application.Context) {
			handleSwitchLibrary(libName, libItems)
		})
		libItems = append(libItems, libItem)
	}

	if len(libItems) == 0 {
		noLibItem := application.NewMenuItem(t("library.noLibrary"))
		noLibItem.SetEnabled(false)
		libItems = append(libItems, noLibItem)
	}

	musicLibMenuItems := append([]*application.MenuItem{}, libItems...)
	musicLibMenuItems = append(musicLibMenuItems, application.NewMenuItemSeparator())
	musicLibMenuItems = append(musicLibMenuItems, refreshLibItem, addLibItem, renameLibItem, deleteLibItem)

	if len(musicLibMenuItems) > 0 {
		musicLibMenu = application.NewMenuFromItems(musicLibMenuItems[0], musicLibMenuItems[1:]...)
	} else {
		musicLibMenu = application.NewMenu()
	}

	musicLibItem = application.NewSubmenu(t("menu.musicLibrary"), musicLibMenu)
}

// handleAddLibrary 处理添加音乐库
func handleAddLibrary() {
	log.Println(t("library.addNew"))

	if musicService == nil {
		log.Println("❌ musicService 为 nil")
		return
	}

	app.Event.Emit("showNotification", map[string]interface{}{
		"title":   t("notification.info"),
		"message": t("library.scanning"),
		"type":    "info",
	})

	if err := musicService.AddLibrary(); err != nil {
		log.Printf("添加音乐库失败：%v", err)
		app.Event.Emit("showNotification", map[string]interface{}{
			"title":   t("notification.error"),
			"message": fmt.Sprintf("添加音乐库失败: %v", err),
			"type":    "error",
		})
		return
	}

	log.Println("✓ 音乐库添加成功，刷新菜单")
	buildMusicLibMenu()

	go func() {
		time.Sleep(2 * time.Second)
		loadLibraryToPlaylist()
	}()
}

// handleRefreshLibrary 处理刷新音乐库
func handleRefreshLibrary() {
	log.Println(t("library.refreshCurrent"))

	if musicService == nil {
		return
	}

	currentLib := musicService.GetCurrentLibrary()
	if currentLib == nil {
		log.Println("当前没有音乐库")
		app.Event.Emit("showNotification", map[string]interface{}{
			"title":   t("notification.info"),
			"message": "当前没有音乐库",
			"type":    "info",
		})
		return
	}

	app.Event.Emit("showNotification", map[string]interface{}{
		"title":   t("notification.info"),
		"message": t("library.scanning"),
		"type":    "info",
	})

	go func() {
		if err := musicService.RefreshLibrary(); err != nil {
			log.Printf("刷新音乐库失败：%v", err)
			app.Event.Emit("showNotification", map[string]interface{}{
				"title":   t("notification.error"),
				"message": fmt.Sprintf("刷新音乐库失败: %v", err),
				"type":    "error",
			})
			return
		}

		loadLibraryToPlaylist()
	}()
}

// handleDeleteLibrary 处理删除音乐库
func handleDeleteLibrary() {
	log.Println(t("library.deleteCurrent"))

	if musicService == nil {
		log.Println("❌ musicService 为 nil")
		return
	}

	currentLib := musicService.GetCurrentLibrary()
	if currentLib == nil {
		log.Println("当前没有音乐库")
		return
	}

	libName := currentLib.Name
	log.Printf("⚠️ 准备删除音乐库：%s", libName)

	if err := musicService.DeleteLibrary(libName); err != nil {
		log.Printf("删除音乐库失败：%v", err)
		return
	}

	log.Printf("✓ 已删除音乐库：%s", libName)
	buildMusicLibMenu()
}

// handleSwitchLibrary 处理切换音乐库
func handleSwitchLibrary(libName string, libItems []*application.MenuItem) {
	log.Printf("切换到音乐库：%s", libName)

	if musicService == nil {
		log.Println("❌ musicService 为 nil")
		return
	}

	if err := musicService.SetCurrentLibrary(libName); err != nil {
		log.Printf("切换音乐库失败：%v", err)
		return
	}

	for _, item := range libItems {
		itemLabel := item.Label()
		if strings.HasPrefix(itemLabel, "✓ ") || strings.HasPrefix(itemLabel, "  ") {
			oldName := strings.TrimPrefix(strings.TrimPrefix(itemLabel, "✓ "), "  ")
			if oldName == libName {
				item.SetLabel("✓ " + oldName)
			} else {
				item.SetLabel("  " + oldName)
			}
		}
	}

	go func() {
		tracks, err := musicService.GetCurrentLibraryTracks()
		if err != nil {
			log.Printf("添加音轨失败  %v", err)
		}

		if len(tracks) == 0 {
			log.Printf("音乐库中没有音轨")
		}

		musicService.ClearPlaylist()

		for _, track := range tracks {
			if err := musicService.AddToPlaylist(track); err != nil {
				log.Printf("添加音轨失败 %s: %v", track, err)
			}
		}

		if len(tracks) > 0 {
			if err := musicService.PlayIndex(0); err != nil {
				log.Printf("播放失败: %v", err)
			}
		}

		log.Printf("已加载音乐库 %s 到播放列表，共 %d 首歌曲", musicService.GetCurrentLibrary().Name, len(tracks))
		log.Printf("✓ 音乐库 %s 加载完成，开始播放", libName)
	}()
}

// loadLibraryToPlaylist 加载音乐库到播放列表
func loadLibraryToPlaylist() {
	tracks, err := musicService.GetCurrentLibraryTracks()
	if err != nil {
		log.Printf("获取音轨失败: %v", err)
		app.Event.Emit("showNotification", map[string]interface{}{
			"title":   t("notification.error"),
			"message": fmt.Sprintf("获取音轨失败: %v", err),
			"type":    "error",
		})
		return
	}

	if len(tracks) == 0 {
		log.Printf("音乐库中没有音轨")
		app.Event.Emit("showNotification", map[string]interface{}{
			"title":   t("notification.info"),
			"message": "音乐库中没有音轨",
			"type":    "info",
		})
		return
	}

	app.Event.Emit("showNotification", map[string]interface{}{
		"title":   t("notification.info"),
		"message": t("library.loadingToPlaylist"),
		"type":    "info",
	})

	musicService.ClearPlaylist()

	for _, track := range tracks {
		if err := musicService.AddToPlaylist(track); err != nil {
			log.Printf("添加音轨失败 %s: %v", track, err)
		}
	}

	if len(tracks) > 0 {
		if err := musicService.PlayIndex(0); err != nil {
			log.Printf("播放失败: %v", err)
		}
	}

	currentLib := musicService.GetCurrentLibrary()
	libName := ""
	if currentLib != nil {
		libName = currentLib.Name
	}

	message := fmt.Sprintf("%s: %s (%d 首歌曲)", t("library.refreshSuccess"), libName, len(tracks))
	app.Event.Emit("showNotification", map[string]interface{}{
		"title":   t("notification.success"),
		"message": message,
		"type":    "success",
	})

	log.Printf("已加载音乐库 %s 到播放列表，共 %d 首歌曲", libName, len(tracks))
}

// buildToolsMenu 构建依赖工具菜单
func buildToolsMenu() {
	log.Println("🔧 构建依赖工具菜单...")

	tools := depManager.GetAllTools()
	var toolItems []*application.MenuItem

	for _, tool := range tools {
		statusIcon := "❌"
		switch tool.Status {
		case backend.ToolInstalled:
			statusIcon = "✅"
		case backend.ToolInstalling:
			statusIcon = "🔧"
		case backend.ToolInstallFailed:
			statusIcon = "⚠️"
		}

		label := fmt.Sprintf("%s %s", statusIcon, tool.Name)
		if tool.Version != "" && tool.Status == backend.ToolInstalled {
			shortVersion := tool.Version
			if len(shortVersion) > 20 {
				shortVersion = shortVersion[:20] + "..."
			}
			label = fmt.Sprintf("%s (%s)", label, shortVersion)
		}

		toolItem := application.NewMenuItem(label)

		if tool.Status == backend.ToolNotInstalled || tool.Status == backend.ToolInstallFailed {
			installSubItem := application.NewMenuItem("📦 安装 " + tool.Name)
			installSubItem.OnClick(func(ctx *application.Context) {
				log.Printf("📦 用户请求安装 %s", tool.Name)

				app.Event.Emit("showNotification", map[string]interface{}{
					"title":   "正在安装",
					"message": fmt.Sprintf("正在后台安装 %s，请稍候...", tool.Name),
					"type":    "info",
				})

				if err := depManager.InstallTool(tool.Command); err != nil {
					log.Printf("❌ 启动安装失败: %v", err)
					app.Event.Emit("showNotification", map[string]interface{}{
						"title":   "安装失败",
						"message": fmt.Sprintf("无法启动安装: %v", err),
						"type":    "error",
					})
				}
			})

			if tool.InstallHint != "" {
				hintItem := application.NewMenuItem("ℹ️ " + tool.InstallHint)
				hintItem.SetEnabled(false)
				toolItem = application.NewSubmenu(label, application.NewMenuFromItems(installSubItem, hintItem))
			} else {
				toolItem = application.NewSubmenu(label, application.NewMenuFromItems(installSubItem))
			}
		} else if tool.Status == backend.ToolInstalling {
			installingItem := application.NewMenuItem("⏳ 安装中...")
			installingItem.SetEnabled(false)
			toolItem = application.NewSubmenu(label, application.NewMenuFromItems(installingItem))
		}

		toolItems = append(toolItems, toolItem)
	}

	toolItems = append(toolItems, application.NewMenuItemSeparator())

	checkUpdatesItem := application.NewMenuItem("🔄 重新检查所有工具")
	checkUpdatesItem.OnClick(func(ctx *application.Context) {
		log.Println("🔄 用户请求重新检查所有工具")

		app.Event.Emit("showNotification", map[string]interface{}{
			"title":   "检查中",
			"message": "正在检查所有依赖工具...",
			"type":    "info",
		})

		go func() {
			depManager.CheckAllTools()
			summary := depManager.GetInstallSummary()
			log.Println(summary)
			buildToolsMenu()

			app.Event.Emit("showNotification", map[string]interface{}{
				"title":   "检查完成",
				"message": "依赖工具状态已更新",
				"type":    "success",
			})
		}()
	})
	toolItems = append(toolItems, checkUpdatesItem)

	if len(toolItems) > 0 {
		toolsMenu := application.NewMenuFromItems(toolItems[0], toolItems[1:]...)
		toolsMenuItem = application.NewSubmenu("🛠️ 依赖工具", toolsMenu)
	} else {
		toolsMenuItem = application.NewSubmenu("🛠️ 依赖工具", application.NewMenu())
	}

	log.Println("✅ 依赖工具菜单构建完成")
}
