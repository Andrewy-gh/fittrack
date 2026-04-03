import type {
  AIChatMessage,
  AIChatTelemetryEvent,
  AIChatTelemetryStage,
} from "@/lib/api/ai-chat";

type RecoveryTelemetryInput = {
  messages?: AIChatMessage[];
  prompt?: string;
  aborted?: boolean;
  error?: unknown;
};

export const recoveryTimeoutMessage =
  "AI chat recovery timed out while waiting for persisted conversation state";

export function classifyStreamInterruption(
  error: unknown,
  streamStarted: boolean,
): Pick<AIChatTelemetryEvent, "outcome" | "stage"> {
  return {
    outcome: isAbortError(error)
      ? "client_aborted"
      : "transport_ended_pre_terminal",
    stage: streamStarted ? "post_start" : "pre_start",
  };
}

export function classifyLoadOutcome(
  aborted: boolean,
): AIChatTelemetryEvent["outcome"] {
  return aborted ? "load_aborted_stale" : "load_failed";
}

export function classifyRecoveryOutcome({
  messages = [],
  prompt,
  aborted,
  error,
}: RecoveryTelemetryInput): AIChatTelemetryEvent["outcome"] {
  if (aborted) {
    return "recovery_aborted";
  }

  if (isRecoveryTimeoutError(error)) {
    return "recovery_timeout";
  }

  const status = findRecoveredAssistantStatus(messages, prompt);
  if (status === "completed") {
    return "recovered_completed";
  }
  if (status === "failed") {
    return "recovered_failed";
  }

  return "recovered_failed";
}

export function isAbortError(error: unknown): boolean {
  return (
    (error instanceof DOMException && error.name === "AbortError") ||
    (error instanceof Error && error.name === "AbortError")
  );
}

export function isRecoveryTimeoutError(error: unknown): boolean {
  return (
    error instanceof Error && error.message.includes(recoveryTimeoutMessage)
  );
}

export function terminalStreamStage(): AIChatTelemetryStage {
  return "terminal";
}

function findRecoveredAssistantStatus(
  messages: AIChatMessage[],
  prompt?: string,
): AIChatMessage["status"] | null {
  const normalizedPrompt = prompt?.trim();
  if (normalizedPrompt) {
    for (let index = messages.length - 1; index >= 0; index -= 1) {
      const message = messages[index];
      if (
        message.role !== "user" ||
        message.content.trim() !== normalizedPrompt
      ) {
        continue;
      }

      const assistant = messages[index + 1];
      if (assistant?.role === "assistant") {
        return assistant.status;
      }

      return null;
    }
  }

  for (let index = messages.length - 1; index >= 0; index -= 1) {
    const message = messages[index];
    if (message.role === "assistant") {
      return message.status;
    }
  }

  return null;
}
