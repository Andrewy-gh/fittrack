import { describe, expect, it } from "vitest";
import {
  buildDemoContributionData,
  type DemoContributionWorkout,
} from "@/lib/demo-data/contribution-data";

function workout(
  overrides: Partial<DemoContributionWorkout>,
): DemoContributionWorkout {
  return {
    id: 0,
    date: "",
    workout_focus: null,
    volume: 0,
    workingSetCount: 0,
    ...overrides,
  };
}

describe("demo contribution data", () => {
  it("builds sorted contribution data from demo workouts", () => {
    const data = buildDemoContributionData([
      workout({
        id: 2,
        date: "2026-03-12T08:00:00.000Z",
        workout_focus: "Upper",
        volume: 3600,
        workingSetCount: 3,
      }),
      workout({
        id: 1,
        date: "2026-03-10T07:00:00.000Z",
        workout_focus: "Lower",
        volume: 2800,
        workingSetCount: 2,
      }),
      workout({
        id: 3,
        date: "2026-03-12T17:00:00.000Z",
        workout_focus: "Conditioning",
        volume: 1800,
        workingSetCount: 4,
      }),
    ]);

    expect(data).toEqual({
      days: [
        {
          date: "2026-03-10",
          count: 2,
          level: 1,
          workouts: [
            {
              id: 1,
              focus: "Lower",
              time: "2026-03-10T07:00:00.000Z",
              volume: 2800,
            },
          ],
        },
        {
          date: "2026-03-12",
          count: 7,
          level: 2,
          workouts: [
            {
              id: 2,
              focus: "Upper",
              time: "2026-03-12T08:00:00.000Z",
              volume: 3600,
            },
            {
              id: 3,
              focus: "Conditioning",
              time: "2026-03-12T17:00:00.000Z",
              volume: 1800,
            },
          ],
        },
      ],
    });
  });
});
