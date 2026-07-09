import { createFileRoute } from "@tanstack/react-router";
import { TrainingProfilePage } from "@/features/training-profile/pages/training-profile-page";

export const Route = createFileRoute("/_layout/settings/training-profile")({
  component: TrainingProfileRoute,
});

function TrainingProfileRoute() {
  const { user } = Route.useRouteContext();

  if (!user) {
    return (
      <main className="mx-auto max-w-3xl px-4 py-6">
        <h1 className="text-2xl font-semibold tracking-tight">
          Training Profile
        </h1>
        <p className="mt-2 text-sm text-muted-foreground">
          Sign in to manage your training profile.
        </p>
      </main>
    );
  }

  return <TrainingProfilePage />;
}
