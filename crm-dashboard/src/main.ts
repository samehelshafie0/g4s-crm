import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import { useAuthStore } from './stores/auth'

import './assets/main.css'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)

// Rehydrate user profile from stored token before first route renders
const authStore = useAuthStore()
authStore.fetchMe().finally(() => {
  app.mount('#app')
})
