import { beforeEach, describe, expect, it, vi } from "vitest";

const { getUser } = vi.hoisted(() => ({
  getUser: vi.fn(),
}));

const { applyLocalDevAuthHeader } = vi.hoisted(() => ({
  applyLocalDevAuthHeader: vi.fn((headers: Headers) => {
    headers.set("x-fittrack-dev-e2e-user", "local-e2e-user");
    return headers;
  }),
}));

vi.mock("@/stack", () => ({
  stackClientApp: {
    getUser,
  },
}));

vi.mock("@/lib/local-dev-auth", () => ({
  applyLocalDevAuthHeader,
}));

import {
  createBillingCheckoutSession,
  createBillingCustomerPortalSession,
  getBillingStatus,
} from "./billing";

function latestRequest(): Request {
  return vi.mocked(fetch).mock.calls.at(-1)?.[0] as Request;
}

describe("billing api wrapper", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    getUser.mockReset();
    applyLocalDevAuthHeader.mockClear();
    getUser.mockResolvedValue({
      getAuthJson: vi.fn().mockResolvedValue({ accessToken: "token-123" }),
    });
  });

  it("loads current AI chat billing status", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          feature_key: "ai_chatbot",
          has_access: true,
          subscription: {
            stripe_subscription_id: "sub_123",
            status: "trialing",
            cancel_at_period_end: false,
          },
          trial_usage: {
            used: 3,
            limit: 30,
          },
        }),
        {
          status: 200,
          headers: {
            "Content-Type": "application/json",
          },
        },
      ),
    );

    const status = await getBillingStatus();

    expect(fetch).toHaveBeenCalledWith(expect.any(Request));
    expect(latestRequest().url).toContain("/api/billing/status");
    expect(latestRequest().method).toBe("GET");
    expect(status.trial_usage).toEqual({ used: 3, limit: 30 });
  });

  it("creates a Stripe-hosted Checkout session", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          url: "https://checkout.stripe.test/session",
        }),
        {
          status: 200,
          headers: {
            "Content-Type": "application/json",
          },
        },
      ),
    );

    const session = await createBillingCheckoutSession();

    expect(fetch).toHaveBeenCalledWith(expect.any(Request));
    expect(latestRequest().url).toContain("/api/billing/checkout-session");
    expect(latestRequest().method).toBe("POST");
    expect(session.url).toBe("https://checkout.stripe.test/session");
  });

  it("creates a Stripe-hosted billing portal session", async () => {
    vi.spyOn(globalThis, "fetch").mockResolvedValue(
      new Response(
        JSON.stringify({
          url: "https://billing.stripe.test/session",
        }),
        {
          status: 200,
          headers: {
            "Content-Type": "application/json",
          },
        },
      ),
    );

    const session = await createBillingCustomerPortalSession();

    expect(fetch).toHaveBeenCalledWith(expect.any(Request));
    expect(latestRequest().url).toContain(
      "/api/billing/customer-portal-session",
    );
    expect(latestRequest().method).toBe("POST");
    expect(session.url).toBe("https://billing.stripe.test/session");
  });
});
