import { Link } from '@tanstack/react-router';
import { CustomUserButton } from './custom-user-button';
import { MobileBottomNav } from './mobile-bottom-nav';
import { MobileNavDrawer } from './mobile-nav-drawer';

export function Header() {
  return (
    <>
      <header
        className="flex items-center justify-between gap-2 border-b p-2"
        data-app-header
      >
        <div className="flex items-center gap-2">
          <MobileNavDrawer includeChat />

          <nav className="hidden md:flex md:flex-row">
            <div className="px-2 font-bold">
              <Link to="/">Home</Link>
            </div>
            <div className="px-2 font-bold">
              <Link to="/workouts">Workouts</Link>
            </div>
            <div className="px-2 font-bold">
              <Link to="/exercises">Exercises</Link>
            </div>
            <div className="px-2 font-bold">
              <Link to="/analytics">Analytics</Link>
            </div>
            <div className="px-2 font-bold">
              <Link to="/chat">AI Chat</Link>
            </div>
          </nav>
        </div>

        <CustomUserButton />
      </header>

      <MobileBottomNav includeChat />
    </>
  );
}
