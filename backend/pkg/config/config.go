package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/yhao521/haoyun-music-player/backend/pkg/i18n"
)

// AppConfig 应用程序配置结构
type AppConfig struct {
	Language        string `json:"language"`        // 语言设置: zh-CN, en-US
	Theme           string `json:"theme"`           // 主题: auto, light, dark
	AutoLaunch      bool   `json:"autoLaunch"`      // 开机启动
	KeepAwake       bool   `json:"keepAwake"`       // 保持唤醒
	DefaultVolume   int    `json:"defaultVolume"`   // 默认音量 (0-100)
	ShowLyrics      bool   `json:"showLyrics"`      // 显示歌词
	EnableMediaKeys bool   `json:"enableMediaKeys"` // 启用媒体键
	DefaultPlayMode string `json:"defaultPlayMode"` // 默认播放模式: order, loop, random, single
}

// DefaultConfig 返回默认配置
func DefaultConfig() *AppConfig {
	return &AppConfig{
		Language:        "zh-CN",
		Theme:           "auto",
		AutoLaunch:      false,
		KeepAwake:       true,
		DefaultVolume:   80,
		ShowLyrics:      true,
		EnableMediaKeys: true,
		DefaultPlayMode: "loop",
	}
}

// ConfigManager 配置管理器（单例模式）
type ConfigManager struct {
	mu     sync.RWMutex
	config *AppConfig
	path   string
}

var (
	instance *ConfigManager
	once     sync.Once
)

// GetConfigManager 获取配置管理器实例（单例）
func GetConfigManager() *ConfigManager {
	once.Do(func() {
		instance = &ConfigManager{
			config: DefaultConfig(),
		}
		// 初始化时加载配置
		if err := instance.Load(); err != nil {
			log.Printf("⚠️ 加载配置文件失败，使用默认配置: %v", err)
		}
	})
	return instance
}

// getConfigPath 获取配置文件路径
func (cm *ConfigManager) getConfigPath() string {
	if cm.path != "" {
		return cm.path
	}

	// 获取用户配置目录
	configDir, err := os.UserConfigDir()
	if err != nil {
		// 降级到当前目录
		configDir = "."
	}

	// 创建应用配置目录
	appConfigDir := filepath.Join(configDir, "haoyun-music-player")
	if err := os.MkdirAll(appConfigDir, 0755); err != nil {
		log.Printf("⚠️ 创建配置目录失败: %v", err)
		appConfigDir = "."
	}

	cm.path = filepath.Join(appConfigDir, "config.json")
	return cm.path
}

// Load 从文件加载配置
func (cm *ConfigManager) Load() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	configPath := cm.getConfigPath()

	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("📝 配置文件不存在，创建默认配置: %s", configPath)
		return cm.saveUnsafe()
	}

	// 读取文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析 JSON
	var loadedConfig AppConfig
	if err := json.Unmarshal(data, &loadedConfig); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 应用加载的配置
	cm.config = &loadedConfig

	log.Printf("✅ 配置已加载: %s", configPath)
	log.Printf("   - 语言: %s", cm.config.Language)
	log.Printf("   - 主题: %s", cm.config.Theme)
	log.Printf("   - 音量: %d", cm.config.DefaultVolume)

	return nil
}

// Save 保存配置到文件
func (cm *ConfigManager) Save() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	return cm.saveUnsafe()
}

// saveUnsafe 内部保存方法（调用者需持有锁）
func (cm *ConfigManager) saveUnsafe() error {
	configPath := cm.getConfigPath()

	// 序列化为 JSON（格式化输出，便于阅读）
	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	log.Printf("💾 配置已保存: %s", configPath)
	return nil
}

// Get 获取当前配置（线程安全）
func (cm *ConfigManager) Get() *AppConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// 返回副本，避免外部修改
	configCopy := *cm.config
	return &configCopy
}

// SetLanguage 设置语言并保存
func (cm *ConfigManager) SetLanguage(locale string) error {
	cm.mu.Lock()
	cm.config.Language = locale
	cm.mu.Unlock()

	// 同时更新翻译器
	translator := i18n.GetTranslator()
	if err := translator.SetLocale(locale); err != nil {
		return fmt.Errorf("设置翻译器语言失败: %w", err)
	}

	// 保存到文件
	return cm.Save()
}

// SetTheme 设置主题并保存
func (cm *ConfigManager) SetTheme(theme string) error {
	cm.mu.Lock()
	cm.config.Theme = theme
	cm.mu.Unlock()

	return cm.Save()
}

// SetAutoLaunch 设置开机启动并保存
func (cm *ConfigManager) SetAutoLaunch(enabled bool) error {
	cm.mu.Lock()
	cm.config.AutoLaunch = enabled
	cm.mu.Unlock()

	return cm.Save()
}

// SetKeepAwake 设置保持唤醒并保存
func (cm *ConfigManager) SetKeepAwake(enabled bool) error {
	cm.mu.Lock()
	cm.config.KeepAwake = enabled
	cm.mu.Unlock()

	return cm.Save()
}

// SetDefaultVolume 设置默认音量并保存
func (cm *ConfigManager) SetDefaultVolume(volume int) error {
	if volume < 0 || volume > 100 {
		return fmt.Errorf("音量必须在 0-100 之间")
	}

	cm.mu.Lock()
	cm.config.DefaultVolume = volume
	cm.mu.Unlock()

	return cm.Save()
}

// SetShowLyrics 设置显示歌词并保存
func (cm *ConfigManager) SetShowLyrics(show bool) error {
	cm.mu.Lock()
	cm.config.ShowLyrics = show
	cm.mu.Unlock()

	return cm.Save()
}

// SetEnableMediaKeys 设置启用媒体键并保存
func (cm *ConfigManager) SetEnableMediaKeys(enabled bool) error {
	cm.mu.Lock()
	cm.config.EnableMediaKeys = enabled
	cm.mu.Unlock()

	return cm.Save()
}

// SetDefaultPlayMode 设置默认播放模式并保存
func (cm *ConfigManager) SetDefaultPlayMode(mode string) error {
	validModes := map[string]bool{
		"order":  true,
		"loop":   true,
		"random": true,
		"single": true,
	}

	if !validModes[mode] {
		return fmt.Errorf("无效的播放模式: %s", mode)
	}

	cm.mu.Lock()
	cm.config.DefaultPlayMode = mode
	cm.mu.Unlock()

	return cm.Save()
}

// ApplyLanguageToTranslator 将配置的语言应用到翻译器
func (cm *ConfigManager) ApplyLanguageToTranslator() {
	cm.mu.RLock()
	locale := cm.config.Language
	cm.mu.RUnlock()

	translator := i18n.GetTranslator()
	if err := translator.SetLocale(locale); err != nil {
		log.Printf("⚠️ 应用语言设置失败: %v", err)
		return
	}

	log.Printf("✓ 已应用语言设置: %s", locale)
}
