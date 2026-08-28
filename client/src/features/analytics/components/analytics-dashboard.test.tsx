import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { AnalyticsDashboard } from "./analytics-dashboard";

vi.mock("@/components/generic-combobox", () => ({
  GenericCombobox: ({
    onChange,
  }: {
    onChange: (value: { id: number }) => void;
  }) => <button onClick={() => onChange({ id: 2 })}>Exercise selector</button>,
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

  it("restores the scroll position after changed exercise metrics load", () => {
    const exercises = [
      {
        id: 1,
        name: "Bench press",
        created_at: "2026-08-21T00:00:00Z",
        updated_at: "2026-08-21T00:00:00Z",
        user_id: "user-1",
      },
      {
        id: 2,
        name: "Squat",
        created_at: "2026-08-21T00:00:00Z",
        updated_at: "2026-08-21T00:00:00Z",
        user_id: "user-1",
      },
    ];
    const scrollTo = vi.spyOn(window, "scrollTo").mockImplementation(() => {});
    vi.spyOn(window, "scrollX", "get").mockReturnValue(12);
    vi.spyOn(window, "scrollY", "get").mockReturnValue(480);

    const { rerender } = render(
      <AnalyticsDashboard
        isLoadingExercises={false}
        exercises={exercises}
        selectedExerciseId={1}
        onSelectExercise={vi.fn()}
        isLoadingDetails={false}
        isDemoMode={false}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Exercise selector" }));
    rerender(
      <AnalyticsDashboard
        isLoadingExercises={false}
        exercises={exercises}
        selectedExerciseId={2}
        onSelectExercise={vi.fn()}
        isLoadingDetails={false}
        isDemoMode={false}
      />,
    );

    expect(scrollTo).toHaveBeenCalledWith({
      left: 12,
      top: 480,
      behavior: "auto",
    });
  });
});
