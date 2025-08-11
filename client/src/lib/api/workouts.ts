import { queryOptions } from '@tanstack/react-query';
import { WorkoutsService, OpenAPI } from '../../generated';
import type { 
  workout_CreateWorkoutRequest,
  workout_WorkoutResponse,
  workout_WorkoutWithSetsResponse
} from '../../generated';

// Re-export generated types directly
export type { 
  workout_CreateWorkoutRequest as WorkoutFormValues,
  workout_WorkoutResponse as WorkoutData,
  workout_WorkoutWithSetsResponse as WorkoutWithSets
} from '../../generated';

export const getWorkouts = () => WorkoutsService.getWorkouts();
export const createWorkout = (data: workout_CreateWorkoutRequest) =>
  WorkoutsService.postWorkouts(data);
export const getWorkoutWithSets = (id: number) =>
  WorkoutsService.getWorkouts1(id);

export function workoutsQueryOptions(accessToken: string) {
  return queryOptions<workout_WorkoutResponse[], Error>({
    queryKey: ['workouts', 'list'],
    queryFn: async () => {
      OpenAPI.HEADERS = {
        'x-stack-access-token': accessToken,
      };
      return WorkoutsService.getWorkouts();
    },
  });
}

export function workoutByIdQueryOptions(workoutId: number, accessToken: string) {
  return queryOptions<workout_WorkoutWithSetsResponse[], Error>({
    queryKey: ['workouts', 'details', workoutId],
    queryFn: async () => {
      OpenAPI.HEADERS = {
        'x-stack-access-token': accessToken,
      };
      return WorkoutsService.getWorkouts1(workoutId);
    },
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
