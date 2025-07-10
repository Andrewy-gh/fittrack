import { StackClientApp } from '@stackframe/react';
import { useNavigate } from '@tanstack/react-router';

export const stackClientApp = new StackClientApp({
  // You should store these in environment variables
  projectId: import.meta.env.VITE_PROJECT_ID,
  publishableClientKey: import.meta.env.VITE_PUBLISHABLE_CLIENT_KEY,
  tokenStore: 'cookie',
  redirectMethod: {
    useNavigate: () => {
      const navigate = useNavigate();
      return (to: string) => {
        navigate({ to });
      };
    },
  },
});