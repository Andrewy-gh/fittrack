import { createFileRoute, redirect } from "@tanstack/react-router";
import type { ApplicationUser } from "@/lib/application-user";
import { HomePage } from "@/features/home/pages/home-page";

export const Route = createFileRoute("/")({
  beforeLoad: redirectPwaSignedInUser,
  component: RouteComponent,
});

export function isStandalone() {
  const standaloneDisplayMode =
    typeof window !== "undefined" &&
    window.matchMedia("(display-mode: standalone)").matches;
  const iosStandalone =
    typeof navigator !== "undefined" &&
    "standalone" in navigator &&
    navigator.standalone === true;

  return standaloneDisplayMode || iosStandalone;
}

export function redirectPwaSignedInUser({
  context,
}: {
  context: { user: ApplicationUser | null };
}) {
  if (isStandalone() && context.user) {
    throw redirect({ to: "/workouts" });
  }
}

function RouteComponent() {
  const { user } = Route.useRouteContext();

  return <HomePage user={user} />;
}
