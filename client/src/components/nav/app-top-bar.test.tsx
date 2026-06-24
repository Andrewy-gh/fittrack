import type { ComponentPropsWithoutRef } from "react";
import { render, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { CurrentUser } from "@stackframe/react";
import { AppTopBar } from "@/components/nav/app-top-bar";

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

vi.mock("@/components/nav/nav-side-drawer", () => ({
  NavSideDrawer: () => (
    <button
      type="button"
      aria-label="Open navigation menu"
    >
      menu
    </button>
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

describe("AppTopBar", () => {
  beforeEach(() => {
    routerMock.pathname = "/workouts";
  });

  it("renders the app header contract, logo, inline links, and active route", () => {
    routerMock.pathname = "/chat/sessions/today";

    render(<AppTopBar user={{} as CurrentUser} />);

    const header = screen.getByRole("banner");
    const nav = screen.getByRole("navigation", { name: "Primary navigation" });

    expect(header).toHaveAttribute("data-app-header", "true");
    expect(screen.getByRole("link", { name: "FitTrack" })).toHaveAttribute(
      "href",
      "/",
    );
    expect(
      within(nav).getByRole("link", { name: "Workouts" }),
    ).toBeInTheDocument();
    expect(within(nav).getByRole("link", { name: "AI Chat" })).toHaveAttribute(
      "aria-current",
      "page",
    );
    expect(
      screen.getByRole("button", { name: "Open navigation menu" }),
    ).toBeInTheDocument();
  });

  it("renders the signed-in account slot for authenticated users", () => {
    render(<AppTopBar user={{} as CurrentUser} />);

    expect(
      screen.getByRole("button", { name: "Signed-in user menu" }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Guest user menu" }),
    ).not.toBeInTheDocument();
  });

  it("renders the guest account slot without a user", () => {
    render(<AppTopBar user={null} />);

    expect(
      screen.getByRole("button", { name: "Guest user menu" }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Signed-in user menu" }),
    ).not.toBeInTheDocument();
  });
});
