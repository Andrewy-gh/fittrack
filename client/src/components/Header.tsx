import { Link } from '@tanstack/react-router';
import { CustomUserButton } from './custom-user-button';

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
      <CustomUserButton />
    </header>
  );
}
