import {
  reportAIChatTelemetry,
  stopAIChatRun,
  type AIChatConversationDetail,
  type AIChatStopResponse,
  type AIChatTelemetryEvent,
} from "@/features/chat/api/ai-chat";
import {
  classifyLoadOutcome,
  classifyRecoveryOutcome,
} from "@/features/chat/utils/ai-chat-observability";
import { showErrorToast } from "@/lib/errors";
import { clearResumeCursor } from "./chat-resume";
import {
  loadConversation as loadConversationRequest,
  recoverConversation as recoverConversationRequest,
  recoverLoadedConversation,
  resumeConversation as resumeConversationRequest,
} from "./chat-session-recovery";
import { submitPrompt as submitPromptRequest } from "./chat-session-submit";
import type {
  ChatSessionState,
  ConversationRequestOptions,
  ConversationRequestResult,
  ConversationResumeResult,
} from "./chat-session-types";

import {
  transitionChatOperation,
  classifyChatConversation,
  type ChatControllerAttachment,
  type ChatOperation,
  type ChatOperationAttempt,
  type ChatOperationAttemptKind,
  type ChatOperationEvent,
  type ChatOperationState,
  type ChatOperationTarget,
  type ChatStopRequest,
  type ChatStopOutcome,
} from "./chat-operation-state";

import { ChatRequestFailure } from "./chat-session-types";

/** Public phases projected from the controller's internal lifecycle. */
export type ChatOperationPhase = ChatOperationState["phase"];

/** Stable screen-facing projection of lifecycle state. */
export type ChatOperationSnapshot = Readonly<{
  phase: ChatOperationPhase;
  conversationId: number | null;
  runId: number | null;
  assistantMessageId: number | null;
}>;

type ControllerOptions = {
  state: ChatSessionState;
  onConversationCreated: (conversationId: number) => Promise<void>;
  onPromptStarted: (conversationId: number) => void;
  onNewConversationCreated: (conversationId: number) => void;
};

type RouteLoad = {
  readonly attachment: ChatControllerAttachment;
  readonly controller: AbortController;
  readonly silent: boolean;
};

const idleSnapshot: ChatOperationSnapshot = Object.freeze({
  phase: "idle",
  conversationId: null,
  runId: null,
  assistantMessageId: null,
});

const stopRequestTimeoutMs = 15_000;
const stopReconciliationTimeoutMs = 15_000;

type BoundedRequestResult<T> =
  | { readonly kind: "completed"; readonly value: T }
  | { readonly kind: "failed"; readonly error: ChatRequestFailure };

function boundRequest<T>(
  operation: "load" | "stop",
  request: () => Promise<T>,
  timeoutMs: number,
  timeoutMessage: string,
  onTimeout?: () => void,
): Promise<BoundedRequestResult<T>> {
  return new Promise((resolve) => {
    const timeoutId = globalThis.setTimeout(() => {
      onTimeout?.();
      resolve({
        kind: "failed",
        error: new ChatRequestFailure(
          operation,
          new Error(timeoutMessage),
          "deadline",
        ),
      });
    }, timeoutMs);
    const failed = (cause: unknown) => {
      globalThis.clearTimeout(timeoutId);
      resolve({
        kind: "failed",
        error: new ChatRequestFailure(operation, cause),
      });
    };
    try {
      void request().then((value) => {
        globalThis.clearTimeout(timeoutId);
        resolve({ kind: "completed", value });
      }, failed);
    } catch (cause) {
      failed(cause);
    }
  });
}

/** Owns the browser-side lifecycle of the chat operation shown by one chat screen. */
export class ChatOperationController {
  /** UI state sinks owned by this screen. */
  readonly state: ChatSessionState;
  private readonly onConversationCreatedCallback: (
    conversationId: number,
  ) => Promise<void>;
  private readonly onPromptStartedCallback: (conversationId: number) => void;
  private readonly onNewConversationCreatedCallback: (
    conversationId: number,
  ) => void;
  private readonly listeners = new Set<() => void>();
  private attachmentGeneration = 0;
  private attachment: ChatControllerAttachment | null = null;
  private operationState: ChatOperationState = { phase: "idle" };
  private adoptedConversation: { operation: ChatOperation; id: number } | null =
    null;

  private routeLoad: RouteLoad | null = null;
  private snapshot: ChatOperationSnapshot = idleSnapshot;

  private get operation(): ChatOperation | null {
    return this.operationState.phase === "idle"
      ? null
      : this.operationState.operation;
  }

  private currentAttempt(): ChatOperationAttempt | null {
    const current = this.operationState;
    return current.phase !== "idle" &&
      current.phase !== "stopping" &&
      current.activity.kind === "attempt"
      ? current.activity.attempt
      : null;
  }

  private transition(event: ChatOperationEvent): boolean {
    const previousAttempt = this.currentAttempt();
    const next = transitionChatOperation(this.operationState, event);
    if (next === this.operationState) return false;
    this.operationState = next;
    if (this.adoptedConversation?.operation !== this.operation)
      this.adoptedConversation = null;
    // Invalidate ownership before abort: abort listeners may run synchronously.
    if (previousAttempt !== this.currentAttempt())
      previousAttempt?.controller.abort();
    this.publishOperation();
    return true;
  }

  /** Bind the lifecycle owner to one screen and its navigation callbacks. */
  constructor({
    state,
    onConversationCreated,
    onPromptStarted,
    onNewConversationCreated,
  }: ControllerOptions) {
    this.state = state;
    this.onConversationCreatedCallback = onConversationCreated;
    this.onPromptStartedCallback = onPromptStarted;
    this.onNewConversationCreatedCallback = onNewConversationCreated;
  }

  /** Subscribe React to changes in the screen-facing snapshot. */
  readonly subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  };

  /** Return a cached snapshot suitable for useSyncExternalStore. */
  readonly getSnapshot = () => this.snapshot;

  /** Start an attachment generation, invalidating work from earlier mounts. */
  attach(): ChatControllerAttachment {
    if (this.attachment?.active) {
      this.detach(this.attachment);
    }
    const attachment = {
      generation: ++this.attachmentGeneration,
      active: true,
    };
    this.attachment = attachment;
    return attachment;
  }

  /** Detach only the matching mount; this does not stop server generation. */
  detach(attachment: ChatControllerAttachment): void {
    if (this.attachment !== attachment || !attachment.active) {
      return;
    }
    attachment.active = false;
    this.abortRouteLoad();
    this.detachOperation();
    this.attachment = null;
  }

  /** Reserve this attached screen for a new prompt. */
  beginOperation(conversationId: number | null): ChatOperation | null {
    const attachment = this.attachment;
    if (!attachment?.active || this.operation) return null;
    const operation: ChatOperation = { attachment };
    return this.transition({ type: "begin", operation, conversationId })
      ? operation
      : null;
  }

  private beginRecovery(target: ChatOperationTarget): ChatOperation | null {
    const attachment = this.attachment;
    if (!attachment?.active || this.operation) return null;
    const operation: ChatOperation = { attachment };
    return this.transition({ type: "recover-opened", operation, target })
      ? operation
      : null;
  }

  /** Replace the active transport attempt without transferring operation ownership. */
  beginAttempt(
    operation: ChatOperation,
    kind: ChatOperationAttemptKind,
  ): ChatOperationAttempt | null {
    if (!this.ownsOperation(operation)) return null;
    const attempt: ChatOperationAttempt = {
      operation,
      attachment: operation.attachment,
      kind,
      controller: new AbortController(),
    };
    return this.transition({ type: "attempt-started", attempt })
      ? attempt
      : null;
  }

  /** Whether a callback still belongs to the attached screen's operation. */
  ownsOperation(operation: ChatOperation): boolean {
    return (
      this.operation === operation &&
      this.attachment === operation.attachment &&
      operation.attachment.active
    );
  }

  /** Whether an asynchronous callback still belongs to the current attempt. */
  ownsAttempt(
    operation: ChatOperation,
    attempt: ChatOperationAttempt,
  ): boolean {
    return (
      this.ownsOperation(operation) &&
      this.currentAttempt() === attempt &&
      attempt.operation === operation &&
      attempt.attachment === operation.attachment
    );
  }

  /** End an attempt without releasing a retained Stop-resolution operation. */
  finishAttempt(
    operation: ChatOperation,
    attempt: ChatOperationAttempt,
  ): boolean {
    return (
      this.ownsAttempt(operation, attempt) &&
      this.transition({ type: "attempt-finished", attempt })
    );
  }

  /** Ordinary workflow cleanup cannot release an unresolved Stop. */
  finishOperation(operation: ChatOperation): boolean {
    return (
      this.ownsOperation(operation) &&
      this.transition({ type: "workflow-finished", operation })
    );
  }

  /** Accept server stream identity only from the active attempt. */
  markStreaming(
    operation: ChatOperation,
    attempt: ChatOperationAttempt,
    runId: number | null,
    assistantMessageId: number,
  ): boolean {
    return (
      this.ownsAttempt(operation, attempt) &&
      this.transition({
        type: "stream-started",
        attempt,
        runId,
        assistantMessageId,
      })
    );
  }

  /** Associate an optimistic assistant message with its transport attempt. */
  setAssistantMessage(
    operation: ChatOperation,
    attempt: ChatOperationAttempt,
    assistantMessageId: number,
  ): boolean {
    return (
      this.ownsAttempt(operation, attempt) &&
      this.transition({
        type: "assistant-selected",
        attempt,
        assistantMessageId,
      })
    );
  }

  /** Keep a newly created conversation's route change attached to its operation. */
  adoptConversation(operation: ChatOperation, conversationId: number): boolean {
    if (!this.ownsOperation(operation)) return false;
    // Route adoption is independent of transport state and survives its updates.
    this.adoptedConversation = { operation, id: conversationId };
    return this.transition({
      type: "conversation-created",
      operation,
      conversationId,
    });
  }

  /** Notify the screen when a prompt has been acknowledged. */
  onPromptStarted(conversationId: number): void {
    this.onPromptStartedCallback(conversationId);
  }

  /** Migrate the draft into the newly created conversation. */
  onNewConversationCreated(conversationId: number): void {
    this.onNewConversationCreatedCallback(conversationId);
  }

  /** Navigate to a created conversation through the screen callback. */
  onConversationCreated(conversationId: number): Promise<void> {
    return this.onConversationCreatedCallback(conversationId);
  }

  /** Start the submission workflow while this controller owns the screen. */
  submitPrompt = ({
    conversationId,
    prompt,
  }: {
    conversationId: number | null;
    prompt: string;
  }): Promise<void> =>
    submitPromptRequest({
      conversationId,
      prompt,
      controller: this,
      recordTelemetry: this.recordTelemetry,
      state: this.state,
    });

  /** Adopt an expected creation route or load a different conversation. */
  handleRoute = (conversationId: number | null, prompt = ""): void => {
    const adoption = this.adoptedConversation;
    if (
      adoption &&
      this.ownsOperation(adoption.operation) &&
      (conversationId === null || conversationId === adoption.id)
    ) {
      if (conversationId === adoption.id) this.adoptedConversation = null;
      return;
    }

    if (!conversationId) {
      this.resetConversation(prompt);
      return;
    }
    this.state.setPrompt(prompt);
    void this.loadRouteConversation(conversationId);
  };

  /** Invalidate and cancel the current detail request. */
  cancelRouteLoad(): void {
    this.abortRouteLoad();
  }

  /** Clear local ownership and presentation without issuing a server Stop. */
  resetConversation = (prompt = ""): void => {
    this.abortRouteLoad();
    this.detachOperation();
    this.state.setConversation(null);
    this.state.setMessages([]);
    this.state.setPrompt(prompt);
    this.state.setLatestWorkoutDraftMessageId(null);
    this.state.setLoadError(null);
    this.state.setIsLoadingConversation(false);
  };

  /** Load a route and recover any observed server work under fresh ownership. */
  async loadRouteConversation(conversationId: number): Promise<void> {
    this.detachOperation();
    const loadResult = await this.loadConversation(conversationId);
    this.recordTelemetry({
      category: "load",
      outcome:
        loadResult.kind === "loaded"
          ? "load_completed"
          : classifyLoadOutcome(loadResult.kind === "cancelled"),
    });
    if (loadResult.kind !== "loaded" || !this.isAttached()) return;
    const activity = classifyChatConversation(loadResult.detail);
    if (activity.kind === "settled") return;
    const operation = this.beginRecovery(activity);
    if (!operation) return;
    try {
      if (activity.kind === "run") {
        await this.resumeOrRecoverActiveRun(
          conversationId,
          loadResult.detail,
          operation,
        );
      } else {
        await this.recoverOpenedConversation(conversationId, operation);
      }
    } finally {
      this.finishOperation(operation);
    }
  }

  /** Reload detail while the attachment, route request and optional owner remain current. */
  async loadConversation(
    id: number,
    opts?: ConversationRequestOptions,
    guard?: () => boolean,
  ): Promise<ConversationRequestResult> {
    const attachment = this.attachment;
    if (!attachment?.active) {
      return { kind: "cancelled" };
    }
    this.abortRouteLoad();
    const routeLoad: RouteLoad = {
      attachment,
      controller: new AbortController(),
      silent: opts?.silent === true,
    };
    this.routeLoad = routeLoad;
    if (!opts?.silent) {
      this.state.setIsLoadingConversation(true);
    }
    const isCurrent = () =>
      this.routeLoad === routeLoad &&
      this.attachment === attachment &&
      attachment.active &&
      (guard?.() ?? true);

    try {
      const request = () =>
        loadConversationRequest(id, opts, {
          state: this.state,
          signal: routeLoad.controller.signal,
          isCurrent,
        });
      if (!opts?.timeoutMs) return await request();
      const bounded = await boundRequest(
        "load",
        request,
        opts.timeoutMs,
        "AI chat reconciliation reload timed out",
        () => routeLoad.controller.abort(),
      );
      if (bounded.kind === "completed") return bounded.value;
      if (!isCurrent()) return { kind: "cancelled" };
      this.state.setLoadError(bounded.error.message);
      return { kind: "failed", error: bounded.error };
    } finally {
      if (this.routeLoad === routeLoad) {
        this.routeLoad = null;
        if (!opts?.silent) {
          this.state.setIsLoadingConversation(false);
        }
      }
    }
  }

  /** Poll recovery under an attempt that navigation or Stop can invalidate. */
  async recoverConversation(
    id: number,
    operation: ChatOperation,
    opts?: ConversationRequestOptions,
  ): Promise<ConversationRequestResult> {
    const attempt = this.beginAttempt(operation, "recovery");
    if (!attempt) return { kind: "cancelled" };
    try {
      return await recoverConversationRequest(id, opts, {
        state: this.state,
        signal: attempt.controller.signal,
        isCurrent: () => this.ownsAttempt(operation, attempt),
      });
    } finally {
      this.finishAttempt(operation, attempt);
    }
  }

  /** Request persisted Stop, retaining ownership until its outcome is reconciled. */
  stopRun = async (): Promise<void> => {
    const current = this.operationState;
    if (
      (current.phase !== "streaming" && current.phase !== "recovering") ||
      current.target.kind !== "run" ||
      !this.ownsOperation(current.operation)
    )
      return;
    const request: ChatStopRequest = {
      kind: current.completion === "stop-resolution" ? "retry" : "initial",
      operation: current.operation,
      target: current.target,
    };
    if (!this.transition({ type: "stop-requested", request })) return;
    const { conversationId, runId, assistantMessageId } = request.target;
    this.markAssistantPresentationStopped(assistantMessageId);

    const response = await boundRequest(
      "stop",
      () => stopAIChatRun(conversationId, runId),
      stopRequestTimeoutMs,
      "AI chat Stop request timed out",
    );
    const outcome: ChatStopOutcome =
      response.kind === "completed"
        ? { kind: "terminal", result: response.value }
        : { kind: "uncertain", error: response.error };
    if (
      !this.ownsStopRequest(request) ||
      !this.transition({ type: "stop-replied", request, outcome })
    )
      return;
    if (outcome.kind === "uncertain") {
      showErrorToast(outcome.error.cause, "Failed to stop AI chat response");
      await this.reconcileFailedStop(request);
      return;
    }
    clearResumeCursor(conversationId);
    if (outcome.result.status === "stopped" && request.kind === "initial") {
      this.state.setLoadError(null);
      this.applyStopResult(outcome.result);
      this.releaseOperation(request.operation);
      return;
    }
    await this.reconcileTerminalStopResult(request, outcome.result);
  };

  private ownsStopRequest(request: ChatStopRequest): boolean {
    return (
      this.ownsOperation(request.operation) &&
      this.operationState.phase === "stopping" &&
      this.operationState.request === request
    );
  }

  private markAssistantPresentationStopped(messageId: number | null): void {
    if (messageId === null) return;
    this.state.setMessages((current) =>
      current.map((message) =>
        message.id === messageId && message.status === "streaming"
          ? {
              ...message,
              status: "stopped",
              completed_at: new Date().toISOString(),
            }
          : message,
      ),
    );
  }

  private restoreAssistantStreamingPresentation(
    messageId: number | null,
  ): void {
    if (messageId === null) return;
    this.state.setMessages((current) =>
      current.map((message) =>
        message.id === messageId && message.status === "stopped"
          ? {
              ...message,
              status: "streaming",
              completed_at: undefined,
            }
          : message,
      ),
    );
  }

  private applyStopResult(result: AIChatStopResponse): void {
    this.state.setMessages((current) =>
      current.map((message) =>
        message.id === result.message_id
          ? {
              ...message,
              content: result.text,
              status: result.status,
              completed_at: new Date().toISOString(),
            }
          : message,
      ),
    );
  }

  private async reconcileTerminalStopResult(
    request: ChatStopRequest,
    result: AIChatStopResponse,
  ): Promise<void> {
    const { operation } = request;
    const { conversationId, runId } = request.target;
    const loadResult = await this.loadConversation(
      conversationId,
      { timeoutMs: stopReconciliationTimeoutMs },
      () => this.ownsStopRequest(request),
    );
    if (!this.ownsStopRequest(request)) {
      return;
    }

    if (loadResult.kind !== "loaded") {
      this.applyStopResult(result);
      this.releaseOperation(operation);
      return;
    }

    const nextRun = loadResult.detail.active_run;
    if (nextRun?.id === runId) {
      this.applyStopResult(result);
      this.releaseOperation(operation);
      return;
    }
    if (nextRun) {
      this.releaseOperation(operation);
      const nextOperation = this.beginRecovery({
        kind: "run",
        conversationId,
        runId: nextRun.id,
        assistantMessageId: nextRun.assistant_message_id,
      });
      if (!nextOperation) return;
      this.keepOperationUnresolved(nextOperation);
      await this.resumeOrRecoverAfterFailedStop(
        loadResult.detail,
        nextOperation,
      );
      return;
    }

    const streamingMessage = loadResult.detail.messages.find(
      (message) => message.status === "streaming",
    );
    if (streamingMessage) {
      this.keepOperationUnresolved(
        operation,
        this.messageTarget(conversationId, streamingMessage.id),
      );
      await this.recoverAfterFailedStop(conversationId, operation);
      return;
    }

    this.releaseOperation(operation);
  }

  private async reconcileFailedStop(request: ChatStopRequest): Promise<void> {
    const { operation } = request;
    const { conversationId } = request.target;
    const loadResult = await this.loadConversation(
      conversationId,
      { timeoutMs: stopReconciliationTimeoutMs },
      () => this.ownsStopRequest(request),
    );
    if (!this.ownsStopRequest(request)) {
      return;
    }

    if (loadResult.kind !== "loaded") {
      this.keepOperationUnresolved(operation);
      return;
    }

    await this.reconcileLoadedAfterFailedStop(operation, loadResult.detail);
  }

  private async reconcileLoadedAfterFailedStop(
    targetOperation: ChatOperation,
    detail: AIChatConversationDetail,
  ): Promise<void> {
    if (!this.ownsOperation(targetOperation)) return;

    const operation = this.retainOperationAfterFailedStop(
      targetOperation,
      detail,
    );
    if (!operation) {
      this.releaseOperation(targetOperation);
      return;
    }
    if (detail.active_run) {
      await this.resumeOrRecoverAfterFailedStop(detail, operation);
    } else {
      await this.recoverAfterFailedStop(detail.conversation.id, operation);
    }
  }

  private async resumeOrRecoverAfterFailedStop(
    detail: AIChatConversationDetail,
    operation: ChatOperation,
  ): Promise<void> {
    const attempt = this.beginAttempt(operation, "resume");
    if (!attempt) return;
    let resumeResult: ConversationResumeResult;
    try {
      resumeResult = await resumeConversationRequest(
        detail,
        (id, opts) =>
          this.loadConversation(id, opts, () =>
            this.ownsAttempt(operation, attempt),
          ),
        { controller: this, state: this.state },
        operation,
        attempt,
      );
    } finally {
      this.finishAttempt(operation, attempt);
    }

    if (!this.ownsOperation(operation) || resumeResult.kind === "cancelled")
      return;
    const observed =
      resumeResult.kind === "not-needed"
        ? resumeResult.detail
        : resumeResult.kind === "finished" &&
            resumeResult.reload.kind === "loaded"
          ? resumeResult.reload.detail
          : null;
    const terminalStatus =
      resumeResult.kind === "finished" &&
      resumeResult.stream.kind === "terminal"
        ? resumeResult.stream.status
        : null;
    if (observed && classifyChatConversation(observed).kind !== "settled") {
      const retained = this.retainOperationAfterFailedStop(operation, observed);
      if (!retained) return;
      if (retained === operation) {
        await this.recoverAfterFailedStop(observed.conversation.id, retained);
      } else {
        await this.resumeOrRecoverAfterFailedStop(observed, retained);
      }
      return;
    }
    if (terminalStatus || observed) {
      this.recordTelemetry({
        category: "recovery",
        outcome: terminalStatus
          ? terminalStatus === "failed"
            ? "recovered_failed"
            : "recovered_completed"
          : classifyRecoveryOutcome({ messages: observed?.messages }),
      });
      this.releaseOperation(operation);
      return;
    }

    await this.recoverAfterFailedStop(detail.conversation.id, operation);
  }

  private async recoverAfterFailedStop(
    conversationId: number,
    operation: ChatOperation,
  ): Promise<void> {
    const recoveryResult = await this.recoverOpenedConversation(
      conversationId,
      operation,
    );
    if (!this.ownsOperation(operation) || recoveryResult.kind === "cancelled")
      return;

    if (recoveryResult.kind !== "loaded") {
      this.keepOperationUnresolved(operation);
      return;
    }

    if (classifyChatConversation(recoveryResult.detail).kind === "settled") {
      this.releaseOperation(operation);
      return;
    }

    this.retainOperationAfterFailedStop(operation, recoveryResult.detail);
  }

  // Retain ownership without choosing whether the caller resumes or only polls.
  private retainOperationAfterFailedStop(
    operation: ChatOperation,
    detail: AIChatConversationDetail,
  ): ChatOperation | null {
    if (!this.ownsOperation(operation)) return null;
    const activity = classifyChatConversation(detail);
    if (activity.kind === "settled") return null;
    if (activity.kind === "run") {
      const replacement =
        this.snapshot.runId === activity.runId
          ? operation
          : { attachment: operation.attachment };
      this.keepOperationUnresolved(operation, activity, replacement);
      return replacement;
    }
    this.keepOperationUnresolved(
      operation,
      this.messageTarget(activity.conversationId, activity.assistantMessageId),
    );
    return operation;
  }

  // An orphan streaming message must not erase the known target for Stop retries.
  private messageTarget(
    conversationId: number,
    assistantMessageId: number,
  ): ChatOperationTarget {
    const current = this.operationState;
    return current.phase !== "idle" && current.target.kind === "run"
      ? { ...current.target, assistantMessageId }
      : { kind: "message", conversationId, assistantMessageId };
  }

  private keepOperationUnresolved(
    operation: ChatOperation,
    target?: ChatOperationTarget,
    replacement = operation,
  ): void {
    const current = this.operationState;
    if (
      !this.ownsOperation(operation) ||
      current.phase === "idle" ||
      current.target.kind === "new-conversation"
    )
      return;
    const nextTarget = target ?? current.target;
    if (
      this.transition({
        type: "recovery-unresolved",
        operation,
        target: nextTarget,
        replacement,
      })
    ) {
      this.restoreAssistantStreamingPresentation(
        "assistantMessageId" in nextTarget
          ? nextTarget.assistantMessageId
          : null,
      );
    }
  }

  private async resumeOrRecoverActiveRun(
    conversationId: number,
    detail: AIChatConversationDetail,
    operation: ChatOperation,
  ): Promise<void> {
    const attempt = this.beginAttempt(operation, "resume");
    if (!attempt) return;
    let resumeResult: ConversationResumeResult;
    try {
      resumeResult = await resumeConversationRequest(
        detail,
        (id, opts) =>
          this.loadConversation(id, opts, () =>
            this.ownsAttempt(operation, attempt),
          ),
        { controller: this, state: this.state },
        operation,
        attempt,
      );
    } finally {
      this.finishAttempt(operation, attempt);
    }

    if (!this.ownsOperation(operation)) return;
    if (
      resumeResult.kind === "cancelled" ||
      (resumeResult.kind === "finished" &&
        resumeResult.reload.kind === "cancelled")
    ) {
      this.recordTelemetry({
        category: "recovery",
        outcome: "recovery_aborted",
      });
      return;
    }
    const observed =
      resumeResult.kind === "not-needed"
        ? resumeResult.detail
        : resumeResult.kind === "finished" &&
            resumeResult.reload.kind === "loaded"
          ? resumeResult.reload.detail
          : null;
    const activity = observed ? classifyChatConversation(observed) : null;
    // A terminal event settles its run, not work started before the reload.
    if (observed && activity && activity.kind !== "settled") {
      if (activity.kind === "run" && activity.runId !== detail.active_run?.id) {
        this.releaseOperation(operation);
        const nextOperation = this.beginRecovery(activity);
        if (!nextOperation) return;
        try {
          await this.resumeOrRecoverActiveRun(
            conversationId,
            observed,
            nextOperation,
          );
        } finally {
          this.finishOperation(nextOperation);
        }
      } else {
        await this.recoverOpenedConversation(conversationId, operation);
      }
      return;
    }
    if (
      resumeResult.kind === "finished" &&
      resumeResult.stream.kind === "terminal"
    ) {
      this.recordTelemetry({
        category: "recovery",
        outcome:
          resumeResult.stream.status === "failed"
            ? "recovered_failed"
            : "recovered_completed",
      });
      return;
    }
    if (observed) {
      const resumeOutcome = classifyRecoveryOutcome({
        messages: observed.messages,
      });
      this.recordTelemetry({ category: "recovery", outcome: resumeOutcome });
      if (
        resumeOutcome !== "recovered_completed" &&
        resumeOutcome !== "recovery_aborted"
      ) {
        const error =
          resumeResult.kind === "finished" &&
          resumeResult.stream.kind === "unconfirmed"
            ? resumeResult.stream.error.cause
            : new Error("Failed to resume AI chat conversation");
        this.recordTelemetry({
          category: "ux",
          outcome: "failure_toast_shown",
        });
        showErrorToast(error, "Failed to recover AI chat conversation");
      }
      return;
    }

    await this.recoverOpenedConversation(conversationId, operation);
  }

  private async recoverOpenedConversation(
    id: number,
    operation: ChatOperation,
  ) {
    return recoverLoadedConversation({
      id,
      recoverConversation: (conversationId, opts) =>
        this.recoverConversation(conversationId, operation, opts),
      recordTelemetry: this.recordTelemetry,
      setLoadError: this.state.setLoadError,
    });
  }

  private readonly recordTelemetry = (event: AIChatTelemetryEvent): void => {
    void Promise.resolve(reportAIChatTelemetry(event)).catch((error) => {
      if (import.meta.env.DEV) {
        console.warn("Failed to record AI chat telemetry", error, event);
      }
    });
  };

  private isAttached(): boolean {
    return Boolean(this.attachment?.active);
  }

  private detachOperation(): void {
    this.transition({ type: "detached" });
  }

  private releaseOperation(operation: ChatOperation): boolean {
    return (
      this.ownsOperation(operation) &&
      this.transition({ type: "resolved", operation })
    );
  }

  private abortRouteLoad(): void {
    const routeLoad = this.routeLoad;
    if (!routeLoad) return;
    routeLoad.controller.abort();
    this.routeLoad = null;
    if (!routeLoad.silent) {
      this.state.setIsLoadingConversation(false);
    }
  }

  private publishOperation(): void {
    const current = this.operationState;
    const next: ChatOperationSnapshot =
      current.phase === "idle"
        ? idleSnapshot
        : {
            phase: current.phase,
            conversationId:
              current.target.kind === "new-conversation"
                ? null
                : current.target.conversationId,
            runId: current.target.kind === "run" ? current.target.runId : null,
            assistantMessageId:
              "assistantMessageId" in current.target
                ? current.target.assistantMessageId
                : null,
          };
    if (
      next.phase === this.snapshot.phase &&
      next.conversationId === this.snapshot.conversationId &&
      next.runId === this.snapshot.runId &&
      next.assistantMessageId === this.snapshot.assistantMessageId
    )
      return;
    this.snapshot = Object.freeze(next);
    for (const listener of this.listeners) listener();
  }
}
