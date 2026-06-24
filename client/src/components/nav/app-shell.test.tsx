import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { CurrentUser } from "@stackframe/react";
import { AppShell } from "@/components/nav/app-shell";

const displayModeMock = vi.hoisted(() => ({
  displayMode: "web" as "web" | "pwa",
}));

vi.mock("@/hooks/use-display-mode", () => ({
  useDisplayMode: () => displayModeMock.displayMode,
}));

vi.mock("@/components/nav/app-top-bar", () => ({
  AppTopBar: ({ user }: { user: CurrentUser | null }) => (
    <div data-testid="top-bar">{user ? "authed web" : "guest web"}</div>
  ),
}));

vi.mock("@/components/nav/app-bottom-bar", () => ({
  AppBottomBar: ({ user }: { user: CurrentUser | null }) => (
    <div data-testid="bottom-bar">{user ? "authed pwa" : "guest pwa"}</div>
  ),
}));

describe("AppShell", () => {
  beforeEach(() => {
    displayModeMock.displayMode = "web";
  });

  it("renders the top bar in web display mode", () => {
    render(<AppShell user={{} as CurrentUser} />);

    expect(screen.getByTestId("top-bar")).toHaveTextContent("authed web");
    expect(screen.queryByTestId("bottom-bar")).not.toBeInTheDocument();
  });

  it("renders the bottom bar in pwa display mode", () => {
    displayModeMock.displayMode = "pwa";

    render(<AppShell user={null} />);

    expect(screen.getByTestId("bottom-bar")).toHaveTextContent("guest pwa");
    expect(screen.queryByTestId("top-bar")).not.toBeInTheDocument();
  });
});
