import { useEffect, useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import type { NavigateFn } from "@tanstack/react-router";
import {
  billingStatusQueryOptions,
  createBillingCheckoutSession,
  createBillingCustomerPortalSession,
  getBillingStatus,
  redirectToBillingCheckout,
  redirectToBillingPortal,
  type BillingStatusResponse,
} from "@/lib/api/billing";
import {
  featureAccessQueryOptions,
  getFeatureAccess,
  hasAIChatFeatureAccess,
  type FeatureAccessGrant,
} from "@/lib/api/feature-access";
import { showErrorToast } from "@/lib/errors";

type CheckoutSearch = "success" | "cancelled";

type UseAIChatBillingAccessOptions = {
  userId?: string;
  checkout?: CheckoutSearch;
  conversationId?: string;
  navigate: NavigateFn;
};

export type CheckoutAccessPollResult = {
  billingStatus: BillingStatusResponse;
  featureAccess: FeatureAccessGrant[];
};

export type CheckoutAccessPollingView =
  | { status: "idle" }
  | { status: "polling" }
  | { status: "ready"; result: CheckoutAccessPollResult }
  | { status: "activating"; result: CheckoutAccessPollResult }
  | { status: "failed"; error: unknown };

export type AIChatBillingAccess = {
  billingStatus?: BillingStatusResponse;
  checkoutNotice: CheckoutSearch | null;
  accessState: AIChatAccessState;
  hasChatAccess: boolean;
  isCheckingAccess: boolean;
  isBillingCardLoading: boolean;
  isBillingError: boolean;
  isRefreshingAccess: boolean;
  isCheckoutLoading: boolean;
  isBillingPortalLoading: boolean;
  startCheckout: () => void;
  manageBilling: () => void;
  refreshAccess: () => void;
};

const checkoutAccessRetryDelaysMs =
  import.meta.env.MODE === "test" ? [1, 1, 1, 1] : [1000, 2000, 4000, 8000];

export type AIChatAccessState =
  | "checking"
  | "ready"
  | "activating"
  | "blocked"
  | "billing-error"
  | "checkout-activation-error";

type AIChatAccessErrorSource = "billing" | "checkout-activation";

class CheckoutAccessPendingError extends Error {
  result: CheckoutAccessPollResult;

  constructor(result: CheckoutAccessPollResult) {
    super("checkout access is still pending");
    this.name = "CheckoutAccessPendingError";
    this.result = result;
  }
}

export function useAIChatBillingAccess({
  userId,
  checkout,
  conversationId,
  navigate,
}: UseAIChatBillingAccessOptions): AIChatBillingAccess {
  const [checkoutNotice, setCheckoutNotice] = useState<CheckoutSearch | null>(
    checkout ?? null,
  );
  const [shouldPollCheckoutAccess, setShouldPollCheckoutAccess] = useState(
    checkout === "success",
  );
  const isSignedIn = Boolean(userId);

  const billingQuery = useQuery({
    ...billingStatusQueryOptions(userId),
    enabled: isSignedIn,
  });
  const featureAccessQuery = useQuery({
    ...featureAccessQueryOptions(userId),
    enabled: isSignedIn,
  });
  const checkoutAccessQuery = useQuery({
    queryKey: [
      "billing",
      "ai-chatbot",
      "checkout-access",
      userId,
      conversationId,
    ],
    queryFn: waitForCheckoutAccess,
    enabled: isSignedIn && shouldPollCheckoutAccess,
    retry: checkoutAccessRetryDelaysMs.length,
    retryDelay: (failureCount) =>
      checkoutAccessRetryDelaysMs[
        Math.min(
          Math.max(failureCount - 1, 0),
          checkoutAccessRetryDelaysMs.length - 1,
        )
      ],
  });

  const checkoutMutation = useMutation({
    mutationFn: createBillingCheckoutSession,
    onSuccess: (session) => redirectToBillingCheckout(session.url),
    onError: (error) => showErrorToast(error, "Could not open Checkout"),
  });
  const billingPortalMutation = useMutation({
    mutationFn: createBillingCustomerPortalSession,
    onSuccess: (session) => redirectToBillingPortal(session.url),
    onError: (error) => showErrorToast(error, "Could not open billing"),
  });

  useEffect(() => {
    if (!checkout) {
      return;
    }

    setCheckoutNotice(checkout);
    if (checkout === "success") {
      setShouldPollCheckoutAccess(true);
    }

    void navigate({
      to: "/chat",
      search: { conversationId },
      replace: true,
    });
  }, [checkout, conversationId, navigate]);

  useEffect(() => {
    if (checkoutAccessQuery.isSuccess || checkoutAccessQuery.isError) {
      setShouldPollCheckoutAccess(false);
    }
  }, [checkoutAccessQuery.isError, checkoutAccessQuery.isSuccess]);

  const checkoutPollingView = resolveCheckoutAccessPollingView({
    data: checkoutAccessQuery.data,
    error: checkoutAccessQuery.error,
    isFetching: checkoutAccessQuery.isFetching,
  });
  const checkoutAccessResult =
    getCheckoutAccessPollingResult(checkoutPollingView);
  const errorSource = getAIChatAccessErrorSource({
    isBillingError: billingQuery.isError || featureAccessQuery.isError,
    checkoutPollingView,
  });
  const accessView = resolveAIChatAccessView({
    billingStatus: checkoutAccessResult?.billingStatus ?? billingQuery.data,
    featureAccess:
      checkoutAccessResult?.featureAccess ?? featureAccessQuery.data,
    isChecking:
      billingQuery.isLoading ||
      billingQuery.isPending ||
      featureAccessQuery.isLoading ||
      featureAccessQuery.isPending ||
      checkoutPollingView.status === "polling",
    errorSource,
  });
  const isRefreshingAccess =
    checkoutAccessQuery.isFetching ||
    Boolean(billingQuery.isFetching) ||
    Boolean(featureAccessQuery.isFetching);

  return {
    billingStatus: accessView.billingStatus,
    checkoutNotice,
    accessState: accessView.state,
    hasChatAccess: accessView.hasChatAccess,
    isCheckingAccess: accessView.state === "checking",
    isBillingCardLoading: accessView.state === "checking",
    isBillingError: isAIChatAccessErrorState(accessView.state),
    isRefreshingAccess,
    isCheckoutLoading: checkoutMutation.isPending,
    isBillingPortalLoading: billingPortalMutation.isPending,
    startCheckout: () => checkoutMutation.mutate(),
    manageBilling: () => billingPortalMutation.mutate(),
    refreshAccess: () => {
      if (
        accessView.state === "activating" ||
        accessView.state === "checkout-activation-error"
      ) {
        setShouldPollCheckoutAccess(true);
      }
      void billingQuery.refetch();
      void featureAccessQuery.refetch();
    },
  };
}

export function resolveAIChatAccessView({
  billingStatus,
  featureAccess,
  isChecking,
  errorSource,
}: {
  billingStatus?: BillingStatusResponse;
  featureAccess?: FeatureAccessGrant[];
  isChecking?: boolean;
  errorSource?: AIChatAccessErrorSource;
}) {
  const hasChatAccess = hasAIChatFeatureAccess(featureAccess);
  const state = resolveAIChatAccessState({
    billingStatus,
    hasChatAccess,
    isChecking,
    errorSource,
  });

  return {
    billingStatus,
    state,
    hasChatAccess,
  };
}

export function resolveCheckoutAccessPollingView({
  data,
  error,
  isFetching,
}: {
  data?: CheckoutAccessPollResult;
  error?: unknown;
  isFetching?: boolean;
}): CheckoutAccessPollingView {
  if (isFetching) {
    return { status: "polling" };
  }

  if (data) {
    return hasAIChatFeatureAccess(data.featureAccess)
      ? { status: "ready", result: data }
      : { status: "activating", result: data };
  }

  if (error instanceof CheckoutAccessPendingError) {
    return { status: "activating", result: error.result };
  }

  if (error) {
    return { status: "failed", error };
  }

  return { status: "idle" };
}

function getCheckoutAccessPollingResult(
  view: CheckoutAccessPollingView,
): CheckoutAccessPollResult | undefined {
  if (view.status === "ready" || view.status === "activating") {
    return view.result;
  }

  return undefined;
}

function resolveAIChatAccessState({
  billingStatus,
  hasChatAccess,
  isChecking,
  errorSource,
}: {
  billingStatus?: BillingStatusResponse;
  hasChatAccess: boolean;
  isChecking?: boolean;
  errorSource?: AIChatAccessErrorSource;
}): AIChatAccessState {
  if (hasChatAccess) {
    return "ready";
  }

  if (errorSource === "checkout-activation") {
    return "checkout-activation-error";
  }

  if (errorSource === "billing") {
    return "billing-error";
  }

  if (isChecking) {
    return "checking";
  }

  if (billingStatus?.has_access === true) {
    return "activating";
  }

  return "blocked";
}

function getAIChatAccessErrorSource({
  isBillingError,
  checkoutPollingView,
}: {
  isBillingError: boolean;
  checkoutPollingView: CheckoutAccessPollingView;
}): AIChatAccessErrorSource | undefined {
  if (checkoutPollingView.status === "failed") {
    return "checkout-activation";
  }

  return isBillingError ? "billing" : undefined;
}

function isAIChatAccessErrorState(state: AIChatAccessState): boolean {
  return state === "billing-error" || state === "checkout-activation-error";
}

async function waitForCheckoutAccess(): Promise<CheckoutAccessPollResult> {
  const [featureAccess, billingStatus] = await Promise.all([
    getFeatureAccess(),
    getBillingStatus(),
  ]);

  if (!hasAIChatFeatureAccess(featureAccess)) {
    throw new CheckoutAccessPendingError({ billingStatus, featureAccess });
  }

  return { billingStatus, featureAccess };
}
