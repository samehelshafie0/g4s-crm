import { ref } from 'vue'
import type { Toast } from '../components/UI/BaseToast.vue'

// Global toast instance
const toastInstance = ref<any>(null)

export function useToast() {
  const setToastInstance = (instance: any) => {
    toastInstance.value = instance
  }

  const showToast = (toast: Omit<Toast, 'id'>) => {
    if (toastInstance.value) {
      return toastInstance.value.addToast(toast)
    }
    console.warn('Toast instance not initialized')
    return null
  }

  const success = (message: string, title?: string, duration?: number) => {
    return showToast({ type: 'success', message, title, duration })
  }

  const error = (message: string, title?: string, duration?: number) => {
    return showToast({ type: 'error', message, title, duration })
  }

  const warning = (message: string, title?: string, duration?: number) => {
    return showToast({ type: 'warning', message, title, duration })
  }

  const info = (message: string, title?: string, duration?: number) => {
    return showToast({ type: 'info', message, title, duration })
  }

  return {
    setToastInstance,
    showToast,
    success,
    error,
    warning,
    info
  }
}
