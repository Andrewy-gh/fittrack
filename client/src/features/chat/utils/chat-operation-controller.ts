import {
  reportAIChatTelemetry,
  stopAIChatRun,
  type AIChatConversationDetail,
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
} from "./chat-session-types";

export type ChatOperationPhase =
  | "idle"
  | "starting"
  | "streaming"
  | "recovering"
  | "stopping";

export type ChatOperationAttemptKind = "stream" | "resume" | "recovery";

export type ChatControllerAttachment = {
  readonly generation: number;
  active: boolean;
};

export type ChatOperationAttempt = {
  readonly operation: ChatOperation;
  readonly attachment: ChatControllerAttachment;
  readonly kind: ChatOperationAttemptKind;
  readonly controller: AbortController;
};

export type ChatOperation = {
  readonly attachment: ChatControllerAttachment;
  conversationId: number | null;
  runId: number | null;
  assistantMessageId: number | null;
  phase: Exclude<ChatOperationPhase, "idle">;
  currentAttempt: ChatOperationAttempt | null;
  adoptedConversationId: number | null;
};

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

/** Owns the browser-side lifecycle of the chat operation shown by one chat screen. */
export class ChatOperationController {
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
  private operation: ChatOperation | null = null;
  private routeLoad: RouteLoad | null = null;
  private snapshot: ChatOperationSnapshot = idleSnapshot;

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

  readonly subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  };

  readonly getSnapshot = () => this.snapshot;

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

  detach(attachment: ChatControllerAttachment): void {
    if (this.attachment !== attachment || !attachment.active) {
      return;
    }
    attachment.active = false;
    this.abortRouteLoad();
    this.detachOperation();
    this.attachment = null;
  }

  beginOperation(
    conversationId: number | null,
    phase: Exclude<ChatOperationPhase, "idle" | "stopping"> = "starting",
    activeRun?: { runId: number; assistantMessageId: number | null },
  ): ChatOperation | null {
    const attachment = this.attachment;
    if (!attachment?.active || this.operation) {
      return null;
    }
    const operation: ChatOperation = {
      attachment,
      conversationId,
      runId: activeRun?.runId ?? null,
      assistantMessageId: activeRun?.assistantMessageId ?? null,
      phase,
      currentAttempt: null,
      adoptedConversationId: null,
    };
    this.operation = operation;
    this.publishOperation();
    return operation;
  }

  beginAttempt(
    operation: ChatOperation,
    kind: ChatOperationAttemptKind,
  ): ChatOperationAttempt | null {
    if (!this.ownsOperation(operation)) {
      return null;
    }
    operation.currentAttempt?.controller.abort();
    const attempt: ChatOperationAttempt = {
      operation,
      attachment: operation.attachment,
      kind,
      controller: new AbortController(),
    };
    operation.currentAttempt = attempt;
    if (kind === "recovery" && operation.phase !== "recovering") {
      operation.phase = "recovering";
      this.publishOperation();
    }
    return attempt;
  }

  ownsOperation(operation: ChatOperation): boolean {
    return (
      this.operation === operation &&
      this.attachment === operation.attachment &&
      operation.attachment.active
    );
  }

  ownsAttempt(
    operation: ChatOperation,
    attempt: ChatOperationAttempt,
  ): boolean {
    return (
      this.ownsOperation(operation) &&
      operation.currentAttempt === attempt &&
      attempt.operation === operation &&
      attempt.attachment === operation.attachment
    );
  }

  finishAttempt(
    operation: ChatOperation,
    attempt: ChatOperationAttempt,
  ): boolean {
    if (!this.ownsAttempt(operation, attempt)) {
      return false;
    }
    operation.currentAttempt = null;
    return true;
  }

  finishOperation(operation: ChatOperation): boolean {
    if (!this.ownsOperation(operation)) {
      return false;
    }
    operation.currentAttempt?.controller.abort();
    operation.currentAttempt = null;
    this.operation = null;
    this.publishOperation();
    return true;
  }

  markStreaming(
    operation: ChatOperation,
    attempt: ChatOperationAttempt,
    runId: number | null,
    assistantMessageId: number,
  ): boolean {
    if (!this.ownsAttempt(operation, attempt)) {
      return false;
    }
    operation.runId = runId;
    operation.assistantMessageId = assistantMessageId;
    operation.phase = "streaming";
    this.publishOperation();
    return true;
  }

  setAssistantMessage(
    operation: ChatOperation,
    attempt: ChatOperationAttempt,
    assistantMessageId: number,
  ): boolean {
    if (!this.ownsAttempt(operation, attempt)) {
      return false;
    }
    operation.assistantMessageId = assistantMessageId;
    this.publishOperation();
    return true;
  }

  adoptConversation(operation: ChatOperation, conversationId: number): boolean {
    if (!this.ownsOperation(operation)) {
      return false;
    }
    operation.conversationId = conversationId;
    operation.adoptedConversationId = conversationId;
    this.publishOperation();
    return true;
  }

  onPromptStarted(conversationId: number): void {
    this.onPromptStartedCallback(conversationId);
  }

  onNewConversationCreated(conversationId: number): void {
    this.onNewConversationCreatedCallback(conversationId);
  }

  onConversationCreated(conversationId: number): Promise<void> {
    return this.onConversationCreatedCallback(conversationId);
  }

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

  handleRoute = (conversationId: number | null, prompt = ""): void => {
    const operation = this.operation;
    if (
      operation &&
      operation.adoptedConversationId !== null &&
      (conversationId === null ||
        conversationId === operation.adoptedConversationId)
    ) {
      if (conversationId === operation.adoptedConversationId) {
        operation.adoptedConversationId = null;
      }
      return;
    }

    if (!conversationId) {
      this.resetConversation(prompt);
      return;
    }
    void this.loadRouteConversation(conversationId);
  };

  cancelRouteLoad(): void {
    this.abortRouteLoad();
  }

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

  async loadRouteConversation(conversationId: number): Promise<void> {
    this.detachOperation();
    const loadResult = await this.loadConversation(conversationId);
    this.recordTelemetry({
      category: "load",
      outcome: loadResult.detail
        ? "load_completed"
        : classifyLoadOutcome(loadResult.aborted),
    });
    if (!loadResult.detail || !this.isAttached()) {
      return;
    }

    const activeRun = loadResult.detail.active_run;
    if (activeRun) {
      const operation = this.beginOperation(conversationId, "recovering", {
        runId: activeRun.id,
        assistantMessageId: activeRun.assistant_message_id,
      });
      if (!operation) return;
      try {
        await this.resumeOrRecoverActiveRun(
          conversationId,
          loadResult.detail,
          operation,
        );
      } finally {
        this.finishOperation(operation);
      }
      return;
    }

    const streamingMessage = loadResult.detail.messages.find(
      (message) => message.status === "streaming",
    );
    if (streamingMessage) {
      const operation = this.beginOperation(conversationId, "recovering");
      if (!operation) return;
      operation.assistantMessageId = streamingMessage.id;
      this.publishOperation();
      try {
        await this.recoverOpenedConversation(conversationId, operation);
      } finally {
        this.finishOperation(operation);
      }
    }
  }

  async loadConversation(
    id: number,
    opts?: ConversationRequestOptions,
    guard?: () => boolean,
  ): Promise<ConversationRequestResult> {
    const attachment = this.attachment;
    if (!attachment?.active) {
      return { detail: null, aborted: true };
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
      return await loadConversationRequest(id, opts, {
        state: this.state,
        signal: routeLoad.controller.signal,
        isCurrent,
      });
    } finally {
      if (this.routeLoad === routeLoad) {
        this.routeLoad = null;
        if (!opts?.silent) {
          this.state.setIsLoadingConversation(false);
        }
      }
    }
  }

  async recoverConversation(
    id: number,
    operation: ChatOperation,
    opts?: ConversationRequestOptions,
  ): Promise<ConversationRequestResult> {
    const attempt = this.beginAttempt(operation, "recovery");
    if (!attempt) return { detail: null, aborted: true };
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

  stopRun = async (): Promise<void> => {
    const operation = this.operation;
    const attachment = this.attachment;
    const conversationId = operation?.conversationId;
    const runId = operation?.runId;
    if (!operation || !attachment || !conversationId || !runId) return;

    try {
      const result = await stopAIChatRun(conversationId, runId);
      if (
        !attachment.active ||
        this.attachment !== attachment ||
        !this.ownsOperation(operation) ||
        operation.conversationId !== conversationId ||
        operation.runId !== runId
      ) {
        return;
      }
      this.finishOperation(operation);
      clearResumeCursor(conversationId);
      if (result.status === "stopped") {
        this.state.setMessages((current) =>
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
      } else {
        await this.loadRouteConversation(conversationId);
      }
    } catch (error) {
      if (
        attachment.active &&
        this.attachment === attachment &&
        this.ownsOperation(operation) &&
        operation.conversationId === conversationId &&
        operation.runId === runId
      ) {
        showErrorToast(error, "Failed to stop AI chat response");
      }
    }
  };

  private async resumeOrRecoverActiveRun(
    conversationId: number,
    detail: AIChatConversationDetail,
    operation: ChatOperation,
  ): Promise<void> {
    const attempt = this.beginAttempt(operation, "resume");
    if (!attempt) return;
    let resumeResult: ConversationRequestResult;
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
    if (!resumeResult.aborted && resumeResult.detail) {
      const resumeOutcome = classifyRecoveryOutcome({
        messages: resumeResult.detail.messages,
        aborted: false,
        error: resumeResult.error,
      });
      this.recordTelemetry({ category: "recovery", outcome: resumeOutcome });
      if (
        resumeOutcome !== "recovered_completed" &&
        resumeOutcome !== "recovery_aborted"
      ) {
        const resumeError =
          resumeResult.error ??
          new Error("Failed to resume AI chat conversation");
        this.recordTelemetry({
          category: "ux",
          outcome: "failure_toast_shown",
        });
        showErrorToast(resumeError, "Failed to recover AI chat conversation");
      }
      return;
    }
    if (resumeResult.aborted) {
      this.recordTelemetry({
        category: "recovery",
        outcome: "recovery_aborted",
      });
      return;
    }
    await this.recoverOpenedConversation(conversationId, operation);
  }

  private async recoverOpenedConversation(
    id: number,
    operation: ChatOperation,
  ): Promise<void> {
    await recoverLoadedConversation({
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
    const operation = this.operation;
    if (!operation) return;
    operation.currentAttempt?.controller.abort();
    operation.currentAttempt = null;
    this.operation = null;
    this.publishOperation();
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
    const operation = this.operation;
    this.snapshot = operation
      ? Object.freeze({
          phase: operation.phase,
          conversationId: operation.conversationId,
          runId: operation.runId,
          assistantMessageId: operation.assistantMessageId,
        })
      : idleSnapshot;
    for (const listener of this.listeners) listener();
  }
}
