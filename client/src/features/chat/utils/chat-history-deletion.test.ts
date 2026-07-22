import { beforeEach, describe, expect, it } from "vitest";
import { ChatDraftStore } from "./chat-draft-store";
import { clearDeletedAIChatClientState } from "./chat-history-deletion";

describe("clearDeletedAIChatClientState", () => {
  beforeEach(() => {
    window.sessionStorage.clear();
  });

  it("clears every tab-local chat draft and resume cursor without touching unrelated state", () => {
    const store = new ChatDraftStore();
    store.setDraft({ type: "new" }, "new chat draft");
    store.setDraft(
      { type: "conversation", conversationId: 41 },
      "private draft",
    );
    window.sessionStorage.setItem("fittrack.ai-chat.resume:41", "cursor");
    window.sessionStorage.setItem("fittrack.ai-chat.resume:99", "cursor");
    window.sessionStorage.setItem("unrelated", "keep");

    clearDeletedAIChatClientState(store);

    expect(store.getDraft({ type: "new" })).toBe("");
    expect(store.getDraft({ type: "conversation", conversationId: 41 })).toBe(
      "",
    );
    expect(
      window.sessionStorage.getItem("fittrack.ai-chat.resume:41"),
    ).toBeNull();
    expect(
      window.sessionStorage.getItem("fittrack.ai-chat.resume:99"),
    ).toBeNull();
    expect(window.sessionStorage.getItem("unrelated")).toBe("keep");
  });
});
