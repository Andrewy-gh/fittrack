import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import type { RecommendationRange, RecommendationSnapshot } from "@/client";
import type { RecommendationApi } from "@/features/workouts/api/recommendations";
import { ExerciseRecommendationPanel } from "../exercise-recommendation-panel";

function memoryApi(): RecommendationApi {
  let baseline: RecommendationRange | null = null;
  return {
    async get(exerciseId, readiness) {
      return {
        ok: true,
        value: {
          exerciseId,
          readiness,
          source: baseline ? "prescription" : "none",
          baseline,
          range: baseline
            ? {
                min:
                  readiness === "sluggish"
                    ? Math.max(1, baseline.min - 1)
                    : baseline.min,
                max:
                  readiness === "sluggish"
                    ? Math.max(1, baseline.max - 1)
                    : baseline.max,
              }
            : null,
          previous: null,
          explanation: baseline
            ? "Use your baseline."
            : "No eligible recent working-set history.",
          policyVersion: "working-sets-v1",
        },
      };
    },
    async prescribe(_id, next) {
      baseline = next;
      return { ok: true, value: null };
    },
    async saved() {
      return { ok: true, value: [] };
    },
  };
}

describe("exercise recommendation controls", () => {
  it("persists an explicit baseline and records readiness and optional feedback independently", async () => {
    const api = memoryApi();
    const user = userEvent.setup();
    let recorded: RecommendationSnapshot | null = null;
    const view = () =>
      render(
        <QueryClientProvider client={new QueryClient()}>
          <ExerciseRecommendationPanel
            exerciseId={1}
            exerciseName="Press"
            userId="owner"
            api={api}
            onChange={(exerciseName, recommendation, feedback) => {
              recorded = recommendation
                ? { exerciseName, recommendation, feedback }
                : null;
            }}
          />
        </QueryClientProvider>,
      );
    const first = view();
    expect(
      await screen.findByText("No recommended range yet"),
    ).toBeInTheDocument();
    await user.type(
      screen.getByRole("spinbutton", { name: "Baseline minimum sets" }),
      "3",
    );
    await user.type(
      screen.getByRole("spinbutton", { name: "Baseline maximum sets" }),
      "4",
    );
    await user.click(screen.getByRole("button", { name: "Save baseline" }));
    expect(
      await screen.findByText("Recommended: 3–4 working sets"),
    ).toBeInTheDocument();
    await user.click(screen.getByRole("radio", { name: "Great" }));
    expect(
      await screen.findByText("Recommended: 3–4 working sets"),
    ).toBeInTheDocument();
    await user.click(screen.getByRole("radio", { name: "Sluggish" }));
    expect(
      await screen.findByText("Recommended: 2–3 working sets"),
    ).toBeInTheDocument();
    await user.selectOptions(
      screen.getByRole("combobox", { name: "After this exercise (optional)" }),
      "about_right",
    );
    await waitFor(() =>
      expect(recorded).toMatchObject({
        exerciseName: "Press",
        feedback: "about_right",
        recommendation: {
          readiness: "sluggish",
          range: { min: 2, max: 3 },
          baseline: { min: 3, max: 4 },
          source: "prescription",
        },
      }),
    );
    first.unmount();
    view();
    expect(
      await screen.findByText("Recommended: 3–4 working sets"),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("spinbutton", { name: "Baseline minimum sets" }),
    ).toHaveValue(3);
    await user.click(screen.getByRole("button", { name: "Clear baseline" }));
    expect(
      await screen.findByText("No recommended range yet"),
    ).toBeInTheDocument();
  });

  it("rejects invalid baseline ranges while leaving guidance available", async () => {
    render(
      <QueryClientProvider client={new QueryClient()}>
        <ExerciseRecommendationPanel
          exerciseId={1}
          exerciseName="Press"
          userId="owner"
          api={memoryApi()}
          onChange={() => undefined}
        />
      </QueryClientProvider>,
    );
    const user = userEvent.setup();
    await screen.findByText("No recommended range yet");
    await user.type(
      screen.getByRole("spinbutton", { name: "Baseline minimum sets" }),
      "5",
    );
    await user.type(
      screen.getByRole("spinbutton", { name: "Baseline maximum sets" }),
      "2",
    );
    await user.click(screen.getByRole("button", { name: "Save baseline" }));
    expect(screen.getByRole("status")).toHaveTextContent(
      "Enter an ordered range between 1 and 20",
    );
    expect(screen.getByText("No recommended range yet")).toBeInTheDocument();
  });
});
