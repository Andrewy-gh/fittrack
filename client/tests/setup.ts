// Vitest setup file
// This file runs before all tests

import { beforeEach, afterEach } from 'vitest';
import '@testing-library/jest-dom/vitest';
import { cleanup } from '@testing-library/react';

// Clean up after each test
afterEach(() => {
  cleanup();
});

// jsdom provides localStorage, but we want to ensure it's clean before each test
beforeEach(() => {
  localStorage.clear();
});
