import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { CurrentUser } from "@stackframe/react";
import {
  isPwaInstallPromptRoute,
  PwaInstallPrompt,
  PWA_INSTALL_PROMPT_DISMISS_KEY,
} from "@/components/pwa-install-prompt";

const matchMediaState = {
  coarsePointer: true,
  mobileViewport: true,
};

function setTouchDevice() {
  matchMediaState.coarsePointer = true;
  matchMediaState.mobileViewport = true;
  Object.defineProperty(navigator, "maxTouchPoints", {
    configurable: true,
    value: 1,
  });
}

function setMouseDevice() {
  matchMediaState.coarsePointer = false;
  matchMediaState.mobileViewport = true;
  Object.defineProperty(navigator, "maxTouchPoints", {
    configurable: true,
    value: 0,
  });
}

const user = { id: "user_1" } as CurrentUser;

beforeEach(() => {
  setTouchDevice();
  Object.defineProperty(window, "matchMedia", {
    configurable: true,
    value: vi.fn((query: string) => ({
      matches:
        (query === "(pointer: coarse)" && matchMediaState.coarsePointer) ||
        (query === "(max-width: 1024px)" && matchMediaState.mobileViewport),
      media: query,
      onchange: null,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      addListener: vi.fn(),
      removeListener: vi.fn(),
      dispatchEvent: vi.fn(),
    })),
  });
});

describe("isPwaInstallPromptRoute", () => {
  it("matches app routes and child routes only", () => {
    expect(isPwaInstallPromptRoute("/workouts")).toBe(true);
    expect(isPwaInstallPromptRoute("/workouts/123")).toBe(true);
    expect(isPwaInstallPromptRoute("/exercises/barbell-squat")).toBe(true);
    expect(isPwaInstallPromptRoute("/analytics")).toBe(true);
    expect(isPwaInstallPromptRoute("/chat/sessions/today")).toBe(true);

    expect(isPwaInstallPromptRoute("/")).toBe(false);
    expect(isPwaInstallPromptRoute("/settings")).toBe(false);
    expect(isPwaInstallPromptRoute("/privacy")).toBe(false);
    expect(isPwaInstallPromptRoute("/handler/sign-in")).toBe(false);
  });
});

describe("PwaInstallPrompt", () => {
  it("renders for signed-in mobile web users on eligible app routes", async () => {
    render(
      <PwaInstallPrompt
        displayMode="web"
        pathname="/workouts"
        user={user}
      />,
    );

    expect(await screen.findByRole("status")).toHaveTextContent(
      "Install FitTrack",
    );
  });

  it("does not render for guest, standalone, ineligible route, or mouse sessions", async () => {
    const { rerender } = render(
      <PwaInstallPrompt
        displayMode="web"
        pathname="/workouts"
        user={null}
      />,
    );

    await waitFor(() =>
      expect(screen.queryByRole("status")).not.toBeInTheDocument(),
    );

    rerender(
      <PwaInstallPrompt
        displayMode="pwa"
        pathname="/workouts"
        user={user}
      />,
    );
    await waitFor(() =>
      expect(screen.queryByRole("status")).not.toBeInTheDocument(),
    );

    rerender(
      <PwaInstallPrompt
        displayMode="web"
        pathname="/settings"
        user={user}
      />,
    );
    await waitFor(() =>
      expect(screen.queryByRole("status")).not.toBeInTheDocument(),
    );

    setMouseDevice();
    rerender(
      <PwaInstallPrompt
        displayMode="web"
        pathname="/chat"
        user={{ id: "user_2" } as CurrentUser}
      />,
    );
    await waitFor(() =>
      expect(screen.queryByRole("status")).not.toBeInTheDocument(),
    );
  });

  it("does not show again after appearing once during the current login session", async () => {
    const { rerender } = render(
      <PwaInstallPrompt
        displayMode="web"
        pathname="/workouts"
        user={user}
      />,
    );

    expect(await screen.findByRole("status")).toBeInTheDocument();

    rerender(
      <PwaInstallPrompt
        displayMode="web"
        pathname="/settings"
        user={user}
      />,
    );
    await waitFor(() =>
      expect(screen.queryByRole("status")).not.toBeInTheDocument(),
    );

    rerender(
      <PwaInstallPrompt
        displayMode="web"
        pathname="/exercises"
        user={user}
      />,
    );
    await waitFor(() =>
      expect(screen.queryByRole("status")).not.toBeInTheDocument(),
    );
  });

  it("uses durable dismissal when the user chooses not now", async () => {
    render(
      <PwaInstallPrompt
        displayMode="web"
        pathname="/analytics"
        user={user}
      />,
    );

    fireEvent.click(await screen.findByRole("button", { name: "Not now" }));

    expect(localStorage.getItem(PWA_INSTALL_PROMPT_DISMISS_KEY)).toBe("1");
    await waitFor(() =>
      expect(screen.queryByRole("status")).not.toBeInTheDocument(),
    );
  });
});
