import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type {
  WorkoutCreateWorkoutRequest,
  WorkoutExerciseInput,
} from "@/client";
import { useAppForm } from "@/hooks/form";
import { AddExerciseScreen } from "../add-exercise-screen";

const navigateMock = vi.fn();

vi.mock("@tanstack/react-router", async (importOriginal) => {
  const actual =
    await importOriginal<typeof import("@tanstack/react-router")>();

  return {
    ...actual,
    useNavigate: () => navigateMock,
  };
});

vi.mock("@/routes/_layout/workouts/new", () => ({
  Route: {
    fullPath: "/_layout/workouts/new",
  },
}));

function AddExerciseScreenHarness() {
  const defaultValues: WorkoutCreateWorkoutRequest = {
    date: new Date().toISOString(),
    exercises: [] as WorkoutExerciseInput[],
    notes: "",
    workoutFocus: "",
  };
  const form = useAppForm({
    defaultValues,
    onSubmit: async () => undefined,
  });

  return (
    <AddExerciseScreen
      form={form}
      exercises={[{ id: 1, name: "Incline Hammer Curl" }]}
      onBack={vi.fn()}
    />
  );
}

describe("AddExerciseScreen", () => {
  it("allows creating an exercise when the search is only a partial match", () => {
    render(<AddExerciseScreenHarness />);

    fireEvent.change(screen.getByLabelText("Search exercises"), {
      target: { value: "Hammer Curl" },
    });

    expect(screen.getByRole("button", { name: "Add" })).toBeInTheDocument();
    expect(screen.getByText("Incline Hammer Curl")).toBeInTheDocument();
  });

  it("does not show Add when the search exactly matches an existing exercise", () => {
    render(<AddExerciseScreenHarness />);

    fireEvent.change(screen.getByLabelText("Search exercises"), {
      target: { value: " incline hammer curl " },
    });

    expect(
      screen.queryByRole("button", { name: "Add" }),
    ).not.toBeInTheDocument();
  });
});
