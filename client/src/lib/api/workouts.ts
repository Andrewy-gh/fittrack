import { queryOptions } from '@tanstack/react-query';
import { WorkoutsService } from '../../generated';
import type { 
  workout_CreateWorkoutRequest,
  workout_WorkoutResponse,
  workout_WorkoutWithSetsResponse
} from '../../generated';

// Type aliases for better compatibility with existing code
export type WorkoutFormValues = workout_CreateWorkoutRequest;
export type WorkoutData = workout_WorkoutResponse;
export type WorkoutWithSets = workout_WorkoutWithSetsResponse;

// Delegated functions using generated service
export const getWorkouts = () => WorkoutsService.getWorkouts();
export const createWorkout = (data: WorkoutFormValues) =>
  WorkoutsService.postWorkouts(data);
export const getWorkoutWithSets = (id: number) =>
  WorkoutsService.getWorkouts1(id);

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
  return queryOptions<WorkoutData[], Error>({
    queryKey: ['workouts', 'list'],
    queryFn: () => fetchWorkouts(accessToken),
  });
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
      date: new Date().toISOString(),
      notes: '',
      exercises: [],
    };
  }

  // Group sets by exercise
  const exercisesMap = new Map<number, Exercise>();
  
  // Sort all workouts by set_id first to ensure consistent ordering
  const sortedWorkouts = [...workouts].sort((a, b) => (a.set_id || 0) - (b.set_id || 0));

  for (const workout of sortedWorkouts) {
    const exerciseId = workout.exercise_id || 0;
    if (!exercisesMap.has(exerciseId)) {
      exercisesMap.set(exerciseId, {
        name: workout.exercise_name || '',
        sets: [],
      });
    }

    const exercise = exercisesMap.get(exerciseId)!;
    exercise.sets.push({
      weight: workout.weight || 0,
      reps: workout.reps || 0,
      setType: (workout.set_type as 'warmup' | 'working') || 'working',
    });
  }

  return {
    date: workouts[0].workout_date || new Date().toISOString(),
    notes: workouts[0].workout_notes || '',
    exercises: Array.from(exercisesMap.values()),
  };
}
