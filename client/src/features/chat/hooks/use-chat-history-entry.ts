import { useEffect, useState } from "react";
import {
  listAIChatConversations,
  type AIChatConversationSummary,
} from "@/features/chat/api/ai-chat";
import { getErrorMessage } from "@/lib/errors";

type OpenConversationOptions = {
  replace?: boolean;
};

type UseChatHistoryEntryOptions = {
  userId?: string;
  conversationId: number | null;
  onOpenConversation: (
    conversationId: number,
    options?: OpenConversationOptions,
  ) => void;
};

export function useChatHistoryEntry({
  userId,
  conversationId,
  onOpenConversation,
}: UseChatHistoryEntryOptions) {
  const [conversations, setConversations] = useState<
    AIChatConversationSummary[]
  >([]);
  const [isLoading, setIsLoading] = useState(Boolean(userId));
  const [error, setError] = useState<string | null>(null);
  const [isCollapsed, setIsCollapsed] = useState(false);
  const [isMobileOpen, setIsMobileOpen] = useState(false);

  useEffect(() => {
    if (!userId) {
      return;
    }

    const controller = new AbortController();
    setIsLoading(true);
    setError(null);

    void listAIChatConversations({ signal: controller.signal })
      .then((loadedConversations) => {
        setConversations(loadedConversations);
      })
      .catch((listError: unknown) => {
        if (controller.signal.aborted) {
          return;
        }
        setConversations([]);
        setError(getErrorMessage(listError, "Could not load recent chats."));
      })
      .finally(() => {
        if (!controller.signal.aborted) {
          setIsLoading(false);
        }
      });

    return () => controller.abort();
  }, [userId]);

  useEffect(() => {
    if (conversationId || isLoading || error || conversations.length === 0) {
      return;
    }

    onOpenConversation(conversations[0].id, { replace: true });
  }, [conversationId, conversations, error, isLoading, onOpenConversation]);

  const isPreparingEntry = !conversationId && isLoading && !error;
  const isAutoOpeningRecentChat =
    !conversationId && !isLoading && !error && conversations.length > 0;

  return {
    conversations,
    error,
    isAutoOpeningRecentChat,
    isCollapsed,
    isLoading,
    isMobileOpen,
    isPreparingEntry,
    setIsCollapsed,
    setIsMobileOpen,
  };
}
