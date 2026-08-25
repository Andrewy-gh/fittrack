import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { CurrentUser } from "@stackframe/react";
import { LayoutComponent } from "@/components/nav/layout-component";

const routerMock = vi.hoisted(() => ({
  pathname: "/workouts",
}));

const displayModeMock = vi.hoisted(() => ({
  displayMode: "web" as "web" | "pwa",
}));

vi.mock("@tanstack/react-router", () => ({
  Outlet: () => <main data-testid="route-outlet" />,
  useRouterState: ({
    select,
  }: {
    select: (state: { location: { pathname: string } }) => string;
  }) => select({ location: { pathname: routerMock.pathname } }),
}));

vi.mock("@/hooks/use-display-mode", () => ({
  useDisplayMode: () => displayModeMock.displayMode,
}));

vi.mock("@/components/nav/app-shell", () => ({
  AppShell: ({ user }: { user: CurrentUser | null }) => (
    <div data-testid="app-shell">{user ? "authed" : "guest"}</div>
  ),
}));

vi.mock("@/components/pwa-install-prompt", () => ({
  PwaInstallPrompt: ({
    displayMode,
    pathname,
    user,
  }: {
    displayMode: "web" | "pwa";
    pathname: string;
    user: CurrentUser | null;
  }) => (
    <div data-testid="install-prompt-props">
      {displayMode}:{pathname}:{user ? "authed" : "guest"}
    </div>
  ),
}));

describe("LayoutComponent", () => {
  beforeEach(() => {
    routerMock.pathname = "/workouts";
    displayModeMock.displayMode = "web";
  });

  it("renders a single app shell for the current user and route outlet", () => {
    const user = { id: "user_1" } as CurrentUser;

    render(<LayoutComponent user={user} />);

    expect(screen.getByTestId("app-shell")).toHaveTextContent("authed");
    expect(screen.getAllByTestId("app-shell")).toHaveLength(1);
    expect(screen.getByTestId("route-outlet")).toBeInTheDocument();
  });

  it("reserves safe-area space only for PWA sessions", () => {
    displayModeMock.displayMode = "pwa";

    const { container, rerender } = render(<LayoutComponent user={null} />);

    expect(container.firstElementChild).toHaveClass(
      "pt-[env(safe-area-inset-top)]",
      "pb-[calc(5rem+env(safe-area-inset-bottom))]",
    );

    displayModeMock.displayMode = "web";
    rerender(<LayoutComponent user={null} />);

    expect(container.firstElementChild).not.toHaveClass(
      "pt-[env(safe-area-inset-top)]",
      "pb-[calc(5rem+env(safe-area-inset-bottom))]",
    );
  });

  it("passes route, display mode, and user state to the install prompt", () => {
    const user = { id: "user_1" } as CurrentUser;
    routerMock.pathname = "/chat/sessions/today";

    render(<LayoutComponent user={user} />);

    expect(screen.getByTestId("install-prompt-props")).toHaveTextContent(
      "web:/chat/sessions/today:authed",
    );
  });
});
