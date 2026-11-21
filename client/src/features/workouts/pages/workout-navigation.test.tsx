import { QueryClientProvider } from "@tanstack/react-query";
import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  RouterProvider,
} from "@tanstack/react-router";
import { act, render, screen, waitFor } from "@testing-library/react";
import { StrictMode, Suspense } from "react";
import { afterEach, expect, it, vi } from "vitest";
import { queryClient } from "@/lib/api/api";
import { createWorkoutDraftStorage } from "@/lib/local-storage";
import { NewWorkoutPage } from "./new-workout-page";
import { EditWorkoutPage } from "./edit-workout-page";

const workout = {
  date: "2026-10-08T12:00:00Z",
  exercises: [
    { name: "Pushups", sets: [{ reps: 10, setType: "working" as const }] },
  ],
};
const user = {
  id: "owner",
  displayName: "Owner",
  primaryEmail: null,
  profileImageUrl: null,
  signOut: async () => {},
};

afterEach(() => {
  queryClient.clear();
  vi.restoreAllMocks();
});

it.each(["new", "edit"] as const)(
  "%s workout does not redirect an uncommitted suspended render",
  async (page) => {
    vi.spyOn(window, "scrollTo").mockImplementation(() => {});
    let release = () => {};
    let ready = false;
    const pending = new Promise<void>((resolve) => {
      release = resolve;
    });
    function Suspend() {
      if (!ready) throw pending;
      return null;
    }
    const root = createRootRoute();
    const path = page === "new" ? "/workouts/new" : "/workouts/$workoutId/edit";
    const route = createRoute({
      getParentRoute: () => root,
      path,
      validateSearch: (search: Record<string, unknown>) => ({
        exerciseIndex:
          typeof search.exerciseIndex === "number"
            ? search.exerciseIndex
            : undefined,
      }),
      component: () => {
        const search = route.useSearch();
        return (
          <Suspense fallback={<p>Loading pending workout</p>}>
            {page === "new" ? (
              <NewWorkoutPage
                user={user}
                exercises={[]}
                newWorkoutContext={{ focusTemplates: [] }}
                workoutsFocus={[]}
                search={search}
              />
            ) : (
              <EditWorkoutPage
                user={user}
                exercises={[]}
                workout={workout}
                workoutId={1}
                workoutsFocus={[]}
                search={search}
              />
            )}
            <Suspend />
          </Suspense>
        );
      },
    });
    const url = page === "new" ? "/workouts/new" : "/workouts/1/edit";
    const router = createRouter({
      routeTree: root.addChildren([route]),
      history: createMemoryHistory({
        initialEntries: [`${url}?exerciseIndex=2`],
      }),
    });
    render(
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>,
    );
    await screen.findByText("Loading pending workout");
    expect(router.state.location.search).toEqual({ exerciseIndex: 2 });
    ready = true;
    await act(async () => {
      release();
    });
    await waitFor(() => expect(router.state.location.search).toEqual({}));
  },
  60000,
);

it.each(["new", "edit"] as const)(
  "%s workout redirects invalid indices without losing the workout and retains valid exercise navigation",
  async (page) => {
    vi.spyOn(window, "scrollTo").mockImplementation(() => {});
    const errors = vi.spyOn(console, "error").mockImplementation(() => {});
    const entries = new Map<string, string>();
    const draftStorage = createWorkoutDraftStorage({
      getItem: (key) => entries.get(key) ?? null,
      setItem: (key, value) => {
        entries.set(key, value);
      },
      removeItem: (key) => {
        entries.delete(key);
      },
    });
    draftStorage.save(workout, user.id);
    const root = createRootRoute();
    const path = page === "new" ? "/workouts/new" : "/workouts/$workoutId/edit";
    const route = createRoute({
      getParentRoute: () => root,
      path,
      validateSearch: (search: Record<string, unknown>) => ({
        exerciseIndex:
          typeof search.exerciseIndex === "number"
            ? search.exerciseIndex
            : undefined,
        newExercise: search.newExercise === true ? true : undefined,
      }),
      component: () => {
        const search = route.useSearch();
        return page === "new" ? (
          <NewWorkoutPage
            user={user}
            exercises={[]}
            newWorkoutContext={{ focusTemplates: [] }}
            workoutsFocus={[]}
            search={search}
            draftStorage={draftStorage}
          />
        ) : (
          <EditWorkoutPage
            user={user}
            exercises={[]}
            workout={workout}
            workoutId={1}
            workoutsFocus={[]}
            search={search}
          />
        );
      },
    });
    const url = page === "new" ? "/workouts/new" : "/workouts/1/edit";
    const router = createRouter({
      routeTree: root.addChildren([route]),
      history: createMemoryHistory({
        initialEntries: [`${url}?exerciseIndex=2&newExercise=true`],
      }),
    });
    render(
      <StrictMode>
        <QueryClientProvider client={queryClient}>
          <RouterProvider router={router} />
        </QueryClientProvider>
      </StrictMode>,
    );
    await screen.findByRole("button", { name: /Edit Pushups/ });
    expect(router.state.location.pathname).toBe(url);
    expect(router.state.location.search).toEqual({});
    await router.navigate({
      to: path,
      params: { workoutId: 1 },
      search: { exerciseIndex: 0 },
    });
    await screen.findByText("Pushups", { selector: "h1" });
    expect(router.state.location.search).toMatchObject({ exerciseIndex: 0 });
    await router.navigate({
      to: path,
      params: { workoutId: 1 },
      search: { exerciseIndex: -1, newExercise: true },
    });
    await screen.findByRole("button", { name: /Edit Pushups/ });
    await waitFor(() => expect(router.state.location.search).toEqual({}));
    expect(errors.mock.calls.flat().join(" ")).not.toMatch(
      /Cannot update a component.*while rendering/i,
    );
  },
  60000,
);
