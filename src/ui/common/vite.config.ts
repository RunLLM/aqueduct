import { resolve } from 'path';
import { defineConfig } from 'vite';
import { externalizeDeps } from 'vite-plugin-externalize-deps';
import react from '@vitejs/plugin-react';

export default defineConfig(() => {
    return {
        plugins: [
            react(),
            externalizeDeps(),
        ],
        css: {
            modules: {
                localsConvention: 'camelCase',
                generateScopedName: "[name]__[local]___[hash:base64:5]",
            }
        },
        build: {
            sourcemap: 'inline',
            commonjsOptions: {
                include: ['src/index.tsx'],
            },
            lib: {
                entry: resolve(__dirname, 'src/index.tsx'),
                name: '@aqueducthq/common',
                fileName: 'common',
            }
        },
    }
})