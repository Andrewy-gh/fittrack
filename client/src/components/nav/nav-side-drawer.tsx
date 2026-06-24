import type { CurrentInternalUser, CurrentUser } from "@stackframe/react";
import { Link, useRouterState } from "@tanstack/react-router";
import { Menu, X } from "lucide-react";
import { type ReactNode, useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerHeader,
  DrawerTitle,
  DrawerTrigger,
} from "@/components/ui/drawer";
import { AccountSlot } from "@/components/nav/account-slot";
import { isActivePath } from "@/components/nav/use-active-path";
import { navItems } from "@/components/nav/nav-items";
import { cn } from "@/lib/utils";

interface NavSideDrawerProps {
  user: CurrentUser | CurrentInternalUser | null;
  children?: ReactNode;
}

export function NavSideDrawer({ user, children }: NavSideDrawerProps) {
  const [open, setOpen] = useState(false);
  const pathname = useRouterState({
    select: (state) => state.location.pathname,
  });

  return (
    <Drawer
      open={open}
      onOpenChange={setOpen}
      direction="left"
    >
      <DrawerTrigger asChild>
        {children ?? (
          <Button
            variant="ghost"
            size="icon"
            aria-label="Open navigation menu"
          >
            <Menu className="h-5 w-5" />
          </Button>
        )}
      </DrawerTrigger>

      <DrawerContent className="h-full max-w-[280px] p-0">
        <DrawerHeader className="border-b px-4 py-3 text-left">
          <div className="flex items-center justify-between gap-3">
            <DrawerTitle className="text-base font-semibold">
              FitTrack
            </DrawerTitle>
            <DrawerClose asChild>
              <Button
                variant="ghost"
                size="icon"
                aria-label="Close navigation menu"
              >
                <X className="h-5 w-5" />
              </Button>
            </DrawerClose>
          </div>
        </DrawerHeader>

        <nav
          aria-label="Main navigation"
          className="flex flex-col p-3"
        >
          {navItems.map(({ to, label, icon: Icon }) => {
            const active = isActivePath(pathname, to);

            return (
              <DrawerClose
                asChild
                key={to}
              >
                <Link
                  to={to}
                  aria-current={active ? "page" : undefined}
                  className={cn(
                    "flex items-center gap-3 rounded-lg px-3 py-3 text-sm font-medium transition-colors",
                    active
                      ? "bg-primary text-primary-foreground"
                      : "text-foreground hover:bg-muted",
                  )}
                >
                  <Icon className="h-4 w-4" />
                  <span>{label}</span>
                </Link>
              </DrawerClose>
            );
          })}
        </nav>

        <div className="mt-auto flex items-center justify-between border-t p-4">
          <span className="text-sm font-medium text-muted-foreground">
            Account
          </span>
          <AccountSlot user={user} />
        </div>
      </DrawerContent>
    </Drawer>
  );
}
