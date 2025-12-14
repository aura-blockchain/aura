import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./tests/setup.js'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html', 'lcov'],
      exclude: [
        'node_modules/',
        'tests/',
        'dist/',
        '*.config.js',
      ],
      statements: 90,
      branches: 85,
      functions: 90,
      lines: 90,
    },
    include: ['tests/**/*.test.js'],
    exclude: ['node_modules', 'dist'],
  },
});
