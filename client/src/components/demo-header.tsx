import { Link } from "@tanstack/react-router";
import { GuestUserButton } from "./guest-user-button";
import { MobileBottomNav } from "./mobile-bottom-nav";

export function DemoHeader() {
  return (
    <>
      <header
        className="hidden items-center justify-end gap-2 border-b p-2 md:flex"
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
        </nav>

        <GuestUserButton />
      </header>

      <MobileBottomNav
        includeChat
        isAuthenticated={false}
      />
    </>
  );
}
