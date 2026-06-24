import { useRouterState } from "@tanstack/react-router";

export function isActivePath(pathname: string, to: string) {
  return pathname === to || (to !== "/" && pathname.startsWith(`${to}/`));
}

export function useActivePath(to: string) {
  const pathname = useRouterState({
    select: (state) => state.location.pathname,
  });

  return isActivePath(pathname, to);
}
