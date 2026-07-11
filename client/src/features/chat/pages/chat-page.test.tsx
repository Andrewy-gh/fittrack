import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { StrictMode } from "react";
import { beforeEach, describe, expect, it } from "vitest";
import {
  ChatRouteComponent,
  conversationDetail,
  deferredPromise,
  mockCreateConversation,
  mockGetConversation,
  mockListConversations,
  mockNavigate,
  mockPollConversation,
  mockReportTelemetry,
  mockRequestRecovery,
  mockSearch,
  mockShowErrorToast,
  mockStreamMessage,
  resetChatRouteMocks,
} from "../test/chat-page-test-utils";

describe("ChatRouteComponent", () => {
  beforeEach(resetChatRouteMocks);

  it("opens the newest recent chat when entering without a conversation id", async () => {
    mockSearch.conversationId = undefined;
    mockListConversations.mockResolvedValue([
      {
        id: 72,
        title: "Leg day plan",
        created_at: "2026-06-25T17:00:00Z",
        updated_at: "2026-06-25T17:05:00Z",
        last_message_at: "2026-06-25T17:05:00Z",
      },
    ]);

    render(<ChatRouteComponent />);

    expect(await screen.findByText("Leg day plan")).toBeInTheDocument();
    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith({
        to: "/chat",
        search: { conversationId: "72" },
        replace: true,
      });
    });
    expect(
      screen.queryByText("What should we train today?"),
    ).not.toBeInTheDocument();
    expect(mockGetConversation).not.toHaveBeenCalled();
  });

  it("opens a blank draft without persisting a conversation from chat navigation", async () => {
    mockSearch.conversationId = undefined;
    mockSearch.createChat = true;
    mockListConversations.mockResolvedValue([
      {
        id: 72,
        title: "Leg day plan",
        created_at: "2026-06-25T17:00:00Z",
        updated_at: "2026-06-25T17:05:00Z",
        last_message_at: "2026-06-25T17:05:00Z",
      },
    ]);

    render(
      <StrictMode>
        <ChatRouteComponent />
      </StrictMode>,
    );

    expect(
      await screen.findByText("What should we train today?"),
    ).toBeInTheDocument();
    expect(mockCreateConversation).not.toHaveBeenCalled();
    expect(mockNavigate).not.toHaveBeenCalledWith(
      expect.objectContaining({
        search: { conversationId: "72" },
      }),
    );
    expect(
      screen.getByPlaceholderText(
        "Ask about training, recovery, exercise choices, or FitTrack usage...",
      ),
    ).toBeEnabled();
  });

  it("does not auto-open stale history after the signed-in user changes", async () => {
    mockSearch.conversationId = undefined;
    let resolveNextHistory!: (value: Array<Record<string, unknown>>) => void;
    mockListConversations
      .mockResolvedValueOnce([
        {
          id: 72,
          title: "Leg day plan",
          created_at: "2026-06-25T17:00:00Z",
          updated_at: "2026-06-25T17:05:00Z",
          last_message_at: "2026-06-25T17:05:00Z",
        },
      ])
      .mockReturnValueOnce(
        new Promise((resolve) => {
          resolveNextHistory = resolve;
        }),
      );

    const view = render(<ChatRouteComponent userId="user-123" />);

    expect(await screen.findByText("Leg day plan")).toBeInTheDocument();
    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith({
        to: "/chat",
        search: { conversationId: "72" },
        replace: true,
      });
    });

    mockNavigate.mockClear();
    view.rerender(<ChatRouteComponent userId="user-456" />);

    expect(screen.queryByText("Leg day plan")).not.toBeInTheDocument();
    expect(mockNavigate).not.toHaveBeenCalled();

    resolveNextHistory([
      {
        id: 88,
        title: "Back day plan",
        created_at: "2026-06-26T17:00:00Z",
        updated_at: "2026-06-26T17:05:00Z",
        last_message_at: "2026-06-26T17:05:00Z",
      },
    ]);

    expect(await screen.findByText("Back day plan")).toBeInTheDocument();
    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith({
        to: "/chat",
        search: { conversationId: "88" },
        replace: true,
      });
    });
  });

  it("opens a blank draft from chat history without persisting it", async () => {
    const user = userEvent.setup();
    mockGetConversation.mockResolvedValue(conversationDetail([]));
    mockListConversations.mockResolvedValue([
      {
        id: 41,
        title: "Leg day plan",
        created_at: "2026-06-25T17:00:00Z",
        updated_at: "2026-06-25T17:05:00Z",
        last_message_at: "2026-06-25T17:05:00Z",
      },
    ]);

    render(<ChatRouteComponent />);

    expect(await screen.findByText("Leg day plan")).toBeInTheDocument();
    const newChatButtons = screen.getAllByRole("button", { name: "New Chat" });
    await user.click(newChatButtons.at(-1)!);

    expect(mockCreateConversation).not.toHaveBeenCalled();
    expect(mockNavigate).toHaveBeenCalledWith({
      to: "/chat",
      search: { createChat: true },
    });
    expect(mockListConversations).toHaveBeenCalledTimes(1);
  });

  it("refreshes chat history after the first prompt creates a conversation", async () => {
    const user = userEvent.setup();
    mockSearch.conversationId = undefined;
    mockSearch.createChat = true;
    mockListConversations
      .mockResolvedValueOnce([])
      .mockResolvedValueOnce([
        {
          id: 73,
          created_at: "2026-06-26T17:00:00Z",
          updated_at: "2026-06-26T17:00:00Z",
        },
      ])
      .mockResolvedValueOnce([
        {
          id: 73,
          title: "What should we train",
          created_at: "2026-06-26T17:00:00Z",
          updated_at: "2026-06-26T17:05:00Z",
          last_message_at: "2026-06-26T17:05:00Z",
        },
      ]);
    mockCreateConversation.mockResolvedValue({
      id: 73,
      created_at: "2026-06-26T17:00:00Z",
      updated_at: "2026-06-26T17:00:00Z",
    });
    mockGetConversation.mockResolvedValue({
      conversation: {
        id: 73,
        title: "What should we train",
        created_at: "2026-06-26T17:00:00Z",
        updated_at: "2026-06-26T17:05:00Z",
        last_message_at: "2026-06-26T17:05:00Z",
      },
      messages: [],
    });
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
          text: "Start with squats.",
        });

        return {
          doneEvent: {
            type: "done",
            message_id: 72,
            text: "Start with squats.",
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
      "What should we train?",
    );
    await user.click(screen.getByRole("button", { name: "Send" }));

    expect(mockCreateConversation).toHaveBeenCalledTimes(1);
    expect(mockNavigate).toHaveBeenCalledWith({
      to: "/chat",
      search: { conversationId: "73" },
      replace: true,
    });
    expect(await screen.findByText("What should we train")).toBeInTheDocument();
    await waitFor(() => {
      expect(mockListConversations).toHaveBeenCalledTimes(3);
    });
  });

  it("drafts an example prompt without submitting it", async () => {
    const user = userEvent.setup();
    mockSearch.conversationId = undefined;
    mockListConversations.mockResolvedValue([]);

    render(<ChatRouteComponent />);

    await user.click(
      await screen.findByRole("button", {
        name: "Build me a 45-min push day",
      }),
    );

    expect(
      screen.getByPlaceholderText(
        "Ask about training, recovery, exercise choices, or FitTrack usage...",
      ),
    ).toHaveValue("Build me a 45-min push day");
    expect(mockCreateConversation).not.toHaveBeenCalled();
    expect(mockStreamMessage).not.toHaveBeenCalled();
    expect(mockListConversations).toHaveBeenCalledTimes(1);
  });

  it("shows every recent chat returned by the history endpoint", async () => {
    mockGetConversation.mockResolvedValue(conversationDetail([]));
    mockListConversations.mockResolvedValue(
      Array.from({ length: 50 }, (_, index) => ({
        id: index + 1,
        title: `Recent chat ${index + 1}`,
        created_at: "2026-06-25T17:00:00Z",
        updated_at: "2026-06-25T17:05:00Z",
        last_message_at: "2026-06-25T17:05:00Z",
      })),
    );

    render(<ChatRouteComponent />);

    expect(await screen.findByText("Recent chat 50")).toBeInTheDocument();
  });

  it("lets desktop users collapse and expand chat history", async () => {
    const user = userEvent.setup();
    mockGetConversation.mockResolvedValue(conversationDetail([]));
    mockListConversations.mockResolvedValue([
      {
        id: 41,
        title: "Leg day plan",
        created_at: "2026-06-25T17:00:00Z",
        updated_at: "2026-06-25T17:05:00Z",
        last_message_at: "2026-06-25T17:05:00Z",
      },
    ]);

    render(<ChatRouteComponent />);

    const expandedHistory = await screen.findByLabelText("Chat history");
    expect(expandedHistory).toHaveClass("lg:fixed", "lg:left-0");
    expect(screen.getByTestId("chat-page-layout")).toHaveClass(
      "lg:px-chat-gutter",
    );
    expect(screen.getByTestId("chat-main-pane")).toHaveClass(
      "mx-auto",
      "max-w-3xl",
    );

    await user.click(
      screen.getByRole("button", { name: "Collapse chat history" }),
    );

    expect(
      screen.getByRole("button", { name: "Expand chat history" }),
    ).toBeInTheDocument();
    const collapsedHistory = screen.getByLabelText("Collapsed chat history");
    expect(collapsedHistory).toHaveClass("lg:fixed", "lg:left-0");
    expect(screen.getByTestId("chat-page-layout")).toHaveClass(
      "lg:px-chat-gutter",
    );
    expect(screen.getByTestId("chat-main-pane")).toHaveClass(
      "mx-auto",
      "max-w-3xl",
    );
    expect(
      within(collapsedHistory)
        .getAllByRole("button")
        .map((button) => button.getAttribute("aria-label")),
    ).toEqual(["Expand chat history", "New Chat"]);
    expect(
      screen.queryByLabelText("Collapse chat history"),
    ).not.toBeInTheDocument();
  });

  it("omits ready access chrome from the chat page", async () => {
    mockGetConversation.mockResolvedValue(conversationDetail([]));
    mockListConversations.mockResolvedValue([
      {
        id: 41,
        title: "Leg day plan",
        created_at: "2026-06-25T17:00:00Z",
        updated_at: "2026-06-25T17:05:00Z",
        last_message_at: "2026-06-25T17:05:00Z",
      },
    ]);

    render(<ChatRouteComponent />);

    expect(await screen.findByText("Leg day plan")).toBeInTheDocument();
    expect(screen.queryByText("Access active")).not.toBeInTheDocument();
    expect(screen.getAllByRole("button", { name: "New Chat" })).toHaveLength(1);
  });

  it("adds top padding when showing an existing conversation", async () => {
    mockGetConversation.mockResolvedValue(
      conversationDetail([
        {
          id: 61,
          conversation_id: 41,
          role: "user",
          content: "Hello anybody there?",
          status: "completed",
          created_at: "2026-03-26T17:00:00Z",
          updated_at: "2026-03-26T17:00:00Z",
          completed_at: "2026-03-26T17:00:00Z",
        },
      ]),
    );
    mockListConversations.mockResolvedValue([]);

    render(<ChatRouteComponent />);

    expect(await screen.findByText("Hello anybody there?")).toBeInTheDocument();
    expect(screen.getByTestId("chat-conversation-body")).toHaveClass("pt-4");
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
    expect(screen.getByText("What should we train today?")).toBeInTheDocument();
    expect(
      screen.getByPlaceholderText(
        "Ask about training, recovery, exercise choices, or FitTrack usage...",
      ),
    ).toHaveValue("hello");
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
        screen.getByText("What should we train today?"),
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
        screen.getByText("What should we train today?"),
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
    expect(screen.getByText("What should we train today?")).toBeInTheDocument();
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
});
