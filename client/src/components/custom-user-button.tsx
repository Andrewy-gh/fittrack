'use client';

import { 
  DropdownMenu, 
  DropdownMenuContent, 
  DropdownMenuItem, 
  DropdownMenuLabel, 
  DropdownMenuSeparator, 
  DropdownMenuTrigger 
} from '@/components/ui/dropdown-menu';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { LogOut, Sun, Moon } from 'lucide-react';
import React from 'react';
import { useUser } from '@stackframe/react';
import { useTheme } from '@/components/theme-provider';

// Simple Typography component since we don't have one from shadcn
const Typography: React.FC<{ children: React.ReactNode; className?: string; variant?: 'secondary' }> = ({ 
  children, 
  className = '',
  variant 
}) => {
  const baseClasses = 'text-sm';
  const variantClasses = variant === 'secondary' ? 'text-gray-500' : '';
  
  return (
    <span className={`${baseClasses} ${variantClasses} ${className}`}>
      {children}
    </span>
  );
};

export function CustomUserButton() {
  const user = useUser();
  const { theme, setTheme } = useTheme();

  // If there's no user, don't render anything
  if (!user) {
    return null;
  }

  const iconProps = { size: 16 };

  // Function to toggle theme
  const toggleTheme = () => {
    setTheme(theme === 'light' ? 'dark' : 'light');
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger className="outline-none">
        {/* Button shows only the avatar */}
        <Avatar className="h-8.5 w-8.5">
          {user.profileImageUrl && (
            <AvatarImage src={user.profileImageUrl} alt={user.displayName || 'User'} />
          )}
          <AvatarFallback>
            {user.displayName
              ?.split(' ')
              .map(name => name.charAt(0))
              .join('')
              .toUpperCase() || 'U'}
          </AvatarFallback>
        </Avatar>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuLabel>
          <div className="flex items-center gap-3">
            <Avatar className="h-8 w-8">
              {user.profileImageUrl && (
                <AvatarImage src={user.profileImageUrl} alt={user.displayName || 'User'} />
              )}
              <AvatarFallback>
                {user.displayName
                  ?.split(' ')
                  .map(name => name.charAt(0))
                  .join('')
                  .toUpperCase() || 'U'}
              </AvatarFallback>
            </Avatar>
            <div className="flex flex-col">
              <Typography className="max-w-40 truncate">{user.displayName}</Typography>
              <Typography variant="secondary" className="max-w-40 truncate">
                {user.primaryEmail}
              </Typography>
            </div>
          </div>
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        
        {/* Theme Toggle */}
        <DropdownMenuItem onClick={toggleTheme}>
          <div className="flex gap-2 items-center">
            {theme === 'light' ? <Moon {...iconProps} /> : <Sun {...iconProps} />}
            <Typography>{theme === 'light' ? 'Dark' : 'Light'}</Typography>
          </div>
        </DropdownMenuItem>
        
        {/* Sign Out */}
        <DropdownMenuItem onClick={() => user.signOut()}>
          <div className="flex gap-2 items-center">
            <LogOut {...iconProps} />
            <Typography>Sign out</Typography>
          </div>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}