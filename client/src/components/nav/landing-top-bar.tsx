import type { CurrentInternalUser, CurrentUser } from "@stackframe/react";
import { Link } from "@tanstack/react-router";
import { AccountSlot } from "@/components/nav/account-slot";
import { navItems } from "@/components/nav/nav-items";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

interface LandingTopBarProps {
  user: CurrentUser | CurrentInternalUser | null;
}

export function LandingTopBar({ user }: LandingTopBarProps) {
  const menuLabel = user ? "Open ▾" : "Try ▾";

  return (
    <nav
      aria-label="Landing navigation"
      className="border-b border-border bg-background/90 backdrop-blur-sm"
    >
      <div className="flex items-center justify-between gap-3 px-2 py-4">
        <Link
          to="/"
          className="flex items-center gap-2"
        >
          <img
            src="/favicon.svg"
            alt=""
            aria-hidden="true"
            className="h-6 w-6 rounded-sm"
          />
          <span className="text-xl font-bold tracking-wide text-foreground">
            FITTRACK
          </span>
        </Link>

        <div className="flex items-center gap-2">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                type="button"
                variant="outline"
                className="min-w-20"
              >
                {menuLabel}
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              {navItems.map(({ to, label, icon: Icon, search }) => (
                <DropdownMenuItem
                  asChild
                  key={to}
                >
                  <Link
                    to={to}
                    search={search}
                    preload={false}
                    className="flex items-center gap-2"
                  >
                    <Icon className="h-4 w-4" />
                    <span>{label}</span>
                  </Link>
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>

          <AccountSlot user={user} />
        </div>
      </div>
    </nav>
  );
}
