import { useEffect, useMemo, useState } from 'react';
import { Target } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  formatExerciseGoalSummary,
  getExerciseGoal,
  parseExerciseGoalInput,
  saveExerciseGoal,
  type ExerciseGoal,
} from '@/lib/exercise-goals';

type GoalFieldErrors = {
  targetWeight?: string;
  targetReps?: string;
  frequencyPerWeek?: string;
};

function toInputValue(value?: number): string {
  return value == null ? '' : String(value);
}

export function ExerciseGoalsCard({
  exerciseId,
  exerciseName,
}: {
  exerciseId: number;
  exerciseName: string;
}) {
  const [isOpen, setIsOpen] = useState(false);
  const [goal, setGoal] = useState<ExerciseGoal>(() =>
    getExerciseGoal({ exerciseId, exerciseName }) ?? {}
  );
  const [targetWeight, setTargetWeight] = useState('');
  const [targetReps, setTargetReps] = useState('');
  const [frequencyPerWeek, setFrequencyPerWeek] = useState('');
  const [errors, setErrors] = useState<GoalFieldErrors>({});

  useEffect(() => {
    const storedGoal = getExerciseGoal({ exerciseId, exerciseName }) ?? {};
    setGoal(storedGoal);
  }, [exerciseId, exerciseName]);

  useEffect(() => {
    if (!isOpen) {
      setTargetWeight(toInputValue(goal.targetWeight));
      setTargetReps(toInputValue(goal.targetReps));
      setFrequencyPerWeek(toInputValue(goal.frequencyPerWeek));
      setErrors({});
    }
  }, [goal, isOpen]);

  const goalSummary = useMemo(
    () => formatExerciseGoalSummary(goal),
    [goal]
  );

  return (
    <>
      <Card>
        <CardContent className="pt-6">
          <div className="flex items-start justify-between gap-4">
            <div className="space-y-1">
              <div className="flex items-center gap-2 text-[10px] font-medium uppercase tracking-wide text-muted-foreground">
                <Target className="h-4 w-4 text-primary" />
                Simple Goals
              </div>
              <div className="text-2xl font-semibold tracking-tight">
                {goalSummary ?? 'No goal set'}
              </div>
              <p className="text-xs text-muted-foreground">
                Save a target weight, reps, or weekly frequency for this lift.
              </p>
            </div>
            <Button size="sm" variant="outline" onClick={() => setIsOpen(true)}>
              {goalSummary ? 'Edit' : 'Set'}
            </Button>
          </div>
        </CardContent>
      </Card>

      <Dialog open={isOpen} onOpenChange={setIsOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Exercise Goals</DialogTitle>
            <DialogDescription>
              Keep goals lightweight. Add only the targets you actually use.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-2">
            <div className="space-y-2">
              <Label htmlFor="exercise-goal-weight">Target Weight (lb)</Label>
              <Input
                id="exercise-goal-weight"
                inputMode="decimal"
                placeholder="225"
                value={targetWeight}
                onChange={(event) => {
                  setTargetWeight(event.target.value);
                  setErrors((prev) => ({ ...prev, targetWeight: undefined }));
                }}
              />
              {errors.targetWeight && (
                <p className="text-sm text-destructive">{errors.targetWeight}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="exercise-goal-reps">Target Reps</Label>
              <Input
                id="exercise-goal-reps"
                inputMode="numeric"
                placeholder="5"
                value={targetReps}
                onChange={(event) => {
                  setTargetReps(event.target.value);
                  setErrors((prev) => ({ ...prev, targetReps: undefined }));
                }}
              />
              {errors.targetReps && (
                <p className="text-sm text-destructive">{errors.targetReps}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="exercise-goal-frequency">Weekly Frequency</Label>
              <Input
                id="exercise-goal-frequency"
                inputMode="numeric"
                placeholder="2"
                value={frequencyPerWeek}
                onChange={(event) => {
                  setFrequencyPerWeek(event.target.value);
                  setErrors((prev) => ({ ...prev, frequencyPerWeek: undefined }));
                }}
              />
              {errors.frequencyPerWeek && (
                <p className="text-sm text-destructive">
                  {errors.frequencyPerWeek}
                </p>
              )}
            </div>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setIsOpen(false)}>
              Cancel
            </Button>
            <Button
              type="button"
              onClick={() => {
                const parsedTargetWeight = parseExerciseGoalInput(
                  targetWeight,
                  'Target Weight'
                );
                const parsedTargetReps = parseExerciseGoalInput(
                  targetReps,
                  'Target Reps',
                  { integer: true }
                );
                const parsedFrequencyPerWeek = parseExerciseGoalInput(
                  frequencyPerWeek,
                  'Weekly Frequency',
                  { integer: true }
                );

                const nextErrors: GoalFieldErrors = {
                  targetWeight: parsedTargetWeight.error,
                  targetReps: parsedTargetReps.error,
                  frequencyPerWeek: parsedFrequencyPerWeek.error,
                };

                if (Object.values(nextErrors).some(Boolean)) {
                  setErrors(nextErrors);
                  return;
                }

                const nextGoal = {
                  targetWeight: parsedTargetWeight.value,
                  targetReps: parsedTargetReps.value,
                  frequencyPerWeek: parsedFrequencyPerWeek.value,
                };

                saveExerciseGoal({ exerciseId, exerciseName }, nextGoal);
                setGoal(nextGoal);
                setIsOpen(false);
              }}
            >
              Save
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
