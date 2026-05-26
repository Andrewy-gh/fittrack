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
  hasChatAccess: boolean;
  hasBillingDisplayAccess: boolean;
  isCheckingAccess: boolean;
  isBillingCardLoading: boolean;
  isBillingError: boolean;
  isCheckoutLoading: boolean;
  isBillingPortalLoading: boolean;
  startCheckout: () => void;
  manageBilling: () => void;
  refreshBillingStatus: () => void;
};

const checkoutAccessRetryDelaysMs = [1000, 2000, 4000, 8000];

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
  });
  const isFeatureAccessLoading =
    featureAccessQuery.isLoading || featureAccessQuery.isPending;
  const isBillingLoading = billingQuery.isLoading || billingQuery.isPending;
  const isCheckingAccess =
    isFeatureAccessLoading || checkoutAccessQuery.isFetching;

  return {
    billingStatus: accessView.billingStatus,
    checkoutNotice,
    hasChatAccess: accessView.hasChatAccess,
    hasBillingDisplayAccess: accessView.hasBillingDisplayAccess,
    isCheckingAccess,
    isBillingCardLoading: isBillingLoading || isCheckingAccess,
    isBillingError: billingQuery.isError,
    isCheckoutLoading: checkoutMutation.isPending,
    isBillingPortalLoading: billingPortalMutation.isPending,
    startCheckout: () => checkoutMutation.mutate(),
    manageBilling: () => billingPortalMutation.mutate(),
    refreshBillingStatus: () => {
      void billingQuery.refetch();
    },
  };
}

export function resolveAIChatAccessView({
  billingStatus,
  featureAccess,
}: {
  billingStatus?: BillingStatusResponse;
  featureAccess?: FeatureAccessGrant[];
}) {
  const hasChatAccess = hasAIChatFeatureAccess(featureAccess);

  return {
    billingStatus,
    hasChatAccess,
    hasBillingDisplayAccess:
      hasChatAccess || billingStatus?.has_access === true,
  };
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
