package main

import (
	"embed"
	_ "embed"
	"log"
	"runtime/debug"
	"time"

	"github.com/yhao521/wailsMusicPlay/backend"

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
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	musicService.SetApp(app)

	// 初始化音乐服务
	if err := musicService.Init(); err != nil {
		log.Printf("初始化音乐服务失败：%v", err)
	}

	// 声明窗口变量（先初始化为 nil）
	var mainWindow *application.WebviewWindow

	// 创建系统托盘（在窗口创建之前）
	tray := app.SystemTray.New()
	log.Println("✓ System tray initialized")

	// 设置托盘图标
	tray.SetIcon(trayIcon)
	tray.SetTooltip("Haoyun Music Player")

	// 先声明所有菜单项变量（以便在闭包中使用）
	var playPauseItem, prevItem, nextItem, showItem, browseItem *application.MenuItem
	var downloadItem, wakeItem, launchItem, settingItem, versionItem, quitItem *application.MenuItem
	var playModeItem, musicLibItem *application.MenuItem
	var musicLibMenu *application.Menu
	var menu *application.Menu

	// 创建基本播放控制菜单项
	playPauseItem = application.NewMenuItem("播放")
	playPauseItem.OnClick(func(ctx *application.Context) {
		// 检查当前是否有播放列表
		playlist, _ := musicService.GetPlaylist()

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
	prevItem.OnClick(func(ctx *application.Context) {
		if err := musicService.Previous(); err != nil {
			log.Printf("切换上一曲失败：%v", err)
		}
	})

	nextItem = application.NewMenuItem("下一曲")
	nextItem.OnClick(func(ctx *application.Context) {
		if err := musicService.Next(); err != nil {
			log.Printf("切换下一曲失败：%v", err)
		}
	})

	showItem = application.NewMenuItem("显示主窗口")
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
	browseItem = application.NewMenuItem("浏览歌曲")
	browseItem.SetAccelerator("CmdOrCtrl+F")
	browseItem.OnClick(func(ctx *application.Context) {
		// TODO: 实现浏览歌曲功能
		log.Println("浏览歌曲")
	})

	// 创建播放模式子菜单（使用复选框显示当前模式）
	var playModeOrder, playModeLoop, playModeRandom *application.MenuItem

	playModeOrder = application.NewMenuItemCheckbox("顺序播放", true)
	playModeOrder.OnClick(func(ctx *application.Context) {
		musicService.SetPlayMode("order")
		log.Println("✓ 切换到顺序播放")
		// 更新复选框状态
		playModeOrder.SetChecked(true)
		playModeLoop.SetChecked(false)
		playModeRandom.SetChecked(false)
	})

	playModeLoop = application.NewMenuItemCheckbox("循环播放", true)
	playModeLoop.OnClick(func(ctx *application.Context) {
		musicService.SetPlayMode("loop")
		log.Println("✓ 切换到循环播放")
		// 更新复选框状态
		playModeOrder.SetChecked(false)
		playModeLoop.SetChecked(true)
		playModeRandom.SetChecked(false)
	})

	playModeRandom = application.NewMenuItemCheckbox("随机播放", true)
	playModeRandom.OnClick(func(ctx *application.Context) {
		musicService.SetPlayMode("random")
		log.Println("✓ 切换到随机播放")
		// 更新复选框状态
		playModeOrder.SetChecked(false)
		playModeLoop.SetChecked(false)
		playModeRandom.SetChecked(true)
	})

	playModeMenu := application.NewMenuFromItems(
		playModeOrder,
		playModeLoop,
		playModeRandom,
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

		// 动态生成音乐库列表菜单
		var libItems []*application.MenuItem
		libraries := musicService.GetLibraries()
		for _, libName := range libraries {
			libItem := application.NewMenuItemCheckbox(libName, true)
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
		musicLibMenuItems = append(musicLibMenuItems, refreshLibItem, addLibItem, renameLibItem)

		if len(musicLibMenuItems) > 0 {
			musicLibMenu = application.NewMenuFromItems(musicLibMenuItems[0], musicLibMenuItems[1:]...)
		} else {
			musicLibMenu = application.NewMenu()
		}

		// 更新音乐库子菜单
		musicLibItem = application.NewSubmenu("音乐库", musicLibMenu)

		// 重新创建并设置托盘菜单
		menu = application.NewMenuFromItems(
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

		// 重新设置托盘菜单
		tray.SetMenu(menu)
	}

	// 创建下载音乐菜单项（带快捷键 Cmd+D）
	downloadItem = application.NewMenuItem("下载音乐")
	downloadItem.SetAccelerator("CmdOrCtrl+D")
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
	settingItem.OnClick(func(ctx *application.Context) {
		// TODO: 实现设置功能
		log.Println("设置")
	})

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
	// mainWindow.Hide()
	// log.Println("✓ Main window created (Hide)")
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
