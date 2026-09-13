import path from 'node:path';
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: { '@': path.resolve(import.meta.dirname, './src') },
  },
  server: {
    host: '0.0.0.0',
    port: 5173,
    // Bind-mounted source on Linux hosts doesn't always emit inotify events.
    watch: { usePolling: true },
    proxy: {
      '/api': {
        target: process.env.VITE_API_PROXY || 'http://server:4000',
        changeOrigin: true,
      },
    },
  },
});
