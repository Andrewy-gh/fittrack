import { createFileRoute, Outlet } from '@tanstack/react-router';
import { Header } from '@/components/header';
import { stackClientApp } from '@/stack';

export const Route = createFileRoute('/_auth')({
  beforeLoad: async () => {
    const user = await stackClientApp.getUser({ or: 'redirect' });
    return { user };
  },
  component: LayoutComponent,
});

function LayoutComponent() {
  return (
    <>
      <Header />
      <Outlet />
    </>
  );
}
