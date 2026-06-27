import { useCallback, useEffect, useState } from "react";
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
  const [loadedUserId, setLoadedUserId] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(Boolean(userId));
  const [error, setError] = useState<string | null>(null);
  const [isCollapsed, setIsCollapsed] = useState(false);
  const [isMobileOpen, setIsMobileOpen] = useState(false);

  const loadConversations = useCallback(
    async (requestUserId: string, signal?: AbortSignal) => {
      setIsLoading(true);
      setError(null);

      try {
        const loadedConversations = await listAIChatConversations({ signal });
        if (!signal?.aborted) {
          setConversations(loadedConversations);
          setLoadedUserId(requestUserId);
        }
      } catch (listError: unknown) {
        if (!signal?.aborted) {
          setConversations([]);
          setLoadedUserId(null);
          setError(getErrorMessage(listError, "Could not load recent chats."));
        }
      } finally {
        if (!signal?.aborted) {
          setIsLoading(false);
        }
      }
    },
    [],
  );

  useEffect(() => {
    if (!userId) {
      setConversations([]);
      setLoadedUserId(null);
      setError(null);
      setIsLoading(false);
      return;
    }

    const controller = new AbortController();
    setConversations([]);
    setLoadedUserId(null);
    void loadConversations(userId, controller.signal);

    return () => controller.abort();
  }, [loadConversations, userId]);

  const refreshConversations = useCallback(async () => {
    if (!userId) {
      return;
    }

    await loadConversations(userId);
  }, [loadConversations, userId]);

  const loadedCurrentUserConversations =
    userId && loadedUserId === userId ? conversations : [];
  const hasLoadedCurrentUser = Boolean(userId && loadedUserId === userId);

  useEffect(() => {
    if (
      conversationId ||
      isLoading ||
      error ||
      !hasLoadedCurrentUser ||
      loadedCurrentUserConversations.length === 0
    ) {
      return;
    }

    onOpenConversation(loadedCurrentUserConversations[0].id, {
      replace: true,
    });
  }, [
    conversationId,
    error,
    hasLoadedCurrentUser,
    isLoading,
    loadedCurrentUserConversations,
    onOpenConversation,
  ]);

  const isPreparingEntry =
    !conversationId && (isLoading || !hasLoadedCurrentUser) && !error;
  const isAutoOpeningRecentChat =
    !conversationId &&
    !isLoading &&
    !error &&
    hasLoadedCurrentUser &&
    loadedCurrentUserConversations.length > 0;

  return {
    conversations: loadedCurrentUserConversations,
    error,
    isAutoOpeningRecentChat,
    isCollapsed,
    isLoading,
    isMobileOpen,
    isPreparingEntry,
    refreshConversations,
    setIsCollapsed,
    setIsMobileOpen,
  };
}
