import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it } from "vitest";
import {
  ChatRouteComponent,
  conversationDetail,
  mockBillingQueryResult,
  mockCheckoutAccessQueryResult,
  mockCreateBillingCheckoutSession,
  mockCreateBillingCustomerPortalSession,
  mockFeatureAccessQueryResult,
  mockGetConversation,
  mockNavigate,
  mockRedirectToBillingCheckout,
  mockRefetchBillingStatus,
  mockRefetchFeatureAccess,
  mockSearch,
  mockStreamMessage,
  resetChatRouteMocks,
} from "./-chat-route-test-utils";

describe("ChatRouteComponent", () => {
  beforeEach(resetChatRouteMocks);

  it("shows checkout success and cleans the query state", async () => {
    mockSearch.checkout = "success";
    mockGetConversation.mockResolvedValue(conversationDetail([]));

    render(<ChatRouteComponent />);

    expect(
      await screen.findByText(
        "Checkout complete. We are refreshing your AI chat access.",
      ),
    ).toBeInTheDocument();
    expect(mockNavigate).toHaveBeenCalledWith({
      to: "/chat",
      search: { conversationId: "41" },
      replace: true,
    });
  });

  it("uses checkout access polling results when the webhook grant appears", async () => {
    mockSearch.checkout = "success";
    mockFeatureAccessQueryResult.value = {
      data: [],
      isLoading: false,
      isPending: false,
      refetch: mockRefetchFeatureAccess,
    };
    mockCheckoutAccessQueryResult.value = {
      data: {
        billingStatus: {
          feature_key: "ai_chatbot",
          has_access: true,
          subscription: {
            stripe_subscription_id: "sub_123",
            status: "active",
            cancel_at_period_end: false,
          },
        },
        featureAccess: [{ feature_key: "ai_chatbot" }],
      },
      isFetching: false,
      isError: false,
      isSuccess: true,
    };
    mockGetConversation.mockResolvedValue(conversationDetail([]));

    render(<ChatRouteComponent />);

    expect(
      await screen.findByPlaceholderText(
        "Ask about training, recovery, exercise choices, or FitTrack usage...",
      ),
    ).toBeEnabled();
    expect(screen.getByText("Premium")).toBeInTheDocument();
    expect(mockRefetchFeatureAccess).not.toHaveBeenCalled();
    expect(mockRefetchBillingStatus).not.toHaveBeenCalled();
  });

  it("shows an activation recovery action when checkout polling returns active billing before the feature grant", async () => {
    const user = userEvent.setup();
    mockSearch.checkout = "success";
    mockBillingQueryResult.value = {
      data: {
        feature_key: "ai_chatbot",
        has_access: false,
      },
      isLoading: false,
      isPending: false,
      refetch: mockRefetchBillingStatus,
    };
    mockFeatureAccessQueryResult.value = {
      data: [],
      isLoading: false,
      isPending: false,
      refetch: mockRefetchFeatureAccess,
    };
    mockCheckoutAccessQueryResult.value = {
      data: {
        billingStatus: {
          feature_key: "ai_chatbot",
          has_access: true,
          subscription: {
            stripe_subscription_id: "sub_active",
            status: "active",
            cancel_at_period_end: false,
          },
        },
        featureAccess: [],
      },
      isFetching: false,
      isError: false,
      isSuccess: true,
    };
    mockGetConversation.mockResolvedValue(conversationDetail([]));

    render(<ChatRouteComponent />);

    expect(
      await screen.findByText("AI chat activation is still finishing."),
    ).toBeInTheDocument();
    expect(screen.getByText("Activating")).toBeInTheDocument();
    expect(
      screen.queryByText("Checking your AI chat access..."),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Start 7-day trial" }),
    ).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Refresh access" }));

    expect(mockRefetchBillingStatus).toHaveBeenCalledTimes(1);
    expect(mockRefetchFeatureAccess).toHaveBeenCalledTimes(1);
  });

  it("confirms payment without offering Checkout when the checkout poll has not received billing access yet", async () => {
    const user = userEvent.setup();
    mockSearch.checkout = "success";
    mockBillingQueryResult.value = {
      data: {
        feature_key: "ai_chatbot",
        has_access: false,
      },
      isLoading: false,
      isPending: false,
      refetch: mockRefetchBillingStatus,
    };
    mockFeatureAccessQueryResult.value = {
      data: [],
      isLoading: false,
      isPending: false,
      refetch: mockRefetchFeatureAccess,
    };
    mockCheckoutAccessQueryResult.value = {
      data: {
        billingStatus: {
          feature_key: "ai_chatbot",
          has_access: false,
        },
        featureAccess: [],
      },
      isFetching: false,
      isError: false,
      isSuccess: true,
    };
    mockGetConversation.mockResolvedValue(conversationDetail([]));

    render(<ChatRouteComponent />);

    expect(
      await screen.findByText(
        "Payment complete. We are confirming your AI chat access.",
      ),
    ).toBeInTheDocument();
    expect(screen.getByText("Confirming")).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Start 7-day trial" }),
    ).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Refresh access" }));

    expect(mockRefetchBillingStatus).toHaveBeenCalledTimes(1);
    expect(mockRefetchFeatureAccess).toHaveBeenCalledTimes(1);
    expect(mockCreateBillingCheckoutSession).not.toHaveBeenCalled();
  });

  it("offers refresh instead of Checkout when checkout activation verification fails", async () => {
    const user = userEvent.setup();
    mockSearch.checkout = "success";
    mockBillingQueryResult.value = {
      data: {
        feature_key: "ai_chatbot",
        has_access: false,
      },
      isLoading: false,
      isPending: false,
      refetch: mockRefetchBillingStatus,
    };
    mockFeatureAccessQueryResult.value = {
      data: [],
      isLoading: false,
      isPending: false,
      refetch: mockRefetchFeatureAccess,
    };
    mockCheckoutAccessQueryResult.value = {
      data: undefined,
      error: new Error("billing status unavailable"),
      isFetching: false,
      isError: true,
      isSuccess: false,
    };
    mockGetConversation.mockResolvedValue(conversationDetail([]));

    render(<ChatRouteComponent />);

    expect(
      await screen.findByText(
        "Checkout finished, but we could not refresh AI chat access. Try refreshing access.",
      ),
    ).toBeInTheDocument();
    expect(screen.getByText("Unavailable")).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Start 7-day trial" }),
    ).not.toBeInTheDocument();
    expect(
      screen.getByText("AI chat activation needs another access refresh."),
    ).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Refresh access" }));

    expect(mockRefetchBillingStatus).toHaveBeenCalledTimes(1);
    expect(mockRefetchFeatureAccess).toHaveBeenCalledTimes(1);
    expect(mockCreateBillingCheckoutSession).not.toHaveBeenCalled();
  });

  it("does not offer Checkout when billing status is active but the feature grant has not refreshed", async () => {
    mockBillingQueryResult.value = {
      data: {
        feature_key: "ai_chatbot",
        has_access: true,
        subscription: {
          stripe_subscription_id: "sub_active",
          status: "active",
          cancel_at_period_end: false,
        },
      },
      isLoading: false,
      isPending: false,
      refetch: mockRefetchBillingStatus,
    };
    mockFeatureAccessQueryResult.value = {
      data: [],
      isLoading: false,
      isPending: false,
      refetch: mockRefetchFeatureAccess,
    };
    mockGetConversation.mockResolvedValue(conversationDetail([]));

    render(<ChatRouteComponent />);

    expect(await screen.findByText("Activating")).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Start 7-day trial" }),
    ).not.toBeInTheDocument();
    expect(
      screen.getByPlaceholderText(
        "Ask about training, recovery, exercise choices, or FitTrack usage...",
      ),
    ).toBeDisabled();
    expect(screen.getByText("AI chat activation is still finishing."));
  });

  it("starts Checkout from the no-access trial CTA", async () => {
    const user = userEvent.setup();
    mockBillingQueryResult.value = {
      data: {
        feature_key: "ai_chatbot",
        has_access: false,
      },
      isLoading: false,
      isPending: false,
      refetch: mockRefetchBillingStatus,
    };
    mockFeatureAccessQueryResult.value = {
      data: [],
      isLoading: false,
      isPending: false,
      refetch: mockRefetchFeatureAccess,
    };
    mockGetConversation.mockResolvedValue(conversationDetail([]));

    render(<ChatRouteComponent />);

    await user.click(
      await screen.findByRole("button", { name: "Start 7-day trial" }),
    );

    await waitFor(() => {
      expect(mockCreateBillingCheckoutSession).toHaveBeenCalledTimes(1);
    });
    expect(mockRedirectToBillingCheckout).toHaveBeenCalledWith(
      "https://checkout.stripe.test/session",
    );
    expect(
      screen.getByPlaceholderText(
        "Ask about training, recovery, exercise choices, or FitTrack usage...",
      ),
    ).toBeDisabled();
  });

  it("keeps the trial CTA hidden while feature access is still loading", async () => {
    mockBillingQueryResult.value = {
      data: {
        feature_key: "ai_chatbot",
        has_access: false,
      },
      isLoading: false,
      isPending: false,
      refetch: mockRefetchBillingStatus,
    };
    mockFeatureAccessQueryResult.value = {
      data: undefined,
      isLoading: true,
      isPending: true,
      refetch: mockRefetchFeatureAccess,
    };
    mockGetConversation.mockResolvedValue(conversationDetail([]));

    render(<ChatRouteComponent />);

    expect(
      await screen.findByText("Checking your AI chat access..."),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Start 7-day trial" }),
    ).not.toBeInTheDocument();
    expect(
      screen.getByPlaceholderText(
        "Ask about training, recovery, exercise choices, or FitTrack usage...",
      ),
    ).toBeDisabled();
    expect(mockCreateBillingCheckoutSession).not.toHaveBeenCalled();
  });

  it("keeps chat enabled when feature access is granted without a Stripe subscription", async () => {
    mockBillingQueryResult.value = {
      data: {
        feature_key: "ai_chatbot",
        has_access: false,
      },
      isLoading: false,
      isPending: false,
      refetch: mockRefetchBillingStatus,
    };
    mockFeatureAccessQueryResult.value = {
      data: [
        {
          feature_key: "ai_chatbot",
        },
      ],
      isLoading: false,
      isPending: false,
      refetch: mockRefetchFeatureAccess,
    };
    mockGetConversation.mockResolvedValue(conversationDetail([]));

    render(<ChatRouteComponent />);

    expect(
      await screen.findByPlaceholderText(
        "Ask about training, recovery, exercise choices, or FitTrack usage...",
      ),
    ).toBeEnabled();
    expect(
      screen.queryByRole("button", { name: "Start 7-day trial" }),
    ).not.toBeInTheDocument();
    expect(screen.getByText("Access active")).toBeInTheDocument();
  });

  it("keeps chat enabled when a manual feature grant overrides blocked Stripe billing", async () => {
    const user = userEvent.setup();
    mockBillingQueryResult.value = {
      data: {
        feature_key: "ai_chatbot",
        has_access: false,
        subscription: {
          stripe_subscription_id: "sub_past_due",
          status: "past_due",
          cancel_at_period_end: false,
        },
      },
      isLoading: false,
      isPending: false,
      refetch: mockRefetchBillingStatus,
    };
    mockFeatureAccessQueryResult.value = {
      data: [
        {
          feature_key: "ai_chatbot",
        },
      ],
      isLoading: false,
      isPending: false,
      refetch: mockRefetchFeatureAccess,
    };
    mockGetConversation.mockResolvedValue(conversationDetail([]));

    render(<ChatRouteComponent />);

    expect(
      await screen.findByPlaceholderText(
        "Ask about training, recovery, exercise choices, or FitTrack usage...",
      ),
    ).toBeEnabled();
    expect(screen.getByRole("button", { name: "Send" })).toBeDisabled();
    expect(screen.getByText("Access active")).toBeInTheDocument();
    expect(
      screen.getByText(
        "AI chat access is active for this account. Billing still needs attention.",
      ),
    ).toBeInTheDocument();
    expect(
      screen.queryByText(
        "AI chat is paused until the payment issue is resolved.",
      ),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByText("Start or restore premium access to use AI chat."),
    ).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Update billing" }));

    await waitFor(() => {
      expect(mockCreateBillingCustomerPortalSession).toHaveBeenCalledTimes(1);
    });
    expect(mockCreateBillingCheckoutSession).not.toHaveBeenCalled();
  });

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
