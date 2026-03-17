import { Activity, BarChart3, Calendar, TrendingUp } from 'lucide-react';
import { StatsGrid } from '@/components/stats-grid';
import type { AnalyticsWorkoutSummary } from '@/lib/analytics';

export interface AnalyticsSummaryCardsProps {
  summary: AnalyticsWorkoutSummary;
}

export function AnalyticsSummaryCards({
  summary,
}: AnalyticsSummaryCardsProps) {
  return (
    <StatsGrid
      columns={4}
      items={[
        {
          label: 'Workouts (30d)',
          value: summary.totalWorkouts30d,
          icon: Calendar,
          helperText: 'Last 30 days',
        },
        {
          label: 'Avg / Week',
          value: summary.avgWorkoutsPerWeek,
          icon: BarChart3,
          helperText: '30-day pace',
        },
        {
          label: 'Current Streak',
          value: summary.currentStreak,
          icon: Activity,
          valueSuffix: 'd',
          helperText: 'Consecutive days',
        },
        {
          label: 'Longest Streak',
          value: summary.longestStreak,
          icon: TrendingUp,
          valueSuffix: 'd',
          helperText: 'Best run so far',
        },
      ]}
    />
  );
}
