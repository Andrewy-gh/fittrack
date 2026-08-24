import { createFileRoute } from "@tanstack/react-router";
import { PrivacyPage } from "@/features/privacy/privacy-page";

export const Route = createFileRoute("/privacy")({
  component: RouteComponent,
});

function RouteComponent() {
  const { user } = Route.useRouteContext();

  return <PrivacyPage user={user} />;
}
