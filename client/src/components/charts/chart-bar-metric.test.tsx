import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ChartBarMetric, formatTooltipDateLabel } from "./chart-bar-metric";
import { useBreakpoint } from "./chart-bar-vol.utils";

const { mockScrollableChart, mockYAxis } = vi.hoisted(() => ({
  mockScrollableChart: vi.fn(),
  mockYAxis: vi.fn(),
}));

vi.mock("./chart-bar-vol.components", () => ({
  ScrollableChart: (props: {
    children: React.ReactNode;
    visibleBarCount?: number;
  }) => {
    mockScrollableChart(props);
    return <div>{props.children}</div>;
  },
}));

vi.mock("recharts", () => ({
  Bar: ({ onClick }: { onClick?: (payload: unknown) => void }) =>
    onClick ? (
      <button
        type="button"
        data-testid="metric-bar"
        onClick={() => onClick({ payload: { workout_id: 42 } })}
      >
        Metric bar
      </button>
    ) : (
      <div data-testid="metric-bar-static" />
    ),
  BarChart: ({ children }: { children: React.ReactNode }) => (
    <div>{children}</div>
  ),
  CartesianGrid: () => null,
  ResponsiveContainer: ({ children }: { children: React.ReactNode }) => (
    <div>{children}</div>
  ),
  Tooltip: ({
    formatter,
  }: {
    formatter?: (
      value: number,
      name: string,
      item: { payload: Record<string, never> },
    ) => [React.ReactNode, string];
  }) => {
    const [content] = formatter?.(5400, "", { payload: {} }) ?? [];
    return <div data-testid="metric-tooltip-value">{content}</div>;
  },
  XAxis: () => null,
  YAxis: (props: { tickFormatter?: (value: number) => string }) => {
    mockYAxis(props);
    return null;
  },
}));

vi.mock("./chart-bar-vol.utils", async () => {
  const actual = await vi.importActual<typeof import("./chart-bar-vol.utils")>(
    "./chart-bar-vol.utils",
  );

  return {
    ...actual,
    useBreakpoint: vi.fn(() => "desktop"),
  };
});

describe("ChartBarMetric", () => {
  beforeEach(() => {
    mockScrollableChart.mockClear();
    mockYAxis.mockClear();
    vi.mocked(useBreakpoint).mockReturnValue("desktop");
  });

  it.each([
    ["lb", 227.50000000000003, "228"],
    ["%", 85.71428571428571, "85.7%"],
  ] as const)(
    "formats %s axis values for people to read",
    (unit, value, label) => {
      render(
        <ChartBarMetric
          title="Progress"
          range="M"
          data={[{ x: "1", date: "2026-03-01", value }]}
          unit={unit}
        />,
      );

      const axisWithLabels = mockYAxis.mock.calls.find(
        ([props]) => props.tickFormatter,
      )?.[0];

      expect(axisWithLabels?.tickFormatter?.(value)).toBe(label);
    },
  );

  it("labels volume tooltip values as pounds times reps", () => {
    render(
      <ChartBarMetric
        title="Total Weight Moved"
        range="M"
        data={[{ x: "1", date: "2026-03-01", value: 5400 }]}
        unit="vol"
      />,
    );

    expect(screen.getByTestId("metric-tooltip-value")).toHaveTextContent(
      "5,400 lb × reps",
    );
  });

  it("navigates to the workout when a desktop bar is clicked", async () => {
    const user = userEvent.setup();
    const onWorkoutClick = vi.fn();

    render(
      <ChartBarMetric
        title="Session Best 1RM"
        range="M"
        data={[{ x: "1", date: "2026-03-01", workout_id: 42, value: 225 }]}
        unit="lb"
        onWorkoutClick={onWorkoutClick}
      />,
    );

    await user.click(screen.getByTestId("metric-bar"));

    expect(onWorkoutClick).toHaveBeenCalledWith(42);
  });

  it.each([["W", 7]] as const)(
    "shows %s session bars in the initial viewport",
    (range, count) => {
      render(
        <ChartBarMetric
          title="Session Best 1RM"
          range={range}
          data={[{ x: "1", date: "2026-03-01", value: 225 }]}
          unit="lb"
        />,
      );

      expect(mockScrollableChart).toHaveBeenCalledWith(
        expect.objectContaining({ visibleBarCount: count }),
      );
    },
  );

  it("does not navigate when the chart is rendered on mobile", async () => {
    const user = userEvent.setup();
    const onWorkoutClick = vi.fn();
    vi.mocked(useBreakpoint).mockReturnValue("mobile");

    render(
      <ChartBarMetric
        title="Session Best 1RM"
        range="M"
        data={[{ x: "1", date: "2026-03-01", workout_id: 42, value: 225 }]}
        unit="lb"
        onWorkoutClick={onWorkoutClick}
      />,
    );

    expect(screen.queryByTestId("metric-bar")).not.toBeInTheDocument();
    const staticBars = screen.getAllByTestId("metric-bar-static");
    expect(staticBars).toHaveLength(2);

    await user.click(staticBars[1]);

    expect(onWorkoutClick).not.toHaveBeenCalled();
  });
});

describe("formatTooltipDateLabel", () => {
  it("formats a weekly range within one month", () => {
    expect(formatTooltipDateLabel(new Date(2026, 2, 2), "week")).toBe(
      "Mar 2 - 8 2026",
    );
  });

  it("formats both month names when a weekly range crosses months", () => {
    expect(formatTooltipDateLabel(new Date(2026, 2, 30), "week")).toBe(
      "Mar 30 - Apr 5 2026",
    );
  });
});
