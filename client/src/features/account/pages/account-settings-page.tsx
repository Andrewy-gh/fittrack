import { useState } from "react";
import type { CurrentInternalUser, CurrentUser } from "@stackframe/react";
import { Link, useNavigate } from "@tanstack/react-router";
import { CreditCard, Dumbbell, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { deleteAccount } from "@/features/account/api/account";
import { clearCurrentDeviceAccountState } from "@/features/account/utils/current-device-state";
import {
  createBillingCustomerPortalSession,
  redirectToBillingPortal,
} from "@/features/chat/api/billing";

type AccountSettingsPageProps = {
  user: CurrentUser | CurrentInternalUser;
};

export const accountDeletionBillingCopy =
  "Deleting your account cancels your AI chat subscription so it will not renew. Your current billing period may already have been charged. If you recently paid and want FitTrack to review a refund, contact support@fittrack.andrewy.me.";

export function AccountSettingsPage({ user }: AccountSettingsPageProps) {
  const navigate = useNavigate();
  const [isDeletionConfirmed, setIsDeletionConfirmed] = useState(false);
  const [isOpeningBilling, setIsOpeningBilling] = useState(false);
  const [billingError, setBillingError] = useState<string | null>(null);
  const [isDeletingAccount, setIsDeletingAccount] = useState(false);
  const [isAccountDeleted, setIsAccountDeleted] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  async function handleManageBilling() {
    setIsOpeningBilling(true);
    setBillingError(null);

    try {
      const session = await createBillingCustomerPortalSession("settings");
      redirectToBillingPortal(session.url);
    } catch {
      setBillingError("Could not open billing. Please try again.");
    } finally {
      setIsOpeningBilling(false);
    }
  }

  async function handleDeleteAccount() {
    if (!isDeletionConfirmed) {
      return;
    }

    setIsDeletingAccount(true);
    setDeleteError(null);

    try {
      await deleteAccount();
    } catch {
      setDeleteError("Could not delete your account. Please try again.");
      setIsDeletingAccount(false);
      return;
    }

    setIsAccountDeleted(true);

    let didFinishCurrentDeviceExit = true;

    try {
      clearCurrentDeviceAccountState(user.id);
    } catch {
      didFinishCurrentDeviceExit = false;
    }

    try {
      await user.signOut();
    } catch {
      didFinishCurrentDeviceExit = false;
    }

    if (!didFinishCurrentDeviceExit) {
      setDeleteError(
        "Your account was deleted, but FitTrack could not finish signing you out on this device. Refresh the page if you still see account details.",
      );
      setIsDeletingAccount(false);
      return;
    }

    try {
      await navigate({ to: "/" });
    } catch {
      setDeleteError(
        "Your account was deleted, but FitTrack could not finish signing you out on this device. Refresh the page if you still see account details.",
      );
    }

    setIsDeletingAccount(false);
  }

  return (
    <main className="mx-auto flex w-full max-w-3xl flex-col gap-6 px-4 py-6">
      <header className="space-y-2">
        <h1 className="text-2xl font-semibold tracking-tight">
          Account settings
        </h1>
        <p className="text-sm text-muted-foreground">
          Manage billing separately from destructive account deletion.
        </p>
      </header>

      <section className="rounded-lg border bg-card p-4">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="space-y-2">
            <h2 className="text-lg font-semibold tracking-tight">
              Training Profile
            </h2>
            <p className="text-sm text-muted-foreground">
              Review and correct the durable workout preferences used by the AI
              coach.
            </p>
          </div>
          <Button
            asChild
            variant="outline"
          >
            <Link to="/settings/training-profile">
              <Dumbbell className="size-4" />
              Open profile
            </Link>
          </Button>
        </div>
      </section>

      <section className="rounded-lg border bg-card p-4">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="space-y-2">
            <h2 className="text-lg font-semibold tracking-tight">Billing</h2>
            <p className="text-sm text-muted-foreground">
              Open Stripe to manage your AI chat subscription, payment method,
              invoices, or normal subscription cancellation.
            </p>
          </div>
          <Button
            type="button"
            onClick={handleManageBilling}
            disabled={isOpeningBilling}
          >
            <CreditCard className="size-4" />
            {isOpeningBilling ? "Opening billing..." : "Manage billing"}
          </Button>
        </div>
        {billingError ? (
          <p className="mt-3 text-sm text-destructive">{billingError}</p>
        ) : null}
      </section>

      <section className="rounded-lg border border-destructive/30 bg-card p-4">
        <div className="space-y-4">
          <div className="space-y-2">
            <h2 className="text-lg font-semibold tracking-tight text-destructive">
              Delete account
            </h2>
            <p className="text-sm text-muted-foreground">
              This permanently deletes FitTrack app data for this signed-in
              user. If you want a copy of your data, contact
              privacy@fittrack.andrewy.me before deleting.
            </p>
            <p className="text-sm text-muted-foreground">
              {accountDeletionBillingCopy}
            </p>
          </div>

          <label className="flex items-start gap-3 text-sm">
            <input
              type="checkbox"
              className="mt-1"
              checked={isDeletionConfirmed}
              onChange={(event) =>
                setIsDeletionConfirmed(event.currentTarget.checked)
              }
            />
            <span>
              I understand this deletes my FitTrack app data and signs me out on
              this device.
            </span>
          </label>

          <Button
            type="button"
            variant="destructive"
            onClick={handleDeleteAccount}
            disabled={
              !isDeletionConfirmed || isDeletingAccount || isAccountDeleted
            }
          >
            <Trash2 className="size-4" />
            {isAccountDeleted
              ? "Account deleted"
              : isDeletingAccount
                ? "Deleting account..."
                : "Delete account"}
          </Button>
          {deleteError ? (
            <p className="text-sm text-destructive">{deleteError}</p>
          ) : null}
        </div>
      </section>
    </main>
  );
}
