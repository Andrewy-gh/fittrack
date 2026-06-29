import { useCallback, useEffect, useRef, useState } from "react";
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

export type ChatHistoryEntryState =
  | { status: "openingLatestChat" }
  | { status: "historyLoadError"; message: string }
  | { status: "ready" };

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
  const requestIdRef = useRef(0);

  const loadConversations = useCallback(
    async (requestUserId: string, signal?: AbortSignal) => {
      const requestId = requestIdRef.current + 1;
      requestIdRef.current = requestId;
      const isCurrentRequest = () =>
        !signal?.aborted && requestIdRef.current === requestId;

      setIsLoading(true);
      setError(null);

      try {
        const loadedConversations = await listAIChatConversations({ signal });
        if (isCurrentRequest()) {
          setConversations(loadedConversations);
          setLoadedUserId(requestUserId);
        }
      } catch (listError: unknown) {
        if (isCurrentRequest()) {
          setConversations([]);
          setLoadedUserId(null);
          setError(getErrorMessage(listError, "Could not load recent chats."));
        }
      } finally {
        if (isCurrentRequest()) {
          setIsLoading(false);
        }
      }
    },
    [],
  );

  useEffect(() => {
    if (!userId) {
      requestIdRef.current += 1;
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

  const entryState = getChatHistoryEntryState({
    conversationId,
    conversations: loadedCurrentUserConversations,
    error,
    hasLoadedCurrentUser,
    isLoading,
  });

  return {
    conversations: loadedCurrentUserConversations,
    entryState,
    error,
    isCollapsed,
    isLoading,
    isMobileOpen,
    refreshConversations,
    setIsCollapsed,
    setIsMobileOpen,
  };
}

function getChatHistoryEntryState({
  conversationId,
  conversations,
  error,
  hasLoadedCurrentUser,
  isLoading,
}: {
  conversationId: number | null;
  conversations: AIChatConversationSummary[];
  error: string | null;
  hasLoadedCurrentUser: boolean;
  isLoading: boolean;
}): ChatHistoryEntryState {
  if (conversationId !== null) {
    return { status: "ready" };
  }

  if (error) {
    return { status: "historyLoadError", message: error };
  }

  if (isLoading || !hasLoadedCurrentUser || conversations.length > 0) {
    return { status: "openingLatestChat" };
  }

  return { status: "ready" };
}
