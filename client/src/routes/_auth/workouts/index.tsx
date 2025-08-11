import { createFileRoute, Link } from '@tanstack/react-router';
import { useSuspenseQuery } from '@tanstack/react-query';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Calendar, ChevronRight, Dumbbell, Plus } from 'lucide-react';
import { workoutsQueryOptions } from '@/lib/api/workouts';
import { formatDate, formatTime } from '@/lib/utils';
import { getAccessToken } from '@/lib/api/auth';
import type { workout_WorkoutResponse } from '@/generated';

function WorkoutsDisplay({ workouts }: { workouts: workout_WorkoutResponse[] }) {
  const totalWorkouts = workouts.length;
  const thisWeekWorkouts = workouts.filter((workout) => {
    const workoutDate = new Date(workout.date);
    const weekAgo = new Date();
    weekAgo.setDate(weekAgo.getDate() - 7);
    return workoutDate >= weekAgo;
  }).length;

  const workoutTypes = workouts.reduce(
    (acc, workout) => {
      if (workout.notes) {
        acc[workout.notes] = (acc[workout.notes] || 0) + 1;
      }
      return acc;
    },
    {} as Record<string, number>
  );

  return (
    <main>
      <div className="max-w-lg mx-auto space-y-6 px-4 pb-8">
        {/* Header */}
        <div className="flex items-center justify-between pt-4">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Workouts</h1>
          </div>
          <Button size="sm" asChild>
            <Link to="/workouts/new-2">
              <Plus className="w-4 h-4 mr-2" />
              New Workout
            </Link>
          </Button>
        </div>
        {/* MARK: Summary Cards */}
        <div className="grid grid-cols-2 gap-4">
          <Card className="p-4">
            <div className="flex items-center gap-2 mb-2">
              <Dumbbell className="w-5 h-5 text-primary" />
              <span className="text-sm font-semibold">Total Workouts</span>
            </div>
            <div className="text-2xl text-card-foreground font-bold">
              {totalWorkouts}
            </div>
          </Card>

          <Card className="p-4">
            <div className="flex items-center gap-2 mb-2">
              <Calendar className="w-5 h-5 text-primary" />
              <span className="text-sm font-semibold">This Week</span>
            </div>
            <div className="text-2xl text-card-foreground font-bold">
              {thisWeekWorkouts}
            </div>
          </Card>
        </div>
        {/* MARK: List */}
        <Card className="border-0 shadow-sm backdrop-blur-sm">
          <CardHeader>
            <CardTitle className="text-xl font-semibold">
              Recent Workouts
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {workouts.map((workout) => (
              <Link
                key={workout.id}
                to="/workouts/$workoutId"
                params={{
                  workoutId: workout.id,
                }}
                className="flex cursor-pointer items-center justify-between rounded-lg bg-muted/50 px-3 py-2"
              >
                <div className="flex items-center space-x-4">
                  {workout.notes && (
                    <div className="flex items-center space-x-2">
                      <Badge
                        variant="outline"
                        className="border-border bg-muted text-xs"
                      >
                        {workout.notes.toUpperCase()}
                      </Badge>
                    </div>
                  )}
                  <div className="flex items-center space-x-2 text-sm mt-1">
                    <span>{formatDate(workout.date)}</span>
                    <span>•</span>
                    <span className="text-muted-foreground">
                      {formatTime(workout.created_at)}
                    </span>
                  </div>
                </div>
                <ChevronRight className="w-5 h-5 text-muted-foreground" />
              </Link>
            ))}
          </CardContent>
        </Card>

        {/* MARK: Distribution */}
        <Card className="p-4">
          <CardTitle className="text-xl font-semibold">
            Workout Distribution
          </CardTitle>
          <CardContent className="px-0">
            <div className="flex flex-wrap gap-4">
              {Object.entries(workoutTypes).map(([type, count]) => (
                <div key={type} className="text-center p-4 rounded-xl">
                  <p className="font-semibold text-lg">{count}</p>
                  <p className="text-sm uppercase">{type}</p>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </main>
  );
}

export const Route = createFileRoute('/_auth/workouts/')({
  loader: async ({ context }): Promise<string> => {
    const accessToken = await getAccessToken(context.user);
    context.queryClient.ensureQueryData(workoutsQueryOptions(accessToken));
    return accessToken;
  },
  component: RouteComponent,
});

function RouteComponent() {
  const accessToken = Route.useLoaderData();
  const { data: workouts } = useSuspenseQuery(
    workoutsQueryOptions(accessToken)
  );
  return <WorkoutsDisplay workouts={workouts} />;
}
