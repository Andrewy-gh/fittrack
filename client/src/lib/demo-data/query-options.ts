import { queryOptions } from '@tanstack/react-query';
import { DEMO_EXERCISES, DEMO_EXERCISE_SETS, getDemoExerciseById, getDemoRecentSetsForExercise } from './exercises';
import { DEMO_WORKOUTS, DEMO_WORKOUT_SETS, DEMO_WORKOUT_FOCUS_VALUES, getDemoWorkoutById } from './workouts';
import type { WorkoutFocus } from '@/lib/api/workouts';

// Demo Exercise Query Options
export function demoExercisesQueryOptions() {
  return queryOptions({
    queryKey: ['demo', 'exercises'],
    queryFn: async () => DEMO_EXERCISES,
  });
}

export function demoExerciseByIdQueryOptions(id: number) {
  return queryOptions({
    queryKey: ['demo', 'exercises', id],
    queryFn: async () => getDemoExerciseById(id),
  });
}

export function demoRecentExerciseSetsQueryOptions(id: number) {
  return queryOptions({
    queryKey: ['demo', 'exercises', id, 'recent-sets'],
    queryFn: async () => getDemoRecentSetsForExercise(id),
  });
}

// Demo Workout Query Options
export function demoWorkoutsQueryOptions() {
  return queryOptions({
    queryKey: ['demo', 'workouts'],
    queryFn: async () => DEMO_WORKOUTS,
  });
}

export function demoWorkoutQueryOptions(id: number) {
  return queryOptions({
    queryKey: ['demo', 'workouts', id],
    queryFn: async () => getDemoWorkoutById(id),
  });
}

export function demoWorkoutsFocusValuesQueryOptions() {
  return queryOptions({
    queryKey: ['demo', 'workouts', 'focus-values'],
    queryFn: async () => DEMO_WORKOUT_FOCUS_VALUES,
  });
}

// Demo User
export const DEMO_USER = {
  id: 'demo-user',
  name: 'Demo User',
  email: 'demo@example.com',
} as const;