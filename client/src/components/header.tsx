import { Link } from '@tanstack/react-router';
import { CustomUserButton } from './custom-user-button';
import { MobileBottomNav } from './mobile-bottom-nav';

export function Header() {
  return (
    <>
      <header
        className="flex items-center justify-end gap-2 border-b p-2"
        data-app-header
      >
        <nav className="hidden md:flex md:flex-row md:mr-auto">
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

        <CustomUserButton />
      </header>

      <MobileBottomNav includeChat />
    </>
  );
}
