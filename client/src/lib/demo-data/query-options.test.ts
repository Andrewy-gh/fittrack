import { beforeEach, describe, expect, it, vi } from 'vitest';

const {
  invalidateQueries,
  removeQueries,
  setQueryData,
} = vi.hoisted(() => ({
  invalidateQueries: vi.fn(),
  removeQueries: vi.fn(),
  setQueryData: vi.fn(),
}));

vi.mock('@tanstack/react-query', () => ({
  queryOptions: vi.fn((options) => options),
}));

vi.mock('../api/api', () => ({
  queryClient: {
    invalidateQueries,
    removeQueries,
    setQueryData,
  },
}));

vi.mock('./storage', () => ({
  getAllExercises: vi.fn(() => []),
  getAllWorkoutsForContribution: vi.fn(() => []),
  createExercise: vi.fn(),
  updateExercise: vi.fn(),
  deleteExercise: vi.fn(),
  getExerciseDetail: vi.fn(),
  getExerciseRecentSets: vi.fn(() => []),
  getAllWorkouts: vi.fn(() => []),
  getWorkoutById: vi.fn((id) => [{ id }]),
  createWorkout: vi.fn(),
  updateWorkout: vi.fn(() => ({ success: true })),
  deleteWorkout: vi.fn(() => true),
  getWorkoutFocusValues: vi.fn(() => []),
  initializeDemoData: vi.fn(),
}));

import {
  deleteDemoWorkoutsByIdMutation,
  postDemoWorkoutsMutation,
  putDemoWorkoutsByIdMutation,
} from './query-options';

const mutationContext = {} as never;

describe('demo workout mutation cache invalidation', () => {
  beforeEach(() => {
    invalidateQueries.mockClear();
    removeQueries.mockClear();
    setQueryData.mockClear();
  });

  it('invalidates analytics queries after creating a demo workout', () => {
    const mutation = postDemoWorkoutsMutation();

    mutation.onSuccess?.(
      { success: true },
      { body: {} as never },
      undefined,
      mutationContext
    );

    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: [{ _id: 'demo_getWorkouts' }],
    });
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: [{ _id: 'demo_getExercises' }],
    });
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: [{ _id: 'demo_getWorkoutsContributionData' }],
    });
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: [{ _id: 'demo_getWorkoutsFocusValues' }],
    });
  });

  it('invalidates analytics queries after updating a demo workout', () => {
    const mutation = putDemoWorkoutsByIdMutation();

    mutation.onSuccess?.(
      undefined,
      { path: { id: 42 }, body: {} as never },
      undefined,
      mutationContext
    );

    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: [{ _id: 'demo_getWorkouts' }],
    });
    expect(setQueryData).toHaveBeenCalledWith(
      [{ _id: 'demo_getWorkoutsById', path: { id: 42 } }],
      [{ id: 42 }]
    );
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: [{ _id: 'demo_getWorkoutsContributionData' }],
    });
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: [{ _id: 'demo_getWorkoutsFocusValues' }],
    });
  });

  it('invalidates analytics queries after deleting a demo workout', () => {
    const mutation = deleteDemoWorkoutsByIdMutation();

    mutation.onSuccess?.(
      undefined,
      { path: { id: 42 } },
      undefined,
      mutationContext
    );

    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: [{ _id: 'demo_getWorkouts' }],
    });
    expect(removeQueries).toHaveBeenCalledWith({
      queryKey: [{ _id: 'demo_getWorkoutsById', path: { id: 42 } }],
    });
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: [{ _id: 'demo_getWorkoutsContributionData' }],
    });
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: [{ _id: 'demo_getWorkoutsFocusValues' }],
    });
  });
});
