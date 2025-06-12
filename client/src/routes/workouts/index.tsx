import { createFileRoute, Link } from '@tanstack/react-router';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Calendar, Clock, FileText, Dumbbell } from 'lucide-react';
import { format, formatDistanceToNow, parseISO } from 'date-fns';

interface Workout {
  id: number;
  date: string;
  notes: string | null;
  created_at: string;
  updated_at: string | null;
}

function WorkoutDisplay({ workouts }: { workouts: Workout[] }) {
  const formatWorkoutDate = (dateString: string) => {
    const date = parseISO(dateString);
    return format(date, 'EEEE, MMMM d, yyyy');
  };

  const formatWorkoutTime = (dateString: string) => {
    const date = parseISO(dateString);
    return format(date, 'h:mm a');
  };

  const getRelativeTime = (dateString: string) => {
    const date = parseISO(dateString);
    return formatDistanceToNow(date, { addSuffix: true });
  };

  const sortedWorkouts = [...workouts].sort(
    (a, b) => new Date(b.date).getTime() - new Date(a.date).getTime()
  );

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 p-4 md:p-8">
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <div className="text-center mb-8">
          <div className="flex items-center justify-center gap-3 mb-4">
            <div className="p-3 bg-blue-600 rounded-full">
              <Dumbbell className="h-8 w-8 text-white" />
            </div>
            <h1 className="text-4xl font-bold text-gray-900">My Workouts</h1>
          </div>
          <p className="text-gray-600 text-lg">
            Track your fitness journey with {workouts.length} recorded workouts
          </p>
        </div>

        {/* Stats Overview */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
          <Card className="bg-white/80 backdrop-blur-sm border-0 shadow-lg">
            <CardContent className="p-6 text-center">
              <div className="text-3xl font-bold text-blue-600 mb-2">
                {workouts.length}
              </div>
              <div className="text-gray-600">Total Workouts</div>
            </CardContent>
          </Card>
          <Card className="bg-white/80 backdrop-blur-sm border-0 shadow-lg">
            <CardContent className="p-6 text-center">
              <div className="text-3xl font-bold text-green-600 mb-2">
                {
                  new Set(
                    workouts.map((w) => format(parseISO(w.date), 'yyyy-MM-dd'))
                  ).size
                }
              </div>
              <div className="text-gray-600">Unique Days</div>
            </CardContent>
          </Card>
          <Card className="bg-white/80 backdrop-blur-sm border-0 shadow-lg">
            <CardContent className="p-6 text-center">
              <div className="text-3xl font-bold text-purple-600 mb-2">
                {getRelativeTime(
                  sortedWorkouts[0]?.date || new Date().toISOString()
                )}
              </div>
              <div className="text-gray-600">Last Workout</div>
            </CardContent>
          </Card>
        </div>

        {/* Workouts Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {sortedWorkouts.map((workout) => (
            <Card
              key={workout.id}
              className="bg-white/90 backdrop-blur-sm border-0 shadow-lg hover:shadow-xl transition-all duration-300 hover:-translate-y-1"
            >
              <Link to={`/workouts/$workoutId`} params={{ workoutId: workout.id }}>
                <CardHeader className="pb-3">
                  <div className="flex items-center justify-between">
                    <CardTitle className="text-lg font-semibold text-gray-800">
                      Workout #{workout.id}
                    </CardTitle>
                    <Badge
                      variant="secondary"
                      className="bg-blue-100 text-blue-800"
                    >
                      {getRelativeTime(workout.date)}
                    </Badge>
                  </div>
                </CardHeader>
                <CardContent className="space-y-4">
                  {/* Date */}
                  <div className="flex items-center gap-3 text-gray-700">
                    <Calendar className="h-4 w-4 text-blue-600" />
                    <div>
                      <div className="font-medium">
                        {formatWorkoutDate(workout.date)}
                      </div>
                      <div className="text-sm text-gray-500">
                        {formatWorkoutTime(workout.date)}
                      </div>
                    </div>
                  </div>

                  {/* Created At */}
                  <div className="flex items-center gap-3 text-gray-700">
                    <Clock className="h-4 w-4 text-green-600" />
                    <div>
                      <div className="text-sm font-medium">Created</div>
                      <div className="text-sm text-gray-500">
                        {format(parseISO(workout.created_at), 'MMM d, h:mm a')}
                      </div>
                    </div>
                  </div>

                  {/* Notes */}
                  <div className="flex items-start gap-3 text-gray-700">
                    <FileText className="h-4 w-4 text-purple-600 mt-0.5" />
                    <div className="flex-1">
                      <div className="text-sm font-medium">Notes</div>
                      <div className="text-sm text-gray-500">
                        {workout.notes || 'No notes added'}
                      </div>
                    </div>
                  </div>

                  {/* Updated Status */}
                  {workout.updated_at && (
                    <div className="pt-2 border-t border-gray-200">
                      <div className="text-xs text-gray-500">
                        Updated {getRelativeTime(workout.updated_at)}
                      </div>
                    </div>
                  )}
                </CardContent>
              </Link>
            </Card>
          ))}
        </div>

        {/* Empty State */}
        {workouts.length === 0 && (
          <div className="text-center py-12">
            <Dumbbell className="h-16 w-16 text-gray-400 mx-auto mb-4" />
            <h3 className="text-xl font-semibold text-gray-600 mb-2">
              No workouts yet
            </h3>
            <p className="text-gray-500">
              Start your fitness journey by adding your first workout!
            </p>
          </div>
        )}
      </div>
    </div>
  );
}

export const Route = createFileRoute('/workouts/')({
  loader: async (): Promise<Workout[]> => {
    const res = await fetch('/api/workouts');
    if (!res.ok) {
      throw new Error('Failed to fetch workouts');
    }
    const data = await res.json();
    return data;
  },
  component: RouteComponent,
});

function RouteComponent() {
  const workouts = Route.useLoaderData();
  return <WorkoutDisplay workouts={workouts} />;
}
