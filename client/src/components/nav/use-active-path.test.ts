import { describe, expect, it } from "vitest";
import { isActivePath } from "@/components/nav/use-active-path";

describe("isActivePath", () => {
  it("does not match sibling paths with the same prefix", () => {
    expect(isActivePath("/workouts-old", "/workouts")).toBe(false);
  });

  it("matches root only on the exact root path", () => {
    expect(isActivePath("/", "/")).toBe(true);
    expect(isActivePath("/workouts", "/")).toBe(false);
    expect(isActivePath("/workouts/123", "/")).toBe(false);
  });
});
