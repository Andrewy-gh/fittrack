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
import { useNavigate } from '@tanstack/react-router';
import { Route } from '../new';

type AddExerciseScreenProps = {
  exercises: DbExercise[]; // Database exercises with guaranteed IDs
  onBack: () => void;
  onAddExercise?: (index: number) => void; // Optional callback for when exercise is added (used by edit.tsx)
};

export const AddExerciseScreen = withForm({
  defaultValues: MOCK_VALUES,
  props: {} as AddExerciseScreenProps,
  render: function Render({ form, exercises, onBack, onAddExercise }) {
    const navigate = useNavigate({ from: Route.fullPath });
    const [searchQuery, setSearchQuery] = useState('');
    const [workingExercises, setWorkingExercises] = useState<ExerciseOption[]>(
      exercises.map((ex) => ({ id: ex.id, name: ex.name }))
    ); // ExerciseOption type allows null ids for new exercises not yet in the database

    const filteredExercises = workingExercises.filter((exercise) =>
      exercise.name.toLowerCase().includes(searchQuery.toLowerCase())
    );

    return (
      <main>
        <div className="max-w-lg mx-auto space-y-6 px-4 pb-8">
          {/* Header */}
          <div className="flex items-center justify-between pt-4">
            <button onClick={onBack} className="cursor-pointer">
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
                        if (onAddExercise) {
                          onAddExercise(exerciseIndex);
                        } else {
                          navigate({ search: { exerciseIndex } });
                        }
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
              placeholder="Search/add exercises"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10 text-base"
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
                          data-testid="exercise-card"
                          className="flex items-center justify-between p-4 hover:bg-gray-100/50 transition-colors cursor-pointer border-b border-border last:border-b-0"
                          onClick={() => {
                            field.pushValue({
                              name: exercise.name,
                              sets: [],
                            });
                            const exerciseIndex = field.state.value.length - 1;
                            if (onAddExercise) {
                              onAddExercise(exerciseIndex);
                            } else {
                              navigate({ search: { exerciseIndex } });
                            }
                          }}
                        >
                          <h3 className="font-semibold md:text-sm">{exercise.name}</h3>
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

