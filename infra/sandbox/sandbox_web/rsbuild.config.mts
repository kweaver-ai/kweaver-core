import { defineConfig } from '@rsbuild/core';
import { pluginReact } from '@rsbuild/plugin-react';
import { pluginLess } from '@rsbuild/plugin-less';

export default defineConfig({
  plugins: [pluginReact(), pluginLess()],
  source: {
    entry: {
      index: './src/main.tsx',
    },
  },
  html: {
    template: './public/index.html',
  },
  output: {
    distPath: {
      root: 'dist',
    },
    target: 'web',
    cleanDistPath: true,
  },
  server: {
    port: 1101,
    proxy: {
      '/api': {
        target: process.env.VITE_API_BASE_URL || 'http://localhost:8000',
        changeOrigin: true,
      },
    },
  },
  resolve: {
    alias: {
      '@': './src',
      '@components': './src/components',
      '@pages': './src/pages',
      '@utils': './src/utils',
      '@hooks': './src/hooks',
      '@types': './src/types',
      '@constants': './src/constants',
      '@apis': './src/apis',
      '@styles': './src/styles',
    },
  },
});
