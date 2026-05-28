import { createFileRoute } from "@tanstack/react-router";
import { HomePage } from "@/features/home/pages/home-page";

export const Route = createFileRoute("/")({
  component: RouteComponent,
});

function RouteComponent() {
  const { user } = Route.useRouteContext();

  return <HomePage user={user} />;
}
