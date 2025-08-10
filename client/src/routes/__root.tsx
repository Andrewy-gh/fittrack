import { createRootRouteWithContext, Outlet } from '@tanstack/react-router';
import { QueryClient } from '@tanstack/react-query';
import { TanStackRouterDevtools } from '@tanstack/react-router-devtools';
import Header from '../components/Header';
import type { User } from '@/lib/api/auth';

interface RouteContext {
  user: User;
  queryClient: QueryClient;
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
