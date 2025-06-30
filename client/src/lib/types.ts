// GET /exercises
export interface ExerciseOption {
  id: number;
  name: string;
}

// POST /workouts
export type Exercise = {
  name: string;
  sets: Set[];
};

type Set = {
  weight: number | undefined;
  reps: number | undefined;
  setType: string;
};

export type Workout = {
  date: Date;
  exercises: Exercise[];
};

export type WorkoutFormValues = {
  date: Date;
  notes: string;
  exercises: Exercise[];
};

// GET /exercises/:exerciseName/sets
export interface ExerciseSet {
  id: number;
  exercise_id: number;
  workout_id: number;
  weight: number;
  reps: number;
  set_type: string;
  created_at: string;
  updated_at: string | null;
}