import { describe, it, expect } from "vitest";
import { formatWeight } from "./utils";

describe("formatWeight", () => {
  it('should return "0" for null', () => {
    expect(formatWeight(null)).toBe("0");
  });

  it("should return whole number without decimal", () => {
    expect(formatWeight(45)).toBe("45");
    expect(formatWeight(135)).toBe("135");
    expect(formatWeight(225)).toBe("225");
  });

  it("should round to one decimal place for numbers with more precision", () => {
    // Note: Due to JavaScript floating point precision, 45.55 may be stored as 45.549999...
    expect(formatWeight(45.55)).toBe("45.5");
    expect(formatWeight(45.54)).toBe("45.5");
    expect(formatWeight(45.56)).toBe("45.6");
    expect(formatWeight(45.99)).toBe("46.0");
  });
});
