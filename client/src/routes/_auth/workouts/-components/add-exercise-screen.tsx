import { withForm } from '@/hooks/form';
import { Plus } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import type { DbExercise, ExerciseOption } from '@/lib/api/exercises';
import { MOCK_VALUES } from '../-components/form-options';
import { useState } from 'react';
import { Input } from '@/components/ui/input';
import { ChevronLeft, Search } from 'lucide-react';
import { CardContent } from '@/components/ui/card';
import { ChevronRight } from 'lucide-react';

type AddExerciseScreenProps = {
  exercises: DbExercise[]; // Database exercises with guaranteed IDs
  onAddExercise: (exerciseIndex: number, exerciseId?: number) => void;
  onBack: () => void;
};

export const AddExerciseScreen = withForm({
  defaultValues: MOCK_VALUES,
  props: {} as AddExerciseScreenProps,
  render: function Render({ form, exercises, onAddExercise, onBack }) {
    const [searchQuery, setSearchQuery] = useState('');
    const [workingExercises, setWorkingExercises] = useState<ExerciseOption[]>(
      exercises.map((ex) => ({ id: ex.id, name: ex.name }))
    );

    const filteredExercises = workingExercises.filter((exercise) =>
      exercise.name.toLowerCase().includes(searchQuery.toLowerCase())
    );

    return (
      <main>
        <div className="max-w-lg mx-auto space-y-6 px-4 pb-8">
          {/* Header */}
          <div className="flex items-center justify-between pt-4">
            <button onClick={onBack}>
              <ChevronLeft className="text-primary" />
            </button>
            <div>
              <h1 className="text-2xl font-bold tracking-tight">
                Choose Exercise
              </h1>
            </div>
            <form.AppField
              name="exercises"
              mode="array"
              children={(field) => (
                <>
                  {filteredExercises.length === 0 && (
                    <Button
                      size="sm"
                      onClick={() => {
                        console.log('searchQuery', searchQuery);
                        const newExercise: ExerciseOption = {
                          id: null, // null ID for new exercises not yet in the database
                          name: searchQuery,
                        };
                        setWorkingExercises((prev) => [...prev, newExercise]);
                        field.pushValue({
                          name: searchQuery,
                          sets: [],
                        });
                        const exerciseIndex = field.state.value.length - 1;
                        onAddExercise(exerciseIndex, undefined);
                      }}
                    >
                      <Plus className="w-4 h-4 mr-2" />
                      Add
                    </Button>
                  )}
                </>
              )}
            />
          </div>

          {/* MARK: Search */}
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground w-4 h-4" />
            <Input
              placeholder="Search exercises..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10"
            />
          </div>

          {/* MARK: Exercise List */}
          <Card className="py-0">
            <CardContent className="p-0">
              <form.AppField
                name="exercises"
                mode="array"
                children={(field) => (
                  <>
                    {filteredExercises.length === 0 ? (
                      <div>
                        <div className="text-center py-8 text-muted-foreground">
                          No exercises found matching "{searchQuery}"
                        </div>
                      </div>
                    ) : (
                      filteredExercises.map((exercise) => (
                        // MARK: List items
                        <div
                          key={exercise.id}
                          className="flex items-center justify-between p-4 hover:bg-gray-100/50 transition-colors cursor-pointer border-b border-border last:border-b-0"
                          onClick={() => {
                            field.pushValue({
                              name: exercise.name,
                              sets: [],
                            });
                            const exerciseIndex = field.state.value.length - 1;
                            // should not be undefined but check for type safety
                            const exerciseId = exercise.id || undefined;
                            onAddExercise(exerciseIndex, exerciseId);
                          }}
                        >
                          <h3 className="font-semibold">{exercise.name}</h3>
                          <ChevronRight className="w-5 h-5 text-muted-foreground" />
                        </div>
                      ))
                    )}
                  </>
                )}
              />
            </CardContent>
          </Card>

          {/* MARK: Results Count */}
          {searchQuery && (
            <div className="text-center text-sm text-muted-foreground">
              {filteredExercises.length} exercise
              {filteredExercises.length !== 1 ? 's' : ''} found
            </div>
          )}
        </div>
      </main>
    );
  },
});

// type AddExerciseScreenProps = {
//   exercises: DbExercise[]; // Database exercises with guaranteed IDs
//   onAddExercise: (exerciseIndex: number, exerciseId?: number) => void;
//   onBack: () => void;
// };

// export const AddExerciseScreen = withForm({
//   defaultValues: MOCK_VALUES,
//   props: {} as AddExerciseScreenProps,
//   render: function Render({ form, exercises, onAddExercise, onBack }) {
//     return (
//       <div className="min-h-screen bg-background">
//         <div className="px-4 pb-8">
//           <div className="max-w-md mx-auto space-y-6">
//             {/* Header */}
//             <div className="flex items-center pt-6 pb-2">
//               <Button
//                 variant="ghost"
//                 onClick={onBack}
//                 className="p-0 h-auto text-primary hover:text-primary/80"
//               >
//                 <ArrowLeft className="w-6 h-6" />
//               </Button>
//               <h1 className="font-bold text-3xl tracking-tight text-foreground flex-1 text-center">
//                 Add Exercise
//               </h1>
//             </div>

//             {/* Exercise Name Input */}
//             <form.AppField
//               name="exercises"
//               mode="array"
//               children={(field) => (
//                 <>
//                   <field.AddExerciseField2
//                     exercises={exercises}
//                     onAddExercise={onAddExercise}
//                   />
//                   <div className="space-y-4">
//                     <h3 className="font-semibold text-sm tracking-tight uppercase text-muted-foreground text-center">
//                       OR CHOOSE FROM COMMON EXERCISES:
//                     </h3>
//                     <div className="grid grid-cols-2 gap-3">
//                       {exercises
//                         .filter(
//                           (exercise) =>
//                             !field.state.value.some(
//                               (e) => e.name === exercise.name
//                             )
//                         )
//                         .map((exercise) => (
//                           <Card
//                             key={exercise.id} // ! TODO: handle button hover
//                             className="shadow-sm hover:bg-accent hover:text-accent-foreground dark:hover:bg-accent/50"
//                           >
//                             <Button
//                               variant="ghost"
//                               className="h-auto w-full text-sm font-medium whitespace-normal text-card-foreground hover:bg-transparent"
//                               onClick={() => {
//                                 field.pushValue({
//                                   name: exercise.name,
//                                   sets: [],
//                                 });
//                                 const exerciseIndex =
//                                   field.state.value.length - 1;
//                                 onAddExercise(exerciseIndex, exercise.id);
//                               }}
//                             >
//                               {exercise.name}
//                             </Button>
//                           </Card>
//                         ))}
//                     </div>
//                   </div>
//                 </>
//               )}
//             />
//           </div>
//         </div>
//       </div>
//     );
//   },
// });
