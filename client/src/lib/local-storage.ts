import type { WorkoutFormValues } from '@/lib/types';

const STORAGE_KEY = 'workout-entry-form-data';

export const saveToLocalStorage = (data: WorkoutFormValues) => {
  try {
    // Convert Date objects to ISO strings for JSON serialization
    const serializedData = {
      ...data,
      date: data.date instanceof Date ? data.date.toISOString() : data.date,
    };
    localStorage.setItem(STORAGE_KEY, JSON.stringify(serializedData));
  } catch (error) {
    console.warn('Failed to save to localStorage:', error);
  }
};

export const loadFromLocalStorage = (): WorkoutFormValues | null => {
  console.log('Loading form data from localStorage');
  try {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (saved) {
      const parsed = JSON.parse(saved);
      // Convert ISO string back to Date object
      if (parsed.date) {
        parsed.date = new Date(parsed.date);
      }
      return parsed as WorkoutFormValues;
    }
  } catch (error) {
    console.warn('Failed to load from localStorage:', error);
  }
  return null;
};

export const clearLocalStorage = () => {
  try {
    localStorage.removeItem(STORAGE_KEY);
  } catch (error) {
    console.warn('Failed to clear localStorage:', error);
  }
};