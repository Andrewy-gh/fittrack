import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  createMemoryHistory,
  createRouter,
  RouterProvider,
} from "@tanstack/react-router";
import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";
import { routeTree } from "@/routeTree.gen";
import type { ApplicationUser } from "@/lib/application-user";

function renderRoute(path: string, user: ApplicationUser | null = null) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  const router = createRouter({
    routeTree,
    history: createMemoryHistory({ initialEntries: [path] }),
    context: { user, queryClient },
  });
  return render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
}

describe("settings route nesting", () => {
  beforeEach(() => {
    window.scrollTo = () => undefined;
    window.matchMedia = (query) => ({
      matches: false,
      media: query,
      onchange: null,
      addListener: () => undefined,
      removeListener: () => undefined,
      addEventListener: () => undefined,
      removeEventListener: () => undefined,
      dispatchEvent: () => false,
    });
  });
  it.each([
    [
      "/settings",
      "Account settings",
      "Sign in to manage your account settings.",
    ],
    [
      "/settings/training-profile",
      "Training Profile",
      "Sign in to manage your training profile.",
    ],
  ])("renders the correct child at %s", async (path, heading, message) => {
    renderRoute(path);
    expect(
      await screen.findByRole("heading", { name: heading }),
    ).toBeInTheDocument();
    expect(screen.getByText(message)).toBeInTheDocument();
    expect(screen.getAllByRole("heading", { name: heading })).toHaveLength(1);
    expect(
      screen.queryByRole("heading", {
        name:
          heading === "Training Profile"
            ? "Account settings"
            : "Training Profile",
      }),
    ).not.toBeInTheDocument();
  });
});
