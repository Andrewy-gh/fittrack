import { afterEach, describe, expect, it, vi } from "vitest";
import type { CurrentUser } from "@stackframe/react";
import {
  getExerciseDetailQueryOptions,
  getExerciseListQueryOptions,
  getRecentExerciseSetsQueryOptions,
} from "@/features/exercises/api/exercise-query-options";
import * as apiExercises from "@/features/exercises/api/exercises";
import * as demoQueryOptions from "@/lib/demo-data/query-options";

const exerciseId = 42;

// SAFETY: Query-option selectors only use the user value for its truthiness.
const authenticatedUser = { id: "user-1" } as CurrentUser;

const selectors = [
  {
    name: "exercise list",
    select: (user: CurrentUser | null) => getExerciseListQueryOptions(user),
    expectedArgs: [] as const,
    spyOnApi: () => vi.spyOn(apiExercises, "exercisesQueryOptions"),
    spyOnDemo: () => vi.spyOn(demoQueryOptions, "getDemoExercisesQueryOptions"),
  },
  {
    name: "exercise detail",
    select: (user: CurrentUser | null) =>
      getExerciseDetailQueryOptions(user, exerciseId),
    expectedArgs: [exerciseId] as const,
    spyOnApi: () => vi.spyOn(apiExercises, "exerciseByIdQueryOptions"),
    spyOnDemo: () =>
      vi.spyOn(demoQueryOptions, "getDemoExercisesByIdQueryOptions"),
  },
  {
    name: "recent exercise sets",
    select: (user: CurrentUser | null) =>
      getRecentExerciseSetsQueryOptions(user, exerciseId),
    expectedArgs: [exerciseId] as const,
    spyOnApi: () => vi.spyOn(apiExercises, "recentExerciseSetsQueryOptions"),
    spyOnDemo: () =>
      vi.spyOn(demoQueryOptions, "getDemoExercisesByIdRecentSetsQueryOptions"),
  },
] as const;

const querySources = [
  { name: "authenticated users", user: authenticatedUser, expected: "api" },
  { name: "demo users", user: null, expected: "demo" },
] as const;

describe("exercise query options", () => {
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
