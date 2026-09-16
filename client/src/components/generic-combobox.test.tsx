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
import { GenericCombobox } from "@/components/generic-combobox";

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

async function renderGenericCombobox({
  isDesktop = false,
  onChange = vi.fn(),
  onCreate = vi.fn(),
}: {
  isDesktop?: boolean;
  onChange?: ReturnType<typeof vi.fn>;
  onCreate?: ReturnType<typeof vi.fn>;
} = {}) {
  mediaQueryMatches = isDesktop;

  render(
    <GenericCombobox
      options={[{ name: "Squat" }]}
      selected=""
      ariaLabel="Exercise type"
      inputAriaLabel="Search options"
      onChange={onChange}
      onCreate={onCreate}
    />,
  );

  fireEvent.click(screen.getByRole("combobox", { name: "Exercise type" }));

  return { onChange, onCreate };
}

describe("GenericCombobox", () => {
  it("keeps mobile list gestures from dragging the drawer", async () => {
    await renderGenericCombobox();

    const option = await screen.findByRole("option", { name: "Squat" });
    const list = option.closest('[data-slot="command-list"]');

    expect(list?.hasAttribute("data-vaul-no-drag")).toBe(true);
  });

  it("selects an existing option on touch release in the mobile drawer path", async () => {
    const onChange = vi.fn();
    await renderGenericCombobox({ onChange });

    const option = await screen.findByRole("option", { name: "Squat" });
    const releasedTouch = touchTap(option);

    expect(onChange).toHaveBeenCalledTimes(1);
    expect(onChange).toHaveBeenCalledWith({ name: "Squat" });
    expect(releasedTouch.defaultPrevented).toBe(true);
  });

  it("does not select an existing option after a drag gesture", async () => {
    const onChange = vi.fn();
    await renderGenericCombobox({ onChange });

    const option = await screen.findByRole("option", { name: "Squat" });

    touchStart(option);
    fireEvent.touchMove(option, {
      touches: [{ clientX: 8, clientY: 40 }],
      changedTouches: [{ clientX: 8, clientY: 40 }],
    });
    touchEnd(option, 8, 40);

    expect(onChange).not.toHaveBeenCalled();
  });

  it("does not select an existing option twice after a follow-up click", async () => {
    const onChange = vi.fn();
    await renderGenericCombobox({ onChange });

    const option = await screen.findByRole("option", { name: "Squat" });

    touchTap(option);
    fireEvent.click(option);

    expect(onChange).toHaveBeenCalledTimes(1);
  });

  it("uses the generic search placeholder by default", async () => {
    await renderGenericCombobox();

    expect(screen.getByPlaceholderText("Search options...")).toBeTruthy();
  });

  it("supports exercise-specific labels, search text, and touch creation", async () => {
    const onCreate = vi.fn();

    render(
      <GenericCombobox
        options={[{ id: 1, name: "Squat" }]}
        selected=""
        ariaLabel="Select exercise"
        inputAriaLabel="Search exercises"
        placeholder="Select exercise..."
        searchPlaceholder="Search exercises..."
        onChange={vi.fn()}
        onCreate={onCreate}
      />,
    );

    const trigger = screen.getByRole("combobox", { name: "Select exercise" });
    expect(trigger.textContent).toContain("Select exercise...");
    fireEvent.click(trigger);

    expect(
      await screen.findByRole("dialog", { name: "Select exercise" }),
    ).toBeTruthy();
    expect(screen.getByText("Search and select an option.")).toBeTruthy();

    const searchInput = await screen.findByRole("combobox", {
      name: "Search exercises",
    });
    expect(searchInput.getAttribute("placeholder")).toBe("Search exercises...");
    fireEvent.change(searchInput, { target: { value: "Bench" } });

    const [createRow] = await screen.findAllByRole("option", {
      name: 'Create "Bench"',
    });
    const releasedTouch = touchTap(createRow);

    expect(onCreate).toHaveBeenCalledTimes(1);
    expect(onCreate).toHaveBeenCalledWith("Bench");
    expect(releasedTouch.defaultPrevented).toBe(true);
  });

  it("uses configured search text in the desktop popover", async () => {
    mediaQueryMatches = true;

    render(
      <GenericCombobox
        options={[{ name: "Squat" }]}
        selected=""
        ariaLabel="Select exercise"
        inputAriaLabel="Search exercises"
        searchPlaceholder="Search exercises..."
        onChange={vi.fn()}
      />,
    );

    fireEvent.click(screen.getByRole("combobox", { name: "Select exercise" }));

    const searchInput = await screen.findByRole("combobox", {
      name: "Search exercises",
    });
    expect(searchInput.getAttribute("placeholder")).toBe("Search exercises...");
    expect(
      document.querySelector('[data-slot="popover-content"]'),
    ).not.toBeNull();
  });

  it("does not create an option after touch tracking is canceled", async () => {
    const { onCreate } = await renderGenericCombobox();

    fireEvent.change(screen.getByPlaceholderText("Search options..."), {
      target: { value: "Bench" },
    });

    const [createRow] = await screen.findAllByRole("option", {
      name: 'Create "Bench"',
    });

    touchStart(createRow);
    fireEvent.touchCancel(createRow);
    touchEnd(createRow);

    expect(onCreate).not.toHaveBeenCalled();
  });

  it("does not create a duplicate option after a follow-up click", async () => {
    const { onCreate } = await renderGenericCombobox();

    fireEvent.change(screen.getByPlaceholderText("Search options..."), {
      target: { value: "Bench" },
    });

    const [createRow] = await screen.findAllByRole("option", {
      name: 'Create "Bench"',
    });

    touchTap(createRow);
    fireEvent.click(createRow);

    expect(onCreate).toHaveBeenCalledTimes(1);
  });
});
