import { useCallback } from "react";
import { History } from "lucide-react";
import { useNavigate } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type {
  AIChatMessage,
  AIWorkoutDraft,
} from "@/features/chat/api/ai-chat";
import { saveAIWorkoutDraftToWorkoutForm } from "@/features/chat/utils/ai-workout-draft";
import { workoutDraftStorage } from "@/lib/local-storage";
import { cn } from "@/lib/utils";
import { toast } from "sonner";
import {
  AIChatBillingActions,
  AIChatBillingCard,
} from "../components/ai-chat-billing-card";
import { ChatComposer } from "../components/chat-composer";
import { ChatEmptyState } from "../components/chat-empty-state";
import { ChatHistoryEntry } from "../components/chat-history-entry";
import { ChatMessageActions } from "../components/chat-message-actions";
import { ChatTypingIndicator } from "../components/chat-typing-indicator";
import { ChatWorkoutDraftCard } from "../components/chat-workout-draft-card";
import { useAIChatBillingAccess } from "../hooks/use-ai-chat-billing-access";
import { useAIChatSession } from "../hooks/use-ai-chat-session";
import { useChatHistoryEntry } from "../hooks/use-chat-history-entry";

type ChatCheckoutSearch = "success" | "cancelled";
type ChatBillingSearch = "cancelled" | "portal-return";

type ChatPageProps = {
  userId?: string;
  conversationId: number | null;
  conversationIdSearch?: string;
  checkout?: ChatCheckoutSearch;
  billing?: ChatBillingSearch;
};

const COMPOSER_PLACEHOLDER =
  "Ask about training, recovery, exercise choices, or FitTrack usage...";

const EXAMPLE_PROMPTS = ["Build me a 45-min push day", "Plan leg day"];

export function ChatPage({
  userId,
  conversationId,
  conversationIdSearch,
  checkout,
  billing,
}: ChatPageProps) {
  const navigate = useNavigate({ from: "/chat" });
  const openConversation = useCallback(
    (
      selectedConversationId: number,
      options: {
        replace?: boolean;
      } = {},
    ) => {
      const target = {
        to: "/chat",
        search: { conversationId: String(selectedConversationId) },
      } as const;

      void navigate(
        options.replace === undefined
          ? target
          : { ...target, replace: options.replace },
      );
    },
    [navigate],
  );
  const historyEntry = useChatHistoryEntry({
    userId,
    conversationId,
    onOpenConversation: openConversation,
  });
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
    submitPromptValue,
    saveLatestWorkoutDraft,
  } = useAIChatSession({
    conversationId,
    onConversationCreated: async (createdConversationId) => {
      await navigate({
        to: "/chat",
        search: { conversationId: String(createdConversationId) },
      });
      await historyEntry.refreshConversations();
    },
  });
  const billingAccess = useAIChatBillingAccess({
    userId,
    checkout,
    billing,
    conversationId: conversationIdSearch,
    navigate,
  });

  if (!userId) {
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
  const currentUserId = userId;

  const hasChatAccess = billingAccess.hasChatAccess;
  const isComposerDisabled =
    isSubmitting || billingAccess.isCheckingAccess || !hasChatAccess;
  const showBillingAccessPanel = billingAccess.accessState !== "ready";

  async function handleNewChat() {
    if (!hasChatAccess || billingAccess.isCheckingAccess) {
      return;
    }

    await createNewChat();
  }

  function handleResumeConversation(selectedConversationId: number) {
    openConversation(selectedConversationId);
  }

  async function handleSubmit() {
    if (isComposerDisabled) {
      return;
    }

    await submitPrompt();
    billingAccess.refreshAccess();
  }

  async function handleSelectExample(text: string) {
    if (isComposerDisabled) {
      return;
    }

    setPrompt(text);
    await submitPromptValue(text);
    billingAccess.refreshAccess();
  }

  async function handleRetry() {
    if (isComposerDisabled) {
      return;
    }

    let lastUserPrompt: string | undefined;
    for (let index = messages.length - 1; index >= 0; index -= 1) {
      if (messages[index].role === "user") {
        lastUserPrompt = messages[index].content;
        break;
      }
    }
    if (!lastUserPrompt) {
      return;
    }

    await submitPromptValue(lastUserPrompt);
    billingAccess.refreshAccess();
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

  let lastAssistantMessageId: number | null = null;
  for (let index = messages.length - 1; index >= 0; index -= 1) {
    if (messages[index].role === "assistant") {
      lastAssistantMessageId = messages[index].id;
      break;
    }
  }

  const composer = (
    <ChatComposer
      value={prompt}
      onChange={setPrompt}
      onSubmit={handleSubmit}
      disabled={isComposerDisabled}
      isSubmitting={isSubmitting}
      placeholder={COMPOSER_PLACEHOLDER}
      autoFocus={messages.length === 0}
    />
  );

  const showEmptyState =
    !isLoadingConversation &&
    !loadError &&
    messages.length === 0 &&
    !conversation?.latest_workout_draft;
  const emptyState = (
    <ChatEmptyState
      heading={
        hasChatAccess
          ? "What should we train today?"
          : "Unlock your AI training partner"
      }
      examples={hasChatAccess ? EXAMPLE_PROMPTS : []}
      onSelectExample={handleSelectExample}
      composer={composer}
      disabled={isComposerDisabled}
    />
  );

  return (
    <div className="flex flex-col gap-4 pb-10">
      {showBillingAccessPanel ? (
        <div className="mx-auto w-full max-w-6xl px-4 pt-4">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
            <div className="sm:flex-1">
              <AIChatBillingCard
                status={billingAccess.billingStatus}
                accessState={billingAccess.accessState}
                isLoading={billingAccess.isBillingCardLoading}
                isError={billingAccess.isBillingError}
              />
            </div>
            <div className="flex w-full flex-col gap-2 sm:w-auto sm:flex-row sm:items-start">
              <AIChatBillingActions
                status={billingAccess.billingStatus}
                accessState={billingAccess.accessState}
                isLoading={billingAccess.isBillingCardLoading}
                isError={billingAccess.isBillingError}
                isRefreshingAccess={billingAccess.isRefreshingAccess}
                isCheckoutLoading={billingAccess.isCheckoutLoading}
                isBillingPortalLoading={billingAccess.isBillingPortalLoading}
                onStartCheckout={billingAccess.startCheckout}
                onManageBilling={billingAccess.manageBilling}
                onRefreshAccess={billingAccess.refreshAccess}
              />
              <Button
                type="button"
                variant="outline"
                aria-label="Open chat history"
                onClick={() => historyEntry.setIsMobileOpen(true)}
                className="w-full lg:hidden"
              >
                <History className="size-4" />
                History
              </Button>
            </div>
          </div>
        </div>
      ) : (
        <div className="mx-auto w-full max-w-6xl px-4 pt-4 lg:hidden">
          <div className="flex w-full flex-col gap-2 sm:w-auto sm:flex-row sm:items-start sm:justify-end">
            <Button
              type="button"
              variant="outline"
              aria-label="Open chat history"
              onClick={() => historyEntry.setIsMobileOpen(true)}
              className="w-full"
            >
              <History className="size-4" />
              History
            </Button>
          </div>
        </div>
      )}

      {billingAccess.checkoutNotice ? (
        <div className="mx-auto w-full max-w-6xl px-4">
          <CheckoutNotice checkout={billingAccess.checkoutNotice} />
        </div>
      ) : null}
      {billingAccess.billingNotice ? (
        <div className="mx-auto w-full max-w-6xl px-4">
          <BillingReturnNotice billing={billingAccess.billingNotice} />
        </div>
      ) : null}

      <div
        data-testid="chat-page-layout"
        className={cn(
          "w-full px-4",
          historyEntry.isCollapsed
            ? "mx-auto max-w-3xl"
            : "mx-auto max-w-6xl lg:mx-0 lg:max-w-none lg:pl-76 lg:pr-4",
        )}
      >
        <ChatHistoryEntry
          conversations={historyEntry.conversations}
          activeConversationId={conversationId}
          isLoading={historyEntry.isLoading}
          error={historyEntry.error}
          isCollapsed={historyEntry.isCollapsed}
          isMobileOpen={historyEntry.isMobileOpen}
          onMobileOpenChange={historyEntry.setIsMobileOpen}
          onToggleCollapsed={() =>
            historyEntry.setIsCollapsed((value) => !value)
          }
          onResumeConversation={handleResumeConversation}
          onNewChat={() => void handleNewChat()}
          isNewChatDisabled={billingAccess.isCheckingAccess || !hasChatAccess}
        />

        <div className="min-w-0 lg:max-w-3xl">
          {historyEntry.isPreparingEntry ||
          historyEntry.isAutoOpeningRecentChat ? (
            <div className="text-sm text-muted-foreground">
              Opening latest chat...
            </div>
          ) : historyEntry.error && !conversationId ? (
            <div className="flex flex-col gap-4">
              <div className="rounded-md border border-destructive/30 bg-destructive/5 p-3 text-sm text-destructive">
                {historyEntry.error}
              </div>
              {emptyState}
            </div>
          ) : showEmptyState ? (
            emptyState
          ) : (
            <div
              data-testid="chat-conversation-body"
              className="flex flex-col gap-6 pt-4"
            >
              {loadError ? (
                <div className="rounded-md border border-destructive/30 bg-destructive/5 p-3 text-sm text-destructive">
                  {loadError}
                </div>
              ) : isLoadingConversation && messages.length === 0 ? (
                <div className="text-sm text-muted-foreground">
                  Loading conversation...
                </div>
              ) : (
                <div className="flex flex-col gap-6">
                  {messages.map((message) => (
                    <MessageBubble
                      key={`${message.id}-${message.updated_at}`}
                      message={message}
                      isLastAssistant={message.id === lastAssistantMessageId}
                      onRetry={handleRetry}
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

              <div className="bg-background pt-2 md:sticky md:bottom-4">
                {composer}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function CheckoutNotice({ checkout }: { checkout: ChatCheckoutSearch }) {
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

function BillingReturnNotice({ billing }: { billing: ChatBillingSearch }) {
  const isCancelled = billing === "cancelled";

  return (
    <div
      className={`rounded-md border p-3 text-sm ${
        isCancelled
          ? "border-primary/30 bg-primary/5 text-foreground"
          : "border-border bg-muted/40 text-muted-foreground"
      }`}
    >
      {isCancelled
        ? "Cancellation received. We are refreshing your AI chat billing status."
        : "Returned from billing. We are refreshing your AI chat billing status."}
    </div>
  );
}

function MessageBubble({
  message,
  isLastAssistant,
  onRetry,
  workoutDraft,
  isSavingWorkoutDraft,
  isSavedWorkoutDraft,
  savedWorkoutId,
  onSaveWorkoutDraft,
  onEditWorkoutDraft,
  onOpenSavedWorkout,
}: {
  message: AIChatMessage;
  isLastAssistant?: boolean;
  onRetry?: () => void;
  workoutDraft?: AIWorkoutDraft;
  isSavingWorkoutDraft?: boolean;
  isSavedWorkoutDraft?: boolean;
  savedWorkoutId?: number;
  onSaveWorkoutDraft?: () => void;
  onEditWorkoutDraft?: () => void;
  onOpenSavedWorkout?: () => void;
}) {
  const isUser = message.role === "user";

  if (isUser) {
    return (
      <div
        data-testid={`chat-message-${message.id}`}
        className="flex justify-end"
      >
        <div className="max-w-[80%] whitespace-pre-wrap rounded-2xl bg-muted px-4 py-2.5 text-sm leading-relaxed">
          {message.content}
        </div>
      </div>
    );
  }

  const isStreaming = message.status === "streaming";
  const isFailed = message.status === "failed";
  const showActions =
    !isStreaming && (message.content.trim().length > 0 || isFailed);

  return (
    <div
      data-testid={`chat-message-${message.id}`}
      className="flex flex-col gap-2"
    >
      <div className="whitespace-pre-wrap text-sm leading-relaxed text-foreground">
        {message.content}
        {isStreaming ? (
          <ChatTypingIndicator
            className={message.content ? "ml-1" : undefined}
          />
        ) : null}
      </div>

      {message.error_message ? (
        <div className="text-xs text-destructive">{message.error_message}</div>
      ) : null}

      {showActions ? (
        <ChatMessageActions
          content={message.content}
          canRetry={Boolean(isLastAssistant && onRetry)}
          onRetry={onRetry}
        />
      ) : null}

      {workoutDraft && onSaveWorkoutDraft && onEditWorkoutDraft ? (
        <ChatWorkoutDraftCard
          className={cn("mt-1 w-full")}
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
