// Vitest setup file
// This file runs before all tests

import { beforeEach } from 'vitest';

// jsdom provides localStorage, but we want to ensure it's clean before each test
beforeEach(() => {
  localStorage.clear();
});
