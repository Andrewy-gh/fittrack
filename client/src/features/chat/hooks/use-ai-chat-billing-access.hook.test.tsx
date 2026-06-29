import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { BillingStatusResponse } from "@/features/chat/api/billing";
import type { FeatureAccessGrant } from "@/features/chat/api/feature-access";

const mocks = vi.hoisted(() => ({
  mockCreateBillingCheckoutSession: vi.fn(),
  mockCreateBillingCustomerPortalSession: vi.fn(),
  mockCreateBillingSubscriptionCancelPortalSession: vi.fn(),
  mockGetBaseBillingStatus: vi.fn(),
  mockGetBaseFeatureAccess: vi.fn(),
  mockGetCheckoutBillingStatus: vi.fn(),
  mockGetCheckoutFeatureAccess: vi.fn(),
  mockRedirectToBillingCheckout: vi.fn(),
  mockRedirectToBillingPortal: vi.fn(),
  mockShowErrorToast: vi.fn(),
}));

vi.mock("@/features/chat/api/billing", async () => {
  const { queryOptions } = await import("@tanstack/react-query");

  return {
    billingStatusQueryOptions: (userId?: string) =>
      queryOptions({
        queryKey: ["billing", "ai-chatbot", "status", userId],
        queryFn: () => mocks.mockGetBaseBillingStatus(),
      }),
    createBillingCheckoutSession: mocks.mockCreateBillingCheckoutSession,
    createBillingCustomerPortalSession:
      mocks.mockCreateBillingCustomerPortalSession,
    createBillingSubscriptionCancelPortalSession:
      mocks.mockCreateBillingSubscriptionCancelPortalSession,
    getBillingStatus: mocks.mockGetCheckoutBillingStatus,
    redirectToBillingCheckout: mocks.mockRedirectToBillingCheckout,
    redirectToBillingPortal: mocks.mockRedirectToBillingPortal,
  };
});

vi.mock("@/features/chat/api/feature-access", async () => {
  const { queryOptions } = await import("@tanstack/react-query");

  return {
    featureAccessQueryOptions: (userId?: string) =>
      queryOptions({
        queryKey: ["feature-access", userId],
        queryFn: () => mocks.mockGetBaseFeatureAccess(),
      }),
    getFeatureAccess: mocks.mockGetCheckoutFeatureAccess,
    hasAIChatFeatureAccess: (grants?: FeatureAccessGrant[]) =>
      grants?.some((grant) => grant.feature_key === "ai_chatbot") ?? false,
  };
});

vi.mock("@/lib/errors", () => ({
  showErrorToast: mocks.mockShowErrorToast,
}));

const { useAIChatBillingAccess } =
  await import("@/features/chat/hooks/use-ai-chat-billing-access");

const blockedBillingStatus: BillingStatusResponse = {
  feature_key: "ai_chatbot",
  has_access: false,
};

const activeBillingStatus: BillingStatusResponse = {
  feature_key: "ai_chatbot",
  has_access: true,
  subscription: {
    stripe_subscription_id: "sub_active",
    status: "active",
    cancellation_scheduled: false,
  },
};

const aiChatFeatureGrant: FeatureAccessGrant = {
  created_at: "2026-05-26T12:00:00Z",
  feature_key: "ai_chatbot",
  source: "subscription",
  starts_at: "2026-05-26T12:00:00Z",
};

describe("useAIChatBillingAccess checkout polling", () => {
  beforeEach(() => {
    mocks.mockCreateBillingCheckoutSession.mockReset();
    mocks.mockCreateBillingCustomerPortalSession.mockReset();
    mocks.mockCreateBillingSubscriptionCancelPortalSession.mockReset();
    mocks.mockGetBaseBillingStatus.mockReset();
    mocks.mockGetBaseFeatureAccess.mockReset();
    mocks.mockGetCheckoutBillingStatus.mockReset();
    mocks.mockGetCheckoutFeatureAccess.mockReset();
    mocks.mockRedirectToBillingCheckout.mockReset();
    mocks.mockRedirectToBillingPortal.mockReset();
    mocks.mockShowErrorToast.mockReset();
  });

  it("keeps exhausted checkout polling in activating when billing is active but the feature grant is still pending", async () => {
    mocks.mockGetBaseBillingStatus.mockResolvedValue(blockedBillingStatus);
    mocks.mockGetBaseFeatureAccess.mockResolvedValue([]);
    mocks.mockGetCheckoutBillingStatus.mockResolvedValue(activeBillingStatus);
    mocks.mockGetCheckoutFeatureAccess.mockResolvedValue([]);

    const { result } = renderBillingAccessHook();

    await waitFor(() => {
      expect(result.current.accessState).toBe("activating");
      expect(result.current.billingStatus?.has_access).toBe(true);
    });
    expect(result.current.hasChatAccess).toBe(false);
    expect(result.current.isBillingError).toBe(false);
  });

  it("passes the React Query cancellation signal to checkout polling requests", async () => {
    mocks.mockGetBaseBillingStatus.mockResolvedValue(blockedBillingStatus);
    mocks.mockGetBaseFeatureAccess.mockResolvedValue([]);
    mocks.mockGetCheckoutBillingStatus.mockResolvedValue(activeBillingStatus);
    mocks.mockGetCheckoutFeatureAccess.mockResolvedValue([aiChatFeatureGrant]);

    const { result } = renderBillingAccessHook();

    await waitFor(() => {
      expect(result.current.accessState).toBe("ready");
    });
    expectSharedCancellationSignal(
      mocks.mockGetCheckoutFeatureAccess.mock.calls[0]?.[0],
      mocks.mockGetCheckoutBillingStatus.mock.calls[0]?.[0],
    );
  });

  it("keeps exhausted checkout polling in payment confirming when billing has not caught up", async () => {
    mocks.mockGetBaseBillingStatus.mockResolvedValue(blockedBillingStatus);
    mocks.mockGetBaseFeatureAccess.mockResolvedValue([]);
    mocks.mockGetCheckoutBillingStatus.mockResolvedValue(blockedBillingStatus);
    mocks.mockGetCheckoutFeatureAccess.mockResolvedValue([]);

    const { result } = renderBillingAccessHook();

    await waitFor(() => {
      expect(result.current.accessState).toBe("payment-confirming");
      expect(result.current.billingStatus?.has_access).toBe(false);
    });
    expect(result.current.hasChatAccess).toBe(false);
    expect(result.current.isBillingError).toBe(false);
  });

  it("keeps refreshing checkout access automatically while payment is confirming", async () => {
    mocks.mockGetBaseBillingStatus.mockResolvedValue(blockedBillingStatus);
    mocks.mockGetBaseFeatureAccess.mockResolvedValue([]);
    mocks.mockGetCheckoutBillingStatus.mockResolvedValue(blockedBillingStatus);
    mocks.mockGetCheckoutFeatureAccess.mockResolvedValue([]);

    const { result } = renderBillingAccessHook();

    await waitFor(
      () => {
        expect(result.current.accessState).toBe("payment-confirming");
      },
      { interval: 1 },
    );
    const checkoutAttemptsBeforeRefresh =
      mocks.mockGetCheckoutFeatureAccess.mock.calls.length;

    mocks.mockGetCheckoutBillingStatus.mockResolvedValue(activeBillingStatus);
    mocks.mockGetCheckoutFeatureAccess.mockResolvedValue([aiChatFeatureGrant]);

    await waitFor(() => {
      expect(result.current.accessState).toBe("ready");
      expect(
        mocks.mockGetCheckoutFeatureAccess.mock.calls.length,
      ).toBeGreaterThan(checkoutAttemptsBeforeRefresh);
    });
    expect(result.current.hasChatAccess).toBe(true);
  });

  it("keeps refreshing access automatically when premium billing is active but the feature grant is stale", async () => {
    mocks.mockGetBaseBillingStatus.mockResolvedValue(activeBillingStatus);
    mocks.mockGetBaseFeatureAccess.mockResolvedValue([]);
    mocks.mockGetCheckoutBillingStatus.mockResolvedValue(activeBillingStatus);
    mocks.mockGetCheckoutFeatureAccess.mockResolvedValue([]);

    const { result } = renderBillingAccessHook({ checkout: undefined });

    await waitFor(
      () => {
        expect(result.current.accessState).toBe("activating");
      },
      { interval: 1 },
    );
    const accessAttemptsBeforeRefresh =
      mocks.mockGetCheckoutFeatureAccess.mock.calls.length;

    mocks.mockGetCheckoutFeatureAccess.mockResolvedValue([aiChatFeatureGrant]);

    await waitFor(() => {
      expect(result.current.accessState).toBe("ready");
      expect(
        mocks.mockGetCheckoutFeatureAccess.mock.calls.length,
      ).toBeGreaterThan(accessAttemptsBeforeRefresh);
    });
    expect(result.current.hasChatAccess).toBe(true);
  });

  it("restarts checkout polling when payment confirmation is refreshed", async () => {
    mocks.mockGetBaseBillingStatus.mockResolvedValue(blockedBillingStatus);
    mocks.mockGetBaseFeatureAccess.mockResolvedValue([]);
    mocks.mockGetCheckoutBillingStatus.mockResolvedValue(blockedBillingStatus);
    mocks.mockGetCheckoutFeatureAccess.mockResolvedValue([]);

    const { result } = renderBillingAccessHook();

    await waitFor(() => {
      expect(result.current.accessState).toBe("payment-confirming");
    });
    const checkoutAttemptsBeforeRefresh =
      mocks.mockGetCheckoutBillingStatus.mock.calls.length;

    mocks.mockGetCheckoutBillingStatus.mockResolvedValue(activeBillingStatus);
    mocks.mockGetCheckoutFeatureAccess.mockResolvedValue([aiChatFeatureGrant]);

    act(() => {
      result.current.refreshAccess();
    });

    await waitFor(() => {
      expect(result.current.accessState).toBe("ready");
      expect(
        mocks.mockGetCheckoutBillingStatus.mock.calls.length,
      ).toBeGreaterThan(checkoutAttemptsBeforeRefresh);
    });
    expect(result.current.hasChatAccess).toBe(true);
  });

  it("uses refreshed base billing data after checkout reaches ready", async () => {
    const checkoutReadyStatus: BillingStatusResponse = {
      ...activeBillingStatus,
      trial_usage: {
        used: 0,
        limit: 30,
      },
    };
    const refreshedBillingStatus: BillingStatusResponse = {
      ...activeBillingStatus,
      trial_usage: {
        used: 1,
        limit: 30,
      },
    };
    mocks.mockGetBaseBillingStatus
      .mockResolvedValueOnce(blockedBillingStatus)
      .mockResolvedValue(refreshedBillingStatus);
    mocks.mockGetBaseFeatureAccess
      .mockResolvedValueOnce([])
      .mockResolvedValue([aiChatFeatureGrant]);
    mocks.mockGetCheckoutBillingStatus.mockResolvedValue(checkoutReadyStatus);
    mocks.mockGetCheckoutFeatureAccess.mockResolvedValue([aiChatFeatureGrant]);

    const { result } = renderBillingAccessHook();

    await waitFor(() => {
      expect(result.current.accessState).toBe("ready");
      expect(result.current.billingStatus?.trial_usage?.used).toBe(0);
    });

    act(() => {
      result.current.refreshAccess();
    });

    await waitFor(() => {
      expect(result.current.accessState).toBe("ready");
      expect(result.current.billingStatus?.trial_usage?.used).toBe(1);
    });
  });

  it("returns a recoverable error when checkout polling fails unexpectedly instead of falling back to stale blocked data", async () => {
    mocks.mockGetBaseBillingStatus.mockResolvedValue(blockedBillingStatus);
    mocks.mockGetBaseFeatureAccess.mockResolvedValue([]);
    mocks.mockGetCheckoutBillingStatus.mockRejectedValue(
      new Error("billing status unavailable"),
    );
    mocks.mockGetCheckoutFeatureAccess.mockResolvedValue([]);

    const { result } = renderBillingAccessHook();

    await waitFor(
      () => {
        expect(result.current.accessState).toBe("checkout-activation-error");
      },
      { timeout: 5000 },
    );
    expect(result.current.hasChatAccess).toBe(false);
    expect(result.current.billingStatus?.has_access).toBe(false);
    expect(result.current.isBillingError).toBe(true);
  });

  it("keeps checkout activation recoverable when the base access queries fail during polling", async () => {
    mocks.mockGetBaseBillingStatus.mockRejectedValue(
      new Error("base billing unavailable"),
    );
    mocks.mockGetBaseFeatureAccess.mockRejectedValue(
      new Error("base feature access unavailable"),
    );
    mocks.mockGetCheckoutBillingStatus.mockResolvedValue(activeBillingStatus);
    mocks.mockGetCheckoutFeatureAccess.mockResolvedValue([]);

    const { result } = renderBillingAccessHook();

    await waitFor(() => {
      expect(result.current.accessState).toBe("activating");
      expect(result.current.billingStatus?.has_access).toBe(true);
    });
    expect(result.current.hasChatAccess).toBe(false);
    expect(result.current.isBillingError).toBe(false);
  });

  it("keeps normal billing failures distinct outside checkout return", async () => {
    mocks.mockGetBaseBillingStatus.mockRejectedValue(
      new Error("base billing unavailable"),
    );
    mocks.mockGetBaseFeatureAccess.mockResolvedValue([]);
    mocks.mockGetCheckoutBillingStatus.mockResolvedValue(activeBillingStatus);
    mocks.mockGetCheckoutFeatureAccess.mockResolvedValue([]);

    const { result } = renderBillingAccessHook({ checkout: undefined });

    await waitFor(() => {
      expect(result.current.accessState).toBe("billing-error");
    });
    expect(result.current.isBillingError).toBe(true);
  });

  it("polls cancellation return until billing reflects the canceled state", async () => {
    mocks.mockGetBaseBillingStatus.mockResolvedValue(activeBillingStatus);
    mocks.mockGetBaseFeatureAccess.mockResolvedValue([aiChatFeatureGrant]);
    mocks.mockGetCheckoutBillingStatus
      .mockResolvedValueOnce(activeBillingStatus)
      .mockResolvedValue({
        ...activeBillingStatus,
        subscription: {
          ...activeBillingStatus.subscription!,
          cancellation_scheduled: true,
          access_ends_at: "2026-06-10T12:00:00Z",
        },
      });
    mocks.mockGetCheckoutFeatureAccess.mockResolvedValue([aiChatFeatureGrant]);

    const { result } = renderBillingAccessHook({
      checkout: undefined,
      billing: "cancelled",
    });

    await waitFor(() => {
      expect(
        result.current.billingStatus?.subscription?.cancellation_scheduled,
      ).toBe(true);
    });
    expect(mocks.mockGetCheckoutBillingStatus).toHaveBeenCalledTimes(2);
  });

  it("treats scheduled cancellation as reflected on portal return", async () => {
    mocks.mockGetBaseBillingStatus.mockResolvedValue(activeBillingStatus);
    mocks.mockGetBaseFeatureAccess.mockResolvedValue([aiChatFeatureGrant]);
    mocks.mockGetCheckoutBillingStatus.mockResolvedValue({
      ...activeBillingStatus,
      subscription: {
        ...activeBillingStatus.subscription!,
        cancellation_scheduled: true,
        access_ends_at: "2026-07-10T12:00:00Z",
      },
    });
    mocks.mockGetCheckoutFeatureAccess.mockResolvedValue([aiChatFeatureGrant]);

    const { result } = renderBillingAccessHook({
      checkout: undefined,
      billing: "cancelled",
    });

    await waitFor(() => {
      expect(result.current.billingStatus?.subscription?.access_ends_at).toBe(
        "2026-07-10T12:00:00Z",
      );
    });
    expect(mocks.mockGetCheckoutBillingStatus).toHaveBeenCalledTimes(1);
  });

  it("keeps active access ready after a non-cancellation portal return", async () => {
    mocks.mockGetBaseBillingStatus.mockResolvedValue(activeBillingStatus);
    mocks.mockGetBaseFeatureAccess.mockResolvedValue([aiChatFeatureGrant]);
    mocks.mockGetCheckoutBillingStatus.mockResolvedValue(activeBillingStatus);
    mocks.mockGetCheckoutFeatureAccess.mockResolvedValue([aiChatFeatureGrant]);

    const { result } = renderBillingAccessHook({
      checkout: undefined,
      billing: "portal-return",
    });

    await waitFor(() => {
      expect(result.current.accessState).toBe("ready");
    });
    expect(result.current.hasChatAccess).toBe(true);
    expect(mocks.mockGetCheckoutBillingStatus).toHaveBeenCalled();
    expect(mocks.mockGetBaseBillingStatus).toHaveBeenCalled();
    expect(mocks.mockGetBaseFeatureAccess).toHaveBeenCalled();
  });

  it("passes the React Query cancellation signal to billing cancellation polling requests", async () => {
    mocks.mockGetBaseBillingStatus.mockResolvedValue(activeBillingStatus);
    mocks.mockGetBaseFeatureAccess.mockResolvedValue([aiChatFeatureGrant]);
    mocks.mockGetCheckoutBillingStatus.mockResolvedValue({
      ...activeBillingStatus,
      subscription: {
        stripe_subscription_id: "sub_canceled",
        status: "canceled",
        cancellation_scheduled: false,
      },
    });
    mocks.mockGetCheckoutFeatureAccess.mockResolvedValue([aiChatFeatureGrant]);

    const { result } = renderBillingAccessHook({
      checkout: undefined,
      billing: "cancelled",
    });

    await waitFor(() => {
      expect(result.current.billingStatus?.subscription?.status).toBe(
        "canceled",
      );
    });
    expectSharedCancellationSignal(
      mocks.mockGetCheckoutFeatureAccess.mock.calls[0]?.[0],
      mocks.mockGetCheckoutBillingStatus.mock.calls[0]?.[0],
    );
  });
});

function expectSharedCancellationSignal(
  featureAccessOptions: unknown,
  billingStatusOptions: unknown,
) {
  expect(featureAccessOptions).toEqual({
    signal: expect.any(AbortSignal),
  });
  expect(billingStatusOptions).toEqual({
    signal: expect.any(AbortSignal),
  });
  expect((featureAccessOptions as { signal?: AbortSignal }).signal).toBe(
    (billingStatusOptions as { signal?: AbortSignal }).signal,
  );
}

function renderBillingAccessHook(
  options: {
    checkout?: "success";
    billing?: "cancelled" | "portal-return";
  } = { checkout: "success" },
) {
  const navigate = vi.fn();
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });

  return renderHook(
    () =>
      useAIChatBillingAccess({
        userId: "user-123",
        checkout: options.checkout,
        billing: options.billing,
        conversationId: "41",
        navigate,
      }),
    {
      wrapper: ({ children }: { children: ReactNode }) => (
        <QueryClientProvider client={queryClient}>
          {children}
        </QueryClientProvider>
      ),
    },
  );
}
