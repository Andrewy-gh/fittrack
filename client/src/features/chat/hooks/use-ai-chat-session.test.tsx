import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

const {
  mockCreateConversation,
  mockGetConversation,
  mockPollConversation,
  mockResumeStream,
  mockReportTelemetry,
  mockRequestRecovery,
  mockSaveLatestWorkoutDraft,
  mockStopRun,
  mockStreamMessage,
  mockShowErrorToast,
  mockToastSuccess,
} = vi.hoisted(() => ({
  mockCreateConversation: vi.fn(),
  mockGetConversation: vi.fn(),
  mockPollConversation: vi.fn(),
  mockResumeStream: vi.fn(),
  mockReportTelemetry: vi.fn(),
  mockRequestRecovery: vi.fn(),
  mockSaveLatestWorkoutDraft: vi.fn(),
  mockStopRun: vi.fn(),
  mockStreamMessage: vi.fn(),
  mockShowErrorToast: vi.fn(),
  mockToastSuccess: vi.fn(),
}));

vi.mock("@/features/chat/api/ai-chat", () => ({
  createAIChatConversation: mockCreateConversation,
  getAIChatConversation: mockGetConversation,
  pollAIChatConversationUntilSettled: mockPollConversation,
  resumeAIChatMessageStream: mockResumeStream,
  reportAIChatTelemetry: mockReportTelemetry,
  requestAIChatMessageRecovery: mockRequestRecovery,
  saveAIChatLatestWorkoutDraft: mockSaveLatestWorkoutDraft,
  stopAIChatRun: mockStopRun,
  streamAIChatMessage: mockStreamMessage,
}));

vi.mock("@/lib/errors", () => ({
  getErrorMessage: (
    error: unknown,
    fallback = "An unexpected error occurred",
  ) => (error instanceof Error ? error.message : fallback),
  showErrorToast: mockShowErrorToast,
}));

vi.mock("sonner", () => ({
  toast: {
    success: mockToastSuccess,
  },
}));

import { useAIChatSession } from "@/features/chat/hooks/use-ai-chat-session";
import type {
  AIChatStreamDeltaEvent,
  AIChatStreamDoneEvent,
  AIChatStreamErrorEvent,
  AIChatStreamStartEvent,
} from "@/features/chat/api/ai-chat";

type ResumeStreamHandlers = {
  onStart?: (event: AIChatStreamStartEvent) => void;
  onDelta?: (event: AIChatStreamDeltaEvent) => void;
  onDone?: (event: AIChatStreamDoneEvent) => void;
  onErrorEvent?: (event: AIChatStreamErrorEvent) => void;
};

function conversationDetail(
  messages: Array<Record<string, unknown>>,
  activeRun?: Record<string, unknown>,
) {
  return {
    conversation: {
      id: 41,
      created_at: "2026-03-26T17:00:00Z",
      updated_at: "2026-03-26T17:00:00Z",
    },
    messages,
    active_run: activeRun,
  };
}

function renderSession(conversationId: number | null = 41) {
  return renderHook(
    ({ id }) =>
      useAIChatSession({
        conversationId: id,
        initialPrompt: "",
        onPromptChange: vi.fn(),
        onPromptStarted: vi.fn(),
        onNewConversationCreated: vi.fn(),
        onConversationCreated: vi.fn().mockResolvedValue(undefined),
      }),
    {
      initialProps: { id: conversationId },
    },
  );
}

describe("useAIChatSession", () => {
  beforeEach(() => {
    window.sessionStorage.clear();
    mockCreateConversation.mockReset();
    mockGetConversation.mockReset();
    mockPollConversation.mockReset();
    mockResumeStream.mockReset();
    mockReportTelemetry.mockReset();
    mockRequestRecovery.mockReset();
    mockSaveLatestWorkoutDraft.mockReset();
    mockStopRun.mockReset();
    mockStreamMessage.mockReset();
    mockShowErrorToast.mockReset();
    mockToastSuccess.mockReset();
    mockReportTelemetry.mockResolvedValue(undefined);
    mockGetConversation.mockResolvedValue(conversationDetail([]));
    mockRequestRecovery.mockResolvedValue({
      conversation_id: 41,
      run_id: 61,
      status: "queued",
    });
  });

  it("creates optimistic user and assistant messages, then completes from a done stream event", async () => {
    mockGetConversation
      .mockResolvedValueOnce(conversationDetail([]))
      .mockResolvedValueOnce(
        conversationDetail([
          {
            id: 71,
            conversation_id: 41,
            role: "user",
            content: "hello",
            status: "completed",
            created_at: "2026-03-26T17:00:01Z",
            updated_at: "2026-03-26T17:00:01Z",
            completed_at: "2026-03-26T17:00:01Z",
          },
          {
            id: 72,
            conversation_id: 41,
            role: "assistant",
            content: "Completed answer",
            status: "completed",
            created_at: "2026-03-26T17:00:01Z",
            updated_at: "2026-03-26T17:00:02Z",
            completed_at: "2026-03-26T17:00:02Z",
          },
        ]),
      );
    mockStreamMessage.mockImplementation(
      async (
        _conversationId: number,
        _prompt: string,
        options?: {
          onStart?: (event: Record<string, unknown>) => void;
          onDelta?: (event: Record<string, unknown>) => void;
          onDone?: (event: Record<string, unknown>) => void;
        },
      ) => {
        options?.onStart?.({ type: "start", message_id: 72, sequence: 1 });
        options?.onDelta?.({ type: "delta", delta: "Partial ", sequence: 2 });
        options?.onDone?.({
          type: "done",
          message_id: 72,
          text: "Completed answer",
          sequence: 3,
        });

        return {
          doneEvent: {
            type: "done",
            message_id: 72,
            text: "Completed answer",
          },
          endedWithError: false,
        };
      },
    );

    const { result } = renderSession();

    await waitFor(() => {
      expect(result.current.isLoadingConversation).toBe(false);
    });

    act(() => {
      result.current.setPrompt("hello");
    });
    await act(async () => {
      await result.current.submitPrompt();
    });

    expect(result.current.messages).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          role: "user",
          content: "hello",
          status: "completed",
        }),
        expect.objectContaining({
          id: 72,
          role: "assistant",
          content: "Completed answer",
          status: "completed",
        }),
      ]),
    );
    expect(mockReportTelemetry).toHaveBeenCalledWith({
      category: "stream",
      outcome: "completed",
      stage: "terminal",
    });
  });

  it("does not let an old terminal refresh overwrite a newly submitted stream", async () => {
    let resolveOldRefresh:
      | ((value: ReturnType<typeof conversationDetail>) => void)
      | undefined;
    let resolveNewStream:
      | ((value: { doneEvent: null; endedWithError: false }) => void)
      | undefined;
    mockGetConversation
      .mockResolvedValueOnce(conversationDetail([]))
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveOldRefresh = resolve;
          }),
      );
    mockStreamMessage
      .mockImplementationOnce(async (_conversationId, _prompt, options) => {
        options?.onStart?.({ type: "start", run_id: 91, message_id: 71 });
        options?.onDone?.({
          type: "done",
          run_id: 91,
          message_id: 71,
          text: "First answer",
        });
        return { doneEvent: null, endedWithError: false };
      })
      .mockImplementationOnce((_conversationId, _prompt, options) => {
        options?.onStart?.({ type: "start", run_id: 92, message_id: 72 });
        options?.onDelta?.({
          type: "delta",
          run_id: 92,
          message_id: 72,
          delta: "New partial",
        });
        return new Promise((resolve) => {
          resolveNewStream = resolve;
        });
      });

    const { result } = renderSession();
    await waitFor(() => {
      expect(result.current.isLoadingConversation).toBe(false);
    });

    act(() => {
      result.current.setPrompt("first prompt");
    });
    let firstSubmit: Promise<void> | undefined;
    act(() => {
      firstSubmit = result.current.submitPrompt();
    });
    await waitFor(() => {
      expect(mockGetConversation).toHaveBeenCalledTimes(2);
      expect(result.current.isSubmitting).toBe(false);
    });

    act(() => {
      result.current.setPrompt("second prompt");
    });
    let secondSubmit: Promise<void> | undefined;
    act(() => {
      secondSubmit = result.current.submitPrompt();
    });
    await waitFor(() => {
      expect(result.current.messages.at(-1)).toEqual(
        expect.objectContaining({
          id: 72,
          content: "New partial",
          status: "streaming",
        }),
      );
    });

    await act(async () => {
      resolveOldRefresh?.(
        conversationDetail([
          {
            id: 71,
            conversation_id: 41,
            role: "assistant",
            content: "Stale first answer",
            status: "completed",
            created_at: "2026-03-26T17:00:01Z",
            updated_at: "2026-03-26T17:00:02Z",
            completed_at: "2026-03-26T17:00:02Z",
          },
        ]),
      );
      await firstSubmit;
    });

    expect(result.current.isSubmitting).toBe(true);
    expect(result.current.messages.at(-1)).toEqual(
      expect.objectContaining({
        id: 72,
        content: "New partial",
        status: "streaming",
      }),
    );
    expect(result.current.messages).not.toEqual(
      expect.arrayContaining([
        expect.objectContaining({ content: "Stale first answer" }),
      ]),
    );

    await act(async () => {
      resolveNewStream?.({ doneEvent: null, endedWithError: false });
      await secondSubmit;
    });
  });

  it("recovers an interrupted stream and suppresses the failure toast when recovery succeeds", async () => {
    mockStreamMessage.mockRejectedValue(
      new Error("AI chat stream ended before a terminal event"),
    );
    mockPollConversation.mockResolvedValue(
      conversationDetail([
        {
          id: 71,
          conversation_id: 41,
          role: "user",
          content: "hello",
          status: "completed",
          created_at: "2026-03-26T17:00:01Z",
          updated_at: "2026-03-26T17:00:01Z",
          completed_at: "2026-03-26T17:00:01Z",
        },
        {
          id: 72,
          conversation_id: 41,
          role: "assistant",
          content: "Recovered answer",
          status: "completed",
          created_at: "2026-03-26T17:00:01Z",
          updated_at: "2026-03-26T17:00:02Z",
          completed_at: "2026-03-26T17:00:02Z",
        },
      ]),
    );
    const { result } = renderSession();

    await waitFor(() => {
      expect(result.current.isLoadingConversation).toBe(false);
    });

    act(() => {
      result.current.setPrompt("hello");
    });
    await act(async () => {
      await result.current.submitPrompt();
    });

    expect(result.current.messages.at(-1)?.content).toBe("Recovered answer");
    expect(mockShowErrorToast).not.toHaveBeenCalled();
    expect(mockReportTelemetry).toHaveBeenCalledWith({
      category: "ux",
      outcome: "failure_toast_suppressed_due_to_successful_recovery",
    });
  });

  it("treats a persisted stopped response as successful recovery", async () => {
    mockStreamMessage.mockRejectedValue(
      new Error("AI chat stream ended before a terminal event"),
    );
    mockPollConversation.mockResolvedValue(
      conversationDetail([
        {
          id: 71,
          conversation_id: 41,
          role: "user",
          content: "hello",
          status: "completed",
          created_at: "2026-03-26T17:00:01Z",
          updated_at: "2026-03-26T17:00:01Z",
          completed_at: "2026-03-26T17:00:01Z",
        },
        {
          id: 72,
          conversation_id: 41,
          role: "assistant",
          content: "Partial answer",
          status: "stopped",
          created_at: "2026-03-26T17:00:01Z",
          updated_at: "2026-03-26T17:00:02Z",
          completed_at: "2026-03-26T17:00:02Z",
        },
      ]),
    );
    const { result } = renderSession();

    await waitFor(() => {
      expect(result.current.isLoadingConversation).toBe(false);
    });

    act(() => {
      result.current.setPrompt("hello");
    });
    await act(async () => {
      await result.current.submitPrompt();
    });

    expect(result.current.messages.at(-1)).toEqual(
      expect.objectContaining({
        content: "Partial answer",
        status: "stopped",
      }),
    );
    expect(mockShowErrorToast).not.toHaveBeenCalled();
    expect(mockReportTelemetry).toHaveBeenCalledWith({
      category: "recovery",
      outcome: "recovered_completed",
    });
    expect(mockReportTelemetry).toHaveBeenCalledWith({
      category: "ux",
      outcome: "failure_toast_suppressed_due_to_successful_recovery",
    });
  });

  it("removes optimistic messages and restores the prompt after a pre-start server error", async () => {
    mockStreamMessage.mockRejectedValue({
      message: "ai chat runtime is not configured",
      request_id: "req-123",
    });
    const { result } = renderSession();

    await waitFor(() => {
      expect(result.current.isLoadingConversation).toBe(false);
    });

    act(() => {
      result.current.setPrompt("hello");
    });
    await act(async () => {
      await result.current.submitPrompt();
    });

    expect(result.current.messages).toEqual([]);
    expect(result.current.prompt).toBe("hello");
    expect(mockRequestRecovery).not.toHaveBeenCalled();
    expect(mockShowErrorToast).toHaveBeenCalledWith(
      expect.objectContaining({
        message: "ai chat runtime is not configured",
      }),
      "Failed to stream AI chat response",
    );
  });

  it("attempts active-run resume before recovery polling fallback", async () => {
    mockGetConversation.mockResolvedValue(
      conversationDetail(
        [
          {
            id: 61,
            conversation_id: 41,
            role: "assistant",
            content: "partial",
            status: "streaming",
            created_at: "2026-03-26T17:00:00Z",
            updated_at: "2026-03-26T17:00:01Z",
          },
        ],
        {
          id: 91,
          assistant_message_id: 61,
          status: "streaming",
          latest_sequence: 1,
        },
      ),
    );
    mockResumeStream.mockRejectedValue({ message: "resume unavailable" });
    mockPollConversation.mockResolvedValue(
      conversationDetail([
        {
          id: 61,
          conversation_id: 41,
          role: "assistant",
          content: "Recovered answer",
          status: "completed",
          created_at: "2026-03-26T17:00:00Z",
          updated_at: "2026-03-26T17:00:02Z",
          completed_at: "2026-03-26T17:00:02Z",
        },
      ]),
    );

    const { result } = renderSession();

    await waitFor(() => {
      expect(result.current.messages.at(-1)?.content).toBe("Recovered answer");
    });

    expect(mockResumeStream).toHaveBeenCalledWith(
      41,
      91,
      1,
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      }),
    );
    expect(mockResumeStream.mock.invocationCallOrder[0]).toBeLessThan(
      mockRequestRecovery.mock.invocationCallOrder[0],
    );
    expect(mockPollConversation).toHaveBeenCalledWith(
      41,
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      }),
    );
  });

  it.each(["completed", "failed"] as const)(
    "releases a hung active request when Stop reports an already %s run",
    async (status) => {
      mockGetConversation
        .mockResolvedValueOnce(
          conversationDetail([], {
            id: 91,
            assistant_message_id: 61,
            status: "streaming",
            latest_sequence: 1,
          }),
        )
        .mockResolvedValueOnce(
          conversationDetail([
            {
              id: 61,
              conversation_id: 41,
              role: "assistant",
              content: `${status} response`,
              status,
              created_at: "2026-03-26T17:00:00Z",
              updated_at: "2026-03-26T17:00:02Z",
              completed_at: "2026-03-26T17:00:02Z",
            },
          ]),
        );
      mockResumeStream.mockImplementation(() => new Promise(() => undefined));
      mockStopRun.mockResolvedValue({
        conversation_id: 41,
        run_id: 91,
        message_id: 61,
        status,
        text: `${status} response`,
        sequence: 2,
      });

      const { result } = renderSession();

      await waitFor(() => {
        expect(result.current.canStop).toBe(true);
      });

      await act(async () => {
        await result.current.stopRun();
      });

      expect(result.current.isSubmitting).toBe(false);
      expect(result.current.canStop).toBe(false);
      expect(result.current.messages).toEqual([
        expect.objectContaining({
          id: 61,
          content: `${status} response`,
          status,
        }),
      ]);
    },
  );

  it("keeps a newly loaded Stop target when an aborted route resume settles late", async () => {
    const resolveResume: Array<
      (value: { doneEvent: null; endedWithError: false }) => void
    > = [];
    mockResumeStream.mockImplementation(
      () =>
        new Promise((resolve) => {
          resolveResume.push(resolve);
        }),
    );
    mockGetConversation
      .mockResolvedValueOnce(
        conversationDetail([], {
          id: 91,
          assistant_message_id: 61,
          status: "streaming",
          latest_sequence: 1,
        }),
      )
      .mockResolvedValueOnce({
        ...conversationDetail([], {
          id: 92,
          assistant_message_id: 82,
          status: "streaming",
          latest_sequence: 1,
        }),
        conversation: {
          id: 42,
          created_at: "2026-03-26T17:01:00Z",
          updated_at: "2026-03-26T17:01:00Z",
        },
      });

    const { result, rerender } = renderSession();

    await waitFor(() => {
      expect(result.current.canStop).toBe(true);
      expect(resolveResume).toHaveLength(1);
    });

    rerender({ id: 42 });
    await waitFor(() => {
      expect(resolveResume).toHaveLength(2);
      expect(result.current.canStop).toBe(true);
      expect(result.current.isSubmitting).toBe(true);
    });

    await act(async () => {
      resolveResume[0]?.({ doneEvent: null, endedWithError: false });
      await Promise.resolve();
    });

    expect(result.current.canStop).toBe(true);
    expect(result.current.isSubmitting).toBe(true);
    mockStopRun.mockResolvedValue({
      conversation_id: 42,
      run_id: 92,
      message_id: 82,
      status: "stopped",
      text: "new response",
      sequence: 2,
    });
    await act(async () => {
      await result.current.stopRun();
    });
    expect(mockStopRun).toHaveBeenCalledWith(42, 92);
  });

  it("ignores queued resume callbacks after Stop clears operation ownership", async () => {
    let resumeHandlers: ResumeStreamHandlers | undefined;
    mockResumeStream.mockImplementation(
      (
        _conversationId: number,
        _runId: number,
        _afterSequence: number,
        handlers: ResumeStreamHandlers,
      ) => {
        resumeHandlers = handlers;
        return new Promise(() => undefined);
      },
    );
    mockGetConversation.mockResolvedValue(
      conversationDetail(
        [
          {
            id: 61,
            conversation_id: 41,
            role: "assistant",
            content: "partial response",
            status: "streaming",
            created_at: "2026-03-26T17:00:00Z",
            updated_at: "2026-03-26T17:00:01Z",
          },
        ],
        {
          id: 91,
          assistant_message_id: 61,
          status: "streaming",
          latest_sequence: 1,
        },
      ),
    );
    mockStopRun.mockResolvedValue({
      conversation_id: 41,
      run_id: 91,
      message_id: 61,
      status: "stopped",
      text: "authoritative stopped response",
      sequence: 2,
    });

    const { result } = renderSession();
    await waitFor(() => {
      expect(resumeHandlers).toBeDefined();
      expect(result.current.canStop).toBe(true);
    });

    await act(async () => {
      await result.current.stopRun();
    });

    const assertStoppedStateIsUnchanged = () => {
      expect(result.current.messages).toEqual([
        expect.objectContaining({
          id: 61,
          content: "authoritative stopped response",
          status: "stopped",
        }),
      ]);
      expect(result.current.conversation?.latest_workout_draft).toBeUndefined();
      expect(result.current.latestWorkoutDraftMessageId).toBeNull();
      expect(
        window.sessionStorage.getItem("fittrack.ai-chat.resume:41"),
      ).toBeNull();
    };

    assertStoppedStateIsUnchanged();
    act(() => {
      resumeHandlers?.onStart?.({
        type: "start",
        run_id: 91,
        message_id: 161,
        sequence: 3,
      });
    });
    assertStoppedStateIsUnchanged();
    act(() => {
      resumeHandlers?.onDelta?.({
        type: "delta",
        run_id: 91,
        message_id: 61,
        delta: " stale delta",
        sequence: 4,
      });
    });
    assertStoppedStateIsUnchanged();
    act(() => {
      resumeHandlers?.onDone?.({
        type: "done",
        run_id: 91,
        message_id: 61,
        text: "stale completion",
        sequence: 5,
        workout_draft: {
          date: "2026-03-26",
          exercises: [],
        },
      });
    });
    assertStoppedStateIsUnchanged();
    act(() => {
      resumeHandlers?.onErrorEvent?.({
        type: "error",
        run_id: 91,
        message_id: 61,
        message: "stale failure",
        sequence: 6,
      });
    });
    assertStoppedStateIsUnchanged();
  });

  it("ignores queued resume callbacks after a new operation takes ownership", async () => {
    const resumeHandlers: ResumeStreamHandlers[] = [];
    mockResumeStream.mockImplementation(
      (
        _conversationId: number,
        _runId: number,
        _afterSequence: number,
        handlers: ResumeStreamHandlers,
      ) => {
        resumeHandlers.push(handlers);
        return new Promise(() => undefined);
      },
    );
    mockGetConversation
      .mockResolvedValueOnce(
        conversationDetail(
          [
            {
              id: 61,
              conversation_id: 41,
              role: "assistant",
              content: "old partial response",
              status: "streaming",
              created_at: "2026-03-26T17:00:00Z",
              updated_at: "2026-03-26T17:00:01Z",
            },
          ],
          {
            id: 91,
            assistant_message_id: 61,
            status: "streaming",
            latest_sequence: 1,
          },
        ),
      )
      .mockResolvedValueOnce({
        ...conversationDetail(
          [
            {
              id: 82,
              conversation_id: 42,
              role: "assistant",
              content: "new partial response",
              status: "streaming",
              created_at: "2026-03-26T17:01:00Z",
              updated_at: "2026-03-26T17:01:01Z",
            },
          ],
          {
            id: 92,
            assistant_message_id: 82,
            status: "streaming",
            latest_sequence: 1,
          },
        ),
        conversation: {
          id: 42,
          created_at: "2026-03-26T17:01:00Z",
          updated_at: "2026-03-26T17:01:00Z",
        },
      });

    const { result, rerender } = renderSession();
    await waitFor(() => {
      expect(resumeHandlers).toHaveLength(1);
      expect(result.current.canStop).toBe(true);
    });

    rerender({ id: 42 });
    await waitFor(() => {
      expect(resumeHandlers).toHaveLength(2);
      expect(result.current.messages.at(-1)?.content).toBe(
        "new partial response",
      );
      expect(result.current.canStop).toBe(true);
    });

    const assertNewOperationStateIsUnchanged = () => {
      expect(result.current.messages).toEqual([
        expect.objectContaining({
          id: 82,
          content: "new partial response",
          status: "streaming",
        }),
      ]);
      expect(result.current.conversation?.id).toBe(42);
      expect(result.current.conversation?.latest_workout_draft).toBeUndefined();
      expect(result.current.latestWorkoutDraftMessageId).toBeNull();
      expect(
        window.sessionStorage.getItem("fittrack.ai-chat.resume:41"),
      ).toBeNull();
    };

    const oldHandlers = resumeHandlers[0];
    assertNewOperationStateIsUnchanged();
    act(() => {
      oldHandlers?.onStart?.({
        type: "start",
        run_id: 91,
        message_id: 161,
        sequence: 2,
      });
    });
    assertNewOperationStateIsUnchanged();
    act(() => {
      oldHandlers?.onDelta?.({
        type: "delta",
        run_id: 91,
        message_id: 61,
        delta: " stale delta",
        sequence: 3,
      });
    });
    assertNewOperationStateIsUnchanged();
    act(() => {
      oldHandlers?.onDone?.({
        type: "done",
        run_id: 91,
        message_id: 61,
        text: "stale completion",
        sequence: 4,
        workout_draft: {
          date: "2026-03-26",
          exercises: [],
        },
      });
    });
    assertNewOperationStateIsUnchanged();
    act(() => {
      oldHandlers?.onErrorEvent?.({
        type: "error",
        run_id: 91,
        message_id: 61,
        message: "stale failure",
        sequence: 5,
      });
    });
    assertNewOperationStateIsUnchanged();
  });

  it("keeps a newly loaded Stop target when an aborted submit settles late", async () => {
    let resolveSubmit:
      | ((value: { doneEvent: null; endedWithError: false }) => void)
      | undefined;
    mockStreamMessage.mockImplementation(
      (
        _conversationId: number,
        _prompt: string,
        options?: {
          onStart?: (event: Record<string, unknown>) => void;
        },
      ) => {
        options?.onStart?.({
          type: "start",
          run_id: 91,
          message_id: 61,
          sequence: 1,
        });
        return new Promise((resolve) => {
          resolveSubmit = resolve;
        });
      },
    );
    mockResumeStream.mockImplementation(() => new Promise(() => undefined));
    mockGetConversation
      .mockResolvedValueOnce(conversationDetail([]))
      .mockResolvedValueOnce({
        ...conversationDetail([], {
          id: 92,
          assistant_message_id: 82,
          status: "streaming",
          latest_sequence: 1,
        }),
        conversation: {
          id: 42,
          created_at: "2026-03-26T17:01:00Z",
          updated_at: "2026-03-26T17:01:00Z",
        },
      });

    const { result, rerender } = renderSession();
    await waitFor(() => {
      expect(result.current.isLoadingConversation).toBe(false);
    });

    act(() => {
      result.current.setPrompt("hello");
    });
    let submitPromise: Promise<void> | undefined;
    act(() => {
      submitPromise = result.current.submitPrompt();
    });
    await waitFor(() => {
      expect(result.current.canStop).toBe(true);
    });

    rerender({ id: 42 });
    await waitFor(() => {
      expect(mockResumeStream).toHaveBeenCalled();
      expect(result.current.canStop).toBe(true);
      expect(result.current.isSubmitting).toBe(true);
    });

    await act(async () => {
      resolveSubmit?.({ doneEvent: null, endedWithError: false });
      await submitPromise;
    });

    expect(result.current.canStop).toBe(true);
    expect(result.current.isSubmitting).toBe(true);
    mockStopRun.mockResolvedValue({
      conversation_id: 42,
      run_id: 92,
      message_id: 82,
      status: "stopped",
      text: "new response",
      sequence: 2,
    });
    await act(async () => {
      await result.current.stopRun();
    });
    expect(mockStopRun).toHaveBeenCalledWith(42, 92);
  });

  it("ignores a Stop response after navigating to a different active run", async () => {
    let resolveStop: ((value: Record<string, unknown>) => void) | undefined;
    mockStopRun.mockImplementation(
      () =>
        new Promise((resolve) => {
          resolveStop = resolve;
        }),
    );
    mockGetConversation
      .mockResolvedValueOnce(
        conversationDetail([], {
          id: 91,
          assistant_message_id: 61,
          status: "streaming",
          latest_sequence: 1,
        }),
      )
      .mockResolvedValueOnce({
        ...conversationDetail(
          [
            {
              id: 82,
              conversation_id: 42,
              role: "assistant",
              content: "new response",
              status: "streaming",
              created_at: "2026-03-26T17:01:00Z",
              updated_at: "2026-03-26T17:01:01Z",
            },
          ],
          {
            id: 92,
            assistant_message_id: 82,
            status: "streaming",
            latest_sequence: 1,
          },
        ),
        conversation: {
          id: 42,
          created_at: "2026-03-26T17:01:00Z",
          updated_at: "2026-03-26T17:01:00Z",
        },
      });
    mockResumeStream.mockImplementation(() => new Promise(() => undefined));

    const { result, rerender } = renderSession();

    await waitFor(() => {
      expect(result.current.canStop).toBe(true);
    });

    let stopPromise: Promise<void> | undefined;
    act(() => {
      stopPromise = result.current.stopRun();
    });
    expect(mockStopRun).toHaveBeenCalledWith(41, 91);

    rerender({ id: 42 });
    await waitFor(() => {
      expect(mockResumeStream).toHaveBeenCalledTimes(2);
      expect(result.current.canStop).toBe(true);
      expect(result.current.isSubmitting).toBe(true);
    });
    const newRunSignal = mockResumeStream.mock.calls[1]?.[3]?.signal;
    expect(newRunSignal).toBeInstanceOf(AbortSignal);
    expect(newRunSignal.aborted).toBe(false);

    await act(async () => {
      resolveStop?.({
        conversation_id: 41,
        run_id: 91,
        message_id: 61,
        status: "stopped",
        text: "old response",
        sequence: 2,
      });
      await stopPromise;
    });

    expect(newRunSignal.aborted).toBe(false);
    expect(result.current.isSubmitting).toBe(true);
    expect(result.current.messages).toEqual([
      expect.objectContaining({
        id: 82,
        content: "new response",
        status: "streaming",
      }),
    ]);
  });

  it("ignores a rejected Stop request after navigating to a different active run", async () => {
    let rejectStop: ((reason?: unknown) => void) | undefined;
    mockStopRun.mockImplementation(
      () =>
        new Promise((_resolve, reject) => {
          rejectStop = reject;
        }),
    );
    mockGetConversation
      .mockResolvedValueOnce(
        conversationDetail([], {
          id: 91,
          assistant_message_id: 61,
          status: "streaming",
          latest_sequence: 1,
        }),
      )
      .mockResolvedValueOnce({
        ...conversationDetail(
          [
            {
              id: 82,
              conversation_id: 42,
              role: "assistant",
              content: "new response",
              status: "streaming",
              created_at: "2026-03-26T17:01:00Z",
              updated_at: "2026-03-26T17:01:01Z",
            },
          ],
          {
            id: 92,
            assistant_message_id: 82,
            status: "streaming",
            latest_sequence: 1,
          },
        ),
        conversation: {
          id: 42,
          created_at: "2026-03-26T17:01:00Z",
          updated_at: "2026-03-26T17:01:00Z",
        },
      });
    mockResumeStream.mockImplementation(() => new Promise(() => undefined));

    const { result, rerender } = renderSession();

    await waitFor(() => {
      expect(result.current.canStop).toBe(true);
    });

    let stopPromise: Promise<void> | undefined;
    act(() => {
      stopPromise = result.current.stopRun();
    });

    rerender({ id: 42 });
    await waitFor(() => {
      expect(result.current.canStop).toBe(true);
      expect(result.current.isSubmitting).toBe(true);
    });

    await act(async () => {
      rejectStop?.(new Error("old Stop request failed"));
      await stopPromise;
    });

    expect(mockShowErrorToast).not.toHaveBeenCalled();
    expect(result.current.isSubmitting).toBe(true);
  });

  it("reports a rejected Stop request for the current active run", async () => {
    const error = new Error("Stop request failed");
    mockStopRun.mockRejectedValue(error);
    mockGetConversation.mockResolvedValue(
      conversationDetail([], {
        id: 91,
        assistant_message_id: 61,
        status: "streaming",
        latest_sequence: 1,
      }),
    );
    mockResumeStream.mockImplementation(() => new Promise(() => undefined));

    const { result } = renderSession();

    await waitFor(() => {
      expect(result.current.canStop).toBe(true);
    });

    await act(async () => {
      await result.current.stopRun();
    });

    expect(mockShowErrorToast).toHaveBeenCalledWith(
      error,
      "Failed to stop AI chat response",
    );
  });
});
