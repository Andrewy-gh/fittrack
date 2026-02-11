import { useMemo, useState } from 'react';
import { useRouter } from '@tanstack/react-router';
import { toast } from 'sonner';

import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';

import type {
  ExerciseExerciseDetailExerciseResponse,
  ExerciseExerciseWithSetsResponse,
} from '@/client';
import {
  useUpdateExerciseHistorical1RmMutation,
} from '@/lib/api/exercises';
import {
  getDemoExerciseHistorical1Rm,
  recomputeDemoExerciseHistorical1Rm,
  setDemoExerciseHistorical1RmManual,
} from '@/lib/demo-data/historical-1rm';

function computeBestE1rmFromSets(sets: ExerciseExerciseWithSetsResponse[]): { best: number; workoutId: number } | null {
  let best: number | null = null;
  let workoutId: number | null = null;
  for (const s of sets) {
    if (s.set_type !== 'working') continue;
    const w = s.weight ?? 0;
    const reps = s.reps ?? 0;
    const e1rm = w * (1 + reps / 30);
    if (!Number.isFinite(e1rm)) continue;
    if (best == null || e1rm > best) {
      best = e1rm;
      workoutId = s.workout_id;
    }
  }
  if (best == null || workoutId == null) return null;
  return { best, workoutId };
}

export function resolveBestE1rmForDisplay({
  isDemoMode,
  apiBestE1RM,
  exerciseSets,
}: {
  isDemoMode: boolean;
  apiBestE1RM: number | null | undefined;
  exerciseSets: ExerciseExerciseWithSetsResponse[];
}): { best: number | null; workoutId: number | null } {
  const shouldFallbackFromSets = isDemoMode || apiBestE1RM == null;
  const bestFromSets = shouldFallbackFromSets ? computeBestE1rmFromSets(exerciseSets) : null;

  return {
    best: apiBestE1RM ?? bestFromSets?.best ?? null,
    workoutId: bestFromSets?.workoutId ?? null,
  };
}

function fmtLb(n: number | null | undefined): string {
  if (n == null || !Number.isFinite(n)) return '—';
  return `${n.toFixed(1)} lb`;
}

export function ExerciseHistorical1RmCard({
  exercise,
  exerciseId,
  exerciseSets,
  isDemoMode,
}: {
  exercise: ExerciseExerciseDetailExerciseResponse;
  exerciseId: number;
  exerciseSets: ExerciseExerciseWithSetsResponse[];
  isDemoMode: boolean;
}) {
  const [isOpen, setIsOpen] = useState(false);

  const computedBest = useMemo(
    () =>
      resolveBestE1rmForDisplay({
        isDemoMode,
        apiBestE1RM: exercise.best_e1rm,
        exerciseSets,
      }),
    [isDemoMode, exercise.best_e1rm, exerciseSets]
  );
  const demoStored = isDemoMode ? getDemoExerciseHistorical1Rm(exerciseId) : null;

  const stored = isDemoMode
    ? demoStored?.historical_1rm ?? null
    : exercise.historical_1rm ?? null;
  const storedUpdatedAt = isDemoMode
    ? demoStored?.updated_at ?? null
    : exercise.historical_1rm_updated_at ?? null;
  const storedSourceWorkoutId = isDemoMode
    ? demoStored?.source_workout_id ?? null
    : exercise.historical_1rm_source_workout_id ?? null;
  const computed = computedBest.best;
  const computedWorkoutId = computedBest.workoutId;

  const primaryValue = stored ?? computed ?? null;

  return (
    <>
      <Card>
        <CardContent className="pt-6">
          <div className="flex items-start justify-between gap-4">
            <div className="space-y-1">
              <div className="text-[10px] font-medium text-muted-foreground uppercase tracking-wide">
                Historical 1RM
              </div>
              <div className="text-2xl font-semibold tracking-tight">
                {fmtLb(primaryValue)}
              </div>
              <MetaLine
                stored={stored ?? null}
                storedUpdatedAt={storedUpdatedAt ?? null}
                storedSourceWorkoutId={storedSourceWorkoutId ?? null}
                computed={computed ?? null}
                computedWorkoutId={computedWorkoutId ?? null}
              />
            </div>

            <Button
              size="sm"
              variant="outline"
              onClick={() => setIsOpen(true)}
              title="Set historical 1RM"
            >
              Set
            </Button>
          </div>
        </CardContent>
      </Card>

      <ExerciseHistorical1RmDialog
        isOpen={isOpen}
        onOpenChange={setIsOpen}
        exerciseId={exerciseId}
        isDemoMode={isDemoMode}
        stored={stored ?? null}
        computed={computed ?? null}
        computedWorkoutId={computedWorkoutId ?? null}
      />
    </>
  );
}

function MetaLine({
  stored,
  storedUpdatedAt,
  storedSourceWorkoutId,
  computed,
  computedWorkoutId,
}: {
  stored: number | null;
  storedUpdatedAt: string | null;
  storedSourceWorkoutId: number | null;
  computed: number | null;
  computedWorkoutId: number | null;
}) {
  const router = useRouter();
  const onWorkoutClick = (workoutId: number) =>
    router.navigate({ to: '/workouts/$workoutId', params: { workoutId } });

  if (stored != null) {
    const updated = storedUpdatedAt ? new Date(storedUpdatedAt) : null;
    const updatedText = updated && !isNaN(updated.getTime())
      ? `Updated ${updated.toLocaleDateString()}`
      : 'Updated';

    return (
      <div className="text-xs text-muted-foreground">
        {updatedText}
        {storedSourceWorkoutId != null ? (
          <>
            {' '}
            •{' '}
            <button
              type="button"
              className="underline underline-offset-2"
              onClick={() => onWorkoutClick(storedSourceWorkoutId)}
            >
              Workout {storedSourceWorkoutId}
            </button>
          </>
        ) : (
          <> • Manual</>
        )}
      </div>
    );
  }

  if (computed != null) {
    return (
      <div className="text-xs text-muted-foreground">
        Not set. Best e1RM: {fmtLb(computed)}
        {computedWorkoutId != null ? (
          <>
            {' '}
            •{' '}
            <button
              type="button"
              className="underline underline-offset-2"
              onClick={() => onWorkoutClick(computedWorkoutId)}
            >
              Workout {computedWorkoutId}
            </button>
          </>
        ) : null}
      </div>
    );
  }

  return (
    <div className="text-xs text-muted-foreground">
      Not set yet.
    </div>
  );
}

function ExerciseHistorical1RmDialog({
  isOpen,
  onOpenChange,
  exerciseId,
  isDemoMode,
  stored,
  computed,
  computedWorkoutId,
}: {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  exerciseId: number;
  isDemoMode: boolean;
  stored: number | null;
  computed: number | null;
  computedWorkoutId: number | null;
}) {
  const [value, setValue] = useState<string>('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const mutation = useUpdateExerciseHistorical1RmMutation();

  const defaultValue = useMemo(() => {
    if (stored != null) return String(stored);
    if (computed != null) return String(computed);
    return '';
  }, [stored, computed]);

  const handleClose = (open: boolean) => {
    onOpenChange(open);
    if (!open) {
      setValue('');
      setIsSubmitting(false);
    }
  };

  const submitManual = async () => {
    const trimmed = value.trim();
    const parsed = trimmed === '' ? null : Number(trimmed);
    if (parsed != null && (!Number.isFinite(parsed) || parsed < 0)) {
      toast.error('Enter a valid number');
      return;
    }

    setIsSubmitting(true);
    try {
      if (isDemoMode) {
        setDemoExerciseHistorical1RmManual(exerciseId, parsed);
      } else {
        await mutation.mutateAsync({
          path: { id: exerciseId },
          body: parsed == null
            ? { mode: 'manual' }
            : { mode: 'manual', historical_1rm: parsed },
        });
      }
      toast.success('Historical 1RM updated');
      handleClose(false);
    } catch {
      toast.error('Failed to update historical 1RM');
    } finally {
      setIsSubmitting(false);
    }
  };

  const submitRecompute = async () => {
    setIsSubmitting(true);
    try {
      if (isDemoMode) {
        recomputeDemoExerciseHistorical1Rm(exerciseId);
      } else {
        await mutation.mutateAsync({
          path: { id: exerciseId },
          body: { mode: 'recompute' },
        });
      }
      toast.success('Historical 1RM recomputed');
      handleClose(false);
    } catch {
      toast.error('Failed to recompute historical 1RM');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Historical 1RM</DialogTitle>
          <DialogDescription>
            Set a manual value, or recompute from your best working-set e1RM.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-3 py-2">
          <div className="space-y-1.5">
            <Label htmlFor="historical-1rm-input">Manual 1RM (lb)</Label>
            <Input
              id="historical-1rm-input"
              type="number"
              inputMode="decimal"
              step="0.1"
              placeholder={defaultValue || 'e.g. 315'}
              value={value}
              onChange={(e) => setValue(e.target.value)}
              disabled={isSubmitting}
            />
            <div className="text-xs text-muted-foreground">
              Stored: {fmtLb(stored)} • Best e1RM: {fmtLb(computed)}
              {computedWorkoutId != null ? ` (workout ${computedWorkoutId})` : ''}
            </div>
          </div>
        </div>

        <DialogFooter className="flex gap-2 sm:gap-2">
          <Button
            type="button"
            variant="outline"
            onClick={() => handleClose(false)}
            disabled={isSubmitting}
          >
            Cancel
          </Button>
          <Button
            type="button"
            variant="outline"
            onClick={submitRecompute}
            disabled={isSubmitting}
          >
            Use Best
          </Button>
          <Button
            type="button"
            onClick={submitManual}
            disabled={isSubmitting}
          >
            Save
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
