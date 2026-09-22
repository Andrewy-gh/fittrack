import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  createMemoryHistory,
  createRootRoute,
  createRouter,
  RouterProvider,
} from "@tanstack/react-router";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { client } from "@/client/client.gen";
import { TrainingAssessment } from "./training-assessment";

const result = {
  exerciseId: 1,
  period: {
    startDate: "2026-09-14",
    endDate: "2026-09-21",
    timezone: "UTC",
    partial: false,
  },
  workingSets: 6,
  trainingDays: 2,
  sessionsMeetingReference: 1,
  reference: "2–3 working sets per exercise per session",
  sourceUrl: "https://acsm.org/resistance-training-guidelines-update-2026/",
  applicability:
    "General reference for healthy adults. Personal applicability has not been verified.",
  limitations: [
    "Training outside FitTrack and intentional deloads are unknown.",
  ],
  policyVersion: "strength-session-reference-v1",
  sessions: [
    {
      workoutId: 1,
      date: "2026-09-15T12:00:00Z",
      workingSets: 1,
      invalidWorkingSets: 0,
      state: "below_reference",
      reason: "Fewer than the lower reference of 2 working sets.",
    },
    {
      workoutId: 2,
      date: "2026-09-17T12:00:00Z",
      workingSets: 5,
      invalidWorkingSets: 0,
      state: "reference_met",
      reason: "At least the lower reference of 2 working sets logged.",
    },
  ],
};

async function mount(
  goals: string[] = ["strength", "hypertrophy"],
  partial = false,
) {
  const api = vi.spyOn(client, "get").mockImplementation(
    async (options: any) =>
      ({
        data:
          options.url === "/training-profile"
            ? { goals, primary_goal: goals[0] ?? null }
            : options.url === "/analytics/muscle-contributions"
              ? {
                  period: { ...result.period, partial },
                  muscles: [],
                  exercises: [],
                  workingSets: 0,
                  invalidWorkingSets: 0,
                  unclassifiedSets: 0,
                  unreviewedSets: 0,
                }
              : { ...result, period: { ...result.period, partial } },
      }) as any,
  );
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  const root = createRootRoute({
    component: () => (
      <TrainingAssessment
        userId="owner"
        exerciseId={1}
      />
    ),
  });
  const router = createRouter({
    routeTree: root,
    history: createMemoryHistory({ initialEntries: ["/"] }),
  });
  await router.load();
  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
  return api;
}

beforeEach(() => vi.restoreAllMocks());
describe("Training assessment", () => {
  it("shows mixed session findings and the reference instead of an overall verdict", async () => {
    await mount();
    expect(
      await screen.findByText("1 of 2 logged sessions reached 2 working sets."),
    ).toBeInTheDocument();
    expect(screen.getByText("Below reference")).toBeInTheDocument();
    expect(
      screen.getByText("Reference met in logged training"),
    ).toBeInTheDocument();
    fireEvent.click(screen.getByText("About this comparison"));
    expect(
      screen.getByText(/Personal applicability has not been verified/),
    ).toBeVisible();
    fireEvent.click(screen.getByText("Other goal coverage"));
    expect(screen.getByText("Muscle growth · counts available")).toBeVisible();
  });
  it("shows only counts for an unfinished week", async () => {
    await mount(["strength"], true);
    expect(
      await screen.findByText("Week in progress · counts so far only."),
    ).toBeInTheDocument();
    expect(screen.queryByText("Below reference")).not.toBeInTheDocument();
    expect(
      screen.queryByText("Reference met in logged training"),
    ).not.toBeInTheDocument();
  });
  it("does not substitute strength for other selected goals", async () => {
    const api = await mount([
      "mobility",
      "endurance",
      "weight_loss",
      "general_fitness",
    ]);
    await screen.findByText("Goal coverage");
    expect(
      screen.queryByRole("region", { name: "Strength reference" }),
    ).not.toBeInTheDocument();
    expect(api).toHaveBeenCalledTimes(1);
    expect(screen.getByText("Mobility · not assessed")).toBeVisible();
  });
  it("navigates weeks with a new period request and disables future weeks", async () => {
    const api = await mount();
    await screen.findByText("Below reference");
    fireEvent.click(
      screen.getByRole("button", { name: "Next assessment week" }),
    );
    expect(
      screen.getByRole("button", { name: "Next assessment week" }),
    ).toBeDisabled();
    await waitFor(() =>
      expect(
        api.mock.calls.filter(([options]) =>
          options.url.endsWith("/assessment"),
        ),
      ).toHaveLength(2),
    );
  });
});

it("offers week navigation for hypertrophy alone without a strength verdict", async () => {
  const api = await mount(["hypertrophy"], true);
  await screen.findByRole("region", { name: "Muscle contributions" });
  expect(
    screen.queryByRole("region", { name: "Strength reference" }),
  ).not.toBeInTheDocument();
  expect(
    await screen.findByText("Week in progress · contributions so far."),
  ).toBeInTheDocument();
  fireEvent.click(
    screen.getByRole("button", { name: "Previous assessment week" }),
  );
  await waitFor(() =>
    expect(
      api.mock.calls.filter(
        ([options]) => options.url === "/analytics/muscle-contributions",
      ),
    ).toHaveLength(2),
  );
  expect(
    api.mock.calls.some(([options]) => options.url.endsWith("/assessment")),
  ).toBe(false);
});
