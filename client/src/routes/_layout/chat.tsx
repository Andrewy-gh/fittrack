import { createFileRoute, useNavigate } from "@tanstack/react-router";
import {
  ChatPage,
  normalizeConversationSearchValue,
  parseConversationId,
} from "@/features/chat/chat-page";

type ChatSearch = {
  conversationId?: string;
};

export const Route = createFileRoute("/_layout/chat")({
  validateSearch: (search): ChatSearch => ({
    conversationId: normalizeConversationSearchValue(search.conversationId),
  }),
  component: ChatRouteComponent,
});

export function ChatRouteComponent() {
  const { user } = Route.useRouteContext();
  const search = Route.useSearch();
  const navigate = useNavigate({ from: Route.fullPath });
  const conversationId = parseConversationId(search.conversationId);

  return (
    <ChatPage
      user={user}
      conversationId={conversationId}
      onConversationCreated={async (createdConversationId) => {
        await navigate({
          to: "/chat",
          search: { conversationId: String(createdConversationId) },
        });
      }}
      onOpenWorkoutForm={() => {
        void navigate({ to: "/workouts/new" });
      }}
      onOpenSavedWorkout={(workoutId) => {
        void navigate({
          to: "/workouts/$workoutId",
          params: { workoutId },
        });
      }}
    />
  );
}
