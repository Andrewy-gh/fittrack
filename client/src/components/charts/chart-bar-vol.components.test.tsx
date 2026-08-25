import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { ScrollableChart } from "./chart-bar-vol.components";

vi.mock("./chart-bar-vol.utils", async () => {
  const actual = await vi.importActual<typeof import("./chart-bar-vol.utils")>(
    "./chart-bar-vol.utils",
  );

  return {
    ...actual,
    useBreakpoint: () => "desktop",
  };
});

let clientWidth = 500;
let coarsePointer = false;
let resizeCallbacks: Array<() => void> = [];
const scrollBy = vi.fn();
const originalClientWidth = Object.getOwnPropertyDescriptor(
  HTMLElement.prototype,
  "clientWidth",
);
const originalMatchMedia = window.matchMedia;
const originalResizeObserver = globalThis.ResizeObserver;
const originalRequestAnimationFrame = globalThis.requestAnimationFrame;
const originalCancelAnimationFrame = globalThis.cancelAnimationFrame;
const originalScrollBy = HTMLElement.prototype.scrollBy;

class ResizeObserverMock implements ResizeObserver {
  readonly callback: ResizeObserverCallback;

  constructor(callback: ResizeObserverCallback) {
    this.callback = callback;
    resizeCallbacks.push(() => callback([], this));
  }

  observe() {}
  unobserve() {}
  disconnect() {}
  takeRecords(): ResizeObserverEntry[] {
    return [];
  }
}

beforeEach(() => {
  clientWidth = 500;
  coarsePointer = false;
  resizeCallbacks = [];
  scrollBy.mockReset();

  Object.defineProperty(HTMLElement.prototype, "clientWidth", {
    configurable: true,
    get: () => clientWidth,
  });
  Object.defineProperty(HTMLElement.prototype, "scrollBy", {
    configurable: true,
    value: scrollBy,
  });
  Object.defineProperty(window, "matchMedia", {
    configurable: true,
    value: vi.fn((query: string) => ({
      matches: query === "(pointer: coarse)" && coarsePointer,
      media: query,
      onchange: null,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      addListener: vi.fn(),
      removeListener: vi.fn(),
      dispatchEvent: vi.fn(),
    })),
  });
  Object.defineProperty(globalThis, "ResizeObserver", {
    configurable: true,
    value: ResizeObserverMock,
  });
  Object.defineProperty(globalThis, "requestAnimationFrame", {
    configurable: true,
    value: (callback: FrameRequestCallback) => {
      callback(0);
      return 1;
    },
  });
  Object.defineProperty(globalThis, "cancelAnimationFrame", {
    configurable: true,
    value: vi.fn(),
  });
});

afterEach(() => {
  if (originalClientWidth) {
    Object.defineProperty(
      HTMLElement.prototype,
      "clientWidth",
      originalClientWidth,
    );
  }
  Object.defineProperty(HTMLElement.prototype, "scrollBy", {
    configurable: true,
    value: originalScrollBy,
  });
  Object.defineProperty(window, "matchMedia", {
    configurable: true,
    value: originalMatchMedia,
  });
  Object.defineProperty(globalThis, "ResizeObserver", {
    configurable: true,
    value: originalResizeObserver,
  });
  Object.defineProperty(globalThis, "requestAnimationFrame", {
    configurable: true,
    value: originalRequestAnimationFrame,
  });
  Object.defineProperty(globalThis, "cancelAnimationFrame", {
    configurable: true,
    value: originalCancelAnimationFrame,
  });
});

function getScrollContainer(container: HTMLElement): HTMLElement {
  const element = container.querySelector(".overflow-x-auto");
  if (!(element instanceof HTMLElement)) {
    throw new Error("Expected a scrollable chart container");
  }
  return element;
}

describe("ScrollableChart", () => {
  it("fits the visible bars, starts at the latest data, and clamps after resize", async () => {
    const user = userEvent.setup();
    const { container } = render(
      <ScrollableChart
        dataLength={20}
        visibleBarCount={5}
      >
        <div>Chart</div>
      </ScrollableChart>,
    );
    const scrollContainer = getScrollContainer(container);

    await waitFor(() => {
      expect(scrollContainer.firstElementChild).toHaveStyle({
        width: "2000px",
      });
      expect(scrollContainer.scrollLeft).toBe(1500);
    });
    expect(screen.getByRole("button", { name: "Scroll left" })).toBeVisible();
    expect(
      screen.queryByRole("button", { name: "Scroll right" }),
    ).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Scroll left" }));
    expect(scrollBy).toHaveBeenCalledWith({
      left: -400,
      behavior: "smooth",
    });

    clientWidth = 300;
    for (const notifyResize of resizeCallbacks) notifyResize();

    await waitFor(() => {
      expect(scrollContainer.firstElementChild).toHaveStyle({
        width: "1200px",
      });
      expect(scrollContainer.scrollLeft).toBe(900);
    });
  });

  it("preserves a historical scroll position after resize", async () => {
    const { container } = render(
      <ScrollableChart
        dataLength={20}
        visibleBarCount={5}
      >
        <div>Chart</div>
      </ScrollableChart>,
    );
    const scrollContainer = getScrollContainer(container);

    await waitFor(() => expect(scrollContainer.scrollLeft).toBe(1500));
    scrollContainer.scrollLeft = 200;
    fireEvent.scroll(scrollContainer);

    clientWidth = 300;
    for (const notifyResize of resizeCallbacks) notifyResize();

    await waitFor(() => {
      expect(scrollContainer.firstElementChild).toHaveStyle({
        width: "1200px",
      });
      expect(scrollContainer.scrollLeft).toBe(200);
    });
  });

  it("keeps an unfitted chart pinned to the latest data after resize", async () => {
    const { container } = render(
      <ScrollableChart dataLength={24}>
        <div>Chart</div>
      </ScrollableChart>,
    );
    const scrollContainer = getScrollContainer(container);

    await waitFor(() => expect(scrollContainer.scrollLeft).toBe(700));

    clientWidth = 300;
    for (const notifyResize of resizeCallbacks) notifyResize();

    await waitFor(() => {
      expect(scrollContainer.scrollLeft).toBe(900);
      expect(
        screen.queryByRole("button", { name: "Scroll right" }),
      ).not.toBeInTheDocument();
    });
  });

  it("hides desktop scroll controls for touch devices", async () => {
    coarsePointer = true;
    const { container } = render(
      <ScrollableChart
        dataLength={20}
        visibleBarCount={5}
      >
        <div>Chart</div>
      </ScrollableChart>,
    );
    const scrollContainer = getScrollContainer(container);

    await waitFor(() => expect(scrollContainer.scrollLeft).toBe(1500));
    expect(
      screen.queryByRole("button", { name: "Scroll left" }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Scroll right" }),
    ).not.toBeInTheDocument();
  });
});
