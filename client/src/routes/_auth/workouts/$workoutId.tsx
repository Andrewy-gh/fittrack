import { createFileRoute, useRouter, Link } from '@tanstack/react-router';
import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  ArrowLeft,
  Calendar,
  Clock,
  Target,
  TrendingUp,
  Weight,
  RotateCcw,
  Plus,
  Edit,
  Trash2,
  BarChart3,
} from 'lucide-react';
import { fetchWorkoutById, type WorkoutSet } from '@/lib/api/workouts';

interface ExerciseGroup {
  exercise_id: number;
  exercise_name: string;
  sets: WorkoutSet[];
  totalVolume: number;
  maxWeight: number;
  totalReps: number;
}

export const Route = createFileRoute('/_auth/workouts/$workoutId')({
  params: {
    parse: (params) => {
      const workoutId = parseInt(params.workoutId, 10);
      if (isNaN(workoutId) || !Number.isInteger(workoutId)) {
        throw new Error('Invalid workoutId');
      }
      return { workoutId };
    },
  },
  loader: async ({ context, params }) => {
    const user = context.user;
    if (!user) {
      throw new Error('User not found');
    }
    const { accessToken } = await user.getAuthJson();
    if (!accessToken) {
      throw new Error('Access token not found');
    }
    const workoutId = params.workoutId;
    const workout = await fetchWorkoutById(workoutId, accessToken);
    return workout;
  },
  component: RouteComponent,
});

function RouteComponent() {
  const workout = Route.useLoaderData();
  return <IndividualWorkoutPage workoutData={workout} />;
}

function IndividualWorkoutPage({ workoutData }: { workoutData: WorkoutSet[] }) {
  const router = useRouter();
  const [selectedExercise, setSelectedExercise] =
    useState<ExerciseGroup | null>(null);

  // Group exercises and calculate stats
  const exerciseGroups: ExerciseGroup[] = workoutData.reduce((acc, set) => {
    const existingExercise = acc.find(
      (ex) => ex.exercise_id === set.exercise_id
    );

    if (existingExercise) {
      existingExercise.sets.push(set);
      existingExercise.totalVolume += set.weight * set.reps;
      existingExercise.maxWeight = Math.max(
        existingExercise.maxWeight,
        set.weight
      );
      existingExercise.totalReps += set.reps;
    } else {
      acc.push({
        exercise_id: set.exercise_id,
        exercise_name: set.exercise_name,
        sets: [set],
        totalVolume: set.weight * set.reps,
        maxWeight: set.weight,
        totalReps: set.reps,
      });
    }

    return acc;
  }, [] as ExerciseGroup[]);

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      weekday: 'long',
      month: 'long',
      day: 'numeric',
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
      case 'dropset':
        return 'bg-red-500/20 text-red-500';
      default:
        return 'bg-neutral-500/20 text-neutral-300';
    }
  };

  // Calculate workout totals
  const totalVolume = exerciseGroups.reduce(
    (sum, ex) => sum + ex.totalVolume,
    0
  );
  const totalSets = workoutData.length;
  const totalReps = exerciseGroups.reduce((sum, ex) => sum + ex.totalReps, 0);
  const exerciseCount = exerciseGroups.length;

  const workoutInfo = workoutData[0]; // Get workout info from first set

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div className="flex items-center gap-4">
          <Button
            variant="ghost"
            onClick={() => {
              router.history.back();
            }}
            className="text-neutral-400 hover:text-orange-500 p-2"
          >
            <ArrowLeft className="w-5 h-5" />
          </Button>
          <div>
            <h1 className="text-2xl font-bold text-white tracking-wider">
              WORKOUT SESSION WO-
              {workoutInfo.workout_id.toString().padStart(3, '0')}
            </h1>
            <p className="text-sm text-neutral-400">
              {workoutInfo.workout_notes} Training Protocol
            </p>
          </div>
        </div>
        <div className="flex gap-2">
          <Button className="bg-orange-500 hover:bg-orange-600 text-white">
            <Edit className="w-4 h-4 mr-2" />
            Edit Session
          </Button>
          <Button className="bg-orange-500 hover:bg-orange-600 text-white">
            <BarChart3 className="w-4 h-4 mr-2" />
            Analytics
          </Button>
        </div>
      </div>

      {/* Workout Info */}
      <Card>
        <CardHeader>
          <CardTitle className="text-sm font-medium text-neutral-300 tracking-wider">
            SESSION OVERVIEW
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
            <div className="text-center">
              <div className="flex items-center justify-center gap-2 mb-2">
                <Calendar className="w-4 h-4 text-neutral-400" />
              </div>
              <div className="text-lg font-bold text-white font-mono">
                {formatDate(workoutInfo.workout_date)}
              </div>
              <div className="text-xs text-neutral-400">DATE</div>
            </div>
            <div className="text-center">
              <div className="flex items-center justify-center gap-2 mb-2">
                <Clock className="w-4 h-4 text-neutral-400" />
              </div>
              <div className="text-lg font-bold text-white font-mono">
                {formatTime(workoutInfo.workout_date)}
              </div>
              <div className="text-xs text-neutral-400">TIME</div>
            </div>
            <div className="text-center">
              <div className="flex items-center justify-center gap-2 mb-2">
                <Target className="w-4 h-4 text-neutral-400" />
              </div>
              <div className="text-lg font-bold text-white font-mono">
                {exerciseCount}
              </div>
              <div className="text-xs text-neutral-400">EXERCISES</div>
            </div>
            <div className="text-center">
              <div className="flex items-center justify-center gap-2 mb-2">
                <RotateCcw className="w-4 h-4 text-neutral-400" />
              </div>
              <div className="text-lg font-bold text-white font-mono">
                {totalSets}
              </div>
              <div className="text-xs text-neutral-400">TOTAL SETS</div>
            </div>
            <div className="text-center">
              <div className="flex items-center justify-center gap-2 mb-2">
                <TrendingUp className="w-4 h-4 text-neutral-400" />
              </div>
              <div className="text-lg font-bold text-white font-mono">
                {totalReps}
              </div>
              <div className="text-xs text-neutral-400">TOTAL REPS</div>
            </div>
            <div className="text-center">
              <div className="flex items-center justify-center gap-2 mb-2">
                <Weight className="w-4 h-4 text-neutral-400" />
              </div>
              <div className="text-lg font-bold text-orange-500 font-mono">
                {totalVolume.toLocaleString()}
              </div>
              <div className="text-xs text-neutral-400">VOLUME (LBS)</div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Exercise List */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {exerciseGroups.map((exercise, index) => (
          <Card
            key={exercise.exercise_id}
            className="hover:border-orange-500/50 transition-colors cursor-pointer"
            onClick={() => setSelectedExercise(exercise)}
          >
            <CardHeader className="pb-3">
              <div className="flex items-start justify-between">
                <div>
                  <CardTitle className="text-sm font-bold text-white tracking-wider">
                    <Link
                      to="/exercises/$exerciseId"
                      params={{ exerciseId: exercise.exercise_id }}
                      className="hover:text-orange-500 transition-colors"
                      onClick={(e) => e.stopPropagation()}
                    >
                      {exercise.exercise_name.toUpperCase()}
                    </Link>
                  </CardTitle>
                  <p className="text-xs text-neutral-400 font-mono">
                    EX-{exercise.exercise_id.toString().padStart(3, '0')} •{' '}
                    {exercise.sets.length} Sets
                  </p>
                </div>
                <Badge className="bg-orange-500/20 text-orange-500">
                  {exercise.sets.length} SETS
                </Badge>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-3 gap-4 text-xs">
                <div>
                  <div className="text-neutral-400 mb-1">MAX WEIGHT</div>
                  <div className="text-white font-mono">
                    {exercise.maxWeight} lbs
                  </div>
                </div>
                <div>
                  <div className="text-neutral-400 mb-1">TOTAL REPS</div>
                  <div className="text-white font-mono">
                    {exercise.totalReps}
                  </div>
                </div>
                <div>
                  <div className="text-neutral-400 mb-1">VOLUME</div>
                  <div className="text-orange-500 font-mono">
                    {exercise.totalVolume.toLocaleString()}
                  </div>
                </div>
              </div>

              <div className="space-y-2">
                <div className="text-xs text-neutral-400 tracking-wider">
                  SET BREAKDOWN
                </div>
                {exercise.sets.map((set, setIndex) => (
                  <div
                    key={set.set_id}
                    className="flex items-center justify-between p-2 bg-neutral-800 rounded text-xs"
                  >
                    <div className="flex items-center gap-3">
                      <span className="text-neutral-400 font-mono">
                        #{setIndex + 1}
                      </span>
                      <Badge className={getSetTypeColor(set.set_type)}>
                        {set.set_type.toUpperCase()}
                      </Badge>
                    </div>
                    <div className="flex items-center gap-4 font-mono">
                      <span className="text-white">
                        {set.weight > 0 ? `${set.weight} lbs` : 'BW'}
                      </span>
                      <span className="text-neutral-400">×</span>
                      <span className="text-white">{set.reps} reps</span>
                      <span className="text-orange-500">
                        {(set.weight * set.reps).toLocaleString()} vol
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* MARK: Modal */}
      {selectedExercise && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <Card className="w-full max-w-4xl max-h-[90vh] overflow-y-auto">
            <CardHeader className="flex flex-row items-center justify-between">
              <div>
                <CardTitle className="text-xl font-bold text-white tracking-wider">
                  <Link
                    to="/exercises/$exerciseId"
                    params={{ exerciseId: selectedExercise.exercise_id }}
                    className="hover:text-orange-500 transition-colors"
                    onClick={(e) => e.stopPropagation()}
                  >
                    {selectedExercise.exercise_name.toUpperCase()}
                  </Link>
                </CardTitle>
                <p className="text-sm text-neutral-400 font-mono">
                  EX-{selectedExercise.exercise_id.toString().padStart(3, '0')}{' '}
                  • {selectedExercise.sets.length} Sets
                </p>
              </div>
              <Button
                variant="ghost"
                onClick={() => setSelectedExercise(null)}
                className="text-neutral-400 hover:text-white"
              >
                ✕
              </Button>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div className="space-y-4">
                  <div>
                    <h3 className="text-sm font-medium text-neutral-300 tracking-wider mb-2">
                      PERFORMANCE METRICS
                    </h3>
                    <div className="space-y-2 text-sm">
                      <div className="flex justify-between">
                        <span className="text-neutral-400">Max Weight:</span>
                        <span className="text-white font-mono">
                          {selectedExercise.maxWeight} lbs
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-neutral-400">Total Volume:</span>
                        <span className="text-orange-500 font-mono">
                          {selectedExercise.totalVolume.toLocaleString()} lbs
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-neutral-400">Total Reps:</span>
                        <span className="text-white font-mono">
                          {selectedExercise.totalReps}
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-neutral-400">
                          Average Reps/Set:
                        </span>
                        <span className="text-white font-mono">
                          {(
                            selectedExercise.totalReps /
                            selectedExercise.sets.length
                          ).toFixed(1)}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>

                <div className="md:col-span-2">
                  <h3 className="text-sm font-medium text-neutral-300 tracking-wider mb-2">
                    SET DETAILS
                  </h3>
                  <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                      <thead>
                        <tr className="border-b border-neutral-700">
                          <th className="text-left py-2 text-xs font-medium text-neutral-400 tracking-wider">
                            SET
                          </th>
                          <th className="text-left py-2 text-xs font-medium text-neutral-400 tracking-wider">
                            TYPE
                          </th>
                          <th className="text-left py-2 text-xs font-medium text-neutral-400 tracking-wider">
                            WEIGHT
                          </th>
                          <th className="text-left py-2 text-xs font-medium text-neutral-400 tracking-wider">
                            REPS
                          </th>
                          <th className="text-left py-2 text-xs font-medium text-neutral-400 tracking-wider">
                            VOLUME
                          </th>
                        </tr>
                      </thead>
                      <tbody>
                        {selectedExercise.sets.map((set, index) => (
                          <tr
                            key={set.set_id}
                            className="border-b border-neutral-800"
                          >
                            <td className="py-2 text-white font-mono">
                              #{index + 1}
                            </td>
                            <td className="py-2">
                              <Badge className={getSetTypeColor(set.set_type)}>
                                {set.set_type.toUpperCase()}
                              </Badge>
                            </td>
                            <td className="py-2 text-white font-mono">
                              {set.weight > 0
                                ? `${set.weight} lbs`
                                : 'Bodyweight'}
                            </td>
                            <td className="py-2 text-white font-mono">
                              {set.reps}
                            </td>
                            <td className="py-2 text-orange-500 font-mono">
                              {(set.weight * set.reps).toLocaleString()} lbs
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              </div>

              <div className="flex gap-2 pt-4 border-t border-neutral-700">
                <Button className="bg-orange-500 hover:bg-orange-600 text-white">
                  <Plus className="w-4 h-4 mr-2" />
                  Add Set
                </Button>
                <Button
                  variant="outline"
                  className="border-neutral-700 text-neutral-400 hover:bg-neutral-800 hover:text-neutral-300 bg-transparent"
                >
                  <Edit className="w-4 h-4 mr-2" />
                  Edit Exercise
                </Button>
                <Button
                  variant="outline"
                  className="border-neutral-700 text-red-400 hover:bg-red-900/20 hover:text-red-300 bg-transparent"
                >
                  <Trash2 className="w-4 h-4 mr-2" />
                  Remove
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}
