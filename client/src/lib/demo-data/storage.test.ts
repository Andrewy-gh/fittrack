import { afterEach, describe, expect, it, vi } from "vitest";
import {
  updateExercise,
  createExercise,
  getExerciseById,
  getExerciseDetail,
  getAllExercises,
  getAllWorkouts,
  getWorkoutById,
} from "./storage";
import { recomputeDemoExerciseHistorical1Rm } from "./historical-1rm";
import { STORAGE_KEYS } from "./types";

afterEach(() => {
  vi.useRealTimers();
  localStorage.clear();
});

describe("updateExercise", () => {
  it("should successfully update exercise name", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-03-15T09:00:00.000Z"));
    const exercise = createExercise("Bench Press");
    vi.setSystemTime(new Date("2026-03-15T09:00:01.000Z"));
    const success = updateExercise(exercise.id, "Incline Bench Press");

    expect(success).toBe(true);

    const exercises = getAllExercises();
    const updated = exercises.find((ex) => ex.id === exercise.id);

    expect(updated).toBeDefined();
    expect(updated?.name).toBe("Incline Bench Press");
    expect(updated?.updated_at).not.toBe(exercise.updated_at);
  });

  it("should throw error when duplicate name exists", () => {
    const exercise1 = createExercise("Squat");
    createExercise("Deadlift");

    expect(() => {
      updateExercise(exercise1.id, "Deadlift");
    }).toThrow('Exercise name "Deadlift" already exists');
  });

  it("should throw error when duplicate name exists (case-insensitive)", () => {
    const exercise1 = createExercise("Squat");
    createExercise("Deadlift");

    expect(() => {
      updateExercise(exercise1.id, "deadlift");
    }).toThrow('Exercise name "deadlift" already exists');
  });

  it("should return false when exercise not found", () => {
    const success = updateExercise(999, "Non-existent Exercise");
    expect(success).toBe(false);
  });

  it("should allow updating to the same name", () => {
    const exercise = createExercise("Bench Press");

    const success = updateExercise(exercise.id, "Bench Press");
    expect(success).toBe(true);

    const exercises = getAllExercises();
    const updated = exercises.find((ex) => ex.id === exercise.id);
    expect(updated?.name).toBe("Bench Press");
  });

  it("should update updated_at timestamp", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-03-15T10:00:00.000Z"));
    const exercise = createExercise("Bench Press");
    const oldUpdatedAt = exercise.updated_at;
    vi.setSystemTime(new Date("2026-03-15T10:00:01.000Z"));
    updateExercise(exercise.id, "New Name");

    const exercises = getAllExercises();
    const updated = exercises.find((ex) => ex.id === exercise.id);

    expect(updated?.updated_at).not.toBe(oldUpdatedAt);
  });

  it("reads valid persisted demo collections through storage APIs", () => {
    localStorage.setItem(
      STORAGE_KEYS.EXERCISES,
      JSON.stringify([
        {
          id: 1,
          name: "Bench Press",
          user_id: "demo-user",
          created_at: "2026-03-15T10:00:00.000Z",
          updated_at: "2026-03-15T10:00:00.000Z",
        },
      ]),
    );
    localStorage.setItem(
      STORAGE_KEYS.WORKOUTS,
      JSON.stringify([
        {
          id: 2,
          date: "2026-03-16",
          notes: "Heavy triples",
          workout_focus: "Upper",
          user_id: "demo-user",
          created_at: "2026-03-16T10:00:00.000Z",
          updated_at: "2026-03-16T10:00:00.000Z",
        },
      ]),
    );
    localStorage.setItem(
      STORAGE_KEYS.SETS,
      JSON.stringify([
        {
          id: 3,
          exercise_id: 1,
          workout_id: 2,
          weight: 185,
          reps: 3,
          set_type: "working",
          exercise_order: 0,
          set_order: 0,
          user_id: "demo-user",
          created_at: "2026-03-16T10:00:00.000Z",
        },
      ]),
    );

    expect(getAllExercises()).toEqual([
      {
        id: 1,
        name: "Bench Press",
        user_id: "demo-user",
        created_at: "2026-03-15T10:00:00.000Z",
        updated_at: "2026-03-15T10:00:00.000Z",
      },
    ]);
    expect(getAllWorkouts()).toEqual([
      {
        id: 2,
        date: "2026-03-16",
        notes: "Heavy triples",
        workout_focus: "Upper",
        user_id: "demo-user",
        created_at: "2026-03-16T10:00:00.000Z",
        updated_at: "2026-03-16T10:00:00.000Z",
      },
    ]);
    expect(getWorkoutById(2)).toEqual([
      {
        workout_id: 2,
        workout_date: "2026-03-16",
        workout_notes: "Heavy triples",
        workout_focus: "Upper",
        set_id: 3,
        exercise_id: 1,
        exercise_name: "Bench Press",
        exercise_order: 0,
        set_order: 0,
        reps: 3,
        weight: 185,
        set_type: "working",
        volume: 555,
      },
    ]);
  });

  it("defaults corrupt demo storage collections", () => {
    localStorage.setItem(
      STORAGE_KEYS.EXERCISES,
      JSON.stringify([{ id: 1, name: 123 }]),
    );
    localStorage.setItem(
      STORAGE_KEYS.WORKOUTS,
      JSON.stringify([{ id: "2", date: "2026-03-16" }]),
    );
    localStorage.setItem(
      STORAGE_KEYS.SETS,
      JSON.stringify([{ id: 3, set_type: "drop" }]),
    );

    expect(getAllExercises()).toEqual([]);
    expect(getAllWorkouts()).toEqual([]);
    expect(getWorkoutById(2)).toEqual([]);
  });

  it("defaults malformed historical 1rm map storage when reading exercise detail", () => {
    localStorage.setItem(
      STORAGE_KEYS.EXERCISES,
      JSON.stringify([
        {
          id: 1,
          name: "Bench Press",
          user_id: "demo-user",
          created_at: "2026-03-15T10:00:00.000Z",
          updated_at: "2026-03-15T10:00:00.000Z",
        },
      ]),
    );
    localStorage.setItem(STORAGE_KEYS.HISTORICAL_1RM, JSON.stringify(null));

    expect(getExerciseById(1)?.name).toBe("Bench Press");
    expect(getExerciseDetail(1).exercise.historical_1rm).toBeUndefined();
  });

  it("defaults malformed demo set storage when recomputing historical 1rm", () => {
    localStorage.setItem(
      STORAGE_KEYS.SETS,
      JSON.stringify({
        id: 3,
        exercise_id: 1,
        workout_id: 2,
        weight: 185,
        reps: 3,
        set_type: "working",
      }),
    );

    expect(recomputeDemoExerciseHistorical1Rm(1)).toBeNull();
  });
});
