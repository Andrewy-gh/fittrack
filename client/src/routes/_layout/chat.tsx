import { createFileRoute } from "@tanstack/react-router";
import { ChatPage } from "@/features/chat/pages/chat-page";

type ChatSearch = {
  conversationId?: string;
  createChat?: true;
  checkout?: "success" | "cancelled";
  billing?: "cancelled" | "portal-return";
};

export const Route = createFileRoute("/_layout/chat")({
  validateSearch: (search): ChatSearch => ({
    conversationId: normalizeConversationSearchValue(search.conversationId),
    createChat: normalizeCreateChatSearchValue(search.createChat),
    checkout: normalizeCheckoutSearchValue(search.checkout),
    billing: normalizeBillingSearchValue(search.billing),
  }),
  component: RouteComponent,
});

function RouteComponent() {
  const { user } = Route.useRouteContext();
  const search = Route.useSearch();

  return (
    <ChatPage
      userId={user?.id}
      conversationId={parseConversationId(search.conversationId)}
      conversationIdSearch={search.conversationId}
      createChat={search.createChat}
      checkout={search.checkout}
      billing={search.billing}
    />
  );
}

function normalizeCreateChatSearchValue(value: unknown): true | undefined {
  return value === true || value === "true" ? true : undefined;
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

function normalizeBillingSearchValue(
  value: unknown,
): ChatSearch["billing"] | undefined {
  return value === "cancelled" || value === "portal-return" ? value : undefined;
}
