import {
  History,
  MessageSquare,
  PanelLeftClose,
  PanelLeftOpen,
  Plus,
  SquarePen,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Drawer,
  DrawerContent,
  DrawerDescription,
  DrawerHeader,
  DrawerTitle,
} from "@/components/ui/drawer";
import type { AIChatConversationSummary } from "@/features/chat/api/ai-chat";
import { cn } from "@/lib/utils";

type NonEmptyConversations = [
  AIChatConversationSummary,
  ...AIChatConversationSummary[],
];

export type ChatHistoryListState =
  | { status: "loading" }
  | { status: "error"; message: string }
  | { status: "empty" }
  | {
      status: "ready";
      conversations: NonEmptyConversations;
      activeConversationId: number | null;
    };

type ChatHistoryListStateInput = {
  conversations: AIChatConversationSummary[];
  activeConversationId: number | null;
  isLoading: boolean;
  error: string | null;
};

type ChatHistoryEntryProps = {
  historyState: ChatHistoryListState;
  isCollapsed: boolean;
  isMobileOpen: boolean;
  onMobileOpenChange: (open: boolean) => void;
  onToggleCollapsed: () => void;
  onResumeConversation: (conversationId: number) => void;
  onNewChat: () => void;
  isNewChatDisabled?: boolean;
};

export function getChatHistoryListState({
  conversations,
  activeConversationId,
  isLoading,
  error,
}: ChatHistoryListStateInput): ChatHistoryListState {
  if (isLoading) {
    return { status: "loading" };
  }

  if (error) {
    return { status: "error", message: error };
  }

  const [firstConversation, ...remainingConversations] = conversations;
  if (!firstConversation) {
    return { status: "empty" };
  }

  return {
    status: "ready",
    conversations: [firstConversation, ...remainingConversations],
    activeConversationId,
  };
}

export function ChatHistoryEntry({
  historyState,
  isCollapsed,
  isMobileOpen,
  onMobileOpenChange,
  onToggleCollapsed,
  onResumeConversation,
  onNewChat,
  isNewChatDisabled,
}: ChatHistoryEntryProps) {
  function handleResumeConversation(conversationId: number) {
    onResumeConversation(conversationId);
    onMobileOpenChange(false);
  }

  function handleNewChat() {
    onNewChat();
    onMobileOpenChange(false);
  }

  return (
    <>
      <DesktopHistoryPanel
        historyState={historyState}
        isCollapsed={isCollapsed}
        onToggleCollapsed={onToggleCollapsed}
        onResumeConversation={onResumeConversation}
        onNewChat={onNewChat}
        isNewChatDisabled={isNewChatDisabled}
      />
      <Drawer
        open={isMobileOpen}
        onOpenChange={onMobileOpenChange}
      >
        <DrawerContent className="lg:hidden">
          <DrawerHeader className="text-left">
            <DrawerTitle>Chat history</DrawerTitle>
            <DrawerDescription>
              Resume a previous chat or start a new one.
            </DrawerDescription>
          </DrawerHeader>
          <div className="max-h-[60vh] overflow-y-auto px-4 pb-4">
            <HistoryList
              historyState={historyState}
              onResumeConversation={handleResumeConversation}
            />
            <Button
              type="button"
              variant="outline"
              onClick={handleNewChat}
              disabled={isNewChatDisabled}
              className="mt-4 w-full"
            >
              <Plus className="size-4" />
              New Chat
            </Button>
          </div>
        </DrawerContent>
      </Drawer>
    </>
  );
}

function DesktopHistoryPanel({
  historyState,
  isCollapsed,
  onToggleCollapsed,
  onResumeConversation,
  onNewChat,
  isNewChatDisabled,
}: Omit<ChatHistoryEntryProps, "isMobileOpen" | "onMobileOpenChange">) {
  if (isCollapsed) {
    return (
      <aside
        aria-label="Collapsed chat history"
        className="hidden w-12 flex-col items-center gap-2 border-r bg-background px-2 py-3 lg:fixed lg:bottom-0 lg:left-0 lg:top-[3.25rem] lg:z-20 lg:flex"
      >
        <Button
          type="button"
          variant="ghost"
          size="icon"
          aria-label="Expand chat history"
          onClick={onToggleCollapsed}
        >
          <PanelLeftOpen className="size-4" />
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          aria-label="New Chat"
          onClick={onNewChat}
          disabled={isNewChatDisabled}
        >
          <SquarePen className="size-4" />
        </Button>
      </aside>
    );
  }

  return (
    <aside
      aria-label="Chat history"
      className="hidden w-chat-sidebar flex-col border-r bg-background lg:fixed lg:bottom-0 lg:left-0 lg:top-[3.25rem] lg:z-20 lg:flex"
    >
      <div className="flex items-center justify-between gap-2 border-b p-3">
        <div className="flex min-w-0 items-center gap-2">
          <History className="size-4 shrink-0 text-muted-foreground" />
          <h2 className="truncate text-sm font-semibold">Chat history</h2>
        </div>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          aria-label="Collapse chat history"
          onClick={onToggleCollapsed}
          className="size-8"
        >
          <PanelLeftClose className="size-4" />
        </Button>
      </div>
      <div className="min-h-0 flex-1 overflow-y-auto p-2">
        <HistoryList
          historyState={historyState}
          onResumeConversation={onResumeConversation}
        />
      </div>
      <div className="border-t p-3">
        <Button
          type="button"
          variant="outline"
          onClick={onNewChat}
          disabled={isNewChatDisabled}
          className="w-full"
        >
          <Plus className="size-4" />
          New Chat
        </Button>
      </div>
    </aside>
  );
}

function HistoryList({
  historyState,
  onResumeConversation,
}: {
  historyState: ChatHistoryListState;
  onResumeConversation: (conversationId: number) => void;
}) {
  switch (historyState.status) {
    case "loading":
      return (
        <div className="px-2 py-3 text-sm text-muted-foreground">
          Loading recent chats...
        </div>
      );
    case "error":
      return (
        <div className="rounded-md border border-destructive/30 bg-destructive/5 p-3 text-sm text-destructive">
          {historyState.message}
        </div>
      );
    case "empty":
      return (
        <div className="px-2 py-3 text-sm text-muted-foreground">
          No recent chats yet.
        </div>
      );
    case "ready":
      return (
        <div className="flex flex-col gap-1">
          {historyState.conversations.map((conversation) => {
            const isActive =
              conversation.id === historyState.activeConversationId;

            return (
              <button
                key={conversation.id}
                type="button"
                onClick={() => onResumeConversation(conversation.id)}
                aria-current={isActive ? "page" : undefined}
                className={cn(
                  "flex w-full items-start gap-3 rounded-md px-3 py-2.5 text-left transition-colors hover:bg-accent hover:text-accent-foreground",
                  isActive && "bg-accent text-accent-foreground",
                )}
              >
                <span className="mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-full bg-muted text-muted-foreground">
                  <MessageSquare className="size-4" />
                </span>
                <span className="min-w-0 flex-1">
                  <span className="block truncate text-sm font-medium">
                    {conversation.title?.trim() || `Chat #${conversation.id}`}
                  </span>
                  <span className="block truncate text-xs text-muted-foreground">
                    {isActive
                      ? "Current conversation"
                      : formatConversationTimestamp(conversation)}
                  </span>
                </span>
              </button>
            );
          })}
        </div>
      );
  }

  const exhaustiveCheck: never = historyState;
  return exhaustiveCheck;
}

function formatConversationTimestamp(
  conversation: AIChatConversationSummary,
): string {
  const timestamp =
    conversation.last_message_at ??
    conversation.updated_at ??
    conversation.created_at;

  if (!timestamp) {
    return "Recent activity";
  }

  return new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  }).format(new Date(timestamp));
}
