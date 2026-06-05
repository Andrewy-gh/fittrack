import type { ComponentPropsWithoutRef } from "react";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { DemoHeader } from "@/components/demo-header";

vi.mock("@tanstack/react-router", () => ({
  Link: ({ children, ...props }: ComponentPropsWithoutRef<"a">) => (
    <a {...props}>{children}</a>
  ),
}));

vi.mock("@/components/guest-user-button", () => ({
  GuestUserButton: () => (
    <button
      type="button"
      aria-label="Guest user menu"
    >
      guest
    </button>
  ),
}));

vi.mock("@/components/mobile-bottom-nav", () => ({
  MobileBottomNav: () => null,
}));

describe("DemoHeader", () => {
  it("uses the guest user button instead of a separate theme button", () => {
    render(<DemoHeader />);

    expect(
      screen.getByRole("button", { name: "Guest user menu" }),
    ).toBeInTheDocument();
    expect(screen.getAllByRole("button")).toHaveLength(1);
  });
});
