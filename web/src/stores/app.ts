import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useAppStore = defineStore('app', () => {
  const locale = ref(localStorage.getItem('locale') || 'zh')
  const isDark = ref(localStorage.getItem('theme') === 'dark')
  const sidebarCollapsed = ref(false)

  function setLocale(lang: string) {
    locale.value = lang
    localStorage.setItem('locale', lang)
  }

  function toggleTheme() {
    isDark.value = !isDark.value
    localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
    document.documentElement.classList.toggle('dark', isDark.value)
  }

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  return {
    locale,
    isDark,
    sidebarCollapsed,
    setLocale,
    toggleTheme,
    toggleSidebar
  }
})
