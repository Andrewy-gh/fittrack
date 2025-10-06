import { Link } from '@tanstack/react-router';
import { Sun, Moon } from 'lucide-react';
import { useTheme } from '@/components/theme-provider';
import { Button } from '@/components/ui/button';

export function DemoHeader() {
  const { theme, setTheme } = useTheme();

  const toggleTheme = () => {
    setTheme(theme === 'light' ? 'dark' : 'light');
  };

  return (
    <header className="flex flex-wrap justify-between gap-2 border-b p-2">
      <nav className="flex flex-wrap flex-row">
        <div className="px-2 font-bold">
          <Link to="/">Home</Link>
        </div>
        <div className="px-2 font-bold">
          <Link to="/demo/workouts">Workouts</Link>
        </div>
        <div className="px-2 font-bold">
          <Link to="/demo/exercises">Exercises</Link>
        </div>
      </nav>
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
  );
}
