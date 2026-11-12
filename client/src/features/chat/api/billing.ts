import { queryOptions } from "@tanstack/react-query";
import { client } from "@/client/client.gen";
import type {
  BillingStatusResponse as GeneratedBillingStatus,
  BillingSubscriptionView,
  PostBillingCustomerPortalSessionData,
  PostBillingCheckoutSessionResponses as BillingCheckoutSessionResponses,
  PostBillingCustomerPortalSessionResponses as BillingCustomerPortalSessionResponses,
  GetBillingStatusResponses as BillingStatusResponses,
} from "@/client/types.gen";
import type { ApiError } from "@/lib/errors";
import "@/lib/api/client-config";

/** Subscription data returned by billing. */
export type BillingSubscription = BillingSubscriptionView;
/** Provider status; unknown future statuses retain the UI's fallback behavior. */
export type BillingSubscriptionStatus = BillingSubscription["status"];
/** Current access and trial usage from the server. */
export type BillingStatusResponse = GeneratedBillingStatus;
/** Hosted checkout session location. */
export type BillingCheckoutSessionResponse =
  BillingCheckoutSessionResponses[200];
/** Hosted billing portal location. */
export type BillingCustomerPortalSessionResponse =
  BillingCustomerPortalSessionResponses[200];

type BillingPortalReturnDestination = NonNullable<
  NonNullable<PostBillingCustomerPortalSessionData["query"]>["return_to"]
>;

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

export async function createBillingCustomerPortalSession(
  destination: BillingPortalReturnDestination = "chat",
): Promise<BillingCustomerPortalSessionResponse> {
  const response = await client.post<
    BillingCustomerPortalSessionResponses,
    ApiError,
    true
  >({
    url:
      destination === "settings"
        ? "/billing/customer-portal-session?return_to=settings"
        : "/billing/customer-portal-session",
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
