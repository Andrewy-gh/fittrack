import { StrictMode } from 'react';
import ReactDOM from 'react-dom/client';
import { createRouter } from '@tanstack/react-router';
import { queryClient } from './lib/api/api.ts';
import { StackProvider, StackTheme, useUser } from '@stackframe/react';
import { stackClientApp } from './stack.ts';
import { ThemeProvider } from './components/theme-provider.tsx';
// Import the generated route tree
import { routeTree } from './routeTree.gen';

import './styles.css';
import reportWebVitals from './reportWebVitals.ts';

import { App } from './app.tsx';

// Create a new router instance
export const router = createRouter({
  routeTree,
  context: {
    queryClient,
    user: null,
  },
  defaultPreload: 'intent',
  scrollRestoration: true,
  defaultStructuralSharing: true,
  defaultPreloadStaleTime: 0,
});

// Register the router instance for type safety
declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}

// Render the app
const rootElement = document.getElementById('app');
if (rootElement && !rootElement.innerHTML) {
  const root = ReactDOM.createRoot(rootElement);

  function AppWithStack() {
    const user = useUser();
    return <App user={user} />;
  }

  root.render(
    <StrictMode>
      {stackClientApp ? (
        <StackProvider app={stackClientApp}>
          <StackTheme>
            <ThemeProvider>
              <AppWithStack />
            </ThemeProvider>
          </StackTheme>
        </StackProvider>
      ) : (
        <ThemeProvider>
          <App user={null} />
        </ThemeProvider>
      )}
    </StrictMode>
  );
}

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
