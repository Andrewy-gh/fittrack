import type { CurrentInternalUser, CurrentUser } from "@stackframe/react";
import { Link, useRouterState } from "@tanstack/react-router";
import { AccountSlot } from "@/components/nav/account-slot";
import { isActivePath } from "@/components/nav/use-active-path";
import { NavSideDrawer } from "@/components/nav/nav-side-drawer";
import { navItems } from "@/components/nav/nav-items";
import { cn } from "@/lib/utils";

interface AppTopBarProps {
  user: CurrentUser | CurrentInternalUser | null;
}

export function AppTopBar({ user }: AppTopBarProps) {
  const pathname = useRouterState({
    select: (state) => state.location.pathname,
  });

  return (
    <header
      className="flex items-center gap-3 border-b bg-background/95 px-4 py-2 backdrop-blur supports-[backdrop-filter]:bg-background/85"
      data-app-header
    >
      <Link
        to="/"
        className="text-base font-bold"
      >
        FitTrack
      </Link>

      <nav
        aria-label="Primary navigation"
        className="hidden items-center gap-1 md:flex"
      >
        {navItems.map(({ to, label, search }) => {
          const active = isActivePath(pathname, to);

          return (
            <Link
              key={to}
              to={to}
              search={search}
              aria-current={active ? "page" : undefined}
              className={cn(
                "rounded-md px-3 py-2 text-sm font-medium transition-colors",
                active
                  ? "bg-primary text-primary-foreground"
                  : "text-muted-foreground hover:bg-muted hover:text-foreground",
              )}
            >
              {label}
            </Link>
          );
        })}
      </nav>

      <div className="ml-auto flex items-center gap-2">
        <div className="md:hidden">
          <NavSideDrawer user={user} />
        </div>
        <AccountSlot user={user} />
      </div>
    </header>
  );
}
