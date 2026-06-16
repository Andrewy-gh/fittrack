import { describe, expect, it } from "vitest";
import {
  resolveAIChatAccessView,
  resolveCheckoutAccessPollingView,
} from "./use-ai-chat-billing-access";

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
          cancellation_scheduled: false,
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

  it("models an active checkout poll result as activating until the feature grant refreshes", () => {
    const pollingView = resolveCheckoutAccessPollingView({
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
    });

    expect(pollingView.status).toBe("activating");
    if (pollingView.status !== "activating") {
      throw new Error("expected activating checkout polling view");
    }

    const view = resolveAIChatAccessView({
      billingStatus: pollingView.result.billingStatus,
      featureAccess: pollingView.result.featureAccess,
    });
    expect(view.hasChatAccess).toBe(false);
    expect(view.state).toBe("activating");
  });

  it("models a checkout poll result without active billing as payment confirming", () => {
    const pollingView = resolveCheckoutAccessPollingView({
      data: {
        billingStatus: {
          feature_key: "ai_chatbot",
          has_access: false,
        },
        featureAccess: [],
      },
      isFetching: false,
    });

    expect(pollingView.status).toBe("payment-confirming");
    if (pollingView.status !== "payment-confirming") {
      throw new Error("expected payment-confirming checkout polling view");
    }

    const view = resolveAIChatAccessView({
      billingStatus: pollingView.result.billingStatus,
      featureAccess: pollingView.result.featureAccess,
      isPaymentConfirming: pollingView.status === "payment-confirming",
    });
    expect(view.hasChatAccess).toBe(false);
    expect(view.state).toBe("payment-confirming");
  });

  it("models unexpected checkout polling failures as checkout activation errors", () => {
    const pollingView = resolveCheckoutAccessPollingView({
      error: new Error("network unavailable"),
      isFetching: false,
    });

    expect(pollingView.status).toBe("failed");
    const view = resolveAIChatAccessView({
      billingStatus: {
        feature_key: "ai_chatbot",
        has_access: false,
      },
      featureAccess: [],
      errorSource:
        pollingView.status === "failed" ? "checkout-activation" : undefined,
    });

    expect(view.hasChatAccess).toBe(false);
    expect(view.state).toBe("checkout-activation-error");
  });

  it("models base billing or feature access failures as billing errors", () => {
    const view = resolveAIChatAccessView({
      billingStatus: {
        feature_key: "ai_chatbot",
        has_access: false,
      },
      featureAccess: [],
      errorSource: "billing",
    });

    expect(view.hasChatAccess).toBe(false);
    expect(view.state).toBe("billing-error");
  });
});
