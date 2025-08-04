import { UserButton } from '@stackframe/react';
import { Link } from '@tanstack/react-router';
import { ModeToggle } from './mode-toggle';

export default function Header() {
  return (
    <header className="flex flex-wrap justify-between gap-2 border-b p-2">
      <nav className="flex flex-wrap flex-row">
        <div className="px-2 font-bold">
          <Link to="/">Home</Link>
        </div>
        <div className="px-2 font-bold">
          <Link to="/workouts">Workouts</Link>
        </div>
        <div className="px-2 font-bold">
          <Link to="/exercises">Exercises</Link>
        </div>
      </nav>
      <div className="flex flex-row gap-2">
        <ModeToggle />
        <UserButton />
      </div>
    </header>
  );
}
