import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { BillingStatusResponse } from "@/lib/api/billing";
import type { FeatureAccessGrant } from "@/lib/api/feature-access";

const mocks = vi.hoisted(() => ({
  mockCreateBillingCheckoutSession: vi.fn(),
  mockCreateBillingCustomerPortalSession: vi.fn(),
  mockGetBillingStatus: vi.fn(),
  mockGetFeatureAccess: vi.fn(),
  mockRedirectToBillingCheckout: vi.fn(),
  mockRedirectToBillingPortal: vi.fn(),
  mockShowErrorToast: vi.fn(),
}));

vi.mock("@/lib/api/billing", async () => {
  const { queryOptions } = await import("@tanstack/react-query");

  return {
    billingStatusQueryOptions: (userId?: string) =>
      queryOptions({
        queryKey: ["billing", "ai-chatbot", "status", userId],
        queryFn: () => mocks.mockGetBillingStatus(),
      }),
    createBillingCheckoutSession: mocks.mockCreateBillingCheckoutSession,
    createBillingCustomerPortalSession:
      mocks.mockCreateBillingCustomerPortalSession,
    getBillingStatus: mocks.mockGetBillingStatus,
    redirectToBillingCheckout: mocks.mockRedirectToBillingCheckout,
    redirectToBillingPortal: mocks.mockRedirectToBillingPortal,
  };
});

vi.mock("@/lib/api/feature-access", async () => {
  const { queryOptions } = await import("@tanstack/react-query");

  return {
    featureAccessQueryOptions: (userId?: string) =>
      queryOptions({
        queryKey: ["feature-access", userId],
        queryFn: () => mocks.mockGetFeatureAccess(),
      }),
    getFeatureAccess: mocks.mockGetFeatureAccess,
    hasAIChatFeatureAccess: (grants?: FeatureAccessGrant[]) =>
      grants?.some((grant) => grant.feature_key === "ai_chatbot") ?? false,
  };
});

vi.mock("@/lib/errors", () => ({
  showErrorToast: mocks.mockShowErrorToast,
}));

const { useAIChatBillingAccess } = await import("./-chat-billing-access");

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
    cancel_at_period_end: false,
  },
};

describe("useAIChatBillingAccess checkout polling", () => {
  beforeEach(() => {
    mocks.mockCreateBillingCheckoutSession.mockReset();
    mocks.mockCreateBillingCustomerPortalSession.mockReset();
    mocks.mockGetBillingStatus.mockReset();
    mocks.mockGetFeatureAccess.mockReset();
    mocks.mockRedirectToBillingCheckout.mockReset();
    mocks.mockRedirectToBillingPortal.mockReset();
    mocks.mockShowErrorToast.mockReset();
  });

  it("keeps exhausted checkout polling in activating when billing is active but the feature grant is still pending", async () => {
    mocks.mockGetBillingStatus
      .mockResolvedValueOnce(blockedBillingStatus)
      .mockResolvedValue(activeBillingStatus);
    mocks.mockGetFeatureAccess.mockResolvedValue([]);

    const { result } = renderBillingAccessHook();

    await waitFor(() => {
      expect(result.current.accessState).toBe("activating");
      expect(result.current.billingStatus?.has_access).toBe(true);
    });
    expect(result.current.hasChatAccess).toBe(false);
    expect(result.current.isBillingError).toBe(false);
  });

  it("returns a recoverable error when checkout polling fails unexpectedly instead of falling back to stale blocked data", async () => {
    mocks.mockGetBillingStatus
      .mockResolvedValueOnce(blockedBillingStatus)
      .mockRejectedValue(new Error("billing status unavailable"));
    mocks.mockGetFeatureAccess.mockResolvedValue([]);

    const { result } = renderBillingAccessHook();

    await waitFor(() => {
      expect(result.current.accessState).toBe("error");
    });
    expect(result.current.hasChatAccess).toBe(false);
    expect(result.current.billingStatus?.has_access).toBe(false);
    expect(result.current.isBillingError).toBe(true);
  });
});

function renderBillingAccessHook() {
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
        checkout: "success",
        conversationId: "41",
        navigate: vi.fn(),
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
