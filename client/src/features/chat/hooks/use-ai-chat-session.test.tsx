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
  mockStreamMessage: vi.fn(),
  mockShowErrorToast: vi.fn(),
  mockToastSuccess: vi.fn(),
}));

vi.mock("@/lib/api/ai-chat", () => ({
  createAIChatConversation: mockCreateConversation,
  getAIChatConversation: mockGetConversation,
  pollAIChatConversationUntilSettled: mockPollConversation,
  resumeAIChatMessageStream: mockResumeStream,
  reportAIChatTelemetry: mockReportTelemetry,
  requestAIChatMessageRecovery: mockRequestRecovery,
  saveAIChatLatestWorkoutDraft: mockSaveLatestWorkoutDraft,
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
});
