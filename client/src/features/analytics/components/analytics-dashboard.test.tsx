import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { AnalyticsDashboard } from "./analytics-dashboard";

vi.mock("@/components/generic-combobox", () => ({
  GenericCombobox: () => <div>Exercise selector</div>,
}));

vi.mock("@/features/analytics/components/analytics-summary-cards", () => ({
  AnalyticsSummaryCards: () => <div>Summary cards</div>,
}));

vi.mock("@/features/analytics/components/workout-volume-chart", () => ({
  WorkoutVolumeChart: () => <div>Workout volume</div>,
}));

vi.mock("@/features/exercises/components/exercise-metric-charts", () => ({
  ExerciseMetricCharts: () => <div>Session metrics</div>,
}));

vi.mock("@/features/workouts/components/workout-contribution-graph", () => ({
  WorkoutContributionGraph: () => <div>Workout trends</div>,
}));

describe("AnalyticsDashboard", () => {
  it("shows a loading status while exercise details load", () => {
    render(
      <AnalyticsDashboard
        isLoadingExercises={false}
        exercises={[
          {
            id: 1,
            name: "Bench press",
            created_at: "2026-08-21T00:00:00Z",
            updated_at: "2026-08-21T00:00:00Z",
            user_id: "user-1",
          },
        ]}
        selectedExerciseId={1}
        onSelectExercise={vi.fn()}
        isLoadingDetails
        isDemoMode={false}
      />,
    );

    const status = screen.getByRole("status");

    expect(status).toHaveTextContent("Loading exercise metrics...");
    expect(screen.queryByText("Session metrics")).not.toBeInTheDocument();
  });
});
