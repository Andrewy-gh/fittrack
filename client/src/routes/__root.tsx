import { createRootRouteWithContext, Outlet } from '@tanstack/react-router';
import { QueryClient } from '@tanstack/react-query';
import { TanStackRouterDevtools } from '@tanstack/react-router-devtools';
import type { CurrentInternalUser, CurrentUser } from '@stackframe/react';
import { RouteError } from '@/components/route-error';

interface RouteContext {
  queryClient: QueryClient;
  user: CurrentUser | CurrentInternalUser | null;
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
