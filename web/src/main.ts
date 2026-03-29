import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createI18n } from 'vue-i18n'
import ElementPlus from 'element-plus'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import 'element-plus/dist/index.css'

import App from './App.vue'
import router from './router'
import messages from './locales'
import { useAppStore } from './stores/app'

import './style.css'

const pinia = createPinia()
const app = createApp(App)

const i18n = createI18n({
  legacy: false,
  locale: localStorage.getItem('locale') || 'zh',
  messages
})

for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.use(pinia)
app.use(router)
app.use(i18n)
app.use(ElementPlus)

const appStore = useAppStore()
if (appStore.isDark) {
  document.documentElement.classList.add('dark')
}

app.mount('#app')
