package main

import (
	"embed"
	_ "embed"
	"fmt"
	"log"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/yhao521/wailsMusicPlay/backend"
	"github.com/yhao521/wailsMusicPlay/backend/pkg/file"
	"github.com/yhao521/wailsMusicPlay/backend/pkg/utils"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// TrackInfo 音乐文件信息（用于事件注册）
type TrackInfo = backend.TrackInfo

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
	application.RegisterEvent[TrackInfo]("currentTrackChanged")
	application.RegisterEvent[map[string]interface{}]("windowUrl")       // 添加窗口 URL 变化事件
	application.RegisterEvent[[]string]("launchArgs")                    // 添加第二实例启动参数事件
	application.RegisterEvent[interface{}]("playbackEnded")              // 添加播放结束事件
	application.RegisterEvent[[]backend.HistoryRecord]("historyUpdated") // 添加播放历史更新事件
	application.RegisterEvent[*backend.LyricInfo]("lyricLoaded")         // 添加歌词加载完成事件
	application.RegisterEvent[int]("currentLyricLineChanged")            // 添加当前歌词行变化事件
}

func main() {
	// 创建统一的音乐服务（MVC Model 层）
	musicService := backend.NewMusicService()

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
			ApplicationShouldTerminateAfterLastWindowClosed: false, // macOS: 关闭窗口时不退出程序
		},
	})

	musicService.SetApp(app)

	// 初始化音乐服务
	if err := musicService.Init(); err != nil {
		log.Printf("初始化音乐服务失败：%v", err)
	}

	// 声明窗口变量（先初始化为 nil）
	var mainWindow *application.WebviewWindow

	menus, playPauseMenuItem, prevMenuItem, nextMenuItem := createMenu(app)
	app.Menu.Set(menus)

	// 在 main 函数中设置播放控制菜单项的回调，以访问 musicService
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

	// 创建系统托盘（在窗口创建之前）
	tray := app.SystemTray.New()
	log.Println("✓ System tray initialized")

	// 设置托盘图标
	tray.SetIcon(trayIcon)
	tray.SetTooltip("Haoyun Music Player")

	// 先声明所有菜单项变量（以便在闭包中使用）
	var playPauseItem, prevItem, nextItem, mainWindowItem, browseItem *application.MenuItem
	var downloadItem, wakeItem, launchItem, settingItem, versionItem, quitItem *application.MenuItem
	var playModeItem, musicLibItem, favoriteItem *application.MenuItem
	var nowPlayingItem *application.MenuItem // 新增：正在播放的音乐名称
	var musicLibMenu *application.Menu
	var menu *application.Menu

	// 创建基本播放控制菜单项（带快捷键）
	playPauseItem = application.NewMenuItem("播放/暂停")
	playPauseItem.SetAccelerator("Space") // 空格键（注意：输入框中会输入空格）
	playPauseItem.OnClick(func(ctx *application.Context) {
		// 检查当前是否有播放列表
		playlist, _ := musicService.GetPlaylist()
		log.Println("GetPlaylist", len(playlist))

		if len(playlist) == 0 {
			// 如果播放列表为空，从当前音乐库加载
			log.Println("播放列表为空，从当前音乐库加载")

			currentLib := musicService.GetCurrentLibrary()
			if currentLib == nil {
				log.Println("当前没有音乐库，请先添加音乐库")
				return
			}

			// 从 JSON 文件加载音乐库到播放列表并播放
			if err := musicService.LoadCurrentLibrary(); err != nil {
				log.Printf("加载音乐库失败：%v", err)
				return
			}

			log.Printf("✓ 已从音乐库 %s 加载并播放", currentLib.Name)
		} else {
			// 如果已有播放列表，直接切换播放/暂停
			musicService.TogglePlayPause()
		}
	})

	prevItem = application.NewMenuItem("上一曲")
	prevItem.SetAccelerator("CmdOrCtrl+[") // Cmd/Ctrl + [ （类似浏览器后退）
	prevItem.OnClick(func(ctx *application.Context) {
		if err := musicService.Previous(); err != nil {
			log.Printf("切换上一曲失败：%v", err)
		}
	})

	nextItem = application.NewMenuItem("下一曲")
	nextItem.SetAccelerator("CmdOrCtrl+]") // Cmd/Ctrl + ] （类似浏览器前进）
	nextItem.OnClick(func(ctx *application.Context) {
		if err := musicService.Next(); err != nil {
			log.Printf("切换下一曲失败：%v", err)
		}
	})

	mainWindowItem = application.NewMenuItem("显示主窗口")
	mainWindowItem.OnClick(func(ctx *application.Context) {
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
		if isvisible {
			log.Println("准备调用 Hide()...")
			mainWindow.Hide()
		} else {
			log.Println("准备调用 Show()...")
			mainWindow.Show()
			log.Println("准备调用 Maximise()...")
			// mainWindow.Maximise()
			log.Println("✓ Maximise() 完成")

			mainWindow.Focus()
			log.Println("✓ Focus() 完成")
			go func() {
				if mainWindow != nil && mainWindow.IsVisible() {
					// 等待窗口完全初始化后再发送事件
					time.Sleep(100 * time.Millisecond)
					app.Event.Emit("windowUrl", map[string]interface{}{
						"type":       "test_message",
						"message":    "后端定时测试消息",
						"serverTime": time.Now().Format(time.RFC1123),
						"url":        "#/main",
					})
					log.Println("📤 [测试消息] 已发送到 mainWindow")
				}
			}()
		}

		log.Println("=== 操作完成 ===")
	})

	// 创建浏览歌曲菜单项（带快捷键 Cmd+F）
	// 注意：OnClick 回调会在后面窗口创建后重新定义
	browseItem = application.NewMenuItem("浏览歌曲")
	browseItem.SetAccelerator("CmdOrCtrl+F")

	// 创建播放模式子菜单（使用普通菜单项，通过标签显示当前模式）
	var playModeOrder, playModeLoop, playModeRandom, playModeSingle *application.MenuItem

	playModeOrder = application.NewMenuItem("  顺序播放")
	playModeOrder.OnClick(func(ctx *application.Context) {
		musicService.SetPlayMode("order")
		log.Println("✓ 切换到顺序播放")
		// 更新菜单标签
		playModeOrder.SetLabel("✓ 顺序播放")
		playModeLoop.SetLabel("  循环播放")
		playModeRandom.SetLabel("  随机播放")
		playModeSingle.SetLabel("  单曲循环")
	})

	playModeLoop = application.NewMenuItem("✓ 循环播放")
	playModeLoop.OnClick(func(ctx *application.Context) {
		musicService.SetPlayMode("loop")
		log.Println("✓ 切换到循环播放")
		// 更新菜单标签
		playModeOrder.SetLabel("  顺序播放")
		playModeLoop.SetLabel("✓ 循环播放")
		playModeRandom.SetLabel("  随机播放")
		playModeSingle.SetLabel("  单曲循环")
	})

	playModeRandom = application.NewMenuItem("  随机播放")
	playModeRandom.OnClick(func(ctx *application.Context) {
		musicService.SetPlayMode("random")
		log.Println("✓ 切换到随机播放")
		// 更新菜单标签
		playModeOrder.SetLabel("  顺序播放")
		playModeLoop.SetLabel("  循环播放")
		playModeRandom.SetLabel("✓ 随机播放")
		playModeSingle.SetLabel("  单曲循环")
	})

	playModeSingle = application.NewMenuItem("  单曲循环")
	playModeSingle.OnClick(func(ctx *application.Context) {
		musicService.SetPlayMode("random")
		log.Println("✓ 切换到单曲循环")
		// 更新菜单标签
		playModeOrder.SetLabel("  顺序播放")
		playModeLoop.SetLabel("  循环播放")
		playModeRandom.SetLabel("  随机播放")
		playModeSingle.SetLabel("✓ 单曲循环")
	})

	playModeMenu := application.NewMenuFromItems(
		playModeOrder,
		playModeLoop,
		playModeRandom,
		playModeSingle,
	)
	playModeItem = application.NewSubmenu("播放模式", playModeMenu)

	// 构建音乐库菜单的辅助函数
	var buildMusicLibMenu func()
	buildMusicLibMenu = func() {
		// 先创建"添加新音乐库"菜单项
		addLibItem := application.NewMenuItem("添加新音乐库")
		addLibItem.OnClick(func(ctx *application.Context) {
			log.Println("添加新音乐库")

			if musicService == nil {
				log.Println("❌ musicService 为 nil")
				return
			}

			// 添加音乐库
			if err := musicService.AddLibrary(); err != nil {
				log.Printf("添加音乐库失败：%v", err)
				return
			}

			// 添加成功后，重建音乐库菜单
			log.Println("✓ 音乐库添加成功，刷新菜单")
			buildMusicLibMenu()

			// 延迟一点时间，等待扫描完成后加载播放
			go func() {
				time.Sleep(2 * time.Second) // 等待扫描完成
				tracks, err := musicService.GetCurrentLibraryTracks()
				if err != nil {
					// return err
					log.Printf("添加音轨失败  %v", err)
				}

				if len(tracks) == 0 {
					log.Printf("音乐库中没有音轨")
				}

				// 清空当前播放列表
				musicService.ClearPlaylist()

				// 将所有音轨添加到播放列表
				for _, track := range tracks {
					if err := musicService.AddToPlaylist(track); err != nil {
						log.Printf("添加音轨失败 %s: %v", track, err)
					}
				}

				// 播放第一首
				if len(tracks) > 0 {
					if err := musicService.PlayIndex(0); err != nil {
						log.Printf("播放失败: %v", err)
					}
				}

				log.Printf("已加载音乐库 %s 到播放列表，共 %d 首歌曲", musicService.GetCurrentLibrary().Name, len(tracks))
			}()
		})

		// 创建"刷新当前音乐库"菜单项
		refreshLibItem := application.NewMenuItem("刷新当前音乐库")
		refreshLibItem.SetAccelerator("CmdOrCtrl+R")
		refreshLibItem.OnClick(func(ctx *application.Context) {
			log.Println("刷新当前音乐库")

			if musicService == nil {
				return
			}

			currentLib := musicService.GetCurrentLibrary()
			if currentLib == nil {
				log.Println("当前没有音乐库")
				return
			}

			go func() {
				// 刷新音乐库（重新扫描）
				if err := musicService.RefreshLibrary(); err != nil {
					log.Printf("刷新音乐库失败：%v", err)
					return
				}

				// 刷新成功后，重新加载到播放列表
				tracks, err := musicService.GetCurrentLibraryTracks()
				if err != nil {
					// return err
					log.Printf("添加音轨失败  %v", err)
				}

				if len(tracks) == 0 {
					log.Printf("音乐库中没有音轨")
				}

				// 清空当前播放列表
				musicService.ClearPlaylist()

				// 将所有音轨添加到播放列表
				for _, track := range tracks {
					if err := musicService.AddToPlaylist(track); err != nil {
						log.Printf("添加音轨失败 %s: %v", track, err)
					}
				}

				// 播放第一首
				if len(tracks) > 0 {
					if err := musicService.PlayIndex(0); err != nil {
						log.Printf("播放失败: %v", err)
					}
				}

				log.Printf("已加载音乐库 %s 到播放列表，共 %d 首歌曲", musicService.GetCurrentLibrary().Name, len(tracks))
			}()
		})

		// 创建"重命名当前音乐库"菜单项
		renameLibItem := application.NewMenuItem("重命名当前音乐库")
		renameLibItem.OnClick(func(ctx *application.Context) {
			log.Println("重命名当前音乐库")
			// TODO: 实现重命名功能
		})

		// 创建"删除当前音乐库"菜单项
		deleteLibItem := application.NewMenuItem("删除当前音乐库")
		deleteLibItem.OnClick(func(ctx *application.Context) {
			log.Println("删除当前音乐库")

			if musicService == nil {
				log.Println("❌ musicService 为 nil")
				return
			}

			currentLib := musicService.GetCurrentLibrary()
			if currentLib == nil {
				log.Println("当前没有音乐库")
				return
			}

			// 确认删除
			libName := currentLib.Name
			log.Printf("⚠️ 准备删除音乐库：%s", libName)

			// 执行删除（仅删除配置，不删除文件）
			if err := musicService.DeleteLibrary(libName); err != nil {
				log.Printf("删除音乐库失败：%v", err)
				return
			}

			log.Printf("✓ 已删除音乐库：%s", libName)

			// 重建菜单
			buildMusicLibMenu()
		})

		// 动态生成音乐库列表菜单
		var libItems []*application.MenuItem
		libraries := musicService.GetLibraries()
		currentLib := musicService.GetCurrentLibrary()
		currentLibName := ""
		if currentLib != nil {
			currentLibName = currentLib.Name
		}

		for _, libName := range libraries {
			// 使用 ✓ 符号标记当前选中的库
			label := libName
			if libName == currentLibName {
				label = "✓ " + libName
			} else {
				label = "  " + libName
			}

			libItem := application.NewMenuItem(label)
			libItem.OnClick(func(ctx *application.Context) {
				log.Printf("切换到音乐库：%s", libName)

				if musicService == nil {
					log.Println("❌ musicService 为 nil")
					return
				}

				// 切换当前音乐库
				if err := musicService.SetCurrentLibrary(libName); err != nil {
					log.Printf("切换音乐库失败：%v", err)
					return
				}

				// 更新所有库菜单项的标签
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

				// 加载音乐库到播放列表并开始播放
				go func() {
					tracks, err := musicService.GetCurrentLibraryTracks()
					if err != nil {
						// return err
						log.Printf("添加音轨失败  %v", err)
					}

					if len(tracks) == 0 {
						log.Printf("音乐库中没有音轨")
					}

					// 清空当前播放列表
					musicService.ClearPlaylist()

					// 将所有音轨添加到播放列表
					for _, track := range tracks {
						if err := musicService.AddToPlaylist(track); err != nil {
							log.Printf("添加音轨失败 %s: %v", track, err)
						}
					}

					// 播放第一首
					if len(tracks) > 0 {
						if err := musicService.PlayIndex(0); err != nil {
							log.Printf("播放失败: %v", err)
						}
					}

					log.Printf("已加载音乐库 %s 到播放列表，共 %d 首歌曲", musicService.GetCurrentLibrary().Name, len(tracks))
					log.Printf("✓ 音乐库 %s 加载完成，开始播放", libName)
				}()
			})
			libItems = append(libItems, libItem)
		}

		// 如果没有音乐库，显示提示
		if len(libItems) == 0 {
			noLibItem := application.NewMenuItem("暂无音乐库")
			noLibItem.SetEnabled(false)
			libItems = append(libItems, noLibItem)
		}

		// 组装音乐库菜单
		musicLibMenuItems := append([]*application.MenuItem{}, libItems...)
		musicLibMenuItems = append(musicLibMenuItems, application.NewMenuItemSeparator())
		musicLibMenuItems = append(musicLibMenuItems, refreshLibItem, addLibItem, renameLibItem, deleteLibItem)

		if len(musicLibMenuItems) > 0 {
			musicLibMenu = application.NewMenuFromItems(musicLibMenuItems[0], musicLibMenuItems[1:]...)
		} else {
			musicLibMenu = application.NewMenu()
		}

		// 更新音乐库子菜单
		musicLibItem = application.NewSubmenu("音乐库", musicLibMenu)

		// 创建"正在播放"菜单项（初始显示"无"）
		nowPlayingItem = application.NewMenuItem("正在播放：无")
		nowPlayingItem.SetEnabled(false) // 禁用点击

		// 重新创建并设置托盘菜单
		menu = application.NewMenuFromItems(
			nowPlayingItem, // 添加正在播放菜单项
			application.NewMenuItemSeparator(),
			playPauseItem,
			prevItem,
			nextItem,
			application.NewMenuItemSeparator(),
			browseItem,
			favoriteItem, // 添加喜爱音乐菜单
			playModeItem,
			musicLibItem,
			downloadItem,
			wakeItem,
			launchItem,
			settingItem,
			mainWindowItem,
			application.NewMenuItemSeparator(),
			versionItem,
			quitItem,
		)

		// 重新设置托盘菜单
		tray.SetMenu(menu)
	}
	favoriteItem = application.NewMenuItem("❤️ 喜爱音乐")
	favoriteItem.SetAccelerator("CmdOrCtrl+H") // Cmd/Ctrl + H (Heart)
	// OnClick 回调将在 favoritesWindow 创建后设置

	// 创建下载音乐菜单项（带快捷键 Cmd+D）
	downloadItem = application.NewMenuItem("下载音乐")
	downloadItem.SetAccelerator("CmdOrCtrl+D")
	downloadItem.SetEnabled(false)
	downloadItem.OnClick(func(ctx *application.Context) {
		// TODO: 实现下载音乐功能
		log.Println("下载音乐")
	})

	// 创建保持系统唤醒菜单项（带复选框）
	wakeItem = application.NewMenuItemCheckbox("保持系统唤醒", true)
	wakeItem.OnClick(func(ctx *application.Context) {
		// TODO: 实现保持唤醒功能
		log.Println("保持系统唤醒")
	})

	// 创建开机启动菜单项（带复选框）
	launchItem = application.NewMenuItemCheckbox("开机启动", true)
	launchItem.OnClick(func(ctx *application.Context) {
		// TODO: 实现开机启动功能
		log.Println("开机启动")
	})

	// 创建设置菜单项（带快捷键 Cmd+S）
	settingItem = application.NewMenuItem("设置")
	settingItem.SetAccelerator("CmdOrCtrl+S")
	// OnClick 回调将在 settingsWindow 创建后设置

	// 创建版本信息（禁用状态）
	versionItem = application.NewMenuItem("Version 0.5.0")
	versionItem.SetEnabled(false)

	// 创建退出菜单项
	quitItem = application.NewMenuItem("退出")
	quitItem.OnClick(func(ctx *application.Context) {
		app.Quit()
	})

	// 初始构建音乐库菜单（已包含菜单创建和设置）
	buildMusicLibMenu()

	// 初始化正在播放的音乐名称菜单项
	updateNowPlayingItem := func() {
		if musicService == nil {
			log.Println("❌ musicService 为 nil")
			return
		}

		trackName, err := musicService.GetCurrentTrackName()
		if err != nil {
			log.Printf("⚠️ 获取歌曲名称失败：%v", err)
			nowPlayingItem.SetLabel("未播放")
			nowPlayingItem.SetEnabled(false)
			return
		}

		log.Printf("✓ 更新正在播放：%s", trackName)

		// 截断过长的文件名（最多显示 30 个字符）
		displayName := trackName
		if len(displayName) > 30 {
			displayName = displayName[:27] + "..."
		}

		newLabel := fmt.Sprintf("🎵 %s", displayName)
		nowPlayingItem.SetLabel(newLabel)
		nowPlayingItem.SetEnabled(true)
		log.Printf("✓ 菜单项已更新为：%s", newLabel)
	}

	// 创建正在播放的音乐名称菜单项（禁用状态，仅展示）
	nowPlayingItem = application.NewMenuItem("未播放")
	nowPlayingItem.SetEnabled(false)

	// 监听当前歌曲变化事件
	app.Event.On("currentTrackChanged", func(event *application.CustomEvent) {
		log.Printf("收到歌曲变化事件：%v", event.Data)
		updateNowPlayingItem()
	})

	// 延迟初始化（等待服务完全启动）
	go func() {
		time.Sleep(500 * time.Millisecond)
		updateNowPlayingItem()
	}()

	// 交互事件
	// 注意：macOS 上单击托盘图标会自动显示菜单
	// 如果需要双击显示窗口，保留 OnDoubleClick

	log.Println("✓ System tray menu created")

	// ==================== 创建所有窗口 ====================

	// 创建主窗口（默认隐藏,通过托盘菜单打开）
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

	// 创建浏览歌曲窗口（用于展示音乐库和歌曲列表）
	var browseWindow *application.WebviewWindow
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

	// 创建喜爱音乐窗口（用于展示按播放次数排序的歌曲列表）
	var favoritesWindow *application.WebviewWindow
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

	// 创建设置窗口（用于应用程序设置）
	var settingsWindow *application.WebviewWindow
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

	// ==================== 注册所有窗口的关闭拦截钩子 ====================

	// 辅助函数：检查是否还有其他可见窗口（用于日志记录）
	hasOtherVisibleWindows := func(currentWindow string) bool {
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

	// 拦截主窗口关闭事件：始终隐藏，不真正关闭
	log.Println("🔧 正在为主窗口注册关闭拦截钩子...")
	mainWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		log.Println("⚠️ [主窗口] 关闭事件触发")

		if hasOtherVisibleWindows("main") {
			log.Println("ℹ️ [主窗口] 检测到其他可见窗口，但仍执行隐藏操作")
		} else {
			log.Println("ℹ️ [主窗口] 这是最后一个可见窗口")
		}

		// 统一行为：所有窗口关闭时都隐藏，不真正关闭
		mainWindow.Hide()
		e.Cancel() // 取消关闭操作
		log.Println("✅ [主窗口] 已隐藏并取消关闭")
	})
	log.Println("✅ 主窗口关闭拦截钩子注册成功")

	// 拦截浏览窗口关闭事件：始终隐藏，不真正关闭
	log.Println("🔧 正在为浏览窗口注册关闭拦截钩子...")
	browseWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		log.Println("⚠️ [浏览窗口] 关闭事件触发")

		if hasOtherVisibleWindows("browse") {
			log.Println("ℹ️ [浏览窗口] 检测到其他可见窗口，但仍执行隐藏操作")
		} else {
			log.Println("ℹ️ [浏览窗口] 这是最后一个可见窗口")
		}

		// 统一行为：所有窗口关闭时都隐藏，不真正关闭
		browseWindow.Hide()
		e.Cancel() // 取消关闭操作
		log.Println("✅ [浏览窗口] 已隐藏并取消关闭")
	})
	log.Println("✅ 浏览窗口关闭拦截钩子注册成功")

	// 拦截喜爱音乐窗口关闭事件：始终隐藏，不真正关闭
	log.Println("🔧 正在为喜爱音乐窗口注册关闭拦截钩子...")
	favoritesWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		log.Println("⚠️ [喜爱音乐窗口] 关闭事件触发")

		if hasOtherVisibleWindows("favorites") {
			log.Println("ℹ️ [喜爱音乐窗口] 检测到其他可见窗口，但仍执行隐藏操作")
		} else {
			log.Println("ℹ️ [喜爱音乐窗口] 这是最后一个可见窗口")
		}

		// 统一行为：所有窗口关闭时都隐藏，不真正关闭
		favoritesWindow.Hide()
		e.Cancel() // 取消关闭操作
		log.Println("✅ [喜爱音乐窗口] 已隐藏并取消关闭")
	})
	log.Println("✅ 喜爱音乐窗口关闭拦截钩子注册成功")

	// 拦截设置窗口关闭事件：始终隐藏，不真正关闭
	log.Println("🔧 正在为设置窗口注册关闭拦截钩子...")
	settingsWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		log.Println("⚠️ [设置窗口] 关闭事件触发")

		if hasOtherVisibleWindows("settings") {
			log.Println("ℹ️ [设置窗口] 检测到其他可见窗口，但仍执行隐藏操作")
		} else {
			log.Println("ℹ️ [设置窗口] 这是最后一个可见窗口")
		}

		// 统一行为：所有窗口关闭时都隐藏，不真正关闭
		settingsWindow.Hide()
		e.Cancel() // 取消关闭操作
		log.Println("✅ [设置窗口] 已隐藏并取消关闭")
	})
	log.Println("✅ 设置窗口关闭拦截钩子注册成功")

	// 设置喜爱音乐菜单项的点击事件（在 favoritesWindow 初始化之后）
	favoriteItem.OnClick(func(ctx *application.Context) {
		log.Println("打开喜爱音乐窗口")

		defer func() {
			if r := recover(); r != nil {
				log.Printf("❌ 打开喜爱音乐窗口时发生 panic: %v", r)
				debug.PrintStack()
			}
		}()

		if favoritesWindow == nil {
			log.Println("❌ favoritesWindow 为 nil")
			return
		}

		isVisible := favoritesWindow.IsVisible()
		log.Printf("✓ favoritesWindow IsVisible() = %v", isVisible)

		if isVisible {
			log.Println("准备调用 Hide()...")
			favoritesWindow.Hide()
		} else {
			log.Println("准备调用 Show()...")
			favoritesWindow.Show()
			log.Println("准备调用 Focus()...")
			favoritesWindow.Focus()
			log.Println("✓ Focus() 完成")
		}

		log.Println("=== 喜爱音乐窗口操作完成 ===")
	})

	// 设置设置菜单项的点击事件（在 settingsWindow 初始化之后）
	settingItem.OnClick(func(ctx *application.Context) {
		log.Println("打开设置窗口")

		defer func() {
			if r := recover(); r != nil {
				log.Printf("❌ 打开设置窗口时发生 panic: %v", r)
				debug.PrintStack()
			}
		}()

		if settingsWindow == nil {
			log.Println("❌ settingsWindow 为 nil")
			return
		}

		isVisible := settingsWindow.IsVisible()
		log.Printf("✓ settingsWindow IsVisible() = %v", isVisible)

		if isVisible {
			log.Println("准备调用 Hide()...")
			settingsWindow.Hide()
		} else {
			log.Println("准备调用 Show()...")
			settingsWindow.Show()
			log.Println("准备调用 Focus()...")
			settingsWindow.Focus()
			log.Println("✓ Focus() 完成")
		}

		log.Println("=== 设置窗口操作完成 ===")
	})

	// 定时向 browseWindow 发送测试消息（用于调试）
	go func() {
		time.Sleep(2 * time.Second) // 等待 2 秒后发送第一条测试消息
		// for {
		// 	time.Sleep(5 * time.Second) // 每 5 秒发送一次
		// }
	}()

	// 重新设置浏览歌曲菜单项的点击事件（在 browseWindow 初始化之后）
	// 托盘菜单
	browseItem.OnClick(func(ctx *application.Context) {
		// TODO: 实现浏览歌曲功能
		log.Println("浏览歌曲")
		defer func() {
			if r := recover(); r != nil {
				log.Printf("❌ 打开浏览窗口时发生 panic: %v", r)
				debug.PrintStack()
			}
		}()

		log.Println("=== 打开浏览歌曲窗口 ===")

		if browseWindow == nil {
			log.Println("❌ browseWindow 为 nil")
			return
		}

		isvisible := browseWindow.IsVisible()
		log.Println("✓ IsVisible() = ", isvisible)
		if isvisible {
			log.Println("准备调用 Hide()...")
			browseWindow.Hide()
		} else {
			log.Println("准备调用 Show()...")
			// 显示并最大化浏览窗口
			browseWindow.Show()
			browseWindow.Focus()
		}

		go func() {
			if browseWindow != nil && browseWindow.IsVisible() {
				// 等待窗口完全初始化后再发送事件
				time.Sleep(100 * time.Millisecond)
				app.Event.Emit("windowUrl", map[string]interface{}{
					"type":       "test_message",
					"message":    "后端定时测试消息",
					"serverTime": time.Now().Format(time.RFC1123),
					"url":        "#/browse",
				})
				log.Println("📤 [测试消息] 已发送到 browseWindow")
			}
		}()

		log.Println("✓ 浏览窗口已打开")
	})

	// ==================== 注册主界面 Music 菜单的事件监听器 ====================
	
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
	
	// 监听播放模式设置事件
	app.Event.On("setPlayMode", func(event *application.CustomEvent) {
		if mode, ok := event.Data.(string); ok {
			musicService.SetPlayMode(mode)
			log.Printf("✓ 切换到%s播放", mode)
		}
	})

	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}

	musicService.Shutdown()
}

func createMenu(app *application.App) (*application.Menu, *application.MenuItem, *application.MenuItem, *application.MenuItem) {
	menu := app.NewMenu()

	if runtime.GOOS == "darwin" {
		menu.AddRole(application.AppMenu) // macOS only
	}
	
	// File menu
	fileMenu := menu.AddSubmenu("File")
	fileMenu.Add("打开目录").
		SetAccelerator("Ctrl+O").OnClick(func(ctx *application.Context) {
		OpenDir()
	})
	fileMenu.Add("New").
		SetAccelerator("Ctrl+N").
		OnClick(func(ctx *application.Context) {
			// Create new document
		})
	fileMenu.Add("Open").
		SetAccelerator("Ctrl+O").
		OnClick(func(ctx *application.Context) {
			// Open file dialog
		})
	fileMenu.Add("Save").
		SetAccelerator("Ctrl+S").
		OnClick(func(ctx *application.Context) {
			// Save document
		})
	fileMenu.AddSeparator()
	fileMenu.Add("Exit").OnClick(func(ctx *application.Context) {
		app.Quit()
	})

	// Music menu (从托盘菜单复制的功能)
	musicMenu := menu.AddSubmenu("Music")
	
	// 正在播放菜单项（禁用状态，仅展示）
	nowPlayingMenuItem := musicMenu.Add("未播放")
	nowPlayingMenuItem.SetEnabled(false)
	
	musicMenu.AddSeparator()
	
	// 播放控制
	menuPlayPauseItem := musicMenu.Add("Play/Pause")
	menuPlayPauseItem.SetAccelerator("CmdOrCtrl+Space")
	// OnClick 会在 main 函数中通过事件机制设置
	
	menuPrevItem := musicMenu.Add("Previous Track")
	menuPrevItem.SetAccelerator("CmdOrCtrl+Shift+[")
	
	menuNextItem := musicMenu.Add("Next Track")
	menuNextItem.SetAccelerator("CmdOrCtrl+Shift+]")
	
	musicMenu.AddSeparator()
	
	// 窗口管理
	menuBrowseItem := musicMenu.Add("浏览歌曲")
	menuBrowseItem.SetAccelerator("CmdOrCtrl+Shift+F")
	
	menuFavoriteItem := musicMenu.Add("❤️ 喜爱音乐")
	menuFavoriteItem.SetAccelerator("CmdOrCtrl+Shift+H")
	
	menuMainWindowItem := musicMenu.Add("显示主窗口")
	menuMainWindowItem.OnClick(func(ctx *application.Context) {
		log.Println("显示主窗口（从主菜单）")
		// 通过事件机制触发,在 main 函数中处理
		app.Event.Emit("openWindow", map[string]interface{}{"type": "main"})
	})
	
	menuSettingItem := musicMenu.Add("设置")
	menuSettingItem.SetAccelerator("CmdOrCtrl+Shift+S")
	menuSettingItem.OnClick(func(ctx *application.Context) {
		log.Println("打开设置窗口（从主菜单）")
		app.Event.Emit("openWindow", map[string]interface{}{"type": "settings"})
	})
	
	musicMenu.AddSeparator()
	
	// 播放模式子菜单
	playModeSubMenu := musicMenu.AddSubmenu("播放模式")
	menuPlayModeOrder := playModeSubMenu.Add("  顺序播放")
	menuPlayModeOrder.OnClick(func(ctx *application.Context) {
		log.Println("切换到顺序播放")
		app.Event.Emit("setPlayMode", "order")
	})
	
	menuPlayModeLoop := playModeSubMenu.Add("✓ 循环播放")
	menuPlayModeLoop.OnClick(func(ctx *application.Context) {
		log.Println("切换到循环播放")
		app.Event.Emit("setPlayMode", "loop")
	})
	
	menuPlayModeRandom := playModeSubMenu.Add("  随机播放")
	menuPlayModeRandom.OnClick(func(ctx *application.Context) {
		log.Println("切换到随机播放")
		app.Event.Emit("setPlayMode", "random")
	})
	
	menuPlayModeSingle := playModeSubMenu.Add("  单曲循环")
	menuPlayModeSingle.OnClick(func(ctx *application.Context) {
		log.Println("切换到单曲循环")
		app.Event.Emit("setPlayMode", "single")
	})
	
	// 注意：这些菜单项的 OnClick 回调需要在 main 函数中设置,因为它们需要访问 musicService
	// 这里只是创建占位符
	
	// 音乐库子菜单（简化版）
	musicLibSubMenu := musicMenu.AddSubmenu("音乐库")
	musicLibSubMenu.Add("刷新当前音乐库").SetAccelerator("CmdOrCtrl+Shift+R")
	musicLibSubMenu.Add("添加新音乐库")
	
	musicMenu.AddSeparator()
	
	// 其他功能
	menuDownloadItem := musicMenu.Add("下载音乐")
	menuDownloadItem.SetAccelerator("CmdOrCtrl+Shift+D")
	menuDownloadItem.SetEnabled(false)
	
	menuWakeItem := musicMenu.AddCheckbox("保持系统唤醒", true)
	menuWakeItem.OnClick(func(ctx *application.Context) {
		log.Println("保持系统唤醒（从主菜单）")
	})
	
	menuLaunchItem := musicMenu.AddCheckbox("开机启动", true)
	menuLaunchItem.OnClick(func(ctx *application.Context) {
		log.Println("开机启动（从主菜单）")
	})
	
	menuVersionItem := musicMenu.Add("Version 0.5.0")
	menuVersionItem.SetEnabled(false)

	// Add development menu
	devMenu := menu.AddSubmenu("Development")
	//重载app
	devMenu.Add("Reload Application").OnClick(func(ctx *application.Context) {
		// Reload the application
		window := app.Window.Current()
		if window != nil {
			window.Reload()
		}
	})
	//打开控制台
	devMenu.Add("Open DevTools").OnClick(func(ctx *application.Context) {
		window := app.Window.Current()
		if window != nil {
			window.OpenDevTools()
		}
	})

	devMenu.Add("Show Environment").OnClick(func(ctx *application.Context) {
		// showEnvironmentDialog(app)
	})

	// Edit menu
	editMenu := menu.AddSubmenu("Edit")
	editMenu.Add("Undo").SetAccelerator("Ctrl+Z")
	editMenu.Add("Redo").SetAccelerator("Ctrl+Y")
	editMenu.AddSeparator()
	editMenu.Add("Cut").SetAccelerator("Ctrl+X")
	editMenu.Add("Copy").SetAccelerator("Ctrl+C")
	editMenu.Add("Paste").SetAccelerator("Ctrl+V")

	// Playback menu (播放控制菜单 - 快捷键必须在这里才能生效)
	playbackMenu := menu.AddSubmenu("Playback")

	playPauseMenuItem := playbackMenu.Add("Play/Pause")
	playPauseMenuItem.SetAccelerator("Space")
	// OnClick 会在 main 函数中设置

	prevMenuItem := playbackMenu.Add("Previous Track")
	prevMenuItem.SetAccelerator("CmdOrCtrl+[")
	// OnClick 会在 main 函数中设置

	nextMenuItem := playbackMenu.Add("Next Track")
	nextMenuItem.SetAccelerator("CmdOrCtrl+]")
	// OnClick 会在 main 函数中设置

	// View menu
	viewMenu := menu.AddSubmenu("View")
	darkMode := viewMenu.AddCheckbox("Dark Mode", false)
	darkMode.OnClick(func(ctx *application.Context) {
		// Toggle dark mode
		isChecked := darkMode.Checked()
		app.Logger.Info("Dark mode", "enabled", isChecked)
	})
	viewMenu.AddSeparator()
	// 必须保存 radio items 的引用，避免 Wails 处理 radio groups 时出现空指针错误
	_ = viewMenu.AddRadio("List View", true)
	_ = viewMenu.AddRadio("Grid View", false)
	_ = viewMenu.AddRadio("Detail View", false)

	// Help menu
	helpMenu := menu.AddSubmenu("Help")
	helpMenu.Add("Documentation").OnClick(func(ctx *application.Context) {
		// Open docs
	})
	helpMenu.Add("About").OnClick(func(ctx *application.Context) {
		// Show about dialog
	})

	return menu, playPauseMenuItem, prevMenuItem, nextMenuItem
}

func OpenDir() {
	appPath := file.GetAppPath()
	if runtime.GOOS == "darwin" {
		utils.OpenMac(appPath)
	} else {
		utils.OpenWin(appPath)
	}
}
