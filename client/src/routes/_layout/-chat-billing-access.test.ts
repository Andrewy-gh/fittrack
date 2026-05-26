import { describe, expect, it } from "vitest";
import { resolveAIChatAccessView } from "./-chat-billing-access";

describe("resolveAIChatAccessView", () => {
  it("uses feature grants as the chat composer access source", () => {
    const view = resolveAIChatAccessView({
      billingStatus: {
        feature_key: "ai_chatbot",
        has_access: false,
      },
      featureAccess: [
        {
          created_at: "2026-05-26T12:00:00Z",
          feature_key: "ai_chatbot",
          source: "manual",
          starts_at: "2026-05-26T12:00:00Z",
        },
      ],
    });

    expect(view.hasChatAccess).toBe(true);
    expect(view.hasBillingDisplayAccess).toBe(true);
  });

  it("keeps billing display active while waiting for the feature grant to refresh", () => {
    const view = resolveAIChatAccessView({
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
    });

    expect(view.hasChatAccess).toBe(false);
    expect(view.hasBillingDisplayAccess).toBe(true);
  });
});
