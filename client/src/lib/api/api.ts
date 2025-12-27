import { QueryClient } from '@tanstack/react-query';
import { showErrorToast } from '@/lib/errors';

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      staleTime: 5 * 60 * 1000, // 5 minutes
    },
    mutations: {
      onError: (error) => {
        // Default: show toast for all mutation errors
        // This can be overridden on a per-mutation basis by providing
        // a custom onError handler in the mutation options
        showErrorToast(error);
      },
    },
  },
});
