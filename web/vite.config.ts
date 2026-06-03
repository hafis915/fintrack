import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { VitePWA } from 'vite-plugin-pwa'
import { fileURLToPath, URL } from 'node:url'

// Vite config — Vue 3 SPA + PWA shell. Backend lives on :8080 and we
// proxy /v1 + /health during dev so the frontend can stay relative-path.
export default defineConfig({
  plugins: [
    vue(),
    VitePWA({
      registerType: 'autoUpdate',
      manifest: {
        name: 'Fintrack',
        short_name: 'Fintrack',
        description: 'Money discipline that feels like training, not bookkeeping.',
        theme_color: '#0a0a0a',
        background_color: '#0a0a0a',
        display: 'standalone',
        start_url: '/',
        icons: [],
      },
    }),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 5173,
    proxy: {
      // Override with VITE_API_PROXY_TARGET when the API isn't on the default
      // dev port (e.g. Playwright runs on 8088 to dodge nginx).
      '/v1': process.env.VITE_API_PROXY_TARGET || 'http://localhost:8080',
      '/health': process.env.VITE_API_PROXY_TARGET || 'http://localhost:8080',
    },
  },
})
