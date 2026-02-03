import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { LogIn, Moon, Sun, User, UserPlus } from 'lucide-react';
import { useStackApp } from '@stackframe/react';
import { useTheme } from '@/components/theme-provider';
import React from 'react';

const Typography: React.FC<{ children: React.ReactNode; className?: string }> =
  ({ children, className = '' }) => (
    <span className={`text-sm ${className}`}>{children}</span>
  );

export function GuestUserButton() {
  const { theme, setTheme } = useTheme();
  const app = useStackApp();

  const toggleTheme = () => {
    setTheme(theme === 'light' ? 'dark' : 'light');
  };

  const iconProps = { size: 16 };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger className="outline-none">
        <Avatar className="h-8.5 w-8.5">
          <AvatarFallback>
            <User {...iconProps} />
          </AvatarFallback>
        </Avatar>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuLabel>
          <div className="flex items-center gap-3">
            <Avatar className="h-8 w-8">
              <AvatarFallback>
                <User {...iconProps} />
              </AvatarFallback>
            </Avatar>
            <div className="flex flex-col">
              <Typography>Not signed in</Typography>
            </div>
          </div>
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={toggleTheme}>
          <div className="flex gap-2 items-center">
            {theme === 'light' ? <Moon {...iconProps} /> : <Sun {...iconProps} />}
            <Typography>{theme === 'light' ? 'Dark' : 'Light'}</Typography>
          </div>
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={() => app.redirectToSignIn()}>
          <div className="flex gap-2 items-center">
            <LogIn {...iconProps} />
            <Typography>Sign in</Typography>
          </div>
        </DropdownMenuItem>
        <DropdownMenuItem onClick={() => app.redirectToSignUp()}>
          <div className="flex gap-2 items-center">
            <UserPlus {...iconProps} />
            <Typography>Sign up</Typography>
          </div>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
