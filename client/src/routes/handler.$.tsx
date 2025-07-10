import { createFileRoute, useLocation } from '@tanstack/react-router';
import { StackHandler } from '@stackframe/react';
import { stackClientApp } from '@/stack';

export const Route = createFileRoute('/handler/$')({
  component: RouteComponent,
});

function RouteComponent() {
  const location = useLocation();
  return (
    <StackHandler app={stackClientApp} location={location.pathname} fullPage />
  );
}
