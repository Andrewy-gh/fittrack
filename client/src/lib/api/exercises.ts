import type { ExerciseSet, ExerciseOption } from '@/lib/types';

export async function fetchExerciseSets(exerciseName: string): Promise<ExerciseSet[]> {
  const response = await fetch(`/api/exercises/${encodeURIComponent(exerciseName)}/sets`);
  
  if (!response.ok) {
    throw new Error('Failed to fetch exercise sets');
  }
  
  return response.json();
}

export async function fetchExerciseOptions(): Promise<ExerciseOption[]> {
  const response = await fetch('/api/exercises');
  
  if (!response.ok) {
    throw new Error(`Failed to fetch exercise options: ${response.status} ${response.statusText}`);
  }
  
  return response.json();
}
