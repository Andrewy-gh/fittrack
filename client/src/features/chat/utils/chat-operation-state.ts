import type {
  AIChatConversationDetail,
  AIChatStopResponse,
} from "@/features/chat/api/ai-chat";
import type { ChatRequestFailure } from "./chat-session-types";

/** One screen attachment; identity also separates React Strict Mode replays. */
export type ChatControllerAttachment = {
  readonly generation: number;
  active: boolean;
};

/** Identity of an operation. Workflow callers cannot mutate its lifecycle. */
export type ChatOperation = {
  readonly attachment: ChatControllerAttachment;
};

/** Transport work owned by an operation. */
export type ChatOperationAttemptKind = "stream" | "resume" | "recovery";

/** Correlates callbacks and owns cancellation for one transport attempt. */
export type ChatOperationAttempt = {
  readonly operation: ChatOperation;
  readonly attachment: ChatControllerAttachment;
  readonly kind: ChatOperationAttemptKind;
  readonly controller: AbortController;
};

/** A known conversation, orphan message, or server run; Stop requires a run. */
export type ChatOperationTarget =
  | { readonly kind: "conversation"; readonly conversationId: number }
  | {
      readonly kind: "message";
      readonly conversationId: number;
      readonly assistantMessageId: number;
    }
  | {
      readonly kind: "run";
      readonly conversationId: number;
      readonly runId: number;
      readonly assistantMessageId: number | null;
    };

/** One Stop request, distinct even when a retry targets the same run. */
export type ChatStopRequest = {
  /** A retained retry must reload even an idempotent stopped response. */
  readonly kind: "initial" | "retry";
  readonly operation: ChatOperation;
  readonly target: Extract<ChatOperationTarget, { kind: "run" }>;
};

/** Stop failure is uncertainty about persistence, not a terminal run outcome. */
export type ChatStopOutcome =
  | { readonly kind: "terminal"; readonly result: AIChatStopResponse }
  | { readonly kind: "uncertain"; readonly error: ChatRequestFailure };

/** Persisted activity observed in a detail response, which is not an atomic snapshot. */
export type ChatConversationActivity =
  | Exclude<ChatOperationTarget, { kind: "conversation" }>
  | { readonly kind: "settled" };

/** A run takes precedence over orphan messages; both must be absent to observe settlement. */
export function classifyChatConversation(
  detail: AIChatConversationDetail,
): ChatConversationActivity {
  const conversationId = detail.conversation.id;
  if (detail.active_run) {
    return {
      kind: "run",
      conversationId,
      runId: detail.active_run.id,
      assistantMessageId: detail.active_run.assistant_message_id,
    };
  }
  const message = detail.messages.find(
    (candidate) => candidate.status === "streaming",
  );
  return message
    ? { kind: "message", conversationId, assistantMessageId: message.id }
    : { kind: "settled" };
}

type Activity =
  | { readonly kind: "waiting" }
  | { readonly kind: "attempt"; readonly attempt: ChatOperationAttempt };

/** Ordinary workflows release on cleanup; failed Stop requires resolution. */
type ChatCompletionPolicy = "workflow" | "stop-resolution";

/** The browser lifecycle, independent of persisted server status and UI text. */
export type ChatOperationState =
  | { readonly phase: "idle" }
  | {
      readonly phase: "starting";
      readonly operation: ChatOperation;
      readonly target:
        | { readonly kind: "new-conversation" }
        | ChatOperationTarget;
      readonly activity: Activity;
    }
  | {
      readonly phase: "streaming" | "recovering";
      readonly operation: ChatOperation;
      readonly target: ChatOperationTarget;
      readonly completion: ChatCompletionPolicy;
      readonly activity: Activity;
    }
  | {
      readonly phase: "stopping";
      readonly operation: ChatOperation;
      readonly target: Extract<ChatOperationTarget, { kind: "run" }>;
      readonly request: ChatStopRequest;
      readonly step:
        | { readonly kind: "requesting" }
        | { readonly kind: "reconciling"; readonly outcome: ChatStopOutcome };
    };

/** Facts accepted by the lifecycle owner; every async fact carries identity. */
export type ChatOperationEvent =
  | {
      readonly type: "begin";
      readonly operation: ChatOperation;
      readonly conversationId: number | null;
    }
  | {
      readonly type: "recover-opened";
      readonly operation: ChatOperation;
      readonly target: ChatOperationTarget;
    }
  | {
      readonly type: "conversation-created";
      readonly operation: ChatOperation;
      readonly conversationId: number;
    }
  | { readonly type: "attempt-started"; readonly attempt: ChatOperationAttempt }
  | {
      readonly type: "attempt-finished";
      readonly attempt: ChatOperationAttempt;
    }
  | {
      readonly type: "assistant-selected";
      readonly attempt: ChatOperationAttempt;
      readonly assistantMessageId: number;
    }
  | {
      readonly type: "stream-started";
      readonly attempt: ChatOperationAttempt;
      readonly runId: number | null;
      readonly assistantMessageId: number;
    }
  | { readonly type: "stop-requested"; readonly request: ChatStopRequest }
  | {
      readonly type: "stop-replied";
      readonly request: ChatStopRequest;
      readonly outcome: ChatStopOutcome;
    }
  | {
      readonly type: "recovery-unresolved";
      readonly operation: ChatOperation;
      readonly target: ChatOperationTarget;
      readonly replacement: ChatOperation;
    }
  | { readonly type: "workflow-finished"; readonly operation: ChatOperation }
  | { readonly type: "resolved"; readonly operation: ChatOperation }
  | { readonly type: "detached" };

const idle: ChatOperationState = { phase: "idle" };
const waiting: Activity = { kind: "waiting" };

function owns(state: ChatOperationState, operation: ChatOperation): boolean {
  return state.phase !== "idle" && state.operation === operation;
}

function ownsAttempt(
  state: ChatOperationState,
  attempt: ChatOperationAttempt,
): boolean {
  return (
    owns(state, attempt.operation) &&
    state.phase !== "idle" &&
    state.phase !== "stopping" &&
    state.activity.kind === "attempt" &&
    state.activity.attempt === attempt
  );
}

function completion(
  state: Exclude<ChatOperationState, { phase: "idle" }>,
): ChatCompletionPolicy {
  return state.phase === "starting"
    ? "workflow"
    : state.phase === "stopping"
      ? "stop-resolution"
      : state.completion;
}

function withMessage(
  target: ChatOperationTarget,
  assistantMessageId: number,
): ChatOperationTarget {
  return target.kind === "run"
    ? { ...target, assistantMessageId }
    : {
        kind: "message",
        conversationId: target.conversationId,
        assistantMessageId,
      };
}

/**
 * Decide a transition without I/O or reading the clock. The controller commits
 * this value before aborting old work, so synchronous callbacks lose ownership.
 * Stale or inapplicable events return the same state and have no effects.
 */
export function transitionChatOperation(
  state: ChatOperationState,
  event: ChatOperationEvent,
): ChatOperationState {
  switch (event.type) {
    case "begin":
      if (state.phase !== "idle") return state;
      return {
        phase: "starting",
        operation: event.operation,
        activity: waiting,
        target:
          event.conversationId === null
            ? { kind: "new-conversation" }
            : { kind: "conversation", conversationId: event.conversationId },
      };
    case "recover-opened":
      if (state.phase !== "idle") return state;
      return {
        phase: "recovering",
        operation: event.operation,
        target: event.target,
        completion: "workflow",
        activity: waiting,
      };
    case "conversation-created":
      if (state.phase !== "starting" || !owns(state, event.operation))
        return state;
      return {
        ...state,
        target: { kind: "conversation", conversationId: event.conversationId },
      };
    case "attempt-started": {
      if (
        state.phase === "idle" ||
        state.phase === "stopping" ||
        !owns(state, event.attempt.operation) ||
        state.target.kind === "new-conversation"
      )
        return state;
      const activity: Activity = { kind: "attempt", attempt: event.attempt };
      switch (event.attempt.kind) {
        case "stream":
          return state.phase === "starting" ? { ...state, activity } : state;
        case "resume":
          return state.phase === "recovering" ? { ...state, activity } : state;
        case "recovery":
          return {
            phase: "recovering",
            operation: state.operation,
            target: state.target,
            completion: completion(state),
            activity,
          };
        default:
          return casesHandled(event.attempt.kind);
      }
    }
    case "attempt-finished":
      if (
        state.phase === "idle" ||
        state.phase === "stopping" ||
        !ownsAttempt(state, event.attempt)
      )
        return state;
      return { ...state, activity: waiting };
    case "assistant-selected":
      if (
        state.phase === "idle" ||
        state.phase === "stopping" ||
        state.target.kind === "new-conversation" ||
        !ownsAttempt(state, event.attempt)
      )
        return state;
      return {
        ...state,
        target: withMessage(state.target, event.assistantMessageId),
      };
    case "stream-started":
      if (
        state.phase === "idle" ||
        state.phase === "stopping" ||
        state.target.kind === "new-conversation" ||
        !ownsAttempt(state, event.attempt)
      )
        return state;
      return {
        phase: "streaming",
        operation: state.operation,
        completion: completion(state),
        activity: state.activity,
        target:
          event.runId === null
            ? {
                kind: "message",
                conversationId: state.target.conversationId,
                assistantMessageId: event.assistantMessageId,
              }
            : {
                kind: "run",
                conversationId: state.target.conversationId,
                runId: event.runId,
                assistantMessageId: event.assistantMessageId,
              },
      };
    case "stop-requested":
      if (
        (state.phase !== "streaming" && state.phase !== "recovering") ||
        state.target.kind !== "run" ||
        !owns(state, event.request.operation) ||
        state.target !== event.request.target
      )
        return state;
      return {
        phase: "stopping",
        operation: state.operation,
        target: state.target,
        request: event.request,
        step: { kind: "requesting" },
      };
    case "stop-replied":
      if (
        state.phase !== "stopping" ||
        state.request !== event.request ||
        state.step.kind !== "requesting"
      )
        return state;
      return {
        ...state,
        step: { kind: "reconciling", outcome: event.outcome },
      };
    case "recovery-unresolved":
      if (state.phase === "idle" || !owns(state, event.operation)) return state;
      return {
        phase: "recovering",
        operation: event.replacement,
        target: event.target,
        completion: "stop-resolution",
        activity: waiting,
      };
    case "workflow-finished":
      if (
        state.phase === "idle" ||
        !owns(state, event.operation) ||
        completion(state) !== "workflow"
      )
        return state;
      return idle;
    case "resolved":
      return owns(state, event.operation) ? idle : state;
    case "detached":
      return state.phase === "idle" ? state : idle;
    default:
      return casesHandled(event);
  }
}

function casesHandled(event: never): never {
  throw new Error(`Unexpected chat lifecycle event: ${String(event)}`);
}
