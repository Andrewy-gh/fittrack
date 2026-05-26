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

type CheckoutAccessPollResult = {
  billingStatus: BillingStatusResponse;
  featureAccess: FeatureAccessGrant[];
};

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

const checkoutAccessRetryDelaysMs = [1000, 2000, 4000, 8000];

export type AIChatAccessState =
  | "checking"
  | "ready"
  | "activating"
  | "blocked"
  | "error";

class CheckoutAccessPendingError extends Error {
  constructor() {
    super("checkout access is still pending");
    this.name = "CheckoutAccessPendingError";
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
    ...billingStatusQueryOptions(),
    enabled: isSignedIn,
  });
  const featureAccessQuery = useQuery({
    ...featureAccessQueryOptions(),
    enabled: isSignedIn,
  });
  const checkoutAccessQuery = useQuery({
    queryKey: ["billing", "ai-chatbot", "checkout-access", conversationId],
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

  const accessView = resolveAIChatAccessView({
    billingStatus: checkoutAccessQuery.data?.billingStatus ?? billingQuery.data,
    featureAccess:
      checkoutAccessQuery.data?.featureAccess ?? featureAccessQuery.data,
    isChecking:
      billingQuery.isLoading ||
      billingQuery.isPending ||
      featureAccessQuery.isLoading ||
      featureAccessQuery.isPending ||
      checkoutAccessQuery.isFetching,
    isError: billingQuery.isError || featureAccessQuery.isError,
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
    isBillingError: accessView.state === "error",
    isRefreshingAccess,
    isCheckoutLoading: checkoutMutation.isPending,
    isBillingPortalLoading: billingPortalMutation.isPending,
    startCheckout: () => checkoutMutation.mutate(),
    manageBilling: () => billingPortalMutation.mutate(),
    refreshAccess: () => {
      if (accessView.state === "activating") {
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
  isError,
}: {
  billingStatus?: BillingStatusResponse;
  featureAccess?: FeatureAccessGrant[];
  isChecking?: boolean;
  isError?: boolean;
}) {
  const hasChatAccess = hasAIChatFeatureAccess(featureAccess);
  const state = resolveAIChatAccessState({
    billingStatus,
    hasChatAccess,
    isChecking,
    isError,
  });

  return {
    billingStatus,
    state,
    hasChatAccess,
  };
}

function resolveAIChatAccessState({
  billingStatus,
  hasChatAccess,
  isChecking,
  isError,
}: {
  billingStatus?: BillingStatusResponse;
  hasChatAccess: boolean;
  isChecking?: boolean;
  isError?: boolean;
}): AIChatAccessState {
  if (hasChatAccess) {
    return "ready";
  }

  if (isError) {
    return "error";
  }

  if (isChecking) {
    return "checking";
  }

  if (billingStatus?.has_access === true) {
    return "activating";
  }

  return "blocked";
}

async function waitForCheckoutAccess(): Promise<CheckoutAccessPollResult> {
  const [featureAccess, billingStatus] = await Promise.all([
    getFeatureAccess(),
    getBillingStatus(),
  ]);

  if (!hasAIChatFeatureAccess(featureAccess)) {
    throw new CheckoutAccessPendingError();
  }

  return { billingStatus, featureAccess };
}
