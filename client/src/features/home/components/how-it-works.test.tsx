import { act, render, screen, within } from "@testing-library/react";
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

function stepObserver() {
  const observer = MockIntersectionObserver.instances.find(
    (instance) => instance.options?.rootMargin,
  );
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

function playedClips(play: ReturnType<typeof vi.fn>) {
  return play.mock.contexts.map((video) =>
    video instanceof HTMLVideoElement ? video.getAttribute("poster") : null,
  );
}

describe("HowItWorks", () => {
  let play: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    MockIntersectionObserver.instances = [];
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
    play = vi.fn().mockResolvedValue(undefined);
    vi.spyOn(HTMLMediaElement.prototype, "play").mockImplementation(play);
    vi.spyOn(HTMLMediaElement.prototype, "pause").mockImplementation(() => {});
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("lists the log, repeat, and review steps with the first one current", () => {
    render(<HowItWorks />);

    const section = screen.getByRole("region", {
      name: /log it\. repeat it\./i,
    });
    const stepHeadings = within(section).getAllByRole("heading", { level: 3 });

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

  it("makes the step crossing the middle of the screen current", () => {
    render(<HowItWorks />);
    const repeatStep = screen
      .getByRole("heading", { name: "Start from what already worked." })
      .closest("li");
    if (!repeatStep) throw new Error("repeat step missing");

    stepObserver().trigger(repeatStep);

    expect(repeatStep).toHaveAttribute("aria-current", "step");
    expect(
      screen.getByRole("button", { name: "Show step 2: Repeat" }),
    ).toHaveClass("bg-primary");
  });

  it("plays only the current step's clip while the section is on screen", () => {
    const { container } = render(<HowItWorks />);
    expect(play).not.toHaveBeenCalled();

    sectionObserver().trigger(container.querySelector("section")!);
    expect(playedClips(play)).toEqual(["/media/how-it-works/log-light.webp"]);

    const reviewStep = screen
      .getByRole("heading", { name: "See where training is going." })
      .closest("li")!;
    stepObserver().trigger(reviewStep);
    expect(playedClips(play).at(-1)).toBe(
      "/media/how-it-works/review-light.webp",
    );
  });

  it("keeps clips on their poster frame when reduced motion is preferred", () => {
    media.reducedMotion = true;
    const { container } = render(<HowItWorks />);

    sectionObserver().trigger(container.querySelector("section")!);

    expect(play).not.toHaveBeenCalled();
    expect(container.querySelector("video")).toHaveAttribute(
      "poster",
      "/media/how-it-works/log-light.webp",
    );
  });
});
