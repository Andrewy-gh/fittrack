import {
  AlertCircle,
  CheckCircle2,
  CreditCard,
  RefreshCw,
  Sparkles,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import type {
  BillingStatusResponse,
  BillingSubscription,
  BillingSubscriptionStatus,
} from "@/lib/api/billing";
import { cn } from "@/lib/utils";

type AIChatBillingCardProps = {
  status?: BillingStatusResponse;
  accessState: AIChatBillingCardAccessState;
  isLoading?: boolean;
  isError?: boolean;
  isRefreshingAccess?: boolean;
  isCheckoutLoading?: boolean;
  isBillingPortalLoading?: boolean;
  onStartCheckout: () => void;
  onManageBilling: () => void;
  onRefreshAccess: () => void;
};

export type AIChatBillingCardAccessState =
  | "checking"
  | "ready"
  | "activating"
  | "blocked"
  | "error";

const checkoutBlockedStatuses = new Set<BillingSubscriptionStatus>([
  "canceled",
  "incomplete",
  "incomplete_expired",
]);

const billingPortalStatuses = new Set<BillingSubscriptionStatus>([
  "past_due",
  "unpaid",
]);

export function AIChatBillingCard({
  status,
  accessState,
  isLoading = false,
  isError = false,
  isRefreshingAccess = false,
  isCheckoutLoading = false,
  isBillingPortalLoading = false,
  onStartCheckout,
  onManageBilling,
  onRefreshAccess,
}: AIChatBillingCardProps) {
  const subscription = status?.subscription;
  const isTrialing = status?.has_access && subscription?.status === "trialing";
  const isActive = status?.has_access && subscription?.status === "active";
  const isBlocked =
    !status?.has_access &&
    subscription?.status !== undefined &&
    isBlockedStatus(subscription.status);
  const hasChatAccess = accessState === "ready";
  const billingAction = getBillingAction(status, accessState);
  const shouldShowBillingAction =
    !isLoading &&
    !isError &&
    billingAction &&
    (accessState !== "ready" || billingAction === "portal");
  const isActionLoading = getActionLoadingState({
    billingAction,
    isBillingPortalLoading,
    isCheckoutLoading,
    isRefreshingAccess,
  });

  return (
    <Card className="rounded-lg">
      <CardContent className="space-y-4 px-4 py-4">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
          <div className="space-y-2">
            <div className="flex flex-wrap items-center gap-2">
              <h2 className="text-lg font-semibold tracking-tight">
                Premium AI Chat
              </h2>
              <BillingStatusBadge
                isLoading={isLoading}
                isError={isError}
                status={status}
                accessState={accessState}
              />
            </div>
            <p className="text-sm text-muted-foreground">
              AI chat is a premium FitTrack feature on one monthly plan.
            </p>
          </div>

          {shouldShowBillingAction ? (
            <Button
              type="button"
              className="w-full sm:w-auto"
              onClick={
                billingAction === "portal"
                  ? onManageBilling
                  : billingAction === "refresh"
                    ? onRefreshAccess
                    : onStartCheckout
              }
              disabled={isActionLoading || isLoading}
            >
              {billingAction === "refresh" ? (
                <RefreshCw className="size-4" />
              ) : (
                <CreditCard className="size-4" />
              )}
              {isActionLoading
                ? billingActionLoadingLabel(billingAction)
                : billingActionLabel(status, billingAction)}
            </Button>
          ) : null}
        </div>

        <BillingMessage
          status={status}
          isLoading={isLoading}
          isError={isError}
          isBlocked={isBlocked}
          isTrialing={Boolean(isTrialing)}
          isActive={Boolean(isActive)}
          accessState={accessState}
          hasChatAccess={hasChatAccess}
        />
      </CardContent>
    </Card>
  );
}

function BillingStatusBadge({
  status,
  isLoading,
  isError,
  accessState,
}: {
  status?: BillingStatusResponse;
  isLoading: boolean;
  isError: boolean;
  accessState: AIChatBillingCardAccessState;
}) {
  if (isLoading) {
    return <Badge variant="secondary">Checking</Badge>;
  }

  if (isError) {
    return (
      <Badge
        variant="outline"
        className="border-destructive/30 text-destructive"
      >
        Unavailable
      </Badge>
    );
  }

  if (accessState === "activating") {
    return <Badge variant="secondary">Activating</Badge>;
  }

  if (!status?.subscription) {
    if (accessState === "ready") {
      return (
        <Badge>
          <CheckCircle2 className="size-3" />
          Access active
        </Badge>
      );
    }

    return <Badge variant="outline">Trial available</Badge>;
  }

  if (status.has_access && status.subscription.status === "trialing") {
    return (
      <Badge>
        <Sparkles className="size-3" />
        Trial
      </Badge>
    );
  }

  if (status.has_access && status.subscription.status === "active") {
    return (
      <Badge>
        <CheckCircle2 className="size-3" />
        Premium
      </Badge>
    );
  }

  if (accessState === "ready") {
    return (
      <Badge>
        <CheckCircle2 className="size-3" />
        Access active
      </Badge>
    );
  }

  return (
    <Badge
      variant="outline"
      className="border-destructive/30 text-destructive"
    >
      Action needed
    </Badge>
  );
}

function BillingMessage({
  status,
  isLoading,
  isError,
  isBlocked,
  isTrialing,
  isActive,
  accessState,
  hasChatAccess,
}: {
  status?: BillingStatusResponse;
  isLoading: boolean;
  isError: boolean;
  isBlocked: boolean;
  isTrialing: boolean;
  isActive: boolean;
  accessState: AIChatBillingCardAccessState;
  hasChatAccess: boolean;
}) {
  if (isLoading) {
    return (
      <p className="text-sm text-muted-foreground">
        Checking your AI chat access...
      </p>
    );
  }

  if (isError) {
    return (
      <p className="text-sm text-muted-foreground">
        We could not confirm billing status. Refresh the page or try again soon.
      </p>
    );
  }

  if (!status) {
    return (
      <p className="text-sm text-muted-foreground">
        Start a 7-day card-required trial to unlock AI chat.
      </p>
    );
  }

  const cancellationMessage = getCancellationMessage(status.subscription);

  if (accessState === "activating") {
    return (
      <p className="text-sm text-muted-foreground">
        Premium is active. We are finishing AI chat activation. Refresh access
        if this takes more than a moment.
      </p>
    );
  }

  if (hasChatAccess && !isTrialing && !isActive) {
    return (
      <p className="text-sm text-muted-foreground">
        {billingPortalStatuses.has(status.subscription?.status ?? "active")
          ? "AI chat access is active for this account. Billing still needs attention."
          : "AI chat access is active for this account."}
      </p>
    );
  }

  if (isTrialing) {
    return (
      <div className="space-y-2">
        <p className="text-sm text-muted-foreground">
          Your 7-day trial is active.
        </p>
        {status.trial_usage ? (
          <p className="text-sm font-medium">
            {status.trial_usage.used} of {status.trial_usage.limit} trial
            prompts used
          </p>
        ) : null}
        {cancellationMessage ? (
          <p className="text-sm text-muted-foreground">{cancellationMessage}</p>
        ) : null}
      </div>
    );
  }

  if (isActive) {
    return (
      <div className="space-y-2">
        <p className="text-sm text-muted-foreground">
          Premium is active. AI chat is available.
        </p>
        {cancellationMessage ? (
          <p className="text-sm text-muted-foreground">{cancellationMessage}</p>
        ) : null}
      </div>
    );
  }

  if (isBlocked) {
    return (
      <div
        className={cn(
          "flex gap-2 rounded-md border border-destructive/30",
          "bg-destructive/5 p-3 text-sm text-destructive",
        )}
      >
        <AlertCircle className="mt-0.5 size-4 shrink-0" />
        <p>{blockedStatusMessage(status.subscription?.status)}</p>
      </div>
    );
  }

  return (
    <p className="text-sm text-muted-foreground">
      Start a 7-day card-required trial to unlock AI chat. The trial includes 30
      AI prompts.
    </p>
  );
}

type BillingAction = "checkout" | "portal" | "refresh";

function getBillingAction(
  status: BillingStatusResponse | undefined,
  accessState: AIChatBillingCardAccessState,
): BillingAction {
  if (accessState === "activating") {
    return "refresh";
  }

  if (billingPortalStatuses.has(status?.subscription?.status ?? "active")) {
    return "portal";
  }

  return "checkout";
}

function billingActionLabel(
  status: BillingStatusResponse | undefined,
  action: BillingAction,
): string {
  if (action === "refresh") {
    return "Refresh access";
  }

  if (action === "portal") {
    return "Update billing";
  }

  if (!status?.subscription) {
    return "Start 7-day trial";
  }

  switch (status.subscription.status) {
    case "incomplete":
      return "Finish Checkout";
    case "incomplete_expired":
      return "Restart Checkout";
    case "canceled":
      return "Restart premium";
    default:
      return "Start 7-day trial";
  }
}

function billingActionLoadingLabel(action: BillingAction): string {
  switch (action) {
    case "portal":
      return "Opening billing...";
    case "refresh":
      return "Refreshing access...";
    case "checkout":
      return "Opening Checkout...";
  }
}

function getActionLoadingState({
  billingAction,
  isBillingPortalLoading,
  isCheckoutLoading,
  isRefreshingAccess,
}: {
  billingAction: BillingAction;
  isBillingPortalLoading: boolean;
  isCheckoutLoading: boolean;
  isRefreshingAccess: boolean;
}): boolean {
  switch (billingAction) {
    case "portal":
      return isBillingPortalLoading;
    case "refresh":
      return isRefreshingAccess;
    case "checkout":
      return isCheckoutLoading;
  }
}

function isBlockedStatus(status: BillingSubscriptionStatus): boolean {
  return (
    checkoutBlockedStatuses.has(status) || billingPortalStatuses.has(status)
  );
}

function blockedStatusMessage(status?: BillingSubscriptionStatus): string {
  switch (status) {
    case "past_due":
    case "unpaid":
      return "AI chat is paused until the payment issue is resolved.";
    case "canceled":
      return "AI chat is blocked because the subscription is canceled.";
    case "incomplete":
      return "AI chat is blocked until Checkout is completed.";
    case "incomplete_expired":
      return "AI chat is blocked because the Checkout session expired.";
    default:
      return "AI chat is blocked until billing is updated.";
  }
}

function getCancellationMessage(
  subscription?: BillingSubscription,
): string | null {
  if (!subscription?.cancel_at_period_end || !subscription.current_period_end) {
    return null;
  }

  return `Access continues until ${formatBillingDate(subscription.current_period_end)}.`;
}

function formatBillingDate(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
  }).format(date);
}
