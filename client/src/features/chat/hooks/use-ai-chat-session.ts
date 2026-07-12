import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import type {
  AIChatConversation,
  AIChatMessage,
} from "@/features/chat/api/ai-chat";
import { createAIChatSessionLifecycle } from "../utils/ai-chat-session-lifecycle";
import type {
  ChatSessionRefs,
  ChatSessionSetters,
} from "../utils/chat-session-types";
import { saveLatestWorkoutDraft as saveLatestWorkoutDraftRequest } from "../utils/chat-session-workout-draft";
import { stopAIChatRun } from "@/features/chat/api/ai-chat";
import { clearResumeCursor } from "../utils/chat-resume";
import { showErrorToast } from "@/lib/errors";

type UseAIChatSessionOptions = {
  conversationId: number | null;
  initialPrompt: string;
  onPromptChange: (prompt: string) => void;
  onPromptStarted: (conversationId: number) => void;
  onNewConversationCreated: (conversationId: number) => void;
  onConversationCreated: (conversationId: number) => Promise<void>;
};

export function useAIChatSession({
  conversationId,
  initialPrompt,
  onPromptChange,
  onPromptStarted,
  onNewConversationCreated,
  onConversationCreated,
}: UseAIChatSessionOptions) {
  const [conversation, setConversation] = useState<AIChatConversation | null>(
    null,
  );
  const [messages, setMessages] = useState<AIChatMessage[]>([]);
  const [prompt, setPromptState] = useState(initialPrompt);
  const [isLoadingConversation, setIsLoadingConversation] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [isSavingWorkoutDraft, setIsSavingWorkoutDraft] = useState(false);
  const [latestWorkoutDraftMessageId, setLatestWorkoutDraftMessageId] =
    useState<number | null>(null);
  const [activeRunId, setActiveRunId] = useState<number | null>(null);

  const pendingAssistantIdRef = useRef<number | null>(null);
  const loadAbortRef = useRef<AbortController | null>(null);
  const recoveryAbortRef = useRef<AbortController | null>(null);
  const resumeAbortRef = useRef<AbortController | null>(null);
  const streamAbortRef = useRef<AbortController | null>(null);
  const stopTargetRef = useRef({ conversationId, activeRunId });
  const callbacksRef = useRef({
    onPromptChange,
    onPromptStarted,
    onNewConversationCreated,
    onConversationCreated,
  });
  const setPrompt = useCallback((value: string) => {
    setPromptState(value);
    callbacksRef.current.onPromptChange(value);
  }, []);

  const refs = useMemo<ChatSessionRefs>(
    () => ({
      pendingAssistantIdRef,
      loadAbortRef,
      recoveryAbortRef,
      resumeAbortRef,
      streamAbortRef,
    }),
    [],
  );

  const setters = useMemo<ChatSessionSetters>(
    () => ({
      setConversation,
      setMessages,
      setPrompt: setPromptState,
      setIsLoadingConversation,
      setIsSubmitting,
      setLoadError,
      setIsSavingWorkoutDraft,
      setLatestWorkoutDraftMessageId,
      setActiveRunId,
    }),
    [],
  );

  useLayoutEffect(() => {
    stopTargetRef.current = { conversationId, activeRunId };
  }, [activeRunId, conversationId]);

  useEffect(() => {
    callbacksRef.current = {
      onPromptChange,
      onPromptStarted,
      onNewConversationCreated,
      onConversationCreated,
    };
  }, [
    onConversationCreated,
    onNewConversationCreated,
    onPromptChange,
    onPromptStarted,
  ]);

  const handlePromptStarted = useCallback(
    (conversationId: number) =>
      callbacksRef.current.onPromptStarted(conversationId),
    [],
  );

  const handleNewConversationCreated = useCallback(
    (conversationId: number) =>
      callbacksRef.current.onNewConversationCreated(conversationId),
    [],
  );

  const handleConversationCreated = useCallback(
    (createdConversationId: number) =>
      callbacksRef.current.onConversationCreated(createdConversationId),
    [],
  );

  const lifecycle = useMemo(
    () =>
      createAIChatSessionLifecycle({
        refs,
        setters,
        onConversationCreated: handleConversationCreated,
        onPromptStarted: handlePromptStarted,
        onNewConversationCreated: handleNewConversationCreated,
      }),
    [
      handleConversationCreated,
      handleNewConversationCreated,
      handlePromptStarted,
      refs,
      setters,
    ],
  );

  useEffect(() => {
    setters.setPrompt(initialPrompt);
    if (!conversationId) {
      lifecycle.resetConversation(initialPrompt);
      return;
    }

    void lifecycle.loadRouteConversation(conversationId);

    return () => {
      lifecycle.abortActiveRequests();
    };
  }, [conversationId, initialPrompt, lifecycle]);

  const submitPrompt = useCallback(
    () =>
      lifecycle.submitPrompt({
        conversationId,
        prompt,
        isSubmitting,
      }),
    [conversationId, isSubmitting, lifecycle, prompt],
  );

  const submitPromptValue = useCallback(
    (value: string) =>
      lifecycle.submitPrompt({
        conversationId,
        prompt: value,
        isSubmitting,
      }),
    [conversationId, isSubmitting, lifecycle],
  );

  const saveLatestWorkoutDraft = useCallback(
    () =>
      saveLatestWorkoutDraftRequest({
        conversation,
        setters,
      }),
    [conversation, setters],
  );

  const stopRun = useCallback(async () => {
    if (!conversationId) return;
    if (!activeRunId) return;
    try {
      const result = await stopAIChatRun(conversationId, activeRunId);
      const currentTarget = stopTargetRef.current;
      if (
        currentTarget.conversationId !== conversationId ||
        currentTarget.activeRunId !== activeRunId
      ) {
        return;
      }
      if (result.status === "stopped") {
        streamAbortRef.current?.abort();
        resumeAbortRef.current?.abort();
        recoveryAbortRef.current?.abort();
        clearResumeCursor(conversationId);
        setMessages((current) =>
          current.map((message) =>
            message.id === result.message_id
              ? {
                  ...message,
                  content: result.text,
                  status: "stopped",
                  completed_at: new Date().toISOString(),
                }
              : message,
          ),
        );
        setIsSubmitting(false);
      } else {
        await lifecycle.loadRouteConversation(conversationId);
      }
    } catch (error) {
      const currentTarget = stopTargetRef.current;
      if (
        currentTarget.conversationId !== conversationId ||
        currentTarget.activeRunId !== activeRunId
      ) {
        return;
      }
      showErrorToast(error, "Failed to stop AI chat response");
    }
  }, [activeRunId, conversationId, lifecycle]);

  return {
    conversation,
    messages,
    prompt,
    setPrompt,
    isLoadingConversation,
    isSubmitting,
    loadError,
    isSavingWorkoutDraft,
    latestWorkoutDraftMessageId,
    resetConversation: lifecycle.resetConversation,
    submitPrompt,
    submitPromptValue,
    saveLatestWorkoutDraft,
    stopRun,
    canStop: activeRunId !== null,
  };
}
