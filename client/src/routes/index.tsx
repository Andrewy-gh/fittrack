import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/')({
  loader: async () => {
    const res = await fetch('/api/hello');
    if (!res.ok) {
      throw new Error('Failed to fetch data');
    }
    const data = await res.json();
    return data;
  },
  component: App,
});

function App() {
  const { message } = Route.useLoaderData();

  return <main>{message}</main>;
}
