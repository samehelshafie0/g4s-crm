import { ref, watch, onMounted } from 'vue'

type Theme = 'light' | 'dark'

const THEME_STORAGE_KEY = 'crm-dashboard-theme'

const currentTheme = ref<Theme>('light')

export function useTheme() {
  const setTheme = (theme: Theme) => {
    currentTheme.value = theme
    document.documentElement.setAttribute('data-theme', theme)
    localStorage.setItem(THEME_STORAGE_KEY, theme)
  }

  const toggleTheme = () => {
    const newTheme = currentTheme.value === 'light' ? 'dark' : 'light'
    setTheme(newTheme)
  }

  const initTheme = () => {
    // Check localStorage first
    const savedTheme = localStorage.getItem(THEME_STORAGE_KEY) as Theme | null

    // Check system preference
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches

    const theme = savedTheme || (prefersDark ? 'dark' : 'light')
    setTheme(theme)
  }

  // Watch for system theme changes
  const watchSystemTheme = () => {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')

    mediaQuery.addEventListener('change', (e) => {
      if (!localStorage.getItem(THEME_STORAGE_KEY)) {
        setTheme(e.matches ? 'dark' : 'light')
      }
    })
  }

  return {
    currentTheme,
    setTheme,
    toggleTheme,
    initTheme,
    watchSystemTheme
  }
}
