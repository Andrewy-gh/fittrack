import { act, renderHook } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { useExerciseReorder } from '../use-exercise-reorder';

type TestExercise = {
  name: string;
  sets: Array<{ reps: number; weight?: number }>;
};

function createExercise(name: string): TestExercise {
  return {
    name,
    sets: [],
  };
}

describe('useExerciseReorder', () => {
  it('keeps an in-progress reorder draft across ordinary rerenders', () => {
    const initialExercises = [
      createExercise('Bench Press'),
      createExercise('Barbell Squat'),
    ];
    const { result, rerender } = renderHook(
      ({ exercises }: { exercises: TestExercise[] }) =>
        useExerciseReorder(exercises),
      {
        initialProps: { exercises: initialExercises },
      }
    );

    act(() => {
      result.current.startReorder();
    });

    const [firstId, secondId] = result.current.displayEntries.map(
      (entry) => entry.id
    );

    act(() => {
      result.current.moveExercise(secondId, firstId);
    });

    rerender({ exercises: [...initialExercises] });

    expect(result.current.isReorderMode).toBe(true);
    expect(
      result.current.displayEntries.map((entry) => entry.exercise.name)
    ).toEqual(['Barbell Squat', 'Bench Press']);
  });

  it('resets the draft when exercises are externally replaced during reorder mode', () => {
    const initialExercises = [
      createExercise('Bench Press'),
      createExercise('Barbell Squat'),
    ];
    const replacementExercises = [
      createExercise('Deadlift'),
      createExercise('Overhead Press'),
    ];
    const { result, rerender } = renderHook(
      ({ exercises }: { exercises: TestExercise[] }) =>
        useExerciseReorder(exercises),
      {
        initialProps: { exercises: initialExercises },
      }
    );

    act(() => {
      result.current.startReorder();
    });

    const [firstId, secondId] = result.current.displayEntries.map(
      (entry) => entry.id
    );

    act(() => {
      result.current.moveExercise(secondId, firstId);
    });

    rerender({ exercises: replacementExercises });

    expect(result.current.isReorderMode).toBe(false);
    expect(result.current.hasPendingOrderChanges).toBe(false);
    expect(
      result.current.displayEntries.map((entry) => entry.exercise.name)
    ).toEqual(['Deadlift', 'Overhead Press']);
  });
});
