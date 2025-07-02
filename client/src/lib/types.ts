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

// GET /exercises/$exerciseId
export interface ExerciseWithSets {
  workout_id: number;
  workout_date: string;
  workout_notes: string | null;
  set_id: number;
  weight: number;
  reps: number;
  set_type: string;
  exercise_id: number;
  exercise_name: string;
}