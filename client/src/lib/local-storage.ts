import type { workout_CreateWorkoutRequest } from '@/generated';

const STORAGE_KEY = 'workout-entry-form-data';

const getStorageKey = (userId?: string): string => {
  return userId ? `${STORAGE_KEY}-${userId}` : STORAGE_KEY;
};

export const saveToLocalStorage = (data: workout_CreateWorkoutRequest, userId?: string) => {
  try {
    // Convert Date objects to ISO strings for JSON serialization
    const serializedData = {
      ...data,
      date: data.date instanceof Date ? data.date.toISOString() : data.date,
    };
    localStorage.setItem(getStorageKey(userId), JSON.stringify(serializedData));
  } catch (error) {
    console.warn('Failed to save to localStorage:', error);
  }
};

export const loadFromLocalStorage = (userId?: string): workout_CreateWorkoutRequest | null => {
  console.log('Loading form data from localStorage');
  try {
    const saved = localStorage.getItem(getStorageKey(userId));
    if (saved) {
      const parsed = JSON.parse(saved);
      // Convert ISO string back to Date object
      if (parsed.date) {
        parsed.date = new Date(parsed.date);
      }
      return parsed as workout_CreateWorkoutRequest;
    }
  } catch (error) {
    console.warn('Failed to load from localStorage:', error);
  }
  return null;
};

export const clearLocalStorage = (userId?: string) => {
  try {
    localStorage.removeItem(getStorageKey(userId));
  } catch (error) {
    console.warn('Failed to clear localStorage:', error);
  }
};