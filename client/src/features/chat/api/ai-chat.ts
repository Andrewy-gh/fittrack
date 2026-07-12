import {
  getAiConversations,
  getAiConversationsById,
  postAiChatTelemetry,
  postAiConversations,
  postAiConversationsByIdLatestWorkoutDraftSave,
  postAiConversationsByIdMessagesRecover,
} from "@/client";
import type { ApiError } from "@/lib/errors";
import { applyLocalDevAuthHeader } from "@/lib/local-dev-auth";
import { stackClientApp } from "@/stack";

const BASE_URL = import.meta.env.VITE_API_BASE_URL || "/api";

export type AIChatConversation = {
  id: number;
  title?: string;
  latest_workout_draft?: AIWorkoutDraft;
  latest_workout_draft_status?: AIWorkoutDraftStatus;
  created_at: string;
  updated_at: string;
  last_message_at?: string;
};

export type AIChatConversationSummary = Pick<
  AIChatConversation,
  "id" | "title" | "created_at" | "updated_at" | "last_message_at"
>;

export type AIWorkoutDraftStatus = {
  source_run_id?: number;
  is_saved: boolean;
  saved_workout_id?: number;
  saved_at?: string;
};

export type AIChatMessage = {
  id: number;
  conversation_id: number;
  role: "user" | "assistant";
  content: string;
  status: "streaming" | "completed" | "failed" | "stopped";
  error_message?: string;
  created_at: string;
  updated_at: string;
  completed_at?: string;
};

export type AIChatActiveRun = {
  id: number;
  assistant_message_id: number;
  status: "streaming" | "completed" | "failed" | "stopped";
  latest_sequence: number;
};

export type AIChatConversationDetail = {
  conversation: AIChatConversation;
  messages: AIChatMessage[];
  active_run?: AIChatActiveRun;
};

export type AIChatRecoveryResponse = {
  conversation_id: number;
  run_id?: number;
  status: "queued" | "not_needed";
};

export type AIChatStopResponse = {
  conversation_id: number;
  run_id: number;
  message_id: number;
  status: "stopped" | "completed" | "failed";
  text: string;
  sequence: number;
};

export async function stopAIChatRun(
  conversationId: number,
  runId: number,
): Promise<AIChatStopResponse> {
  const response = await fetch(
    `${BASE_URL}/ai/conversations/${conversationId}/runs/${runId}/stop`,
    { method: "POST", headers: await getAuthHeaders() },
  );
  if (!response.ok) throw await readApiError(response);
  return (await response.json()) as AIChatStopResponse;
}

export type AISaveLatestWorkoutDraftResponse = {
  conversation: AIChatConversation;
  workout_id: number;
};

export type AIWorkoutSetInput = {
  weight?: number;
  reps: number;
  setType: "warmup" | "working";
};

export type AIWorkoutExerciseInput = {
  name: string;
  sets: AIWorkoutSetInput[];
};

export type AIWorkoutDraft = {
  date: string;
  notes?: string;
  workoutFocus?: string;
  exercises: AIWorkoutExerciseInput[];
};

type AIChatStreamEventContext = {
  request_id?: string;
  conversation_id?: number;
  run_id?: number;
  message_id?: number;
  model?: string;
  sequence?: number;
};

export type AIChatStreamStartEvent = {
  type: "start";
} & AIChatStreamEventContext;

export type AIChatStreamDeltaEvent = {
  type: "delta";
  delta: string;
} & AIChatStreamEventContext;

export type AIChatStreamDoneEvent = {
  type: "done";
  text: string;
  status?: "completed" | "stopped";
  workout_draft?: AIWorkoutDraft;
} & AIChatStreamEventContext;

export type AIChatStreamErrorEvent = {
  type: "error";
  message: string;
} & AIChatStreamEventContext;

export type AIChatStreamEvent =
  | AIChatStreamStartEvent
  | AIChatStreamDeltaEvent
  | AIChatStreamDoneEvent
  | AIChatStreamErrorEvent;

export type AIChatStreamResult = {
  doneEvent?: AIChatStreamDoneEvent | AIChatStreamErrorEvent;
  endedWithError: boolean;
};

export type AIChatTelemetryCategory = "stream" | "recovery" | "load" | "ux";

export type AIChatTelemetryStage = "pre_start" | "post_start" | "terminal";

export type AIChatTelemetryEvent = {
  category: AIChatTelemetryCategory;
  outcome:
    | "completed"
    | "server_error"
    | "transport_ended_pre_terminal"
    | "client_aborted"
    | "recovered_completed"
    | "recovered_failed"
    | "recovery_timeout"
    | "recovery_aborted"
    | "load_completed"
    | "load_failed"
    | "load_aborted_stale"
    | "failure_toast_shown"
    | "failure_toast_suppressed_due_to_successful_recovery";
  stage?: AIChatTelemetryStage;
};

type ParsedSSEChunk = {
  event: AIChatStreamEvent;
  id?: string;
};

type UnknownRecord = Record<string, unknown>;

type StreamHandlers = {
  onStart?: (event: AIChatStreamStartEvent) => void;
  onDelta?: (event: AIChatStreamDeltaEvent) => void;
  onDone?: (event: AIChatStreamDoneEvent) => void;
  onErrorEvent?: (event: AIChatStreamErrorEvent) => void;
  signal?: AbortSignal;
};

type ConversationPollOptions = {
  signal?: AbortSignal;
  intervalMs?: number;
  timeoutMs?: number;
  onStreaming?: (detail: AIChatConversationDetail) => Promise<void> | void;
};

type ConversationRequestOptions = {
  signal?: AbortSignal;
};

type ConversationRecoveryOptions = {
  signal?: AbortSignal;
};

const defaultConversationPollIntervalMs = 1000;
const defaultConversationPollTimeoutMs = 55000;

async function getAuthHeaders(contentType = false): Promise<Headers> {
  const headers = new Headers();
  if (contentType) {
    headers.set("Content-Type", "application/json");
  }

  if (!stackClientApp) {
    return applyLocalDevAuthHeader(headers);
  }

  const user = await stackClientApp.getUser();
  if (!user) {
    return applyLocalDevAuthHeader(headers);
  }

  const { accessToken } = await user.getAuthJson();
  if (accessToken) {
    headers.set("x-stack-access-token", accessToken);
    return headers;
  }

  return applyLocalDevAuthHeader(headers);
}

async function readApiError(response: Response): Promise<ApiError> {
  try {
    return (await response.json()) as ApiError;
  } catch {
    return {
      message: `${response.status} ${response.statusText}`,
    };
  }
}

export async function createAIChatConversation(): Promise<AIChatConversation> {
  const response = await postAiConversations({
    throwOnError: true,
  });

  return response.data as AIChatConversation;
}

export async function listAIChatConversations(
  options: ConversationRequestOptions = {},
): Promise<AIChatConversationSummary[]> {
  const response = await getAiConversations({
    signal: options.signal,
    throwOnError: true,
  });

  return response.data as AIChatConversationSummary[];
}

export async function reportAIChatTelemetry(
  event: AIChatTelemetryEvent,
): Promise<void> {
  await postAiChatTelemetry({
    body: event,
    keepalive: true,
    throwOnError: true,
  });
}

export async function getAIChatConversation(
  conversationId: number,
  options: ConversationRequestOptions = {},
): Promise<AIChatConversationDetail> {
  const response = await getAiConversationsById({
    path: { id: conversationId },
    signal: options.signal,
    throwOnError: true,
  });

  return response.data as AIChatConversationDetail;
}

export async function requestAIChatMessageRecovery(
  conversationId: number,
  options: ConversationRecoveryOptions = {},
): Promise<AIChatRecoveryResponse> {
  const response = await postAiConversationsByIdMessagesRecover({
    path: { id: conversationId },
    signal: options.signal,
    throwOnError: true,
  });

  return response.data as AIChatRecoveryResponse;
}

export async function saveAIChatLatestWorkoutDraft(
  conversationId: number,
  options: ConversationRequestOptions = {},
): Promise<AIChatConversation> {
  const response = await postAiConversationsByIdLatestWorkoutDraftSave({
    path: { id: conversationId },
    signal: options.signal,
    throwOnError: true,
  });

  return (response.data as AISaveLatestWorkoutDraftResponse).conversation;
}

export async function streamAIChatMessage(
  conversationId: number,
  prompt: string,
  handlers: StreamHandlers = {},
): Promise<AIChatStreamResult> {
  const response = await fetch(
    `${BASE_URL}/ai/conversations/${conversationId}/messages/stream`,
    {
      method: "POST",
      headers: await getAuthHeaders(true),
      body: JSON.stringify({ prompt }),
      signal: handlers.signal,
    },
  );

  if (!response.ok) {
    throw await readApiError(response);
  }

  return readAIChatStreamResponse(response, handlers);
}

export async function resumeAIChatMessageStream(
  conversationId: number,
  runId: number,
  afterSequence: number,
  handlers: StreamHandlers = {},
): Promise<AIChatStreamResult> {
  const query = new URLSearchParams({
    runId: String(runId),
    afterSequence: String(afterSequence),
  });
  const response = await fetch(
    `${BASE_URL}/ai/conversations/${conversationId}/messages/stream/resume?${query.toString()}`,
    {
      method: "GET",
      headers: await getAuthHeaders(),
      signal: handlers.signal,
    },
  );

  if (!response.ok) {
    throw await readApiError(response);
  }

  return readAIChatStreamResponse(response, handlers);
}

async function readAIChatStreamResponse(
  response: Response,
  handlers: StreamHandlers,
): Promise<AIChatStreamResult> {
  if (!response.body) {
    throw new Error("No body in SSE response");
  }

  const reader = response.body.pipeThrough(new TextDecoderStream()).getReader();
  let buffer = "";
  let lastEventId = "";
  let sawTerminalEvent = false;

  try {
    while (true) {
      const { done, value } = await reader.read();
      if (done) {
        break;
      }

      buffer += value;
      const chunks = buffer.split("\n\n");
      buffer = chunks.pop() ?? "";

      for (const chunk of chunks) {
        const parsed = parseSSEChunk(chunk);
        if (!parsed) {
          continue;
        }
        if (parsed.id && parsed.id === lastEventId) {
          continue;
        }
        lastEventId = parsed.id ?? lastEventId;

        const event = parsed.event;

        switch (event.type) {
          case "start":
            handlers.onStart?.(event);
            break;
          case "delta":
            handlers.onDelta?.(event);
            break;
          case "done":
            sawTerminalEvent = true;
            handlers.onDone?.(event);
            return { doneEvent: event, endedWithError: false };
          case "error":
            sawTerminalEvent = true;
            handlers.onErrorEvent?.(event);
            return { doneEvent: event, endedWithError: true };
          default:
            return casesHandled(event);
        }
      }
    }
  } finally {
    reader.releaseLock();
  }

  if (!sawTerminalEvent) {
    throw new Error("AI chat stream ended before a terminal event");
  }

  return { endedWithError: false };
}

export async function pollAIChatConversationUntilSettled(
  conversationId: number,
  options: ConversationPollOptions = {},
): Promise<AIChatConversationDetail> {
  const intervalMs = options.intervalMs ?? defaultConversationPollIntervalMs;
  const timeoutMs = options.timeoutMs ?? defaultConversationPollTimeoutMs;
  const deadline = Date.now() + timeoutMs;

  while (true) {
    throwIfAborted(options.signal);
    const detail = await getAIChatConversation(conversationId, {
      signal: options.signal,
    });
    throwIfAborted(options.signal);
    const hasStreamingMessage = detail.messages.some(
      (message) => message.status === "streaming",
    );
    if (!hasStreamingMessage) {
      return detail;
    }
    await options.onStreaming?.(detail);
    throwIfAborted(options.signal);
    if (Date.now() >= deadline) {
      throw new Error(
        "AI chat recovery timed out while waiting for persisted conversation state",
      );
    }
    await delay(intervalMs, options.signal);
  }
}

function parseSSEChunk(chunk: string): ParsedSSEChunk | null {
  const lines = chunk.split("\n");
  const dataLines: string[] = [];
  let id = "";

  for (const line of lines) {
    if (line.startsWith("id:")) {
      id = line.replace(/^id:\s*/, "").trim();
    }
    if (line.startsWith("data:")) {
      dataLines.push(line.replace(/^data:\s*/, ""));
    }
  }

  if (dataLines.length === 0) {
    return null;
  }

  return {
    event: parseAIChatStreamEvent(dataLines.join("\n")),
    id,
  };
}

function parseAIChatStreamEvent(rawEventJson: string): AIChatStreamEvent {
  let rawEvent: unknown;
  try {
    rawEvent = JSON.parse(rawEventJson);
  } catch {
    throw new Error("Invalid AI chat stream event JSON");
  }

  const event = parseRecord(rawEvent, "AI chat stream event");
  const type = parseRequiredString(event, "type");
  switch (type) {
    case "start":
      return { type, ...parseStreamEventContext(event) };
    case "delta":
      return {
        type,
        delta: parseRequiredString(event, "delta"),
        ...parseStreamEventSequence(event),
      };
    case "done": {
      const workoutDraft = parseOptionalWorkoutDraft(event.workout_draft);
      return {
        type,
        text: parseOptionalStringValue(event, "text") ?? "",
        ...(parseOptionalStringValue(event, "status") === "stopped"
          ? { status: "stopped" as const }
          : {}),
        ...parseDoneEventContext(event),
        ...(workoutDraft === undefined ? {} : { workout_draft: workoutDraft }),
      };
    }
    case "error":
      return {
        type,
        message: parseRequiredString(event, "message"),
        ...parseErrorEventContext(event),
      };
    default:
      throw new Error(`Unknown AI chat stream event type: ${type}`);
  }
}

function parseStreamEventContext(
  event: UnknownRecord,
): AIChatStreamEventContext {
  return {
    ...parseStreamEventIdentity(event),
    ...parseOptionalStringProperty(event, "request_id"),
    ...parseOptionalStringProperty(event, "model"),
    ...parseStreamEventSequence(event),
  };
}

function parseDoneEventContext(
  event: UnknownRecord,
): Omit<AIChatStreamEventContext, "request_id"> {
  return {
    ...parseStreamEventIdentity(event),
    ...parseOptionalStringProperty(event, "model"),
    ...parseStreamEventSequence(event),
  };
}

function parseErrorEventContext(
  event: UnknownRecord,
): Omit<AIChatStreamEventContext, "request_id" | "model" | "sequence"> {
  return parseStreamEventIdentity(event);
}

function parseStreamEventIdentity(
  event: UnknownRecord,
): Pick<AIChatStreamEventContext, "conversation_id" | "run_id" | "message_id"> {
  return {
    ...parseOptionalNumberProperty(event, "conversation_id"),
    ...parseOptionalNumberProperty(event, "run_id"),
    ...parseOptionalNumberProperty(event, "message_id"),
  };
}

function parseStreamEventSequence(
  event: UnknownRecord,
): Pick<AIChatStreamEventContext, "sequence"> {
  return parseOptionalNumberProperty(event, "sequence");
}

function parseOptionalWorkoutDraft(
  rawDraft: unknown,
): AIWorkoutDraft | undefined {
  if (rawDraft === undefined) {
    return undefined;
  }

  const draft = parseRecord(rawDraft, "workout_draft");
  const notes = parseOptionalStringProperty(draft, "notes");
  const workoutFocus = parseOptionalStringProperty(draft, "workoutFocus");

  return {
    date: parseRequiredString(draft, "date"),
    ...notes,
    ...workoutFocus,
    exercises: parseWorkoutExercises(draft.exercises),
  };
}

function parseWorkoutExercises(
  rawExercises: unknown,
): AIWorkoutExerciseInput[] {
  if (!Array.isArray(rawExercises)) {
    throw new Error("AI chat workout draft exercises must be an array");
  }

  return rawExercises.map((rawExercise, exerciseIndex) => {
    const exercise = parseRecord(
      rawExercise,
      `workout_draft.exercises[${exerciseIndex}]`,
    );
    return {
      name: parseRequiredString(exercise, "name"),
      sets: parseWorkoutSets(exercise.sets, exerciseIndex),
    };
  });
}

function parseWorkoutSets(
  rawSets: unknown,
  exerciseIndex: number,
): AIWorkoutSetInput[] {
  if (!Array.isArray(rawSets)) {
    throw new Error(
      `AI chat workout draft exercise ${exerciseIndex + 1} sets must be an array`,
    );
  }

  return rawSets.map((rawSet, setIndex) => {
    const set = parseRecord(
      rawSet,
      `workout_draft.exercises[${exerciseIndex}].sets[${setIndex}]`,
    );
    const setType = parseRequiredString(set, "setType");
    if (setType !== "warmup" && setType !== "working") {
      throw new Error("AI chat workout draft setType is invalid");
    }

    return {
      ...parseOptionalNumberProperty(set, "weight"),
      reps: parseRequiredNumber(set, "reps"),
      setType,
    };
  });
}

function parseRecord(value: unknown, label: string): UnknownRecord {
  if (typeof value !== "object" || value === null || Array.isArray(value)) {
    throw new Error(`${label} must be an object`);
  }

  // SAFETY: The runtime object/null/array checks above establish the record shape
  // needed for concrete field parsing in this boundary adapter.
  return value as UnknownRecord;
}

function parseRequiredString(record: UnknownRecord, key: string): string {
  const value = record[key];
  if (typeof value !== "string") {
    throw new Error(`AI chat stream event ${key} must be a string`);
  }

  return value;
}

function parseRequiredNumber(record: UnknownRecord, key: string): number {
  const value = record[key];
  if (typeof value !== "number" || !Number.isFinite(value)) {
    throw new Error(`AI chat stream event ${key} must be a number`);
  }

  return value;
}

function parseOptionalStringProperty<K extends string>(
  record: UnknownRecord,
  key: K,
): Partial<Record<K, string>> {
  const value = parseOptionalStringValue(record, key);
  if (value === undefined) {
    return {};
  }

  // SAFETY: The computed key is the exact key argument, and the value was checked
  // as a string before constructing the single-property record.
  return { [key]: value } as Record<K, string>;
}

function parseOptionalStringValue(
  record: UnknownRecord,
  key: string,
): string | undefined {
  const value = record[key];
  if (value === undefined) {
    return undefined;
  }
  if (typeof value !== "string") {
    throw new Error(`AI chat stream event ${key} must be a string`);
  }

  return value;
}

function parseOptionalNumberProperty<K extends string>(
  record: UnknownRecord,
  key: K,
): Partial<Record<K, number>> {
  const value = record[key];
  if (value === undefined) {
    return {};
  }
  if (typeof value !== "number" || !Number.isFinite(value)) {
    throw new Error(`AI chat stream event ${key} must be a number`);
  }

  // SAFETY: The computed key is the exact key argument, and the value was checked
  // as a finite number before constructing the single-property record.
  return { [key]: value } as Record<K, number>;
}

function casesHandled(value: never): never {
  return value;
}

function throwIfAborted(signal?: AbortSignal): void {
  if (signal?.aborted) {
    throw signal.reason instanceof Error
      ? signal.reason
      : new DOMException("Aborted", "AbortError");
  }
}

function delay(ms: number, signal?: AbortSignal): Promise<void> {
  return new Promise((resolve, reject) => {
    const timer = window.setTimeout(() => {
      cleanup();
      resolve();
    }, ms);

    const onAbort = () => {
      cleanup();
      reject(
        signal?.reason instanceof Error
          ? signal.reason
          : new DOMException("Aborted", "AbortError"),
      );
    };

    const cleanup = () => {
      window.clearTimeout(timer);
      signal?.removeEventListener("abort", onAbort);
    };

    signal?.addEventListener("abort", onAbort, { once: true });
  });
}
