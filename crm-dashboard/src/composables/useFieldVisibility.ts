import { ref, watch } from 'vue'
export function useFieldVisibility(key: string, defaults: Record<string, boolean>) {
 let saved: Record<string, boolean> = {}
 try { const parsed: unknown = JSON.parse(localStorage.getItem(key) ?? '{}'); if (parsed && typeof parsed === 'object') saved = Object.fromEntries(Object.entries(parsed).filter(([name, value]) => name in defaults && typeof value === 'boolean')) }
 catch { /* Invalid or unavailable local storage uses the visible defaults. */ }
 const visible = ref({ ...defaults, ...saved })
 watch(visible, value => { try { localStorage.setItem(key, JSON.stringify(value)) } catch { /* Preferences still work for this session. */ } }, { deep: true })
 return visible
}
