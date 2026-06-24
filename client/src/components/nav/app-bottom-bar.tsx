import type { CurrentInternalUser, CurrentUser } from "@stackframe/react";
import { Link, useRouterState } from "@tanstack/react-router";
import { AccountSlot } from "@/components/nav/account-slot";
import { isActivePath } from "@/components/nav/use-active-path";
import { navItems } from "@/components/nav/nav-items";
import { cn } from "@/lib/utils";

interface AppBottomBarProps {
  user: CurrentUser | CurrentInternalUser | null;
}

export function AppBottomBar({ user }: AppBottomBarProps) {
  const pathname = useRouterState({
    select: (state) => state.location.pathname,
  });

  return (
    <nav
      aria-label="PWA navigation"
      className="pointer-events-none fixed inset-x-0 bottom-0 z-50 px-4"
      data-app-bottom-bar
      style={{ paddingBottom: "calc(env(safe-area-inset-bottom) + 0.75rem)" }}
    >
      <div className="pointer-events-auto mx-auto flex max-w-md items-center justify-between gap-1 rounded-full border bg-background/95 px-2 py-2 shadow-lg backdrop-blur supports-[backdrop-filter]:bg-background/85">
        {navItems.map(({ to, label, icon: Icon }) => {
          const active = isActivePath(pathname, to);

          return (
            <Link
              key={to}
              to={to}
              aria-current={active ? "page" : undefined}
              aria-label={label}
              title={label}
              className={cn(
                "flex size-11 items-center justify-center rounded-full transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
                active
                  ? "bg-primary text-primary-foreground"
                  : "text-muted-foreground hover:bg-muted hover:text-foreground",
              )}
            >
              <Icon className="h-5 w-5" />
              <span className="sr-only">{label}</span>
            </Link>
          );
        })}

        <div className="flex size-11 items-center justify-center">
          <AccountSlot user={user} />
        </div>
      </div>
    </nav>
  );
}
