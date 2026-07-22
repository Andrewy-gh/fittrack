import type { ChatDraftStore } from "./chat-draft-store";
import { clearAllResumeCursors } from "./chat-resume";

export function clearDeletedAIChatClientState(
  chatDraftStore: ChatDraftStore,
): void {
  chatDraftStore.clear();
  clearAllResumeCursors();
}
