import { Button } from '@/components/ui/button';
import type { DbExercise, ExerciseOption } from '@/lib/api/exercises';
import { Plus } from 'lucide-react';
import { ExerciseCombobox } from '@/components/exercise-combobox';
import { useFieldContext } from '@/hooks/form';
import { useState } from 'react';
import type { workout_ExerciseInput } from '@/generated';

export default function AddExerciseField2({
  exercises,
  onAddExercise,
}: {
  exercises: DbExercise[]; // Input: exercises from the database with guaranteed IDs
  onAddExercise: (exerciseIndex: number, exerciseId?: number) => void;
}) {
  const field = useFieldContext<workout_ExerciseInput[]>();
  const [selectedExercise, setSelectedExercise] = useState<ExerciseOption>(); // State: may include manually created exercises
  
  // Working list of exercises that can include both DB and manually created ones
  const [workingExercises, setWorkingExercises] = useState<ExerciseOption[]>(
    exercises.map(ex => ({ id: ex.id, name: ex.name }))
  );

  function handleSelect(option: ExerciseOption) {
    setSelectedExercise(option);
  }

  function handleAppendGroup(name: string) {
    // For new exercises, use null ID to indicate they're not in the database
    const newExercise: ExerciseOption = {
      id: null, // null ID for new exercises not yet in the database
      name,
    };
    setWorkingExercises(prev => [...prev, newExercise]);
    handleSelect(newExercise);
  }

  return (
    <div className="space-y-4">
      <ExerciseCombobox
        options={workingExercises} // Use working list that can include manually created exercises
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
          const exerciseIndex = field.state.value.length - 1;
          // Only pass exerciseId if it's not null (real DB ID)
          const exerciseId = selectedExercise?.id || undefined;
          onAddExercise(exerciseIndex, exerciseId);
          setSelectedExercise(undefined);
        }}
      >
        <Plus className="w-4 h-4 mr-2" />
        Add Exercise
      </Button>
    </div>
  );
}
