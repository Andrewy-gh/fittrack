export type ChatDraftDestination =
  | { type: "new" }
  | { type: "conversation"; conversationId: number };

export type MainChatNavigationPolicy =
  | "restore-latest-intent"
  | "always-new-chat";

export const MAIN_CHAT_NAVIGATION_POLICY: MainChatNavigationPolicy =
  "restore-latest-intent";

type Draft = { text: string; editedAt: number };

export class ChatDraftStore {
  private userId: string | null = null;
  private newChatDraft: Draft | null = null;
  private conversationDrafts = new Map<number, Draft>();
  private latestExplicitDestination: ChatDraftDestination | null = null;
  private editSequence = 0;

  setUser(userId?: string) {
    const nextUserId = userId ?? null;
    if (this.userId === nextUserId) return;
    this.clear();
    this.userId = nextUserId;
  }

  getDraft(destination: ChatDraftDestination): string {
    return this.getDraftRecord(destination)?.text ?? "";
  }

  setDraft(destination: ChatDraftDestination, text: string) {
    const normalized = text.length === 0 ? null : text;
    if (destination.type === "new") {
      this.newChatDraft = normalized
        ? { text: normalized, editedAt: ++this.editSequence }
        : null;
      return;
    }
    if (!normalized) {
      this.conversationDrafts.delete(destination.conversationId);
      return;
    }
    this.conversationDrafts.set(destination.conversationId, {
      text: normalized,
      editedAt: ++this.editSequence,
    });
  }

  startNewChat() {
    this.newChatDraft = null;
    this.latestExplicitDestination = { type: "new" };
  }

  openConversation(conversationId: number) {
    this.latestExplicitDestination = { type: "conversation", conversationId };
  }

  migrateNewChatDraft(conversationId: number) {
    if (this.newChatDraft) {
      this.conversationDrafts.set(conversationId, this.newChatDraft);
      this.newChatDraft = null;
    }
    this.latestExplicitDestination = { type: "conversation", conversationId };
  }

  resolveMainDestination(
    policy: MainChatNavigationPolicy = MAIN_CHAT_NAVIGATION_POLICY,
  ): ChatDraftDestination {
    if (policy === "always-new-chat") return { type: "new" };
    if (this.latestExplicitDestination?.type === "new") return { type: "new" };

    let latest: { conversationId: number; editedAt: number } | null = null;
    for (const [conversationId, draft] of this.conversationDrafts) {
      if (!latest || draft.editedAt > latest.editedAt) {
        latest = { conversationId, editedAt: draft.editedAt };
      }
    }
    return latest
      ? { type: "conversation", conversationId: latest.conversationId }
      : { type: "new" };
  }

  clear() {
    this.newChatDraft = null;
    this.conversationDrafts.clear();
    this.latestExplicitDestination = null;
    this.editSequence = 0;
  }

  private getDraftRecord(destination: ChatDraftDestination): Draft | null {
    return destination.type === "new"
      ? this.newChatDraft
      : (this.conversationDrafts.get(destination.conversationId) ?? null);
  }
}

export const chatDraftStore = new ChatDraftStore();
