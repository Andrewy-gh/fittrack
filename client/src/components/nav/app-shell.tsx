import type { CurrentInternalUser, CurrentUser } from "@stackframe/react";
import { useDisplayMode } from "@/hooks/use-display-mode";
import { AppBottomBar } from "@/components/nav/app-bottom-bar";
import { AppTopBar } from "@/components/nav/app-top-bar";

interface AppShellProps {
  user: CurrentUser | CurrentInternalUser | null;
}

export function AppShell({ user }: AppShellProps) {
  const displayMode = useDisplayMode();

  return displayMode === "pwa" ? (
    <AppBottomBar user={user} />
  ) : (
    <AppTopBar user={user} />
  );
}
