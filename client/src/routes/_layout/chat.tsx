import {
  useCallback,
  useEffect,
  useRef,
  useState,
  type FormEvent,
} from "react";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Textarea } from "@/components/ui/textarea";
import {
  createAIChatConversation,
  getAIChatConversation,
  pollAIChatConversationUntilSettled,
  reportAIChatTelemetry,
  requestAIChatMessageRecovery,
  streamAIChatMessage,
  type AIChatConversation,
  type AIChatConversationDetail,
  type AIChatMessage,
  type AIChatTelemetryEvent,
} from "@/lib/api/ai-chat";
import {
  classifyLoadOutcome,
  classifyRecoveryOutcome,
  classifyStreamInterruption,
  isAbortError,
  terminalStreamStage,
} from "@/lib/ai-chat-observability";
import { getErrorMessage, showErrorToast } from "@/lib/errors";

type ChatSearch = {
  conversationId?: string;
};

type ConversationRequestResult = {
  detail: AIChatConversationDetail | null;
  aborted: boolean;
  error?: unknown;
};

export const Route = createFileRoute("/_layout/chat")({
  validateSearch: (search): ChatSearch => ({
    conversationId:
      typeof search.conversationId === "string" && search.conversationId.trim()
        ? search.conversationId
        : undefined,
  }),
  component: ChatRouteComponent,
});

export function ChatRouteComponent() {
  const { user } = Route.useRouteContext();
  const search = Route.useSearch();
  const navigate = useNavigate({ from: Route.fullPath });
  const conversationId = parseConversationId(search.conversationId);

  const [conversation, setConversation] = useState<AIChatConversation | null>(
    null,
  );
  const [messages, setMessages] = useState<AIChatMessage[]>([]);
  const [prompt, setPrompt] = useState("");
  const [isLoadingConversation, setIsLoadingConversation] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [loadError, setLoadError] = useState<string | null>(null);
  const pendingAssistantIdRef = useRef<number | null>(null);
  const loadAbortRef = useRef<AbortController | null>(null);
  const recoveryAbortRef = useRef<AbortController | null>(null);
  const streamAbortRef = useRef<AbortController | null>(null);

  const recordTelemetry = useCallback((event: AIChatTelemetryEvent) => {
    void Promise.resolve(reportAIChatTelemetry(event)).catch((error) => {
      if (import.meta.env.DEV) {
        console.warn("Failed to record AI chat telemetry", error, event);
      }
    });
  }, []);

  const loadConversation = useCallback(
    async (
      id: number,
      opts?: { silent?: boolean },
    ): Promise<ConversationRequestResult> => {
      loadAbortRef.current?.abort();
      const controller = new AbortController();
      loadAbortRef.current = controller;

      if (!opts?.silent) {
        setIsLoadingConversation(true);
      }

      try {
        const detail = await getAIChatConversation(id, {
          signal: controller.signal,
        });
        if (controller.signal.aborted) {
          return { detail: null, aborted: true };
        }
        setConversation(detail.conversation);
        setMessages(detail.messages);
        setLoadError(null);
        return { detail, aborted: false, error: undefined };
      } catch (error) {
        if (controller.signal.aborted || isAbortError(error)) {
          return { detail: null, aborted: true, error };
        }
        const message = getErrorMessage(error);
        setLoadError(message);
        return { detail: null, aborted: false, error };
      } finally {
        if (loadAbortRef.current === controller) {
          loadAbortRef.current = null;
          if (!opts?.silent) {
            setIsLoadingConversation(false);
          }
        }
      }
    },
    [],
  );

  const recoverConversation = useCallback(
    async (
      id: number,
      opts?: { silent?: boolean },
    ): Promise<ConversationRequestResult> => {
      recoveryAbortRef.current?.abort();
      const controller = new AbortController();
      recoveryAbortRef.current = controller;
      let recoveryRequestError: unknown = null;
      let shouldRetryRecovery = false;

      const requestRecovery = async () => {
        try {
          const response = await requestAIChatMessageRecovery(id, {
            signal: controller.signal,
          });
          shouldRetryRecovery = response.status === "not_needed";
        } catch (error) {
          if (controller.signal.aborted || isAbortError(error)) {
            throw error;
          }
          recoveryRequestError = recoveryRequestError ?? error;
          shouldRetryRecovery = false;
        }
      };

      const retryRecoveryIfNeeded = async () => {
        if (!shouldRetryRecovery) {
          return;
        }
        await requestRecovery();
      };

      try {
        await requestRecovery();

        const detail = await pollAIChatConversationUntilSettled(id, {
          signal: controller.signal,
          onStreaming: retryRecoveryIfNeeded,
        });
        if (controller.signal.aborted) {
          return { detail: null, aborted: true };
        }
        setConversation(detail.conversation);
        setMessages(detail.messages);
        setLoadError(null);
        return { detail, aborted: false, error: undefined };
      } catch (error) {
        if (controller.signal.aborted || isAbortError(error)) {
          return { detail: null, aborted: true, error };
        }
        if (!controller.signal.aborted) {
          const message = getErrorMessage(recoveryRequestError ?? error);
          if (!opts?.silent) {
            setLoadError(message);
          }
        }
        return {
          detail: null,
          aborted: false,
          error: recoveryRequestError ?? error,
        };
      } finally {
        if (recoveryAbortRef.current === controller) {
          recoveryAbortRef.current = null;
        }
      }
    },
    [],
  );

  useEffect(() => {
    if (!conversationId) {
      loadAbortRef.current?.abort();
      recoveryAbortRef.current?.abort();
      streamAbortRef.current?.abort();
      setConversation(null);
      setMessages([]);
      setLoadError(null);
      setIsLoadingConversation(false);
      return;
    }

    void (async () => {
      const loadResult = await loadConversation(conversationId);
      if (loadResult.detail) {
        recordTelemetry({
          category: "load",
          outcome: "load_completed",
        });
      } else {
        recordTelemetry({
          category: "load",
          outcome: classifyLoadOutcome(loadResult.aborted),
        });
      }

      if (
        loadResult.detail?.messages.some(
          (message) => message.status === "streaming",
        )
      ) {
        const recoveryResult = await recoverConversation(conversationId, {
          silent: true,
        });
        const recoveryOutcome = classifyRecoveryOutcome({
          messages: recoveryResult.detail?.messages,
          aborted: recoveryResult.aborted,
          error: recoveryResult.error,
        });
        recordTelemetry({
          category: "recovery",
          outcome: recoveryOutcome,
        });

        if (
          recoveryOutcome !== "recovered_completed" &&
          recoveryOutcome !== "recovery_aborted"
        ) {
          const recoveryError =
            recoveryResult.error ??
            new Error("Failed to recover AI chat conversation");
          recordTelemetry({
            category: "ux",
            outcome: "failure_toast_shown",
          });
          if (!recoveryResult.detail) {
            setLoadError(
              getErrorMessage(
                recoveryError,
                "Failed to recover AI chat conversation",
              ),
            );
          }
          showErrorToast(
            recoveryError,
            "Failed to recover AI chat conversation",
          );
        }
      }
    })();

    return () => {
      loadAbortRef.current?.abort();
      recoveryAbortRef.current?.abort();
      streamAbortRef.current?.abort();
    };
  }, [conversationId, loadConversation, recordTelemetry, recoverConversation]);

  if (!user) {
    return (
      <div className="mx-auto max-w-4xl p-6">
        <Card>
          <CardHeader>
            <CardTitle>AI Chat</CardTitle>
          </CardHeader>
          <CardContent className="text-sm text-muted-foreground">
            Sign in to create a conversation and stream assistant responses.
          </CardContent>
        </Card>
      </div>
    );
  }

  async function handleNewChat() {
    try {
      const created = await createAIChatConversation();
      streamAbortRef.current?.abort();
      recoveryAbortRef.current?.abort();
      loadAbortRef.current?.abort();
      setConversation(created);
      setMessages([]);
      setLoadError(null);
      await navigate({
        to: "/chat",
        search: { conversationId: String(created.id) },
      });
    } catch (error) {
      showErrorToast(error, "Failed to create chat conversation");
    }
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const nextPrompt = prompt.trim();
    if (!nextPrompt || isSubmitting) {
      return;
    }

    setIsSubmitting(true);
    setPrompt("");

    let activeConversationId = conversationId;

    try {
      if (!activeConversationId) {
        const createdConversation = await createAIChatConversation();
        activeConversationId = createdConversation.id;
        setConversation(createdConversation);
        await navigate({
          to: "/chat",
          search: { conversationId: String(activeConversationId) },
        });
      }
    } catch (error) {
      setPrompt(nextPrompt);
      setIsSubmitting(false);
      showErrorToast(error, "Failed to create chat conversation");
      return;
    }

    if (!activeConversationId) {
      setPrompt(nextPrompt);
      setIsSubmitting(false);
      return;
    }

    const baseTimestamp = new Date().toISOString();
    const tempUserId = -Date.now();
    const tempAssistantId = tempUserId - 1;
    let streamStarted = false;
    let shouldRefreshConversation = false;
    const streamController = new AbortController();
    pendingAssistantIdRef.current = tempAssistantId;
    streamAbortRef.current = streamController;

    setMessages((current) => [
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
            streamStarted = true;
            pendingAssistantIdRef.current = event.message_id ?? tempAssistantId;
            setMessages((current) =>
              current.map((message) =>
                message.id === tempAssistantId
                  ? {
                      ...message,
                      id: event.message_id ?? message.id,
                    }
                  : message,
              ),
            );
          },
          onDelta: (event) => {
            const targetId = pendingAssistantIdRef.current ?? tempAssistantId;
            setMessages((current) =>
              current.map((message) =>
                message.id === targetId
                  ? {
                      ...message,
                      content: `${message.content}${event.delta ?? ""}`,
                    }
                  : message,
              ),
            );
          },
          onDone: (event) => {
            const targetId = pendingAssistantIdRef.current ?? tempAssistantId;
            setMessages((current) =>
              current.map((message) =>
                message.id === targetId
                  ? {
                      ...message,
                      id: event.message_id ?? message.id,
                      status: "completed",
                      content: event.text ?? message.content,
                      completed_at: new Date().toISOString(),
                    }
                  : message,
              ),
            );
          },
          onErrorEvent: (event) => {
            const targetId = pendingAssistantIdRef.current ?? tempAssistantId;
            setMessages((current) =>
              current.map((message) =>
                message.id === targetId
                  ? {
                      ...message,
                      id: event.message_id ?? message.id,
                      status: "failed",
                      error_message: event.message,
                    }
                  : message,
              ),
            );
            showErrorToast(
              { message: event.message ?? "AI chat streaming failed" },
              "AI chat streaming failed",
            );
          },
          signal: streamController.signal,
        },
      );

      recordTelemetry({
        category: "stream",
        outcome: streamResult.endedWithError ? "server_error" : "completed",
        stage: terminalStreamStage(),
      });
      if (streamResult.endedWithError) {
        recordTelemetry({
          category: "ux",
          outcome: "failure_toast_shown",
        });
      }
      shouldRefreshConversation = true;
    } catch (error) {
      if (!streamStarted && isPreflightAPIError(error)) {
        setMessages((current) =>
          current.filter(
            (message) =>
              message.id !== tempUserId && message.id !== tempAssistantId,
          ),
        );
        setPrompt(nextPrompt);
        recordTelemetry({
          category: "stream",
          outcome: "server_error",
          stage: "pre_start",
        });
        recordTelemetry({
          category: "ux",
          outcome: "failure_toast_shown",
        });
        showErrorToast(error, "Failed to stream AI chat response");
        return;
      }

      const streamTelemetry = classifyStreamInterruption(error, streamStarted);
      recordTelemetry({
        category: "stream",
        outcome: streamTelemetry.outcome,
        stage: streamTelemetry.stage,
      });

      if (streamTelemetry.outcome === "client_aborted") {
        return;
      }

      const {
        detail: recoveredDetail,
        aborted: recoveryAborted,
        error: recoveryError,
      } = await recoverConversation(activeConversationId, {
        silent: true,
      });
      const recoveryOutcome = classifyRecoveryOutcome({
        messages: recoveredDetail?.messages,
        prompt: nextPrompt,
        aborted: recoveryAborted,
        error: recoveryError,
      });
      recordTelemetry({
        category: "recovery",
        outcome: recoveryOutcome,
      });
      if (recoveryAborted) {
        return;
      }

      const recoveredPromptStatus = findRecoveredPromptStatus(
        recoveredDetail?.messages ?? [],
        nextPrompt,
      );

      if (!recoveredDetail && !streamStarted) {
        setMessages((current) =>
          current.filter(
            (message) =>
              message.id !== tempUserId && message.id !== tempAssistantId,
          ),
        );
      } else if (!recoveredDetail) {
        const targetId = pendingAssistantIdRef.current ?? tempAssistantId;
        setMessages((current) =>
          current.map((message) =>
            message.id === targetId
              ? {
                  ...message,
                  status: "failed",
                  error_message: getErrorMessage(error),
                }
              : message,
          ),
        );
      }

      if (recoveredPromptStatus !== "completed") {
        recordTelemetry({
          category: "ux",
          outcome: "failure_toast_shown",
        });
        showErrorToast(error, "Failed to stream AI chat response");
      } else {
        recordTelemetry({
          category: "ux",
          outcome: "failure_toast_suppressed_due_to_successful_recovery",
        });
      }
    } finally {
      if (streamAbortRef.current === streamController) {
        streamAbortRef.current = null;
      }
      pendingAssistantIdRef.current = null;
      setIsSubmitting(false);
    }

    if (shouldRefreshConversation) {
      await loadConversation(activeConversationId, { silent: true });
    }
  }

  return (
    <div className="mx-auto flex max-w-5xl flex-col gap-6 p-6">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between gap-4">
          <div>
            <CardTitle>AI Chat</CardTitle>
            <p className="text-sm text-muted-foreground">
              Minimal phase-1 slice. Persisted conversations, fetch-based SSE,
              app-owned records.
            </p>
          </div>
          <Button type="button" variant="outline" onClick={handleNewChat}>
            New Chat
          </Button>
        </CardHeader>
      </Card>

      <Card className="min-h-[32rem]">
        <CardContent className="flex h-full flex-col gap-4 p-4">
          {isLoadingConversation ? (
            <div className="text-sm text-muted-foreground">
              Loading conversation...
            </div>
          ) : loadError ? (
            <div className="rounded-md border border-destructive/30 bg-destructive/5 p-3 text-sm text-destructive">
              {loadError}
            </div>
          ) : messages.length === 0 ? (
            <div className="rounded-md border border-dashed p-6 text-sm text-muted-foreground">
              No messages yet. Start a new chat or send the first prompt.
            </div>
          ) : (
            <div className="flex flex-1 flex-col gap-3 overflow-y-auto">
              {messages.map((message) => (
                <MessageBubble
                  key={`${message.id}-${message.updated_at}`}
                  message={message}
                />
              ))}
            </div>
          )}

          <form className="mt-auto flex flex-col gap-3" onSubmit={handleSubmit}>
            <Textarea
              value={prompt}
              onChange={(event) => setPrompt(event.target.value)}
              placeholder="Ask about training, recovery, exercise choices, or FitTrack usage..."
              rows={4}
              disabled={isSubmitting}
            />
            <div className="flex items-center justify-between gap-3">
              <p className="text-xs text-muted-foreground">
                Conversation ID:{" "}
                {conversation?.id ?? conversationId ?? "not created yet"}
              </p>
              <Button type="submit" disabled={isSubmitting || !prompt.trim()}>
                {isSubmitting ? "Streaming..." : "Send"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}

function MessageBubble({ message }: { message: AIChatMessage }) {
  const isUser = message.role === "user";
  const statusLabel =
    message.status === "failed"
      ? "failed"
      : message.status === "streaming"
        ? "streaming"
        : null;

  return (
    <div
      className={`max-w-[85%] rounded-lg border px-4 py-3 text-sm ${
        isUser
          ? "ml-auto border-primary/30 bg-primary/10"
          : "mr-auto border-border bg-background"
      }`}
    >
      <div className="mb-2 flex items-center justify-between gap-3 text-xs uppercase tracking-wide text-muted-foreground">
        <span>{isUser ? "You" : "Assistant"}</span>
        {statusLabel ? <span>{statusLabel}</span> : null}
      </div>
      <div className="whitespace-pre-wrap leading-relaxed">
        {message.content || (message.status === "streaming" ? "..." : "")}
      </div>
      {message.error_message ? (
        <div className="mt-2 text-xs text-destructive">
          {message.error_message}
        </div>
      ) : null}
    </div>
  );
}

function parseConversationId(value?: string): number | null {
  if (!value) {
    return null;
  }

  const parsed = Number(value);
  if (!Number.isInteger(parsed) || parsed <= 0) {
    return null;
  }

  return parsed;
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

function findRecoveredPromptStatus(
  messages: AIChatMessage[],
  prompt: string,
): AIChatMessage["status"] | null {
  const normalizedPrompt = prompt.trim();
  if (!normalizedPrompt) {
    return null;
  }

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

  return null;
}
