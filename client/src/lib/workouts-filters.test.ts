import { describe, it, expect } from "vitest";
import type { WorkoutWorkoutResponse } from "@/client";
import {
  filterWorkoutsByFocus,
  getFocusAreas,
  paginateWorkouts,
  sortWorkoutsByCreatedAt,
} from "./workouts-filters";

const baseWorkout = {
  notes: undefined,
  updated_at: "2025-01-02T10:00:00Z",
  user_id: "user-1",
};

const workouts: WorkoutWorkoutResponse[] = [
  {
    id: 1,
    date: "2025-01-02T00:00:00Z",
    created_at: "2025-01-02T10:00:00Z",
    workout_focus: "Push",
    ...baseWorkout,
  },
  {
    id: 2,
    date: "2025-01-01T00:00:00Z",
    created_at: "2025-01-01T10:00:00Z",
    workout_focus: "Pull",
    ...baseWorkout,
  },
  {
    id: 3,
    date: "2025-01-03T00:00:00Z",
    created_at: "2025-01-03T10:00:00Z",
    workout_focus: undefined,
    ...baseWorkout,
  },
  {
    id: 4,
    date: "2025-01-04T00:00:00Z",
    created_at: "2025-01-04T10:00:00Z",
    workout_focus: "Push",
    ...baseWorkout,
  },
];

describe("workouts-filters", () => {
  it("returns unique sorted focus areas", () => {
    expect(getFocusAreas(workouts)).toEqual(["Pull", "Push"]);
  });

  it("filters, sorts, and paginates workouts", () => {
    const filtered = filterWorkoutsByFocus(workouts, "Push");
    const sorted = sortWorkoutsByCreatedAt(filtered, "asc");
    const result = paginateWorkouts(sorted, 1, 2);

    expect(sorted.map((workout) => workout.id)).toEqual([1, 4]);
    expect(result.totalPages).toBe(2);
    expect(result.currentPage).toBe(2);
    expect(result.pagedWorkouts.map((workout) => workout.id)).toEqual([4]);
  });
});
