import { useCallback, useEffect, useMemo } from "react";
import { History } from "lucide-react";
import { useNavigate } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  deleteAIChatConversation,
  type AIChatDeleteError,
  type AIChatMessage,
  type AIWorkoutDraft,
} from "@/features/chat/api/ai-chat";
import { saveAIWorkoutDraftToWorkoutForm } from "@/features/chat/utils/ai-workout-draft";
import { showErrorToast } from "@/lib/errors";
import { workoutDraftStorage } from "@/lib/local-storage";
import { cn } from "@/lib/utils";
import { toast } from "sonner";
import {
  AIChatBillingActions,
  AIChatBillingCard,
} from "../components/ai-chat-billing-card";
import { ChatComposer } from "../components/chat-composer";
import { ChatEmptyState } from "../components/chat-empty-state";
import {
  ChatHistoryEntry,
  getChatHistoryListState,
} from "../components/chat-history-entry";
import { ChatMessageActions } from "../components/chat-message-actions";
import { ChatTypingIndicator } from "../components/chat-typing-indicator";
import {
  ChatWorkoutDraftCard,
  type ChatWorkoutDraftSaveState,
} from "../components/chat-workout-draft-card";
import { useAIChatBillingAccess } from "../hooks/use-ai-chat-billing-access";
import { useAIChatSession } from "../hooks/use-ai-chat-session";
import { useChatHistoryEntry } from "../hooks/use-chat-history-entry";
import { useChatDraftStore } from "../utils/chat-draft-context";
import { type ChatDraftDestination } from "../utils/chat-draft-store";

type ChatCheckoutSearch = "success" | "cancelled";
type ChatBillingSearch = "cancelled" | "portal-return";

type ChatPageProps = {
  userId?: string;
  conversationId: number | null;
  conversationIdSearch?: string;
  createChat?: true;
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
  createChat,
  checkout,
  billing,
}: ChatPageProps) {
  const chatDraftStore = useChatDraftStore();
  const navigate = useNavigate({ from: "/chat" });
  const draftDestination = useMemo<ChatDraftDestination>(
    () =>
      conversationId === null
        ? { type: "new" }
        : { type: "conversation", conversationId },
    [conversationId],
  );
  const initialPrompt = useMemo(
    () => chatDraftStore.getDraft(draftDestination),
    [draftDestination],
  );
  const handlePromptChange = useCallback(
    (value: string) => chatDraftStore.setDraft(draftDestination, value),
    [draftDestination],
  );
  const handlePromptStarted = useCallback((startedConversationId: number) => {
    chatDraftStore.setDraft(
      { type: "conversation", conversationId: startedConversationId },
      "",
    );
  }, []);
  const handleNewConversationCreated = useCallback((createdId: number) => {
    chatDraftStore.migrateNewChatDraft(createdId);
  }, []);
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
    shouldOpenLatestConversation: false,
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
    resetConversation,
    submitPrompt,
    submitPromptValue,
    saveLatestWorkoutDraft,
    stopRun,
    canStop,
  } = useAIChatSession({
    conversationId,
    initialPrompt,
    onPromptChange: handlePromptChange,
    onPromptStarted: handlePromptStarted,
    onNewConversationCreated: handleNewConversationCreated,
    onConversationCreated: async (createdConversationId) => {
      const target = {
        to: "/chat",
        search: { conversationId: String(createdConversationId) },
      } as const;
      await navigate(createChat ? { ...target, replace: true } : target);
      await historyEntry.refreshConversations();
    },
  });

  useEffect(() => {
    if (conversationId !== null || createChat || !userId) return;
    const destination = chatDraftStore.resolveMainDestination();
    if (destination.type === "conversation") {
      openConversation(destination.conversationId, { replace: true });
      return;
    }
    void navigate({ to: "/chat", search: { createChat: true }, replace: true });
  }, [conversationId, createChat, navigate, openConversation, userId]);
  const billingAccess = useAIChatBillingAccess({
    userId,
    checkout,
    billing,
    conversationId: conversationIdSearch,
    navigate,
  });
  const historyState = getChatHistoryListState({
    conversations: historyEntry.conversations,
    activeConversationId: conversationId,
    isLoading: historyEntry.isLoading,
    error: historyEntry.error,
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

  function handleNewChat() {
    if (!hasChatAccess || billingAccess.isCheckingAccess) {
      return;
    }

    chatDraftStore.startNewChat();
    resetConversation("");
    void navigate({
      to: "/chat",
      search: { createChat: true },
    });
  }

  function handleResumeConversation(selectedConversationId: number) {
    chatDraftStore.openConversation(selectedConversationId);
    openConversation(selectedConversationId);
  }

  async function handleDeleteConversation(selectedConversationId: number) {
    try {
      await deleteAIChatConversation(selectedConversationId);
    } catch (error: unknown) {
      if (isAIChatDeleteError(error, 404)) {
        await historyEntry.refreshConversations();
        return;
      }
      if (isAIChatDeleteError(error, 409)) {
        toast.error(
          "Stop the response before deleting this chat, then try again.",
        );
        return;
      }

      showErrorToast(error, "Could not delete this chat.");
      return;
    }

    if (selectedConversationId === conversationId) {
      chatDraftStore.setDraft(
        { type: "conversation", conversationId: selectedConversationId },
        "",
      );
      chatDraftStore.startNewChat();
      await navigate({ to: "/chat", search: { createChat: true } });
      return;
    }

    await historyEntry.refreshConversations();
  }

  async function handleSubmit() {
    if (isComposerDisabled) {
      return;
    }

    await submitPrompt();
    void historyEntry.refreshConversations();
    billingAccess.refreshAccess();
  }

  function handleSelectExample(text: string) {
    if (isComposerDisabled) {
      return;
    }

    setPrompt(text);
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
    void historyEntry.refreshConversations();
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
  const workoutDraftSaveState = getWorkoutDraftSaveState({
    isSaving: isSavingWorkoutDraft,
    isSaved: isLatestWorkoutDraftSaved,
    savedWorkoutId,
    onOpenSavedWorkout: handleOpenSavedWorkout,
  });
  const latestWorkoutDraft = conversation?.latest_workout_draft;

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
      onStop={canStop ? () => void stopRun() : undefined}
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

      {/* Symmetric gutter (history rail width + 2rem, see --spacing-chat-gutter)
          keeps the reading column centered in the viewport and clear of the rail. */}
      <div
        data-testid="chat-page-layout"
        className="w-full px-4 lg:px-chat-gutter"
      >
        <ChatHistoryEntry
          historyState={historyState}
          isCollapsed={historyEntry.isCollapsed}
          isMobileOpen={historyEntry.isMobileOpen}
          onMobileOpenChange={historyEntry.setIsMobileOpen}
          onToggleCollapsed={() =>
            historyEntry.setIsCollapsed((value) => !value)
          }
          onResumeConversation={handleResumeConversation}
          onDeleteConversation={
            hasChatAccess ? handleDeleteConversation : undefined
          }
          onNewChat={handleNewChat}
          isNewChatDisabled={billingAccess.isCheckingAccess || !hasChatAccess}
        />

        <div
          data-testid="chat-main-pane"
          className="mx-auto w-full min-w-0 max-w-3xl"
        >
          {historyEntry.entryState.status === "openingLatestChat" ? (
            <div className="text-sm text-muted-foreground">
              Opening latest chat...
            </div>
          ) : historyEntry.entryState.status === "historyLoadError" ? (
            <div className="flex flex-col gap-4">
              <div className="rounded-md border border-destructive/30 bg-destructive/5 p-3 text-sm text-destructive">
                {historyEntry.entryState.message}
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
                      key={message.id}
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
                        if (latestWorkoutDraft) {
                          void handleSaveWorkoutDraft();
                        }
                      }}
                      workoutDraftSaveState={workoutDraftSaveState}
                    />
                  ))}
                </div>
              )}

              {latestWorkoutDraft && latestWorkoutDraftMessageId === null ? (
                <ChatWorkoutDraftCard
                  draft={latestWorkoutDraft}
                  saveState={workoutDraftSaveState}
                  onSave={() => void handleSaveWorkoutDraft()}
                  onEdit={() => handleEditInWorkoutForm(latestWorkoutDraft)}
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

function isAIChatDeleteError(
  error: unknown,
  status: number,
): error is AIChatDeleteError {
  return (
    typeof error === "object" &&
    error !== null &&
    "status" in error &&
    error.status === status
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
  workoutDraftSaveState,
  onSaveWorkoutDraft,
  onEditWorkoutDraft,
}: {
  message: AIChatMessage;
  isLastAssistant?: boolean;
  onRetry?: () => void;
  workoutDraft?: AIWorkoutDraft;
  workoutDraftSaveState: ChatWorkoutDraftSaveState;
  onSaveWorkoutDraft?: () => void;
  onEditWorkoutDraft?: () => void;
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
  const isStopped = message.status === "stopped";
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

      {isStopped ? (
        <div className="text-xs text-muted-foreground">Stopped</div>
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
          saveState={workoutDraftSaveState}
          onSave={onSaveWorkoutDraft}
          onEdit={onEditWorkoutDraft}
        />
      ) : null}
    </div>
  );
}

function getWorkoutDraftSaveState({
  isSaving,
  isSaved,
  savedWorkoutId,
  onOpenSavedWorkout,
}: {
  readonly isSaving: boolean;
  readonly isSaved: boolean;
  readonly savedWorkoutId: number | undefined;
  readonly onOpenSavedWorkout: (workoutId: number) => void;
}): ChatWorkoutDraftSaveState {
  if (isSaving) {
    return { status: "saving" };
  }

  if (!isSaved) {
    return { status: "idle" };
  }

  if (savedWorkoutId === undefined) {
    return { status: "saved" };
  }

  return {
    status: "savedWithWorkout",
    workoutId: savedWorkoutId,
    onOpenSavedWorkout: () => onOpenSavedWorkout(savedWorkoutId),
  };
}
