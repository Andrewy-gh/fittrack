import {
  activateTouchTap,
  beginTouchTapTracking,
  cancelTouchTapTracking,
  finishTouchTapTracking,
  hasRecentTouchActivation,
  markRecentTouchActivation,
  shouldSuppressTouchClick,
  updateTouchTapTracking,
} from "@/lib/touch-activation";
import { describe, expect, it, vi } from "vitest";

function createTouchEvent(x: number, y: number) {
  return {
    touches: [{ clientX: x, clientY: y }],
    changedTouches: [{ clientX: x, clientY: y }],
  } as unknown as TouchEvent;
}

function dispatchClick(element: HTMLElement, timeStamp: number) {
  const event = new MouseEvent("click", { bubbles: true, cancelable: true });

  Object.defineProperty(event, "timeStamp", {
    configurable: true,
    value: timeStamp,
  });

  element.dispatchEvent(event);
}

function createTouchTarget(tagName: "button" | "div") {
  const element = document.createElement(tagName);
  const clickSpy = vi.fn();

  element.addEventListener(
    "click",
    (event) => {
      if (!shouldSuppressTouchClick(event.currentTarget, event.timeStamp)) {
        return;
      }

      event.preventDefault();
      event.stopPropagation();
    },
    true,
  );
  element.addEventListener("click", clickSpy);

  return { clickSpy, element };
}

describe("touch activation helpers", () => {
  it("treats a short touch as a tap", () => {
    const element = document.createElement("div");

    beginTouchTapTracking(element, createTouchEvent(10, 20));

    expect(finishTouchTapTracking(element, createTouchEvent(14, 24))).toBe(
      true,
    );
  });

  it("treats a moved touch as a drag", () => {
    const element = document.createElement("div");

    beginTouchTapTracking(element, createTouchEvent(10, 20));
    updateTouchTapTracking(element, createTouchEvent(10, 42));

    expect(finishTouchTapTracking(element, createTouchEvent(10, 42))).toBe(
      false,
    );
  });

  it("does not treat a canceled touch as a tap", () => {
    const element = document.createElement("div");

    beginTouchTapTracking(element, createTouchEvent(10, 20));
    cancelTouchTapTracking(element);

    expect(finishTouchTapTracking(element, createTouchEvent(10, 20))).toBe(
      false,
    );
  });

  it("tracks recent touch activation windows", () => {
    const element = document.createElement("div");

    markRecentTouchActivation(element, 100);

    expect(hasRecentTouchActivation(element, 200)).toBe(true);
    expect(hasRecentTouchActivation(element, 900)).toBe(false);
  });

  it("allows repeated quick taps on the same selectable item while suppressing follow-up clicks", () => {
    const { clickSpy, element } = createTouchTarget("button");
    const preventDefault = vi.fn();

    beginTouchTapTracking(element, createTouchEvent(10, 20));

    expect(
      activateTouchTap(element, {
        ...createTouchEvent(10, 20),
        preventDefault,
        timeStamp: 100,
      } as unknown as TouchEvent),
    ).toBe(true);
    expect(preventDefault).toHaveBeenCalledTimes(1);
    expect(clickSpy).toHaveBeenCalledTimes(1);
    expect(hasRecentTouchActivation(element, 200)).toBe(true);

    dispatchClick(element, 200);
    expect(clickSpy).toHaveBeenCalledTimes(1);

    beginTouchTapTracking(element, createTouchEvent(10, 20));
    expect(
      activateTouchTap(element, {
        ...createTouchEvent(10, 20),
        preventDefault,
        timeStamp: 250,
      } as unknown as TouchEvent),
    ).toBe(true);
    expect(clickSpy).toHaveBeenCalledTimes(2);
  });

  it("allows repeated quick taps on the create row while suppressing follow-up clicks", () => {
    const { clickSpy, element } = createTouchTarget("div");

    beginTouchTapTracking(element, createTouchEvent(10, 20));
    activateTouchTap(element, {
      ...createTouchEvent(10, 20),
      preventDefault: vi.fn(),
      timeStamp: 100,
    } as unknown as TouchEvent);

    dispatchClick(element, 200);
    expect(clickSpy).toHaveBeenCalledTimes(1);

    beginTouchTapTracking(element, createTouchEvent(10, 20));
    activateTouchTap(element, {
      ...createTouchEvent(10, 20),
      preventDefault: vi.fn(),
      timeStamp: 250,
    } as unknown as TouchEvent);

    expect(clickSpy).toHaveBeenCalledTimes(2);
  });
});
