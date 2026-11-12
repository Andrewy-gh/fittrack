import { expect, it } from "vitest";
import { transformToWorkoutFormValues } from "./workouts";
import type { WorkoutWorkoutWithSetsResponse } from "@/client";

it("preserves omitted, zero and positive weights when preparing an edit request", () => {
  const row: WorkoutWorkoutWithSetsResponse = {
    workout_id: 1,
    workout_date: "2026-10-08T12:00:00Z",
    exercise_id: 2,
    exercise_name: "Pushups",
    set_id: 1,
    reps: 10,
    set_type: "working",
    volume: 0,
  };
  const form = transformToWorkoutFormValues([
    row,
    { ...row, set_id: 2, weight: 0 },
    { ...row, set_id: 3, weight: 25 },
  ]);
  expect(form.exercises[0]?.sets.map((set) => set.weight)).toEqual([
    undefined,
    0,
    25,
  ]);
  expect(JSON.parse(JSON.stringify(form)).exercises[0].sets).toEqual([
    { reps: 10, setType: "working" },
    { reps: 10, setType: "working", weight: 0 },
    { reps: 10, setType: "working", weight: 25 },
  ]);
});
