import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ChartBarMetric } from './chart-bar-metric';
import { useBreakpoint } from './chart-bar-vol.utils';

vi.mock('./chart-bar-vol.components', () => ({
  ScrollableChart: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

vi.mock('recharts', () => ({
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
  BarChart: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  CartesianGrid: () => null,
  ResponsiveContainer: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  Tooltip: () => null,
  XAxis: () => null,
  YAxis: () => null,
}));

vi.mock('./chart-bar-vol.utils', async () => {
  const actual = await vi.importActual<typeof import('./chart-bar-vol.utils')>(
    './chart-bar-vol.utils'
  );

  return {
    ...actual,
    useBreakpoint: vi.fn(() => 'desktop'),
  };
});

describe('ChartBarMetric', () => {
  beforeEach(() => {
    vi.mocked(useBreakpoint).mockReturnValue('desktop');
  });

  it('navigates to the workout when a desktop bar is clicked', async () => {
    const user = userEvent.setup();
    const onWorkoutClick = vi.fn();

    render(
      <ChartBarMetric
        title="Session Best 1RM"
        range="M"
        data={[{ x: '1', date: '2026-03-01', workout_id: 42, value: 225 }]}
        unit="lb"
        onWorkoutClick={onWorkoutClick}
      />
    );

    await user.click(screen.getByTestId('metric-bar'));

    expect(onWorkoutClick).toHaveBeenCalledWith(42);
  });

  it('does not navigate when the chart is rendered on mobile', async () => {
    const user = userEvent.setup();
    const onWorkoutClick = vi.fn();
    vi.mocked(useBreakpoint).mockReturnValue('mobile');

    render(
      <ChartBarMetric
        title="Session Best 1RM"
        range="M"
        data={[{ x: '1', date: '2026-03-01', workout_id: 42, value: 225 }]}
        unit="lb"
        onWorkoutClick={onWorkoutClick}
      />
    );

    expect(screen.queryByTestId('metric-bar')).not.toBeInTheDocument();
    const staticBars = screen.getAllByTestId('metric-bar-static');
    expect(staticBars).toHaveLength(2);

    await user.click(staticBars[1]);

    expect(onWorkoutClick).not.toHaveBeenCalled();
  });
});
