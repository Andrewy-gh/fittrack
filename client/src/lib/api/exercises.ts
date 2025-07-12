import type { ExerciseWithSets } from '@/lib/types';

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
