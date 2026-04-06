package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"sync"
)

//go:embed *.json
var localeFiles embed.FS

// Translator 翻译器
type Translator struct {
	mu            sync.RWMutex
	currentLocale string
	translations  map[string]map[string]interface{}
}

var (
	instance *Translator
	once     sync.Once
)

// GetTranslator 获取单例翻译器
func GetTranslator() *Translator {
	once.Do(func() {
		instance = &Translator{
			currentLocale: "zh-CN", // 默认中文
			translations:  make(map[string]map[string]interface{}),
		}
		instance.loadAllLocales()
	})
	return instance
}

// loadAllLocales 加载所有语言文件
func (t *Translator) loadAllLocales() {
	files, err := localeFiles.ReadDir(".")
	if err != nil {
		log.Printf("读取语言文件目录失败: %v", err)
		return
	}

	for _, file := range files {
		if !file.IsDir() && len(file.Name()) > 5 { // zh-CN.json
			locale := file.Name()[:len(file.Name())-5] // 移除 .json
			if err := t.loadLocale(locale); err != nil {
				log.Printf("加载语言文件 %s 失败: %v", file.Name(), err)
			} else {
				log.Printf("✓ 已加载语言文件: %s", file.Name())
			}
		}
	}
}

// loadLocale 加载单个语言文件
func (t *Translator) loadLocale(locale string) error {
	data, err := localeFiles.ReadFile(locale + ".json")
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	var translations map[string]interface{}
	if err := json.Unmarshal(data, &translations); err != nil {
		return fmt.Errorf("解析 JSON 失败: %w", err)
	}

	t.translations[locale] = translations
	return nil
}

// SetLocale 设置当前语言
func (t *Translator) SetLocale(locale string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.translations[locale]; !exists {
		return fmt.Errorf("不支持的语言: %s", locale)
	}

	t.currentLocale = locale
	log.Printf("✓ 语言切换为: %s", locale)
	return nil
}

// GetLocale 获取当前语言
func (t *Translator) GetLocale() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.currentLocale
}

// GetSupportedLocales 获取支持的语言列表
func (t *Translator) GetSupportedLocales() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	locales := make([]string, 0, len(t.translations))
	for locale := range t.translations {
		locales = append(locales, locale)
	}
	return locales
}

// T 翻译指定键（支持点号分隔的嵌套键，如 "menu.playPause"）
func (t *Translator) T(key string) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	translations, exists := t.translations[t.currentLocale]
	if !exists {
		return key
	}

	return t.getNestedValue(translations, key)
}

// getNestedValue 获取嵌套的值
func (t *Translator) getNestedValue(data map[string]interface{}, key string) string {
	keys := splitKey(key)
	current := data

	for i, k := range keys {
		if i == len(keys)-1 {
			// 最后一层，返回值
			if val, ok := current[k]; ok {
				if str, ok := val.(string); ok {
					return str
				}
				return fmt.Sprintf("%v", val)
			}
			return key
		}

		// 中间层，继续深入
		if next, ok := current[k].(map[string]interface{}); ok {
			current = next
		} else {
			return key
		}
	}

	return key
}

// splitKey 按点号分割键
func splitKey(key string) []string {
	result := []string{}
	current := ""
	for _, char := range key {
		if char == '.' {
			result = append(result, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
