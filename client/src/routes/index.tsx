import type { CurrentInternalUser, CurrentUser } from "@stackframe/react";
import { createFileRoute, redirect } from "@tanstack/react-router";
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
    (navigator as Navigator & { standalone?: boolean }).standalone === true;

  return standaloneDisplayMode || iosStandalone;
}

export function redirectPwaSignedInUser({
  context,
}: {
  context: { user: CurrentUser | CurrentInternalUser | null };
}) {
  if (isStandalone() && context.user) {
    throw redirect({ to: "/workouts" });
  }
}

function RouteComponent() {
  const { user } = Route.useRouteContext();

  return <HomePage user={user} />;
}
