import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, fireEvent } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import { client } from "@/client/client.gen";
import { MuscleContributions } from "./muscle-contributions";
const data = {
  period: { partial: false },
  workingSets: 11,
  invalidWorkingSets: 1,
  unclassifiedSets: 1,
  unreviewedSets: 0,
  mappingVersion: "v1",
  catalogVersion: "c1",
  muscles: [
    {
      muscle: "triceps",
      directSets: 4,
      indirectSets: 6,
      unresolvedSets: 1,
      invalidMappedSets: 1,
      invalidUnresolvedSets: 0,
    },
    {
      muscle: "hamstrings",
      directSets: 0,
      indirectSets: 0,
      unresolvedSets: 11,
      invalidMappedSets: 0,
      invalidUnresolvedSets: 1,
    },
  ],
  exercises: [
    {
      exerciseId: 1,
      name: "My bench",
      workingSets: 6,
      invalidWorkingSets: 1,
      roles: { triceps: "indirect" },
    },
    {
      exerciseId: 2,
      name: "Pushdown",
      workingSets: 4,
      invalidWorkingSets: 0,
      roles: { triceps: "direct" },
    },
    {
      exerciseId: 3,
      name: "Custom",
      workingSets: 1,
      invalidWorkingSets: 0,
      roles: {},
    },
  ],
};
afterEach(() => vi.restoreAllMocks());
function mount() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(
    <QueryClientProvider client={queryClient}>
      <MuscleContributions
        userId="owner"
        startDate="2026-09-14"
        timezone="America/New_York"
      />
    </QueryClientProvider>,
  );
}
it("preserves known contributions and exposes unknown and invalid roles without verdicts", async () => {
  vi.spyOn(client, "get").mockResolvedValue({ data } as never);
  mount();
  expect(await screen.findByText("4 direct · 6 indirect")).toBeVisible();
  expect(screen.getByText("1 sets not reviewed for this muscle")).toBeVisible();
  expect(screen.getByText("No supported mapped contributions")).toBeVisible();
  fireEvent.click(screen.getByText("Contributing exercises"));
  expect(screen.getByText("My bench: 6 indirect · 1 invalid")).toBeVisible();
  fireEvent.click(screen.getByText("Coverage and counting"));
  expect(
    screen.getByText("Custom: 1 working sets · mapping unavailable"),
  ).toBeVisible();
  expect(
    screen.queryByText(/Below reference|Reference met/),
  ).not.toBeInTheDocument();
});
it("lets a failed request recover without changing the selected period", async () => {
  const api = vi
    .spyOn(client, "get")
    .mockRejectedValueOnce(new Error("offline"))
    .mockResolvedValue({ data } as never);
  mount();
  expect(await screen.findByRole("alert")).toHaveTextContent(
    "Could not load muscle contributions",
  );
  fireEvent.click(screen.getByRole("button", { name: "Retry" }));
  expect(await screen.findByText("4 direct · 6 indirect")).toBeVisible();
  expect(api.mock.calls.at(-1)?.[0].query).toEqual({
    startDate: "2026-09-14",
    timezone: "America/New_York",
  });
});
