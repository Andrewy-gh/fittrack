import { createFileRoute } from "@tanstack/react-router";
import { LayoutComponent } from "@/components/nav/layout-component";

export const Route = createFileRoute("/_layout")({
  component: RouteComponent,
});

function RouteComponent() {
  const { user } = Route.useRouteContext();

  return <LayoutComponent user={user} />;
}
