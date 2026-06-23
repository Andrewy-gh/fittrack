import { clearExerciseGoals } from "@/features/exercises/utils/exercise-goals";
import { clearDemoData } from "@/lib/demo-data/storage";
import { clearLocalStorage } from "@/lib/local-storage";

const chatResumeStorageKeyPrefix = "fittrack.ai-chat.resume:";

export function clearCurrentDeviceAccountState(userId?: string): void {
  clearLocalStorage();
  clearLocalStorage(userId);
  clearDemoData();
  clearExerciseGoals();
  clearSessionStorageByPrefix(chatResumeStorageKeyPrefix);
}

function clearSessionStorageByPrefix(prefix: string): void {
  if (typeof window === "undefined") {
    return;
  }

  const keysToRemove: string[] = [];
  for (let index = 0; index < window.sessionStorage.length; index += 1) {
    const key = window.sessionStorage.key(index);
    if (key?.startsWith(prefix)) {
      keysToRemove.push(key);
    }
  }

  for (const key of keysToRemove) {
    window.sessionStorage.removeItem(key);
  }
}
