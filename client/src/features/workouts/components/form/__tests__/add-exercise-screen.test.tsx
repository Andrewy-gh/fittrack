import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
const catalog = [
  {
    id: "Pushups",
    name: "Pushups",
    equipment: "body only",
    primaryMuscles: ["chest"],
    secondaryMuscles: ["triceps"],
  },
];
vi.mock("@/client/@tanstack/react-query.gen", () => ({
  getExerciseCatalogQueryOptions: () => ({
    queryKey: ["catalog"],
    queryFn: async () => catalog,
  }),
}));
import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type {
  WorkoutCreateWorkoutRequest,
  WorkoutExerciseInput,
} from "@/client";
import { useAppForm } from "@/hooks/form";
import { AddExerciseScreen } from "../add-exercise-screen";

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
    <QueryClientProvider client={new QueryClient()}>
      <form.Subscribe
        selector={(state) => state.values.exercises}
        children={(exercises) => (
          <output data-testid="draft">{JSON.stringify(exercises)}</output>
        )}
      />
      <AddExerciseScreen
        form={form}
        exercises={[{ id: 1, name: "Incline Hammer Curl" }]}
        onAddExercise={vi.fn()}
        onBack={vi.fn()}
      />
    </QueryClientProvider>
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

it("adds explicit catalog metadata to the workout draft", async () => {
  render(<AddExerciseScreenHarness />);
  fireEvent.click(screen.getByRole("button", { name: "Browse catalog" }));
  fireEvent.click(await screen.findByRole("button", { name: "Pushups" }));
  expect(JSON.parse(screen.getByTestId("draft").textContent ?? "[]")).toEqual([
    { name: "Pushups", catalog_id: "Pushups", sets: [] },
  ]);
});
it("keeps a typed name unclassified even when it matches the catalog", () => {
  render(<AddExerciseScreenHarness />);
  fireEvent.change(screen.getByLabelText("Search exercises"), {
    target: { value: "Pushups" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Add" }));
  expect(JSON.parse(screen.getByTestId("draft").textContent ?? "[]")).toEqual([
    { name: "Pushups", sets: [] },
  ]);
});
