import { describe, expect, it } from "vitest";
import { analyticsSearchValidator } from "./route-search-validation";

describe("Analytics assessment search", () => {
  it("retains a valid explicit assessment week and exercise", () => {
    expect(
      analyticsSearchValidator.parse({
        exerciseId: "9",
        assessmentWeek: "2026-08-10",
      }),
    ).toEqual({ exerciseId: 9, assessmentWeek: "2026-08-10" });
  });
  it("keeps ordinary Analytics navigation free of an explicit week", () => {
    expect(analyticsSearchValidator.parse({})).toEqual({});
  });
  it.each(["not-a-date", "2026-02-30", "2026-13-01", "2026-08-10T12:00:00Z"])(
    "rejects invalid local-date input %s",
    (assessmentWeek) => {
      expect(() =>
        analyticsSearchValidator.parse({ assessmentWeek }),
      ).toThrow();
    },
  );
});
