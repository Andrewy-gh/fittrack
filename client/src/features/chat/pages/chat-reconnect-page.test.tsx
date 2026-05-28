import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";
import {
  ChatRouteComponent,
  conversationDetail,
  mockGetConversation,
  mockPollConversation,
  mockReportTelemetry,
  mockRequestRecovery,
  mockResumeStream,
  mockShowErrorToast,
  resetChatRouteMocks,
} from "../test/chat-page-test-utils";

describe("ChatRouteComponent", () => {
  beforeEach(resetChatRouteMocks);

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
});
