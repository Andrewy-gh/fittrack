**Question how do I implement client interceptor for Hey API with my existing auth flow?**

## Context

I am using the open source auth library "@stackframe/react"
This is how a stack client is defined in my code:

```ts
// stack.ts
import { StackClientApp } from '@stackframe/react';
import { useNavigate } from '@tanstack/react-router';
export const stackClientApp = new StackClientApp({
  // You should store these in environment variables
  projectId: import.meta.env.VITE_PROJECT_ID,
  publishableClientKey: import.meta.env.VITE_PUBLISHABLE_CLIENT_KEY,
  tokenStore: 'cookie',
  redirectMethod: {
    useNavigate: () => {
      const navigate = useNavigate();
      return (to: string) => {
        navigate({ to });
      };
    },
  },
});
```

It's usage in a `_auth.tsx` Layout page for tanstack router:

```ts
import { createFileRoute, Outlet } from '@tanstack/react-router';
import { stackClientApp } from '@/stack';
export const Route = createFileRoute('/_auth')({
  beforeLoad: async () => {
    const user = await stackClientApp.getUser({ or: 'redirect' });
    if (!user) {
      throw new Error('User not found');
    }
    if (!user.id || typeof user.id !== 'string') {
      throw new Error('User ID not found');
    }
    const { accessToken } = await user.getAuthJson();
    if (!accessToken) {
      throw new Error('Access token not found');
    }
  },
  component: LayoutComponent,
});
function LayoutComponent() {
  return <Outlet />;
}
```

This is from hey-api's tanstack query plugin example:

```ts
// App.tsx
// all these are generated imports
import { createClient } from './client/client';
import {
  addPetMutation,
  getPetByIdOptions,
  updatePetMutation,
} from './client/@tanstack/react-query.gen';
import { PetSchema } from './client/schemas.gen';
import type { Pet } from './client/types.gen';
const localClient = createClient({
  // set default base url for requests made by this client
  baseUrl: 'https://petstore3.swagger.io/api/v3',
  /**
   * Set default headers only for requests made by this client. This is to
   * demonstrate local clients and their configuration taking precedence over
   * internal service client.
   */
  headers: {
    Authorization: 'Bearer <token_from_local_client>',
  },
});
localClient.interceptors.request.use((request, options) => {
  // Middleware is great for adding authorization tokens to requests made to
  // protected paths. Headers are set randomly here to allow surfacing the
  // default headers, too.
  if (
    options.url === '/pet/{petId}' &&
    options.method === 'GET' &&
    Math.random() < 0.5
  ) {
    request.headers.set('Authorization', 'Bearer <token_from_interceptor>');
  }
  return request;
});
function App() {
// ... rest of code omitted
```

and 

```ts
// main.tsx
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App.tsx';
import { client } from './client/client.gen';
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 60000,
    },
  },
});
// configure internal service client
client.setConfig({
  // set default base url for requests
  baseUrl: 'https://petstore3.swagger.io/api/v3',
  // set default headers for requests
  headers: {
    Authorization: 'Bearer <token_from_service_client>',
  },
});
ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <Theme appearance="dark">
        <App />
      </Theme>
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
  </React.StrictMode>
);
```

## Answer 

Perfect! With Stackframe and hey-api's TanStack Query plugin, you'll want to set up a request interceptor that dynamically fetches the access token. Here's how to integrate them:

### Setup the Client with Dynamic Auth

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
    // I AM USING A CUSTOM HEADER `x-stack-access-token`
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

### Update Your Main.tsx

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

### Using in Your Components

Now you can use the generated TanStack Query hooks directly:

```ts
// SomeComponent.tsx
import { getPetByIdOptions, addPetMutation } from './client/@tanstack/react-query.gen';
import { useQuery, useMutation } from '@tanstack/react-query';

function SomeComponent() {
  // The interceptor will automatically add the auth token
  const { data: pet } = useQuery(getPetByIdOptions({ 
    path: { petId: '123' } 
  }));
  
  const addPet = useMutation(addPetMutation());
  
  return (
    <div>
      {/* Your component JSX */}
    </div>
  );
}
```

## Follow-up question 

Why does the hey-api example create two clients? Is it because App.tsx deals with authentication and main.tsx does not?

## Answer

Looking at the hey-api example more carefully, they're creating two clients to demonstrate **different configuration scopes** and **precedence**, not specifically for authentication vs non-authentication purposes.
Here's what's happening:

### Two Client Types

Service Client (`main.tsx`): This is the global/default client that all generated hooks use by default
Local Client (`App.tsx`): This is a custom client instance for specific use cases

### Why Two Clients?
The example creates two clients to demonstrate:

1. Global Configuration: The service client in main.tsx sets default config for ALL generated queries/mutations
2. Override Capability: The local client shows you can create custom client instances with different configurations that override the global ones
3. Precedence: Local client config takes precedence over service client config

### The Authentication Demo
In their example, they're showing different auth tokens:

- Service client: 'Bearer <token_from_service_client>' (global default)
- Local client: 'Bearer <token_from_local_client>' (override)
- Interceptor: 'Bearer <token_from_interceptor>' (dynamic override)

The interceptor randomly applies to show that interceptors have the highest precedence.

### For Your Use Case

You actually **don't need two clients**. Since you want dynamic authentication for all requests, you should just configure the single service client:
```typescript
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
### When You Might Want Multiple Clients

You'd only need multiple clients if you have different requirements like:

- Different base URLs (e.g., auth service vs main API)
- Different authentication methods for different endpoints
- Different retry policies or timeouts
- Integration with multiple APIs

For your single API with consistent auth requirements, stick with one client configuration.