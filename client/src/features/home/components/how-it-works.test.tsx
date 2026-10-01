import { act, render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { HowItWorks } from "@/features/home/components/how-it-works";

class MockIntersectionObserver implements IntersectionObserver {
  static instances: MockIntersectionObserver[] = [];
  readonly root = null;
  readonly rootMargin: string;
  readonly scrollMargin = "0px";
  readonly thresholds = [];

  constructor(
    private readonly callback: IntersectionObserverCallback,
    readonly options?: IntersectionObserverInit,
  ) {
    this.rootMargin = options?.rootMargin ?? "0px";
    MockIntersectionObserver.instances.push(this);
  }

  observe() {}
  unobserve() {}
  disconnect() {}
  takeRecords() {
    return [];
  }

  trigger(target: Element, isIntersecting = true) {
    const rect = target.getBoundingClientRect();
    const entry: IntersectionObserverEntry = {
      target,
      isIntersecting,
      boundingClientRect: rect,
      intersectionRect: rect,
      intersectionRatio: isIntersecting ? 1 : 0,
      rootBounds: null,
      time: 0,
    };
    act(() => this.callback([entry], this));
  }
}

const media = { desktop: true, reducedMotion: false };
// Videos currently playing, as the browser would report via `paused`.
const playing = new Set<HTMLMediaElement>();

function stepObserver() {
  const observer = MockIntersectionObserver.instances
    .filter((instance) => instance.options?.rootMargin)
    .at(-1);
  if (!observer) throw new Error("step observer not created");
  return observer;
}

function sectionObserver() {
  const observer = MockIntersectionObserver.instances.find(
    (instance) => !instance.options?.rootMargin,
  );
  if (!observer) throw new Error("section observer not created");
  return observer;
}

function section() {
  return screen.getByRole("region", { name: /log it\. repeat it\./i });
}

function stepItem(title: string) {
  const item = screen.getByRole("heading", { name: title }).closest("li");
  if (!item) throw new Error(`step "${title}" missing`);
  return item;
}

function playingClips() {
  return [...playing].map((video) => video.getAttribute("poster"));
}

describe("HowItWorks", () => {
  beforeEach(() => {
    MockIntersectionObserver.instances = [];
    playing.clear();
    media.desktop = true;
    media.reducedMotion = false;
    vi.stubGlobal("IntersectionObserver", MockIntersectionObserver);
    vi.stubGlobal("matchMedia", (query: string) => ({
      matches: query.includes("reduced-motion")
        ? media.reducedMotion
        : media.desktop,
      addEventListener: () => {},
      removeEventListener: () => {},
    }));
    vi.spyOn(HTMLMediaElement.prototype, "paused", "get").mockImplementation(
      function (this: HTMLMediaElement) {
        return !playing.has(this);
      },
    );
    vi.spyOn(HTMLMediaElement.prototype, "play").mockImplementation(function (
      this: HTMLMediaElement,
    ) {
      playing.add(this);
      return Promise.resolve();
    });
    vi.spyOn(HTMLMediaElement.prototype, "pause").mockImplementation(function (
      this: HTMLMediaElement,
    ) {
      playing.delete(this);
    });
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("lists the log, repeat, and review steps with the first one current", () => {
    render(<HowItWorks />);

    const stepHeadings = within(section()).getAllByRole("heading", {
      level: 3,
    });

    expect(stepHeadings.map((heading) => heading.textContent)).toEqual([
      "Log a set in seconds.",
      "Start from what already worked.",
      "See where training is going.",
    ]);
    expect(stepHeadings[0].closest("li")).toHaveAttribute(
      "aria-current",
      "step",
    );
  });

  it("marks the step crossing the middle of the screen as current, in the list and the rail", () => {
    render(<HowItWorks />);
    const repeatStep = stepItem("Start from what already worked.");

    stepObserver().trigger(repeatStep);

    expect(repeatStep).toHaveAttribute("aria-current", "step");
    expect(
      screen.getByRole("button", { name: "Show step 2: Repeat" }),
    ).toHaveAttribute("aria-current", "step");
    expect(
      screen.getByRole("button", { name: "Show step 1: Log" }),
    ).not.toHaveAttribute("aria-current");
  });

  it("on desktop, plays only the current step's clip and stops everything off screen", () => {
    render(<HowItWorks />);
    expect(playingClips()).toEqual([]);

    sectionObserver().trigger(section());
    expect(playingClips()).toEqual(["/media/how-it-works/log-light.webp"]);

    stepObserver().trigger(stepItem("See where training is going."));
    expect(playingClips()).toEqual(["/media/how-it-works/review-light.webp"]);

    sectionObserver().trigger(section(), false);
    expect(playingClips()).toEqual([]);
  });

  it("below the desktop breakpoint, plays the clip inside the current step", () => {
    media.desktop = false;
    render(<HowItWorks />);

    sectionObserver().trigger(section());

    const [video] = playing;
    expect(playingClips()).toEqual(["/media/how-it-works/log-light.webp"]);
    expect(video.closest("li")).toBe(stepItem("Log a set in seconds."));
  });

  it("lets visitors pause and resume the looping clips", async () => {
    const user = userEvent.setup();
    render(<HowItWorks />);
    sectionObserver().trigger(section());

    await user.click(
      screen.getAllByRole("button", { name: "Pause demo videos" })[0],
    );
    expect(playingClips()).toEqual([]);

    // Staying paused across step changes.
    stepObserver().trigger(stepItem("Start from what already worked."));
    expect(playingClips()).toEqual([]);

    await user.click(
      screen.getAllByRole("button", { name: "Play demo videos" })[0],
    );
    expect(playingClips()).toEqual(["/media/how-it-works/repeat-light.webp"]);
  });

  it("keeps clips on their poster frame when reduced motion is preferred", () => {
    media.reducedMotion = true;
    render(<HowItWorks />);

    sectionObserver().trigger(section());

    expect(playingClips()).toEqual([]);
    expect(
      screen.queryByRole("button", { name: /demo videos/i }),
    ).not.toBeInTheDocument();
    expect(section().querySelector("video")).toHaveAttribute(
      "poster",
      "/media/how-it-works/log-light.webp",
    );
  });

  it("keeps the poster visible when the browser refuses autoplay", () => {
    vi.spyOn(HTMLMediaElement.prototype, "play").mockRejectedValue(
      new DOMException("Autoplay blocked", "NotAllowedError"),
    );
    render(<HowItWorks />);

    expect(() => sectionObserver().trigger(section())).not.toThrow();
    expect(section().querySelector("video")).toHaveAttribute(
      "poster",
      "/media/how-it-works/log-light.webp",
    );
  });
});
