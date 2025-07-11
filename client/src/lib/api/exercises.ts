import type { ExerciseWithSets, ExerciseOption } from '@/lib/types';

export async function fetchExerciseWithSets(exerciseId: number): Promise<ExerciseWithSets[]> {
  const response = await fetch(`/api/exercises/${exerciseId}`);

  if (!response.ok) {
    throw new Error('Failed to fetch exercise with sets');
  }

  const data = await response.json();
  return data;
}

export async function fetchExerciseOptions(accessToken: string): Promise<ExerciseOption[]> {
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
