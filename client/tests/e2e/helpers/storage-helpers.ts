import type { BrowserPage } from '@vitest/browser/context';

export async function getDemoWorkouts(page: BrowserPage) {
  return page.evaluate(() => {
    const data = localStorage.getItem('demo_workouts');
    return data ? JSON.parse(data) : null;
  });
}

export async function getDemoExercises(page: BrowserPage) {
  return page.evaluate(() => {
    const data = localStorage.getItem('demo_exercises');
    return data ? JSON.parse(data) : null;
  });
}

export async function clearAllStorage(page: BrowserPage) {
  await page.evaluate(() => localStorage.clear());
}

export async function verifyDemoDataExists(page: BrowserPage) {
  const workouts = await getDemoWorkouts(page);
  const exercises = await getDemoExercises(page);
  return workouts !== null && exercises !== null;
}
