import { useCallback, useEffect, useMemo, useRef, useState } from "react";
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

  const pendingAssistantIdRef = useRef<number | null>(null);
  const loadAbortRef = useRef<AbortController | null>(null);
  const recoveryAbortRef = useRef<AbortController | null>(null);
  const resumeAbortRef = useRef<AbortController | null>(null);
  const streamAbortRef = useRef<AbortController | null>(null);
  const onConversationCreatedRef = useRef(onConversationCreated);
  const setPrompt = useCallback(
    (value: string) => {
      setPromptState(value);
      onPromptChange(value);
    },
    [onPromptChange],
  );

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
    }),
    [],
  );

  useEffect(() => {
    onConversationCreatedRef.current = onConversationCreated;
  }, [onConversationCreated]);

  const handleConversationCreated = useCallback(
    (createdConversationId: number) =>
      onConversationCreatedRef.current(createdConversationId),
    [],
  );

  const lifecycle = useMemo(
    () =>
      createAIChatSessionLifecycle({
        refs,
        setters,
        onConversationCreated: handleConversationCreated,
        onPromptStarted,
        onNewConversationCreated,
      }),
    [
      handleConversationCreated,
      onNewConversationCreated,
      onPromptStarted,
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
  };
}
