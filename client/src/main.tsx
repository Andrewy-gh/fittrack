import { StrictMode } from 'react';
import ReactDOM from 'react-dom/client';
import { RouterProvider, createRouter } from '@tanstack/react-router';
import {
  StackProvider,
  StackTheme,
} from '@stackframe/react';
import { stackClientApp } from './stack.ts';
import { ThemeProvider } from './components/theme-provider.tsx';
import { useUser } from '@stackframe/react';
// Import the generated route tree
import { routeTree } from './routeTree.gen';

import './styles.css';
import reportWebVitals from './reportWebVitals.ts';

// Create a new router instance
const router = createRouter({
  routeTree,
  context: {
    user: undefined!,
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

function App() {
  const user = useUser();
  return <RouterProvider router={router} context={{ user }} />;
}

// Render the app
const rootElement = document.getElementById('app');
if (rootElement && !rootElement.innerHTML) {
  const root = ReactDOM.createRoot(rootElement);
  root.render(
    <StrictMode>
      <StackProvider app={stackClientApp}>
        <StackTheme>
          <ThemeProvider>
            <App />
          </ThemeProvider>
        </StackTheme>
      </StackProvider>
    </StrictMode>
  );
}

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
