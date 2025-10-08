import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { RecentSets } from '../recent-sets-display';
import * as ReactQuery from '@tanstack/react-query';
import * as unifiedQueryOptions from '@/lib/api/unified-query-options';

// Mock dependencies
vi.mock('@tanstack/react-query', async () => {
  const actual = await vi.importActual('@tanstack/react-query');
  return {
    ...actual,
    useSuspenseQuery: vi.fn(),
  };
});

vi.mock('@tanstack/react-router', () => ({
  Link: ({ children, to, params, className }: any) => (
    <a href={`${to}/${params?.workoutId || ''}`} className={className} role="link">
      {children}
    </a>
  ),
}));

vi.mock('@/components/ui/card', () => ({
  Card: ({ children, className }: any) => <div className={className} data-testid="card">{children}</div>,
  CardHeader: ({ children }: any) => <div data-testid="card-header">{children}</div>,
  CardTitle: ({ children, className }: any) => <div className={className} data-testid="card-title">{children}</div>,
  CardContent: ({ children, className }: any) => <div className={className} data-testid="card-content">{children}</div>,
}));

vi.mock('lucide-react', () => ({
  ChevronRight: () => <span data-testid="chevron">→</span>,
}));

vi.mock('@/lib/api/unified-query-options', async () => {
  const actual = await vi.importActual('@/lib/api/unified-query-options');
  return {
    ...actual,
    getRecentSetsQueryOptions: vi.fn(),
  };
});

vi.mock('@/lib/utils', async () => {
  const actual = await vi.importActual('@/lib/utils');
  return {
    ...actual,
    formatDate: (date: string) => date,
    sortByExerciseAndSetOrder: (sets: any[]) => sets,
  };
});

describe('RecentSets', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('with null exerciseId', () => {
    it('renders nothing when exerciseId is null', () => {
      const { container } = render(<RecentSets exerciseId={null} user={null} />);

      expect(container.firstChild).toBeNull();
    });
  });

  describe('with demo user (null)', () => {
    it('renders recent sets from demo data', async () => {
      const mockRecentSets = [
        {
          set_id: 1,
          workout_id: 100,
          exercise_id: 1,
          set_order: 1,
          reps: 10,
          weight: 135,
          workout_date: '2025-10-01',
        },
        {
          set_id: 2,
          workout_id: 100,
          exercise_id: 1,
          set_order: 2,
          reps: 8,
          weight: 145,
          workout_date: '2025-10-01',
        },
      ];

      const mockQueryOptions = { queryKey: ['demo_recentSets', 1], queryFn: vi.fn() };
      vi.mocked(unifiedQueryOptions.getRecentSetsQueryOptions).mockReturnValue(mockQueryOptions as any);
      vi.mocked(ReactQuery.useSuspenseQuery).mockReturnValue({
        data: mockRecentSets,
      } as any);

      render(<RecentSets exerciseId={1} user={null} />);

      await waitFor(() => {
        expect(screen.getByText('Recent Sets')).toBeInTheDocument();
      });

      expect(screen.getByText('135 lbs')).toBeInTheDocument();
      expect(screen.getByText('10 reps')).toBeInTheDocument();
      expect(screen.getByText('145 lbs')).toBeInTheDocument();
      expect(screen.getByText('8 reps')).toBeInTheDocument();
      expect(unifiedQueryOptions.getRecentSetsQueryOptions).toHaveBeenCalledWith(null, 1);
    });

    it('renders empty state when no recent sets exist', async () => {
      const mockQueryOptions = { queryKey: ['demo_recentSets', 1], queryFn: vi.fn() };
      vi.mocked(unifiedQueryOptions.getRecentSetsQueryOptions).mockReturnValue(mockQueryOptions as any);
      vi.mocked(ReactQuery.useSuspenseQuery).mockReturnValue({
        data: [],
      } as any);

      const { container } = render(<RecentSets exerciseId={1} user={null} />);

      await waitFor(() => {
        // When no sets, should only show suspense wrapper, not the actual content
        expect(container.querySelector('[data-testid="card"]')).not.toBeInTheDocument();
      });
    });
  });

  describe('with authenticated user', () => {
    it('renders recent sets from API data', async () => {
      const mockUser = { id: 'user123' } as any;
      const mockRecentSets = [
        {
          set_id: 3,
          workout_id: 200,
          exercise_id: 2,
          set_order: 1,
          reps: 12,
          weight: 225,
          workout_date: '2025-10-05',
        },
      ];

      const mockQueryOptions = { queryKey: ['recentSets', 2], queryFn: vi.fn() };
      vi.mocked(unifiedQueryOptions.getRecentSetsQueryOptions).mockReturnValue(mockQueryOptions as any);
      vi.mocked(ReactQuery.useSuspenseQuery).mockReturnValue({
        data: mockRecentSets,
      } as any);

      render(<RecentSets exerciseId={2} user={mockUser} />);

      await waitFor(() => {
        expect(screen.getByText('Recent Sets')).toBeInTheDocument();
      });

      expect(screen.getByText('225 lbs')).toBeInTheDocument();
      expect(screen.getByText('12 reps')).toBeInTheDocument();
      expect(unifiedQueryOptions.getRecentSetsQueryOptions).toHaveBeenCalledWith(mockUser, 2);
    });

    it('renders empty state when no recent sets exist', async () => {
      const mockUser = { id: 'user123' } as any;
      const mockQueryOptions = { queryKey: ['recentSets', 2], queryFn: vi.fn() };
      vi.mocked(unifiedQueryOptions.getRecentSetsQueryOptions).mockReturnValue(mockQueryOptions as any);
      vi.mocked(ReactQuery.useSuspenseQuery).mockReturnValue({
        data: [],
      } as any);

      const { container } = render(<RecentSets exerciseId={2} user={mockUser} />);

      await waitFor(() => {
        // When no sets, should only show suspense wrapper, not the actual content
        expect(container.querySelector('[data-testid="card"]')).not.toBeInTheDocument();
      });
    });
  });

  describe('uses correct query options based on user state', () => {
    it('calls getRecentSetsQueryOptions with correct parameters for demo user', async () => {
      const mockQueryOptions = { queryKey: ['demo_recentSets', 5], queryFn: vi.fn() };
      vi.mocked(unifiedQueryOptions.getRecentSetsQueryOptions).mockReturnValue(mockQueryOptions as any);
      vi.mocked(ReactQuery.useSuspenseQuery).mockReturnValue({
        data: [],
      } as any);

      render(<RecentSets exerciseId={5} user={null} />);

      await waitFor(() => {
        expect(unifiedQueryOptions.getRecentSetsQueryOptions).toHaveBeenCalledWith(null, 5);
      });

      expect(ReactQuery.useSuspenseQuery).toHaveBeenCalledWith(mockQueryOptions);
    });

    it('calls getRecentSetsQueryOptions with correct parameters for authenticated user', async () => {
      const mockUser = { id: 'user456' } as any;
      const mockQueryOptions = { queryKey: ['recentSets', 5], queryFn: vi.fn() };
      vi.mocked(unifiedQueryOptions.getRecentSetsQueryOptions).mockReturnValue(mockQueryOptions as any);
      vi.mocked(ReactQuery.useSuspenseQuery).mockReturnValue({
        data: [],
      } as any);

      render(<RecentSets exerciseId={5} user={mockUser} />);

      await waitFor(() => {
        expect(unifiedQueryOptions.getRecentSetsQueryOptions).toHaveBeenCalledWith(mockUser, 5);
      });

      expect(ReactQuery.useSuspenseQuery).toHaveBeenCalledWith(mockQueryOptions);
    });
  });

  describe('displays sets with correct formatting', () => {
    it('calculates and displays volume correctly', async () => {
      const mockRecentSets = [
        {
          set_id: 1,
          workout_id: 100,
          exercise_id: 1,
          set_order: 1,
          reps: 10,
          weight: 100,
          workout_date: '2025-10-01',
        },
      ];

      const mockQueryOptions = { queryKey: ['recentSets', 1], queryFn: vi.fn() };
      vi.mocked(unifiedQueryOptions.getRecentSetsQueryOptions).mockReturnValue(mockQueryOptions as any);
      vi.mocked(ReactQuery.useSuspenseQuery).mockReturnValue({
        data: mockRecentSets,
      } as any);

      render(<RecentSets exerciseId={1} user={null} />);

      await waitFor(() => {
        // Volume = 100 lbs * 10 reps = 1000
        expect(screen.getByText('1,000 vol')).toBeInTheDocument();
      });
    });

    it('groups sets by workout date', async () => {
      const mockRecentSets = [
        {
          set_id: 1,
          workout_id: 100,
          exercise_id: 1,
          set_order: 1,
          reps: 10,
          weight: 135,
          workout_date: '2025-10-01',
        },
        {
          set_id: 2,
          workout_id: 100,
          exercise_id: 1,
          set_order: 2,
          reps: 8,
          weight: 135,
          workout_date: '2025-10-01',
        },
        {
          set_id: 3,
          workout_id: 101,
          exercise_id: 1,
          set_order: 1,
          reps: 12,
          weight: 125,
          workout_date: '2025-09-28',
        },
      ];

      const mockQueryOptions = { queryKey: ['recentSets', 1], queryFn: vi.fn() };
      vi.mocked(unifiedQueryOptions.getRecentSetsQueryOptions).mockReturnValue(mockQueryOptions as any);
      vi.mocked(ReactQuery.useSuspenseQuery).mockReturnValue({
        data: mockRecentSets,
      } as any);

      render(<RecentSets exerciseId={1} user={null} />);

      await waitFor(() => {
        // Should have grouped the sets by date - 2 different cards for 2 dates
        const cards = screen.getAllByTestId('card');
        expect(cards.length).toBe(2);
      });
    });
  });
});
