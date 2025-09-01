import { useState } from 'react';
import { useFieldContext } from '@/hooks/form';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { GenericCombobox } from '@/components/generic-combobox';
import { Target, X } from 'lucide-react';
import type { WorkoutFocus } from '@/lib/api/workouts';

export default function WorkoutFocusCombobox({
  workoutsFocus,
}: {
  workoutsFocus: WorkoutFocus[]; // Input: workout focus values from the database
}) {
  const field = useFieldContext<string>();
  const [open, setOpen] = useState(false);
  const [selectedWorkoutFocus, setSelectedWorkoutFocus] = useState<WorkoutFocus>();

  // Working list of workout focus values that can include both DB and manually created ones
  const [workingWorkoutsFocus, setWorkingWorkoutsFocus] = useState<WorkoutFocus[]>(
    workoutsFocus
  );

  function handleSelect(option: WorkoutFocus) {
    setSelectedWorkoutFocus(option);
  }

  function handleAppendGroup(name: string) {
    // For new workout focus values, just add them to the working list
    const newWorkoutFocus = { name };
    setWorkingWorkoutsFocus((prev) => [...prev, newWorkoutFocus]);
    handleSelect(newWorkoutFocus);
  }

  function handleAddWorkoutFocus() {
    field.setValue(selectedWorkoutFocus?.name ?? '');
    setOpen(false);
    setSelectedWorkoutFocus(undefined);
  }

  function handleReset() {
    field.setValue('');
    setSelectedWorkoutFocus(undefined);
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Card className="p-4" aria-label="Workout Focus">
          <div className="flex items-center gap-2 mb-2">
            <Target className="w-5 h-5 text-primary" />
            <span className="font-semibold text-sm tracking-tight">Workout Focus</span>
            {field.state.value && (
              <Button
                type="button"
                variant="ghost"
                size="icon"
                className="h-4 w-4 ml-auto text-muted-foreground hover:text-foreground"
                onClick={(e) => {
                  e.stopPropagation();
                  handleReset();
                }}
              >
                <X className="h-3 w-3" />
              </Button>
            )}
          </div>
          <div className="text-card-foreground font-semibold text-xs">
            {field.state.value || 'What is your focus for today?'}
          </div>
        </Card>
      </DialogTrigger>
      <DialogContent className="w-[90vw] max-w-md sm:max-w-lg mx-auto my-8">
        <DialogHeader>
          <DialogTitle>Workout Focus</DialogTitle>
          <DialogDescription>
            What are you working on today?
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-4">
          <GenericCombobox
            options={workingWorkoutsFocus} // Use working list that can include manually created values
            selected={selectedWorkoutFocus?.name ?? field.state.value ?? ''}
            onChange={handleSelect}
            onCreate={handleAppendGroup}
          />
          <div className="flex gap-2">
            <Button
              className="flex-1 py-4 text-base font-semibold rounded-lg"
              disabled={!selectedWorkoutFocus?.name.trim() || selectedWorkoutFocus?.name === field.state.value}
              onClick={handleAddWorkoutFocus}
            >
              Add today's focus
            </Button>
            {field.state.value && (
              <Button
                type="button"
                variant="outline"
                className="px-4 py-4 text-base font-semibold rounded-lg"
                onClick={handleReset}
              >
                <X className="h-4 w-4" />
              </Button>
            )}
          </div>
        </div>
        <DialogFooter className="sm:justify-start">
          <DialogClose asChild>
            <Button type="button" variant="outline">
              Close
            </Button>
          </DialogClose>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}