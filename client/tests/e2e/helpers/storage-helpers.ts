import type { Page } from '@playwright/test';

export async function getDemoWorkouts(page: Page) {
  return page.evaluate(() => {
    const data = localStorage.getItem('demo_workouts');
    return data ? JSON.parse(data) : null;
  });
}

export async function getDemoExercises(page: Page) {
  return page.evaluate(() => {
    const data = localStorage.getItem('demo_exercises');
    return data ? JSON.parse(data) : null;
  });
}

export async function clearAllStorage(page: Page) {
  await page.evaluate(() => localStorage.clear());
}

export async function verifyDemoDataExists(page: Page) {
  const workouts = await getDemoWorkouts(page);
  const exercises = await getDemoExercises(page);
  return workouts !== null && exercises !== null;
}
