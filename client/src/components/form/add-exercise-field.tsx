import type { Exercise,ExerciseOption } from '@/lib/types';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { Plus } from 'lucide-react';
import { ExerciseCombobox } from '@/components/exercise-combobox';
import { useFieldContext } from '@/hooks/form';

export default function AddExerciseField({
  showTitle = true,
  onExerciseAdded,
  exercises,
  selectedExercise,
  handleSelect,
  handleAppendGroup,
}: {
  showTitle?: boolean;
  onExerciseAdded?: () => void;
  exercises: ExerciseOption[];
  selectedExercise?: ExerciseOption;
  handleSelect: (option: ExerciseOption) => void;
  handleAppendGroup: (label: ExerciseOption['name']) => void;
}) {
  const field = useFieldContext<Exercise>();
  return (
    <div className="flex flex-col gap-3">
      <div className="space-y-2">
        {showTitle && (
          <Label className="text-xs text-neutral-400 tracking-wider">
            EXERCISE DATABASE
          </Label>
        )}
        <div className="flex flex-col sm:flex-row gap-3">
          <div className="flex-1">
            <ExerciseCombobox
              options={exercises}
              selected={selectedExercise?.name ?? ''}
              onChange={handleSelect}
              onCreate={handleAppendGroup}
            />
          </div>
          <Button
            onClick={() => {
              field.pushValue({
                name: selectedExercise?.name ?? '',
                sets: [],
              });
              onExerciseAdded?.();
            }}
            type="button"
            className="w-full sm:w-auto bg-orange-500 hover:bg-orange-600 text-white"
            disabled={!selectedExercise?.name}
          >
            <Plus className="w-4 h-4 mr-2" />
            Add Exercise
          </Button>
        </div>
      </div>
    </div>
  );
}
