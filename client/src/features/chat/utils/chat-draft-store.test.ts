import { beforeEach, describe, expect, it } from "vitest";
import { ChatDraftStore } from "./chat-draft-store";

describe("ChatDraftStore", () => {
  let store: ChatDraftStore;

  beforeEach(() => {
    store = new ChatDraftStore();
  });

  it("keeps independent drafts and restores the most recently edited conversation", () => {
    store.setDraft({ type: "conversation", conversationId: 1 }, "first");
    store.setDraft({ type: "conversation", conversationId: 2 }, "second");

    expect(store.getDraft({ type: "conversation", conversationId: 1 })).toBe(
      "first",
    );
    expect(store.resolveMainDestination()).toEqual({
      type: "conversation",
      conversationId: 2,
    });
  });

  it("deletes a draft as soon as its text is empty", () => {
    const destination = { type: "conversation" as const, conversationId: 1 };
    store.setDraft(destination, "draft");
    store.setDraft(destination, "");
    expect(store.getDraft(destination)).toBe("");
    expect(store.resolveMainDestination()).toEqual({ type: "new" });
  });

  it("makes explicit New Chat blank and preferred without deleting conversation drafts", () => {
    const conversation = { type: "conversation" as const, conversationId: 1 };
    store.setDraft(conversation, "keep me");
    store.setDraft({ type: "new" }, "discard me");
    store.startNewChat();

    expect(store.getDraft({ type: "new" })).toBe("");
    expect(store.getDraft(conversation)).toBe("keep me");
    expect(store.resolveMainDestination()).toEqual({ type: "new" });
  });

  it("cancels New Chat preference when a drafted conversation is explicitly opened", () => {
    store.setDraft({ type: "conversation", conversationId: 4 }, "restore me");
    store.startNewChat();
    store.openConversation(4);
    expect(store.resolveMainDestination()).toEqual({
      type: "conversation",
      conversationId: 4,
    });
  });

  it("moves a New Chat draft to the created conversation", () => {
    store.setDraft({ type: "new" }, "first prompt");
    store.migrateNewChatDraft(9);
    expect(store.getDraft({ type: "new" })).toBe("");
    expect(store.getDraft({ type: "conversation", conversationId: 9 })).toBe(
      "first prompt",
    );
  });

  it("clears every draft", () => {
    store.setDraft({ type: "conversation", conversationId: 1 }, "secret");
    store.clear();
    expect(store.getDraft({ type: "conversation", conversationId: 1 })).toBe(
      "",
    );
  });

  it("supports the isolated always-new-chat test policy", () => {
    store.setDraft({ type: "conversation", conversationId: 1 }, "draft");
    expect(store.resolveMainDestination("always-new-chat")).toEqual({
      type: "new",
    });
  });
});
