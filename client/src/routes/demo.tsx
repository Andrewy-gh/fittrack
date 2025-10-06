import { createFileRoute, Outlet } from '@tanstack/react-router';
import { Header } from '@/components/header';
import { initializeDemoData } from '@/lib/demo-data/storage';

export const Route = createFileRoute('/demo')({
  beforeLoad: async () => {
    // Initialize demo data in localStorage if not already present
    initializeDemoData();
    return {};
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
