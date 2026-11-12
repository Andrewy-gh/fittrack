import { QueryClientProvider } from "@tanstack/react-query";
import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  RouterProvider,
} from "@tanstack/react-router";
import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import { useState } from "react";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { client } from "@/client/client.gen";
import { queryClient } from "@/lib/api/api";
import { getExerciseCatalogQueryOptions } from "@/client/@tanstack/react-query.gen";
import { AddExerciseScreen } from "../components/form/add-exercise-screen";
import { EditWorkoutPage } from "./edit-workout-page";
import { NewWorkoutPage } from "./new-workout-page";
import { useNewWorkoutFormWorkflow } from "../hooks/use-new-workout-form-workflow";
import { createWorkoutDraftStorage } from "@/lib/local-storage";
import type { ApplicationUser } from "@/lib/application-user";

const user: ApplicationUser = {
  id: "owner",
  displayName: "Owner",
  primaryEmail: null,
  profileImageUrl: null,
  signOut: async () => {},
};

beforeEach(() => {
  vi.spyOn(window, "scrollTo").mockImplementation(() => {});
});

afterEach(() => {
  queryClient.clear();
  vi.restoreAllMocks();
  vi.useRealTimers();
});

async function mount(
  path: "/workouts/$workoutId/edit" | "/workouts/new",
  component: () => React.JSX.Element,
) {
  const root = createRootRoute();
  const route = createRoute({ getParentRoute: () => root, path, component });
  const detail = createRoute({
    getParentRoute: () => root,
    path: "/workouts/$workoutId",
    component: () => <p>Saved workout</p>,
  });
  const router = createRouter({
    routeTree: root.addChildren([route, detail]),
    history: createMemoryHistory({
      initialEntries: [path === "/workouts/new" ? path : "/workouts/1/edit"],
    }),
  });
  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
  await router.load();
  return router;
}

it.each(["new", "edit"] as const)(
  "%s workout cards show only actual set metrics, not invented history",
  async (page) => {
    const workout = {
      date: "2026-10-08T12:00:00Z",
      exercises: [
        {
          name: "Pushups",
          sets: [{ reps: 10, weight: 20, setType: "working" as const }],
        },
        { name: "Squats", sets: [] },
      ],
    };
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
    const router = await mount(
      page === "new" ? "/workouts/new" : "/workouts/$workoutId/edit",
      () =>
        page === "new" ? (
          <NewWorkoutPage
            user={user}
            exercises={[]}
            newWorkoutContext={{ focusTemplates: [] }}
            workoutsFocus={[]}
            search={{}}
            draftStorage={draftStorage}
          />
        ) : (
          <EditWorkoutPage
            user={user}
            exercises={[]}
            workout={workout}
            workoutId={1}
            workoutsFocus={[]}
            search={{}}
          />
        ),
    );
    const metrics = await screen.findByRole("button", {
      name: /Pushups: 1 sets, .*200.* volume/,
    });
    // The former chart was an unlabeled seven-bar cluster with percentage heights.
    // Read the height property, not style text that also matches line-height.
    const fabricatedBars = Array.from(
      metrics.querySelectorAll<HTMLDivElement>("div[style]"),
    ).filter(
      (bar) =>
        bar.style.height.endsWith("%") &&
        bar.parentElement?.children.length === 7,
    );
    expect(fabricatedBars).toHaveLength(0);
    fireEvent.click(screen.getByRole("button", { name: "Edit order" }));
    expect(
      screen.getByRole("button", { name: "Reorder Pushups" }),
    ).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    fireEvent.click(screen.getByRole("button", { name: "Delete Squats" }));
    expect(
      screen.queryByRole("button", { name: "Edit Squats" }),
    ).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Edit Pushups" }));
    await waitFor(() =>
      expect(router.state.location.search).toMatchObject({ exerciseIndex: 0 }),
    );
  },
  60000,
);

it("submits explicit empty notes and focus from the edit form", async () => {
  await Promise.all([
    import("@/components/form/notes-textarea"),
    import("@/components/form/date-picker"),
    import("@/features/workouts/components/form/workout-focus-combobox"),
  ]);
  const update = vi.spyOn(client, "put").mockResolvedValue({ data: undefined });
  await mount("/workouts/$workoutId/edit", () => (
    <EditWorkoutPage
      user={user}
      exercises={[]}
      workoutId={1}
      workoutsFocus={[]}
      search={{}}
      workout={{
        date: "2026-10-08T12:00:00Z",
        notes: "Old notes",
        workoutFocus: "Upper",
        exercises: [
          { name: "Pushups", sets: [{ reps: 10, setType: "working" }] },
        ],
      }}
    />
  ));
  fireEvent.click(
    await screen.findByRole("button", { name: "Notes" }, { timeout: 20000 }),
  );
  fireEvent.change(screen.getByRole("textbox"), { target: { value: "" } });
  fireEvent.click(
    within(screen.getByRole("dialog")).getByText("Close", {
      selector: "button",
    }),
  );
  fireEvent.click(
    within(screen.getByLabelText("Workout Focus")).getByRole("button"),
  );
  fireEvent.click(screen.getByRole("button", { name: "Save" }));
  await waitFor(() => expect(update).toHaveBeenCalled());
  expect(update.mock.calls[0]?.[0].body).toMatchObject({
    notes: "",
    workoutFocus: "",
  });
}, 60000);

it("saves a catalog exercise selected before a draft reload", async () => {
  const entries = new Map<string, string>();
  const storage = {
    getItem: (key: string) => entries.get(key) ?? null,
    setItem: (key: string, value: string) => {
      entries.set(key, value);
    },
    removeItem: (key: string) => {
      entries.delete(key);
    },
  };
  queryClient.setQueryData(getExerciseCatalogQueryOptions().queryKey, [
    {
      id: "Pushups",
      name: "Pushups",
      equipment: "Body only",
      primaryMuscles: ["Chest"],
      secondaryMuscles: [],
    },
  ]);
  const saved = vi
    .spyOn(client, "post")
    .mockResolvedValue({ data: { message: "Saved" } });
  function Form() {
    const [adding, setAdding] = useState(true);
    const workflow = useNewWorkoutFormWorkflow({
      user,
      exercises: [],
      newWorkoutContext: { focusTemplates: [] },
      draftStorage: createWorkoutDraftStorage(storage),
    });
    return adding ? (
      <AddExerciseScreen
        form={workflow.form}
        exercises={[]}
        onBack={() => setAdding(false)}
        onAddExercise={() => setAdding(false)}
      />
    ) : (
      <>
        <workflow.form.AppField
          name="exercises[0].sets"
          children={(field) => (
            <button
              onClick={() =>
                field.handleChange([{ reps: 10, setType: "working" }])
              }
            >
              Log set
            </button>
          )}
        />
        <button onClick={() => workflow.form.handleSubmit()}>
          Save workout
        </button>
      </>
    );
  }
  const router = await mount("/workouts/new", Form);
  fireEvent.click(
    await screen.findByRole("button", { name: "Browse catalog" }),
  );
  fireEvent.click(await screen.findByRole("button", { name: "Pushups" }));
  fireEvent.click(await screen.findByRole("button", { name: "Log set" }));
  // Leaving and re-entering remounts the real workflow and deserializes storage.
  await act(async () => {
    await router.navigate({
      to: "/workouts/$workoutId",
      params: { workoutId: 1 },
    });
  });
  await act(async () => {
    await router.navigate({ to: "/workouts/new" });
  });
  fireEvent.click(
    await screen.findByRole("button", { name: "Back to workout" }),
  );
  fireEvent.click(await screen.findByRole("button", { name: "Save workout" }));
  await waitFor(() => expect(saved).toHaveBeenCalled());
  expect(saved.mock.calls[0]?.[0].body).toMatchObject({
    exercises: [
      {
        name: "Pushups",
        catalog_id: "Pushups",
        sets: [{ reps: 10, setType: "working" }],
      },
    ],
  });
}, 60000);

it("keeps a selected date across rerenders and starts fresh drafts on clear and save after rollover", async () => {
  vi.spyOn(client, "post").mockResolvedValue({ data: { message: "Saved" } });
  vi.useFakeTimers({ toFake: ["Date"] });
  vi.setSystemTime(new Date("2026-10-08T23:59:00Z"));
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
  function Form() {
    const workflow = useNewWorkoutFormWorkflow({
      user,
      exercises: [],
      newWorkoutContext: { focusTemplates: [] },
      draftStorage,
    });
    return (
      <>
        <workflow.form.Subscribe selector={(state) => state.values.date}>
          {(date) => <output>{date}</output>}
        </workflow.form.Subscribe>
        <button
          onClick={() =>
            workflow.form.setFieldValue("date", "2026-10-01T12:00:00Z")
          }
        >
          Select date
        </button>
        <button onClick={() => workflow.setIsClearDialogOpen(true)}>
          Rerender
        </button>
        <button onClick={workflow.clearForm}>Clear draft</button>
        <button
          onClick={() => {
            workflow.form.setFieldValue("exercises", [
              { name: "Pushups", sets: [{ reps: 10, setType: "working" }] },
            ]);
            void workflow.form.handleSubmit();
          }}
        >
          Save workout
        </button>
      </>
    );
  }
  await mount("/workouts/new", Form);
  expect(await screen.findByRole("status")).toHaveTextContent(
    "2026-10-08T23:59:00.000Z",
  );
  fireEvent.click(screen.getByRole("button", { name: "Select date" }));
  vi.setSystemTime(new Date("2026-10-09T00:01:00Z"));
  fireEvent.click(screen.getByRole("button", { name: "Rerender" }));
  expect(screen.getByRole("status")).toHaveTextContent("2026-10-01T12:00:00Z");
  await act(async () => {
    fireEvent.click(screen.getByRole("button", { name: "Clear draft" }));
  });
  expect(screen.getByRole("status")).toHaveTextContent(
    "2026-10-09T00:01:00.000Z",
  );
  vi.setSystemTime(new Date("2026-10-10T00:01:00Z"));
  fireEvent.click(screen.getByRole("button", { name: "Save workout" }));
  await waitFor(() =>
    expect(screen.getByRole("status")).toHaveTextContent(
      "2026-10-10T00:01:00.000Z",
    ),
  );
});
