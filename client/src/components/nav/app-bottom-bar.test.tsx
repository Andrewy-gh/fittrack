import type { ComponentPropsWithoutRef } from "react";
import { render, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { CurrentUser } from "@stackframe/react";
import { AppBottomBar } from "@/components/nav/app-bottom-bar";

const routerMock = vi.hoisted(() => ({ pathname: "/workouts" }));

vi.mock("@tanstack/react-router", () => ({
  Link: ({
    children,
    to,
    ...props
  }: ComponentPropsWithoutRef<"a"> & { to: string }) => (
    <a
      href={to}
      {...props}
    >
      {children}
    </a>
  ),
  useRouterState: ({
    select,
  }: {
    select: (state: { location: { pathname: string } }) => string;
  }) => select({ location: { pathname: routerMock.pathname } }),
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

describe("AppBottomBar", () => {
  beforeEach(() => {
    routerMock.pathname = "/workouts";
  });

  it("renders icon-only nav tabs with the active route highlighted", () => {
    routerMock.pathname = "/analytics/monthly";

    render(<AppBottomBar user={{} as CurrentUser} />);

    const nav = screen.getByRole("navigation", { name: "PWA navigation" });

    expect(
      within(nav).getByRole("link", { name: "Workouts" }),
    ).toBeInTheDocument();
    expect(
      within(nav).getByRole("link", { name: "Exercises" }),
    ).toBeInTheDocument();
    expect(
      within(nav).getByRole("link", { name: "Analytics" }),
    ).toHaveAttribute("aria-current", "page");
    expect(
      within(nav).getByRole("link", { name: "AI Chat" }),
    ).toBeInTheDocument();
    expect(nav).toHaveStyle({
      paddingBottom: "calc(env(safe-area-inset-bottom) + 0.75rem)",
    });
  });

  it("renders the signed-in account slot for authenticated users", () => {
    render(<AppBottomBar user={{} as CurrentUser} />);

    expect(
      screen.getByRole("button", { name: "Signed-in user menu" }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Guest user menu" }),
    ).not.toBeInTheDocument();
  });

  it("renders the guest account slot without a user", () => {
    render(<AppBottomBar user={null} />);

    expect(
      screen.getByRole("button", { name: "Guest user menu" }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Signed-in user menu" }),
    ).not.toBeInTheDocument();
  });
});
