import type { Exercise } from "@/lib/types";
import type { WorkoutFormValues } from "@/lib/types";
import { loadFromLocalStorage } from "@/lib/local-storage";

// MARK: Init values
export const MOCK_VALUES: WorkoutFormValues = {
  date: new Date(), // ! TODO: isoString or not?
  notes: '',
  exercises: [] as Exercise[],
};

export const getInitialValues = (userId: string): WorkoutFormValues => {
  const saved = loadFromLocalStorage(userId);
  return saved || MOCK_VALUES;
};