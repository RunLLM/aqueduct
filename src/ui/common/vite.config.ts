import { resolve } from 'path';
import { defineConfig } from 'vite';
import { externalizeDeps } from 'vite-plugin-externalize-deps';
import react from '@vitejs/plugin-react';

// https://vitejs.dev/config/
// export default defineConfig(({ command, mode }) => {
//     plugins: [react()],

// })

export default defineConfig(({ command, mode }) => {
    if (command === 'serve') {
        console.log('running in serve mode.')
    } else {
        // command === 'build'
        console.log('running vite in build mode.')
    }

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