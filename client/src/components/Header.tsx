import { Link } from '@tanstack/react-router'

export default function Header() {
  return (
    <header className="flex justify-between gap-2 border-b border-neutral-700 bg-neutral-800 p-2 text-white">
      <nav className="flex flex-row">
        <div className="px-2 font-bold">
          <Link to="/">Home</Link>
        </div>
        <div className="px-2 font-bold">
          <Link to="/workouts">Workouts</Link>
        </div>
      </nav>
    </header>
  );
}
