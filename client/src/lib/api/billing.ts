import { queryOptions } from "@tanstack/react-query";
import { client } from "@/client/client.gen";
import type { ApiError } from "@/lib/errors";
import "./client-config";

export type BillingSubscriptionStatus =
  | "trialing"
  | "active"
  | "past_due"
  | "unpaid"
  | "canceled"
  | "incomplete"
  | "incomplete_expired";

export type BillingSubscription = {
  stripe_subscription_id: string;
  status: BillingSubscriptionStatus;
  cancel_at_period_end: boolean;
  current_period_end?: string;
  trial_end?: string;
};

export type BillingTrialUsage = {
  used: number;
  limit: number;
};

export type BillingStatusResponse = {
  feature_key: "ai_chatbot";
  has_access: boolean;
  subscription?: BillingSubscription;
  trial_usage?: BillingTrialUsage;
};

export type BillingCheckoutSessionResponse = {
  url: string;
};

type BillingStatusResponses = {
  200: BillingStatusResponse;
};

type BillingCheckoutSessionResponses = {
  200: BillingCheckoutSessionResponse;
};

export function billingStatusQueryOptions() {
  return queryOptions({
    queryKey: ["billing", "ai-chatbot", "status"],
    queryFn: ({ signal }) => getBillingStatus({ signal }),
  });
}

export async function getBillingStatus(
  options: {
    signal?: AbortSignal;
  } = {},
): Promise<BillingStatusResponse> {
  const response = await client.get<BillingStatusResponses, ApiError, true>({
    url: "/billing/status",
    signal: options.signal,
    throwOnError: true,
  });

  return response.data;
}

export async function createBillingCheckoutSession(): Promise<BillingCheckoutSessionResponse> {
  const response = await client.post<
    BillingCheckoutSessionResponses,
    ApiError,
    true
  >({
    url: "/billing/checkout-session",
    throwOnError: true,
  });

  return response.data;
}

export function redirectToBillingCheckout(url: string): void {
  window.location.assign(url);
}
