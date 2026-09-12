import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  createMemoryHistory,
  createRootRoute,
  createRouter,
  RouterProvider,
} from "@tanstack/react-router";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, expect, it } from "vitest";
import { client } from "@/client/client.gen";
import { TrainingProfilePage, emptyProfile } from "./training-profile-page";

const initialConfig = client.getConfig();
afterEach(() => client.setConfig(initialConfig));

it("persists multiple goals across reload, supports clearing, and isolates cached accounts", async () => {
  let persisted = JSON.stringify(emptyProfile);
  client.setConfig({
    baseUrl: "http://localhost/api",
    fetch: async (input, init) => {
      const request = new Request(input, init);
      if (request.method === "PUT") persisted = await request.text();
      return new Response(persisted, {
        headers: { "Content-Type": "application/json" },
      });
    },
  });
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  function show(userId: string) {
    const root = createRootRoute({
      component: () => <TrainingProfilePage userId={userId} />,
    });
    const router = createRouter({
      routeTree: root,
      history: createMemoryHistory({ initialEntries: ["/"] }),
    });
    return render(
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>,
    );
  }
  const user = userEvent.setup();
  let view = show("account-a");
  await user.click(await screen.findByRole("checkbox", { name: "Strength" }));
  await user.click(screen.getByRole("checkbox", { name: "Mobility" }));
  await user.click(screen.getByRole("button", { name: "Save" }));
  await waitFor(() =>
    expect(screen.getByRole("button", { name: "Save" })).toBeDisabled(),
  );
  view.unmount();
  queryClient.clear();
  view = show("account-a");
  expect(
    await screen.findByRole("checkbox", { name: "Strength" }),
  ).toBeChecked();
  expect(screen.getByRole("checkbox", { name: "Mobility" })).toBeChecked();
  await user.click(screen.getByRole("checkbox", { name: "Strength" }));
  await user.click(screen.getByRole("checkbox", { name: "Mobility" }));
  await user.click(screen.getByRole("button", { name: "Save" }));
  await waitFor(() =>
    expect(screen.getByRole("button", { name: "Save" })).toBeDisabled(),
  );
  view.unmount();
  queryClient.clear();
  view = show("account-a");
  expect(
    await screen.findByRole("checkbox", { name: "Strength" }),
  ).not.toBeChecked();
  expect(screen.getByRole("checkbox", { name: "Mobility" })).not.toBeChecked();
  view.unmount();
  persisted = JSON.stringify({
    ...emptyProfile,
    goals: ["endurance"],
    primary_goal: "endurance",
  });
  show("account-b");
  expect(
    await screen.findByRole("checkbox", { name: "Endurance" }),
  ).toBeChecked();
  expect(screen.getByRole("checkbox", { name: "Strength" })).not.toBeChecked();
});
