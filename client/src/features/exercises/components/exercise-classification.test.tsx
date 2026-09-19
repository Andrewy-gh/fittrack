import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import {
  QueryClient,
  QueryClientProvider,
  useQuery,
} from "@tanstack/react-query";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { ExercisecatalogEntry } from "@/client";
import { ExerciseClassification } from "./exercise-classification";
import { ExerciseCatalogPicker } from "./exercise-catalog-picker";

const mocks = vi.hoisted(() => ({ load: vi.fn(), save: vi.fn() }));
vi.mock("@/client/@tanstack/react-query.gen", () => ({
  getExerciseCatalogQueryOptions: () => ({
    queryKey: ["catalog"],
    queryFn: mocks.load,
  }),
  putExercisesByIdCatalogMutation: () => ({ mutationFn: mocks.save }),
  getExercisesByIdQueryKey: () => ["detail"],
}));

const pushups: ExercisecatalogEntry = {
  id: "Pushups",
  name: "Pushups",
  equipment: "body only",
  primaryMuscles: ["chest"],
  secondaryMuscles: ["shoulders", "triceps"],
};
const curls: ExercisecatalogEntry = {
  id: "Barbell_Curl",
  name: "Barbell Curl",
  equipment: "barbell",
  primaryMuscles: ["biceps"],
  secondaryMuscles: ["forearms"],
};
let saved: ExercisecatalogEntry | undefined;
function Harness() {
  const { data } = useQuery({
    queryKey: ["detail"],
    queryFn: () => ({ catalog: saved }),
  });
  return (
    <>
      <h1>My custom exercise</h1>
      <ExerciseClassification
        exerciseId={42}
        catalog={data?.catalog}
      />
    </>
  );
}
function mount(component = <Harness />) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return render(
    <QueryClientProvider client={client}>{component}</QueryClientProvider>,
  );
}
beforeEach(() => {
  saved = undefined;
  mocks.load.mockReset().mockResolvedValue([pushups, curls]);
  mocks.save
    .mockReset()
    .mockImplementation(async ({ body }: { body: { catalog_id: string } }) => {
      saved = [pushups, curls].find((entry) => entry.id === body.catalog_id);
    });
});
describe("exercise classification", () => {
  it("classifies, changes and clears while keeping the custom name", async () => {
    mount();
    fireEvent.click(screen.getByRole("button", { name: "Classify exercise" }));
    await screen.findByRole("button", { name: "Pushups" });
    fireEvent.change(screen.getByLabelText("Search catalog"), {
      target: { value: "chest" },
    });
    expect(
      screen.queryByRole("button", { name: "Barbell Curl" }),
    ).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Pushups" }));
    await screen.findByRole("button", { name: "Change classification" });
    expect(mocks.save).toHaveBeenLastCalledWith(
      { path: { id: 42 }, body: { catalog_id: "Pushups" } },
      expect.anything(),
    );
    expect(screen.getByText("Equipment: body only")).toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: "My custom exercise" }),
    ).toBeInTheDocument();
    fireEvent.click(
      screen.getByRole("button", { name: "Change classification" }),
    );
    fireEvent.change(screen.getByLabelText("Search catalog"), {
      target: { value: "barbell" },
    });
    fireEvent.click(
      await screen.findByRole("button", { name: "Barbell Curl" }),
    );
    await waitFor(() =>
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument(),
    );
    expect(screen.getByText("Primary: biceps")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Clear" }));
    await screen.findByText("Unclassified");
    expect(mocks.save).toHaveBeenLastCalledWith(
      { path: { id: 42 }, body: { catalog_id: "" } },
      expect.anything(),
    );
  });
  it("shows failed saves and keeps the previous classification", async () => {
    saved = pushups;
    mocks.save.mockRejectedValue(new Error("network"));
    mount();
    fireEvent.click(await screen.findByRole("button", { name: "Clear" }));
    await screen.findByRole("alert");
    expect(screen.getByText("Equipment: body only")).toBeInTheDocument();
  });
  it("supports retry and an empty search without guessing a match", async () => {
    mocks.load.mockRejectedValueOnce(new Error("network"));
    const select = vi.fn();
    mount(<ExerciseCatalogPicker onSelect={select} />);
    await screen.findByRole("alert");
    fireEvent.click(screen.getByRole("button", { name: "Retry" }));
    await screen.findByRole("button", { name: "Pushups" });
    fireEvent.change(screen.getByLabelText("Search catalog"), {
      target: { value: "mystery movement" },
    });
    expect(screen.getByText(/No catalog matches/)).toBeInTheDocument();
    expect(select).not.toHaveBeenCalled();
  });
});
