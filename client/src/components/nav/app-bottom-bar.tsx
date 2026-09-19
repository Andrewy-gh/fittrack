import { useEffect, useState } from "react";
import type { ApplicationUser } from "@/lib/application-user";
import { Link, useRouterState } from "@tanstack/react-router";
import { AccountSlot } from "@/components/nav/account-slot";
import { isActivePath } from "@/components/nav/use-active-path";
import { navItems } from "@/components/nav/nav-items";
import { cn } from "@/lib/utils";

interface AppBottomBarProps {
  user: ApplicationUser | null;
}

export function AppBottomBar({ user }: AppBottomBarProps) {
  const pathname = useRouterState({
    select: (state) => state.location.pathname,
  });

  const [compact, setCompact] = useState(false);

  useEffect(() => {
    // Hysteresis avoids flickering near the top and during iOS overscroll.
    const update = () => {
      const y = Math.max(0, window.scrollY);
      setCompact((current) => (current ? y > 8 : y > 48));
    };
    update();
    window.addEventListener("scroll", update, { passive: true });
    return () => window.removeEventListener("scroll", update);
  }, [pathname]);

  return (
    <nav
      aria-label="PWA navigation"
      className="pointer-events-none fixed inset-x-0 bottom-0 z-50 px-4"
      data-app-bottom-bar
      style={{ paddingBottom: "calc(env(safe-area-inset-bottom) + 0.75rem)" }}
    >
      <div
        className={cn(
          "pointer-events-auto mx-auto flex max-w-md items-center gap-1 rounded-full border bg-background/95 px-1 shadow-lg backdrop-blur supports-[backdrop-filter]:bg-background/85 transition-[width,padding] duration-300 ease-out motion-reduce:transition-none",
          compact ? "w-[84%] py-1" : "w-full py-1.5",
        )}
      >
        {navItems.map(({ to, label, icon: Icon, search }) => {
          const active = isActivePath(pathname, to);

          return (
            <Link
              key={to}
              to={to}
              search={search}
              aria-current={active ? "page" : undefined}
              aria-label={label}
              title={label}
              className={cn(
                "flex min-h-11 min-w-11 flex-1 items-center justify-center rounded-full transition-[height,background-color,color] duration-300 ease-out motion-reduce:transition-none focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
                compact ? "h-11" : "h-12",
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

        <div
          className={cn(
            "flex min-h-11 min-w-11 flex-1 items-center justify-center transition-[height] duration-300 ease-out motion-reduce:transition-none [&_button]:min-h-11 [&_button]:min-w-11 [&_button]:flex [&_button]:items-center [&_button]:justify-center",
            compact ? "h-11" : "h-12",
          )}
        >
          <AccountSlot user={user} />
        </div>
      </div>
    </nav>
  );
}
