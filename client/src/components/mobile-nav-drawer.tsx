import { Link, useRouterState } from '@tanstack/react-router';
import { Menu, X, Home, Dumbbell, Activity, ChartColumn, Bot } from 'lucide-react';
import { useState } from 'react';
import { Button } from '@/components/ui/button';
import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerHeader,
  DrawerTitle,
  DrawerTrigger,
} from '@/components/ui/drawer';
import { cn } from '@/lib/utils';

interface MobileNavDrawerProps {
  includeChat?: boolean;
}

const baseLinks = [
  { to: '/', label: 'Home', icon: Home },
  { to: '/workouts', label: 'Workouts', icon: Dumbbell },
  { to: '/exercises', label: 'Exercises', icon: Activity },
  { to: '/analytics', label: 'Analytics', icon: ChartColumn },
] as const;

export function MobileNavDrawer({ includeChat = false }: MobileNavDrawerProps) {
  const [open, setOpen] = useState(false);
  const pathname = useRouterState({
    select: (state) => state.location.pathname,
  });

  const links = includeChat
    ? [...baseLinks, { to: '/chat', label: 'AI Chat', icon: Bot }]
    : baseLinks;

  return (
    <Drawer open={open} onOpenChange={setOpen} direction="left">
      <DrawerTrigger asChild>
        <Button variant="ghost" size="icon" aria-label="Open navigation menu">
          <Menu className="h-5 w-5" />
        </Button>
      </DrawerTrigger>

      <DrawerContent className="h-full max-w-[280px] p-0">
        <DrawerHeader className="border-b px-4 py-3 text-left">
          <div className="flex items-center justify-between gap-3">
            <DrawerTitle className="text-base font-semibold">FitTrack</DrawerTitle>
            <DrawerClose asChild>
              <Button variant="ghost" size="icon" aria-label="Close navigation menu">
                <X className="h-5 w-5" />
              </Button>
            </DrawerClose>
          </div>
        </DrawerHeader>

        <nav className="flex flex-col p-3">
          {links.map(({ to, label, icon: Icon }) => {
            const active = to === '/'
              ? pathname === '/'
              : pathname === to || pathname.startsWith(`${to}/`);

            return (
              <DrawerClose asChild key={to}>
                <Link
                  to={to}
                  className={cn(
                    'flex items-center gap-3 rounded-lg px-3 py-3 text-sm font-medium transition-colors',
                    active
                      ? 'bg-primary text-primary-foreground'
                      : 'text-foreground hover:bg-muted'
                  )}
                >
                  <Icon className="h-4 w-4" />
                  <span>{label}</span>
                </Link>
              </DrawerClose>
            );
          })}
        </nav>
      </DrawerContent>
    </Drawer>
  );
}
