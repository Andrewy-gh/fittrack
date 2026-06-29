import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { AIChatConversationSummary } from "@/features/chat/api/ai-chat";
import {
  ChatHistoryEntry,
  getChatHistoryListState,
} from "./chat-history-entry";

function conversation(id: number, title: string): AIChatConversationSummary {
  return {
    id,
    title,
    created_at: "2026-06-25T17:00:00Z",
    updated_at: "2026-06-25T17:05:00Z",
    last_message_at: "2026-06-25T17:05:00Z",
  };
}

function renderChatHistoryEntry(
  stateInput: Parameters<typeof getChatHistoryListState>[0],
) {
  render(
    <ChatHistoryEntry
      historyState={getChatHistoryListState(stateInput)}
      isCollapsed={false}
      isMobileOpen={false}
      onMobileOpenChange={vi.fn()}
      onToggleCollapsed={vi.fn()}
      onResumeConversation={vi.fn()}
      onNewChat={vi.fn()}
    />,
  );
}

describe("ChatHistoryEntry", () => {
  it("keeps showing the loading state while a refresh has stale data and an error", () => {
    renderChatHistoryEntry({
      conversations: [conversation(41, "Leg day plan")],
      activeConversationId: 41,
      isLoading: true,
      error: "Could not load recent chats.",
    });

    expect(screen.getByText("Loading recent chats...")).toBeInTheDocument();
    expect(
      screen.queryByText("Could not load recent chats."),
    ).not.toBeInTheDocument();
    expect(screen.queryByText("Leg day plan")).not.toBeInTheDocument();
  });

  it("marks the active ready conversation as current", () => {
    renderChatHistoryEntry({
      conversations: [conversation(41, "Leg day plan")],
      activeConversationId: 41,
      isLoading: false,
      error: null,
    });

    expect(
      screen.getByRole("button", { name: /Leg day plan/ }),
    ).toHaveAttribute("aria-current", "page");
    expect(screen.getByText("Current conversation")).toBeInTheDocument();
  });
});
