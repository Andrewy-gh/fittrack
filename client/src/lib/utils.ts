import { clsx, type ClassValue } from 'clsx';
import { format } from 'date-fns';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatDate(dateString: string){
  return format(new Date(dateString), 'E MM/dd/yyyy')
}


// export const formatDate = (dateString: string) => {
//   const date = new Date(dateString);
//   const today = new Date();
//   const yesterday = new Date(today);
//   yesterday.setDate(yesterday.getDate() - 1);

//   if (date.toDateString() === today.toDateString()) {
//     return 'Today';
//   } else if (date.toDateString() === yesterday.toDateString()) {
//     return 'Yesterday';
//   } else {
//     return date.toLocaleDateString('en-US', {
//       weekday: 'long',
//       month: 'long',
//       day: 'numeric',
//     });
//   }
// };



export const formatTime = (dateString: string) => {
  return new Date(dateString).toLocaleTimeString('en-US', {
    hour: 'numeric',
    minute: '2-digit',
    hour12: true,
  });
};

export function formatWeight(weight: number | null | undefined): string {
  if (weight == null) return '0';
  // Return whole numbers without decimal point, decimals with one decimal place
  return weight % 1 === 0 ? weight.toString() : weight.toFixed(1);
}

export function sortByExerciseAndSetOrder<
  T extends {
    exercise_order?: number;
    set_order?: number;
    exercise_id?: number;
    set_id: number;
  },
>(data: T[]): T[] {
  return [...data].sort((a, b) => {
    const exerciseOrderA = a.exercise_order ?? a.exercise_id ?? 0;
    const exerciseOrderB = b.exercise_order ?? b.exercise_id ?? 0;
    if (exerciseOrderA !== exerciseOrderB) {
      return exerciseOrderA - exerciseOrderB;
    }
    const setOrderA = a.set_order ?? a.set_id ?? 0;
    const setOrderB = b.set_order ?? b.set_id ?? 0;
    return setOrderA - setOrderB;
  });
}