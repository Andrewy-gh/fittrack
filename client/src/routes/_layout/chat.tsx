import { createFileRoute } from "@tanstack/react-router";
import { ChatPage } from "@/features/chat/pages/chat-page";

type ChatSearch = {
  conversationId?: string;
  checkout?: "success" | "cancelled";
};

export const Route = createFileRoute("/_layout/chat")({
  validateSearch: (search): ChatSearch => ({
    conversationId: normalizeConversationSearchValue(search.conversationId),
    checkout: normalizeCheckoutSearchValue(search.checkout),
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
      checkout={search.checkout}
    />
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
