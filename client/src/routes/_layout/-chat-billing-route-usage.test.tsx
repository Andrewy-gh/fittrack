import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it } from "vitest";
import {
  ChatRouteComponent,
  conversationDetail,
  mockBillingQueryResult,
  mockGetConversation,
  mockRefetchBillingStatus,
  mockStreamMessage,
  resetChatRouteMocks,
} from "./-chat-route-test-utils";

describe("ChatRouteComponent billing usage", () => {
  beforeEach(resetChatRouteMocks);

  it("refreshes billing status after a submitted prompt so trial usage updates", async () => {
    const user = userEvent.setup();
    mockBillingQueryResult.value = {
      data: {
        feature_key: "ai_chatbot",
        has_access: true,
        subscription: {
          stripe_subscription_id: "sub_trial",
          status: "trialing",
          cancel_at_period_end: false,
        },
        trial_usage: {
          used: 1,
          limit: 30,
        },
      },
      isLoading: false,
      isPending: false,
      refetch: mockRefetchBillingStatus,
    };
    mockGetConversation
      .mockResolvedValueOnce(conversationDetail([]))
      .mockResolvedValueOnce(
        conversationDetail([
          {
            id: 71,
            conversation_id: 41,
            role: "user",
            content: "hello",
            status: "completed",
            created_at: "2026-03-26T17:00:01Z",
            updated_at: "2026-03-26T17:00:01Z",
            completed_at: "2026-03-26T17:00:01Z",
          },
          {
            id: 72,
            conversation_id: 41,
            role: "assistant",
            content: "Completed answer",
            status: "completed",
            created_at: "2026-03-26T17:00:01Z",
            updated_at: "2026-03-26T17:00:02Z",
            completed_at: "2026-03-26T17:00:02Z",
          },
        ]),
      );
    mockStreamMessage.mockImplementation(
      async (
        _conversationId: number,
        _prompt: string,
        options?: {
          onStart?: (event: Record<string, unknown>) => void;
          onDone?: (event: Record<string, unknown>) => void;
        },
      ) => {
        options?.onStart?.({
          type: "start",
          message_id: 72,
        });
        options?.onDone?.({
          type: "done",
          message_id: 72,
          text: "Completed answer",
        });

        return {
          doneEvent: {
            type: "done",
            message_id: 72,
            text: "Completed answer",
          },
          endedWithError: false,
        };
      },
    );

    render(<ChatRouteComponent />);

    expect(
      await screen.findByText("1 of 30 trial prompts used"),
    ).toBeInTheDocument();
    await user.type(
      screen.getByPlaceholderText(
        "Ask about training, recovery, exercise choices, or FitTrack usage...",
      ),
      "hello",
    );
    await user.click(screen.getByRole("button", { name: "Send" }));

    await waitFor(() => {
      expect(mockRefetchBillingStatus).toHaveBeenCalledTimes(1);
    });
    expect(await screen.findByText("Completed answer")).toBeInTheDocument();
  });
});
