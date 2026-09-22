import { QueryClientProvider } from "@tanstack/react-query";
import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  Outlet,
  RouterProvider,
} from "@tanstack/react-router";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { Toaster, toast } from "sonner";
import { client } from "@/client/client.gen";
import { queryClient } from "@/lib/api/api";
import { exercisesQueryOptions } from "@/features/exercises/api/exercises";
import { useNewWorkoutFormWorkflow } from "@/features/workouts/hooks/use-new-workout-form-workflow";
import { analyticsSearchValidator } from "@/lib/route-search-validation";
import { AnalyticsPage } from "./analytics-page";
import type { WorkoutDraftStorage } from "@/lib/local-storage";
import type { ApplicationUser } from "@/lib/application-user";

const user: ApplicationUser = {
  id: "owner",
  displayName: "Owner",
  primaryEmail: "owner@example.test",
  profileImageUrl: null,
  signOut: async () => {},
};
const firstExercise = {
  id: 1,
  name: "Alphabetical first",
  created_at: "2026-01-01",
  updated_at: "2026-01-01",
  user_id: "owner",
};
const savedExercise = { ...firstExercise, id: 9, name: "Saved press" };

beforeEach(() => {
  queryClient.clear();
  queryClient.setDefaultOptions({ queries: { retry: false } });
  vi.spyOn(window, "scrollTo").mockImplementation(() => {});
  Element.prototype.scrollIntoView = vi.fn();
  vi.useFakeTimers({ toFake: ["Date"] });
  vi.setSystemTime(new Date("2026-09-23T12:00:00Z"));
  vi.stubGlobal(
    "ResizeObserver",
    class {
      observe() {}
      unobserve() {}
      disconnect() {}
    },
  );
  vi.stubGlobal(
    "matchMedia",
    vi.fn(() => ({
      matches: false,
      addEventListener() {},
      removeEventListener() {},
    })),
  );
});
afterEach(() => {
  toast.dismiss();
  queryClient.clear();
  vi.useRealTimers();
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});

async function mountSavedWorkout(
  savedDate: string,
  isNew: boolean,
  lookupFailure = false,
) {
  let saved = false;
  const requests: URL[] = [];
  const oldExercises = isNew ? [firstExercise] : [firstExercise, savedExercise];
  queryClient.setQueryData(exercisesQueryOptions().queryKey, oldExercises);
  // Mock the API adapter boundary; jsdom AbortSignal is incompatible with Node's Request.
  const respond = async (url: URL, method: string) => {
    requests.push(url);
    if (method === "POST" && url.pathname === "/api/workouts") {
      saved = true;
      return { data: { message: "saved" } };
    } else if (url.pathname === "/api/exercises") {
      if (saved && lookupFailure) throw new Error("offline");
      return { data: saved ? [firstExercise, savedExercise] : oldExercises };
    } else if (url.pathname === "/api/training-profile")
      return {
        data: { goals: ["strength", "hypertrophy"], primary_goal: "strength" },
      };
    else if (url.pathname === "/api/analytics/muscle-contributions")
      return {
        data: {
          period: { partial: savedDate.startsWith("2026-09") },
          muscles: [],
          exercises: [],
          workingSets: 0,
          invalidWorkingSets: 0,
          unclassifiedSets: 0,
          unreviewedSets: 0,
        },
      };
    else if (url.pathname.endsWith("/assessment")) {
      const expectedWeek = savedDate.startsWith("2026-08")
        ? "2026-08-10"
        : "2026-09-21";
      const matches =
        url.pathname === "/api/exercises/9/assessment" &&
        url.searchParams.get("startDate") === expectedWeek;
      return {
        data: {
          exerciseId: 9,
          period: {
            startDate: expectedWeek,
            endDate: "2026-09-28",
            timezone: "UTC",
            partial: expectedWeek === "2026-09-21",
          },
          reference: "2–3 working sets per exercise per session",
          workingSets: matches ? 2 : 0,
          trainingDays: matches ? 1 : 0,
          sessionsMeetingReference: matches ? 1 : 0,
          sourceUrl: "https://acsm.org",
          applicability: "General reference",
          limitations: [],
          sessions: matches
            ? [
                {
                  workoutId: 42,
                  date: savedDate,
                  workingSets: 2,
                  invalidWorkingSets: 0,
                  state: "reference_met",
                  reason:
                    "At least the lower reference of 2 working sets logged.",
                },
              ]
            : [],
        },
      };
    } else if (url.pathname.endsWith("/metrics-history"))
      return { data: { bucket: "workout", points: [] } };
    else if (url.pathname.startsWith("/api/exercises/"))
      return { data: { exercise: savedExercise, sets: [] } };
    else if (url.pathname.endsWith("/contribution-data"))
      return { data: { days: [] } };
    return { data: [] };
  };
  vi.spyOn(client, "get").mockImplementation((async (options: any) => {
    const path = options.url.replace("{id}", String(options.path?.id));
    const url = new URL(`http://localhost/api${path}`);
    for (const [key, value] of Object.entries(options.query ?? {}))
      url.searchParams.set(key, String(value));
    return respond(url, "GET");
  }) as any);
  vi.spyOn(client, "post").mockImplementation(
    async (options: any) =>
      respond(new URL(`http://localhost/api${options.url}`), "POST") as any,
  );
  const draftStorage: WorkoutDraftStorage = {
    load: () => ({
      date: savedDate,
      exercises: [
        {
          name: savedExercise.name,
          sets: [
            { reps: 5, weight: 50, setType: "working" },
            { reps: 5, weight: 50, setType: "working" },
          ],
        },
      ],
    }),
    save: vi.fn(),
    clear: vi.fn(),
  };
  function SavePage() {
    const workflow = useNewWorkoutFormWorkflow({
      user,
      exercises: oldExercises,
      newWorkoutContext: { focusTemplates: [] },
      draftStorage,
    });
    return (
      <button onClick={() => workflow.form.handleSubmit()}>Save workout</button>
    );
  }
  const root = createRootRoute({
    component: () => (
      <>
        <Outlet />
        <Toaster />
      </>
    ),
  });
  const saveRoute = createRoute({
    getParentRoute: () => root,
    path: "/workouts/new",
    component: SavePage,
  });
  const analyticsRoute = createRoute({
    getParentRoute: () => root,
    path: "/analytics",
    validateSearch: analyticsSearchValidator,
    component: () => {
      const { exerciseId, assessmentWeek } = analyticsRoute.useSearch();
      return (
        <AnalyticsPage
          user={user}
          exerciseId={exerciseId}
          assessmentWeek={assessmentWeek}
        />
      );
    },
  });
  const router = createRouter({
    routeTree: root.addChildren([saveRoute, analyticsRoute]),
    history: createMemoryHistory({ initialEntries: ["/workouts/new"] }),
  });
  await router.load();
  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
  fireEvent.click(screen.getByRole("button", { name: "Save workout" }));
  await screen.findByRole("button", { name: "View assessment" });
  expect(draftStorage.clear).toHaveBeenCalledWith("owner");
  return { router, requests };
}

describe("Saved workout assessment destination", () => {
  it("opens the saved existing exercise and backdated week after the form was reset", async () => {
    const { requests } = await mountSavedWorkout("2026-08-12T12:00:00Z", false);
    fireEvent.click(screen.getByRole("button", { name: "View assessment" }));
    expect(
      await screen.findByText("1 of 1 logged sessions reached 2 working sets."),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("combobox", { name: "Exercise options" }),
    ).toHaveTextContent("Saved press");
    expect(screen.getByText("Aug 10 – Aug 16, 2026")).toBeInTheDocument();
    await screen.findByRole("region", { name: "Muscle contributions" });
    await waitFor(() =>
      expect(
        requests.some(
          (url) =>
            url.pathname.endsWith("/muscle-contributions") &&
            url.searchParams.get("startDate") === "2026-08-10",
        ),
      ).toBe(true),
    );
    expect(
      screen.getByText("Reference met in logged training"),
    ).toBeInTheDocument();
  });
  it("resolves a new exercise absent from the cached list and opens current-week counts", async () => {
    const { requests } = await mountSavedWorkout("2026-09-22T12:00:00Z", true);
    fireEvent.click(screen.getByRole("button", { name: "View assessment" }));
    expect(
      await screen.findByText("Week in progress · counts so far only."),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("combobox", { name: "Exercise options" }),
    ).toHaveTextContent("Saved press");
    expect(screen.getByText("Sep 21 – Sep 27, 2026")).toBeInTheDocument();
    expect(
      screen.getByText(/2 working sets · 1 training day/),
    ).toBeInTheDocument();
    expect(
      screen.queryByText("Reference met in logged training"),
    ).not.toBeInTheDocument();
    expect(requests.some((url) => url.pathname === "/api/exercises")).toBe(
      true,
    );
  });
  it("keeps the successful save intact when fresh exercise lookup fails", async () => {
    const { router } = await mountSavedWorkout(
      "2026-09-22T12:00:00Z",
      true,
      true,
    );
    fireEvent.click(screen.getByRole("button", { name: "View assessment" }));
    await waitFor(
      () =>
        expect(
          screen.getByText("Could not open the assessment. Try again."),
        ).toBeInTheDocument(),
      { timeout: 4000 },
    );
    expect(router.state.location.pathname).toBe("/workouts/new");
    expect(screen.getByText("Workout saved successfully")).toBeInTheDocument();
  });
});
