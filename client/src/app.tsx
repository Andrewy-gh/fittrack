import { QueryClientProvider } from "@tanstack/react-query";
import { RouterProvider } from "@tanstack/react-router";
import { queryClient } from "./lib/api/api.ts";
import { router } from "./main.tsx";
import "./lib/api/client-config.ts";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { ReloadPrompt } from "./components/reload-prompt";
import { Toaster } from "@/components/ui/sonner";
import type { CurrentInternalUser, CurrentUser } from "@stackframe/react";

export function App({
  user,
}: {
  user: CurrentUser | CurrentInternalUser | null;
}) {
  return (
    <QueryClientProvider client={queryClient}>
      <RouterProvider
        router={router}
        context={{ queryClient, user }}
      />
      <ReactQueryDevtools initialIsOpen={false} />
      <ReloadPrompt />
      <Toaster position="bottom-right" />
    </QueryClientProvider>
  );
}
