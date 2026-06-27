import {
  History,
  MessageSquare,
  PanelLeftClose,
  PanelLeftOpen,
  Plus,
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

const MAX_VISIBLE_RECENT_CHATS = 8;

type ChatHistoryEntryProps = {
  conversations: AIChatConversationSummary[];
  activeConversationId: number | null;
  isLoading: boolean;
  error: string | null;
  isCollapsed: boolean;
  isMobileOpen: boolean;
  onMobileOpenChange: (open: boolean) => void;
  onToggleCollapsed: () => void;
  onResumeConversation: (conversationId: number) => void;
  onNewChat: () => void;
  isNewChatDisabled?: boolean;
};

export function ChatHistoryEntry({
  conversations,
  activeConversationId,
  isLoading,
  error,
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
        conversations={conversations}
        activeConversationId={activeConversationId}
        isLoading={isLoading}
        error={error}
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
              conversations={conversations}
              activeConversationId={activeConversationId}
              isLoading={isLoading}
              error={error}
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
  conversations,
  activeConversationId,
  isLoading,
  error,
  isCollapsed,
  onToggleCollapsed,
  onResumeConversation,
  onNewChat,
  isNewChatDisabled,
}: Omit<ChatHistoryEntryProps, "isMobileOpen" | "onMobileOpenChange">) {
  if (isCollapsed) {
    return (
      <aside className="hidden h-fit flex-col items-center gap-2 rounded-lg border bg-background p-2 lg:sticky lg:top-20 lg:flex">
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
          <Plus className="size-4" />
        </Button>
      </aside>
    );
  }

  return (
    <aside className="hidden h-fit flex-col rounded-lg border bg-background lg:sticky lg:top-20 lg:flex">
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
      <div className="max-h-[calc(100vh-15rem)] overflow-y-auto p-2">
        <HistoryList
          conversations={conversations}
          activeConversationId={activeConversationId}
          isLoading={isLoading}
          error={error}
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
  conversations,
  activeConversationId,
  isLoading,
  error,
  onResumeConversation,
}: {
  conversations: AIChatConversationSummary[];
  activeConversationId: number | null;
  isLoading: boolean;
  error: string | null;
  onResumeConversation: (conversationId: number) => void;
}) {
  if (isLoading) {
    return (
      <div className="px-2 py-3 text-sm text-muted-foreground">
        Loading recent chats...
      </div>
    );
  }

  if (error) {
    return (
      <div className="rounded-md border border-destructive/30 bg-destructive/5 p-3 text-sm text-destructive">
        {error}
      </div>
    );
  }

  if (conversations.length === 0) {
    return (
      <div className="px-2 py-3 text-sm text-muted-foreground">
        No recent chats yet.
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-1">
      {conversations.slice(0, MAX_VISIBLE_RECENT_CHATS).map((conversation) => {
        const isActive = conversation.id === activeConversationId;

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
