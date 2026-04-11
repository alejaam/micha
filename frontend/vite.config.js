import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

export default defineConfig({
    plugins: [react()],
    test: {
        globals: true,
        environment: 'jsdom',
        setupFiles: './src/setupTests.js',
    },
    server: {
        port: 5173,
        proxy: {
            '/v1': {
                target: 'http://localhost:8080',
                changeOrigin: true,
            },
            '/health': {
                target: 'http://localhost:8080',
                changeOrigin: true,
            },
        },
    },
})
