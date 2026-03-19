import { arrayMove } from '@dnd-kit/sortable';
import { useEffect, useMemo, useRef, useState } from 'react';

export type ReorderableExercise<T> = {
  id: string;
  exercise: T;
};

function haveSameExerciseCollection<T extends object>(
  entries: ReorderableExercise<T>[],
  exercises: T[]
) {
  if (entries.length !== exercises.length) {
    return false;
  }

  const exerciseSet = new Set(exercises);
  return entries.every((entry) => exerciseSet.has(entry.exercise));
}

function areExercisesInSameOrder<T extends object>(
  entries: ReorderableExercise<T>[],
  exercises: T[]
) {
  return (
    entries.length === exercises.length &&
    entries.every((entry, index) => entry.exercise === exercises[index])
  );
}

export function useExerciseReorder<T extends object>(exercises: T[]) {
  const exerciseIdsRef = useRef(new WeakMap<T, string>());
  const nextIdRef = useRef(0);
  const [isReorderMode, setIsReorderMode] = useState(false);

  const createEntries = (items: T[]): ReorderableExercise<T>[] =>
    items.map((exercise) => {
      let id = exerciseIdsRef.current.get(exercise);

      if (!id) {
        id = `exercise-${nextIdRef.current++}`;
        exerciseIdsRef.current.set(exercise, id);
      }

      return { id, exercise };
    });

  const [draftEntries, setDraftEntries] = useState(() => createEntries(exercises));
  const canReorder = exercises.length > 1;

  useEffect(() => {
    if (isReorderMode) {
      return;
    }

    setDraftEntries(createEntries(exercises));
  }, [exercises, isReorderMode]);

  useEffect(() => {
    if (!isReorderMode) {
      return;
    }

    if (haveSameExerciseCollection(draftEntries, exercises)) {
      return;
    }

    setIsReorderMode(false);
    setDraftEntries(createEntries(exercises));
  }, [draftEntries, exercises, isReorderMode]);

  useEffect(() => {
    if (canReorder) {
      return;
    }

    setIsReorderMode(false);
    setDraftEntries(createEntries(exercises));
  }, [canReorder, exercises]);

  const displayEntries = isReorderMode ? draftEntries : createEntries(exercises);
  const hasPendingOrderChanges = useMemo(
    () => !areExercisesInSameOrder(draftEntries, exercises),
    [draftEntries, exercises]
  );

  const startReorder = () => {
    if (!canReorder) {
      return;
    }

    setDraftEntries(createEntries(exercises));
    setIsReorderMode(true);
  };

  const cancelReorder = () => {
    setDraftEntries(createEntries(exercises));
    setIsReorderMode(false);
  };

  const moveExercise = (activeId: string, overId: string) => {
    if (activeId === overId) {
      return;
    }

    setDraftEntries((currentEntries) => {
      const oldIndex = currentEntries.findIndex((entry) => entry.id === activeId);
      const newIndex = currentEntries.findIndex((entry) => entry.id === overId);

      if (oldIndex < 0 || newIndex < 0) {
        return currentEntries;
      }

      return arrayMove(currentEntries, oldIndex, newIndex);
    });
  };

  const commitReorder = () => {
    setIsReorderMode(false);
    return draftEntries.map(({ exercise }) => exercise);
  };

  return {
    canReorder,
    commitReorder,
    displayEntries,
    hasPendingOrderChanges,
    isReorderMode,
    cancelReorder,
    moveExercise,
    startReorder,
  };
}
