import { useEffect, useState } from 'react';
import { Menu } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { MobileNavDrawer } from './mobile-nav-drawer';
import { CustomUserButton } from './custom-user-button';

interface MobileBottomNavProps {
  includeChat?: boolean;
}

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

  return (
    <nav
      className="fixed inset-x-0 bottom-0 z-50 border-t bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/85 md:hidden"
      style={{ paddingBottom: 'env(safe-area-inset-bottom)' }}
      aria-label="Mobile navigation"
      data-mobile-bottom-nav
    >
      <div className="mx-auto flex max-w-lg items-center justify-between px-4 py-3">
        <MobileNavDrawer includeChat={includeChat}>
          <Button variant="ghost" size="icon" aria-label="Open navigation menu">
            <Menu className="h-5 w-5" />
          </Button>
        </MobileNavDrawer>

        <CustomUserButton />
      </div>
    </nav>
  );
}
