import { describe, expect, it } from "vitest";
import {
  CheckoutAccessPendingError,
  resolveAIChatAccessView,
  resolveCheckoutAccessResult,
} from "./-chat-billing-access";

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
    expect(view.state).toBe("ready");
  });

  it("keeps active billing in an activating state until the feature grant refreshes", () => {
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
    expect(view.state).toBe("activating");
  });

  it("keeps users blocked when neither billing nor feature grants allow access", () => {
    const view = resolveAIChatAccessView({
      billingStatus: {
        feature_key: "ai_chatbot",
        has_access: false,
      },
      featureAccess: [],
    });

    expect(view.hasChatAccess).toBe(false);
    expect(view.state).toBe("blocked");
  });

  it("preserves active billing from an exhausted checkout poll while the feature grant is still pending", () => {
    const result = resolveCheckoutAccessResult(
      undefined,
      new CheckoutAccessPendingError({
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
      }),
    );

    const view = resolveAIChatAccessView({
      billingStatus: result?.billingStatus,
      featureAccess: result?.featureAccess,
    });

    expect(view.hasChatAccess).toBe(false);
    expect(view.state).toBe("activating");
  });
});
