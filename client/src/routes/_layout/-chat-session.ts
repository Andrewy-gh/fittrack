import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  createAIChatConversation,
  reportAIChatTelemetry,
  type AIChatConversation,
  type AIChatConversationDetail,
  type AIChatMessage,
  type AIChatTelemetryEvent,
} from "@/lib/api/ai-chat";
import {
  classifyLoadOutcome,
  classifyRecoveryOutcome,
} from "@/lib/ai-chat-observability";
import { showErrorToast } from "@/lib/errors";
import {
  loadConversation as loadConversationRequest,
  recoverConversation as recoverConversationRequest,
  recoverLoadedConversation,
  resumeConversation as resumeConversationRequest,
} from "./-chat-session-recovery";
import { submitPrompt as submitPromptRequest } from "./-chat-session-submit";
import type {
  ChatSessionRefs,
  ChatSessionSetters,
} from "./-chat-session-types";
import { saveLatestWorkoutDraft as saveLatestWorkoutDraftRequest } from "./-chat-session-workout-draft";

type UseAIChatSessionOptions = {
  conversationId: number | null;
  onConversationCreated: (conversationId: number) => Promise<void>;
};

export function useAIChatSession({
  conversationId,
  onConversationCreated,
}: UseAIChatSessionOptions) {
  const [conversation, setConversation] = useState<AIChatConversation | null>(
    null,
  );
  const [messages, setMessages] = useState<AIChatMessage[]>([]);
  const [prompt, setPrompt] = useState("");
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
      setPrompt,
      setIsLoadingConversation,
      setIsSubmitting,
      setLoadError,
      setIsSavingWorkoutDraft,
      setLatestWorkoutDraftMessageId,
    }),
    [],
  );

  const recordTelemetry = useCallback((event: AIChatTelemetryEvent) => {
    void Promise.resolve(reportAIChatTelemetry(event)).catch((error) => {
      if (import.meta.env.DEV) {
        console.warn("Failed to record AI chat telemetry", error, event);
      }
    });
  }, []);

  const loadConversation = useCallback(
    (id: number, opts?: { silent?: boolean }) =>
      loadConversationRequest(id, opts, { refs, setters }),
    [refs, setters],
  );

  const recoverConversation = useCallback(
    (id: number, opts?: { silent?: boolean }) =>
      recoverConversationRequest(id, opts, { refs, setters }),
    [refs, setters],
  );

  const resumeConversation = useCallback(
    (detail: Parameters<typeof resumeConversationRequest>[0]) =>
      resumeConversationRequest(detail, loadConversation, { refs, setters }),
    [loadConversation, refs, setters],
  );

  const resetConversation = useCallback(() => {
    loadAbortRef.current?.abort();
    recoveryAbortRef.current?.abort();
    resumeAbortRef.current?.abort();
    streamAbortRef.current?.abort();
    setConversation(null);
    setMessages([]);
    setLatestWorkoutDraftMessageId(null);
    setLoadError(null);
    setIsLoadingConversation(false);
  }, []);

  useEffect(() => {
    if (!conversationId) {
      resetConversation();
      return;
    }

    const activeConversationId = conversationId;
    void loadConversationForRoute();

    return () => {
      loadAbortRef.current?.abort();
      recoveryAbortRef.current?.abort();
      resumeAbortRef.current?.abort();
      streamAbortRef.current?.abort();
    };

    async function loadConversationForRoute() {
      const loadResult = await loadConversation(activeConversationId);
      recordTelemetry({
        category: "load",
        outcome: loadResult.detail
          ? "load_completed"
          : classifyLoadOutcome(loadResult.aborted),
      });

      if (loadResult.detail?.active_run) {
        await resumeOrRecoverActiveRun(loadResult.detail);
        return;
      }

      if (
        loadResult.detail?.messages.some(
          (message) => message.status === "streaming",
        )
      ) {
        await recoverOpenedConversation(activeConversationId);
      }
    }

    async function resumeOrRecoverActiveRun(detail: AIChatConversationDetail) {
      const resumeResult = await resumeConversation(detail);
      if (!resumeResult.aborted && resumeResult.detail) {
        const resumeOutcome = classifyRecoveryOutcome({
          messages: resumeResult.detail.messages,
          aborted: false,
          error: resumeResult.error,
        });
        recordTelemetry({
          category: "recovery",
          outcome: resumeOutcome,
        });
        if (
          resumeOutcome !== "recovered_completed" &&
          resumeOutcome !== "recovery_aborted"
        ) {
          const resumeError =
            resumeResult.error ??
            new Error("Failed to resume AI chat conversation");
          recordTelemetry({
            category: "ux",
            outcome: "failure_toast_shown",
          });
          showErrorToast(resumeError, "Failed to recover AI chat conversation");
        }
        return;
      }

      if (resumeResult.aborted) {
        recordTelemetry({
          category: "recovery",
          outcome: "recovery_aborted",
        });
        return;
      }

      await recoverOpenedConversation(activeConversationId);
    }

    async function recoverOpenedConversation(id: number) {
      await recoverLoadedConversation({
        id,
        recoverConversation,
        recordTelemetry,
        setLoadError,
      });
    }
  }, [
    conversationId,
    loadConversation,
    recoverConversation,
    recordTelemetry,
    resetConversation,
    resumeConversation,
  ]);

  const createNewChat = useCallback(async () => {
    try {
      const created = await createAIChatConversation();
      streamAbortRef.current?.abort();
      recoveryAbortRef.current?.abort();
      loadAbortRef.current?.abort();
      setConversation(created);
      setMessages([]);
      setLatestWorkoutDraftMessageId(null);
      setLoadError(null);
      await onConversationCreated(created.id);
    } catch (error) {
      showErrorToast(error, "Failed to create chat conversation");
    }
  }, [onConversationCreated]);

  const submitPrompt = useCallback(
    () =>
      submitPromptRequest({
        conversationId,
        prompt,
        isSubmitting,
        onConversationCreated,
        loadConversation,
        recoverConversation,
        recordTelemetry,
        refs,
        setters,
      }),
    [
      conversationId,
      isSubmitting,
      loadConversation,
      onConversationCreated,
      prompt,
      recordTelemetry,
      recoverConversation,
      refs,
      setters,
    ],
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
    createNewChat,
    submitPrompt,
    saveLatestWorkoutDraft,
  };
}
