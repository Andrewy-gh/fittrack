import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { useAppForm } from "@/hooks/form";
import { ExerciseSets } from "../exercise-screen";

type SetValue = {
  weight: number;
  reps: number;
  setType: "warmup" | "working";
};

function ExerciseSetsHarness({ initialSets }: { initialSets: SetValue[] }) {
  const form = useAppForm({
    defaultValues: {
      date: "2026-03-24T10:30:00.000Z",
      notes: "",
      workoutFocus: "",
      exercises: [
        {
          name: "Bench Press",
          sets: initialSets,
        },
      ],
    },
    onSubmit: async () => undefined,
  });

  return (
    <ExerciseSets
      form={form as any}
      exerciseIndex={0}
    />
  );
}

describe("ExerciseSets repeat button", () => {
  it("appends a copy of the last logged set", async () => {
    const user = userEvent.setup();

    render(
      <ExerciseSetsHarness
        initialSets={[
          { weight: 95, reps: 10, setType: "warmup" },
          { weight: 135, reps: 5, setType: "working" },
        ]}
      />,
    );

    expect(screen.getAllByTestId("exercise-card")).toHaveLength(2);

    await user.click(screen.getByRole("button", { name: "Repeat last set" }));

    const cards = screen.getAllByTestId("exercise-card");
    expect(cards).toHaveLength(3);
    expect(cards[2]).toHaveTextContent("135lb × 5");
    expect(cards[2]).toHaveTextContent("working");
  });

  it("hides the repeat button when there is no logged set to copy", () => {
    render(<ExerciseSetsHarness initialSets={[]} />);

    expect(
      screen.queryByRole("button", { name: "Repeat last set" }),
    ).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Add Set/ })).toBeInTheDocument();
  });

  it("ignores a blank placeholder set when picking what to repeat", async () => {
    const user = userEvent.setup();

    render(
      <ExerciseSetsHarness
        initialSets={[
          { weight: 185, reps: 3, setType: "working" },
          { weight: 0, reps: 0, setType: "working" },
        ]}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Repeat last set" }));

    const cards = screen.getAllByTestId("exercise-card");
    expect(cards[cards.length - 1]).toHaveTextContent("185lb × 3");
  });
});
