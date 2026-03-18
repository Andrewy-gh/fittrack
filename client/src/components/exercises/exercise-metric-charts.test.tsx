import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ExerciseMetricCharts } from './exercise-metric-charts';

const { mockUseQuery } = vi.hoisted(() => ({
  mockUseQuery: vi.fn(),
}));

vi.mock('@tanstack/react-query', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@tanstack/react-query')>();

  return {
    ...actual,
    useQuery: mockUseQuery,
    keepPreviousData: (previousData: unknown) => previousData,
  };
});

vi.mock('@tanstack/react-router', () => ({
  useRouter: () => ({
    navigate: vi.fn(),
  }),
}));

vi.mock('@/components/charts/chart-bar-metric', () => ({
  ChartBarMetric: ({ title }: { title: string }) => <div>{title}</div>,
}));

vi.mock('@/components/charts/chart-bar-vol.components', () => ({
  RangeSelector: () => <div>Range selector</div>,
}));

describe('ExerciseMetricCharts', () => {
  beforeEach(() => {
    mockUseQuery.mockReset();
  });

  it('shows an initial loading state before metric history resolves', () => {
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
      />
    );

    expect(screen.getByText('Loading session metrics...')).toBeInTheDocument();
    expect(screen.queryByText('Session Best 1RM')).not.toBeInTheDocument();
  });

  it('keeps the current chart visible while a new range is fetching', () => {
    mockUseQuery.mockReturnValue({
      data: {
        points: [
          {
            x: '1',
            date: '2026-03-01',
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
      />
    );

    expect(screen.getByText('Updating chart...')).toBeInTheDocument();
    expect(screen.getByText('Session Best 1RM')).toBeInTheDocument();
  });
});
