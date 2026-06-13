import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  getExerciseDetailQueryOptions,
  getExerciseListQueryOptions,
  getRecentExerciseSetsQueryOptions,
} from "@/features/exercises/api/exercise-query-options";
import * as apiExercises from "@/features/exercises/api/exercises";
import * as demoQueryOptions from "@/lib/demo-data/query-options";

vi.mock("@/features/exercises/api/exercises");
vi.mock("@/lib/demo-data/query-options");

describe("exercise query options", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("uses API exercise list query options for authenticated users", () => {
    const user = { id: "user-1" } as any;
    const apiOptions = { queryKey: ["exercises"], queryFn: vi.fn() };

    vi.mocked(apiExercises.exercisesQueryOptions).mockReturnValue(
      apiOptions as any,
    );

    const result = getExerciseListQueryOptions(user);

    expect(apiExercises.exercisesQueryOptions).toHaveBeenCalled();
    expect(result).toBe(apiOptions);
  });

  it("uses demo exercise list query options for demo users", () => {
    const demoOptions = { queryKey: ["demo_exercises"], queryFn: vi.fn() };

    vi.mocked(demoQueryOptions.getDemoExercisesQueryOptions).mockReturnValue(
      demoOptions as any,
    );

    const result = getExerciseListQueryOptions(null);

    expect(demoQueryOptions.getDemoExercisesQueryOptions).toHaveBeenCalled();
    expect(result).toBe(demoOptions);
  });

  it("uses API exercise detail query options for authenticated users", () => {
    const user = { id: "user-1" } as any;
    const apiOptions = { queryKey: ["exercise", 42], queryFn: vi.fn() };

    vi.mocked(apiExercises.exerciseByIdQueryOptions).mockReturnValue(
      apiOptions as any,
    );

    const result = getExerciseDetailQueryOptions(user, 42);

    expect(apiExercises.exerciseByIdQueryOptions).toHaveBeenCalledWith(42);
    expect(result).toBe(apiOptions);
  });

  it("uses demo exercise detail query options for demo users", () => {
    const demoOptions = { queryKey: ["demo_exercise", 42], queryFn: vi.fn() };

    vi.mocked(
      demoQueryOptions.getDemoExercisesByIdQueryOptions,
    ).mockReturnValue(demoOptions as any);

    const result = getExerciseDetailQueryOptions(null, 42);

    expect(
      demoQueryOptions.getDemoExercisesByIdQueryOptions,
    ).toHaveBeenCalledWith(42);
    expect(result).toBe(demoOptions);
  });

  it("uses API recent sets query options for authenticated users", () => {
    const user = { id: "user-1" } as any;
    const apiOptions = { queryKey: ["recent_sets", 42], queryFn: vi.fn() };

    vi.mocked(apiExercises.recentExerciseSetsQueryOptions).mockReturnValue(
      apiOptions as any,
    );

    const result = getRecentExerciseSetsQueryOptions(user, 42);

    expect(apiExercises.recentExerciseSetsQueryOptions).toHaveBeenCalledWith(
      42,
    );
    expect(result).toBe(apiOptions);
  });

  it("uses demo recent sets query options for demo users", () => {
    const demoOptions = {
      queryKey: ["demo_recent_sets", 42],
      queryFn: vi.fn(),
    };

    vi.mocked(
      demoQueryOptions.getDemoExercisesByIdRecentSetsQueryOptions,
    ).mockReturnValue(demoOptions as any);

    const result = getRecentExerciseSetsQueryOptions(null, 42);

    expect(
      demoQueryOptions.getDemoExercisesByIdRecentSetsQueryOptions,
    ).toHaveBeenCalledWith(42);
    expect(result).toBe(demoOptions);
  });
});
