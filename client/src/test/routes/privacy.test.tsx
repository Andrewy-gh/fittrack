import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

const routeContextMock = vi.hoisted(() => vi.fn());
const displayModeMock = vi.hoisted(() => vi.fn());
const appShellMock = vi.hoisted(() => vi.fn());

vi.mock("@tanstack/react-router", () => ({
  createFileRoute: () => (config: object) => ({
    ...config,
    useRouteContext: routeContextMock,
  }),
}));

vi.mock("@/components/nav/app-shell", () => ({
  AppShell: ({ user }: { user: { id: string } | null }) => {
    appShellMock(user);

    return <div data-testid="app-shell" />;
  },
}));

vi.mock("@/features/privacy/privacy-content", () => ({
  effectiveDate: "June 28, 2026",
  getPolicySections: () => [
    {
      id: "summary",
      title: "Summary",
      content: <p>FitTrack privacy details.</p>,
    },
  ],
}));

vi.mock("@/hooks/use-display-mode", () => ({
  useDisplayMode: displayModeMock,
}));

import { PrivacyPage } from "@/routes/privacy";

describe("privacy route", () => {
  it("renders the shared app shell with the route user", () => {
    const user = { id: "user_1" };
    routeContextMock.mockReturnValue({ user });
    displayModeMock.mockReturnValue("browser");

    render(<PrivacyPage />);

    expect(screen.getByTestId("app-shell")).toBeInTheDocument();
    expect(appShellMock).toHaveBeenCalledWith(user);
    expect(
      screen.getByRole("heading", { name: "Privacy Policy" }),
    ).toBeVisible();
  });
});
