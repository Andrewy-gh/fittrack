import { createFileRoute, Outlet } from '@tanstack/react-router';
import { Header } from '@/components/header';
import { DemoHeader } from '@/components/demo-header';

export const Route = createFileRoute('/_layout')({
  component: LayoutComponent,
});

function LayoutComponent() {
  const { user } = Route.useRouteContext();

  return (
    <div className="pb-[calc(5rem+env(safe-area-inset-bottom))] md:pb-0">
      {user ? <Header /> : <DemoHeader />}
      <Outlet />
    </div>
  );
}
