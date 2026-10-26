import { describe, expect, it } from "vitest";
import {
  buildWorkoutDraftFromHistory,
  formatWeekComparison,
  getWorkoutConsistencySummary,
} from "./workout-insights";

describe("workout-insights", () => {
  it("builds a new workout draft from existing workout structure", () => {
    const draft = buildWorkoutDraftFromHistory(
      [
        {
          workout_id: 10,
          workout_date: "2026-03-03T10:00:00.000Z",
          workout_focus: "Upper",
          exercise_id: 2,
          exercise_name: "Incline Bench",
          exercise_order: 2,
          set_id: 20,
          set_order: 1,
          reps: 8,
          weight: 155,
          set_type: "working",
          volume: 1240,
        },
        {
          workout_id: 10,
          workout_date: "2026-03-03T10:00:00.000Z",
          workout_focus: "Upper",
          exercise_id: 1,
          exercise_name: "Bench Press",
          exercise_order: 1,
          set_id: 10,
          set_order: 1,
          reps: 5,
          weight: 205,
          set_type: "working",
          volume: 1025,
        },
      ],
      new Date("2026-03-13T08:00:00.000Z"),
    );

    expect(draft).toEqual({
      date: "2026-03-13T08:00:00.000Z",
      notes: undefined,
      workoutFocus: "Upper",
      exercises: [
        {
          name: "Bench Press",
          sets: [{ reps: 5, weight: 205, setType: "working" }],
        },
        {
          name: "Incline Bench",
          sets: [{ reps: 8, weight: 155, setType: "working" }],
        },
      ],
    });
  });
  it("computes supportive consistency summaries", () => {
    const summary = getWorkoutConsistencySummary(
      [
        { date: "2026-03-02T12:00:00.000Z" },
        { date: "2026-03-03T12:00:00.000Z" },
        { date: "2026-03-10T12:00:00.000Z" },
        { date: "2026-03-12T12:00:00.000Z" },
      ],
      new Date("2026-03-13T12:00:00.000Z"),
    );

    expect(summary).toEqual({
      totalWorkouts: 4,
      workoutsThisWeek: 2,
      workoutsLastWeek: 2,
      activeDaysThisMonth: 4,
      averageWorkoutsPerWeek: 0.5,
    });
  });

  it("formats a neutral week comparison message", () => {
    expect(formatWeekComparison(3, 1)).toBe("Up 2 from last week");
    expect(formatWeekComparison(1, 3)).toBe("2 fewer than last week");
    expect(formatWeekComparison(0, 0)).toBe("No workouts yet this week");
  });
});
