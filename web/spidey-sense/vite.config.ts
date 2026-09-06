import react from '@vitejs/plugin-react';
import { loadEnv } from 'vite';
import { defineConfig } from 'vitest/config';

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, '.', 'SPIDEY_');
  return {
    plugins: [react()],
    server: {
      proxy: { '/api': env.SPIDEY_API_TARGET || 'http://127.0.0.1:4317' },
    },
    test: {
      environment: 'jsdom',
      setupFiles: './src/test/setup.ts',
      css: true,
      globals: true,
    },
  };
});
