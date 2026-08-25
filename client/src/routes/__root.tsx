import { createRootRouteWithContext, Outlet } from "@tanstack/react-router";
import { QueryClient } from "@tanstack/react-query";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";
import { RouteError } from "@/components/route-error";
import type { ApplicationUser } from "@/lib/application-user";

interface RouteContext {
  queryClient: QueryClient;
  user: ApplicationUser | null;
}

export const Route = createRootRouteWithContext<RouteContext>()({
  component: () => (
    <>
      <Outlet />
      <TanStackRouterDevtools />
    </>
  ),
  errorComponent: RouteError,
});
