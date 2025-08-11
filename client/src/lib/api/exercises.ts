import { queryOptions } from '@tanstack/react-query';
import { ExercisesService, OpenAPI } from '../../generated';
import type { 
  exercise_ExerciseResponse,
  exercise_ExerciseWithSetsResponse,
  exercise_CreateExerciseRequest,
  exercise_CreateExerciseResponse
} from '../../generated';

// Re-export generated types directly
export type { 
  exercise_ExerciseResponse as ExerciseOption,
  exercise_ExerciseWithSetsResponse as ExerciseWithSets,
  exercise_CreateExerciseRequest as CreateExerciseRequest,
  exercise_CreateExerciseResponse as CreateExerciseResponse
} from '../../generated';

// Custom utility types
export type NewExerciseOption = Omit<
  exercise_ExerciseResponse,
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

// Wrapper functions for backward compatibility
export async function fetchExerciseOptions(
  accessToken: string
): Promise<ExerciseOption[]> {
  OpenAPI.HEADERS = {
    'x-stack-access-token': accessToken,
  };
  return ExercisesService.getExercises();
}

export async function fetchExerciseWithSets(
  exerciseId: number,
  accessToken: string
): Promise<ExerciseWithSets[]> {
  OpenAPI.HEADERS = {
    'x-stack-access-token': accessToken,
  };
  return ExercisesService.getExercises1(exerciseId);
}

// Delegated functions using generated service
export const getExercises = () => ExercisesService.getExercises();
export const createExercise = (data: CreateExerciseRequest) =>
  ExercisesService.postExercises(data);
export const getExerciseWithSets = (id: number) =>
  ExercisesService.getExercises1(id);

export function exerciseWithSetsQueryOptions(
  exerciseId: number,
  accessToken: string
) {
  return queryOptions<ExerciseWithSets[], Error>({
    queryKey: ['exercises', 'details', exerciseId],
    queryFn: async () => {
      // Set the custom header for this request
      OpenAPI.HEADERS = {
        'x-stack-access-token': accessToken,
      };
      return ExercisesService.getExercises1(exerciseId);
    },
  });
}

export function exercisesQueryOptions(accessToken: string) {
  return queryOptions<ExerciseOption[], Error>({
    queryKey: ['exercises', 'list'],
    queryFn: async () => {
      // Set the custom header for this request
      OpenAPI.HEADERS = {
        'x-stack-access-token': accessToken,
      };
      return ExercisesService.getExercises();
    },
  });
}
