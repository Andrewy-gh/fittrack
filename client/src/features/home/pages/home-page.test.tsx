import type { ComponentPropsWithoutRef } from "react";
import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { CurrentUser } from "@stackframe/react";

const displayModeMock = vi.hoisted(() => ({
  displayMode: "web" as "web" | "pwa",
}));

const routerMock = vi.hoisted(() => ({
  pathname: "/workouts",
}));

vi.mock("@tanstack/react-router", () => ({
  Link: ({
    children,
    to,
    preload: _preload,
    ...props
  }: ComponentPropsWithoutRef<"a"> & { to?: string; preload?: boolean }) => (
    <a
      href={to}
      {...props}
    >
      {children}
    </a>
  ),
  createFileRoute: () => () => ({
    useRouteContext: () => ({ user: null }),
  }),
  useRouterState: ({
    select,
  }: {
    select: (state: { location: { pathname: string } }) => string;
  }) => select({ location: { pathname: routerMock.pathname } }),
}));

vi.mock("@/hooks/use-display-mode", () => ({
  useDisplayMode: () => displayModeMock.displayMode,
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

import { HomePage } from "@/features/home/pages/home-page";

describe("HomePage feature cards", () => {
  beforeEach(() => {
    displayModeMock.displayMode = "web";
    routerMock.pathname = "/workouts";
  });

  it("renders the PWA bottom bar for guest landing sessions", () => {
    displayModeMock.displayMode = "pwa";

    const { container } = render(<HomePage user={null} />);

    expect(
      screen.getByRole("navigation", { name: "PWA navigation" }),
    ).toBeInTheDocument();
    expect(container.firstElementChild).toHaveClass(
      "pb-[calc(5rem+env(safe-area-inset-bottom))]",
    );
  });

  it("does not render the PWA bottom bar for logged-in landing sessions", () => {
    displayModeMock.displayMode = "pwa";

    render(<HomePage user={{ id: "user_1" } as CurrentUser} />);

    expect(
      screen.queryByRole("navigation", { name: "PWA navigation" }),
    ).not.toBeInTheDocument();
  });
});
