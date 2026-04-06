import { createApp } from 'vue'
import App from './App.vue'
import { initLocale } from './i18n'

// 初始化国际化
initLocale()

createApp(App).mount('#app')
