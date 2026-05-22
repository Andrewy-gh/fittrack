import { useEffect, useState, type FormEvent } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { AIChatBillingCard } from "@/components/billing/ai-chat-billing-card";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Textarea } from "@/components/ui/textarea";
import type { AIChatMessage, AIWorkoutDraft } from "@/lib/api/ai-chat";
import {
  createBillingCustomerPortalSession,
  billingStatusQueryOptions,
  createBillingCheckoutSession,
  redirectToBillingCheckout,
  redirectToBillingPortal,
} from "@/lib/api/billing";
import {
  featureAccessQueryOptions,
  hasAIChatFeatureAccess,
} from "@/lib/api/feature-access";
import { saveAIWorkoutDraftToWorkoutForm } from "@/lib/ai-workout-draft";
import { showErrorToast } from "@/lib/errors";
import { workoutDraftStorage } from "@/lib/local-storage";
import { toast } from "sonner";
import { useAIChatSession } from "./-chat-session";
import { ChatWorkoutDraftCard } from "./-chat-workout-draft";

type ChatSearch = {
  conversationId?: string;
  checkout?: "success" | "cancelled";
};

const checkoutAccessPollDelaysMs = [0, 1000, 2000, 4000, 8000];

export const Route = createFileRoute("/_layout/chat")({
  validateSearch: (search): ChatSearch => ({
    conversationId: normalizeConversationSearchValue(search.conversationId),
    checkout: normalizeCheckoutSearchValue(search.checkout),
  }),
  component: ChatRouteComponent,
});

export function ChatRouteComponent() {
  const { user } = Route.useRouteContext();
  const search = Route.useSearch();
  const navigate = useNavigate({ from: Route.fullPath });
  const conversationId = parseConversationId(search.conversationId);
  const [checkoutNotice, setCheckoutNotice] = useState<
    ChatSearch["checkout"] | null
  >(search.checkout ?? null);
  const [isPollingCheckoutAccess, setIsPollingCheckoutAccess] = useState(
    search.checkout === "success",
  );
  const {
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
  } = useAIChatSession({
    conversationId,
    onConversationCreated: async (createdConversationId) => {
      await navigate({
        to: "/chat",
        search: { conversationId: String(createdConversationId) },
      });
    },
  });
  const billingQuery = useQuery({
    ...billingStatusQueryOptions(),
    enabled: Boolean(user),
  });
  const featureAccessQuery = useQuery({
    ...featureAccessQueryOptions(),
    enabled: Boolean(user),
  });
  const refetchBillingStatus = billingQuery.refetch;
  const refetchFeatureAccess = featureAccessQuery.refetch;
  const checkoutMutation = useMutation({
    mutationFn: createBillingCheckoutSession,
    onSuccess: (session) => redirectToBillingCheckout(session.url),
    onError: (error) => showErrorToast(error, "Could not open Checkout"),
  });
  const billingPortalMutation = useMutation({
    mutationFn: createBillingCustomerPortalSession,
    onSuccess: (session) => redirectToBillingPortal(session.url),
    onError: (error) => showErrorToast(error, "Could not open billing"),
  });

  useEffect(() => {
    if (!search.checkout) {
      return;
    }

    setCheckoutNotice(search.checkout);
    if (search.checkout === "success") {
      setIsPollingCheckoutAccess(true);
    }

    void navigate({
      to: "/chat",
      search: { conversationId: search.conversationId },
      replace: true,
    });
  }, [navigate, search.checkout, search.conversationId]);

  useEffect(() => {
    if (!isPollingCheckoutAccess) {
      return;
    }

    let isCancelled = false;
    void pollCheckoutAccess();

    return () => {
      isCancelled = true;
    };

    async function pollCheckoutAccess() {
      for (const delayMs of checkoutAccessPollDelaysMs) {
        if (delayMs > 0) {
          await waitForCheckoutAccessRetry(delayMs);
        }
        if (isCancelled) {
          return;
        }

        const [featureAccessResult] = await Promise.all([
          refetchFeatureAccess(),
          refetchBillingStatus(),
        ]);
        if (isCancelled) {
          return;
        }
        if (hasAIChatFeatureAccess(featureAccessResult.data)) {
          setIsPollingCheckoutAccess(false);
          return;
        }
      }

      if (!isCancelled) {
        setIsPollingCheckoutAccess(false);
      }
    }
  }, [isPollingCheckoutAccess, refetchBillingStatus, refetchFeatureAccess]);

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
  const currentUserId = user.id;
  const hasAIChatAccess = hasAIChatFeatureAccess(featureAccessQuery.data);
  const isFeatureAccessLoading =
    featureAccessQuery.isLoading || featureAccessQuery.isPending;
  const isBillingLoading = billingQuery.isLoading || billingQuery.isPending;
  const isChatAccessLoading = isFeatureAccessLoading || isPollingCheckoutAccess;

  async function handleNewChat() {
    if (!hasAIChatAccess || isChatAccessLoading) {
      return;
    }

    await createNewChat();
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!hasAIChatAccess || isChatAccessLoading) {
      return;
    }

    await submitPrompt();
  }

  function handleEditInWorkoutForm(draft: AIWorkoutDraft) {
    const hasDraft = workoutDraftStorage.load(currentUserId) !== null;
    if (
      hasDraft &&
      !window.confirm(
        "Replace your current workout draft with the latest AI workout draft?",
      )
    ) {
      return;
    }

    saveAIWorkoutDraftToWorkoutForm(draft, currentUserId, workoutDraftStorage);
    toast.success("Workout draft loaded into the form");
    void navigate({ to: "/workouts/new" });
  }

  async function handleSaveWorkoutDraft() {
    await saveLatestWorkoutDraft();
  }

  function handleOpenSavedWorkout(workoutId: number) {
    void navigate({
      to: "/workouts/$workoutId",
      params: { workoutId },
    });
  }

  const isLatestWorkoutDraftSaved =
    conversation?.latest_workout_draft_status?.is_saved ?? false;
  const savedWorkoutId =
    conversation?.latest_workout_draft_status?.saved_workout_id;

  return (
    <div className="mx-auto flex max-w-5xl flex-col gap-6 p-6">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between gap-4">
          <div>
            <CardTitle>AI Chat</CardTitle>
            <p className="text-sm text-muted-foreground">
              Persisted chat with reopenable workout drafts and a direct handoff
              into the workout form.
            </p>
          </div>
          <Button
            type="button"
            variant="outline"
            onClick={handleNewChat}
            disabled={isChatAccessLoading || !hasAIChatAccess}
          >
            New Chat
          </Button>
        </CardHeader>
      </Card>

      {checkoutNotice ? <CheckoutNotice checkout={checkoutNotice} /> : null}

      <AIChatBillingCard
        status={billingQuery.data}
        hasFeatureAccess={hasAIChatAccess}
        isLoading={isBillingLoading}
        isError={billingQuery.isError}
        isCheckoutLoading={checkoutMutation.isPending}
        isBillingPortalLoading={billingPortalMutation.isPending}
        onStartCheckout={() => checkoutMutation.mutate()}
        onManageBilling={() => billingPortalMutation.mutate()}
      />

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
                  workoutDraft={
                    message.id === latestWorkoutDraftMessageId
                      ? conversation?.latest_workout_draft
                      : undefined
                  }
                  onEditWorkoutDraft={() => {
                    if (conversation?.latest_workout_draft) {
                      handleEditInWorkoutForm(
                        conversation.latest_workout_draft,
                      );
                    }
                  }}
                  onSaveWorkoutDraft={() => {
                    if (conversation?.latest_workout_draft) {
                      void handleSaveWorkoutDraft();
                    }
                  }}
                  onOpenSavedWorkout={
                    savedWorkoutId
                      ? () => handleOpenSavedWorkout(savedWorkoutId)
                      : undefined
                  }
                  isSavingWorkoutDraft={isSavingWorkoutDraft}
                  isSavedWorkoutDraft={isLatestWorkoutDraftSaved}
                  savedWorkoutId={savedWorkoutId}
                />
              ))}
            </div>
          )}

          {conversation?.latest_workout_draft &&
          latestWorkoutDraftMessageId === null ? (
            <ChatWorkoutDraftCard
              className="max-w-[85%]"
              draft={conversation.latest_workout_draft}
              isSaving={isSavingWorkoutDraft}
              isSaved={isLatestWorkoutDraftSaved}
              savedWorkoutId={savedWorkoutId}
              onSave={() => void handleSaveWorkoutDraft()}
              onEdit={() =>
                handleEditInWorkoutForm(conversation.latest_workout_draft!)
              }
              onOpenSavedWorkout={
                savedWorkoutId
                  ? () => handleOpenSavedWorkout(savedWorkoutId)
                  : undefined
              }
            />
          ) : null}

          <form
            className="mt-auto flex flex-col gap-3"
            onSubmit={handleSubmit}
          >
            <Textarea
              value={prompt}
              onChange={(event) => setPrompt(event.target.value)}
              placeholder="Ask about training, recovery, exercise choices, or FitTrack usage..."
              rows={4}
              disabled={isSubmitting || isChatAccessLoading || !hasAIChatAccess}
            />
            <div className="flex items-center justify-between gap-3">
              <p className="text-xs text-muted-foreground">
                {isChatAccessLoading
                  ? "Checking AI chat access..."
                  : hasAIChatAccess
                    ? `Conversation ID: ${
                        conversation?.id ?? conversationId ?? "not created yet"
                      }`
                    : "Start or restore premium access to use AI chat."}
              </p>
              <Button
                type="submit"
                disabled={
                  isSubmitting ||
                  !prompt.trim() ||
                  isChatAccessLoading ||
                  !hasAIChatAccess
                }
              >
                {isSubmitting ? "Streaming..." : "Send"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}

function CheckoutNotice({
  checkout,
}: {
  checkout: NonNullable<ChatSearch["checkout"]>;
}) {
  const isSuccess = checkout === "success";

  return (
    <div
      className={`rounded-md border p-3 text-sm ${
        isSuccess
          ? "border-primary/30 bg-primary/5 text-foreground"
          : "border-border bg-muted/40 text-muted-foreground"
      }`}
    >
      {isSuccess
        ? "Checkout complete. We are refreshing your AI chat access."
        : "Checkout was cancelled. You can restart the trial when you are ready."}
    </div>
  );
}

function MessageBubble({
  message,
  workoutDraft,
  isSavingWorkoutDraft,
  isSavedWorkoutDraft,
  savedWorkoutId,
  onSaveWorkoutDraft,
  onEditWorkoutDraft,
  onOpenSavedWorkout,
}: {
  message: AIChatMessage;
  workoutDraft?: AIWorkoutDraft;
  isSavingWorkoutDraft?: boolean;
  isSavedWorkoutDraft?: boolean;
  savedWorkoutId?: number;
  onSaveWorkoutDraft?: () => void;
  onEditWorkoutDraft?: () => void;
  onOpenSavedWorkout?: () => void;
}) {
  const isUser = message.role === "user";
  const statusLabel =
    message.status === "failed"
      ? "failed"
      : message.status === "streaming"
        ? "streaming"
        : null;

  return (
    <div
      data-testid={`chat-message-${message.id}`}
      className={`flex max-w-[85%] flex-col gap-3 ${
        isUser ? "ml-auto items-end" : "mr-auto items-start"
      }`}
    >
      <div
        className={`w-full rounded-lg border px-4 py-3 text-sm ${
          isUser
            ? "border-primary/30 bg-primary/10"
            : "border-border bg-background"
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

      {!isUser && workoutDraft && onSaveWorkoutDraft && onEditWorkoutDraft ? (
        <ChatWorkoutDraftCard
          className="w-full"
          draft={workoutDraft}
          isSaving={isSavingWorkoutDraft}
          isSaved={isSavedWorkoutDraft}
          savedWorkoutId={savedWorkoutId}
          onSave={onSaveWorkoutDraft}
          onEdit={onEditWorkoutDraft}
          onOpenSavedWorkout={onOpenSavedWorkout}
        />
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

function normalizeConversationSearchValue(value: unknown): string | undefined {
  if (typeof value === "string") {
    const trimmed = value.trim();
    return trimmed ? trimmed : undefined;
  }

  if (
    typeof value === "number" &&
    Number.isInteger(value) &&
    Number.isFinite(value)
  ) {
    return String(value);
  }

  return undefined;
}

function normalizeCheckoutSearchValue(
  value: unknown,
): ChatSearch["checkout"] | undefined {
  return value === "success" || value === "cancelled" ? value : undefined;
}

function waitForCheckoutAccessRetry(delayMs: number): Promise<void> {
  return new Promise((resolve) => window.setTimeout(resolve, delayMs));
}
