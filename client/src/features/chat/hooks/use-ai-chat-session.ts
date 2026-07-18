import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import type {
  AIChatConversation,
  AIChatMessage,
} from "@/features/chat/api/ai-chat";
import { ChatOperationController } from "../utils/chat-operation-controller";
import type { ChatSessionState } from "../utils/chat-session-types";
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
  const [loadError, setLoadError] = useState<string | null>(null);
  const [isSavingWorkoutDraft, setIsSavingWorkoutDraft] = useState(false);
  const [latestWorkoutDraftMessageId, setLatestWorkoutDraftMessageId] =
    useState<number | null>(null);
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

  const state = useMemo<ChatSessionState>(
    () => ({
      setConversation,
      setMessages,
      setPrompt: setPromptState,
      setIsLoadingConversation,
      setLoadError,
      setIsSavingWorkoutDraft,
      setLatestWorkoutDraftMessageId,
    }),
    [],
  );

  useLayoutEffect(() => {
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

  const controller = useMemo(
    () =>
      new ChatOperationController({
        state,
        onConversationCreated: (createdConversationId) =>
          callbacksRef.current.onConversationCreated(createdConversationId),
        onPromptStarted: (startedConversationId) =>
          callbacksRef.current.onPromptStarted(startedConversationId),
        onNewConversationCreated: (createdConversationId) =>
          callbacksRef.current.onNewConversationCreated(createdConversationId),
      }),
    [state],
  );

  const operationSnapshot = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
    controller.getSnapshot,
  );

  useLayoutEffect(() => {
    const attachment = controller.attach();
    return () => controller.detach(attachment);
  }, [controller]);

  useEffect(() => {
    controller.handleRoute(conversationId, initialPrompt);
  }, [conversationId, controller, initialPrompt]);

  const submitPrompt = useCallback(
    () => controller.submitPrompt({ conversationId, prompt }),
    [conversationId, controller, prompt],
  );

  const submitPromptValue = useCallback(
    (value: string) =>
      controller.submitPrompt({ conversationId, prompt: value }),
    [conversationId, controller],
  );

  const saveLatestWorkoutDraft = useCallback(
    () =>
      saveLatestWorkoutDraftRequest({
        conversation,
        state,
      }),
    [conversation, state],
  );

  const isSubmitting = operationSnapshot.phase !== "idle";
  const canStop =
    operationSnapshot.runId !== null &&
    (operationSnapshot.phase === "streaming" ||
      operationSnapshot.phase === "recovering");

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
    resetConversation: controller.resetConversation,
    submitPrompt,
    submitPromptValue,
    saveLatestWorkoutDraft,
    stopRun: controller.stopRun,
    canStop,
  };
}
