import { describe, it, expect } from "vitest";
import { getPaginationItems } from "./pagination-utils";

describe("pagination-utils", () => {
  it("returns a single page when totalPages is 1", () => {
    expect(getPaginationItems(1, 1)).toEqual([1]);
  });

  it("returns both pages when totalPages is 2", () => {
    expect(getPaginationItems(1, 2)).toEqual([1, 2]);
  });

  it("returns all pages when totalPages is 5", () => {
    expect(getPaginationItems(3, 5)).toEqual([1, 2, 3, 4, 5]);
  });

  it("returns start range with ellipsis when current page is near the start", () => {
    expect(getPaginationItems(1, 10)).toEqual([1, 2, 3, 4, "ellipsis", 10]);
  });

  it("returns middle range with ellipses when current page is in the middle", () => {
    expect(getPaginationItems(5, 10)).toEqual([
      1,
      "ellipsis",
      4,
      5,
      6,
      "ellipsis",
      10,
    ]);
  });

  it("returns end range with ellipsis when current page is near the end", () => {
    expect(getPaginationItems(10, 10)).toEqual([1, "ellipsis", 7, 8, 9, 10]);
  });
});
