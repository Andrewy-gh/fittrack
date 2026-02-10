import { createFileRoute, useLocation } from '@tanstack/react-router';
import { StackHandler } from '@stackframe/react';
import { stackClientApp } from '@/stack';

export const Route = createFileRoute('/handler/$')({
  component: RouteComponent,
});

function RouteComponent() {
  const location = useLocation();
  if (!stackClientApp) {
    return (
      <div className="p-6">
        <h1 className="text-lg font-semibold">Auth Not Configured</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          Stack Auth is not configured in this environment.
        </p>
      </div>
    );
  }
  return (
    <StackHandler app={stackClientApp} location={location.pathname} fullPage />
  );
}
