import type { ComponentPropsWithoutRef } from "react";
import { render, screen, within } from "@testing-library/react";
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

  it("uses the favicon for the navbar brand mark", () => {
    render(<HomePage user={null} />);

    const brandText = screen.getByText("FITTRACK");
    const nav = brandText.closest("nav");

    expect(nav).not.toBeNull();
    expect((nav as HTMLElement).querySelector("img")).toHaveAttribute(
      "src",
      "/favicon.svg",
    );
  });

  it("removes highlight badges from the grounded features cards", () => {
    render(<HomePage user={null} />);

    const groundedFeaturesHeading = screen.getByRole("heading", {
      name: /grounded features/i,
    });
    const groundedFeaturesSection = groundedFeaturesHeading.closest("section");

    expect(groundedFeaturesSection).not.toBeNull();

    const section = within(groundedFeaturesSection as HTMLElement);

    expect(
      section.getByRole("heading", { name: "Fast workout logging" }),
    ).toBeInTheDocument();
    expect(
      section.getByRole("heading", { name: "Repeat what you already did" }),
    ).toBeInTheDocument();
    expect(section.queryByText("QUICK ENTRY")).not.toBeInTheDocument();
    expect(section.queryByText("REPEAT LAST")).not.toBeInTheDocument();
    expect(section.queryByText("CLEAR HISTORY")).not.toBeInTheDocument();
    expect(section.queryByText("PLAIN SUMMARY")).not.toBeInTheDocument();
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
