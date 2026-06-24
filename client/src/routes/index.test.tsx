import type { CurrentUser } from "@stackframe/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

const redirectMock = vi.hoisted(() =>
  vi.fn((options: { to: string }) => options),
);

vi.mock("@tanstack/react-router", () => ({
  createFileRoute: () => (config: unknown) => config,
  redirect: redirectMock,
}));

vi.mock("@/features/home/pages/home-page", () => ({
  HomePage: () => null,
}));

import { isStandalone, redirectPwaSignedInUser, Route } from "@/routes/index";

function setStandaloneDisplayMode(matches: boolean) {
  Object.defineProperty(window, "matchMedia", {
    configurable: true,
    value: vi.fn().mockReturnValue({
      matches,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    }),
  });
}

function setIosStandalone(value: boolean | undefined) {
  Object.defineProperty(window.navigator, "standalone", {
    configurable: true,
    value,
  });
}

function runGuard(user: CurrentUser | null) {
  return redirectPwaSignedInUser({ context: { user } });
}

describe("index route PWA redirect", () => {
  beforeEach(() => {
    redirectMock.mockClear();
    setStandaloneDisplayMode(false);
    setIosStandalone(undefined);
  });

  it("wires the synchronous beforeLoad guard onto the landing route", () => {
    expect(Route).toMatchObject({ beforeLoad: redirectPwaSignedInUser });
  });

  it("detects standalone mode from display-mode or iOS navigator state", () => {
    expect(isStandalone()).toBe(false);

    setStandaloneDisplayMode(true);
    expect(isStandalone()).toBe(true);

    setStandaloneDisplayMode(false);
    setIosStandalone(true);
    expect(isStandalone()).toBe(true);
  });

  it("redirects PWA logged-in users before the landing page renders", () => {
    setStandaloneDisplayMode(true);
    const user = { id: "user_1" } as CurrentUser;

    let thrown: unknown;
    try {
      runGuard(user);
    } catch (error) {
      thrown = error;
    }

    expect(thrown).toEqual({ to: "/workouts" });
    expect(redirectMock).toHaveBeenCalledWith({ to: "/workouts" });
  });

  it("does not redirect PWA guests", () => {
    setStandaloneDisplayMode(true);

    expect(() => runGuard(null)).not.toThrow();
    expect(redirectMock).not.toHaveBeenCalled();
  });

  it("does not redirect web users or web guests", () => {
    const user = { id: "user_1" } as CurrentUser;

    expect(() => runGuard(user)).not.toThrow();
    expect(() => runGuard(null)).not.toThrow();
    expect(redirectMock).not.toHaveBeenCalled();
  });
});
