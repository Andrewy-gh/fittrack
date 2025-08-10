import { queryOptions } from '@tanstack/react-query';
import { WorkoutsService } from '../../generated';
import type { workout_CreateWorkoutRequest } from '../../generated';

// Type alias for better compatibility with existing code
export type WorkoutFormValues = workout_CreateWorkoutRequest;
export interface WorkoutData {
  id: number;
  date: string;
  notes: string | null;
  created_at: string;
  updated_at: string | null;
}

// Delegated functions using generated service
export const getWorkouts = () => WorkoutsService.getWorkouts();
export const createWorkout = (data: WorkoutFormValues) =>
  WorkoutsService.postWorkouts(data);

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
  return res.json();
}

export function workoutsQueryOptions(accessToken: string) {
  return queryOptions({
    queryKey: ['workouts', 'list'],
    queryFn: () => fetchWorkouts(accessToken),
  });
}

export interface WorkoutWithSets {
  workout_id: number;
  workout_date: string;
  workout_notes: string;
  exercise_id: number;
  exercise_name: string;
  set_id: number;
  set_type: string;
  weight: number;
  reps: number;
  volume: number;
}

export async function fetchWorkoutById(
  workoutId: number,
  accessToken: string
): Promise<WorkoutWithSets[]> {
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

export function workoutByIdQueryOptions(workoutId: number, accessToken: string) {
  return queryOptions<WorkoutWithSets[], Error>({
    queryKey: ['workouts', 'details', workoutId],
    queryFn: () => fetchWorkoutById(workoutId, accessToken),
  });
}

export interface Exercise {
  name: string;
  sets: {
    weight: number;
    reps: number;
    setType: 'warmup' | 'working';
  }[];
}

export function transformToWorkoutFormValues(workouts: WorkoutWithSets[]): WorkoutFormValues {
  if (workouts.length === 0) {
    return {
      date: new Date(),
      notes: '',
      exercises: [],
    };
  }

  // Group sets by exercise
  const exercisesMap = new Map<number, Exercise>();
  
  // Sort all workouts by set_id first to ensure consistent ordering
  const sortedWorkouts = [...workouts].sort((a, b) => a.set_id - b.set_id);

  for (const workout of sortedWorkouts) {
    if (!exercisesMap.has(workout.exercise_id)) {
      exercisesMap.set(workout.exercise_id, {
        name: workout.exercise_name,
        sets: [],
      });
    }

    const exercise = exercisesMap.get(workout.exercise_id)!;
    exercise.sets.push({
      weight: workout.weight,
      reps: workout.reps,
      setType: workout.set_type as 'warmup' | 'working',
    });
  }

  return {
    date: new Date(workouts[0].workout_date),
    notes: workouts[0].workout_notes || '',
    exercises: Array.from(exercisesMap.values()),
  };
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
