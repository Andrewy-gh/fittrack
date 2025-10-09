import { describe, it, expect, vi, beforeEach } from 'vitest';
import {
  getExercisesQueryOptions,
  getRecentSetsQueryOptions,
  getWorkoutsQueryOptions,
  getWorkoutByIdQueryOptions,
  getWorkoutsFocusQueryOptions,
} from '../unified-query-options';
import * as apiExercises from '../exercises';
import * as apiWorkouts from '../workouts';
import * as demoQueryOptions from '@/lib/demo-data/query-options';

// Mock the API modules
vi.mock('../exercises');
vi.mock('../workouts');
vi.mock('@/lib/demo-data/query-options');

describe('Unified Query Options Factory Functions', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('getExercisesQueryOptions', () => {
    it('returns API query options when user is authenticated', () => {
      const mockUser = { id: 'user123' } as any;
      const mockApiOptions = { queryKey: ['exercises'], queryFn: vi.fn() };

      vi.mocked(apiExercises.exercisesQueryOptions).mockReturnValue(mockApiOptions as any);

      const result = getExercisesQueryOptions(mockUser);

      expect(apiExercises.exercisesQueryOptions).toHaveBeenCalled();
      expect(result).toBe(mockApiOptions);
    });

    it('returns demo query options when user is null', () => {
      const mockDemoOptions = { queryKey: ['demo_exercises'], queryFn: vi.fn() };

      vi.mocked(demoQueryOptions.getDemoExercisesQueryOptions).mockReturnValue(mockDemoOptions as any);

      const result = getExercisesQueryOptions(null);

      expect(demoQueryOptions.getDemoExercisesQueryOptions).toHaveBeenCalled();
      expect(result).toBe(mockDemoOptions);
    });
  });

  describe('getRecentSetsQueryOptions', () => {
    const exerciseId = 42;

    it('returns API query options with exerciseId when user is authenticated', () => {
      const mockUser = { id: 'user123' } as any;
      const mockApiOptions = { queryKey: ['recentSets', exerciseId], queryFn: vi.fn() };

      vi.mocked(apiExercises.recentExerciseSetsQueryOptions).mockReturnValue(mockApiOptions as any);

      const result = getRecentSetsQueryOptions(mockUser, exerciseId);

      expect(apiExercises.recentExerciseSetsQueryOptions).toHaveBeenCalledWith(exerciseId);
      expect(result).toBe(mockApiOptions);
    });

    it('returns demo query options with exerciseId when user is null', () => {
      const mockDemoOptions = { queryKey: ['demo_recentSets', exerciseId], queryFn: vi.fn() };

      vi.mocked(demoQueryOptions.getDemoExercisesByIdRecentSetsQueryOptions).mockReturnValue(mockDemoOptions as any);

      const result = getRecentSetsQueryOptions(null, exerciseId);

      expect(demoQueryOptions.getDemoExercisesByIdRecentSetsQueryOptions).toHaveBeenCalledWith(exerciseId);
      expect(result).toBe(mockDemoOptions);
    });
  });

  describe('getWorkoutsQueryOptions', () => {
    it('returns API query options when user is authenticated', () => {
      const mockUser = { id: 'user123' } as any;
      const mockApiOptions = { queryKey: ['workouts'], queryFn: vi.fn() };

      vi.mocked(apiWorkouts.workoutsQueryOptions).mockReturnValue(mockApiOptions as any);

      const result = getWorkoutsQueryOptions(mockUser);

      expect(apiWorkouts.workoutsQueryOptions).toHaveBeenCalled();
      expect(result).toBe(mockApiOptions);
    });

    it('returns demo query options when user is null', () => {
      const mockDemoOptions = { queryKey: ['demo_workouts'], queryFn: vi.fn() };

      vi.mocked(demoQueryOptions.getDemoWorkoutsQueryOptions).mockReturnValue(mockDemoOptions as any);

      const result = getWorkoutsQueryOptions(null);

      expect(demoQueryOptions.getDemoWorkoutsQueryOptions).toHaveBeenCalled();
      expect(result).toBe(mockDemoOptions);
    });
  });

  describe('getWorkoutByIdQueryOptions', () => {
    const workoutId = 123;

    it('returns API query options with workoutId when user is authenticated', () => {
      const mockUser = { id: 'user123' } as any;
      const mockApiOptions = { queryKey: ['workout', workoutId], queryFn: vi.fn() };

      vi.mocked(apiWorkouts.workoutQueryOptions).mockReturnValue(mockApiOptions as any);

      const result = getWorkoutByIdQueryOptions(mockUser, workoutId);

      expect(apiWorkouts.workoutQueryOptions).toHaveBeenCalledWith(workoutId);
      expect(result).toBe(mockApiOptions);
    });

    it('returns demo query options with workoutId when user is null', () => {
      const mockDemoOptions = { queryKey: ['demo_workout', workoutId], queryFn: vi.fn() };

      vi.mocked(demoQueryOptions.getDemoWorkoutsByIdQueryOptions).mockReturnValue(mockDemoOptions as any);

      const result = getWorkoutByIdQueryOptions(null, workoutId);

      expect(demoQueryOptions.getDemoWorkoutsByIdQueryOptions).toHaveBeenCalledWith(workoutId);
      expect(result).toBe(mockDemoOptions);
    });
  });

  describe('getWorkoutsFocusQueryOptions', () => {
    it('returns API query options when user is authenticated', () => {
      const mockUser = { id: 'user123' } as any;
      const mockApiOptions = { queryKey: ['workouts_focus'], queryFn: vi.fn() };

      vi.mocked(apiWorkouts.workoutsFocusValuesQueryOptions).mockReturnValue(mockApiOptions as any);

      const result = getWorkoutsFocusQueryOptions(mockUser);

      expect(apiWorkouts.workoutsFocusValuesQueryOptions).toHaveBeenCalled();
      expect(result).toBe(mockApiOptions);
    });

    it('returns demo query options when user is null', () => {
      const mockDemoOptions = { queryKey: ['demo_workouts_focus'], queryFn: vi.fn() };

      vi.mocked(demoQueryOptions.getDemoWorkoutsFocusValuesQueryOptions).mockReturnValue(mockDemoOptions as any);

      const result = getWorkoutsFocusQueryOptions(null);

      expect(demoQueryOptions.getDemoWorkoutsFocusValuesQueryOptions).toHaveBeenCalled();
      expect(result).toBe(mockDemoOptions);
    });
  });
});
