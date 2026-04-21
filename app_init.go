package main

import (
	"embed"
	"fmt"
	"log"
	"time"

	"github.com/yhao521/haoyun-music-player/backend"
	"github.com/yhao521/haoyun-music-player/backend/pkg/config"
	"github.com/yhao521/haoyun-music-player/backend/pkg/i18n"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// AppVersion 应用版本信息
const AppVersion = "0.0.36"

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/tubiao.png
var trayIcon []byte

// TrackInfo 音乐文件信息（用于事件注册）
type TrackInfo = backend.TrackInfo

// 全局变量声明（供其他模块使用）
var (
	app           *application.App
	musicService  *backend.MusicService
	depManager    *backend.DependencyManager
	configManager *config.ConfigManager
	translator    *i18n.Translator
	mainWindow    *application.WebviewWindow
)

// t 翻译辅助函数
var t func(key string) string

func init() {
	// 注册所有事件
	registerEvents()
}

// registerEvents 注册所有 Wails 事件
func registerEvents() {
	application.RegisterEvent[string]("time")
	application.RegisterEvent[string]("playbackStateChanged")
	application.RegisterEvent[map[string]interface{}]("playbackProgress")
	application.RegisterEvent[[]string]("playlistUpdated")
	application.RegisterEvent[TrackInfo]("currentTrackChanged")
	application.RegisterEvent[map[string]interface{}]("windowUrl")
	application.RegisterEvent[[]string]("launchArgs")
	application.RegisterEvent[interface{}]("playbackEnded")
	application.RegisterEvent[[]backend.HistoryRecord]("historyUpdated")
	application.RegisterEvent[*backend.LyricInfo]("lyricLoaded")
	application.RegisterEvent[int]("currentLyricLineChanged")
	application.RegisterEvent[map[string]interface{}]("showNotification")
	application.RegisterEvent[map[string]interface{}]("compactLibraries")
	application.RegisterEvent[map[string]interface{}]("migrateToRelativePaths")
	application.RegisterEvent[string]("playModeChanged")
}

// initializeApp 初始化应用核心组件
func initializeApp() error {
	log.Println("🚀 开始初始化应用...")

	// 初始化配置管理器
	configManager = config.GetConfigManager()
	log.Println("✓ 配置管理器已初始化")

	// 应用配置的语言设置
	configManager.ApplyLanguageToTranslator()

	// 初始化国际化模块
	translator = i18n.GetTranslator()
	log.Printf("✓ 国际化模块已初始化，当前语言: %s", translator.GetLocale())

	// 创建翻译辅助函数
	t = func(key string) string {
		return translator.T(key)
	}

	// 创建依赖管理器
	depManager = backend.NewDependencyManager()
	log.Println("✓ 依赖管理器已初始化")

	// 创建统一的音乐服务（包含媒体键服务）
	musicService = backend.NewMusicService()
	log.Println("✓ 音乐服务已创建（包含媒体键服务）")

	// 创建 Wails 应用
	app = application.New(application.Options{
		Name:        "Haoyun Music Player",
		Description: "A menu bar music player built with Wails 3 + Vue 3",
		Services: []application.Service{
			application.NewService(musicService),
			application.NewService(depManager),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
	})

	musicService.SetApp(app)

	// 初始化音乐服务（内部会自动注册媒体键）
	if err := musicService.Init(); err != nil {
		return fmt.Errorf("初始化音乐服务失败：%w", err)
	}

	// 设置 OrganizeService 的 LyricManager 引用
	if organizeService := musicService.GetOrganizeService(); organizeService != nil {
		organizeService.SetLyricManager(musicService.GetLyricManager())
		log.Println("✓ OrganizeService 已关联 LyricManager")
	}

	log.Println("✅ 应用初始化完成")
	return nil
}

// checkDependencies 异步检查依赖工具状态
func checkDependencies() {
	log.Println("🔍 初始检测依赖工具...")
	depManager.CheckAllTools()

	go func() {
		time.Sleep(1 * time.Second)

		log.Println("🔄 后台重新检查依赖工具...")
		depManager.CheckAllTools()

		summary := depManager.GetInstallSummary()
		log.Println(summary)

		if depManager.NeedInstall() {
			missingTools := depManager.GetMissingTools()
			var toolNames []string
			for _, tool := range missingTools {
				toolNames = append(toolNames, tool.Name)
			}

			if app != nil && app.Event != nil {
				app.Event.Emit("missingDependencies", map[string]interface{}{
					"tools":   toolNames,
					"message": fmt.Sprintf("检测到 %d 个依赖工具缺失，建议安装以获得完整功能", len(toolNames)),
				})
			} else if app != nil {
				log.Printf("[checkDependencies] 警告: app.Event 为 nil，跳过事件发送")
			}
		}
	}()
}
