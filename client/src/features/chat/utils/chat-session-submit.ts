import {
  createAIChatConversation,
  streamAIChatMessage,
} from "@/features/chat/api/ai-chat";
import {
  classifyRecoveryOutcome,
  classifyStreamInterruption,
  terminalStreamStage,
} from "@/features/chat/utils/ai-chat-observability";
import { getErrorMessage, showErrorToast } from "@/lib/errors";
import {
  clearResumeCursor,
  saveResumeCursor,
  updateStreamingMessageWithDelta,
  updateStreamingMessageWithDone,
  updateStreamingMessageWithError,
} from "./chat-resume";
import type {
  ChatOperation,
  ChatOperationAttempt,
  ChatOperationController,
} from "./chat-operation-controller";
import type {
  ChatSessionState,
  RecordChatTelemetry,
} from "./chat-session-types";

type SubmitPromptOptions = {
  conversationId: number | null;
  prompt: string;
  controller: ChatOperationController;
  recordTelemetry: RecordChatTelemetry;
  state: ChatSessionState;
};

export async function submitPrompt({
  conversationId,
  prompt,
  controller,
  recordTelemetry,
  state,
}: SubmitPromptOptions): Promise<void> {
  const nextPrompt = prompt.trim();
  if (!nextPrompt) return;

  const operation = controller.beginOperation(conversationId);
  if (!operation) return;
  controller.cancelRouteLoad();
  state.setPrompt("");
  let activeConversationId = conversationId;

  try {
    if (!activeConversationId) {
      const createdConversation = await createAIChatConversation();
      if (!controller.ownsOperation(operation)) return;
      activeConversationId = createdConversation.id;
      controller.adoptConversation(operation, activeConversationId);
      controller.onNewConversationCreated(activeConversationId);
      state.setConversation(createdConversation);
      await controller.onConversationCreated(activeConversationId);
      if (!controller.ownsOperation(operation)) return;
    }
  } catch (error) {
    if (!controller.ownsOperation(operation)) return;
    state.setPrompt(nextPrompt);
    controller.finishOperation(operation);
    showErrorToast(error, "Failed to create chat conversation");
    return;
  }

  if (!activeConversationId) {
    state.setPrompt(nextPrompt);
    controller.finishOperation(operation);
    return;
  }

  const baseTimestamp = new Date().toISOString();
  const tempUserId = -Date.now();
  const tempAssistantId = tempUserId - 1;
  let streamStarted = false;
  let shouldRefreshConversation = false;
  operation.assistantMessageId = tempAssistantId;
  const attempt = controller.beginAttempt(operation, "stream");
  if (!attempt) return;
  clearResumeCursor(activeConversationId);

  state.setMessages((current) => [
    ...current,
    {
      id: tempUserId,
      conversation_id: activeConversationId,
      role: "user",
      content: nextPrompt,
      status: "completed",
      created_at: baseTimestamp,
      updated_at: baseTimestamp,
      completed_at: baseTimestamp,
    },
    {
      id: tempAssistantId,
      conversation_id: activeConversationId,
      role: "assistant",
      content: "",
      status: "streaming",
      created_at: baseTimestamp,
      updated_at: baseTimestamp,
    },
  ]);

  try {
    const streamResult = await streamAIChatMessage(
      activeConversationId,
      nextPrompt,
      {
        onStart: (event) => {
          if (!controller.ownsAttempt(operation, attempt)) return;
          streamStarted = true;
          controller.onPromptStarted(activeConversationId);
          const assistantMessageId = event.message_id ?? tempAssistantId;
          controller.markStreaming(
            operation,
            attempt,
            event.run_id ?? null,
            assistantMessageId,
          );
          state.setMessages((current) =>
            current.map((message) =>
              message.id === tempAssistantId
                ? { ...message, id: assistantMessageId }
                : message,
            ),
          );
          saveResumeCursor(activeConversationId, {
            runId: event.run_id ?? 0,
            sequence: event.sequence ?? 0,
            assistantMessageId,
          });
        },
        onDelta: (event) => {
          if (!controller.ownsAttempt(operation, attempt)) return;
          const targetId = operation.assistantMessageId ?? tempAssistantId;
          state.setMessages((current) =>
            updateStreamingMessageWithDelta(
              current,
              targetId,
              event.delta ?? "",
            ),
          );
          if (event.sequence !== undefined) {
            saveResumeCursor(activeConversationId, {
              runId: event.run_id ?? 0,
              sequence: event.sequence,
              assistantMessageId: targetId,
            });
          }
        },
        onDone: (event) => {
          if (!controller.ownsAttempt(operation, attempt)) return;
          const targetId = operation.assistantMessageId ?? tempAssistantId;
          state.setMessages((current) =>
            updateStreamingMessageWithDone(current, targetId, event),
          );
          if (event.workout_draft) {
            state.setConversation((current) =>
              current
                ? {
                    ...current,
                    latest_workout_draft: event.workout_draft,
                    latest_workout_draft_status: { is_saved: false },
                  }
                : current,
            );
            state.setLatestWorkoutDraftMessageId(event.message_id ?? targetId);
          }
          clearResumeCursor(activeConversationId);
        },
        onErrorEvent: (event) => {
          if (!controller.ownsAttempt(operation, attempt)) return;
          const targetId = operation.assistantMessageId ?? tempAssistantId;
          state.setMessages((current) =>
            updateStreamingMessageWithError(current, targetId, event),
          );
          clearResumeCursor(activeConversationId);
          showErrorToast(
            { message: event.message ?? "AI chat streaming failed" },
            "AI chat streaming failed",
          );
        },
        signal: attempt.controller.signal,
      },
    );

    if (!controller.ownsAttempt(operation, attempt)) return;
    recordTelemetry({
      category: "stream",
      outcome: streamResult.endedWithError ? "server_error" : "completed",
      stage: terminalStreamStage(),
    });
    if (streamResult.endedWithError) {
      recordTelemetry({ category: "ux", outcome: "failure_toast_shown" });
    }
    shouldRefreshConversation = true;
  } catch (error) {
    if (!controller.ownsAttempt(operation, attempt)) return;
    if (!streamStarted && isPreflightAPIError(error)) {
      removeOptimisticMessages(state, tempUserId, tempAssistantId);
      state.setPrompt(nextPrompt);
      recordTelemetry({
        category: "stream",
        outcome: "server_error",
        stage: "pre_start",
      });
      recordTelemetry({ category: "ux", outcome: "failure_toast_shown" });
      showErrorToast(error, "Failed to stream AI chat response");
      return;
    }

    const streamTelemetry = classifyStreamInterruption(error, streamStarted);
    recordTelemetry({
      category: "stream",
      outcome: streamTelemetry.outcome,
      stage: streamTelemetry.stage,
    });
    if (streamTelemetry.outcome === "client_aborted") return;

    await handleStreamFailureRecovery({
      activeConversationId,
      nextPrompt,
      streamStarted,
      tempAssistantId,
      error,
      controller,
      operation,
      streamAttempt: attempt,
      recordTelemetry,
      state,
    });
  } finally {
    controller.finishAttempt(operation, attempt);
    controller.finishOperation(operation);
  }

  if (shouldRefreshConversation) {
    await controller.loadConversation(activeConversationId, { silent: true });
  }
}

async function handleStreamFailureRecovery({
  activeConversationId,
  nextPrompt,
  streamStarted,
  tempAssistantId,
  error,
  controller,
  operation,
  streamAttempt,
  recordTelemetry,
  state,
}: {
  activeConversationId: number;
  nextPrompt: string;
  streamStarted: boolean;
  tempAssistantId: number;
  error: unknown;
  controller: ChatOperationController;
  operation: ChatOperation;
  streamAttempt: ChatOperationAttempt;
  recordTelemetry: RecordChatTelemetry;
  state: ChatSessionState;
}) {
  if (!controller.ownsAttempt(operation, streamAttempt)) return;
  const recoveryPromise = controller.recoverConversation(
    activeConversationId,
    operation,
    { silent: true },
  );
  const {
    detail: recoveredDetail,
    aborted: recoveryAborted,
    error: recoveryError,
  } = await recoveryPromise;
  const submitFailure = recoveryError ?? error;
  const recoveryOutcome = classifyRecoveryOutcome({
    messages: recoveredDetail?.messages,
    prompt: nextPrompt,
    aborted: recoveryAborted,
    error: recoveryError,
  });
  recordTelemetry({ category: "recovery", outcome: recoveryOutcome });
  if (recoveryAborted || !controller.ownsOperation(operation)) return;

  if (!recoveredDetail) {
    const targetId = streamStarted
      ? (operation.assistantMessageId ?? tempAssistantId)
      : tempAssistantId;
    markAssistantMessageFailed(state, targetId, submitFailure);
  }

  if (recoveryOutcome !== "recovered_completed") {
    if (!streamStarted) state.setPrompt(nextPrompt);
    recordTelemetry({ category: "ux", outcome: "failure_toast_shown" });
    showErrorToast(submitFailure, "Failed to stream AI chat response");
    return;
  }

  if (!streamStarted) controller.onPromptStarted(activeConversationId);
  recordTelemetry({
    category: "ux",
    outcome: "failure_toast_suppressed_due_to_successful_recovery",
  });
}

function removeOptimisticMessages(
  state: ChatSessionState,
  tempUserId: number,
  tempAssistantId: number,
) {
  state.setMessages((current) =>
    current.filter(
      (message) => message.id !== tempUserId && message.id !== tempAssistantId,
    ),
  );
}

function markAssistantMessageFailed(
  state: ChatSessionState,
  messageId: number,
  error: unknown,
) {
  state.setMessages((current) =>
    current.map((message) =>
      message.id === messageId
        ? {
            ...message,
            status: "failed",
            error_message: getErrorMessage(error),
          }
        : message,
    ),
  );
}

function isPreflightAPIError(error: unknown): error is { message: string } {
  return (
    !(error instanceof Error) &&
    typeof error === "object" &&
    error !== null &&
    "message" in error &&
    typeof (error as { message?: unknown }).message === "string"
  );
}
