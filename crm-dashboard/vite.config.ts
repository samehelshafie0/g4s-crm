import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    vueDevTools(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
  },
  server: {
    port: Number(process.env.CRM_WEB_PORT ?? 5174),
    strictPort: true,
    proxy: {
      '/api': {
        target: process.env.CRM_API_URL ?? 'http://127.0.0.1:18080',
        changeOrigin: true,
        secure: false,
      },
    },
  },
})
