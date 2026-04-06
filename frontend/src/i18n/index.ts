import zhCN from './locales/zh-CN.json'
import enUS from './locales/en-US.json'

export type Locale = 'zh-CN' | 'en-US'

export interface Translation {
  [key: string]: string | Translation
}

export const translations: Record<Locale, Translation> = {
  'zh-CN': zhCN as unknown as Translation,
  'en-US': enUS as unknown as Translation
}

export let currentLocale: Locale = 'zh-CN'

/**
 * 设置当前语言
 */
export function setLocale(locale: Locale) {
  if (translations[locale]) {
    currentLocale = locale
    localStorage.setItem('preferred-language', locale)
    console.log(`✓ 语言切换为: ${locale}`)
  } else {
    console.error(`不支持的语言: ${locale}`)
  }
}

/**
 * 获取当前语言
 */
export function getLocale(): Locale {
  // 尝试从 localStorage 读取
  const saved = localStorage.getItem('preferred-language') as Locale
  if (saved && translations[saved]) {
    return saved
  }
  return currentLocale
}

/**
 * 翻译函数（支持嵌套键，如 'menu.playPause'）
 * @param key - 翻译键
 * @param defaultValue - 默认值（当找不到翻译时返回）
 */
export function t(key: string, defaultValue?: string): string {
  const keys = key.split('.')
  let value: any = translations[currentLocale]
  
  for (const k of keys) {
    if (value && typeof value === 'object') {
      value = value[k]
    } else {
      return defaultValue || key // 找不到时返回默认值或原键
    }
  }
  
  return typeof value === 'string' ? value : (defaultValue || key)
}

/**
 * 初始化语言（从 localStorage 或默认值）
 */
export function initLocale() {
  const saved = getLocale()
  setLocale(saved)
}
