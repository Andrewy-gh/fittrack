import {
  createFileRoute,
  Outlet,
  useRouterState,
} from "@tanstack/react-router";
import { AppShell } from "@/components/nav/app-shell";
import { PwaInstallPrompt } from "@/components/pwa-install-prompt";
import { useDisplayMode } from "@/hooks/use-display-mode";

export const Route = createFileRoute("/_layout")({
  component: LayoutComponent,
});

export function LayoutComponent() {
  const { user } = Route.useRouteContext();
  const displayMode = useDisplayMode();
  const pathname = useRouterState({
    select: (state) => state.location.pathname,
  });

  return (
    <div
      className={
        displayMode === "pwa"
          ? "pb-[calc(5rem+env(safe-area-inset-bottom))]"
          : undefined
      }
    >
      <AppShell user={user} />
      <PwaInstallPrompt
        displayMode={displayMode}
        pathname={pathname}
        user={user}
      />
      <Outlet />
    </div>
  );
}
