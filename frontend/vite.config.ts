import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:3000', // Go backend adresin
        changeOrigin: true,
        // rewrite: (path) => path.replace(/^\/api/, '') // Eğer backend'de /api prefixi yoksa bunu kullanırsın ama senin backend'inde r.Route("/api"...) olduğu için buna gerek yok.
      }
    }
  }
})
