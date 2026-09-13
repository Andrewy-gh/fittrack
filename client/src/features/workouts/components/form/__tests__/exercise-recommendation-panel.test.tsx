import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import type { RecommendationApi } from "@/features/workouts/api/recommendations";
import { ExerciseRecommendationPanel } from "../exercise-recommendation-panel";

const api: RecommendationApi = {
  async get(exerciseId, readiness) {
    return {
      ok: true,
      value: {
        exerciseId,
        readiness,
        plan: {
          sets: 3,
          reps: readiness === "great" ? 9 : readiness === "sluggish" ? 6 : 8,
          weight: 100,
        },
        explanation: "Repeat your last working sets.",
        policyVersion: "exercise-plan-v2",
      },
    };
  },
};

describe("compact exercise suggestions", () => {
  it("shows guidance without an apply button or extra details", async () => {
    render(
      <QueryClientProvider client={new QueryClient()}>
        <ExerciseRecommendationPanel
          exerciseId={1}
          exerciseName="Press"
          userId="owner"
          api={api}
          onChange={() => undefined}
        />
      </QueryClientProvider>,
    );
    expect(await screen.findByText("3 × 8 at 100 lb")).toBeTruthy();
    expect(
      screen.queryByRole("button", {
        name: /Use suggestion|Ready for your next set/,
      }),
    ).toBeNull();
    expect(screen.queryByText("Details")).toBeNull();
  });

  it("offers three accessible numbered faces and records the chosen feeling", async () => {
    const user = userEvent.setup();
    let recorded: unknown;
    render(
      <QueryClientProvider client={new QueryClient()}>
        <ExerciseRecommendationPanel
          exerciseId={1}
          exerciseName="Press"
          userId="owner"
          api={api}
          onChange={(_name, result) => {
            recorded = result;
          }}
        />
      </QueryClientProvider>,
    );
    await screen.findByText("3 × 8 at 100 lb");
    expect(screen.getByRole("radio", { name: "2 · Okay" })).toHaveProperty(
      "checked",
      true,
    );
    expect(screen.getAllByRole("radio")).toHaveLength(3);
    await user.click(screen.getByRole("radio", { name: "1 · Low energy" }));
    expect(await screen.findByText("3 × 6 at 100 lb")).toBeTruthy();
    await waitFor(() =>
      expect(recorded).toMatchObject({ readiness: "sluggish" }),
    );
  });
});
