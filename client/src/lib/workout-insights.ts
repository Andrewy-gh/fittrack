import {
  endOfMonth,
  endOfWeek,
  isWithinInterval,
  startOfMonth,
  startOfWeek,
  subWeeks,
} from 'date-fns';
import type {
  ExerciseExerciseWithSetsResponse,
  WorkoutCreateWorkoutRequest,
  WorkoutWorkoutResponse,
  WorkoutWorkoutWithSetsResponse,
} from '@/client';
import { sortByExerciseAndSetOrder } from './utils';

export type WorkoutNoteContext = {
  date: string;
  note: string;
  workoutId: number;
};

export type WorkoutConsistencySummary = {
  totalWorkouts: number;
  workoutsThisWeek: number;
  workoutsLastWeek: number;
  activeDaysThisMonth: number;
  averageWorkoutsPerWeek: number;
};

function trimNote(note?: string | null): string | null {
  const trimmed = note?.trim();
  return trimmed ? trimmed : null;
}

export function buildWorkoutDraftFromHistory(
  workout: WorkoutWorkoutWithSetsResponse[],
  now = new Date()
): WorkoutCreateWorkoutRequest {
  const sortedWorkout = sortByExerciseAndSetOrder(workout);
  const groupedExercises = new Map<
    number,
    {
      name: string;
      order: number;
      sets: WorkoutCreateWorkoutRequest['exercises'][number]['sets'];
    }
  >();

  for (const set of sortedWorkout) {
    const exerciseId = set.exercise_id;
    const exerciseOrder = set.exercise_order ?? set.exercise_id ?? 0;

    if (!groupedExercises.has(exerciseId)) {
      groupedExercises.set(exerciseId, {
        name: set.exercise_name,
        order: exerciseOrder,
        sets: [],
      });
    }

    groupedExercises.get(exerciseId)?.sets.push({
      reps: set.reps,
      weight: set.weight,
      setType:
        set.set_type === 'warmup' || set.set_type === 'working'
          ? set.set_type
          : 'working',
    });
  }

  return {
    date: now.toISOString(),
    notes: undefined,
    workoutFocus: sortedWorkout[0]?.workout_focus || undefined,
    exercises: Array.from(groupedExercises.values())
      .sort((left, right) => left.order - right.order)
      .map(({ name, sets }) => ({ name, sets })),
  };
}

export function getLatestWorkoutNote(
  workouts: Pick<WorkoutWorkoutResponse, 'date' | 'id' | 'notes'>[]
): WorkoutNoteContext | null {
  const latestWorkoutWithNote = [...workouts]
    .filter((workout) => trimNote(workout.notes))
    .sort(
      (left, right) =>
        new Date(right.date).getTime() - new Date(left.date).getTime()
    )[0];

  if (!latestWorkoutWithNote) {
    return null;
  }

  return {
    workoutId: latestWorkoutWithNote.id,
    date: latestWorkoutWithNote.date,
    note: trimNote(latestWorkoutWithNote.notes)!,
  };
}

export function getLatestExerciseNote(
  exerciseSets: Pick<
    ExerciseExerciseWithSetsResponse,
    'workout_date' | 'workout_id' | 'workout_notes'
  >[]
): WorkoutNoteContext | null {
  const latestExerciseSetWithNote = [...exerciseSets]
    .filter((set) => trimNote(set.workout_notes))
    .sort(
      (left, right) =>
        new Date(right.workout_date).getTime() -
        new Date(left.workout_date).getTime()
    )[0];

  if (!latestExerciseSetWithNote) {
    return null;
  }

  return {
    workoutId: latestExerciseSetWithNote.workout_id,
    date: latestExerciseSetWithNote.workout_date,
    note: trimNote(latestExerciseSetWithNote.workout_notes)!,
  };
}

export function getWorkoutConsistencySummary(
  workouts: Pick<WorkoutWorkoutResponse, 'date'>[],
  now = new Date()
): WorkoutConsistencySummary {
  const currentWeekInterval = {
    start: startOfWeek(now, { weekStartsOn: 1 }),
    end: endOfWeek(now, { weekStartsOn: 1 }),
  };
  const previousWeekReference = subWeeks(now, 1);
  const previousWeekInterval = {
    start: startOfWeek(previousWeekReference, { weekStartsOn: 1 }),
    end: endOfWeek(previousWeekReference, { weekStartsOn: 1 }),
  };
  const currentMonthInterval = {
    start: startOfMonth(now),
    end: endOfMonth(now),
  };
  const rollingAverageStart = startOfWeek(subWeeks(now, 7), {
    weekStartsOn: 1,
  });

  const workoutsThisWeek = workouts.filter((workout) =>
    isWithinInterval(new Date(workout.date), currentWeekInterval)
  ).length;
  const workoutsLastWeek = workouts.filter((workout) =>
    isWithinInterval(new Date(workout.date), previousWeekInterval)
  ).length;
  const activeDaysThisMonth = new Set(
    workouts
      .filter((workout) =>
        isWithinInterval(new Date(workout.date), currentMonthInterval)
      )
      .map((workout) => workout.date.slice(0, 10))
  ).size;
  const workoutsInRollingWindow = workouts.filter(
    (workout) => new Date(workout.date) >= rollingAverageStart
  ).length;

  return {
    totalWorkouts: workouts.length,
    workoutsThisWeek,
    workoutsLastWeek,
    activeDaysThisMonth,
    averageWorkoutsPerWeek:
      Math.round((workoutsInRollingWindow / 8) * 10) / 10,
  };
}

export function formatWeekComparison(
  workoutsThisWeek: number,
  workoutsLastWeek: number
): string {
  if (workoutsLastWeek === 0) {
    return workoutsThisWeek === 0 ? 'No workouts yet this week' : 'First week back';
  }

  const difference = workoutsThisWeek - workoutsLastWeek;

  if (difference === 0) {
    return `Same as last week (${workoutsLastWeek})`;
  }

  if (difference > 0) {
    return `Up ${difference} from last week`;
  }

  return `${Math.abs(difference)} fewer than last week`;
}
