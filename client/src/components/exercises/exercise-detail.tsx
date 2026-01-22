import { useMemo, useState } from 'react';
import {
  Activity,
  ArrowDownAz,
  ArrowUpAz,
  BarChart3,
  Calendar,
  Edit,
  Hash,
  Trash,
  TrendingUp,
  Weight,
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { ChartBarVol } from '@/components/charts/chart-bar-vol';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Toggle } from '@/components/ui/toggle';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from '@/components/ui/pagination';
import { formatDate, formatTime, formatWeight, sortByExerciseAndSetOrder } from '@/lib/utils';
import type { ExerciseExerciseWithSetsResponse } from '@/client';
import { ExerciseDeleteDialog } from '@/routes/_layout/exercises/-components/exercise-delete-dialog';
import { ExerciseEditDialog } from '@/routes/_layout/exercises/-components/exercise-edit-dialog';

export interface ExerciseDetailProps {
  exerciseSets: ExerciseExerciseWithSetsResponse[];
  exerciseId: number;
  sortOrder: 'asc' | 'desc';
  itemsPerPage: number;
  page?: number;
  onSortOrderChange: (sortOrder: 'asc' | 'desc') => void;
  onItemsPerPageChange: (itemsPerPage: number) => void;
  onPageChange: (page: number) => void;
}

export function ExerciseDetail({
  exerciseSets,
  exerciseId,
  sortOrder,
  itemsPerPage,
  page,
  onSortOrderChange,
  onItemsPerPageChange,
  onPageChange,
}: ExerciseDetailProps) {
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);
  // Calculate summary statistics
  const totalSets = exerciseSets.length;
  const uniqueWorkouts = new Set(exerciseSets.map((set) => set.workout_id))
    .size;
  const weights = exerciseSets.map((set) => set.weight || 0);
  const volumes = exerciseSets.map((set) => set.volume);

  const averageWeight = totalSets > 0
    ? weights.reduce((sum, weight) => sum + weight, 0) / weights.length
    : 0;
  const maxWeight = totalSets > 0 ? Math.max(...weights) : 0;
  const averageVolume = totalSets > 0
    ? Math.round(volumes.reduce((sum, volume) => sum + volume, 0) / volumes.length)
    : 0;
  const maxVolume = totalSets > 0 ? Math.max(...volumes) : 0;

  const sortedExerciseSets = sortByExerciseAndSetOrder(exerciseSets);

  const workoutEntries = useMemo(() => {
    const groups = new Map<
      number,
      { workoutId: number; date: string; notes: string | null; sets: typeof exerciseSets }
    >();
    const ordered: Array<{
      workoutId: number;
      date: string;
      notes: string | null;
      sets: typeof exerciseSets;
    }> = [];

    sortedExerciseSets.forEach((set) => {
      let group = groups.get(set.workout_id);
      if (!group) {
        group = {
          workoutId: set.workout_id,
          date: set.workout_date,
          notes: set.workout_notes || null,
          sets: [],
        };
        groups.set(set.workout_id, group);
        ordered.push(group);
      }
      group.sets.push(set);
    });

    return ordered;
  }, [sortedExerciseSets]);

  const sortedWorkouts = useMemo(() => {
    const direction = sortOrder === 'asc' ? 1 : -1;
    return [...workoutEntries].sort((a, b) => {
      const aTime = new Date(a.date).getTime();
      const bTime = new Date(b.date).getTime();
      return (aTime - bTime) * direction;
    });
  }, [sortOrder, workoutEntries]);

  const totalPages = Math.max(
    1,
    Math.ceil(sortedWorkouts.length / itemsPerPage)
  );
  const currentPage = Math.min(Math.max(1, page ?? 1), totalPages);
  const startIndex = (currentPage - 1) * itemsPerPage;
  const pagedWorkouts = sortedWorkouts.slice(
    startIndex,
    startIndex + itemsPerPage
  );

  const exerciseName = exerciseSets[0]?.exercise_name || 'Exercise';

  const handleOpenEditDialog = () => {
    setIsEditDialogOpen(true);
  };

  const handleOpenDeleteDialog = () => {
    setIsDeleteDialogOpen(true);
  };

  return (
    <main>
      <div className="max-w-lg mx-auto space-y-6 px-4 pb-8">
        {/* Header */}
        <div className="flex items-center justify-between pt-4">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">
              {exerciseName}
            </h1>
          </div>
          <div className="flex flex-col items-center gap-3 md:flex-row">
            <Button size="sm" onClick={handleOpenEditDialog}>
              <Edit className="mr-2 hidden h-4 w-4 md:block" />
              Edit
            </Button>
            <Button
              size="sm"
              variant="outline"
              onClick={handleOpenDeleteDialog}
              data-testid="delete-exercise-button"
            >
              <Trash className="mr-2 hidden h-4 w-4 md:block" />
              Delete
            </Button>
          </div>
        </div>

        {/* MARK: Summary Cards */}
        <div className="grid grid-cols-2 gap-4">
          <Card className="p-4">
            <div className="flex items-center gap-2 mb-2">
              <Hash className="w-5 h-5 text-primary" />
              <span className="text-sm font-semibold">Total Sets</span>
            </div>
            <div className="text-2xl text-card-foreground font-bold">
              {totalSets}
            </div>
          </Card>

          <Card className="p-4">
            <div className="flex items-center gap-2 mb-2">
              <Calendar className="w-5 h-5 text-primary" />
              <span className="text-sm font-semibold">Workouts</span>
            </div>
            <div className="text-2xl text-card-foreground font-bold">
              {uniqueWorkouts}
            </div>
          </Card>

          <Card className="p-4">
            <div className="flex items-center gap-2 mb-2">
              <Weight className="w-5 h-5 text-primary" />
              <span className="text-sm font-semibold">Average Weight</span>
            </div>
            <div className="text-2xl text-card-foreground font-bold">
              {formatWeight(averageWeight)} lbs
            </div>
          </Card>

          <Card className="p-4">
            <div className="flex items-center gap-2 mb-2">
              <TrendingUp className="w-5 h-5 text-primary" />
              <span className="text-sm font-semibold">Max Weight</span>
            </div>
            <div className="text-2xl text-card-foreground font-bold">
              {formatWeight(maxWeight)} lbs
            </div>
          </Card>

          <Card className="p-4">
            <div className="flex items-center gap-2 mb-2">
              <BarChart3 className="w-5 h-5 text-primary" />
              <span className="text-sm font-semibold hidden md:inline">
                Average Volume
              </span>
              <span className="text-sm font-semibold md:hidden">
                Avg. Volume
              </span>
            </div>
            <div className="text-2xl text-card-foreground font-bold">
              {averageVolume.toLocaleString()}
            </div>
          </Card>

          <Card className="p-4">
            <div className="flex items-center gap-2 mb-2">
              <Activity className="w-5 h-5 text-primary" />
              <span className="text-sm font-semibold">Max Volume</span>
            </div>
            <div className="text-2xl text-card-foreground font-bold">
              {maxVolume.toLocaleString()}
            </div>
          </Card>
        </div>
        <ChartBarVol data={exerciseSets} />

        {/* MARK: Workouts */}
        <div className="space-y-4">
          <h2 className="text-xl font-semibold">Workouts</h2>
          <Card>
            <CardContent>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div className="space-y-1.5">
                  <Label className="text-[10px] font-medium text-muted-foreground uppercase tracking-wide">
                    Sort Order
                  </Label>
                  <Toggle
                    pressed={sortOrder === 'asc'}
                    onPressedChange={(pressed) =>
                      onSortOrderChange(pressed ? 'asc' : 'desc')
                    }
                    aria-label="Toggle sort order"
                    className="px-3 py-1.5 inline-flex items-center justify-start gap-2"
                  >
                    <span className="text-sm">
                      {sortOrder === 'asc' ? 'Ascending' : 'Descending'}
                    </span>
                    {sortOrder === 'asc' ? (
                      <ArrowUpAz className="w-4 h-4 opacity-70" />
                    ) : (
                      <ArrowDownAz className="w-4 h-4 opacity-70" />
                    )}
                  </Toggle>
                </div>
              </div>
            </CardContent>
          </Card>

          <div className="flex items-center gap-2 justify-center px-1">
            <Label htmlFor="items-per-page" className="text-xs whitespace-nowrap">
              Show
            </Label>
            <Select
              value={String(itemsPerPage)}
              onValueChange={(value) => onItemsPerPageChange(Number(value))}
            >
              <SelectTrigger id="items-per-page" className="h-8 w-[70px]">
                <SelectValue placeholder="10" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="10">10</SelectItem>
                <SelectItem value="20">20</SelectItem>
                <SelectItem value="50">50</SelectItem>
              </SelectContent>
            </Select>
            <span className="text-xs text-muted-foreground">per page</span>
          </div>

          {pagedWorkouts.length === 0 && (
            <Card>
              <CardContent className="py-6 text-sm text-muted-foreground">
                No workouts logged for this exercise yet.
              </CardContent>
            </Card>
          )}
          {pagedWorkouts.map((workout) => {
            const exerciseReps = workout.sets.reduce(
              (sum, set) => sum + set.reps,
              0
            );
            const exerciseVolume = workout.sets.reduce(
              (sum, set) => sum + set.volume,
              0
            );
            return (
              <Card
                key={workout.workoutId}
                className="border-0 shadow-sm backdrop-blur-sm"
              >
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <div>
                      <CardTitle className="text-lg font-semibold">
                        {formatDate(workout.date)}
                      </CardTitle>
                      <div className="flex items-center gap-2 mt-1">
                        <p className="text-sm text-muted-foreground">
                          {formatTime(workout.date)}
                        </p>
                        {workout.notes && (
                          <>
                            <span className="text-muted-foreground">•</span>
                            <Badge
                              variant="outline"
                              className="border-border bg-muted text-xs"
                            >
                              {workout.notes.toUpperCase()}
                            </Badge>
                          </>
                        )}
                      </div>
                    </div>
                    <div className="flex items-center gap-4 text-sm text-muted-foreground">
                      <span>{exerciseReps} reps</span>
                      <span className="text-primary">
                        {exerciseVolume.toLocaleString()} vol
                      </span>
                    </div>
                  </div>
                </CardHeader>
                <CardContent className="space-y-2">
                  {workout.sets.map((set, index) => (
                    <div
                      key={set.set_id}
                      className="flex items-center justify-between py-2 px-3 rounded-lg bg-muted/50"
                    >
                      <div className="flex items-center space-x-4">
                        <span className="text-sm font-medium text-muted-foreground w-8">
                          {set.set_order ?? index + 1}
                        </span>
                        <div className="flex items-center space-x-4 text-sm">
                          <span className="font-medium">{formatWeight(set.weight)} lbs</span>
                          <span>&times;</span>
                          <span className="font-medium">{set.reps} reps</span>
                        </div>
                      </div>
                      <div className="text-sm text-muted-foreground">
                        {set.volume.toLocaleString()} vol
                      </div>
                    </div>
                  ))}
                </CardContent>
              </Card>
            );
          })}
        </div>

        {totalPages > 1 && (
          <Pagination className="flex justify-center py-2">
            <PaginationContent>
              <PaginationItem>
                <PaginationPrevious
                  href="#"
                  onClick={(event) => {
                    event.preventDefault();
                    onPageChange(Math.max(1, currentPage - 1));
                  }}
                />
              </PaginationItem>
              <PaginationItem className="hidden sm:inline-block">
                <PaginationLink href="#" isActive>
                  {currentPage}
                </PaginationLink>
              </PaginationItem>
              <PaginationItem>
                <PaginationEllipsis />
              </PaginationItem>
              <PaginationItem>
                <PaginationNext
                  href="#"
                  onClick={(event) => {
                    event.preventDefault();
                    onPageChange(Math.min(totalPages, currentPage + 1));
                  }}
                />
              </PaginationItem>
            </PaginationContent>
          </Pagination>
        )}
        {/* MARK: Dialogs */}
        <ExerciseEditDialog
          isOpen={isEditDialogOpen}
          onOpenChange={setIsEditDialogOpen}
          exerciseId={exerciseId}
          exerciseName={exerciseName}
        />
        <ExerciseDeleteDialog
          isOpen={isDeleteDialogOpen}
          onOpenChange={setIsDeleteDialogOpen}
          exerciseId={exerciseId}
        />
      </div>
    </main>
  );
}
