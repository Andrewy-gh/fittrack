import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { AIChatConversationSummary } from "@/features/chat/api/ai-chat";

const { mockListConversations } = vi.hoisted(() => ({
  mockListConversations: vi.fn(),
}));

vi.mock("@/features/chat/api/ai-chat", () => ({
  listAIChatConversations: mockListConversations,
}));

vi.mock("@/lib/errors", () => ({
  getErrorMessage: (
    error: unknown,
    fallback = "An unexpected error occurred",
  ) => (error instanceof Error ? error.message : fallback),
}));

import { useChatHistoryEntry } from "@/features/chat/hooks/use-chat-history-entry";

function conversation(id: number, title: string): AIChatConversationSummary {
  return {
    id,
    title,
    created_at: "2026-06-25T17:00:00Z",
    updated_at: "2026-06-25T17:05:00Z",
    last_message_at: "2026-06-25T17:05:00Z",
  };
}

describe("useChatHistoryEntry", () => {
  beforeEach(() => {
    mockListConversations.mockReset();
  });

  it("ignores a stale manual refresh after the signed-in user changes", async () => {
    let resolveStaleRefresh!: (value: AIChatConversationSummary[]) => void;
    mockListConversations
      .mockResolvedValueOnce([conversation(72, "User 1 initial chat")])
      .mockReturnValueOnce(
        new Promise((resolve) => {
          resolveStaleRefresh = resolve;
        }),
      )
      .mockResolvedValueOnce([conversation(88, "User 2 current chat")]);

    const onOpenConversation = vi.fn();
    const view = renderHook(
      ({ userId }) =>
        useChatHistoryEntry({
          userId,
          conversationId: null,
          onOpenConversation,
        }),
      { initialProps: { userId: "user-1" } },
    );

    await waitFor(() => {
      expect(view.result.current.conversations).toEqual([
        expect.objectContaining({ title: "User 1 initial chat" }),
      ]);
    });

    void act(() => {
      void view.result.current.refreshConversations();
    });
    await waitFor(() => {
      expect(mockListConversations).toHaveBeenCalledTimes(2);
    });

    view.rerender({ userId: "user-2" });

    await waitFor(() => {
      expect(view.result.current.conversations).toEqual([
        expect.objectContaining({ title: "User 2 current chat" }),
      ]);
    });
    expect(view.result.current.entryState).toEqual({
      status: "openingLatestChat",
    });

    await act(async () => {
      resolveStaleRefresh([conversation(73, "User 1 stale chat")]);
    });

    expect(view.result.current.conversations).toEqual([
      expect.objectContaining({ title: "User 2 current chat" }),
    ]);
    expect(view.result.current.entryState).toEqual({
      status: "openingLatestChat",
    });
  });

  it("models the no-conversation loading path as opening the latest chat", () => {
    mockListConversations.mockReturnValue(new Promise(() => {}));

    const view = renderHook(() =>
      useChatHistoryEntry({
        userId: "user-1",
        conversationId: null,
        onOpenConversation: vi.fn(),
      }),
    );

    expect(view.result.current.entryState).toEqual({
      status: "openingLatestChat",
    });
  });

  it("models no-conversation history load failures separately from ready chat content", async () => {
    mockListConversations.mockRejectedValue(new Error("Recent chats failed"));

    const view = renderHook(() =>
      useChatHistoryEntry({
        userId: "user-1",
        conversationId: null,
        onOpenConversation: vi.fn(),
      }),
    );

    await waitFor(() => {
      expect(view.result.current.entryState).toEqual({
        status: "historyLoadError",
        message: "Recent chats failed",
      });
    });

    view.rerender();

    expect(view.result.current.entryState).toEqual({
      status: "historyLoadError",
      message: "Recent chats failed",
    });
  });
});
