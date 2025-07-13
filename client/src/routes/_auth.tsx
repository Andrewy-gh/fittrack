import { createFileRoute, Outlet } from '@tanstack/react-router';
import { stackClientApp } from '@/stack';

export const Route = createFileRoute('/_auth')({
  beforeLoad: async () => {
    const user = await stackClientApp.getUser();
    if (!user) {
      throw new Error('User not found');
    }
    const { accessToken } = await user.getAuthJson();
    console.log('access token', accessToken);
    if (!accessToken) {
      throw new Error('Access token not found');
    }
  },
  component: LayoutComponent,
});

function LayoutComponent() {
  return <Outlet />;
}
