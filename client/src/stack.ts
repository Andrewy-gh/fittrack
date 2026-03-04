import { StackClientApp as StackClientAppCtor, type StackClientApp } from '@stackframe/react';
import { useNavigate } from '@tanstack/react-router';

const projectId = import.meta.env.VITE_PROJECT_ID as string | undefined;
const publishableClientKey = import.meta.env.VITE_PUBLISHABLE_CLIENT_KEY as string | undefined;

export const isStackConfigured = Boolean(projectId && publishableClientKey);

export const stackClientApp: StackClientApp<true, string> | null = (() => {
  if (!isStackConfigured) return null;
  try {
    return new StackClientAppCtor({
      projectId: projectId!,
      publishableClientKey: publishableClientKey!,
      tokenStore: 'cookie',
      redirectMethod: {
        useNavigate: () => {
          const navigate = useNavigate();
          return (to: string) => {
            navigate({ to });
          };
        },
      },
    }) as StackClientApp<true, string>;
  } catch (err) {
    // Missing/invalid Stack Auth config should not take down demo mode.
    console.error('Failed to initialize Stack Auth client; falling back to demo mode.', err);
    return null;
  }
})();
