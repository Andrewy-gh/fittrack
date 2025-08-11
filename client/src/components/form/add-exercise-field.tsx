// import { Label } from '@/components/ui/label';
// import { Button } from '@/components/ui/button';
// import { Plus } from 'lucide-react';
// import { ExerciseCombobox } from '@/components/exercise-combobox';
// import { useFieldContext } from '@/hooks/form';
// import { useState } from 'react';

export default function AddExerciseField() {
  return <div>Ready to delete</div>
//   showTitle = true,
//   onExerciseAdded,
//   exercises,
// }: {
//   showTitle?: boolean;
//   onExerciseAdded?: () => void;
//   exercises: ExerciseOption[];
// }) {
//   const field = useFieldContext<Exercise[]>();
//   const [selectedExercise, setSelectedExercise] = useState<ExerciseOption>();
  
//   function handleSelect(option: ExerciseOption) {
//     console.log('handleSelect');
//     console.log(option);
//     setSelectedExercise(option);
//   }

//   // ! MARK: TODO
//   function handleAppendGroup(name: NewExerciseOption['name']) {
//     const newExercise: ExerciseOption = {
//       id: exercises.length + 1,
//       name,
//       created_at: new Date().toISOString(),
//       updated_at: null,
//     };
//     exercises.push(newExercise);
//     console.log('handleAppendGroup');
//     console.log(newExercise);
//     handleSelect(newExercise);
//   }

//   return (
//     <div className="flex flex-col gap-3">
//       <div className="space-y-2">
//         {showTitle && (
//           <Label className="text-xs text-neutral-400 tracking-wider">
//             EXERCISE DATABASE
//           </Label>
//         )}
//         <div className="flex flex-col sm:flex-row gap-3">
//           <div className="flex-1">
//             <ExerciseCombobox
//               options={exercises}
//               selected={selectedExercise?.name ?? ''}
//               onChange={handleSelect}
//               onCreate={handleAppendGroup}
//             />
//           </div>
//           <Button
//             onClick={() => {
//               field.pushValue({
//                 name: selectedExercise?.name ?? '',
//                 sets: [],
//               });
//               onExerciseAdded?.();
//             }}
//             type="button"
//             className="w-full sm:w-auto bg-orange-500 hover:bg-orange-600 text-white"
//             disabled={!selectedExercise?.name}
//           >
//             <Plus className="w-4 h-4 mr-2" />
//             Add Exercise
//           </Button>
//         </div>
//       </div>
//     </div>
//   );
}
