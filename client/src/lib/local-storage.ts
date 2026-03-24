import type { WorkoutCreateWorkoutRequest } from '../client/types.gen';

const STORAGE_KEY = 'workout-entry-form-data';

export type StorageLike = Pick<Storage, 'getItem' | 'setItem' | 'removeItem'>;

const getStorageKey = (userId?: string): string => {
  return userId ? `${STORAGE_KEY}-${userId}` : STORAGE_KEY;
};

// Type for form data that may contain Date objects
export type FormDataType = Omit<WorkoutCreateWorkoutRequest, 'date'> & {
  date: Date | string;
};

export type WorkoutDraftStorage = {
  save: (data: FormDataType, userId?: string) => void;
  load: (userId?: string) => WorkoutCreateWorkoutRequest | null;
  clear: (userId?: string) => void;
};

function serializeDraft(data: FormDataType): WorkoutCreateWorkoutRequest {
  return {
    ...data,
    date: data.date instanceof Date ? data.date.toISOString() : data.date,
  };
}

function deserializeDraft(
  rawValue: string
): WorkoutCreateWorkoutRequest | null {
  try {
    const parsed = JSON.parse(rawValue);
    if (parsed.date && typeof parsed.date === 'string') {
      parsed.date = new Date(parsed.date);
    }
    return parsed as WorkoutCreateWorkoutRequest;
  } catch {
    return null;
  }
}

export function createWorkoutDraftStorage(
  storage?: StorageLike
): WorkoutDraftStorage {
  return {
    save(data, userId) {
      if (!storage) {
        return;
      }

      try {
        storage.setItem(
          getStorageKey(userId),
          JSON.stringify(serializeDraft(data))
        );
      } catch (error) {
        console.warn('Failed to save to localStorage:', error);
      }
    },
    load(userId) {
      if (!storage) {
        return null;
      }

      try {
        const saved = storage.getItem(getStorageKey(userId));
        return saved ? deserializeDraft(saved) : null;
      } catch (error) {
        console.warn('Failed to load from localStorage:', error);
        return null;
      }
    },
    clear(userId) {
      if (!storage) {
        return;
      }

      try {
        storage.removeItem(getStorageKey(userId));
      } catch (error) {
        console.warn('Failed to clear localStorage:', error);
      }
    },
  };
}

function getBrowserStorage(): StorageLike | undefined {
  if (typeof window === 'undefined') {
    return undefined;
  }

  try {
    return window.localStorage;
  } catch (error) {
    console.warn('Failed to access localStorage:', error);
    return undefined;
  }
}

export const workoutDraftStorage = createWorkoutDraftStorage(
  getBrowserStorage()
);

export const saveToLocalStorage = (
  data: FormDataType,
  userId?: string,
  draftStorage: WorkoutDraftStorage = workoutDraftStorage
) => {
  draftStorage.save(data, userId);
};

export const loadFromLocalStorage = (
  userId?: string,
  draftStorage: WorkoutDraftStorage = workoutDraftStorage
): WorkoutCreateWorkoutRequest | null => {
  return draftStorage.load(userId);
};

export const clearLocalStorage = (
  userId?: string,
  draftStorage: WorkoutDraftStorage = workoutDraftStorage
) => {
  draftStorage.clear(userId);
};
