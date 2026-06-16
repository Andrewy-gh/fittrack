import { queryOptions } from "@tanstack/react-query";
import { client } from "@/client/client.gen";
import type { ApiError } from "@/lib/errors";
import "@/lib/api/client-config";

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
  cancellation_scheduled: boolean;
  access_ends_at?: string;
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

export type BillingCustomerPortalSessionResponse = {
  url: string;
};

type BillingStatusResponses = {
  200: BillingStatusResponse;
};

type BillingCheckoutSessionResponses = {
  200: BillingCheckoutSessionResponse;
};

type BillingCustomerPortalSessionResponses = {
  200: BillingCustomerPortalSessionResponse;
};

export function billingStatusQueryOptions(userId?: string) {
  return queryOptions({
    queryKey: ["billing", "ai-chatbot", "status", userId],
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

export async function createBillingCustomerPortalSession(): Promise<BillingCustomerPortalSessionResponse> {
  const response = await client.post<
    BillingCustomerPortalSessionResponses,
    ApiError,
    true
  >({
    url: "/billing/customer-portal-session",
    throwOnError: true,
  });

  return response.data;
}

export async function createBillingSubscriptionCancelPortalSession(): Promise<BillingCustomerPortalSessionResponse> {
  const response = await client.post<
    BillingCustomerPortalSessionResponses,
    ApiError,
    true
  >({
    url: "/billing/subscription-cancel-portal-session",
    throwOnError: true,
  });

  return response.data;
}

export function redirectToBillingCheckout(url: string): void {
  window.location.assign(url);
}

export function redirectToBillingPortal(url: string): void {
  window.location.assign(url);
}
