import { transferableAbortController } from "node:util";
import { QueryClientProvider, QueryObserver } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
} from "@tanstack/react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { client } from "@/client/client.gen";
import { routeTree } from "@/routeTree.gen";
import { queryClient } from "@/lib/api/api";
import { normalizeQueryOptions } from "@/lib/query-option-adapter";
import {
  exerciseByIdQueryOptions,
  exercisesQueryOptions,
  exerciseMetricsHistoryQueryOptions,
  recentExerciseSetsQueryOptions,
  useDeleteExerciseMutation,
} from "@/features/exercises/api/exercises";
import { trainingAssessmentQueryOptions } from "@/features/analytics/api/training-assessment";
import { muscleContributionsQueryOptions } from "@/features/analytics/api/muscle-contributions";
import { saveLatestWorkoutDraft } from "@/features/chat/utils/chat-session-workout-draft";
import type { AIChatConversation } from "@/features/chat/api/ai-chat";
import type { ChatSessionState } from "@/features/chat/utils/chat-session-types";
import {
  preloadEditWorkoutRouteData,
  preloadNewWorkoutRouteData,
} from "@/features/workouts/pages/workout-route-composition";
import {
  getDemoWorkoutsQueryOptions,
  getDemoNewWorkoutContextQueryOptions,
} from "@/lib/demo-data/query-options";
import {
  contributionDataQueryOptions,
  newWorkoutContextQueryOptions,
  workoutQueryOptions,
  workoutsFocusValuesQueryOptions,
  workoutsQueryOptions,
  useDeleteWorkoutMutation,
  useSaveWorkoutMutation,
  useSaveWorkoutForUserMutation,
  useUpdateWorkoutMutation,
} from "./workouts";

let revision = 0;
let rejectWrite = false;
let rejectRead = false;
let deletedWorkout = false;
let deletedExercise = false;
const reads: string[] = [];
const originalConfig = client.getConfig();
const user = {
  id: "owner",
  displayName: "Owner",
  primaryEmail: null,
  profileImageUrl: null,
  signOut: async () => {},
};
const conversation: AIChatConversation = {
  id: 1,
  created_at: "2026-01-01",
  updated_at: "2026-01-01",
};

function responseFor(path: string) {
  if (path === "/api/workouts") return [{ id: 42, notes: `note-${revision}` }];
  if (path === "/api/exercises")
    return [{ id: 7, name: `exercise-${revision}` }];
  if (path.endsWith("/recent-sets")) return [{ weight: revision, reps: 5 }];
  if (path.endsWith("/metrics-history"))
    return { points: [{ volume: revision }] };
  if (path.endsWith("/contribution-data"))
    return { days: [{ count: revision }] };
  if (path.endsWith("/focus-values")) return [`focus-${revision}`];
  if (path.endsWith("/new-workout-context"))
    return { focusTemplates: [{ focus: `focus-${revision}` }] };
  if (path === "/api/workouts/42")
    return {
      sets: [{ workout_notes: `note-${revision}` }],
      recommendations: [],
    };
  if (path === "/api/exercises/7")
    return {
      exercise: { id: 7, name: `exercise-${revision}` },
      sets: [{ weight: revision }],
    };
  return { workingSets: revision };
}

beforeEach(() => {
  queryClient.clear();
  queryClient.setDefaultOptions({
    queries: { retry: false, staleTime: Infinity },
  });
  // Node Request requires Node's AbortSignal, not jsdom's implementation.
  vi.stubGlobal("AbortController", transferableAbortController().constructor);
  revision = 0;
  rejectWrite = false;
  rejectRead = false;
  deletedWorkout = false;
  deletedExercise = false;
  reads.length = 0;
  client.setConfig({
    fetch: async (input) => {
      if (!(input instanceof Request))
        throw new Error("Expected generated Request");
      const path = new URL(input.url).pathname;
      if (input.method !== "GET") {
        if (rejectWrite)
          return Response.json({ message: "Write rejected" }, { status: 409 });
        revision++;
        if (input.method === "DELETE" && path === "/api/workouts/42")
          deletedWorkout = true;
        if (input.method === "DELETE" && path === "/api/exercises/7")
          deletedExercise = true;
        return Response.json(
          path.includes("latest-workout-draft")
            ? {
                conversation: {
                  ...conversation,
                  latest_workout_draft_status: {
                    is_saved: true,
                    saved_workout_id: 42,
                  },
                },
                workout_id: 42,
              }
            : { message: "saved" },
        );
      }
      reads.push(path);
      if (
        (deletedWorkout && path === "/api/workouts/42") ||
        (deletedExercise &&
          (path === "/api/exercises/7" ||
            path === "/api/exercises/7/metrics-history"))
      )
        return Response.json({ message: "Missing resource" }, { status: 404 });
      if (deletedExercise && path === "/api/exercises/7/recent-sets")
        return Response.json([]);
      if (rejectRead)
        return Response.json({ message: "Read unavailable" }, { status: 503 });
      return Response.json(responseFor(path));
    },
  });
});
afterEach(() => {
  queryClient.clear();
  client.setConfig(originalConfig);
  vi.unstubAllGlobals();
});

function mountMutations() {
  return renderHook(
    ({ userId }) => ({
      create: useSaveWorkoutMutation(userId),
      update: useUpdateWorkoutMutation(userId),
      delete: useDeleteWorkoutMutation(userId),
      deleteExercise: useDeleteExerciseMutation(userId),
    }),
    {
      initialProps: { userId: "owner" },
      wrapper: ({ children }) => (
        <QueryClientProvider client={queryClient}>
          {children}
        </QueryClientProvider>
      ),
    },
  );
}
function chatState() {
  let saved: AIChatConversation | null = conversation;
  let saving = false;
  const state: ChatSessionState = {
    setConversation: (value) => {
      saved = typeof value === "function" ? value(saved) : value;
    },
    setIsSavingWorkoutDraft: (value) => {
      saving = typeof value === "function" ? value(saving) : value;
    },
    setMessages: () => {},
    setPrompt: () => {},
    setIsLoadingConversation: () => {},
    setLoadError: () => {},
    setLatestWorkoutDraftMessageId: () => {},
  };
  return { state, saved: () => saved, saving: () => saving };
}
const body = {
  date: "2026-01-01T12:00:00Z",
  exercises: [
    {
      name: "Press",
      sets: [{ reps: 5, weight: 10, setType: "working" as const }],
    },
  ],
};
const operations = [
  "create",
  "update",
  "delete",
  "deleteExercise",
  "chat",
] as const;
type Operation = (typeof operations)[number];
async function write(
  operation: Operation,
  hooks: ReturnType<typeof mountMutations>,
  chat: ReturnType<typeof chatState>,
) {
  await act(async () => {
    if (operation === "create")
      await hooks.result.current.create.mutateAsync({ body });
    else if (operation === "update")
      await hooks.result.current.update.mutateAsync({ path: { id: 42 }, body });
    else if (operation === "delete")
      await hooks.result.current.delete.mutateAsync({ path: { id: 42 } });
    else if (operation === "deleteExercise")
      await hooks.result.current.deleteExercise.mutateAsync({
        path: { id: 7 },
      });
    else
      await saveLatestWorkoutDraft({
        conversation,
        state: chat.state,
        userId: "owner",
      });
  });
}

function dependents() {
  return [
    normalizeQueryOptions(workoutsQueryOptions()),
    normalizeQueryOptions(workoutQueryOptions(42)),
    normalizeQueryOptions(exercisesQueryOptions()),
    normalizeQueryOptions(exerciseByIdQueryOptions(7)),
    normalizeQueryOptions(recentExerciseSetsQueryOptions(7)),
    normalizeQueryOptions(exerciseMetricsHistoryQueryOptions(7)),
    normalizeQueryOptions(contributionDataQueryOptions()),
    normalizeQueryOptions(workoutsFocusValuesQueryOptions()),
    normalizeQueryOptions(newWorkoutContextQueryOptions()),
    normalizeQueryOptions(
      trainingAssessmentQueryOptions("owner", 7, "2026-01-01", "UTC"),
    ),
    normalizeQueryOptions(
      trainingAssessmentQueryOptions("owner", 7, "2026-02-01", "Europe/Berlin"),
    ),
    normalizeQueryOptions(
      muscleContributionsQueryOptions("owner", "2026-01-01", "UTC"),
    ),
  ].map(({ queryKey, queryFn }) => ({ queryKey, queryFn }));
}

describe("training mutations followed by authoritative reads", () => {
  for (const operation of operations) {
    it(`${operation} refreshes active data and every warmed inactive dependent, without fetching inactive/demo/other-user data eagerly`, async () => {
      const hooks = mountMutations();
      const chat = chatState();
      const options = dependents();
      const old = await Promise.all(
        options.map((option) =>
          queryClient.query<unknown>({ ...option, staleTime: Infinity }),
        ),
      );
      const demo = normalizeQueryOptions(getDemoWorkoutsQueryOptions());
      queryClient.setQueryData<unknown>(demo.queryKey, [
        { id: 99, notes: "demo only" },
      ]);
      const other = normalizeQueryOptions(
        trainingAssessmentQueryOptions("other-user", 7, "2026-01-01", "UTC"),
      );
      queryClient.setQueryData<unknown>(other.queryKey, { workingSets: 999 });
      queryClient.setQueryData(["unrelated"], "unchanged");
      const observer = new QueryObserver(
        queryClient,
        contributionDataQueryOptions(),
      );
      const stop = observer.subscribe(() => {});
      reads.length = 0;
      await write(operation, hooks, chat);
      expect(observer.getCurrentResult().data).toEqual({
        days: [{ count: 1 }],
      });
      expect(reads).toEqual(["/api/workouts/contribution-data"]);
      for (let i = 0; i < options.length; i++) {
        if (
          (operation === "delete" && i === 1) ||
          (operation === "deleteExercise" && (i === 3 || i === 5))
        ) {
          await expect(
            queryClient.query<unknown>({ ...options[i], staleTime: Infinity }),
          ).rejects.toMatchObject({ message: "Missing resource" });
          continue;
        }
        const fresh = await queryClient.query<unknown>({
          ...options[i],
          staleTime: Infinity,
        });
        expect(fresh, `dependent ${i}`).not.toEqual(old[i]);
      }
      expect(queryClient.getQueryData(demo.queryKey)).toEqual([
        { id: 99, notes: "demo only" },
      ]);
      expect(queryClient.getQueryData(other.queryKey)).toEqual({
        workingSets: 999,
      });
      expect(queryClient.getQueryState(other.queryKey)?.isInvalidated).toBe(
        false,
      );
      expect(queryClient.getQueryData(["unrelated"])).toBe("unchanged");
      if (operation === "chat")
        expect(chat.saved()?.latest_workout_draft_status?.is_saved).toBe(true);
      stop();
    });
    it(`${operation} rejection leaves cached reads and saved status unchanged`, async () => {
      const hooks = mountMutations();
      const chat = chatState();
      const before = await queryClient.query(workoutsQueryOptions());
      reads.length = 0;
      rejectWrite = true;
      if (operation === "chat") await write(operation, hooks, chat);
      else
        await expect(write(operation, hooks, chat)).rejects.toMatchObject({
          message: "Write rejected",
        });
      expect(
        await queryClient.query({
          ...workoutsQueryOptions(),
          staleTime: Infinity,
        }),
      ).toEqual(before);
      expect(reads).toEqual([]);
      expect(chat.saved()?.latest_workout_draft_status?.is_saved).not.toBe(
        true,
      );
      expect(chat.saving()).toBe(false);
    });
  }

  it("demo save refreshes warmed local routes without touching authenticated caches or transport", async () => {
    const apiBefore = await queryClient.query(workoutsQueryOptions());
    await preloadNewWorkoutRouteData({ queryClient, user: null });
    const demoBefore = await queryClient.query(getDemoWorkoutsQueryOptions());
    reads.length = 0;
    const demo = renderHook(() => useSaveWorkoutForUserMutation(null), {
      wrapper: ({ children }) => (
        <QueryClientProvider client={queryClient}>
          {children}
        </QueryClientProvider>
      ),
    });
    await act(async () => {
      await demo.result.current.mutateAsync({
        body: { ...body, workoutFocus: "Demo-only-focus" },
      });
    });
    await preloadNewWorkoutRouteData({ queryClient, user: null });
    expect(await queryClient.query(getDemoWorkoutsQueryOptions())).toHaveLength(
      demoBefore.length + 1,
    );
    expect(
      (
        await queryClient.query(getDemoNewWorkoutContextQueryOptions())
      ).focusTemplates?.map((template) => template.focus),
    ).toContain("Demo-only-focus");
    expect(queryClient.getQueryData(workoutsQueryOptions().queryKey)).toEqual(
      apiBefore,
    );
    expect(
      queryClient.getQueryState(workoutsQueryOptions().queryKey)?.isInvalidated,
    ).toBe(false);
    expect(reads).toEqual([]);
    expect(revision).toBe(0);
  });

  for (const destination of [
    {
      href: "/workouts",
      queryKey: workoutsQueryOptions().queryKey,
      expected: [{ id: 42, notes: "note-1" }],
    },
    {
      href: "/analytics",
      queryKey: contributionDataQueryOptions().queryKey,
      expected: { days: [{ count: 1 }] },
    },
    {
      href: "/exercises",
      queryKey: exercisesQueryOptions().queryKey,
      expected: [{ id: 7, name: "exercise-1" }],
    },
    {
      href: "/exercises/7",
      queryKey: exerciseByIdQueryOptions(7).queryKey,
      expected: {
        exercise: { id: 7, name: "exercise-1" },
        sets: [{ weight: 1 }],
      },
    },
    {
      href: "/workouts/42",
      queryKey: workoutQueryOptions(42).queryKey,
      expected: { sets: [{ workout_notes: "note-1" }], recommendations: [] },
    },
  ])
    it(`revisiting production route ${destination.href} reads changed state after chat save`, async () => {
      const hooks = mountMutations();
      const router = createRouter({
        routeTree,
        context: { queryClient, user },
        history: createMemoryHistory({ initialEntries: ["/"] }),
        defaultPreloadStaleTime: 0,
      });
      await router.load();
      await router.navigate({ href: destination.href });
      await waitFor(() => expect(queryClient.isFetching()).toBe(0));
      expect(queryClient.getQueryData(destination.queryKey)).toBeDefined();
      await router.navigate({ to: "/" });
      await write("chat", hooks, chatState());
      await router.navigate({ href: destination.href });
      await waitFor(() =>
        expect(queryClient.getQueryData(destination.queryKey)).toEqual(
          destination.expected,
        ),
      );
    });

  it("new/edit route navigation refreshes invalidated inactive data", async () => {
    const hooks = mountMutations();
    const context = { queryClient, user };
    const root = createRootRoute();
    const home = createRoute({ getParentRoute: () => root, path: "/" });
    const newRoute = createRoute({
      getParentRoute: () => root,
      path: "/workouts/new",
      loader: () => preloadNewWorkoutRouteData(context),
    });
    const editRoute = createRoute({
      getParentRoute: () => root,
      path: "/workouts/$workoutId/edit",
      loader: () => preloadEditWorkoutRouteData({ ...context, workoutId: 42 }),
    });
    const router = createRouter({
      routeTree: root.addChildren([home, newRoute, editRoute]),
      history: createMemoryHistory({ initialEntries: ["/"] }),
      defaultPreloadStaleTime: 0,
    });
    await router.load();
    await router.navigate({ to: "/workouts/new" });
    await router.navigate({
      to: "/workouts/$workoutId/edit",
      params: { workoutId: 42 },
    });
    await router.navigate({ to: "/" });
    await write("update", hooks, chatState());
    await router.navigate({ to: "/workouts/new" });
    await router.navigate({
      to: "/workouts/$workoutId/edit",
      params: { workoutId: 42 },
    });
    // Cached routes revalidate in the background: navigation completion is not
    // a promise that their loaders have finished refreshing Query data.
    await waitFor(() => {
      expect(
        queryClient.getQueryData(exercisesQueryOptions().queryKey),
      ).toEqual([{ id: 7, name: "exercise-1" }]);
      expect(
        queryClient.getQueryData(workoutQueryOptions(42).queryKey),
      ).toEqual({
        sets: [{ workout_notes: "note-1" }],
        recommendations: [],
      });
      expect(
        queryClient.getQueryData(newWorkoutContextQueryOptions().queryKey),
      ).toEqual({ focusTemplates: [{ focus: "focus-1" }] });
    });
  });

  for (const failFirst of [false, true])
    it(`independent concurrent writes complete out of order (first rejected: ${failFirst})`, async () => {
      const hooks = mountMutations();
      await queryClient.query(contributionDataQueryOptions());
      const observer = new QueryObserver(
        queryClient,
        contributionDataQueryOptions(),
      );
      const stop = observer.subscribe(() => {});
      const transport = client.getConfig().fetch;
      if (!transport) throw new Error("Missing test transport");
      let releaseFirst = () => {};
      let releaseSecond = () => {};
      const firstGate = new Promise<void>((resolve) => {
        releaseFirst = resolve;
      });
      const secondGate = new Promise<void>((resolve) => {
        releaseSecond = resolve;
      });
      client.setConfig({
        fetch: async (input, init) => {
          if (input instanceof Request && input.method === "POST") {
            await firstGate;
            if (failFirst)
              return Response.json(
                { message: "Write rejected" },
                { status: 409 },
              );
          }
          if (input instanceof Request && input.method === "PUT")
            await secondGate;
          return transport(input, init);
        },
      });
      let first: Promise<unknown> = Promise.resolve();
      let second: Promise<unknown> = Promise.resolve();
      await act(async () => {
        first = hooks.result.current.create
          .mutateAsync({ body })
          .catch((error: unknown) => error);
        second = hooks.result.current.update.mutateAsync({
          path: { id: 42 },
          body,
        });
      });
      releaseSecond();
      await act(async () => {
        await second;
      });
      expect(observer.getCurrentResult().data).toEqual({
        days: [{ count: 1 }],
      });
      expect(hooks.result.current.create.isPending).toBe(true);
      releaseFirst();
      await act(async () => {
        await first;
      });
      await waitFor(() =>
        expect(hooks.result.current.create.isPending).toBe(false),
      );
      expect(hooks.result.current.update.isSuccess).toBe(true);
      expect(hooks.result.current.create.isError).toBe(failFirst);
      expect(observer.getCurrentResult().data).toEqual({
        days: [{ count: failFirst ? 1 : 2 }],
      });
      expect(
        await queryClient.query({
          ...workoutsQueryOptions(),
          staleTime: Infinity,
        }),
      ).toEqual([{ id: 42, notes: failFirst ? "note-1" : "note-2" }]);
      stop();
    });

  it("keeps a pending mutation's analytics scoped to its starting user after the hook changes user", async () => {
    const hooks = mountMutations();
    const ownerOptions = trainingAssessmentQueryOptions(
      "owner",
      7,
      "2026-01-01",
      "UTC",
    );
    const otherOptions = trainingAssessmentQueryOptions(
      "other-user",
      7,
      "2026-01-01",
      "UTC",
    );
    await queryClient.query(ownerOptions);
    queryClient.setQueryData<unknown>(otherOptions.queryKey, {
      workingSets: 999,
    });
    let release = () => {};
    const gate = new Promise<void>((resolve) => {
      release = resolve;
    });
    const transport = client.getConfig().fetch;
    if (!transport) throw new Error("Missing test transport");
    client.setConfig({
      fetch: async (input, init) => {
        if (input instanceof Request && input.method === "POST") await gate;
        return transport(input, init);
      },
    });
    let pending: Promise<unknown> = Promise.resolve();
    await act(async () => {
      pending = hooks.result.current.create.mutateAsync({ body });
    });
    hooks.rerender({ userId: "other-user" });
    const observer = new QueryObserver(queryClient, otherOptions);
    const stop = observer.subscribe(() => {});
    reads.length = 0;
    release();
    await act(async () => {
      await pending;
    });
    expect(observer.getCurrentResult().data).toEqual({ workingSets: 999 });
    expect(reads).toEqual([]);
    expect(
      queryClient.getQueryState(ownerOptions.queryKey)?.isInvalidated,
    ).toBe(true);
    stop();
  });

  it("a dependent read failure does not turn a persisted chat save into a failed save", async () => {
    const hooks = mountMutations();
    const chat = chatState();
    await queryClient.query(contributionDataQueryOptions());
    const observer = new QueryObserver(
      queryClient,
      contributionDataQueryOptions(),
    );
    const stop = observer.subscribe(() => {});
    rejectRead = true;
    await write("chat", hooks, chat);
    expect(chat.saved()?.latest_workout_draft_status?.is_saved).toBe(true);
    expect(chat.saving()).toBe(false);
    expect(observer.getCurrentResult().isError).toBe(true);
    rejectRead = false;
    expect(await queryClient.query(contributionDataQueryOptions())).toEqual({
      days: [{ count: 1 }],
    });
    stop();
  });
});
