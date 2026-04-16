import tailwindcss from '@tailwindcss/vite';
import react from '@vitejs/plugin-react';
import path from 'path';
import { defineConfig, loadEnv } from 'vite';

export default defineConfig(({ mode }) => {
    const env = loadEnv(mode, process.cwd(), '');
    return {
        plugins: [react(), tailwindcss()],
        define: {
            'process.env.GEMINI_API_KEY': JSON.stringify(env.GEMINI_API_KEY),
        },
        resolve: {
            alias: {
                '@': path.resolve(__dirname, '.'),
            },
        },
        server: {
            proxy: {
                '/all': {
                    target: 'http://app:8080',
                    changeOrigin: true,
                    secure: false,
                },
                '/create': {
                    target: 'http://app:8080',
                    changeOrigin: true,
                    secure: false,
                },
                '/status': {
                    target: 'http://app:8080',
                    changeOrigin: true,
                    secure: false,
                },
                '/cancel': {
                    target: 'http://app:8080',
                    changeOrigin: true,
                    secure: false,
                },
            },
            hmr: process.env.DISABLE_HMR !== 'true',
        },
    };
});