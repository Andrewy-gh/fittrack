import { describe, expect, it } from "vitest";

import {
  classifyLoadOutcome,
  classifyRecoveryOutcome,
  classifyStreamInterruption,
  isRecoveryTimeoutError,
  recoveryTimeoutMessage,
} from "./ai-chat-observability";

describe("ai chat observability helpers", () => {
  it("classifies pre-start transport interruptions separately from aborts", () => {
    expect(
      classifyStreamInterruption(new Error("socket closed"), false),
    ).toEqual({
      outcome: "transport_ended_pre_terminal",
      stage: "pre_start",
    });
  });

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

  it("classifies recovery completion from persisted messages", () => {
    expect(
      classifyRecoveryOutcome({
        prompt: "hello",
        messages: [
          {
            id: 1,
            conversation_id: 41,
            role: "user",
            content: "hello",
            status: "completed",
            created_at: "2026-03-27T00:00:00Z",
            updated_at: "2026-03-27T00:00:00Z",
          },
          {
            id: 2,
            conversation_id: 41,
            role: "assistant",
            content: "Recovered answer",
            status: "completed",
            created_at: "2026-03-27T00:00:00Z",
            updated_at: "2026-03-27T00:00:01Z",
          },
        ],
      }),
    ).toBe("recovered_completed");
  });

  it("classifies recovery timeouts without counting them as aborts", () => {
    const error = new Error(recoveryTimeoutMessage);

    expect(isRecoveryTimeoutError(error)).toBe(true);
    expect(classifyRecoveryOutcome({ error })).toBe("recovery_timeout");
  });

  it("classifies stale load aborts distinctly from load failures", () => {
    expect(classifyLoadOutcome(true)).toBe("load_aborted_stale");
    expect(classifyLoadOutcome(false)).toBe("load_failed");
  });
});
