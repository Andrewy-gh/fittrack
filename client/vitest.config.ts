import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  test: {
    // Browser mode configuration
    browser: {
      enabled: true,
      name: 'chromium',
      provider: 'playwright',
      headless: true,
      screenshotFailures: true,
    },
    // Use the existing dev server
    setupFiles: ['./tests/setup.ts'],
  },
});
