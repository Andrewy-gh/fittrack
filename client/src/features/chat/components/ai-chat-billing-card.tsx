import {
  AlertCircle,
  CheckCircle2,
  CreditCard,
  RefreshCw,
  Sparkles,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import type {
  BillingStatusResponse,
  BillingSubscription,
  BillingSubscriptionStatus,
} from "@/features/chat/api/billing";
import { cn } from "@/lib/utils";

type AIChatBillingCardProps = {
  status?: BillingStatusResponse;
  accessState: AIChatBillingCardAccessState;
  isLoading?: boolean;
  isError?: boolean;
};

type AIChatBillingActionsProps = {
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
  | "payment-confirming"
  | "activating"
  | "blocked"
  | "billing-error"
  | "checkout-activation-error";

const checkoutBlockedStatuses = new Set<BillingSubscriptionStatus>([
  "canceled",
  "incomplete",
  "incomplete_expired",
]);

const billingPortalStatuses = new Set<BillingSubscriptionStatus>([
  "past_due",
  "unpaid",
]);

const planManagementStatuses = new Set<BillingSubscriptionStatus>([
  "trialing",
  "active",
  ...billingPortalStatuses,
]);

export function AIChatBillingCard({
  status,
  accessState,
  isLoading = false,
  isError = false,
}: AIChatBillingCardProps) {
  const subscription = status?.subscription;
  const isTrialing = status?.has_access && subscription?.status === "trialing";
  const isActive = status?.has_access && subscription?.status === "active";
  const isBlocked =
    !status?.has_access &&
    subscription?.status !== undefined &&
    isBlockedStatus(subscription.status);
  const hasChatAccess = accessState === "ready";

  return (
    <section className="space-y-2 px-4 py-4">
      <div className="space-y-2">
        <div className="flex flex-wrap items-center gap-2">
          <h2 className="text-lg font-semibold tracking-tight">AI Chat</h2>
          <BillingStatusBadge
            isLoading={isLoading}
            isError={isError}
            status={status}
            accessState={accessState}
          />
        </div>
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
    </section>
  );
}

export function AIChatBillingActions({
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
}: AIChatBillingActionsProps) {
  const billingAction = getBillingAction(status, accessState);
  const shouldShowBillingAction =
    !isLoading &&
    billingAction &&
    (!isError || billingAction === "refresh") &&
    (accessState !== "ready" || billingAction === "portal");
  const isActionLoading = getActionLoadingState({
    billingAction,
    isBillingPortalLoading,
    isCheckoutLoading,
    isRefreshingAccess,
  });

  if (!shouldShowBillingAction) {
    return null;
  }

  return (
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

  if (accessState === "payment-confirming") {
    return <Badge variant="secondary">Confirming</Badge>;
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
    if (accessState === "checkout-activation-error") {
      return (
        <p className="text-sm text-muted-foreground">
          Checkout finished, but we could not refresh AI chat access. Try
          refreshing access.
        </p>
      );
    }

    return (
      <p className="text-sm text-muted-foreground">
        Could not confirm billing.
      </p>
    );
  }

  if (!status) {
    return (
      <p className="text-sm text-muted-foreground">
        Start a trial to use AI chat.
      </p>
    );
  }

  const cancellationMessage = getCancellationMessage(status.subscription);

  if (accessState === "payment-confirming") {
    return (
      <p className="text-sm text-muted-foreground">
        Checking access after payment.
      </p>
    );
  }

  if (accessState === "activating") {
    return (
      <p className="text-sm text-muted-foreground">Finishing activation.</p>
    );
  }

  if (hasChatAccess && !isTrialing && !isActive) {
    if (billingPortalStatuses.has(status.subscription?.status ?? "active")) {
      return (
        <p className="text-sm text-muted-foreground">
          Update billing to keep chat available.
        </p>
      );
    }

    return null;
  }

  if (isTrialing) {
    if (status.trial_usage || cancellationMessage) {
      return (
        <div className="space-y-2">
          {status.trial_usage ? (
            <p className="text-sm font-medium">
              {status.trial_usage.used} of {status.trial_usage.limit} trial
              prompts used
            </p>
          ) : null}
          {cancellationMessage ? (
            <p className="text-sm text-muted-foreground">
              {cancellationMessage}
            </p>
          ) : null}
        </div>
      );
    }

    return null;
  }

  if (isActive) {
    if (cancellationMessage) {
      return (
        <p className="text-sm text-muted-foreground">{cancellationMessage}</p>
      );
    }

    return null;
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
      Start a trial to use AI chat. The trial includes 30 AI prompts.
    </p>
  );
}

type BillingAction = "checkout" | "portal" | "refresh";

function getBillingAction(
  status: BillingStatusResponse | undefined,
  accessState: AIChatBillingCardAccessState,
): BillingAction | null {
  if (
    accessState === "payment-confirming" ||
    accessState === "activating" ||
    accessState === "checkout-activation-error"
  ) {
    return "refresh";
  }

  if (accessState === "billing-error") {
    return null;
  }

  if (planManagementStatuses.has(status?.subscription?.status ?? "canceled")) {
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
    return "Manage plan";
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
  billingAction: BillingAction | null;
  isBillingPortalLoading: boolean;
  isCheckoutLoading: boolean;
  isRefreshingAccess: boolean;
}): boolean {
  if (!billingAction) {
    return false;
  }

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
  if (!subscription || !subscription.cancellation_scheduled) {
    return null;
  }

  if (!subscription.access_ends_at) {
    return "Access continues until the end of the current billing period.";
  }

  return `Access continues until ${formatBillingDate(subscription.access_ends_at)}.`;
}

export function isSubscriptionCancellationScheduled(
  subscription: BillingSubscription,
): boolean {
  return subscription.cancellation_scheduled;
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
