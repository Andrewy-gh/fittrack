import { createFileRoute } from "@tanstack/react-router";
import { AccountSettingsPage } from "@/features/account/pages/account-settings-page";

export const Route = createFileRoute("/_layout/settings")({
  component: SettingsRoute,
});

function SettingsRoute() {
  const { user } = Route.useRouteContext();

  if (!user) {
    return (
      <main className="mx-auto max-w-3xl px-4 py-6">
        <h1 className="text-2xl font-semibold tracking-tight">
          Account settings
        </h1>
        <p className="mt-2 text-sm text-muted-foreground">
          Sign in to manage your account settings.
        </p>
      </main>
    );
  }

  return <AccountSettingsPage user={user} />;
}
