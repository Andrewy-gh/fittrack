import { describe, expect, it } from "vitest";
import { navItems } from "@/components/nav/nav-items";

describe("navItems", () => {
  it("lists the shared app destinations in display order", () => {
    expect(
      navItems.map(({ icon: _icon, ...destination }) => destination),
    ).toEqual([
      { to: "/workouts", label: "Workouts", search: undefined },
      { to: "/exercises", label: "Exercises", search: undefined },
      { to: "/analytics", label: "Analytics", search: undefined },
      { to: "/chat", label: "AI Chat", search: undefined },
    ]);
  });
});
