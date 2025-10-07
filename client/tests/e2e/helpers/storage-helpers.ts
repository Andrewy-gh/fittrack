import type { Page } from '@playwright/test';

export async function getDemoWorkouts(page: Page) {
  return page.evaluate(() => {
    const data = localStorage.getItem('fittrack-demo-workouts');
    return data ? JSON.parse(data) : null;
  });
}

export async function getDemoExercises(page: Page) {
  return page.evaluate(() => {
    const data = localStorage.getItem('fittrack-demo-exercises');
    return data ? JSON.parse(data) : null;
  });
}

export async function clearAllStorage(page: Page) {
  await page.goto('/');
  await page.evaluate(() => localStorage.clear());
}

export async function verifyDemoDataExists(page: Page) {
  const workouts = await getDemoWorkouts(page);
  const exercises = await getDemoExercises(page);
  return workouts !== null && exercises !== null;
}

export async function resetDemoData(page: Page) {
  await page.evaluate(() => {
    const { resetDemoData } = require('@/lib/demo-data/storage');
    resetDemoData();
  });
}
