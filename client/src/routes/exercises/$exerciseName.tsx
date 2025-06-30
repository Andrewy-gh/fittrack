import { useState } from 'react';
import { createFileRoute, useRouter } from '@tanstack/react-router';
import {
  Activity,
  ArrowLeft,
  BarChart3,
  Calendar,
  Download,
  Edit,
  Filter,
  RotateCcw,
  Target,
  Trash2,
  TrendingUp,
  Weight,
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import type { ExerciseSet } from '@/lib/types';
import { fetchExerciseSets } from '@/lib/api/exercises';

interface WorkoutGroup {
  workout_id: number;
  sets: ExerciseSet[];
  date: string;
  totalVolume: number;
  maxWeight: number;
  totalReps: number;
  setCount: number;
}

export const Route = createFileRoute('/exercises/$exerciseName')({
  loader: ({ params }) => fetchExerciseSets(params.exerciseName),
  component: RouteComponent,
});

function RouteComponent() {
  const exerciseSets = Route.useLoaderData();
  const { exerciseName } = Route.useParams();

  return (
    <ExerciseDisplay exerciseName={exerciseName} exerciseSets={exerciseSets} />
  );
}

function ExerciseDisplay({
  exerciseName,
  exerciseSets,
}: {
  exerciseName: string;
  exerciseSets: ExerciseSet[];
}) {
  const router = useRouter();
  const [selectedWorkout, setSelectedWorkout] = useState<WorkoutGroup | null>(
    null
  );
  const [selectedSet, setSelectedSet] = useState<ExerciseSet | null>(null);

  // Group sets by workout
  const workoutGroups: WorkoutGroup[] = exerciseSets.reduce((acc, set) => {
    const existingWorkout = acc.find((w) => w.workout_id === set.workout_id);

    if (existingWorkout) {
      existingWorkout.sets.push(set);
      existingWorkout.totalVolume += set.weight * set.reps;
      existingWorkout.maxWeight = Math.max(
        existingWorkout.maxWeight,
        set.weight
      );
      existingWorkout.totalReps += set.reps;
      existingWorkout.setCount += 1;
    } else {
      acc.push({
        workout_id: set.workout_id,
        sets: [set],
        date: set.created_at,
        totalVolume: set.weight * set.reps,
        maxWeight: set.weight,
        totalReps: set.reps,
        setCount: 1,
      });
    }

    return acc;
  }, [] as WorkoutGroup[]);

  // Sort by date (most recent first)
  workoutGroups.sort(
    (a, b) => new Date(b.date).getTime() - new Date(a.date).getTime()
  );

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      month: 'short',
      day: '2-digit',
      year: 'numeric',
    });
  };

  const formatTime = (dateString: string) => {
    return new Date(dateString).toLocaleTimeString('en-US', {
      hour: '2-digit',
      minute: '2-digit',
      hour12: false,
    });
  };

  const getSetTypeColor = (setType: string) => {
    switch (setType) {
      case 'working':
        return 'bg-white/20 text-white';
      case 'warmup':
        return 'bg-orange-500/20 text-orange-500';
      case 'workout':
        return 'bg-white/20 text-white';
      case 'dropset':
        return 'bg-red-500/20 text-red-500';
      default:
        return 'bg-neutral-500/20 text-neutral-300';
    }
  };

  // Calculate overall stats
  const totalSets = exerciseSets.length;
  const totalVolume = exerciseSets.reduce(
    (sum, set) => sum + set.weight * set.reps,
    0
  );
  const maxWeight = Math.max(...exerciseSets.map((set) => set.weight));
  const totalReps = exerciseSets.reduce((sum, set) => sum + set.reps, 0);
  const avgWeight =
    exerciseSets.reduce((sum, set) => sum + set.weight, 0) /
    exerciseSets.length;
  const workoutCount = workoutGroups.length;

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div className="flex items-center gap-4">
          <Button
            variant="ghost"
            onClick={() => router.history.back()}
            className="text-neutral-400 hover:text-orange-500 p-2"
          >
            <ArrowLeft className="w-5 h-5" />
          </Button>
          <div>
            <h1 className="text-2xl font-bold text-white tracking-wider">
              {exerciseName}
            </h1>
            <p className="text-sm text-neutral-400">
              Exercise ID: EX-
              {exerciseSets[0]?.exercise_id.toString().padStart(3, '0')} •
              Performance Analysis
            </p>
          </div>
        </div>
        <div className="flex gap-2">
          <Button className="bg-orange-500 hover:bg-orange-600 text-white">
            <BarChart3 className="w-4 h-4 mr-2" />
            Analytics
          </Button>
          <Button className="bg-orange-500 hover:bg-orange-600 text-white">
            <Filter className="w-4 h-4 mr-2" />
            Filter
          </Button>
        </div>
      </div>

      {/* MARK: Exercise Statistics */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-6 gap-4">
        <Card className="bg-neutral-900 border-neutral-700">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-xs text-neutral-400 tracking-wider">
                  TOTAL SETS
                </p>
                <p className="text-2xl font-bold text-white font-mono">
                  {totalSets}
                </p>
              </div>
              <RotateCcw className="w-8 h-8 text-white" />
            </div>
          </CardContent>
        </Card>

        <Card className="bg-neutral-900 border-neutral-700">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-xs text-neutral-400 tracking-wider">
                  WORKOUTS
                </p>
                <p className="text-2xl font-bold text-white font-mono">
                  {workoutCount}
                </p>
              </div>
              <Calendar className="w-8 h-8 text-white" />
            </div>
          </CardContent>
        </Card>

        <Card className="bg-neutral-900 border-neutral-700">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-xs text-neutral-400 tracking-wider">
                  MAX WEIGHT
                </p>
                <p className="text-2xl font-bold text-orange-500 font-mono">
                  {maxWeight}
                </p>
              </div>
              <Weight className="w-8 h-8 text-orange-500" />
            </div>
          </CardContent>
        </Card>

        <Card className="bg-neutral-900 border-neutral-700">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-xs text-neutral-400 tracking-wider">
                  TOTAL VOLUME
                </p>
                <p className="text-2xl font-bold text-white font-mono">
                  {totalVolume.toLocaleString()}
                </p>
              </div>
              <TrendingUp className="w-8 h-8 text-white" />
            </div>
          </CardContent>
        </Card>

        <Card className="bg-neutral-900 border-neutral-700">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-xs text-neutral-400 tracking-wider">
                  TOTAL REPS
                </p>
                <p className="text-2xl font-bold text-white font-mono">
                  {totalReps}
                </p>
              </div>
              <Target className="w-8 h-8 text-white" />
            </div>
          </CardContent>
        </Card>

        <Card className="bg-neutral-900 border-neutral-700">
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-xs text-neutral-400 tracking-wider">
                  AVG WEIGHT
                </p>
                <p className="text-2xl font-bold text-white font-mono">
                  {avgWeight.toFixed(0)}
                </p>
              </div>
              <Activity className="w-8 h-8 text-white" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* MARK: Performance Progression */}
      <Card className="bg-neutral-900 border-neutral-700">
        <CardHeader>
          <CardTitle className="text-sm font-medium text-neutral-300 tracking-wider">
            PERFORMANCE PROGRESSION
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-48 relative">
            {/* Chart Grid */}
            <div className="absolute inset-0 grid grid-cols-8 grid-rows-6 opacity-20">
              {Array.from({ length: 48 }).map((_, i) => (
                <div key={i} className="border border-neutral-700"></div>
              ))}
            </div>

            {/* Chart Line - Max Weight Progression */}
            <svg className="absolute inset-0 w-full h-full">
              <polyline
                points="0,180 100,160 200,140 300,120 400,100 500,80"
                fill="none"
                stroke="#f97316"
                strokeWidth="3"
              />
              {/* Data points */}
              {[0, 100, 200, 300, 400, 500].map((x, i) => (
                <circle key={i} cx={x} cy={180 - i * 20} r="4" fill="#f97316" />
              ))}
            </svg>

            {/* Y-axis labels */}
            <div className="absolute left-0 top-0 h-full flex flex-col justify-between text-xs text-neutral-500 -ml-8 font-mono">
              <span>250</span>
              <span>200</span>
              <span>150</span>
              <span>100</span>
              <span>50</span>
              <span>0</span>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* MARK: Training Sessions */}
      <Card className="bg-neutral-900 border-neutral-700">
        <CardHeader>
          <CardTitle className="text-sm font-medium text-neutral-300 tracking-wider">
            TRAINING SESSIONS
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {workoutGroups.map((workout) => (
              <div
                key={workout.workout_id}
                className="border border-neutral-700 rounded p-4 hover:border-orange-500/50 transition-colors cursor-pointer"
                onClick={() => setSelectedWorkout(workout)}
              >
                <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
                  <div className="flex items-center gap-4">
                    <div className="text-center">
                      <div className="text-lg font-bold text-white font-mono">
                        WO-{workout.workout_id.toString().padStart(3, '0')}
                      </div>
                      <div className="text-xs text-neutral-400">SESSION</div>
                    </div>
                    <div>
                      <div className="text-sm font-medium text-white">
                        {formatDate(workout.date)}
                      </div>
                      <div className="text-xs text-neutral-400 font-mono">
                        {formatTime(workout.date)}
                      </div>
                    </div>
                  </div>

                  <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-center">
                    <div>
                      <div className="text-lg font-bold text-white font-mono">
                        {workout.setCount}
                      </div>
                      <div className="text-xs text-neutral-400">SETS</div>
                    </div>
                    <div>
                      <div className="text-lg font-bold text-orange-500 font-mono">
                        {workout.maxWeight}
                      </div>
                      <div className="text-xs text-neutral-400">MAX LBS</div>
                    </div>
                    <div>
                      <div className="text-lg font-bold text-white font-mono">
                        {workout.totalReps}
                      </div>
                      <div className="text-xs text-neutral-400">REPS</div>
                    </div>
                    <div>
                      <div className="text-lg font-bold text-white font-mono">
                        {workout.totalVolume.toLocaleString()}
                      </div>
                      <div className="text-xs text-neutral-400">VOLUME</div>
                    </div>
                  </div>
                </div>

                {/* Sets Preview */}
                <div className="mt-4 flex flex-wrap gap-2">
                  {workout.sets.map((set, index) => (
                    <div
                      key={set.id}
                      className="flex items-center gap-2 px-2 py-1 bg-neutral-800 rounded text-xs cursor-pointer hover:bg-neutral-700"
                      onClick={(e) => {
                        e.stopPropagation();
                        setSelectedSet(set);
                      }}
                    >
                      <span className="text-neutral-400">#{index + 1}</span>
                      <Badge className={getSetTypeColor(set.set_type)}>
                        {set.set_type.toUpperCase()}
                      </Badge>
                      <span className="text-white font-mono">
                        {set.weight}×{set.reps}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Workout Detail Modal */}
      {selectedWorkout && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <Card className="bg-neutral-900 border-neutral-700 w-full max-w-4xl max-h-[90vh] overflow-y-auto">
            <CardHeader className="flex flex-row items-center justify-between">
              <div>
                <CardTitle className="text-xl font-bold text-white tracking-wider">
                  WORKOUT SESSION WO-
                  {selectedWorkout.workout_id.toString().padStart(3, '0')}
                </CardTitle>
                <p className="text-sm text-neutral-400">
                  {formatDate(selectedWorkout.date)} at{' '}
                  {formatTime(selectedWorkout.date)}
                </p>
              </div>
              <Button
                variant="ghost"
                onClick={() => setSelectedWorkout(null)}
                className="text-neutral-400 hover:text-white"
              >
                ✕
              </Button>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                <div className="text-center">
                  <div className="text-2xl font-bold text-white font-mono">
                    {selectedWorkout.setCount}
                  </div>
                  <div className="text-xs text-neutral-400">TOTAL SETS</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-orange-500 font-mono">
                    {selectedWorkout.maxWeight}
                  </div>
                  <div className="text-xs text-neutral-400">MAX WEIGHT</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-white font-mono">
                    {selectedWorkout.totalReps}
                  </div>
                  <div className="text-xs text-neutral-400">TOTAL REPS</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-white font-mono">
                    {selectedWorkout.totalVolume.toLocaleString()}
                  </div>
                  <div className="text-xs text-neutral-400">TOTAL VOLUME</div>
                </div>
              </div>

              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="border-b border-neutral-700">
                      <th className="text-left py-3 px-4 text-xs font-medium text-neutral-400 tracking-wider">
                        SET
                      </th>
                      <th className="text-left py-3 px-4 text-xs font-medium text-neutral-400 tracking-wider">
                        TYPE
                      </th>
                      <th className="text-left py-3 px-4 text-xs font-medium text-neutral-400 tracking-wider">
                        WEIGHT
                      </th>
                      <th className="text-left py-3 px-4 text-xs font-medium text-neutral-400 tracking-wider">
                        REPS
                      </th>
                      <th className="text-left py-3 px-4 text-xs font-medium text-neutral-400 tracking-wider">
                        VOLUME
                      </th>
                      <th className="text-left py-3 px-4 text-xs font-medium text-neutral-400 tracking-wider">
                        TIME
                      </th>
                      <th className="text-left py-3 px-4 text-xs font-medium text-neutral-400 tracking-wider">
                        ACTIONS
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {selectedWorkout.sets.map((set, index) => (
                      <tr
                        key={set.id}
                        className="border-b border-neutral-800 hover:bg-neutral-800 transition-colors"
                      >
                        <td className="py-3 px-4 text-sm text-white font-mono">
                          #{index + 1}
                        </td>
                        <td className="py-3 px-4">
                          <Badge className={getSetTypeColor(set.set_type)}>
                            {set.set_type.toUpperCase()}
                          </Badge>
                        </td>
                        <td className="py-3 px-4 text-sm text-white font-mono">
                          {set.weight} lbs
                        </td>
                        <td className="py-3 px-4 text-sm text-white font-mono">
                          {set.reps}
                        </td>
                        <td className="py-3 px-4 text-sm text-orange-500 font-mono">
                          {(set.weight * set.reps).toLocaleString()}
                        </td>
                        <td className="py-3 px-4 text-sm text-neutral-300 font-mono">
                          {formatTime(set.created_at)}
                        </td>
                        <td className="py-3 px-4">
                          <div className="flex gap-1">
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-6 w-6 text-neutral-400 hover:text-orange-500"
                            >
                              <Edit className="h-3 w-3" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-6 w-6 text-neutral-400 hover:text-red-500"
                            >
                              <Trash2 className="h-3 w-3" />
                            </Button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              <div className="flex gap-2 pt-4 border-t border-neutral-700">
                <Button className="bg-orange-500 hover:bg-orange-600 text-white">
                  <Download className="w-4 h-4 mr-2" />
                  Export Data
                </Button>
                <Button
                  variant="outline"
                  className="border-neutral-700 text-neutral-400 hover:bg-neutral-800 hover:text-neutral-300 bg-transparent"
                >
                  View Full Workout
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* MARK: Set Detail Modal */}
      {selectedSet && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <Card className="bg-neutral-900 border-neutral-700 w-full max-w-md">
            <CardHeader className="flex flex-row items-center justify-between">
              <div>
                <CardTitle className="text-lg font-bold text-white tracking-wider">
                  SET DETAILS
                </CardTitle>
                <p className="text-sm text-neutral-400 font-mono">
                  SET-{selectedSet.id.toString().padStart(3, '0')}
                </p>
              </div>
              <Button
                variant="ghost"
                onClick={() => setSelectedSet(null)}
                className="text-neutral-400 hover:text-white"
              >
                ✕
              </Button>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <p className="text-xs text-neutral-400 tracking-wider mb-1">
                    WEIGHT
                  </p>
                  <p className="text-lg font-bold text-white font-mono">
                    {selectedSet.weight} lbs
                  </p>
                </div>
                <div>
                  <p className="text-xs text-neutral-400 tracking-wider mb-1">
                    REPS
                  </p>
                  <p className="text-lg font-bold text-white font-mono">
                    {selectedSet.reps}
                  </p>
                </div>
                <div>
                  <p className="text-xs text-neutral-400 tracking-wider mb-1">
                    SET TYPE
                  </p>
                  <Badge className={getSetTypeColor(selectedSet.set_type)}>
                    {selectedSet.set_type.toUpperCase()}
                  </Badge>
                </div>
                <div>
                  <p className="text-xs text-neutral-400 tracking-wider mb-1">
                    VOLUME
                  </p>
                  <p className="text-lg font-bold text-orange-500 font-mono">
                    {(selectedSet.weight * selectedSet.reps).toLocaleString()}
                  </p>
                </div>
              </div>

              <div>
                <p className="text-xs text-neutral-400 tracking-wider mb-1">
                  CREATED
                </p>
                <p className="text-sm text-white font-mono">
                  {formatDate(selectedSet.created_at)}
                </p>
                <p className="text-sm text-neutral-400 font-mono">
                  {formatTime(selectedSet.created_at)}
                </p>
              </div>

              <div className="flex gap-2 pt-4">
                <Button className="bg-orange-500 hover:bg-orange-600 text-white flex-1">
                  <Edit className="w-4 h-4 mr-2" />
                  Edit Set
                </Button>
                <Button
                  variant="outline"
                  className="border-red-700 text-red-400 hover:bg-red-900/20 hover:text-red-300 bg-transparent"
                >
                  <Trash2 className="w-4 h-4" />
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}
