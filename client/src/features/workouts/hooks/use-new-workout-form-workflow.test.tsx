import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { WorkoutNewWorkoutContextResponse } from "@/client";
import type { WorkoutDraftStorage } from "@/lib/local-storage";

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
  beforeEach(() => {
    mockApiMutateAsync.mockReset();
    mockDemoMutateAsync.mockReset();
    mockFetchQuery.mockReset();
    mockNavigate.mockReset();
    mockToastSuccess.mockReset();
  });

  it("uses the authenticated save mutation and clears the saved draft after submit", async () => {
    const draftStorage = createDraftStorage();
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
    expect(mockToastSuccess).toHaveBeenCalledWith("Workout saved successfully");
  });

  it("asks before replacing an existing draft with a focus-area template", async () => {
    const draftStorage = createDraftStorage({
      date: "2026-06-13T10:00:00.000Z",
      notes: "Keep this",
      exercises: [],
      workoutFocus: "",
    });
    mockFetchQuery.mockResolvedValue([
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
    ]);

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
