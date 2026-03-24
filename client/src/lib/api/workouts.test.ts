import { beforeEach, describe, expect, it, vi } from 'vitest';

const {
  invalidateQueries,
  removeQueries,
  useMutation,
} = vi.hoisted(() => ({
  invalidateQueries: vi.fn(),
  removeQueries: vi.fn(),
  useMutation: vi.fn((options) => options),
}));

vi.mock('@tanstack/react-query', () => ({
  useMutation,
}));

vi.mock('./api', () => ({
  queryClient: {
    invalidateQueries,
    removeQueries,
  },
}));

vi.mock('@/client/@tanstack/react-query.gen', () => ({
  getExercisesQueryKey: vi.fn(() => ['exercises']),
  getWorkoutsByIdQueryKey: vi.fn(({ path: { id } }) => ['workout', id]),
  getWorkoutsByIdQueryOptions: vi.fn(),
  getWorkoutsContributionDataQueryKey: vi.fn(() => ['contribution']),
  getWorkoutsContributionDataQueryOptions: vi.fn(),
  getWorkoutsFocusValuesQueryKey: vi.fn(() => ['focus-values']),
  getWorkoutsFocusValuesQueryOptions: vi.fn(),
  getWorkoutsQueryKey: vi.fn(() => ['workouts']),
  getWorkoutsQueryOptions: vi.fn(),
  postWorkoutsMutation: vi.fn(() => ({ mutationFn: vi.fn() })),
  putWorkoutsByIdMutation: vi.fn(() => ({ mutationFn: vi.fn() })),
  deleteWorkoutsByIdMutation: vi.fn(() => ({ mutationFn: vi.fn() })),
}));

import {
  useDeleteWorkoutMutation,
  useSaveWorkoutMutation,
  useUpdateWorkoutMutation,
} from './workouts';

describe('workout mutation cache invalidation', () => {
  beforeEach(() => {
    invalidateQueries.mockClear();
    removeQueries.mockClear();
    useMutation.mockClear();
  });

  it('invalidates analytics queries after creating a workout', () => {
    const mutation = useSaveWorkoutMutation() as any;

    mutation.onSuccess?.(undefined, undefined, undefined);

    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ['workouts'] });
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ['exercises'] });
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ['contribution'] });
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ['focus-values'] });
  });

  it('invalidates analytics queries after updating a workout', () => {
    const mutation = useUpdateWorkoutMutation() as any;

    mutation.onSuccess?.(undefined, { path: { id: 42 } }, undefined);

    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ['workouts'] });
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ['workout', 42] });
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ['contribution'] });
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ['focus-values'] });
  });

  it('invalidates analytics queries after deleting a workout', () => {
    const mutation = useDeleteWorkoutMutation() as any;

    mutation.onSuccess?.(undefined, { path: { id: 42 } }, undefined);

    expect(removeQueries).toHaveBeenCalledWith({ queryKey: ['workouts'] });
    expect(removeQueries).toHaveBeenCalledWith({ queryKey: ['workout', 42] });
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ['workouts'] });
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ['contribution'] });
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: ['focus-values'] });
  });
});
