import { queryOptions } from '@tanstack/react-query';
import { ExercisesService, OpenAPI } from '../../generated';
import type { 
  exercise_ExerciseResponse,
  exercise_ExerciseWithSetsResponse,
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

export function exercisesQueryOptions(accessToken: string) {
  return queryOptions<exercise_ExerciseResponse[], Error>({
    queryKey: ['exercises', 'list'],
    queryFn: async () => {
      OpenAPI.HEADERS = {
        'x-stack-access-token': accessToken,
      };
      return ExercisesService.getExercises();
    },
  });
}

export function exerciseWithSetsQueryOptions(
  exerciseId: number,
  accessToken: string
) {
  return queryOptions<exercise_ExerciseWithSetsResponse[], Error>({
    queryKey: ['exercises', 'details', exerciseId],
    queryFn: async () => {
      OpenAPI.HEADERS = {
        'x-stack-access-token': accessToken,
      };
      return ExercisesService.getExercises1(exerciseId);
    },
  });
}

