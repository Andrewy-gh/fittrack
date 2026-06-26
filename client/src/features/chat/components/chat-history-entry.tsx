import { useEffect, useState, type ReactNode } from "react";
import { MessageSquare, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  listAIChatConversations,
  type AIChatConversationSummary,
} from "@/features/chat/api/ai-chat";
import { getErrorMessage } from "@/lib/errors";

const MAX_VISIBLE_RECENT_CHATS = 5;

type ChatHistoryEntryProps = {
  userId: string;
  fallback: ReactNode;
  onResumeConversation: (conversationId: number) => void;
  onNewChat: () => void;
  isNewChatDisabled?: boolean;
};

export function ChatHistoryEntry({
  userId,
  fallback,
  onResumeConversation,
  onNewChat,
  isNewChatDisabled,
}: ChatHistoryEntryProps) {
  const [recentConversations, setRecentConversations] = useState<
    AIChatConversationSummary[]
  >([]);
  const [isLoadingRecentConversations, setIsLoadingRecentConversations] =
    useState(false);
  const [recentConversationsError, setRecentConversationsError] = useState<
    string | null
  >(null);

  useEffect(() => {
    const controller = new AbortController();
    setIsLoadingRecentConversations(true);
    setRecentConversationsError(null);

    void listAIChatConversations({ signal: controller.signal })
      .then((conversations) => {
        setRecentConversations(conversations);
      })
      .catch((error: unknown) => {
        if (controller.signal.aborted) {
          return;
        }
        setRecentConversations([]);
        setRecentConversationsError(
          getErrorMessage(error, "Could not load recent chats."),
        );
      })
      .finally(() => {
        if (!controller.signal.aborted) {
          setIsLoadingRecentConversations(false);
        }
      });

    return () => controller.abort();
  }, [userId]);

  if (isLoadingRecentConversations) {
    return (
      <div className="text-sm text-muted-foreground">
        Loading recent chats...
      </div>
    );
  }

  if (recentConversations.length === 0) {
    return recentConversationsError ? (
      <div className="flex flex-col gap-4">
        <div className="rounded-md border border-destructive/30 bg-destructive/5 p-3 text-sm text-destructive">
          {recentConversationsError}
        </div>
        {fallback}
      </div>
    ) : (
      fallback
    );
  }

  return (
    <div className="flex min-h-[70vh] flex-col justify-center gap-5 py-8">
      <div className="flex flex-col gap-2">
        <h1 className="text-2xl font-semibold tracking-tight text-foreground sm:text-3xl">
          AI Chat
        </h1>
        <p className="text-sm text-muted-foreground">
          Resume a recent conversation or start a new chat.
        </p>
      </div>

      <div className="overflow-hidden rounded-lg border bg-background">
        <div className="border-b px-4 py-3 text-sm font-medium">
          Recent chats
        </div>
        <div className="divide-y">
          {recentConversations
            .slice(0, MAX_VISIBLE_RECENT_CHATS)
            .map((conversation) => (
              <button
                key={conversation.id}
                type="button"
                onClick={() => onResumeConversation(conversation.id)}
                className="flex w-full items-center gap-3 px-4 py-3 text-left transition-colors hover:bg-accent hover:text-accent-foreground"
              >
                <span className="flex size-9 shrink-0 items-center justify-center rounded-full bg-muted text-muted-foreground">
                  <MessageSquare className="size-4" />
                </span>
                <span className="min-w-0 flex-1">
                  <span className="block truncate text-sm font-medium">
                    {conversation.title?.trim() || `Chat #${conversation.id}`}
                  </span>
                  <span className="block text-xs text-muted-foreground">
                    {formatConversationTimestamp(conversation)}
                  </span>
                </span>
              </button>
            ))}
        </div>
      </div>

      <Button
        type="button"
        variant="outline"
        onClick={onNewChat}
        disabled={isNewChatDisabled}
        className="w-full sm:w-fit"
      >
        <Plus className="size-4" />
        New Chat
      </Button>
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
