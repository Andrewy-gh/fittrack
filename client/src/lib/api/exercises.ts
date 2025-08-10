import { queryOptions } from '@tanstack/react-query';
import { ExercisesService } from '../../generated';
import type { 
  exercise_ExerciseResponse,
  exercise_ExerciseWithSetsResponse,
  exercise_CreateExerciseRequest,
  exercise_CreateExerciseResponse
} from '../../generated';

// Type aliases for better compatibility with existing code
export type ExerciseWithSets = exercise_ExerciseWithSetsResponse;
export type ExerciseOption = exercise_ExerciseResponse;
export type CreateExerciseRequest = exercise_CreateExerciseRequest;
export type CreateExerciseResponse = exercise_CreateExerciseResponse;

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

// Delegated functions using generated service
export const getExercises = () => ExercisesService.getExercises();
export const createExercise = (data: CreateExerciseRequest) =>
  ExercisesService.postExercises(data);
export const getExerciseWithSets = (id: number) =>
  ExercisesService.getExercises1(id);

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
