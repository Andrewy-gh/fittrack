import { createContext, type ReactNode, useContext, useState } from "react";
import { ChatDraftStore } from "./chat-draft-store";

const ChatDraftContext = createContext<ChatDraftStore | null>(null);

export function ChatDraftProvider({ children }: { children: ReactNode }) {
  const [store] = useState(() => new ChatDraftStore());

  return <ChatDraftContext value={store}>{children}</ChatDraftContext>;
}

export function useChatDraftStore() {
  const store = useContext(ChatDraftContext);
  if (!store) {
    throw new Error(
      "useChatDraftStore must be used within a ChatDraftProvider",
    );
  }
  return store;
}
