import { act, renderHook } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useDisplayMode } from "@/hooks/use-display-mode";

const originalMatchMedia = window.matchMedia;

function setStandaloneNavigator(value: boolean | undefined) {
  Object.defineProperty(navigator, "standalone", {
    configurable: true,
    value,
  });
}

function installDisplayModeMatchMedia(matches: boolean) {
  let currentMatches = matches;
  const listeners = new Set<(event: MediaQueryListEvent) => void>();

  Object.defineProperty(window, "matchMedia", {
    configurable: true,
    value: vi.fn((query: string) => ({
      get matches() {
        return currentMatches;
      },
      media: query,
      onchange: null,
      addEventListener: (
        eventName: string,
        listener: (event: MediaQueryListEvent) => void,
      ) => {
        if (eventName === "change") listeners.add(listener);
      },
      removeEventListener: (
        eventName: string,
        listener: (event: MediaQueryListEvent) => void,
      ) => {
        if (eventName === "change") listeners.delete(listener);
      },
      addListener: () => {},
      removeListener: () => {},
      dispatchEvent: () => false,
    })),
  });

  return {
    setMatches(nextMatches: boolean) {
      currentMatches = nextMatches;
      const event = { matches: nextMatches } as MediaQueryListEvent;
      listeners.forEach((listener) => listener(event));
    },
  };
}

describe("useDisplayMode", () => {
  beforeEach(() => {
    setStandaloneNavigator(undefined);
  });

  afterEach(() => {
    Object.defineProperty(window, "matchMedia", {
      configurable: true,
      value: originalMatchMedia,
    });
    setStandaloneNavigator(undefined);
  });

  it("returns web when the app is running in a browser tab", () => {
    installDisplayModeMatchMedia(false);

    const { result } = renderHook(() => useDisplayMode());

    expect(result.current).toBe("web");
  });

  it("returns pwa when display-mode is standalone", () => {
    installDisplayModeMatchMedia(true);

    const { result } = renderHook(() => useDisplayMode());

    expect(result.current).toBe("pwa");
  });

  it("returns pwa for iOS standalone mode", () => {
    installDisplayModeMatchMedia(false);
    setStandaloneNavigator(true);

    const { result } = renderHook(() => useDisplayMode());

    expect(result.current).toBe("pwa");
  });

  it("reacts when the display-mode media query changes", () => {
    const media = installDisplayModeMatchMedia(false);
    const { result } = renderHook(() => useDisplayMode());

    expect(result.current).toBe("web");

    act(() => media.setMatches(true));

    expect(result.current).toBe("pwa");
  });
});
