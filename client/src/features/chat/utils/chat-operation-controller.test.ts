import { beforeEach, describe, expect, it, vi } from "vitest";
import type { ChatSessionState } from "./chat-session-types";

const { mockGetConversation, mockReportTelemetry, mockStopRun } = vi.hoisted(
  () => ({
    mockGetConversation: vi.fn(),
    mockReportTelemetry: vi.fn(),
    mockStopRun: vi.fn(),
  }),
);

vi.mock("@/features/chat/api/ai-chat", () => ({
  getAIChatConversation: mockGetConversation,
  reportAIChatTelemetry: mockReportTelemetry,
  stopAIChatRun: mockStopRun,
}));

import { ChatOperationController } from "./chat-operation-controller";

function createController() {
  const state: ChatSessionState = {
    setConversation: vi.fn(),
    setMessages: vi.fn(),
    setPrompt: vi.fn(),
    setIsLoadingConversation: vi.fn(),
    setLoadError: vi.fn(),
    setIsSavingWorkoutDraft: vi.fn(),
    setLatestWorkoutDraftMessageId: vi.fn(),
  };
  const controller = new ChatOperationController({
    state,
    onConversationCreated: vi.fn().mockResolvedValue(undefined),
    onPromptStarted: vi.fn(),
    onNewConversationCreated: vi.fn(),
  });
  return { controller, state };
}

describe("ChatOperationController", () => {
  beforeEach(() => {
    mockGetConversation.mockReset();
    mockReportTelemetry.mockReset().mockResolvedValue(undefined);
    mockStopRun.mockReset();
  });

  it("publishes cached snapshots for starting and streaming transitions", () => {
    const { controller } = createController();
    controller.attach();
    const listener = vi.fn();
    controller.subscribe(listener);

    const idle = controller.getSnapshot();
    expect(controller.getSnapshot()).toBe(idle);
    const operation = controller.beginOperation(null);
    expect(operation).not.toBeNull();
    expect(controller.getSnapshot()).toEqual({
      phase: "starting",
      conversationId: null,
      runId: null,
      assistantMessageId: null,
    });
    expect(controller.getSnapshot()).toBe(controller.getSnapshot());

    const attempt = controller.beginAttempt(operation!, "stream");
    expect(attempt).not.toBeNull();
    controller.markStreaming(operation!, attempt!, 91, 61);
    expect(controller.getSnapshot()).toEqual({
      phase: "streaming",
      conversationId: null,
      runId: 91,
      assistantMessageId: 61,
    });
    expect(listener).toHaveBeenCalledTimes(2);
  });

  it("uses attempt identity so late cleanup cannot finish a replacement", () => {
    const { controller } = createController();
    controller.attach();
    const operation = controller.beginOperation(41)!;
    const stream = controller.beginAttempt(operation, "stream")!;
    const recovery = controller.beginAttempt(operation, "recovery")!;

    expect(stream.controller.signal.aborted).toBe(true);
    expect(controller.finishAttempt(operation, stream)).toBe(false);
    expect(controller.ownsAttempt(operation, recovery)).toBe(true);
    expect(controller.getSnapshot().phase).toBe("recovering");
  });

  it("adopts the created route but detaches for different navigation", async () => {
    mockGetConversation.mockResolvedValue({
      conversation: { id: 42 },
      messages: [],
      active_run: null,
    });
    const { controller } = createController();
    controller.attach();
    const operation = controller.beginOperation(null)!;
    const attempt = controller.beginAttempt(operation, "stream")!;

    expect(controller.adoptConversation(operation, 41)).toBe(true);
    controller.handleRoute(41);
    expect(controller.ownsAttempt(operation, attempt)).toBe(true);
    expect(mockGetConversation).not.toHaveBeenCalled();

    controller.handleRoute(42);
    expect(attempt.controller.signal.aborted).toBe(true);
    expect(controller.getSnapshot().phase).toBe("idle");
    expect(mockGetConversation).toHaveBeenCalledWith(
      42,
      expect.objectContaining({ signal: expect.any(AbortSignal) }),
    );
  });

  it("invalidates old callbacks across Strict Mode attachment replay", () => {
    const { controller } = createController();
    const firstAttachment = controller.attach();
    const oldOperation = controller.beginOperation(41)!;
    const oldAttempt = controller.beginAttempt(oldOperation, "stream")!;

    controller.detach(firstAttachment);
    const secondAttachment = controller.attach();
    const currentOperation = controller.beginOperation(42)!;
    const currentAttempt = controller.beginAttempt(currentOperation, "stream")!;

    expect(oldAttempt.controller.signal.aborted).toBe(true);
    expect(controller.markStreaming(oldOperation, oldAttempt, 91, 61)).toBe(
      false,
    );
    expect(controller.ownsAttempt(currentOperation, currentAttempt)).toBe(true);
    expect(secondAttachment.active).toBe(true);
  });

  it("preserves visible Stop timing until the existing request resolves", async () => {
    let resolveStop!: (value: {
      conversation_id: number;
      run_id: number;
      message_id: number;
      status: "stopped";
      text: string;
      sequence: number;
    }) => void;
    mockStopRun.mockReturnValue(
      new Promise((resolve) => {
        resolveStop = resolve;
      }),
    );
    const { controller, state } = createController();
    controller.attach();
    const operation = controller.beginOperation(41)!;
    const attempt = controller.beginAttempt(operation, "stream")!;
    controller.markStreaming(operation, attempt, 91, 61);

    const stopPromise = controller.stopRun();
    expect(mockStopRun).toHaveBeenCalledWith(41, 91);
    expect(controller.getSnapshot().phase).toBe("streaming");
    expect(attempt.controller.signal.aborted).toBe(false);

    resolveStop({
      conversation_id: 41,
      run_id: 91,
      message_id: 61,
      status: "stopped",
      text: "partial response",
      sequence: 2,
    });
    await stopPromise;

    expect(attempt.controller.signal.aborted).toBe(true);
    expect(controller.getSnapshot().phase).toBe("idle");
    expect(state.setMessages).toHaveBeenCalledTimes(1);
  });
});
