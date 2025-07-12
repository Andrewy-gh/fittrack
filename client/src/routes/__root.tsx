import { Outlet, createRootRouteWithContext } from '@tanstack/react-router';
import { TanStackRouterDevtools } from '@tanstack/react-router-devtools';
import type { CurrentUser, CurrentInternalUser } from '@stackframe/react';
import Header from '../components/Header';

interface RouteContext {
  user: CurrentUser | CurrentInternalUser | null;
}

export const Route = createRootRouteWithContext<RouteContext>()({
  component: () => (
    <>
      <Header />
      <Outlet />
      <TanStackRouterDevtools />
    </>
  ),
});
