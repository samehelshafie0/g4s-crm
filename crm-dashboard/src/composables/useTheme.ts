import { ref, watch, readonly } from 'vue'

export type ThemeMode = 'light' | 'dark' | 'system'

const STORAGE_KEY = 'g4s-crm-theme'

const mode = ref<ThemeMode>((localStorage.getItem(STORAGE_KEY) as ThemeMode) || 'system')
const isDark = ref(false)

function getSystemPreference(): boolean {
  return window.matchMedia('(prefers-color-scheme: dark)').matches
}

function applyTheme(animate = false) {
  const dark = mode.value === 'dark' || (mode.value === 'system' && getSystemPreference())
  isDark.value = dark

  if (animate) {
    document.documentElement.classList.add('theme-transition')
    window.setTimeout(() => document.documentElement.classList.remove('theme-transition'), 400)
  }

  document.documentElement.setAttribute('data-theme', dark ? 'dark' : 'light')
}

function setMode(newMode: ThemeMode) {
  mode.value = newMode
  localStorage.setItem(STORAGE_KEY, newMode)
  applyTheme(true)
}

function toggle() {
  if (mode.value === 'system') {
    setMode(isDark.value ? 'light' : 'dark')
  } else {
    setMode(isDark.value ? 'light' : 'dark')
  }
}

function cycleMode() {
  const order: ThemeMode[] = ['light', 'dark', 'system']
  const idx = order.indexOf(mode.value)
  setMode(order[(idx + 1) % order.length] ?? 'system')
}

window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
  if (mode.value === 'system') applyTheme()
})

applyTheme()

export function useTheme() {
  return {
    mode: readonly(mode),
    isDark: readonly(isDark),
    setMode,
    toggle,
    cycleMode,
  }
}
