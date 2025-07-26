import { Button } from '@/components/ui/button';
import type { Exercise, ExerciseOption } from '@/lib/types';
import { Plus } from 'lucide-react';
import { ExerciseCombobox } from '@/components/exercise-combobox';
import { useFieldContext } from '@/hooks/form';
import { useState } from 'react';

export default function AddExerciseField2({
  exercises,
  onAddExercise,
}: {
  exercises: ExerciseOption[];
  onAddExercise: (exerciseIndex: number) => void;
}) {
  const field = useFieldContext<Exercise[]>();
  const [selectedExercise, setSelectedExercise] = useState<ExerciseOption>();

  function handleSelect(option: ExerciseOption) {
    setSelectedExercise(option);
  }

  // ! MARK: TODO handle id
  function handleAppendGroup(name: ExerciseOption['name']) {
    const newExercise: ExerciseOption = {
      id: exercises.length + 1,
      name,
      created_at: new Date().toISOString(),
      updated_at: null,
    };
    exercises.push(newExercise);
    handleSelect(newExercise);
  }

  return (
    <div className="space-y-4">
      <ExerciseCombobox
        options={exercises}
        selected={selectedExercise?.name ?? ''}
        onChange={handleSelect}
        onCreate={handleAppendGroup}
      />
      <Button
        className="bg-primary text-primary-foreground hover:bg-primary/90 w-full py-4 text-base font-semibold rounded-lg"
        disabled={!selectedExercise?.name.trim()}
        onClick={() => {
          field.pushValue({
            name: selectedExercise?.name ?? '',
            sets: [],
          });
          onAddExercise(field.state.value.length - 1);
          setSelectedExercise(undefined);
        }}
      >
        <Plus className="w-4 h-4 mr-2" />
        Add Exercise
      </Button>
    </div>
  );
}
