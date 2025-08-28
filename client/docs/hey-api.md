Example implementation:

```ts
// client-config.ts
import { client } from './client/client.gen';
import { stackClientApp } from '@/stack';

// Helper function to get current access token
const getAccessToken = async (): Promise<string | null> => {
  try {
    const user = await stackClientApp.getUser();
    if (!user) return null;
    
    const { accessToken } = await user.getAuthJson();
    return accessToken || null;
  } catch (error) {
    console.warn('Failed to get access token:', error);
    return null;
  }
};

// Configure the service client
client.setConfig({
  baseUrl: import.meta.env.VITE_API_BASE_URL, // your API base URL
});

// Add request interceptor for dynamic auth
client.interceptors.request.use(async (request, options) => {
  // Get fresh token for each request
  const accessToken = await getAccessToken();
  
  if (accessToken) {
    request.headers.set('Authorization', `Bearer ${accessToken}`);
    // or if you need a custom header:
    // request.headers.set('X-Custom-Auth', accessToken);
  }
  
  return request;
});

// Optional: Add response interceptor to handle auth errors
client.interceptors.response.use(
  (response) => response,
  async (error) => {
    // Handle 401 unauthorized - redirect to login
    if (error.status === 401) {
      await stackClientApp.signOut();
      // Redirect will be handled by your auth layout
    }
    throw error;
  }
);
```

**main.tsx**

```ts
// main.tsx
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App.tsx';
import './client-config'; // Import to initialize client configuration

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 60000,
      retry: (failureCount, error) => {
        // Don't retry on auth errors
        if (error && typeof error === 'object' && 'status' in error && error.status === 401) {
          return false;
        }
        return failureCount < 3;
      },
    },
  },
});

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <App />
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
  </React.StrictMode>
);
```

**For my use case**

```ts
// main.tsx - Configure the global service client
import { client } from './client/client.gen';
import { stackClientApp } from '@/stack';

// Configure the global client that all generated hooks will use
client.setConfig({
  baseUrl: import.meta.env.VITE_API_BASE_URL,
});

// Add dynamic auth interceptor to the global client
client.interceptors.request.use(async (request) => {
  const user = await stackClientApp.getUser();
  if (user) {
    const { accessToken } = await user.getAuthJson();
    if (accessToken) {
      request.headers.set('Authorization', `Bearer ${accessToken}`);
    }
  }
  return request;
});
```

