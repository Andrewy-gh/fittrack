import { queryOptions } from '@tanstack/react-query';

// API response type for exercise details
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
  volume: number;
}

export async function fetchExerciseWithSets(
  exerciseId: number,
  accessToken: string
): Promise<ExerciseWithSets[]> {
  const response = await fetch(`/api/exercises/${exerciseId}`, {
    headers: {
      'x-stack-access-token': accessToken,
    },
  });
  if (!response.ok) {
    throw new Error('Failed to fetch exercise sets');
  }
  return response.json();
}

export function exerciseWithSetsQueryOptions(
  exerciseId: number,
  accessToken: string
) {
  return queryOptions<ExerciseWithSets[], Error>({
    queryKey: ['exercises', 'details', exerciseId],
    queryFn: () => fetchExerciseWithSets(exerciseId, accessToken),
  });
}

export interface ExerciseOption {
  id: number;
  name: string;
  created_at: string;
  updated_at: string | null;
}

export type NewExerciseOption = Omit<
  ExerciseOption,
  'created_at' | 'updated_at'
>;

// Form types for workout creation
export interface Set {
  weight?: number;
  reps?: number;
  type?: string;
}

export interface Exercise {
  name: string;
  sets: Set[];
}

export async function fetchExerciseOptions(
  accessToken: string
): Promise<ExerciseOption[]> {
  const response = await fetch('/api/exercises', {
    headers: {
      'x-stack-access-token': accessToken,
    },
  });
  if (!response.ok) {
    throw new Error(
      `Failed to fetch exercise options: ${response.status} ${response.statusText}`
    );
  }
  return response.json();
}

export function exercisesQueryOptions(accessToken: string) {
  return queryOptions<ExerciseOption[], Error>({
    queryKey: ['exercises', 'list'],
    queryFn: () => fetchExerciseOptions(accessToken),
  });
}
