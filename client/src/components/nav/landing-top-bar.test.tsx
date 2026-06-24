import type { ComponentPropsWithoutRef } from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { CurrentUser } from "@stackframe/react";
import { LandingTopBar } from "@/components/nav/landing-top-bar";

vi.mock("@tanstack/react-router", () => ({
  Link: ({
    children,
    to,
    preload: _preload,
    ...props
  }: ComponentPropsWithoutRef<"a"> & { to: string; preload?: boolean }) => (
    <a
      href={to}
      {...props}
    >
      {children}
    </a>
  ),
}));

vi.mock("@/components/custom-user-button", () => ({
  CustomUserButton: () => (
    <button
      type="button"
      aria-label="Signed-in user menu"
    >
      user
    </button>
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

describe("LandingTopBar", () => {
  it('labels the quick-jump menu "Try ▾" for guests and links to every app destination', async () => {
    render(<LandingTopBar user={null} />);

    const menu = screen.getByRole("button", { name: "Try ▾" });
    expect(menu).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Guest user menu" }),
    ).toBeInTheDocument();

    await userEvent.click(menu);

    expect(screen.getByRole("menuitem", { name: /workouts/i })).toHaveAttribute(
      "href",
      "/workouts",
    );
    expect(
      screen.getByRole("menuitem", { name: /exercises/i }),
    ).toHaveAttribute("href", "/exercises");
    expect(
      screen.getByRole("menuitem", { name: /analytics/i }),
    ).toHaveAttribute("href", "/analytics");
    expect(screen.getByRole("menuitem", { name: /ai chat/i })).toHaveAttribute(
      "href",
      "/chat",
    );
  });

  it('labels the quick-jump menu "Open ▾" for logged-in users', () => {
    render(<LandingTopBar user={{ id: "user_1" } as CurrentUser} />);

    expect(screen.getByRole("button", { name: "Open ▾" })).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Signed-in user menu" }),
    ).toBeInTheDocument();
  });
});
