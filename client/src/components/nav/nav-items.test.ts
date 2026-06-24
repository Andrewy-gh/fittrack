import { describe, expect, it } from "vitest";
import { navItems } from "@/components/nav/nav-items";

describe("navItems", () => {
  it("lists the shared app destinations in display order", () => {
    expect(
      navItems.map(({ icon: _icon, ...destination }) => destination),
    ).toEqual([
      { to: "/workouts", label: "Workouts" },
      { to: "/exercises", label: "Exercises" },
      { to: "/analytics", label: "Analytics" },
      { to: "/chat", label: "AI Chat" },
    ]);
  });
});
