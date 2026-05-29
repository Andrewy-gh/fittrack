import { describe, expect, it } from "vitest";
import { featureAccessQueryOptions } from "./feature-access";

describe("feature access api wrapper", () => {
  it("scopes the feature access query cache by user", () => {
    expect(featureAccessQueryOptions("user-123").queryKey).toEqual([
      "feature-access",
      "user-123",
    ]);
  });
});
