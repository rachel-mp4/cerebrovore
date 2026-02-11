import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

// https://vite.dev/config/
export default defineConfig({
  build: {
    outDir: '../dist',
    emptyOutDir: true,
    manifest: true,
    rollupOptions: {
      input: {
        chat: 'src/chat.ts',
        beep: 'src/beep.ts',
      }
    }
  },
  plugins: [svelte()],
})
