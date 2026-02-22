// Re-export types from generated API client
// DO NOT define types manually - always import from @/client to ensure compatibility

export type {
  // Workout types
  WorkoutWorkoutResponse,
  WorkoutWorkoutWithSetsResponse,
  WorkoutCreateWorkoutRequest,
  WorkoutUpdateWorkoutRequest,
  WorkoutExerciseInput,
  WorkoutSetInput,
  WorkoutUpdateExercise,
  WorkoutUpdateSet,

  // Exercise types
  ExerciseExerciseResponse,
  ExerciseExerciseDetailResponse,
  ExerciseExerciseDetailExerciseResponse,
  ExerciseExerciseWithSetsResponse,
  ExerciseRecentSetsResponse,
  ExerciseCreateExerciseRequest,
  ExerciseCreateExerciseResponse,

  // Response types
  ResponseSuccessResponse,
  ResponseErrorResponse,
} from '@/client';

// Demo-specific constants
export const DEMO_USER_ID = "demo-user" as const;

export const STORAGE_KEYS = {
  WORKOUTS: 'fittrack-demo-workouts',
  EXERCISES: 'fittrack-demo-exercises',
  SETS: 'fittrack-demo-sets',
  HISTORICAL_1RM: 'fittrack-demo-historical-1rm',
} as const;

// Internal storage types (not exposed via API)
export type StoredSet = {
  id: number;
  exercise_id: number;
  workout_id: number;
  weight?: number;
  reps: number;
  set_type: "warmup" | "working";
  exercise_order: number;
  set_order: number;
  user_id: string;
  created_at: string;
};

export type StoredWorkout = {
  id: number;
  date: string;
  notes?: string;
  workout_focus?: string;
  user_id: string;
  created_at: string;
  updated_at: string;
};

export type StoredExercise = {
  id: number;
  name: string;
  user_id: string;
  created_at: string;
  updated_at: string;
};
