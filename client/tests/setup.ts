import { beforeEach } from 'vitest';

// Clear localStorage before each test
beforeEach(async () => {
  if (typeof localStorage !== 'undefined') {
    localStorage.clear();
  }
});
