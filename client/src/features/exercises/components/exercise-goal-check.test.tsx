import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  createRootRoute,
  createRoute,
  createRouter,
  createMemoryHistory,
  RouterProvider,
} from "@tanstack/react-router";
import { beforeEach, expect, it, vi } from "vitest";
import {
  getExercisesByIdGoalCheck,
  type ExerciseStrengthGoalCheck,
} from "@/client";
import { ExerciseGoalCheck } from "./exercise-goal-check";
import { invalidateGoalChecks } from "../api/goal-check-cache";
vi.mock("@/client", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/client")>()),
  getExercisesByIdGoalCheck: vi.fn(),
}));
const fixture: ExerciseStrengthGoalCheck = {
  exercise_id: 1,
  applicable: true,
  policy_version: "strength-observations-v1",
  window: {
    start: "2026-09-07T00:00:00Z",
    end: "2026-09-13T12:00:00Z",
    timezone: "UTC",
  },
  sessions: [
    { workout_id: 1, date: "2026-09-10T12:00:00Z", working_sets: 1 },
    { workout_id: 2, date: "2026-09-12T12:00:00Z", working_sets: 5 },
  ],
  frequency: { training_days: 2, reference_min: 2, status: "reference_met" },
  working_sets: {
    total: 6,
    average: 3,
    reference_min: 2,
    reference_max: 3,
    status: "reference_met",
  },
  loading: {
    reps_min: 5,
    reps_max: 8,
    sets_with_load: 0,
    reference: null,
    percent_min: null,
    percent_max: null,
  },
};
const api = vi.mocked(getExercisesByIdGoalCheck);
function reply(data: ExerciseStrengthGoalCheck) {
  return {
    data,
    request: new Request("http://localhost/api/exercises/1/goal-check"),
    response: new Response(),
  };
}
async function mount(demo = false) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  const root = createRootRoute();
  const route = createRoute({
    getParentRoute: () => root,
    path: "/",
    component: () => (
      <ExerciseGoalCheck
        exerciseId={1}
        isDemoMode={demo}
        userId="owner"
      />
    ),
  });
  const router = createRouter({
    routeTree: root.addChildren([route]),
    history: createMemoryHistory({ initialEntries: ["/"] }),
  });
  await router.load();
  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
  return queryClient;
}
beforeEach(() => {
  api.mockReset();
  api.mockResolvedValue(reply(fixture));
});
it("shows measurable comparisons, session distribution and research links", async () => {
  await mount();
  expect(await screen.findByText("Working sets per session")).toBeVisible();
  expect(screen.getAllByText("Meets guide")).toHaveLength(2);
  expect(screen.getByText(/Not enough weight data/)).toBeVisible();
  await userEvent.click(screen.getByText("Your sessions"));
  expect(screen.getByText("1 working set")).toBeVisible();
  expect(screen.getByText("5 working sets")).toBeVisible();
  await userEvent.click(screen.getByText("About this check"));
  expect(
    screen.getByRole("link", { name: "ACSM position stand (2026)" }),
  ).toHaveAttribute("href", "https://pubmed.ncbi.nlm.nih.gov/41843416/");
  expect(
    screen.queryByText(/on track|heavy loading sufficient/i),
  ).not.toBeInTheDocument();
});
it("shows loading, recovers from an error, and refreshes after invalidation", async () => {
  api.mockImplementationOnce(
    () =>
      new Promise<never>((_, reject) =>
        setTimeout(() => reject(new Error("offline")), 100),
      ),
  );
  const client = await mount();
  expect(await screen.findByRole("status")).toHaveTextContent(
    "Loading Goal Check",
  );
  expect(await screen.findByRole("alert")).toHaveTextContent(
    "Couldn’t load this check.",
  );
  await userEvent.click(
    screen.getByRole("button", { name: "Retry Goal Check" }),
  );
  expect(await screen.findByText("Working sets per session")).toBeVisible();
  api.mockResolvedValue(
    reply({
      ...fixture,
      sessions: [],
      frequency: {
        ...fixture.frequency,
        training_days: 0,
        status: "insufficient_data",
      },
      working_sets: {
        ...fixture.working_sets,
        average: null,
        total: 0,
        status: "insufficient_data",
      },
    }),
  );
  await invalidateGoalChecks(client);
  expect(
    await screen.findByText(/No working sets logged in the last 7 days/),
  ).toBeVisible();
});
it("shows an informational saved-reference comparison without clamping", async () => {
  api.mockResolvedValue(
    reply({
      ...fixture,
      loading: {
        ...fixture.loading,
        sets_with_load: 6,
        reference: {
          value: 100,
          origin: "workout_estimate",
          updated_at: null,
          source_workout_id: 1,
        },
        percent_min: 85,
        percent_max: 120,
      },
    }),
  );
  await mount();
  expect(await screen.findByText("85–120% of saved max")).toBeVisible();
  await userEvent.click(screen.getByText("About this check"));
  expect(screen.getByText(/doesn’t tell us whether/)).toBeVisible();
});
it("does not assess a hypertrophy-only profile", async () => {
  api.mockResolvedValue(reply({ ...fixture, applicable: false }));
  await mount();
  expect(
    await screen.findByRole("link", { name: "Training profile" }),
  ).toBeVisible();
  expect(screen.queryByText("Meets guide")).not.toBeInTheDocument();
});
it("explains demo availability without making an authenticated request", async () => {
  await mount(true);
  await screen.findByRole("heading", {
    name: "Strength Goal Check",
  });
  expect(screen.getByText(/Sign in and choose strength/)).toBeVisible();
  expect(api).not.toHaveBeenCalled();
});
