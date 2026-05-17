import { createEvent, fireEvent, render, screen } from "@testing-library/react";
import {
  afterAll,
  beforeAll,
  beforeEach,
  describe,
  expect,
  it,
  vi,
} from "vitest";
import { ExerciseCombobox } from "@/features/workouts/form/exercise-combobox";

let mediaQueryMatches = false;
const originalResizeObserver = globalThis.ResizeObserver;
const originalScrollIntoView = HTMLElement.prototype.scrollIntoView;
const originalMatchMedia = globalThis.matchMedia;

beforeAll(() => {
  class ResizeObserverMock {
    observe() {}
    unobserve() {}
    disconnect() {}
  }

  Object.defineProperty(globalThis, "ResizeObserver", {
    value: ResizeObserverMock,
    writable: true,
  });

  Object.defineProperty(HTMLElement.prototype, "scrollIntoView", {
    value: () => {},
    writable: true,
  });

  Object.defineProperty(globalThis, "matchMedia", {
    value: (query: string) =>
      ({
        matches: mediaQueryMatches,
        media: query,
        onchange: null,
        addEventListener: () => {},
        removeEventListener: () => {},
        addListener: () => {},
        removeListener: () => {},
        dispatchEvent: () => false,
      }) as MediaQueryList,
    writable: true,
  });
});

beforeEach(() => {
  mediaQueryMatches = false;
});

afterAll(() => {
  Object.defineProperty(globalThis, "ResizeObserver", {
    value: originalResizeObserver,
    writable: true,
  });

  Object.defineProperty(HTMLElement.prototype, "scrollIntoView", {
    value: originalScrollIntoView,
    writable: true,
  });

  Object.defineProperty(globalThis, "matchMedia", {
    value: originalMatchMedia,
    writable: true,
  });
});

function touchStart(target: Element, x = 8, y = 12) {
  fireEvent.touchStart(target, {
    touches: [{ clientX: x, clientY: y }],
    changedTouches: [{ clientX: x, clientY: y }],
  });
}

function touchEnd(target: Element, x = 8, y = 12) {
  const event = createEvent.touchEnd(target, {
    changedTouches: [{ clientX: x, clientY: y }],
  });

  fireEvent(target, event);
  return event;
}

function touchTap(target: Element, x = 8, y = 12) {
  touchStart(target, x, y);
  return touchEnd(target, x, y);
}

async function renderMobileExerciseCombobox(onCreate = vi.fn()) {
  render(
    <ExerciseCombobox
      options={[{ id: 1, name: "Squat" }]}
      selected=""
      onChange={vi.fn()}
      onCreate={onCreate}
    />,
  );

  const trigger = screen.getByText("Select exercise...").closest("button");

  if (!trigger) {
    throw new Error("Expected exercise combobox trigger button");
  }

  fireEvent.click(trigger);
  fireEvent.change(screen.getByPlaceholderText("Search exercises..."), {
    target: { value: "Bench" },
  });

  const [createRow] = await screen.findAllByText('Create "Bench"');

  return { createRow, onCreate };
}

describe("ExerciseCombobox create row", () => {
  it("creates an option on touch release in the mobile drawer path", async () => {
    const { createRow, onCreate } = await renderMobileExerciseCombobox();

    const releasedTouch = touchTap(createRow);

    expect(onCreate).toHaveBeenCalledTimes(1);
    expect(onCreate).toHaveBeenCalledWith("Bench");
    expect(releasedTouch.defaultPrevented).toBe(true);
  });

  it("does not create an option after touch tracking is canceled", async () => {
    const { createRow, onCreate } = await renderMobileExerciseCombobox();

    touchStart(createRow);
    fireEvent.touchCancel(createRow);
    touchEnd(createRow);

    expect(onCreate).not.toHaveBeenCalled();
  });

  it("does not create a duplicate option after a follow-up click", async () => {
    const { createRow, onCreate } = await renderMobileExerciseCombobox();

    touchTap(createRow);
    fireEvent.click(createRow);

    expect(onCreate).toHaveBeenCalledTimes(1);
  });
});
