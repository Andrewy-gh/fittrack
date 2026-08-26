import { afterEach, describe, expect, it, vi } from "vitest";
import type { CurrentUser } from "@stackframe/react";
import {
  getNewWorkoutContextQueryOptions,
  getWorkoutContributionQueryOptions,
  getWorkoutByIdQueryOptions,
  getWorkoutListQueryOptions,
  getWorkoutsFocusQueryOptions,
} from "@/features/workouts/api/workout-query-options";
import * as apiWorkouts from "@/features/workouts/api/workouts";
import * as demoQueryOptions from "@/lib/demo-data/query-options";

const workoutId = 42;

// SAFETY: Query-option selectors only use the user value for its truthiness.
const authenticatedUser = { id: "user-1" } as CurrentUser;

const selectors = [
  {
    name: "workout list",
    select: (user: CurrentUser | null) => getWorkoutListQueryOptions(user),
    expectedArgs: [] as const,
    spyOnApi: () => vi.spyOn(apiWorkouts, "workoutsQueryOptions"),
    spyOnDemo: () => vi.spyOn(demoQueryOptions, "getDemoWorkoutsQueryOptions"),
  },
  {
    name: "workout detail",
    select: (user: CurrentUser | null) =>
      getWorkoutByIdQueryOptions(user, workoutId),
    expectedArgs: [workoutId] as const,
    spyOnApi: () => vi.spyOn(apiWorkouts, "workoutQueryOptions"),
    spyOnDemo: () =>
      vi.spyOn(demoQueryOptions, "getDemoWorkoutsByIdQueryOptions"),
  },
  {
    name: "new-workout context",
    select: (user: CurrentUser | null) =>
      getNewWorkoutContextQueryOptions(user),
    expectedArgs: [] as const,
    spyOnApi: () => vi.spyOn(apiWorkouts, "newWorkoutContextQueryOptions"),
    spyOnDemo: () =>
      vi.spyOn(demoQueryOptions, "getDemoNewWorkoutContextQueryOptions"),
  },
  {
    name: "workout focus",
    select: (user: CurrentUser | null) => getWorkoutsFocusQueryOptions(user),
    expectedArgs: [] as const,
    spyOnApi: () => vi.spyOn(apiWorkouts, "workoutsFocusValuesQueryOptions"),
    spyOnDemo: () =>
      vi.spyOn(demoQueryOptions, "getDemoWorkoutsFocusValuesQueryOptions"),
  },
  {
    name: "workout contribution",
    select: (user: CurrentUser | null) =>
      getWorkoutContributionQueryOptions(user),
    expectedArgs: [] as const,
    spyOnApi: () => vi.spyOn(apiWorkouts, "contributionDataQueryOptions"),
    spyOnDemo: () =>
      vi.spyOn(demoQueryOptions, "getDemoContributionDataQueryOptions"),
  },
] as const;

const querySources = [
  { name: "authenticated users", user: authenticatedUser, expected: "api" },
  { name: "demo users", user: null, expected: "demo" },
] as const;

describe("workout query options", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  for (const selector of selectors) {
    describe(selector.name, () => {
      for (const source of querySources) {
        it(`uses ${source.name} query options`, () => {
          const apiSpy = selector.spyOnApi();
          const demoSpy = selector.spyOnDemo();

          const result = selector.select(source.user);
          const expectedSpy = source.expected === "api" ? apiSpy : demoSpy;
          const unusedSpy = source.expected === "api" ? demoSpy : apiSpy;

          expect(expectedSpy).toHaveBeenCalledWith(...selector.expectedArgs);
          expect(unusedSpy).not.toHaveBeenCalled();
          expect(result.queryKey).toEqual(
            expectedSpy.mock.results[0]?.value?.queryKey,
          );
        });
      }
    });
  }
});
