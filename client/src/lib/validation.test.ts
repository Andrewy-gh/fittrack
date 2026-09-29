import { describe, it, expect } from "vitest";
import { minValue, maxValue, compose } from "./validation";

describe("minValue", () => {
  it("handles non-number inputs gracefully", () => {
    const validator = minValue(10);
    expect(validator("test")).toBeUndefined();
    expect(validator(null)).toBeUndefined();
    expect(validator(undefined)).toBeUndefined();
  });
});

describe("maxValue", () => {
  it("handles non-number inputs gracefully", () => {
    const validator = maxValue(100);
    expect(validator("test")).toBeUndefined();
    expect(validator(null)).toBeUndefined();
    expect(validator(undefined)).toBeUndefined();
  });
});

describe("compose", () => {
  it("returns first error from multiple validators", () => {
    const validator = compose(minValue(5), maxValue(10));
    expect(validator(4)).toBe("This field must be at least 5");
    expect(validator(11)).toBe("This field must be at most 10");
  });

  it("returns undefined when all pass", () => {
    const validator = compose(minValue(3), maxValue(10));
    expect(validator(5)).toBeUndefined();
  });

  it("passes field name to all validators", () => {
    const validator = compose(minValue(5), maxValue(10));
    expect(validator(4, "Score")).toBe("Score must be at least 5");
    expect(validator(11, "Score")).toBe("Score must be at most 10");
  });
});
