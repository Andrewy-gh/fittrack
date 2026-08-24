import type { CurrentInternalUser, CurrentUser } from "@stackframe/react";
import { Outlet, useRouterState } from "@tanstack/react-router";
import { PwaInstallPrompt } from "@/components/pwa-install-prompt";
import { ChatDraftProvider } from "@/features/chat/utils/chat-draft-context";
import { useDisplayMode } from "@/hooks/use-display-mode";
import { AppShell } from "./app-shell";

type LayoutComponentProps = {
  readonly user: CurrentUser | CurrentInternalUser | null;
};

/** Renders the shared authenticated application layout. */
export function LayoutComponent({ user }: LayoutComponentProps) {
  const displayMode = useDisplayMode();
  const pathname = useRouterState({
    select: (state) => state.location.pathname,
  });

  return (
    <div
      className={
        displayMode === "pwa"
          ? "pt-[env(safe-area-inset-top)] pb-[calc(5rem+env(safe-area-inset-bottom))]"
          : undefined
      }
    >
      <AppShell user={user} />
      <PwaInstallPrompt
        displayMode={displayMode}
        pathname={pathname}
        user={user}
      />
      <ChatDraftProvider key={user?.id ?? "signed-out"}>
        <Outlet />
      </ChatDraftProvider>
    </div>
  );
}
