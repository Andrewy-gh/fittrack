import {
  createAIChatConversation,
  reportAIChatTelemetry,
  type AIChatConversationDetail,
  type AIChatTelemetryEvent,
} from "@/features/chat/api/ai-chat";
import {
  classifyLoadOutcome,
  classifyRecoveryOutcome,
} from "@/features/chat/utils/ai-chat-observability";
import { showErrorToast } from "@/lib/errors";
import {
  loadConversation as loadConversationRequest,
  recoverConversation as recoverConversationRequest,
  recoverLoadedConversation,
  resumeConversation as resumeConversationRequest,
} from "./chat-session-recovery";
import { submitPrompt as submitPromptRequest } from "./chat-session-submit";
import type { ChatSessionRefs, ChatSessionSetters } from "./chat-session-types";

type AIChatSessionLifecycleOptions = {
  refs: ChatSessionRefs;
  setters: ChatSessionSetters;
  onConversationCreated: (conversationId: number) => Promise<void>;
};

type SubmitPromptOptions = {
  conversationId: number | null;
  prompt: string;
  isSubmitting: boolean;
};

export function createAIChatSessionLifecycle({
  refs,
  setters,
  onConversationCreated,
}: AIChatSessionLifecycleOptions) {
  const recordTelemetry = (event: AIChatTelemetryEvent) => {
    void Promise.resolve(reportAIChatTelemetry(event)).catch((error) => {
      if (import.meta.env.DEV) {
        console.warn("Failed to record AI chat telemetry", error, event);
      }
    });
  };

  const loadConversation = (
    id: number,
    opts?: Parameters<typeof loadConversationRequest>[1],
  ) => loadConversationRequest(id, opts, { refs, setters });

  const recoverConversation = (
    id: number,
    opts?: Parameters<typeof recoverConversationRequest>[1],
  ) => recoverConversationRequest(id, opts, { refs, setters });

  const resumeConversation = (detail: AIChatConversationDetail) =>
    resumeConversationRequest(detail, loadConversation, { refs, setters });

  const abortActiveRequests = () => {
    refs.loadAbortRef.current?.abort();
    refs.recoveryAbortRef.current?.abort();
    refs.resumeAbortRef.current?.abort();
    refs.streamAbortRef.current?.abort();
  };

  const resetConversation = () => {
    abortActiveRequests();
    setters.setConversation(null);
    setters.setMessages([]);
    setters.setLatestWorkoutDraftMessageId(null);
    setters.setLoadError(null);
    setters.setIsLoadingConversation(false);
  };

  const loadRouteConversation = async (conversationId: number) => {
    const loadResult = await loadConversation(conversationId);
    recordTelemetry({
      category: "load",
      outcome: loadResult.detail
        ? "load_completed"
        : classifyLoadOutcome(loadResult.aborted),
    });

    if (loadResult.detail?.active_run) {
      await resumeOrRecoverActiveRun(conversationId, loadResult.detail);
      return;
    }

    if (
      loadResult.detail?.messages.some(
        (message) => message.status === "streaming",
      )
    ) {
      await recoverOpenedConversation(conversationId);
    }
  };

  const createNewChat = async () => {
    try {
      const created = await createAIChatConversation();
      refs.streamAbortRef.current?.abort();
      refs.recoveryAbortRef.current?.abort();
      refs.loadAbortRef.current?.abort();
      setters.setConversation(created);
      setters.setMessages([]);
      setters.setLatestWorkoutDraftMessageId(null);
      setters.setLoadError(null);
      await onConversationCreated(created.id);
    } catch (error) {
      showErrorToast(error, "Failed to create chat conversation");
    }
  };

  const submitPrompt = ({
    conversationId,
    prompt,
    isSubmitting,
  }: SubmitPromptOptions) =>
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
    });

  async function resumeOrRecoverActiveRun(
    conversationId: number,
    detail: AIChatConversationDetail,
  ) {
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

    await recoverOpenedConversation(conversationId);
  }

  async function recoverOpenedConversation(id: number) {
    await recoverLoadedConversation({
      id,
      recoverConversation,
      recordTelemetry,
      setLoadError: setters.setLoadError,
    });
  }

  return {
    abortActiveRequests,
    createNewChat,
    loadRouteConversation,
    resetConversation,
    submitPrompt,
  };
}
