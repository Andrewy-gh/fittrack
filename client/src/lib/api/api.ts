import { MutationCache, QueryClient } from '@tanstack/react-query';
import { showErrorToast } from '@/lib/errors';

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      staleTime: 5 * 60 * 1000, // 5 minutes
    },
  },
  mutationCache: new MutationCache({
    onError: (error, _variables, _context, mutation) => {
      // Skip global error handler if mutation opts out via meta
      if (mutation.options.meta?.skipGlobalErrorHandler) return;
      showErrorToast(error);
    },
  }),
});
