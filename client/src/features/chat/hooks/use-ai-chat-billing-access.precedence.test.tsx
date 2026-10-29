import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import type { BillingStatusResponse } from "@/features/chat/api/billing";
import type { FeatureAccessGrant } from "@/features/chat/api/feature-access";
import { useAIChatBillingAccess } from "./use-ai-chat-billing-access";

const grant: FeatureAccessGrant = {
  created_at: "2026-05-26T12:00:00Z",
  feature_key: "ai_chatbot",
  source: "subscription",
  starts_at: "2026-05-26T12:00:00Z",
};

function billingStatus(used: number): BillingStatusResponse {
  return {
    feature_key: "ai_chatbot",
    has_access: false,
    trial_usage: { used, limit: 30 },
  };
}

describe("billing access source precedence", () => {
  it.each([
    {
      name: "checkout grants override a conflicting cancellation and base view",
      baseGrants: [],
      checkoutGrants: [grant],
      cancellationGrants: [],
      expectedState: "ready",
      expectedUsage: 1,
      expectedAccess: true,
    },
    {
      name: "ready checkout yields to cancellation once base grants catch up",
      baseGrants: [grant],
      checkoutGrants: [grant],
      cancellationGrants: [],
      expectedState: "blocked",
      expectedUsage: 2,
      expectedAccess: false,
    },
    {
      name: "pending checkout keeps priority even after base grants catch up",
      baseGrants: [grant],
      checkoutGrants: [],
      cancellationGrants: [grant],
      expectedState: "payment-confirming",
      expectedUsage: 1,
      expectedAccess: false,
    },
  ])(
    "$name",
    async ({
      baseGrants,
      checkoutGrants,
      cancellationGrants,
      expectedState,
      expectedUsage,
      expectedAccess,
    }) => {
      // All four real queries are settled independently. No HTTP or provider
      // behavior is simulated; this exercises the hook's conflicting cache views.
      const queryClient = new QueryClient({
        defaultOptions: {
          queries: { staleTime: Infinity, retry: false },
        },
      });
      queryClient.setQueryData(
        ["billing", "ai-chatbot", "status", "user-123"],
        billingStatus(3),
      );
      queryClient.setQueryData(["feature-access", "user-123"], baseGrants);
      queryClient.setQueryData(
        ["billing", "ai-chatbot", "checkout-access", "user-123", "41"],
        { billingStatus: billingStatus(1), featureAccess: checkoutGrants },
      );
      queryClient.setQueryData(
        ["billing", "ai-chatbot", "billing-cancellation", "user-123", "41"],
        { billingStatus: billingStatus(2), featureAccess: cancellationGrants },
      );

      const { result, unmount } = renderHook(
        () =>
          useAIChatBillingAccess({
            userId: "user-123",
            conversationId: "41",
            checkout: "success",
            billing: "cancelled",
            navigate: async () => {},
          }),
        {
          wrapper: ({ children }: { children: ReactNode }) => (
            <QueryClientProvider client={queryClient}>
              {children}
            </QueryClientProvider>
          ),
        },
      );
      try {
        await waitFor(() => {
          expect(result.current.accessState).toBe(expectedState);
          expect(result.current.billingStatus?.trial_usage?.used).toBe(
            expectedUsage,
          );
          expect(result.current.hasChatAccess).toBe(expectedAccess);
          expect(result.current.isBillingError).toBe(false);
        });
      } finally {
        unmount();
        queryClient.clear();
      }
    },
  );
});
