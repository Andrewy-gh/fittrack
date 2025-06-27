export interface ExerciseOption {
  id: number;
  name: string;
}

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