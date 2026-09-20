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
import { getExerciseCatalogQueryKey } from "@/client/@tanstack/react-query.gen";
import { useState } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type {
  WorkoutCreateWorkoutRequest,
  WorkoutExerciseInput,
} from "@/client";
import { useAppForm } from "@/hooks/form";
import { AddExerciseScreen } from "../add-exercise-screen";

function AddExerciseScreenHarness() {
  const [closed, setClosed] = useState(false);
  const [queryClient] = useState(() => {
    const client = new QueryClient({
      defaultOptions: { queries: { staleTime: Infinity } },
    });
    client.setQueryData(getExerciseCatalogQueryKey(), catalog);
    return client;
  });
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

  if (closed) return <h1>Workout editor</h1>;

  return (
    <QueryClientProvider client={queryClient}>
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
        onBack={() => setClosed(true)}
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

it("selects an existing custom exercise through the shared list", () => {
  render(<AddExerciseScreenHarness />);
  fireEvent.click(screen.getByRole("button", { name: "Incline Hammer Curl" }));
  expect(JSON.parse(screen.getByTestId("draft").textContent ?? "[]")).toEqual([
    { name: "Incline Hammer Curl", sets: [] },
  ]);
});

it("adds explicit catalog metadata to the workout draft", async () => {
  render(<AddExerciseScreenHarness />);
  fireEvent.click(screen.getByRole("button", { name: "Browse catalog" }));
  fireEvent.click(await screen.findByRole("button", { name: "Pushups" }));
  expect(JSON.parse(screen.getByTestId("draft").textContent ?? "[]")).toEqual([
    { name: "Pushups", catalog_id: "Pushups", sets: [] },
  ]);
});
it("uses Back to leave the catalog before leaving exercise selection", async () => {
  render(<AddExerciseScreenHarness />);
  fireEvent.change(screen.getByLabelText("Search exercises"), {
    target: { value: "Hammer" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Browse catalog" }));
  await screen.findByLabelText("Search catalog");
  expect(
    screen.queryByRole("button", { name: "Hide catalog" }),
  ).not.toBeInTheDocument();
  fireEvent.click(screen.getByRole("button", { name: "Back to exercises" }));
  expect(screen.getByLabelText("Search exercises")).toHaveValue("Hammer");
  expect(screen.queryByLabelText("Search catalog")).not.toBeInTheDocument();
  fireEvent.click(screen.getByRole("button", { name: "Back to workout" }));
  expect(
    screen.getByRole("heading", { name: "Workout editor" }),
  ).toBeInTheDocument();
});

it("expands catalog metadata without adding an exercise", async () => {
  render(<AddExerciseScreenHarness />);
  fireEvent.click(screen.getByRole("button", { name: "Browse catalog" }));
  const details = await screen.findByRole("button", {
    name: "Details for Pushups",
  });
  expect(details).toHaveAttribute("aria-expanded", "false");
  expect(screen.queryByText("Equipment: body only")).not.toBeInTheDocument();
  fireEvent.click(details);
  expect(details).toHaveAttribute("aria-expanded", "true");
  expect(screen.getByText("Equipment: body only")).toBeVisible();
  expect(screen.getByText("Primary: chest")).toBeVisible();
  expect(screen.getByText("Secondary: triceps")).toBeVisible();
  expect(screen.getByTestId("draft")).toHaveTextContent("[]");
  fireEvent.click(details);
  expect(screen.queryByText("Equipment: body only")).not.toBeInTheDocument();
  fireEvent.click(screen.getByRole("button", { name: "Pushups" }));
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
