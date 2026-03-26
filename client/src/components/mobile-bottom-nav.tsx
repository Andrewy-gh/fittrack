import { Link, useRouterState } from '@tanstack/react-router';
import { useEffect, useState } from 'react';
import { Home, Dumbbell, Activity, ChartColumn, Bot } from 'lucide-react';
import { cn } from '@/lib/utils';

interface MobileBottomNavProps {
  includeChat?: boolean;
}

const baseLinks = [
  { to: '/', label: 'Home', icon: Home },
  { to: '/workouts', label: 'Workouts', icon: Dumbbell },
  { to: '/exercises', label: 'Exercises', icon: Activity },
  { to: '/analytics', label: 'Analytics', icon: ChartColumn },
] as const;

function shouldUseBottomNav() {
  if (typeof window === 'undefined') return false;

  const standalone =
    window.matchMedia('(display-mode: standalone)').matches ||
    (navigator as Navigator & { standalone?: boolean }).standalone === true;
  const coarsePointer = window.matchMedia('(pointer: coarse)').matches;
  const touchPoints = navigator.maxTouchPoints > 0;
  const mobileViewport = window.matchMedia('(max-width: 1024px)').matches;

  return (coarsePointer || touchPoints) && (standalone || mobileViewport);
}

export function MobileBottomNav({ includeChat = false }: MobileBottomNavProps) {
  const pathname = useRouterState({
    select: (state) => state.location.pathname,
  });
  const [enabled, setEnabled] = useState(false);

  useEffect(() => {
    const update = () => setEnabled(shouldUseBottomNav());
    update();

    const queries = [
      window.matchMedia('(display-mode: standalone)'),
      window.matchMedia('(pointer: coarse)'),
      window.matchMedia('(max-width: 1024px)'),
    ];

    queries.forEach((query) => query.addEventListener('change', update));
    window.addEventListener('resize', update);

    return () => {
      queries.forEach((query) => query.removeEventListener('change', update));
      window.removeEventListener('resize', update);
    };
  }, []);

  if (!enabled) {
    return null;
  }

  const links = includeChat
    ? [...baseLinks, { to: '/chat', label: 'Chat', icon: Bot }]
    : baseLinks;

  return (
    <nav
      className="fixed inset-x-0 bottom-0 z-50 border-t bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/85 md:hidden"
      style={{ paddingBottom: 'env(safe-area-inset-bottom)' }}
      aria-label="Mobile navigation"
      data-mobile-bottom-nav
    >
      <div className="mx-auto grid max-w-lg grid-cols-5 px-2 py-2">
        {links.map(({ to, label, icon: Icon }) => {
          const active = to === '/'
            ? pathname === '/'
            : pathname === to || pathname.startsWith(`${to}/`);

          return (
            <Link
              key={to}
              to={to}
              className={cn(
                'flex min-w-0 flex-col items-center justify-center gap-1 rounded-md px-1 py-2 text-[11px] font-medium transition-colors',
                active ? 'text-primary' : 'text-muted-foreground hover:text-foreground'
              )}
            >
              <Icon className="h-4 w-4" />
              <span className="truncate">{label}</span>
            </Link>
          );
        })}
      </div>
    </nav>
  );
}
