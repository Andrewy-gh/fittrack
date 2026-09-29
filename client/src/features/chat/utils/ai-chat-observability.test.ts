import { describe, expect, it } from "vitest";

import {
  classifyLoadOutcome,
  classifyStreamInterruption,
} from "./ai-chat-observability";

describe("ai chat observability helpers", () => {
  it("classifies aborts as client_aborted", () => {
    expect(
      classifyStreamInterruption(
        new DOMException("Aborted", "AbortError"),
        true,
      ),
    ).toEqual({
      outcome: "client_aborted",
      stage: "post_start",
    });
  });

  it("classifies stale load aborts distinctly from load failures", () => {
    expect(classifyLoadOutcome(true)).toBe("load_aborted_stale");
    expect(classifyLoadOutcome(false)).toBe("load_failed");
  });
});
