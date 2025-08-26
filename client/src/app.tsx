import { QueryClientProvider } from "@tanstack/react-query";
import { RouterProvider } from "@tanstack/react-router";
// import { useUser } from "@stackframe/react";
import { queryClient } from "./lib/api/api.ts";
import { router } from "./main.tsx";
import "./lib/api/client-config.ts";

export function App() {
  // const user = useUser();
  return (
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} 
      // context={{ user }} 
      />
    </QueryClientProvider>
  );
}