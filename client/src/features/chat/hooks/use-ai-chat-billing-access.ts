import { useCallback, useEffect, useMemo, useState } from "react";
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
} from "@/features/chat/api/billing";
import {
  featureAccessQueryOptions,
  getFeatureAccess,
  hasAIChatFeatureAccess,
  type FeatureAccessGrant,
} from "@/features/chat/api/feature-access";
import { showErrorToast } from "@/lib/errors";

type CheckoutSearch = "success" | "cancelled";
type BillingSearch = "cancelled" | "portal-return";

type UseAIChatBillingAccessOptions = {
  userId?: string;
  checkout?: CheckoutSearch;
  billing?: BillingSearch;
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
  | { status: "payment-confirming"; result: CheckoutAccessPollResult }
  | { status: "activating"; result: CheckoutAccessPollResult }
  | { status: "failed"; error: unknown };

export type AIChatBillingAccess = {
  billingStatus?: BillingStatusResponse;
  checkoutNotice: CheckoutSearch | null;
  billingNotice: BillingSearch | null;
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
const checkoutAccessAutoRefreshDelayMs =
  import.meta.env.MODE === "test" ? 25 : 5000;

export type AIChatAccessState =
  | "checking"
  | "ready"
  | "payment-confirming"
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

class BillingCancellationPendingError extends Error {
  result: CheckoutAccessPollResult;

  constructor(result: CheckoutAccessPollResult) {
    super("billing cancellation is still pending");
    this.name = "BillingCancellationPendingError";
    this.result = result;
  }
}

export function useAIChatBillingAccess({
  userId,
  checkout,
  billing,
  conversationId,
  navigate,
}: UseAIChatBillingAccessOptions): AIChatBillingAccess {
  const [checkoutNotice, setCheckoutNotice] = useState<CheckoutSearch | null>(
    checkout ?? null,
  );
  const [billingNotice, setBillingNotice] = useState<BillingSearch | null>(
    billing ?? null,
  );
  const [shouldPollCheckoutAccess, setShouldPollCheckoutAccess] = useState(
    checkout === "success",
  );
  const [shouldPollBillingCancellation, setShouldPollBillingCancellation] =
    useState(billing === "cancelled");
  const [settledCheckoutPollingView, setSettledCheckoutPollingView] =
    useState<CheckoutAccessPollingView>({ status: "idle" });
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
  const billingCancellationQuery = useQuery({
    queryKey: [
      "billing",
      "ai-chatbot",
      "billing-cancellation",
      userId,
      conversationId,
    ],
    queryFn: waitForBillingCancellation,
    enabled: isSignedIn && shouldPollBillingCancellation,
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
  const restartCheckoutAccessPolling = useCallback(() => {
    setSettledCheckoutPollingView({ status: "idle" });
    setShouldPollCheckoutAccess(true);
    void billingQuery.refetch();
    void featureAccessQuery.refetch();
  }, [billingQuery.refetch, featureAccessQuery.refetch]);

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
    if (!billing) {
      return;
    }

    setBillingNotice(billing);
    if (billing === "cancelled") {
      setShouldPollBillingCancellation(true);
    } else {
      void billingQuery.refetch();
      void featureAccessQuery.refetch();
    }
    void navigate({
      to: "/chat",
      search: { conversationId },
      replace: true,
    });
  }, [
    billing,
    billingQuery.refetch,
    conversationId,
    featureAccessQuery.refetch,
    navigate,
  ]);

  useEffect(() => {
    if (
      billingCancellationQuery.isSuccess ||
      billingCancellationQuery.isError
    ) {
      setShouldPollBillingCancellation(false);
    }
  }, [billingCancellationQuery.isError, billingCancellationQuery.isSuccess]);

  const currentCheckoutPollingView = useMemo(
    () =>
      resolveCheckoutAccessPollingView({
        data: checkoutAccessQuery.data,
        error: checkoutAccessQuery.error,
        isFetching: checkoutAccessQuery.isFetching,
        isError: checkoutAccessQuery.isError,
      }),
    [
      checkoutAccessQuery.data,
      checkoutAccessQuery.error,
      checkoutAccessQuery.isError,
      checkoutAccessQuery.isFetching,
    ],
  );

  useEffect(() => {
    if (checkoutNotice !== "success") {
      setSettledCheckoutPollingView((previousView) =>
        previousView.status === "idle" ? previousView : { status: "idle" },
      );
      return;
    }

    if (isTerminalCheckoutPollingView(currentCheckoutPollingView)) {
      setSettledCheckoutPollingView(currentCheckoutPollingView);
    }

    if (checkoutAccessQuery.isSuccess || checkoutAccessQuery.isError) {
      setShouldPollCheckoutAccess(false);
    }
  }, [
    checkoutAccessQuery.isError,
    checkoutAccessQuery.isSuccess,
    checkoutNotice,
    currentCheckoutPollingView,
  ]);

  const checkoutPollingView =
    checkoutNotice === "success" &&
    isTerminalCheckoutPollingView(settledCheckoutPollingView) &&
    !isTerminalCheckoutPollingView(currentCheckoutPollingView)
      ? settledCheckoutPollingView
      : currentCheckoutPollingView;
  const checkoutAccessOverride = getCheckoutAccessOverride({
    checkoutPollingView,
    featureAccess: featureAccessQuery.data,
  });
  const billingCancellationOverride = getBillingCancellationOverride({
    data: billingCancellationQuery.data,
    error: billingCancellationQuery.error,
    isError: billingCancellationQuery.isError,
  });
  const errorSource = getAIChatAccessErrorSource({
    isBillingError: billingQuery.isError || featureAccessQuery.isError,
    checkoutPollingView,
    isCheckoutReturn: checkoutNotice === "success",
  });
  const accessView = resolveAIChatAccessView({
    billingStatus:
      checkoutAccessOverride?.billingStatus ??
      billingCancellationOverride?.billingStatus ??
      billingQuery.data,
    featureAccess:
      checkoutAccessOverride?.featureAccess ??
      billingCancellationOverride?.featureAccess ??
      featureAccessQuery.data,
    isPaymentConfirming: checkoutPollingView.status === "payment-confirming",
    isChecking:
      billingQuery.isLoading ||
      billingQuery.isPending ||
      featureAccessQuery.isLoading ||
      featureAccessQuery.isPending ||
      checkoutPollingView.status === "polling" ||
      billingCancellationQuery.isFetching,
    errorSource,
  });
  const isRefreshingAccess =
    checkoutAccessQuery.isFetching ||
    billingCancellationQuery.isFetching ||
    Boolean(billingQuery.isFetching) ||
    Boolean(featureAccessQuery.isFetching);

  useEffect(() => {
    if (checkoutNotice !== "success" && accessView.state !== "activating") {
      return;
    }

    if (
      accessView.state !== "payment-confirming" &&
      accessView.state !== "activating"
    ) {
      return;
    }

    if (isRefreshingAccess) {
      return;
    }

    const timeoutId = window.setTimeout(
      restartCheckoutAccessPolling,
      checkoutAccessAutoRefreshDelayMs,
    );

    return () => window.clearTimeout(timeoutId);
  }, [
    accessView.state,
    checkoutNotice,
    isRefreshingAccess,
    restartCheckoutAccessPolling,
  ]);

  return {
    billingStatus: accessView.billingStatus,
    checkoutNotice,
    billingNotice,
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
        accessView.state === "payment-confirming" ||
        accessView.state === "checkout-activation-error"
      ) {
        restartCheckoutAccessPolling();
        return;
      }
      void billingQuery.refetch();
      void featureAccessQuery.refetch();
    },
  };
}

export function resolveAIChatAccessView({
  billingStatus,
  featureAccess,
  isPaymentConfirming,
  isChecking,
  errorSource,
}: {
  billingStatus?: BillingStatusResponse;
  featureAccess?: FeatureAccessGrant[];
  isPaymentConfirming?: boolean;
  isChecking?: boolean;
  errorSource?: AIChatAccessErrorSource;
}) {
  const hasChatAccess = hasAIChatFeatureAccess(featureAccess);
  const state = resolveAIChatAccessState({
    billingStatus,
    hasChatAccess,
    isPaymentConfirming,
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
  isError,
  isFetching,
}: {
  data?: CheckoutAccessPollResult;
  error?: unknown;
  isError?: boolean;
  isFetching?: boolean;
}): CheckoutAccessPollingView {
  if (isFetching) {
    return { status: "polling" };
  }

  if (data) {
    return resolveCheckoutAccessSettledView(data);
  }

  if (error instanceof CheckoutAccessPendingError && isError !== false) {
    return resolveCheckoutAccessSettledView(error.result);
  }

  if (error && isError !== false) {
    return { status: "failed", error };
  }

  if (error) {
    return { status: "polling" };
  }

  return { status: "idle" };
}

function getCheckoutAccessOverride({
  checkoutPollingView,
  featureAccess,
}: {
  checkoutPollingView: CheckoutAccessPollingView;
  featureAccess?: FeatureAccessGrant[];
}): CheckoutAccessPollResult | undefined {
  switch (checkoutPollingView.status) {
    case "payment-confirming":
    case "activating":
      return checkoutPollingView.result;
    case "ready":
      return hasAIChatFeatureAccess(featureAccess)
        ? undefined
        : checkoutPollingView.result;
    case "idle":
    case "polling":
    case "failed":
      return undefined;
  }
}

function getBillingCancellationOverride({
  data,
  error,
  isError,
}: {
  data?: CheckoutAccessPollResult;
  error?: unknown;
  isError?: boolean;
}): CheckoutAccessPollResult | undefined {
  if (data) {
    return data;
  }

  if (error instanceof BillingCancellationPendingError && isError !== false) {
    return error.result;
  }

  return undefined;
}

function resolveCheckoutAccessSettledView(
  result: CheckoutAccessPollResult,
): Exclude<
  CheckoutAccessPollingView,
  { status: "idle" | "polling" | "failed" }
> {
  if (hasAIChatFeatureAccess(result.featureAccess)) {
    return { status: "ready", result };
  }

  if (result.billingStatus.has_access) {
    return { status: "activating", result };
  }

  return { status: "payment-confirming", result };
}

function isTerminalCheckoutPollingView(
  view: CheckoutAccessPollingView,
): view is Extract<
  CheckoutAccessPollingView,
  { status: "ready" | "payment-confirming" | "activating" | "failed" }
> {
  return (
    view.status === "ready" ||
    view.status === "payment-confirming" ||
    view.status === "activating" ||
    view.status === "failed"
  );
}

function resolveAIChatAccessState({
  billingStatus,
  hasChatAccess,
  isPaymentConfirming,
  isChecking,
  errorSource,
}: {
  billingStatus?: BillingStatusResponse;
  hasChatAccess: boolean;
  isPaymentConfirming?: boolean;
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

  if (isPaymentConfirming) {
    return "payment-confirming";
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
  isCheckoutReturn,
}: {
  isBillingError: boolean;
  checkoutPollingView: CheckoutAccessPollingView;
  isCheckoutReturn: boolean;
}): AIChatAccessErrorSource | undefined {
  switch (checkoutPollingView.status) {
    case "failed":
      return "checkout-activation";
    case "polling":
    case "ready":
    case "payment-confirming":
    case "activating":
      return undefined;
    case "idle":
      if (isCheckoutReturn && isBillingError) {
        return "checkout-activation";
      }

      return isBillingError ? "billing" : undefined;
  }
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

async function waitForBillingCancellation(): Promise<CheckoutAccessPollResult> {
  const [featureAccess, billingStatus] = await Promise.all([
    getFeatureAccess(),
    getBillingStatus(),
  ]);

  if (!isBillingCancellationReflected(billingStatus)) {
    throw new BillingCancellationPendingError({ billingStatus, featureAccess });
  }

  return { billingStatus, featureAccess };
}

function isBillingCancellationReflected(
  billingStatus: BillingStatusResponse,
): boolean {
  const subscription = billingStatus.subscription;
  return (
    !subscription ||
    subscription.cancel_at_period_end ||
    Boolean(subscription.cancel_at) ||
    subscription.status === "canceled"
  );
}
