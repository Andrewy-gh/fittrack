import { describe, expect, it } from "vitest";
import type { WorkoutContributionDataResponse } from "@/client";
import {
  buildWorkoutVolumeChartData,
  getWorkoutVolumeBucketLabel,
  getWorkoutVolumeTitle,
  getWorkoutSummary,
} from "@/features/analytics/utils/analytics-workouts";

function contributionDays(
  days: WorkoutContributionDataResponse["days"],
): WorkoutContributionDataResponse["days"] {
  return days;
}

describe("analytics workout helpers", () => {
  it("builds workout volume buckets for focus-filtered chart ranges", () => {
    const days = contributionDays([
      {
        date: "2026-03-21",
        count: 3,
        level: 1,
        workouts: [
          {
            id: 1,
            focus: "Push",
            time: "2026-03-21T08:00:00.000Z",
            volume: 1800,
          },
        ],
      },
      {
        date: "2026-03-22",
        count: 4,
        level: 1,
        workouts: [
          {
            id: 2,
            focus: "Pull",
            time: "2026-03-22T08:00:00.000Z",
            volume: 2200,
          },
        ],
      },
      {
        date: "2026-03-23",
        count: 5,
        level: 1,
        workouts: [
          {
            id: 3,
            focus: "Push",
            time: "2026-03-23T08:00:00.000Z",
            volume: 2600,
          },
        ],
      },
    ]);

    expect(
      buildWorkoutVolumeChartData(
        days,
        "W",
        "Push",
        new Date("2026-03-23T12:00:00.000Z"),
      ),
    ).toEqual([
      {
        x: "2026-03-21",
        date: "2026-03-21",
        focusType: "Push",
        value: 1800,
      },
      {
        x: "2026-03-23",
        date: "2026-03-23",
        focusType: "Push",
        value: 2600,
      },
    ]);
  });

  it("adds focus types to all-focus daily workout volume buckets", () => {
    const days = contributionDays([
      {
        date: "2026-03-21",
        count: 7,
        level: 2,
        workouts: [
          {
            id: 1,
            focus: "Push",
            time: "2026-03-21T08:00:00.000Z",
            volume: 1800,
          },
          {
            id: 2,
            focus: "Pull",
            time: "2026-03-21T18:00:00.000Z",
            volume: 2200,
          },
        ],
      },
      {
        date: "2026-03-23",
        count: 5,
        level: 1,
        workouts: [
          {
            id: 3,
            focus: "Push",
            time: "2026-03-23T08:00:00.000Z",
            volume: 2600,
          },
        ],
      },
    ]);

    expect(
      buildWorkoutVolumeChartData(
        days,
        "W",
        undefined,
        new Date("2026-03-23T12:00:00.000Z"),
      ),
    ).toEqual([
      {
        x: "2026-03-21",
        date: "2026-03-21",
        focusType: "Pull, Push",
        value: 4000,
      },
      {
        x: "2026-03-23",
        date: "2026-03-23",
        focusType: "Push",
        value: 2600,
      },
    ]);
  });

  it("adds focus types to all-focus weekly and monthly volume buckets", () => {
    const days = contributionDays([
      {
        date: "2026-03-16",
        count: 3,
        level: 1,
        workouts: [
          {
            id: 1,
            focus: "Push",
            time: "2026-03-16T08:00:00.000Z",
            volume: 1800,
          },
        ],
      },
      {
        date: "2026-03-17",
        count: 4,
        level: 1,
        workouts: [
          {
            id: 2,
            focus: "Pull",
            time: "2026-03-17T18:00:00.000Z",
            volume: 2200,
          },
        ],
      },
      {
        date: "2026-04-01",
        count: 5,
        level: 1,
        workouts: [
          {
            id: 3,
            focus: "Legs",
            time: "2026-04-01T08:00:00.000Z",
            volume: 2600,
          },
        ],
      },
    ]);

    expect(
      buildWorkoutVolumeChartData(
        days,
        "6M",
        undefined,
        new Date("2026-04-05T12:00:00.000Z"),
      ).filter((point) => point.value > 0),
    ).toEqual([
      {
        x: "2026-03-16",
        date: "2026-03-16",
        focusType: "Pull, Push",
        value: 4000,
      },
      {
        x: "2026-03-30",
        date: "2026-03-30",
        focusType: "Legs",
        value: 2600,
      },
    ]);

    expect(
      buildWorkoutVolumeChartData(
        days,
        "Y",
        undefined,
        new Date("2026-04-05T12:00:00.000Z"),
      ).filter((point) => point.value > 0),
    ).toEqual([
      {
        x: "2026-03-01",
        date: "2026-03-01",
        focusType: "Pull, Push",
        value: 4000,
      },
      {
        x: "2026-04-01",
        date: "2026-04-01",
        focusType: "Legs",
        value: 2600,
      },
    ]);
  });

  it("returns explicit bucket labels for each workout volume range", () => {
    expect(getWorkoutVolumeBucketLabel("W")).toBe(
      "Daily bars for the last 7 days",
    );
    expect(getWorkoutVolumeBucketLabel("M")).toBe(
      "Daily bars for the last 30 days",
    );
    expect(getWorkoutVolumeBucketLabel("6M")).toBe(
      "Weekly bars for the last 26 weeks",
    );
    expect(getWorkoutVolumeBucketLabel("Y")).toBe(
      "Monthly bars for the last 12 months",
    );
  });

  it("returns range-aware workout volume titles", () => {
    expect(getWorkoutVolumeTitle("W")).toBe("Daily Volume");
    expect(getWorkoutVolumeTitle("M", "Push")).toBe("Daily Push Volume");
    expect(getWorkoutVolumeTitle("6M")).toBe("Weekly Volume");
    expect(getWorkoutVolumeTitle("Y", "Pull")).toBe("Monthly Pull Volume");
  });

  it("computes analytics workout summaries from workout counts when workout metadata is present", () => {
    const summary = getWorkoutSummary(
      [
        {
          date: "2026-03-10",
          count: 9,
          level: 2,
          workouts: [
            {
              id: 1,
              time: "2026-03-10T08:00:00.000Z",
              volume: 1200,
            },
            {
              id: 2,
              time: "2026-03-10T17:00:00.000Z",
              volume: 1400,
            },
          ],
        },
        {
          date: "2026-03-11",
          count: 4,
          level: 1,
          workouts: [{ id: 3, time: "2026-03-11T08:00:00.000Z", volume: 1800 }],
        },
      ],
      new Date("2026-03-11T12:00:00.000Z"),
    );

    expect(summary).toEqual({
      totalWorkouts30d: 3,
      avgWorkoutsPerWeek: 0.7,
      currentStreak: 2,
      longestStreak: 2,
    });
  });

  it("computes analytics workout summaries with streaks", () => {
    const summary = getWorkoutSummary(
      [
        { date: "2026-03-05", count: 1 },
        { date: "2026-03-08", count: 1 },
        { date: "2026-03-10", count: 1 },
        { date: "2026-03-11", count: 2 },
        { date: "2026-03-12", count: 1 },
      ],
      new Date("2026-03-12T12:00:00.000Z"),
    );

    expect(summary).toEqual({
      totalWorkouts30d: 6,
      avgWorkoutsPerWeek: 1.4,
      currentStreak: 3,
      longestStreak: 3,
    });
  });
});
