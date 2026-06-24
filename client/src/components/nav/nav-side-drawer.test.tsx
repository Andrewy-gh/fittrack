import type { ComponentPropsWithoutRef, ReactNode } from "react";
import { render, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { CurrentUser } from "@stackframe/react";
import { NavSideDrawer } from "@/components/nav/nav-side-drawer";

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

vi.mock("@/components/ui/drawer", () => ({
  Drawer: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  DrawerTrigger: ({ children }: { children: ReactNode }) => (
    <div>{children}</div>
  ),
  DrawerContent: ({ children, ...props }: ComponentPropsWithoutRef<"div">) => (
    <div {...props}>{children}</div>
  ),
  DrawerHeader: ({ children, ...props }: ComponentPropsWithoutRef<"div">) => (
    <div {...props}>{children}</div>
  ),
  DrawerTitle: ({ children, ...props }: ComponentPropsWithoutRef<"h2">) => (
    <h2 {...props}>{children}</h2>
  ),
  DrawerClose: ({ children }: { children: ReactNode }) => <>{children}</>,
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

describe("NavSideDrawer", () => {
  beforeEach(() => {
    routerMock.pathname = "/workouts";
  });

  it("renders the shared nav items without Home and marks the active route", () => {
    routerMock.pathname = "/exercises/bench-press";

    render(<NavSideDrawer user={{} as CurrentUser} />);

    const nav = screen.getByRole("navigation", { name: "Main navigation" });

    expect(
      screen.queryByRole("link", { name: "Home" }),
    ).not.toBeInTheDocument();
    expect(
      within(nav).getByRole("link", { name: "Workouts" }),
    ).toBeInTheDocument();
    expect(
      within(nav).getByRole("link", { name: "Exercises" }),
    ).toHaveAttribute("aria-current", "page");
    expect(
      within(nav).getByRole("link", { name: "Analytics" }),
    ).toBeInTheDocument();
    expect(
      within(nav).getByRole("link", { name: "AI Chat" }),
    ).toBeInTheDocument();
  });

  it("renders the signed-in account slot for authenticated users", () => {
    render(<NavSideDrawer user={{} as CurrentUser} />);

    expect(
      screen.getByRole("button", { name: "Signed-in user menu" }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Guest user menu" }),
    ).not.toBeInTheDocument();
  });

  it("renders the guest account slot without a user", () => {
    render(<NavSideDrawer user={null} />);

    expect(
      screen.getByRole("button", { name: "Guest user menu" }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Signed-in user menu" }),
    ).not.toBeInTheDocument();
  });
});
