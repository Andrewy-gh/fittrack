// import type { WorkoutFormValues } from '@/lib/types';
export interface WorkoutData {
  id: number;
  date: string;
  notes: string | null;
  created_at: string;
  updated_at: string | null;
}

export async function fetchWorkouts(
  accessToken: string
): Promise<WorkoutData[]> {
  const res = await fetch('/api/workouts', {
    headers: {
      'x-stack-access-token': accessToken,
    },
  });
  if (!res.ok) {
    throw new Error('Failed to fetch workouts');
  }
  const data = await res.json();
  return data;
}

export interface WorkoutSet {
  workout_id: number;
  workout_date: string;
  workout_notes: string;
  exercise_id: number;
  exercise_name: string;
  set_id: number;
  weight: number;
  reps: number;
  set_type: string;
}

export async function fetchWorkoutById(
  workoutId: number,
  accessToken: string
): Promise<WorkoutSet[]> {
  const res = await fetch(`/api/workouts/${workoutId}`, {
    headers: {
      'x-stack-access-token': accessToken,
    },
  });
  if (!res.ok) {
    throw new Error(`Failed to fetch workout: ${res.status} ${res.statusText}`);
  }
  return res.json();
}

// export async function createWorkout(
//   workoutData: WorkoutFormValues,
//   accessToken: string
// ): Promise<any> {
//   const response = await fetch('/api/workouts', {
//     method: 'POST',
//     headers: {
//       'Content-Type': 'application/json',
//       'x-stack-access-token': accessToken,
//     },
//     body: JSON.stringify(workoutData),
//   });

//   if (!response.ok) {
//     const errorText = await response.text();
//     throw new Error(errorText ?? 'Failed to submit workout');
//   }

//   return response.json();
// }
