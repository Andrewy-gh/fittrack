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
      <div className="bg-yellow-100 border-b border-yellow-300 px-4 py-2 text-center text-sm text-yellow-800">
        <strong>Demo Mode:</strong> You're viewing sample data. Changes won't be
        saved.
      </div>
      <Outlet />
    </>
  );
}
