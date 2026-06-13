import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  getNewWorkoutContextQueryOptions,
  getWorkoutContributionQueryOptions,
  getWorkoutByIdQueryOptions,
  getWorkoutListQueryOptions,
  getWorkoutsFocusQueryOptions,
} from "@/features/workouts/api/workout-query-options";
import * as apiWorkouts from "@/features/workouts/api/workouts";
import * as demoQueryOptions from "@/lib/demo-data/query-options";

vi.mock("@/features/workouts/api/workouts");
vi.mock("@/lib/demo-data/query-options");

describe("workout query options", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("uses API workout list query options for authenticated users", () => {
    const user = { id: "user-1" } as any;
    const apiOptions = { queryKey: ["workouts"], queryFn: vi.fn() };

    vi.mocked(apiWorkouts.workoutsQueryOptions).mockReturnValue(
      apiOptions as any,
    );

    const result = getWorkoutListQueryOptions(user);

    expect(apiWorkouts.workoutsQueryOptions).toHaveBeenCalled();
    expect(result).toBe(apiOptions);
  });

  it("uses demo workout list query options for demo users", () => {
    const demoOptions = { queryKey: ["demo_workouts"], queryFn: vi.fn() };

    vi.mocked(demoQueryOptions.getDemoWorkoutsQueryOptions).mockReturnValue(
      demoOptions as any,
    );

    const result = getWorkoutListQueryOptions(null);

    expect(demoQueryOptions.getDemoWorkoutsQueryOptions).toHaveBeenCalled();
    expect(result).toBe(demoOptions);
  });

  it("uses API workout detail query options for authenticated users", () => {
    const user = { id: "user-1" } as any;
    const apiOptions = { queryKey: ["workout", 42], queryFn: vi.fn() };

    vi.mocked(apiWorkouts.workoutQueryOptions).mockReturnValue(
      apiOptions as any,
    );

    const result = getWorkoutByIdQueryOptions(user, 42);

    expect(apiWorkouts.workoutQueryOptions).toHaveBeenCalledWith(42);
    expect(result).toBe(apiOptions);
  });

  it("uses demo workout detail query options for demo users", () => {
    const demoOptions = { queryKey: ["demo_workout", 42], queryFn: vi.fn() };

    vi.mocked(demoQueryOptions.getDemoWorkoutsByIdQueryOptions).mockReturnValue(
      demoOptions as any,
    );

    const result = getWorkoutByIdQueryOptions(null, 42);

    expect(
      demoQueryOptions.getDemoWorkoutsByIdQueryOptions,
    ).toHaveBeenCalledWith(42);
    expect(result).toBe(demoOptions);
  });

  it("uses API new-workout context query options for authenticated users", () => {
    const user = { id: "user-1" } as any;
    const apiOptions = { queryKey: ["new_workout_context"], queryFn: vi.fn() };

    vi.mocked(apiWorkouts.newWorkoutContextQueryOptions).mockReturnValue(
      apiOptions as any,
    );

    const result = getNewWorkoutContextQueryOptions(user);

    expect(apiWorkouts.newWorkoutContextQueryOptions).toHaveBeenCalled();
    expect(result).toBe(apiOptions);
  });

  it("uses demo new-workout context query options for demo users", () => {
    const demoOptions = {
      queryKey: ["demo_new_workout_context"],
      queryFn: vi.fn(),
    };

    vi.mocked(
      demoQueryOptions.getDemoNewWorkoutContextQueryOptions,
    ).mockReturnValue(demoOptions as any);

    const result = getNewWorkoutContextQueryOptions(null);

    expect(
      demoQueryOptions.getDemoNewWorkoutContextQueryOptions,
    ).toHaveBeenCalled();
    expect(result).toBe(demoOptions);
  });

  it("uses API focus query options for authenticated users", () => {
    const user = { id: "user-1" } as any;
    const apiOptions = { queryKey: ["workout_focus"], queryFn: vi.fn() };

    vi.mocked(apiWorkouts.workoutsFocusValuesQueryOptions).mockReturnValue(
      apiOptions as any,
    );

    const result = getWorkoutsFocusQueryOptions(user);

    expect(apiWorkouts.workoutsFocusValuesQueryOptions).toHaveBeenCalled();
    expect(result).toBe(apiOptions);
  });

  it("uses demo focus query options for demo users", () => {
    const demoOptions = { queryKey: ["demo_workout_focus"], queryFn: vi.fn() };

    vi.mocked(
      demoQueryOptions.getDemoWorkoutsFocusValuesQueryOptions,
    ).mockReturnValue(demoOptions as any);

    const result = getWorkoutsFocusQueryOptions(null);

    expect(
      demoQueryOptions.getDemoWorkoutsFocusValuesQueryOptions,
    ).toHaveBeenCalled();
    expect(result).toBe(demoOptions);
  });

  it("uses API contribution query options for authenticated users", () => {
    const user = { id: "user-1" } as any;
    const apiOptions = { queryKey: ["workout_contribution"], queryFn: vi.fn() };

    vi.mocked(apiWorkouts.contributionDataQueryOptions).mockReturnValue(
      apiOptions as any,
    );

    const result = getWorkoutContributionQueryOptions(user);

    expect(apiWorkouts.contributionDataQueryOptions).toHaveBeenCalled();
    expect(result).toBe(apiOptions);
  });

  it("uses demo contribution query options for demo users", () => {
    const demoOptions = {
      queryKey: ["demo_workout_contribution"],
      queryFn: vi.fn(),
    };

    vi.mocked(
      demoQueryOptions.getDemoContributionDataQueryOptions,
    ).mockReturnValue(demoOptions as any);

    const result = getWorkoutContributionQueryOptions(null);

    expect(
      demoQueryOptions.getDemoContributionDataQueryOptions,
    ).toHaveBeenCalled();
    expect(result).toBe(demoOptions);
  });
});
