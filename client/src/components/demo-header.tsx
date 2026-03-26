import { Link } from '@tanstack/react-router';
import { Sun, Moon } from 'lucide-react';
import { useTheme } from '@/components/theme-provider';
import { Button } from '@/components/ui/button';
import { MobileBottomNav } from './mobile-bottom-nav';

export function DemoHeader() {
  const { theme, setTheme } = useTheme();

  const toggleTheme = () => {
    setTheme(theme === 'light' ? 'dark' : 'light');
  };

  return (
    <>
      <header
        className="flex items-center justify-between gap-2 border-b p-2"
        data-app-header
      >
        <div className="flex items-center gap-2">
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
        </nav>
      </div>

        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="icon"
            onClick={toggleTheme}
            className="h-8 w-8"
          >
            {theme === 'light' ? <Moon size={16} /> : <Sun size={16} />}
          </Button>
        </div>
      </header>

      <MobileBottomNav />
    </>
  );
}
