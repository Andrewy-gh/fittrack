import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { AIWorkoutDraft } from "@/lib/api/ai-chat";

const {
  mockSearch,
  mockNavigate,
  mockCreateConversation,
  mockGetConversation,
  mockPollConversation,
  mockResumeStream,
  mockReportTelemetry,
  mockRequestRecovery,
  mockStreamMessage,
  mockShowErrorToast,
} = vi.hoisted(() => ({
  mockSearch: { conversationId: "41" as string | undefined },
  mockNavigate: vi.fn(),
  mockCreateConversation: vi.fn(),
  mockGetConversation: vi.fn(),
  mockPollConversation: vi.fn(),
  mockResumeStream: vi.fn(),
  mockReportTelemetry: vi.fn(),
  mockRequestRecovery: vi.fn(),
  mockStreamMessage: vi.fn(),
  mockShowErrorToast: vi.fn(),
}));

vi.mock("@tanstack/react-router", () => ({
  createFileRoute: () => () => ({
    useRouteContext: () => ({ user: { id: "user-123" } }),
    useSearch: () => mockSearch,
    fullPath: "/chat",
  }),
  useNavigate: () => mockNavigate,
}));

vi.mock("@/lib/api/ai-chat", () => ({
  createAIChatConversation: mockCreateConversation,
  getAIChatConversation: mockGetConversation,
  pollAIChatConversationUntilSettled: mockPollConversation,
  resumeAIChatMessageStream: mockResumeStream,
  reportAIChatTelemetry: mockReportTelemetry,
  requestAIChatMessageRecovery: mockRequestRecovery,
  streamAIChatMessage: mockStreamMessage,
}));

vi.mock("@/lib/errors", () => ({
  getErrorMessage: (
    error: unknown,
    fallback = "An unexpected error occurred",
  ) => (error instanceof Error ? error.message : fallback),
  showErrorToast: mockShowErrorToast,
}));

import { ChatRouteComponent } from "./chat";

function conversationDetail(
  messages: Array<Record<string, unknown>>,
  activeRun?: Record<string, unknown>,
  latestWorkoutDraft?: AIWorkoutDraft,
) {
  return {
    conversation: {
      id: 41,
      created_at: "2026-03-26T17:00:00Z",
      updated_at: "2026-03-26T17:00:00Z",
      latest_workout_draft: latestWorkoutDraft,
    },
    messages,
    active_run: activeRun,
  };
}

function deferredPromise<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason?: unknown) => void;

  const promise = new Promise<T>((res, rej) => {
    resolve = res;
    reject = rej;
  });

  return { promise, resolve, reject };
}

describe("ChatRouteComponent", () => {
  beforeEach(() => {
    window.sessionStorage.clear();
    window.localStorage.clear();
    mockSearch.conversationId = "41";
    mockNavigate.mockReset();
    mockCreateConversation.mockReset();
    mockGetConversation.mockReset();
    mockPollConversation.mockReset();
    mockResumeStream.mockReset();
    mockReportTelemetry.mockReset();
    mockRequestRecovery.mockReset();
    mockStreamMessage.mockReset();
    mockShowErrorToast.mockReset();
    mockReportTelemetry.mockResolvedValue(undefined);
    mockResumeStream.mockResolvedValue({
      doneEvent: {
        type: "done",
        message_id: 61,
        text: "resumed",
      },
      endedWithError: false,
    });
    mockRequestRecovery.mockResolvedValue({
      conversation_id: 41,
      run_id: 61,
      status: "queued",
    });
  });

  it("recovers a completed reply when the stream dies before the start event reaches the client", async () => {
    const user = userEvent.setup();
    mockGetConversation.mockResolvedValue(conversationDetail([]));
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

    render(<ChatRouteComponent />);

    await user.type(
      await screen.findByPlaceholderText(
        "Ask about training, recovery, exercise choices, or FitTrack usage...",
      ),
      "hello",
    );
    await user.click(screen.getByRole("button", { name: "Send" }));

    expect(await screen.findByText("Recovered answer")).toBeInTheDocument();
    expect(screen.getByText("hello")).toBeInTheDocument();
    expect(mockPollConversation).toHaveBeenCalledWith(
      41,
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      }),
    );
    expect(mockRequestRecovery).toHaveBeenCalledWith(
      41,
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      }),
    );
    expect(mockReportTelemetry).toHaveBeenCalledWith({
      category: "stream",
      outcome: "transport_ended_pre_terminal",
      stage: "pre_start",
    });
    expect(mockReportTelemetry).toHaveBeenCalledWith({
      category: "recovery",
      outcome: "recovered_completed",
    });
    expect(mockReportTelemetry).toHaveBeenCalledWith({
      category: "ux",
      outcome: "failure_toast_suppressed_due_to_successful_recovery",
    });
    expect(mockShowErrorToast).not.toHaveBeenCalled();
  });

  it("does not treat preflight api failures as transport interruptions", async () => {
    const user = userEvent.setup();
    mockGetConversation.mockResolvedValue(conversationDetail([]));
    mockStreamMessage.mockRejectedValue({
      message: "ai chat runtime is not configured",
      request_id: "req-123",
    });

    render(<ChatRouteComponent />);

    const promptBox = await screen.findByPlaceholderText(
      "Ask about training, recovery, exercise choices, or FitTrack usage...",
    );
    await user.type(promptBox, "hello");
    await user.click(screen.getByRole("button", { name: "Send" }));

    await waitFor(() => {
      expect(mockShowErrorToast).toHaveBeenCalledWith(
        expect.objectContaining({
          message: "ai chat runtime is not configured",
        }),
        "Failed to stream AI chat response",
      );
    });
    expect(mockRequestRecovery).not.toHaveBeenCalled();
    expect(mockPollConversation).not.toHaveBeenCalled();
    expect(mockReportTelemetry).toHaveBeenCalledWith({
      category: "stream",
      outcome: "server_error",
      stage: "pre_start",
    });
    expect(mockReportTelemetry).toHaveBeenCalledWith({
      category: "ux",
      outcome: "failure_toast_shown",
    });
    expect(
      screen.getByText(
        "No messages yet. Start a new chat or send the first prompt.",
      ),
    ).toBeInTheDocument();
    expect(promptBox).toHaveValue("hello");
  });

  it("keeps the prompt visible and shows the recovery failure when submit recovery fails before stream start", async () => {
    const user = userEvent.setup();
    mockGetConversation.mockResolvedValue(conversationDetail([]));
    mockStreamMessage.mockRejectedValue(
      new Error("AI chat stream ended before a terminal event"),
    );
    mockRequestRecovery.mockRejectedValue(
      new Error("ai chat recovery is not configured"),
    );

    render(<ChatRouteComponent />);

    await user.type(
      await screen.findByPlaceholderText(
        "Ask about training, recovery, exercise choices, or FitTrack usage...",
      ),
      "hello",
    );
    await user.click(screen.getByRole("button", { name: "Send" }));

    expect(await screen.findByText("hello")).toBeInTheDocument();
    expect(
      await screen.findByText("ai chat recovery is not configured"),
    ).toBeInTheDocument();
    expect(mockShowErrorToast).toHaveBeenCalledWith(
      expect.objectContaining({
        message: "ai chat recovery is not configured",
      }),
      "Failed to stream AI chat response",
    );
    expect(mockReportTelemetry).toHaveBeenCalledWith({
      category: "recovery",
      outcome: "recovered_failed",
    });
  });

  it("ignores late initial load results after the conversation is cleared", async () => {
    const initialLoad =
      deferredPromise<ReturnType<typeof conversationDetail>>();
    mockGetConversation.mockReturnValue(initialLoad.promise);

    const view = render(<ChatRouteComponent />);

    await waitFor(() => {
      expect(mockGetConversation).toHaveBeenCalledTimes(1);
    });

    mockSearch.conversationId = undefined;
    view.rerender(<ChatRouteComponent />);

    await waitFor(() => {
      expect(
        screen.getByText(
          "No messages yet. Start a new chat or send the first prompt.",
        ),
      ).toBeInTheDocument();
    });

    initialLoad.resolve(
      conversationDetail([
        {
          id: 61,
          conversation_id: 41,
          role: "assistant",
          content: "stale reply",
          status: "streaming",
          created_at: "2026-03-26T17:00:00Z",
          updated_at: "2026-03-26T17:00:01Z",
        },
      ]),
    );

    await waitFor(() => {
      expect(screen.queryByText("stale reply")).not.toBeInTheDocument();
    });
    expect(mockPollConversation).not.toHaveBeenCalled();
  });

  it("ignores late recovery results after the conversation is cleared", async () => {
    const recovery = deferredPromise<ReturnType<typeof conversationDetail>>();
    mockGetConversation.mockResolvedValue(
      conversationDetail([
        {
          id: 61,
          conversation_id: 41,
          role: "assistant",
          content: "partial",
          status: "streaming",
          created_at: "2026-03-26T17:00:00Z",
          updated_at: "2026-03-26T17:00:01Z",
        },
      ]),
    );
    mockPollConversation.mockReturnValue(recovery.promise);

    const view = render(<ChatRouteComponent />);

    await waitFor(() => {
      expect(mockPollConversation).toHaveBeenCalledTimes(1);
    });

    mockSearch.conversationId = undefined;
    view.rerender(<ChatRouteComponent />);

    await waitFor(() => {
      expect(
        screen.getByText(
          "No messages yet. Start a new chat or send the first prompt.",
        ),
      ).toBeInTheDocument();
    });

    recovery.resolve(
      conversationDetail([
        {
          id: 71,
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

    await waitFor(() => {
      expect(screen.queryByText("Recovered answer")).not.toBeInTheDocument();
    });
  });

  it("does not toast when submit recovery is aborted by clearing the conversation", async () => {
    const user = userEvent.setup();
    mockGetConversation.mockResolvedValue(conversationDetail([]));
    mockStreamMessage.mockRejectedValue(
      new Error("AI chat stream ended before a terminal event"),
    );
    mockPollConversation.mockImplementation(
      (_conversationId: number, options?: { signal?: AbortSignal }) =>
        new Promise((_resolve, reject) => {
          options?.signal?.addEventListener(
            "abort",
            () => reject(new DOMException("Aborted", "AbortError")),
            { once: true },
          );
        }),
    );

    const view = render(<ChatRouteComponent />);

    await user.type(
      await screen.findByPlaceholderText(
        "Ask about training, recovery, exercise choices, or FitTrack usage...",
      ),
      "hello",
    );
    await user.click(screen.getByRole("button", { name: "Send" }));

    await waitFor(() => {
      expect(mockPollConversation).toHaveBeenCalledTimes(1);
    });

    mockSearch.conversationId = undefined;
    view.rerender(<ChatRouteComponent />);

    await waitFor(() => {
      expect(
        screen.getByPlaceholderText(
          "Ask about training, recovery, exercise choices, or FitTrack usage...",
        ),
      ).toBeEnabled();
    });
    expect(
      screen.getByText(
        "No messages yet. Start a new chat or send the first prompt.",
      ),
    ).toBeInTheDocument();
    expect(mockShowErrorToast).not.toHaveBeenCalled();
    expect(mockReportTelemetry).toHaveBeenCalledWith({
      category: "recovery",
      outcome: "recovery_aborted",
    });
  });

  it("stops retrying load-triggered recovery after the handoff is queued", async () => {
    mockGetConversation.mockResolvedValue(
      conversationDetail([
        {
          id: 61,
          conversation_id: 41,
          role: "assistant",
          content: "partial",
          status: "streaming",
          created_at: "2026-03-26T17:00:00Z",
          updated_at: "2026-03-26T17:00:01Z",
        },
      ]),
    );
    mockRequestRecovery
      .mockResolvedValueOnce({
        conversation_id: 41,
        status: "not_needed",
      })
      .mockResolvedValueOnce({
        conversation_id: 41,
        run_id: 61,
        status: "queued",
      });
    mockPollConversation.mockImplementation(
      async (
        _conversationId: number,
        options?: {
          onStreaming?: (
            detail: ReturnType<typeof conversationDetail>,
          ) => Promise<void> | void;
        },
      ) => {
        await options?.onStreaming?.(
          conversationDetail([
            {
              id: 61,
              conversation_id: 41,
              role: "assistant",
              content: "partial",
              status: "streaming",
              created_at: "2026-03-26T17:00:00Z",
              updated_at: "2026-03-26T17:00:01Z",
            },
          ]),
        );
        await options?.onStreaming?.(
          conversationDetail([
            {
              id: 61,
              conversation_id: 41,
              role: "assistant",
              content: "partial",
              status: "streaming",
              created_at: "2026-03-26T17:00:00Z",
              updated_at: "2026-03-26T17:00:01Z",
            },
          ]),
        );

        return conversationDetail([
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
        ]);
      },
    );

    render(<ChatRouteComponent />);

    expect(await screen.findByText("Recovered answer")).toBeInTheDocument();
    expect(mockRequestRecovery).toHaveBeenCalledTimes(2);
    expect(mockRequestRecovery).toHaveBeenNthCalledWith(
      1,
      41,
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      }),
    );
    expect(mockRequestRecovery).toHaveBeenNthCalledWith(
      2,
      41,
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      }),
    );
    expect(mockShowErrorToast).not.toHaveBeenCalled();
    expect(mockReportTelemetry).toHaveBeenCalledWith({
      category: "recovery",
      outcome: "recovered_completed",
    });
  });

  it("resumes an active run from the last seen sequence without duplicating replayed text", async () => {
    window.sessionStorage.setItem(
      "fittrack.ai-chat.resume:41",
      JSON.stringify({
        runId: 91,
        sequence: 3,
        assistantMessageId: 61,
      }),
    );
    mockGetConversation
      .mockResolvedValueOnce(
        conversationDetail(
          [
            {
              id: 61,
              conversation_id: 41,
              role: "assistant",
              content: "hello",
              status: "streaming",
              created_at: "2026-03-26T17:00:00Z",
              updated_at: "2026-03-26T17:00:01Z",
            },
          ],
          {
            id: 91,
            assistant_message_id: 61,
            status: "streaming",
            latest_sequence: 3,
          },
        ),
      )
      .mockResolvedValueOnce(
        conversationDetail([
          {
            id: 61,
            conversation_id: 41,
            role: "assistant",
            content: "hello world",
            status: "completed",
            created_at: "2026-03-26T17:00:00Z",
            updated_at: "2026-03-26T17:00:02Z",
            completed_at: "2026-03-26T17:00:02Z",
          },
        ]),
      );
    mockResumeStream.mockImplementation(
      async (
        conversationId: number,
        runId: number,
        afterSequence: number,
        options?: {
          onDelta?: (event: Record<string, unknown>) => void;
          onDone?: (event: Record<string, unknown>) => void;
        },
      ) => {
        expect(conversationId).toBe(41);
        expect(runId).toBe(91);
        expect(afterSequence).toBe(3);
        options?.onDelta?.({
          type: "delta",
          delta: " world",
          sequence: 4,
        });
        options?.onDone?.({
          type: "done",
          message_id: 61,
          text: "hello world",
          sequence: 4,
        });

        return {
          doneEvent: {
            type: "done",
            message_id: 61,
            text: "hello world",
          },
          endedWithError: false,
        };
      },
    );

    render(<ChatRouteComponent />);

    expect(await screen.findByText("hello world")).toBeInTheDocument();
    expect(screen.queryByText("hello world world")).not.toBeInTheDocument();
    expect(mockResumeStream).toHaveBeenCalledWith(
      41,
      91,
      3,
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      }),
    );
    expect(mockRequestRecovery).not.toHaveBeenCalled();
    expect(mockPollConversation).not.toHaveBeenCalled();
  });

  it("falls back to recovery polling when resume is unavailable", async () => {
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

    render(<ChatRouteComponent />);

    expect(await screen.findByText("Recovered answer")).toBeInTheDocument();
    expect(mockResumeStream).toHaveBeenCalledTimes(1);
    expect(mockRequestRecovery).toHaveBeenCalledWith(
      41,
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      }),
    );
    expect(mockPollConversation).toHaveBeenCalledWith(
      41,
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      }),
    );
  });

  it("falls back to recovery polling when the resume stream exits before a terminal event", async () => {
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
    mockResumeStream.mockRejectedValue(
      new Error("AI chat stream ended before a terminal event"),
    );
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

    render(<ChatRouteComponent />);

    expect(await screen.findByText("Recovered answer")).toBeInTheDocument();
    expect(mockResumeStream).toHaveBeenCalledTimes(1);
    expect(mockRequestRecovery).toHaveBeenCalledWith(
      41,
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      }),
    );
    expect(mockPollConversation).toHaveBeenCalledWith(
      41,
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      }),
    );
    expect(mockReportTelemetry).toHaveBeenCalledWith({
      category: "recovery",
      outcome: "recovered_completed",
    });
    expect(mockShowErrorToast).not.toHaveBeenCalled();
  });

  it("finishes reconnect without recovery when the resume stream returns the completed reply", async () => {
    mockGetConversation
      .mockResolvedValueOnce(
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
      )
      .mockResolvedValueOnce(
        conversationDetail([
          {
            id: 61,
            conversation_id: 41,
            role: "assistant",
            content: "Completed answer",
            status: "completed",
            created_at: "2026-03-26T17:00:00Z",
            updated_at: "2026-03-26T17:00:02Z",
            completed_at: "2026-03-26T17:00:02Z",
          },
        ]),
      );
    mockResumeStream.mockImplementation(
      async (
        _conversationId: number,
        _runId: number,
        _afterSequence: number,
        options?: {
          onDone?: (event: Record<string, unknown>) => void;
        },
      ) => {
        options?.onDone?.({
          type: "done",
          message_id: 61,
          text: "Completed answer",
          sequence: 2,
        });

        return {
          doneEvent: {
            type: "done",
            message_id: 61,
            text: "Completed answer",
          },
          endedWithError: false,
        };
      },
    );

    render(<ChatRouteComponent />);

    expect(await screen.findByText("Completed answer")).toBeInTheDocument();
    expect(mockResumeStream).toHaveBeenCalledTimes(1);
    expect(mockRequestRecovery).not.toHaveBeenCalled();
    expect(mockPollConversation).not.toHaveBeenCalled();
  });

  it("shows a user-visible failure when load-triggered recovery times out", async () => {
    mockGetConversation.mockResolvedValue(
      conversationDetail([
        {
          id: 61,
          conversation_id: 41,
          role: "assistant",
          content: "partial",
          status: "streaming",
          created_at: "2026-03-26T17:00:00Z",
          updated_at: "2026-03-26T17:00:01Z",
        },
      ]),
    );
    mockPollConversation.mockRejectedValue(
      new Error(
        "AI chat recovery timed out while waiting for persisted conversation state",
      ),
    );

    render(<ChatRouteComponent />);

    expect(
      await screen.findByText(
        "AI chat recovery timed out while waiting for persisted conversation state",
      ),
    ).toBeInTheDocument();
    expect(mockRequestRecovery).toHaveBeenCalledWith(
      41,
      expect.objectContaining({
        signal: expect.any(AbortSignal),
      }),
    );
    expect(mockReportTelemetry).toHaveBeenCalledWith({
      category: "recovery",
      outcome: "recovery_timeout",
    });
    expect(mockReportTelemetry).toHaveBeenCalledWith({
      category: "ux",
      outcome: "failure_toast_shown",
    });
    expect(mockShowErrorToast).toHaveBeenCalledWith(
      expect.objectContaining({
        message:
          "AI chat recovery timed out while waiting for persisted conversation state",
      }),
      "Failed to recover AI chat conversation",
    );
  });

  it("does not reclassify a completed stream when the follow-up refresh fails", async () => {
    const user = userEvent.setup();
    mockGetConversation
      .mockResolvedValueOnce(conversationDetail([]))
      .mockRejectedValueOnce(new Error("refresh failed"));
    mockStreamMessage.mockImplementation(
      async (
        _conversationId: number,
        _prompt: string,
        options?: {
          onStart?: (event: Record<string, unknown>) => void;
          onDone?: (event: Record<string, unknown>) => void;
        },
      ) => {
        options?.onStart?.({
          type: "start",
          message_id: 72,
        });
        options?.onDone?.({
          type: "done",
          message_id: 72,
          text: "Completed answer",
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

    render(<ChatRouteComponent />);

    await user.type(
      await screen.findByPlaceholderText(
        "Ask about training, recovery, exercise choices, or FitTrack usage...",
      ),
      "hello",
    );
    await user.click(screen.getByRole("button", { name: "Send" }));

    expect(await screen.findByText("refresh failed")).toBeInTheDocument();
    expect(mockReportTelemetry).toHaveBeenCalledWith({
      category: "stream",
      outcome: "completed",
      stage: "terminal",
    });
    expect(mockReportTelemetry).not.toHaveBeenCalledWith(
      expect.objectContaining({
        category: "stream",
        outcome: "transport_ended_pre_terminal",
      }),
    );
    expect(mockPollConversation).not.toHaveBeenCalled();
    expect(mockShowErrorToast).not.toHaveBeenCalled();
  });

  it("keeps the active stream running when new chat creation fails", async () => {
    const user = userEvent.setup();
    mockGetConversation.mockResolvedValue(conversationDetail([]));

    let streamSignal: AbortSignal | undefined;
    mockStreamMessage.mockImplementation(
      (
        _conversationId: number,
        _prompt: string,
        options?: { signal?: AbortSignal },
      ) => {
        streamSignal = options?.signal;
        return new Promise(() => {});
      },
    );
    mockCreateConversation.mockRejectedValue(new Error("create failed"));

    const view = render(<ChatRouteComponent />);

    await user.type(
      await screen.findByPlaceholderText(
        "Ask about training, recovery, exercise choices, or FitTrack usage...",
      ),
      "hello",
    );
    await user.click(screen.getByRole("button", { name: "Send" }));

    await waitFor(() => {
      expect(mockStreamMessage).toHaveBeenCalledTimes(1);
    });

    await user.click(screen.getByRole("button", { name: "New Chat" }));

    await waitFor(() => {
      expect(mockShowErrorToast).toHaveBeenCalledWith(
        expect.objectContaining({ message: "create failed" }),
        "Failed to create chat conversation",
      );
    });
    expect(streamSignal?.aborted).toBe(false);
    expect(screen.getByText("hello")).toBeInTheDocument();
    expect(screen.getByText("...")).toBeInTheDocument();

    view.unmount();
  });

  it("shows the latest workout draft on reopen and imports it into the workout form flow", async () => {
    const user = userEvent.setup();
    const confirmSpy = vi.spyOn(window, "confirm").mockReturnValue(true);
    const latestWorkoutDraft: AIWorkoutDraft = {
      date: "2026-04-21T12:00:00Z",
      notes: "Keep rest short",
      workoutFocus: "pull",
      exercises: [
        {
          name: "Chest Supported Row",
          sets: [{ reps: 10, setType: "working" }],
        },
      ],
    };

    mockGetConversation.mockResolvedValue(
      conversationDetail([], undefined, latestWorkoutDraft),
    );

    render(<ChatRouteComponent />);

    expect(
      await screen.findByText("Latest structured workout draft"),
    ).toBeInTheDocument();
    expect(screen.getByText(/Chest Supported Row/)).toBeInTheDocument();
    expect(screen.getByText("Keep rest short")).toBeInTheDocument();

    await user.click(
      screen.getByRole("button", { name: "Edit in workout form" }),
    );

    expect(confirmSpy).not.toHaveBeenCalled();
    await waitFor(() => {
      expect(window.localStorage.getItem("workout-entry-form-data-user-123")).toBe(
        JSON.stringify({
          date: "2026-04-21T12:00:00Z",
          notes: "Keep rest short",
          workoutFocus: "pull",
          exercises: [
            {
              name: "Chest Supported Row",
              sets: [{ reps: 10, setType: "working" }],
            },
          ],
        }),
      );
    });
    expect(mockNavigate).toHaveBeenCalledWith({ to: "/workouts/new" });
  });

  it("overwrites the latest workout draft CTA after a regenerated structured workout", async () => {
    const user = userEvent.setup();
    const originalDraft: AIWorkoutDraft = {
      date: "2026-04-20T12:00:00Z",
      notes: "Original draft",
      workoutFocus: "push",
      exercises: [
        {
          name: "Bench Press",
          sets: [{ reps: 8, setType: "working", weight: 185 }],
        },
      ],
    };
    const regeneratedDraft: AIWorkoutDraft = {
      date: "2026-04-21T12:00:00Z",
      notes: "Regenerated draft",
      workoutFocus: "pull",
      exercises: [
        {
          name: "Chest Supported Row",
          sets: [{ reps: 10, setType: "working" }],
        },
      ],
    };

    mockGetConversation
      .mockResolvedValueOnce(conversationDetail([], undefined, originalDraft))
      .mockResolvedValueOnce(
        conversationDetail(
          [
            {
              id: 71,
              conversation_id: 41,
              role: "user",
              content: "regenerate it",
              status: "completed",
              created_at: "2026-03-26T17:00:01Z",
              updated_at: "2026-03-26T17:00:01Z",
              completed_at: "2026-03-26T17:00:01Z",
            },
            {
              id: 72,
              conversation_id: 41,
              role: "assistant",
              content: "I put together a structured workout draft for you.",
              status: "completed",
              created_at: "2026-03-26T17:00:01Z",
              updated_at: "2026-03-26T17:00:02Z",
              completed_at: "2026-03-26T17:00:02Z",
            },
          ],
          undefined,
          regeneratedDraft,
        ),
      );
    mockStreamMessage.mockImplementation(
      async (
        _conversationId: number,
        _prompt: string,
        options?: {
          onStart?: (event: Record<string, unknown>) => void;
          onDone?: (event: Record<string, unknown>) => void;
        },
      ) => {
        options?.onStart?.({
          type: "start",
          message_id: 72,
        });
        options?.onDone?.({
          type: "done",
          message_id: 72,
          text: "I put together a structured workout draft for you.",
          workout_draft: regeneratedDraft,
        });

        return {
          doneEvent: {
            type: "done",
            message_id: 72,
            text: "I put together a structured workout draft for you.",
            workout_draft: regeneratedDraft,
          },
          endedWithError: false,
        };
      },
    );

    render(<ChatRouteComponent />);

    expect(await screen.findByText(/Bench Press/)).toBeInTheDocument();

    await user.type(
      await screen.findByPlaceholderText(
        "Ask about training, recovery, exercise choices, or FitTrack usage...",
      ),
      "regenerate it",
    );
    await user.click(screen.getByRole("button", { name: "Send" }));

    await user.click(
      await screen.findByRole("button", { name: "Edit in workout form" }),
    );

    expect(screen.getByText(/Chest Supported Row/)).toBeInTheDocument();
    expect(screen.queryByText(/Bench Press/)).not.toBeInTheDocument();
    expect(
      screen.queryAllByRole("button", { name: "Edit in workout form" }),
    ).toHaveLength(1);

    await waitFor(() => {
      expect(window.localStorage.getItem("workout-entry-form-data-user-123")).toBe(
        JSON.stringify({
          date: "2026-04-21T12:00:00Z",
          notes: "Regenerated draft",
          workoutFocus: "pull",
          exercises: [
            {
              name: "Chest Supported Row",
              sets: [{ reps: 10, setType: "working" }],
            },
          ],
        }),
      );
    });
  });

  it("keeps the workout draft card with the draft-producing reply after a non-draft follow-up", async () => {
    const user = userEvent.setup();
    const generatedDraft: AIWorkoutDraft = {
      date: "2026-04-21T12:00:00Z",
      notes: "Generated draft",
      workoutFocus: "pull",
      exercises: [
        {
          name: "Chest Supported Row",
          sets: [{ reps: 10, setType: "working" }],
        },
      ],
    };
    const generatedMessages = [
      {
        id: 71,
        conversation_id: 41,
        role: "user",
        content: "build a pull workout",
        status: "completed",
        created_at: "2026-03-26T17:00:01Z",
        updated_at: "2026-03-26T17:00:01Z",
        completed_at: "2026-03-26T17:00:01Z",
      },
      {
        id: 72,
        conversation_id: 41,
        role: "assistant",
        content: "I put together a structured workout draft for you.",
        status: "completed",
        created_at: "2026-03-26T17:00:01Z",
        updated_at: "2026-03-26T17:00:02Z",
        completed_at: "2026-03-26T17:00:02Z",
      },
    ];
    const followUpMessages = [
      ...generatedMessages,
      {
        id: 73,
        conversation_id: 41,
        role: "user",
        content: "how long should I rest?",
        status: "completed",
        created_at: "2026-03-26T17:01:01Z",
        updated_at: "2026-03-26T17:01:01Z",
        completed_at: "2026-03-26T17:01:01Z",
      },
      {
        id: 74,
        conversation_id: 41,
        role: "assistant",
        content: "Rest 90 seconds between these working sets.",
        status: "completed",
        created_at: "2026-03-26T17:01:01Z",
        updated_at: "2026-03-26T17:01:02Z",
        completed_at: "2026-03-26T17:01:02Z",
      },
    ];

    mockGetConversation
      .mockResolvedValueOnce(conversationDetail([]))
      .mockResolvedValueOnce(
        conversationDetail(generatedMessages, undefined, generatedDraft),
      )
      .mockResolvedValueOnce(
        conversationDetail(followUpMessages, undefined, generatedDraft),
      );
    mockStreamMessage
      .mockImplementationOnce(
        async (
          _conversationId: number,
          _prompt: string,
          options?: {
            onStart?: (event: Record<string, unknown>) => void;
            onDone?: (event: Record<string, unknown>) => void;
          },
        ) => {
          options?.onStart?.({ type: "start", message_id: 72 });
          options?.onDone?.({
            type: "done",
            message_id: 72,
            text: "I put together a structured workout draft for you.",
            workout_draft: generatedDraft,
          });

          return {
            doneEvent: {
              type: "done",
              message_id: 72,
              text: "I put together a structured workout draft for you.",
              workout_draft: generatedDraft,
            },
            endedWithError: false,
          };
        },
      )
      .mockImplementationOnce(
        async (
          _conversationId: number,
          _prompt: string,
          options?: {
            onStart?: (event: Record<string, unknown>) => void;
            onDone?: (event: Record<string, unknown>) => void;
          },
        ) => {
          options?.onStart?.({ type: "start", message_id: 74 });
          options?.onDone?.({
            type: "done",
            message_id: 74,
            text: "Rest 90 seconds between these working sets.",
          });

          return {
            doneEvent: {
              type: "done",
              message_id: 74,
              text: "Rest 90 seconds between these working sets.",
            },
            endedWithError: false,
          };
        },
      );

    render(<ChatRouteComponent />);

    await user.type(
      await screen.findByPlaceholderText(
        "Ask about training, recovery, exercise choices, or FitTrack usage...",
      ),
      "build a pull workout",
    );
    await user.click(screen.getByRole("button", { name: "Send" }));

    await waitFor(() => {
      expect(screen.getByTestId("chat-message-72")).toBeInTheDocument();
    });
    expect(
      within(screen.getByTestId("chat-message-72")).getByText(
        "Latest structured workout draft",
      ),
    ).toBeInTheDocument();

    await user.type(
      screen.getByPlaceholderText(
        "Ask about training, recovery, exercise choices, or FitTrack usage...",
      ),
      "how long should I rest?",
    );
    await user.click(screen.getByRole("button", { name: "Send" }));

    expect(
      await screen.findByText("Rest 90 seconds between these working sets."),
    ).toBeInTheDocument();
    expect(
      within(screen.getByTestId("chat-message-72")).getByText(
        "Latest structured workout draft",
      ),
    ).toBeInTheDocument();
    expect(
      within(screen.getByTestId("chat-message-74")).queryByText(
        "Latest structured workout draft",
      ),
    ).not.toBeInTheDocument();
  });
});
