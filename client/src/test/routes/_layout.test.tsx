import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { CurrentUser } from "@stackframe/react";
import { LayoutComponent } from "@/routes/_layout";

const routeContextMock = vi.hoisted(() => ({
  user: null as CurrentUser | null,
}));

const routerMock = vi.hoisted(() => ({
  pathname: "/workouts",
}));

const displayModeMock = vi.hoisted(() => ({
  displayMode: "web" as "web" | "pwa",
}));

vi.mock("@tanstack/react-router", () => ({
  createFileRoute: () => (config: unknown) => ({
    ...(config as object),
    useRouteContext: () => ({ user: routeContextMock.user }),
  }),
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
    routeContextMock.user = null;
    routerMock.pathname = "/workouts";
    displayModeMock.displayMode = "web";
  });

  it("renders a single app shell for the current user and route outlet", () => {
    routeContextMock.user = { id: "user_1" } as CurrentUser;

    render(<LayoutComponent />);

    expect(screen.getByTestId("app-shell")).toHaveTextContent("authed");
    expect(screen.getAllByTestId("app-shell")).toHaveLength(1);
    expect(screen.getByTestId("route-outlet")).toBeInTheDocument();
  });

  it("does not reserve bottom-nav space for web sessions", () => {
    const { container } = render(<LayoutComponent />);

    expect(container.firstElementChild).not.toHaveClass(
      "pb-[calc(5rem+env(safe-area-inset-bottom))]",
    );
    expect(container.firstElementChild).not.toHaveClass(
      "pt-[env(safe-area-inset-top)]",
    );
  });

  it("reserves safe-area space for PWA sessions only", () => {
    displayModeMock.displayMode = "pwa";

    const { container } = render(<LayoutComponent />);

    expect(container.firstElementChild).toHaveClass(
      "pt-[env(safe-area-inset-top)]",
      "pb-[calc(5rem+env(safe-area-inset-bottom))]",
    );
  });

  it("passes route, display mode, and user state to the install prompt", () => {
    routeContextMock.user = { id: "user_1" } as CurrentUser;
    routerMock.pathname = "/chat/sessions/today";

    render(<LayoutComponent />);

    expect(screen.getByTestId("install-prompt-props")).toHaveTextContent(
      "web:/chat/sessions/today:authed",
    );
  });
});
