import type { StoredExercise, StoredSet, StoredWorkout } from './types';
import { DEMO_USER_ID } from './types';

// Helper to generate ISO timestamps
const now = new Date();
const daysAgo = (days: number) => {
  const date = new Date(now);
  date.setDate(date.getDate() - days);
  return date.toISOString();
};

const dateOnly = (days: number) => {
  const date = new Date(now);
  date.setDate(date.getDate() - days);
  return date.toISOString().split('T')[0];
};

// Demo exercises (5 common compound movements)
export const INITIAL_EXERCISES: StoredExercise[] = [
  {
    id: 1,
    name: "Barbell Squat",
    user_id: DEMO_USER_ID,
    created_at: daysAgo(30),
    updated_at: daysAgo(30),
  },
  {
    id: 2,
    name: "Bench Press",
    user_id: DEMO_USER_ID,
    created_at: daysAgo(30),
    updated_at: daysAgo(30),
  },
  {
    id: 3,
    name: "Deadlift",
    user_id: DEMO_USER_ID,
    created_at: daysAgo(30),
    updated_at: daysAgo(30),
  },
  {
    id: 4,
    name: "Overhead Press",
    user_id: DEMO_USER_ID,
    created_at: daysAgo(30),
    updated_at: daysAgo(30),
  },
  {
    id: 5,
    name: "Pull-ups",
    user_id: DEMO_USER_ID,
    created_at: daysAgo(25),
    updated_at: daysAgo(25),
  },
];

// Demo workouts (3 recent sessions showing progression)
export const INITIAL_WORKOUTS: StoredWorkout[] = [
  {
    id: 1,
    date: dateOnly(7), // 1 week ago
    notes: "Felt strong today, good depth on squats",
    workout_focus: "Lower Body",
    user_id: DEMO_USER_ID,
    created_at: daysAgo(7),
    updated_at: daysAgo(7),
  },
  {
    id: 2,
    date: dateOnly(5), // 5 days ago
    notes: "Bench press PR! Hit 185lbs for 3 reps",
    workout_focus: "Upper Body Push",
    user_id: DEMO_USER_ID,
    created_at: daysAgo(5),
    updated_at: daysAgo(5),
  },
  {
    id: 3,
    date: dateOnly(2), // 2 days ago
    notes: "Deadlifts moving well, grip was limiting factor",
    workout_focus: "Lower Body",
    user_id: DEMO_USER_ID,
    created_at: daysAgo(2),
    updated_at: daysAgo(2),
  },
];

// Demo sets (realistic progression across workouts)
export const INITIAL_SETS: StoredSet[] = [
  // Workout 1: Lower Body (Squat + Deadlift)
  // Exercise 1: Barbell Squat
  {
    id: 1,
    exercise_id: 1,
    workout_id: 1,
    weight: 135,
    reps: 8,
    set_type: "warmup",
    exercise_order: 0,
    set_order: 0,
    user_id: DEMO_USER_ID,
    created_at: daysAgo(7),
  },
  {
    id: 2,
    exercise_id: 1,
    workout_id: 1,
    weight: 185,
    reps: 5,
    set_type: "working",
    exercise_order: 0,
    set_order: 1,
    user_id: DEMO_USER_ID,
    created_at: daysAgo(7),
  },
  {
    id: 3,
    exercise_id: 1,
    workout_id: 1,
    weight: 185,
    reps: 5,
    set_type: "working",
    exercise_order: 0,
    set_order: 2,
    user_id: DEMO_USER_ID,
    created_at: daysAgo(7),
  },
  {
    id: 4,
    exercise_id: 1,
    workout_id: 1,
    weight: 185,
    reps: 5,
    set_type: "working",
    exercise_order: 0,
    set_order: 3,
    user_id: DEMO_USER_ID,
    created_at: daysAgo(7),
  },
  // Exercise 2: Deadlift
  {
    id: 5,
    exercise_id: 3,
    workout_id: 1,
    weight: 135,
    reps: 5,
    set_type: "warmup",
    exercise_order: 1,
    set_order: 0,
    user_id: DEMO_USER_ID,
    created_at: daysAgo(7),
  },
  {
    id: 6,
    exercise_id: 3,
    workout_id: 1,
    weight: 225,
    reps: 5,
    set_type: "working",
    exercise_order: 1,
    set_order: 1,
    user_id: DEMO_USER_ID,
    created_at: daysAgo(7),
  },
  {
    id: 7,
    exercise_id: 3,
    workout_id: 1,
    weight: 225,
    reps: 5,
    set_type: "working",
    exercise_order: 1,
    set_order: 2,
    user_id: DEMO_USER_ID,
    created_at: daysAgo(7),
  },

  // Workout 2: Upper Body Push (Bench Press + Overhead Press)
  // Exercise 1: Bench Press
  {
    id: 8,
    exercise_id: 2,
    workout_id: 2,
    weight: 95,
    reps: 10,
    set_type: "warmup",
    exercise_order: 0,
    set_order: 0,
    user_id: DEMO_USER_ID,
    created_at: daysAgo(5),
  },
  {
    id: 9,
    exercise_id: 2,
    workout_id: 2,
    weight: 135,
    reps: 5,
    set_type: "warmup",
    exercise_order: 0,
    set_order: 1,
    user_id: DEMO_USER_ID,
    created_at: daysAgo(5),
  },
  {
    id: 10,
    exercise_id: 2,
    workout_id: 2,
    weight: 185,
    reps: 3,
    set_type: "working",
    exercise_order: 0,
    set_order: 2,
    user_id: DEMO_USER_ID,
    created_at: daysAgo(5),
  },
  {
    id: 11,
    exercise_id: 2,
    workout_id: 2,
    weight: 165,
    reps: 8,
    set_type: "working",
    exercise_order: 0,
    set_order: 3,
    user_id: DEMO_USER_ID,
    created_at: daysAgo(5),
  },
  {
    id: 12,
    exercise_id: 2,
    workout_id: 2,
    weight: 165,
    reps: 8,
    set_type: "working",
    exercise_order: 0,
    set_order: 4,
    user_id: DEMO_USER_ID,
    created_at: daysAgo(5),
  },
  // Exercise 2: Overhead Press
  {
    id: 13,
    exercise_id: 4,
    workout_id: 2,
    weight: 65,
    reps: 8,
    set_type: "warmup",
    exercise_order: 1,
    set_order: 0,
    user_id: DEMO_USER_ID,
    created_at: daysAgo(5),
  },
  {
    id: 14,
    exercise_id: 4,
    workout_id: 2,
    weight: 95,
    reps: 5,
    set_type: "working",
    exercise_order: 1,
    set_order: 1,
    user_id: DEMO_USER_ID,
    created_at: daysAgo(5),
  },
  {
    id: 15,
    exercise_id: 4,
    workout_id: 2,
    weight: 95,
    reps: 5,
    set_type: "working",
    exercise_order: 1,
    set_order: 2,
    user_id: DEMO_USER_ID,
    created_at: daysAgo(5),
  },
  {
    id: 16,
    exercise_id: 4,
    workout_id: 2,
    weight: 95,
    reps: 4,
    set_type: "working",
    exercise_order: 1,
    set_order: 3,
    user_id: DEMO_USER_ID,
    created_at: daysAgo(5),
  },

  // Workout 3: Lower Body (Squat + Deadlift - progression from workout 1)
  // Exercise 1: Barbell Squat (increased weight)
  {
    id: 17,
    exercise_id: 1,
    workout_id: 3,
    weight: 135,
    reps: 8,
    set_type: "warmup",
    exercise_order: 0,
    set_order: 0,
    user_id: DEMO_USER_ID,
    created_at: daysAgo(2),
  },
  {
    id: 18,
    exercise_id: 1,
    workout_id: 3,
    weight: 195,
    reps: 5,
    set_type: "working",
    exercise_order: 0,
    set_order: 1,
    user_id: DEMO_USER_ID,
    created_at: daysAgo(2),
  },
  {
    id: 19,
    exercise_id: 1,
    workout_id: 3,
    weight: 195,
    reps: 5,
    set_type: "working",
    exercise_order: 0,
    set_order: 2,
    user_id: DEMO_USER_ID,
    created_at: daysAgo(2),
  },
  {
    id: 20,
    exercise_id: 1,
    workout_id: 3,
    weight: 195,
    reps: 4,
    set_type: "working",
    exercise_order: 0,
    set_order: 3,
    user_id: DEMO_USER_ID,
    created_at: daysAgo(2),
  },
  // Exercise 2: Deadlift (increased weight)
  {
    id: 21,
    exercise_id: 3,
    workout_id: 3,
    weight: 135,
    reps: 5,
    set_type: "warmup",
    exercise_order: 1,
    set_order: 0,
    user_id: DEMO_USER_ID,
    created_at: daysAgo(2),
  },
  {
    id: 22,
    exercise_id: 3,
    workout_id: 3,
    weight: 245,
    reps: 5,
    set_type: "working",
    exercise_order: 1,
    set_order: 1,
    user_id: DEMO_USER_ID,
    created_at: daysAgo(2),
  },
  {
    id: 23,
    exercise_id: 3,
    workout_id: 3,
    weight: 245,
    reps: 5,
    set_type: "working",
    exercise_order: 1,
    set_order: 2,
    user_id: DEMO_USER_ID,
    created_at: daysAgo(2),
  },
  // Exercise 3: Pull-ups (bodyweight - no weight value)
  {
    id: 24,
    exercise_id: 5,
    workout_id: 3,
    reps: 8,
    set_type: "working",
    exercise_order: 2,
    set_order: 0,
    user_id: DEMO_USER_ID,
    created_at: daysAgo(2),
  },
  {
    id: 25,
    exercise_id: 5,
    workout_id: 3,
    reps: 7,
    set_type: "working",
    exercise_order: 2,
    set_order: 1,
    user_id: DEMO_USER_ID,
    created_at: daysAgo(2),
  },
  {
    id: 26,
    exercise_id: 5,
    workout_id: 3,
    reps: 6,
    set_type: "working",
    exercise_order: 2,
    set_order: 2,
    user_id: DEMO_USER_ID,
    created_at: daysAgo(2),
  },
];

// Helper to get next available ID
export const getNextId = (items: Array<{ id: number }>): number => {
  if (items.length === 0) return 1;
  return Math.max(...items.map((item) => item.id)) + 1;
};
