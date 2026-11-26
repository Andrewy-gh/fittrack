import { useField } from "@tanstack/react-form";
import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { WorkoutNewWorkoutContextResponse } from "@/client";
import {
  createWorkoutDraftStorage,
  type WorkoutDraftStorage,
} from "@/lib/local-storage";
import { saveAIWorkoutDraftToWorkoutForm } from "@/features/chat/utils/ai-workout-draft";
import { featureFlags } from "@/lib/feature-flags";

vi.mock("@/lib/feature-flags", () => ({
  featureFlags: { trainingGuidance: false },
}));

const {
  mockApiMutateAsync,
  mockDemoMutateAsync,
  mockFetchQuery,
  mockNavigate,
  mockToastSuccess,
} = vi.hoisted(() => ({
  mockApiMutateAsync: vi.fn(),
  mockDemoMutateAsync: vi.fn(),
  mockFetchQuery: vi.fn(),
  mockNavigate: vi.fn(),
  mockToastSuccess: vi.fn(),
}));

vi.mock("@tanstack/react-router", () => ({
  useNavigate: () => mockNavigate,
}));

vi.mock("@/features/workouts/api/workouts", () => ({
  useSaveWorkoutForUserMutation: (user: unknown) => ({
    mutateAsync: user ? mockApiMutateAsync : mockDemoMutateAsync,
  }),
}));

vi.mock("@/lib/api/api", () => ({
  queryClient: {
    fetchQuery: mockFetchQuery,
  },
}));

vi.mock("@/features/workouts/api/workout-query-options", () => ({
  getWorkoutByIdQueryOptions: (_user: unknown, workoutId: number) => ({
    queryKey: ["workout", workoutId],
  }),
}));

vi.mock("sonner", () => ({
  toast: {
    success: mockToastSuccess,
  },
}));

import { useNewWorkoutFormWorkflow } from "@/features/workouts/hooks/use-new-workout-form-workflow";

function createDraftStorage(
  loadValue: ReturnType<WorkoutDraftStorage["load"]> = null,
): WorkoutDraftStorage {
  return {
    save: vi.fn(),
    load: vi.fn(() => loadValue),
    clear: vi.fn(),
  };
}

const newWorkoutContext: WorkoutNewWorkoutContextResponse = {
  latestWorkoutNote: undefined,
  focusTemplates: [
    {
      date: "2026-06-12",
      focus: "Push",
      workoutId: 42,
    },
  ],
};

describe("useNewWorkoutFormWorkflow", () => {
  it.each([true, false])(
    "saves original chat text, edits and undo across reload (historical context: %s)",
    async (withContext) => {
      const entries = new Map<string, string>();
      const storage = createWorkoutDraftStorage({
        getItem: (key) => entries.get(key) ?? null,
        setItem: (key, value) => {
          entries.set(key, value);
        },
        removeItem: (key) => {
          entries.delete(key);
        },
      });
      const snapshot = {
        exerciseName: "Pushups",
        recommendation: {
          exerciseId: 77,
          readiness: "normal" as const,
          explanation: "Original context",
          policyVersion: "exercise-plan-v2",
          plan: { sets: 3, reps: 8, weight: null },
        },
      };
      const draft = {
        date: "2026-09-11T12:00:00-04:00",
        // Go's chat normalization preserves edge U+FEFF; ECMAScript trim does not.
        notes: "\uFEFFOriginal note\uFEFF",
        workoutFocus: "\uFEFFPush\uFEFF",
        exercises: [
          {
            name: "Pushups",
            catalog_id: "Pushups",
            sets: [{ reps: 8, setType: "working" as const }],
          },
        ],
        ...(withContext && { recommendations: [snapshot] }),
      };
      const original = structuredClone(draft);
      expect(
        saveAIWorkoutDraftToWorkoutForm(draft, "user-1", storage),
      ).not.toBeNull();
      expect(storage.load("user-2")).toBeNull();
      const mountWorkflow = () =>
        renderHook(() => {
          const workflow = useNewWorkoutFormWorkflow({
            user: {
              id: "user-1",
              displayName: "Owner",
              primaryEmail: null,
              profileImageUrl: null,
              signOut: async () => {},
            },
            // A different owned ID must not be substituted for snapshot 77.
            exercises: [{ id: 88, name: "Pushups" }],
            newWorkoutContext,
            draftStorage: storage,
          });
          const notes = useField({ form: workflow.form, name: "notes" });
          const focus = useField({ form: workflow.form, name: "workoutFocus" });
          return { ...workflow, notes, focus };
        });
      const { result, unmount } = mountWorkflow();
      if (withContext) {
        await act(async () => {
          result.current.recordRecommendation("Pushups", {
            ...snapshot.recommendation,
            exerciseId: 88,
            explanation: "Late response",
          });
        });
      }
      await act(async () => {
        await result.current.form.handleSubmit();
      });
      expect(mockApiMutateAsync.mock.calls.at(-1)?.[0].body).toEqual(original);
      await act(async () => {
        // Removing the edge characters changes what is saved, even when the
        // visible words stay the same. Leave the other text field untouched.
        result.current.notes.handleChange("Original note");
      });
      await act(async () => {
        await result.current.form.handleSubmit();
      });
      expect(mockApiMutateAsync.mock.calls.at(-1)?.[0].body).toEqual({
        ...original,
        notes: "Original note",
        recommendations: undefined,
      });
      unmount();
      const reloaded = mountWorkflow();
      await act(async () => {
        await reloaded.result.current.form.handleSubmit();
      });
      expect(mockApiMutateAsync.mock.calls.at(-1)?.[0].body).toEqual({
        ...original,
        notes: "Original note",
        recommendations: undefined,
      });
      await act(async () => {
        reloaded.result.current.notes.handleChange(original.notes);
        reloaded.result.current.focus.handleChange(" Pull ");
      });
      await act(async () => {
        await reloaded.result.current.form.handleSubmit();
      });
      expect(mockApiMutateAsync.mock.calls.at(-1)?.[0].body).toEqual({
        ...original,
        workoutFocus: "Pull",
        recommendations: undefined,
      });
      await act(async () => {
        reloaded.result.current.focus.handleChange(original.workoutFocus);
        await reloaded.result.current.form.handleSubmit();
      });
      expect(mockApiMutateAsync.mock.calls.at(-1)?.[0].body).toEqual(original);
      for (const changedExercises of [
        [{ ...draft.exercises[0], catalog_id: "Dumbbell_Bench_Press" }],
        [
          {
            ...draft.exercises[0],
            sets: [{ reps: 9, setType: "working" as const }],
          },
        ],
        [
          {
            ...draft.exercises[0],
            sets: [{ reps: 8, setType: "working" as const, weight: 0 }],
          },
        ],
      ]) {
        await act(async () => {
          reloaded.result.current.form.setFieldValue(
            "exercises",
            changedExercises,
          );
        });
        await act(async () => {
          await reloaded.result.current.form.handleSubmit();
        });
        expect(
          mockApiMutateAsync.mock.calls.at(-1)?.[0].body.recommendations,
        ).toBeUndefined();
        await act(async () => {
          reloaded.result.current.form.setFieldValue(
            "exercises",
            draft.exercises,
          );
        });
        await act(async () => {
          await reloaded.result.current.form.handleSubmit();
        });
        expect(
          mockApiMutateAsync.mock.calls.at(-1)?.[0].body.recommendations,
        ).toEqual(original.recommendations);
      }
      expect(draft).toEqual(original);
    },
  );
  it("records guidance without creating or overwriting actual sets and discards it on repeat", async () => {
    const draftStorage = createDraftStorage({
      date: "2026-09-11T12:00:00Z",
      exercises: [
        { name: "Press", sets: [{ reps: 5, weight: 50, setType: "working" }] },
      ],
    });
    const { result } = renderHook(() =>
      useNewWorkoutFormWorkflow({
        user: null,
        exercises: [{ id: 1, name: "Press" }],
        newWorkoutContext,
        draftStorage,
      }),
    );
    const actualSets = structuredClone(
      result.current.form.state.values.exercises,
    );
    await act(async () => {
      result.current.recordRecommendation("Press", {
        exerciseId: 1,
        readiness: "sluggish",
        plan: { sets: 2, reps: 6, weight: 50 },
        explanation: "Reduced work",
        policyVersion: "exercise-plan-v2",
      });
    });
    expect(result.current.form.state.values.exercises).toEqual(actualSets);
    expect(result.current.form.state.values.recommendations).toMatchObject([
      {
        recommendation: { readiness: "sluggish", plan: { sets: 2, reps: 6 } },
      },
    ]);
    await act(async () => {
      await result.current.form.handleSubmit();
    });
    expect(mockDemoMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({
        body: expect.objectContaining({
          exercises: actualSets,
          recommendations: expect.arrayContaining([
            expect.objectContaining({ exerciseName: "Press" }),
          ]),
        }),
      }),
      expect.anything(),
    );
    mockFetchQuery.mockResolvedValue({ sets: [], recommendations: [] });
    await act(async () => {
      await result.current.repeatWorkout(42);
    });
    await act(async () => {
      await result.current.replaceDraftWithPendingTemplate();
    });
    expect(result.current.form.state.values.recommendations).toBeUndefined();
  });
  beforeEach(() => {
    featureFlags.trainingGuidance = false;
    mockApiMutateAsync.mockReset();
    mockDemoMutateAsync.mockReset();
    mockFetchQuery.mockReset();
    mockNavigate.mockReset();
    mockToastSuccess.mockReset();
  });

  it.each([false, true])(
    "saves an authenticated workout with training guidance enabled=%s",
    async (enabled) => {
      featureFlags.trainingGuidance = enabled;
      const draftStorage = createDraftStorage({
        date: "2026-08-12T12:00:00Z",
        exercises: [
          {
            name: "New press",
            sets: [{ reps: 5, weight: 50, setType: "working" }],
          },
        ],
      });
      mockFetchQuery.mockResolvedValue([
        { id: 1, name: "Alphabetical first" },
        { id: 9, name: "New press" },
      ]);
      mockApiMutateAsync.mockImplementation(async (_variables, options) => {
        options.onSuccess();
      });

      const { result } = renderHook(() =>
        useNewWorkoutFormWorkflow({
          user: { id: "user-1" } as any,
          exercises: [],
          newWorkoutContext,
          draftStorage,
        }),
      );

      await act(async () => {
        await result.current.form.handleSubmit();
      });

      expect(mockApiMutateAsync).toHaveBeenCalledWith(
        expect.objectContaining({
          body: expect.objectContaining({
            notes: undefined,
            workoutFocus: undefined,
          }),
        }),
        expect.any(Object),
      );
      expect(mockDemoMutateAsync).not.toHaveBeenCalled();
      expect(draftStorage.clear).toHaveBeenCalledWith("user-1");
      expect(mockNavigate).toHaveBeenCalledWith({ search: {} });
      if (!enabled) {
        expect(mockToastSuccess).toHaveBeenCalledWith(
          "Workout saved successfully",
        );
        return;
      }
      expect(mockToastSuccess).toHaveBeenCalledWith(
        "Workout saved successfully",
        {
          action: { label: "View assessment", onClick: expect.any(Function) },
        },
      );
      await act(async () => {
        await mockToastSuccess.mock.calls[0][1].action.onClick();
      });
      expect(mockNavigate).toHaveBeenCalledWith({
        to: "/analytics",
        hash: "training-assessment",
        search: { exerciseId: 9, assessmentWeek: "2026-08-10" },
      });
    },
  );

  it("asks before replacing an existing draft with a focus-area template", async () => {
    const draftStorage = createDraftStorage({
      date: "2026-06-13T10:00:00.000Z",
      notes: "Keep this",
      exercises: [],
      workoutFocus: "",
    });
    mockFetchQuery.mockResolvedValue({
      recommendations: [],
      sets: [
        {
          exercise_id: 1,
          exercise_name: "Bench Press",
          exercise_order: 0,
          reps: 5,
          set_id: 10,
          set_order: 0,
          set_type: "working",
          volume: 500,
          weight: 100,
          workout_date: "2026-06-12",
          workout_id: 42,
          workout_notes: "",
          workout_focus: "Push",
        },
      ],
    });

    const { result } = renderHook(() =>
      useNewWorkoutFormWorkflow({
        user: null,
        exercises: [{ id: 1, name: "Bench Press" }],
        newWorkoutContext,
        draftStorage,
      }),
    );

    await act(async () => {
      await result.current.repeatWorkout(42);
    });

    expect(result.current.pendingTemplateWorkoutId).toBe(42);
    expect(mockFetchQuery).not.toHaveBeenCalled();

    await act(async () => {
      await result.current.replaceDraftWithPendingTemplate();
    });

    expect(draftStorage.save).toHaveBeenCalledWith(
      expect.objectContaining({
        workoutFocus: "Push",
        exercises: [
          {
            name: "Bench Press",
            sets: [{ reps: 5, setType: "working", weight: 100 }],
          },
        ],
      }),
      undefined,
    );
    expect(mockNavigate).toHaveBeenCalledWith({ search: {} });
    expect(mockToastSuccess).toHaveBeenCalledWith("Loaded workout structure");
  });

  it("clears the current draft and closes the clear dialog", () => {
    const draftStorage = createDraftStorage();
    const { result } = renderHook(() =>
      useNewWorkoutFormWorkflow({
        user: null,
        exercises: [],
        newWorkoutContext,
        draftStorage,
      }),
    );

    act(() => {
      result.current.setIsClearDialogOpen(true);
    });
    act(() => {
      result.current.clearForm();
    });

    expect(draftStorage.clear).toHaveBeenCalledWith(undefined);
    expect(mockNavigate).toHaveBeenCalledWith({ search: {} });
    expect(result.current.isClearDialogOpen).toBe(false);
  });
});
