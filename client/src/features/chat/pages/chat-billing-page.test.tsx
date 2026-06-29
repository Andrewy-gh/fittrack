import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it } from "vitest";
import {
  ChatRouteComponent,
  conversationDetail,
  mockBillingQueryResult,
  mockBillingCancellationQueryResult,
  mockCheckoutAccessQueryResult,
  mockCreateBillingCheckoutSession,
  mockCreateBillingCustomerPortalSession,
  mockFeatureAccessQueryResult,
  mockGetConversation,
  mockNavigate,
  mockRedirectToBillingCheckout,
  mockRedirectToBillingPortal,
  mockRefetchBillingStatus,
  mockRefetchFeatureAccess,
  mockSearch,
  resetChatRouteMocks,
} from "../test/chat-page-test-utils";

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
            cancellation_scheduled: false,
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
    expect(screen.queryByText("Premium")).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Start 7-day trial" }),
    ).not.toBeInTheDocument();
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
            cancellation_scheduled: false,
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

    expect(await screen.findByText("Activating")).toBeInTheDocument();
    expect(
      screen.queryByText("Checking your AI chat access..."),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Start 7-day trial" }),
    ).not.toBeInTheDocument();

    const billingRefreshesBeforeClick =
      mockRefetchBillingStatus.mock.calls.length;
    const featureAccessRefreshesBeforeClick =
      mockRefetchFeatureAccess.mock.calls.length;

    await user.click(screen.getByRole("button", { name: "Refresh access" }));

    expect(mockRefetchBillingStatus.mock.calls.length).toBeGreaterThanOrEqual(
      billingRefreshesBeforeClick + 1,
    );
    expect(mockRefetchFeatureAccess.mock.calls.length).toBeGreaterThanOrEqual(
      featureAccessRefreshesBeforeClick + 1,
    );
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

    expect(await screen.findByText("Confirming")).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Start 7-day trial" }),
    ).not.toBeInTheDocument();

    const billingRefreshesBeforeClick =
      mockRefetchBillingStatus.mock.calls.length;
    const featureAccessRefreshesBeforeClick =
      mockRefetchFeatureAccess.mock.calls.length;

    await user.click(screen.getByRole("button", { name: "Refresh access" }));

    expect(mockRefetchBillingStatus.mock.calls.length).toBeGreaterThanOrEqual(
      billingRefreshesBeforeClick + 1,
    );
    expect(mockRefetchFeatureAccess.mock.calls.length).toBeGreaterThanOrEqual(
      featureAccessRefreshesBeforeClick + 1,
    );
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
          cancellation_scheduled: false,
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
  });

  it("keeps ordinary billing management out of AI chat for active subscribers", async () => {
    mockBillingQueryResult.value = {
      data: {
        feature_key: "ai_chatbot",
        has_access: true,
        subscription: {
          stripe_subscription_id: "sub_active",
          status: "active",
          cancellation_scheduled: false,
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
    expect(screen.queryByText("Premium")).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Manage plan" }),
    ).not.toBeInTheDocument();

    expect(mockCreateBillingCustomerPortalSession).not.toHaveBeenCalled();
    expect(mockRedirectToBillingPortal).not.toHaveBeenCalled();
    expect(mockCreateBillingCheckoutSession).not.toHaveBeenCalled();
  });

  it("shows a cancellation return notice and refreshes billing state", async () => {
    mockSearch.billing = "cancelled";
    mockGetConversation.mockResolvedValue(conversationDetail([]));

    render(<ChatRouteComponent />);

    expect(
      await screen.findByText(
        "Cancellation received. We are refreshing your AI chat billing status.",
      ),
    ).toBeInTheDocument();
    expect(mockNavigate).toHaveBeenCalledWith({
      to: "/chat",
      search: { conversationId: "41" },
      replace: true,
    });
  });

  it("shows a billing return notice and refreshes billing state", async () => {
    mockSearch.billing = "portal-return";
    mockGetConversation.mockResolvedValue(conversationDetail([]));

    render(<ChatRouteComponent />);

    expect(
      await screen.findByText(
        "Returned from billing. We are refreshing your AI chat billing status.",
      ),
    ).toBeInTheDocument();
    expect(mockRefetchBillingStatus).toHaveBeenCalledTimes(1);
    expect(mockRefetchFeatureAccess).toHaveBeenCalledTimes(1);
    expect(mockNavigate).toHaveBeenCalledWith({
      to: "/chat",
      search: { conversationId: "41" },
      replace: true,
    });
  });

  it("keeps active access ready after returning from billing management", async () => {
    mockSearch.billing = "portal-return";
    mockBillingCancellationQueryResult.value = {
      data: undefined,
      error: null,
      isFetching: true,
      isError: false,
      isSuccess: false,
    };
    mockGetConversation.mockResolvedValue(conversationDetail([]));

    render(<ChatRouteComponent />);

    expect(
      await screen.findByText(
        "Returned from billing. We are refreshing your AI chat billing status.",
      ),
    ).toBeInTheDocument();
    expect(screen.queryByText("Premium")).not.toBeInTheDocument();
    expect(
      screen
        .getAllByRole("button", { name: "New Chat" })
        .some((button) => !button.hasAttribute("disabled")),
    ).toBe(true);
    expect(
      screen.queryByText("Access continues until Jun 10, 2026."),
    ).not.toBeInTheDocument();
  });

  it("uses cancellation polling data after Stripe returns before the webhook has refreshed base billing", async () => {
    mockSearch.billing = "cancelled";
    mockBillingCancellationQueryResult.value = {
      data: {
        billingStatus: {
          feature_key: "ai_chatbot",
          has_access: true,
          subscription: {
            stripe_subscription_id: "sub_active",
            status: "active",
            cancellation_scheduled: true,
            access_ends_at: "2026-06-10T12:00:00Z",
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
      await screen.findByText(
        "Cancellation received. We are refreshing your AI chat billing status.",
      ),
    ).toBeInTheDocument();
    expect(
      screen.queryByText("Access continues until Jun 10, 2026."),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Cancel plan" }),
    ).not.toBeInTheDocument();
  });

  it("uses access-end polling data after Stripe returns from scheduled cancellation", async () => {
    mockSearch.billing = "cancelled";
    mockBillingCancellationQueryResult.value = {
      data: {
        billingStatus: {
          feature_key: "ai_chatbot",
          has_access: true,
          subscription: {
            stripe_subscription_id: "sub_active",
            status: "active",
            cancellation_scheduled: true,
            access_ends_at: "2026-07-10T12:00:00Z",
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
      await screen.findByText(
        "Cancellation received. We are refreshing your AI chat billing status.",
      ),
    ).toBeInTheDocument();
    expect(
      screen.queryByText("Access continues until Jul 10, 2026."),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Cancel plan" }),
    ).not.toBeInTheDocument();
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
    expect(screen.queryByText("Access active")).not.toBeInTheDocument();
  });

  it("keeps chat enabled without billing chrome when a manual feature grant overrides blocked Stripe billing", async () => {
    mockBillingQueryResult.value = {
      data: {
        feature_key: "ai_chatbot",
        has_access: false,
        subscription: {
          stripe_subscription_id: "sub_past_due",
          status: "past_due",
          cancellation_scheduled: false,
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
    expect(screen.queryByText("Access active")).not.toBeInTheDocument();
    expect(
      screen.queryByText("Update billing to keep chat available."),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByText(
        "AI chat is paused until the payment issue is resolved.",
      ),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByText("Start or restore premium access to use AI chat."),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Manage plan" }),
    ).not.toBeInTheDocument();
    expect(mockCreateBillingCustomerPortalSession).not.toHaveBeenCalled();
    expect(mockCreateBillingCheckoutSession).not.toHaveBeenCalled();
  });
});
