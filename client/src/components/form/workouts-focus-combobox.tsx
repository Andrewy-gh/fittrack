import { Button } from '@/components/ui/button';
import { Plus } from 'lucide-react';
import { GenericCombobox } from '@/components/generic-combobox';
import { useFieldContext } from '@/hooks/form';
import { useState } from 'react';

export default function WorkoutsFocusCombobox({
  workoutsFocus,
}: {
  workoutsFocus: Array<{ name: string }>; // Input: workout focus values from the database
}) {
  const field = useFieldContext<string>();
  const [selectedWorkoutFocus, setSelectedWorkoutFocus] = useState<{ name: string }>();

  // Working list of workout focus values that can include both DB and manually created ones
  const [workingWorkoutsFocus, setWorkingWorkoutsFocus] = useState<Array<{ name: string }>>(
    workoutsFocus
  );

  function handleSelect(option: { name: string }) {
    setSelectedWorkoutFocus(option);
  }

  function handleAppendGroup(name: string) {
    // For new workout focus values, just add them to the working list
    const newWorkoutFocus = { name };
    setWorkingWorkoutsFocus((prev) => [...prev, newWorkoutFocus]);
    handleSelect(newWorkoutFocus);
  }

  return (
    <div className="space-y-4">
      <GenericCombobox
        options={workingWorkoutsFocus} // Use working list that can include manually created values
        selected={selectedWorkoutFocus?.name ?? field.state.value ?? ''}
        onChange={handleSelect}
        onCreate={handleAppendGroup}
      />
      <Button
        className="bg-primary text-primary-foreground hover:bg-primary/90 w-full py-4 text-base font-semibold rounded-lg"
        disabled={!selectedWorkoutFocus?.name.trim()}
        onClick={() => {
          field.setValue(selectedWorkoutFocus?.name ?? '');
          setSelectedWorkoutFocus(undefined);
        }}
      >
        <Plus className="w-4 h-4 mr-2" />
        Add workout focus value
      </Button>
    </div>
  );
}