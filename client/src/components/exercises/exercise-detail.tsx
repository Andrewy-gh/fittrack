import { useMemo, useState } from 'react';
import { ArrowDownAz, ArrowUpAz } from 'lucide-react';
import { Card, CardContent } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Toggle } from '@/components/ui/toggle';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { PaginationControl } from '@/components/ui/pagination-control';
import { sortByExerciseAndSetOrder } from '@/lib/utils';
import type { ExerciseExerciseWithSetsResponse } from '@/client';
import { ExerciseDeleteDialog } from '@/routes/_layout/exercises/-components/exercise-delete-dialog';
import { ExerciseEditDialog } from '@/routes/_layout/exercises/-components/exercise-edit-dialog';
import { ExerciseDetailHeader } from '@/components/exercises/exercise-detail-header';
import { ExerciseSummaryCards } from '@/components/exercises/exercise-summary-cards';
import { ExerciseMetricCharts } from '@/components/exercises/exercise-metric-charts';
import {
  ExerciseWorkoutCards,
  type ExerciseWorkoutEntry,
} from '@/components/exercises/exercise-workout-cards';
import { ExerciseHistorical1RmCard } from '@/routes/_layout/exercises/-components/exercise-historical-1rm';

export interface ExerciseDetailProps {
  exerciseSets: ExerciseExerciseWithSetsResponse[];
  exerciseId: number;
  isDemoMode: boolean;
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
  isDemoMode,
  sortOrder,
  itemsPerPage,
  page,
  onSortOrderChange,
  onItemsPerPageChange,
  onPageChange,
}: ExerciseDetailProps) {
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);

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

  const workoutEntries = useMemo<ExerciseWorkoutEntry[]>(() => {
    const groups = new Map<
      number,
      { workoutId: number; date: string; notes: string | null; sets: ExerciseExerciseWithSetsResponse[] }
    >();
    const ordered: ExerciseWorkoutEntry[] = [];

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
        <ExerciseDetailHeader
          exerciseName={exerciseName}
          onEdit={handleOpenEditDialog}
          onDelete={handleOpenDeleteDialog}
        />

        <ExerciseHistorical1RmCard
          exerciseId={exerciseId}
          exerciseSets={exerciseSets}
          isDemoMode={isDemoMode}
        />

        <ExerciseSummaryCards
          totalSets={totalSets}
          uniqueWorkouts={uniqueWorkouts}
          averageWeight={averageWeight}
          maxWeight={maxWeight}
          averageVolume={averageVolume}
          maxVolume={maxVolume}
        />

        <ExerciseMetricCharts
          exerciseId={exerciseId}
          exerciseSets={exerciseSets}
          isDemoMode={isDemoMode}
        />

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

          <ExerciseWorkoutCards workouts={pagedWorkouts} />
        </div>

        <PaginationControl
          currentPage={currentPage}
          totalPages={totalPages}
          onPageChange={onPageChange}
        />
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
