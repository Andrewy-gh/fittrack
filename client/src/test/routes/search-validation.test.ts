import { describe, expect, it } from "vitest";
import {
  exerciseDetailSearchValidator,
  workoutEditorSearchValidator,
  workoutsSearchValidator,
} from "@/lib/route-search-validation";

describe("route search validation", () => {
  it("coerces optional numeric workout list search params and strips unknown keys", () => {
    expect(
      workoutsSearchValidator.parse({
        focusArea: "push",
        sortOrder: "desc",
        itemsPerPage: "25",
        page: "2",
        ignored: "value",
      }),
    ).toEqual({
      focusArea: "push",
      sortOrder: "desc",
      itemsPerPage: 25,
      page: 2,
    });
  });

  it("keeps omitted optional search params absent", () => {
    expect(workoutsSearchValidator.parse({})).toEqual({});
    expect(exerciseDetailSearchValidator.parse({ sortOrder: "asc" })).toEqual({
      sortOrder: "asc",
    });
  });

  it("preserves workout editor boolean params and coerces exercise indexes", () => {
    expect(
      workoutEditorSearchValidator.parse({
        addExercise: true,
        exerciseIndex: "0",
        newExercise: false,
      }),
    ).toEqual({
      addExercise: true,
      exerciseIndex: 0,
      newExercise: false,
    });
  });

  it("rejects invalid numeric and enum query values", () => {
    expect(() => workoutsSearchValidator.parse({ page: "0" })).toThrow();
    expect(() => workoutsSearchValidator.parse({ page: "2.5" })).toThrow();
    expect(() =>
      exerciseDetailSearchValidator.parse({ sortOrder: "newest" }),
    ).toThrow();
  });
});
