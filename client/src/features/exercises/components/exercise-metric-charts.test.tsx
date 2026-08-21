import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ExerciseMetricCharts } from "./exercise-metric-charts";

const { mockUseQuery } = vi.hoisted(() => ({
  mockUseQuery: vi.fn(),
}));

vi.mock("@tanstack/react-query", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@tanstack/react-query")>();

  return {
    ...actual,
    useQuery: mockUseQuery,
    keepPreviousData: (previousData: unknown) => previousData,
  };
});

vi.mock("@tanstack/react-router", () => ({
  useRouter: () => ({
    navigate: vi.fn(),
  }),
}));

vi.mock("@/components/charts/chart-bar-metric", () => ({
  ChartBarMetric: ({ title }: { title: string }) => <div>{title}</div>,
}));

vi.mock("@/components/charts/chart-bar-vol.components", () => ({
  RangeSelector: () => <div>Range selector</div>,
}));

describe("ExerciseMetricCharts", () => {
  beforeEach(() => {
    mockUseQuery.mockReset();
  });

  it("shows an initial loading state before metric history resolves", () => {
    mockUseQuery.mockReturnValue({
      data: undefined,
      isFetching: true,
      isPending: true,
    });

    render(
      <ExerciseMetricCharts
        exerciseId={1}
        exerciseSets={[]}
        isDemoMode={false}
      />,
    );

    const status = screen.getByRole("status");
    const spinner = status.querySelector("svg");

    expect(status).toHaveTextContent("Loading session metrics...");
    expect(status.closest('[data-slot="card"]')).toBeNull();
    expect(status).toHaveClass("justify-center");
    expect(spinner).toHaveClass("size-6", "text-primary");
    expect(screen.queryByText("Session Best 1RM")).not.toBeInTheDocument();
  });

  it("shows an empty state when the range has no working-set sessions", () => {
    mockUseQuery.mockReturnValue({
      data: { points: [] },
      isFetching: false,
      isPending: false,
    });

    const { container } = render(
      <ExerciseMetricCharts
        exerciseId={1}
        exerciseSets={[]}
        isDemoMode={false}
      />,
    );

    const message = screen.getByText("No working-set sessions in this range.");

    expect(message).toHaveClass("text-center");
    expect(message.closest('[data-slot="card"]')).toBeNull();
    expect(container.querySelector('[data-slot="card"]')).toBeNull();
    expect(screen.queryByText("Session Best 1RM")).not.toBeInTheDocument();
  });

  it("shows an empty state when every weighted metric is zero", () => {
    mockUseQuery.mockReturnValue({
      data: {
        points: [
          {
            x: "1",
            date: "2026-03-01",
            workout_id: 42,
            session_best_e1rm: 0,
            session_avg_e1rm: 0,
            session_avg_intensity: 0,
            session_best_intensity: 0,
            total_volume_working: 0,
          },
          {
            x: "2",
            date: "2026-03-08",
            workout_id: 43,
          },
        ],
      },
      isFetching: false,
      isPending: false,
    });

    const { container } = render(
      <ExerciseMetricCharts
        exerciseId={1}
        exerciseSets={[]}
        isDemoMode={false}
      />,
    );

    const message = screen.getByText(
      "No weighted metrics for this exercise/range.",
    );

    expect(message).toHaveClass("text-center");
    expect(message.closest('[data-slot="card"]')).toBeNull();
    expect(container.querySelector('[data-slot="card"]')).toBeNull();
    expect(screen.queryByText("Session Best 1RM")).not.toBeInTheDocument();
  });

  it("shows charts when any session has a weighted metric", () => {
    mockUseQuery.mockReturnValue({
      data: {
        points: [
          {
            x: "1",
            date: "2026-03-01",
            workout_id: 42,
            session_best_e1rm: 0,
            session_avg_e1rm: 0,
            session_avg_intensity: 0,
            session_best_intensity: 0,
            total_volume_working: 0,
          },
          {
            x: "2",
            date: "2026-03-08",
            workout_id: 43,
            session_best_e1rm: 225,
            session_avg_e1rm: 220,
            session_avg_intensity: 84.5,
            session_best_intensity: 91.2,
            total_volume_working: 5400,
          },
        ],
      },
      isFetching: false,
      isPending: false,
    });

    render(
      <ExerciseMetricCharts
        exerciseId={1}
        exerciseSets={[]}
        isDemoMode={false}
      />,
    );

    expect(screen.getByText("Session Best 1RM")).toBeInTheDocument();
    expect(
      screen.queryByText("No weighted metrics for this exercise/range."),
    ).not.toBeInTheDocument();
  });

  it("keeps the current chart visible while a new range is fetching", () => {
    mockUseQuery.mockReturnValue({
      data: {
        points: [
          {
            x: "1",
            date: "2026-03-01",
            workout_id: 42,
            session_best_e1rm: 225,
            session_avg_e1rm: 220,
            session_avg_intensity: 84.5,
            session_best_intensity: 91.2,
            total_volume_working: 5400,
          },
        ],
      },
      isFetching: true,
      isPending: false,
    });

    render(
      <ExerciseMetricCharts
        exerciseId={1}
        exerciseSets={[]}
        isDemoMode={false}
      />,
    );

    expect(screen.getByText("Updating chart...")).toBeInTheDocument();
    expect(screen.getByText("Session Best 1RM")).toBeInTheDocument();
  });

  it("rethrows the initial load error when no chart data is available", () => {
    const error = new Error("metrics history failed");
    const consoleError = vi
      .spyOn(console, "error")
      .mockImplementation(() => {});

    mockUseQuery.mockReturnValue({
      data: undefined,
      error,
      isFetching: false,
      isPending: false,
    });

    expect(() =>
      render(
        <ExerciseMetricCharts
          exerciseId={1}
          exerciseSets={[]}
          isDemoMode={false}
        />,
      ),
    ).toThrow(error);

    consoleError.mockRestore();
  });

  it("shows a stale-data warning when refreshing fails after prior data loaded", () => {
    mockUseQuery.mockReturnValue({
      data: {
        points: [
          {
            x: "1",
            date: "2026-03-01",
            workout_id: 42,
            session_best_e1rm: 225,
            session_avg_e1rm: 220,
            session_avg_intensity: 84.5,
            session_best_intensity: 91.2,
            total_volume_working: 5400,
          },
        ],
      },
      error: new Error("refresh failed"),
      isFetching: false,
      isPending: false,
    });

    render(
      <ExerciseMetricCharts
        exerciseId={1}
        exerciseSets={[]}
        isDemoMode={false}
      />,
    );

    expect(
      screen.getByText("Couldn't update chart. Showing previous data."),
    ).toBeInTheDocument();
    expect(screen.getByText("Session Best 1RM")).toBeInTheDocument();
  });

  it("labels the metrics section as working-set based", () => {
    mockUseQuery.mockReturnValue({
      data: {
        points: [
          {
            x: "1",
            date: "2026-03-01",
            workout_id: 42,
            session_best_e1rm: 225,
            session_avg_e1rm: 220,
            session_avg_intensity: 84.5,
            session_best_intensity: 91.2,
            total_volume_working: 5400,
          },
        ],
      },
      isFetching: false,
      isPending: false,
    });

    render(
      <ExerciseMetricCharts
        exerciseId={1}
        exerciseSets={[]}
        isDemoMode={false}
      />,
    );

    expect(
      screen.getByText(
        "Each bar represents one workout session. e1RM, intensity, and volume are computed from working sets. Intensity can exceed 100%.",
      ),
    ).toBeInTheDocument();
  });
});
