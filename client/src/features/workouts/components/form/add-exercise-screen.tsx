import type { WorkoutExerciseInput } from "@/client";
import { ExerciseCatalogPicker } from "@/features/exercises/components/exercise-catalog-picker";
import { withForm } from "@/hooks/form";
import { Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  ExerciseList,
  ExerciseListSearch,
} from "@/features/exercises/components/exercise-selection-list";
import type {
  DbExercise,
  ExerciseOption,
} from "@/features/exercises/api/exercises";
import { MOCK_VALUES } from "@/features/workouts/components/form/form-options";
import { useState } from "react";
import { ChevronLeft } from "lucide-react";

type AddExerciseScreenProps = {
  isDemoMode?: boolean;
  exercises: DbExercise[]; // Database exercises with guaranteed IDs
  onBack: () => void;
  onAddExercise: (index: number, isNewExercise?: boolean) => void;
};

function normalizeExerciseName(name: string) {
  return name.trim().toLowerCase();
}

export const AddExerciseScreen = withForm({
  defaultValues: MOCK_VALUES,
  // SAFETY: TanStack withForm spreads this empty defaults object before the props
  // supplied by every AddExerciseScreen caller, so it is only a type witness for the generic.
  props: {} as AddExerciseScreenProps,
  render: function Render({
    form,
    exercises,
    onBack,
    onAddExercise,
    isDemoMode = false,
  }) {
    const [showCatalog, setShowCatalog] = useState(false);
    const [searchQuery, setSearchQuery] = useState("");
    const [workingExercises, setWorkingExercises] = useState<ExerciseOption[]>(
      exercises.map((ex) => ({ id: ex.id, name: ex.name })),
    ); // ExerciseOption type allows null ids for new exercises not yet in the database

    const filteredExercises = workingExercises.filter((exercise) =>
      exercise.name.toLowerCase().includes(searchQuery.toLowerCase()),
    );
    const trimmedSearchQuery = searchQuery.trim();
    const canCreateExercise =
      normalizeExerciseName(trimmedSearchQuery).length > 0 &&
      !workingExercises.some(
        (exercise) =>
          normalizeExerciseName(exercise.name) ===
          normalizeExerciseName(trimmedSearchQuery),
      );

    return (
      <main>
        <div className="max-w-lg mx-auto space-y-6 px-4 pb-8">
          {/* Header */}
          <div className="flex items-center justify-between pt-4">
            <button
              type="button"
              aria-label={showCatalog ? "Back to exercises" : "Back to workout"}
              onClick={() => (showCatalog ? setShowCatalog(false) : onBack())}
              className="cursor-pointer"
            >
              <ChevronLeft className="text-primary" />
            </button>
            <div>
              <h1 className="text-2xl font-bold tracking-tight">
                {showCatalog ? "Exercise Catalog" : "Choose Exercise"}
              </h1>
            </div>
            <form.AppField
              name="exercises"
              mode="array"
              children={(field) => (
                <>
                  {canCreateExercise && !showCatalog && (
                    <Button
                      size="sm"
                      onClick={() => {
                        const newExercise: ExerciseOption = {
                          id: null, // null ID for new exercises not yet in the database
                          name: trimmedSearchQuery,
                        };
                        setWorkingExercises((prev) => [...prev, newExercise]);
                        field.pushValue({
                          name: trimmedSearchQuery,
                          sets: [],
                        });
                        const exerciseIndex = field.state.value.length - 1;
                        onAddExercise(exerciseIndex, true);
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

          {!isDemoMode && (
            <div className="space-y-3">
              {!showCatalog && (
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setShowCatalog(true)}
                >
                  Browse catalog
                </Button>
              )}
              {showCatalog && (
                <form.AppField
                  name="exercises"
                  mode="array"
                  children={(field) => (
                    <ExerciseCatalogPicker
                      onSelect={(entry) => {
                        const existing = workingExercises.find(
                          (exercise) =>
                            normalizeExerciseName(exercise.name) ===
                            normalizeExerciseName(entry.name),
                        );
                        const draft: WorkoutExerciseInput = {
                          name: existing?.name ?? entry.name,
                          sets: [],
                        };
                        if (!existing) draft.catalog_id = entry.id;
                        field.pushValue(draft);
                        onAddExercise(field.state.value.length - 1, true);
                      }}
                    />
                  )}
                />
              )}
            </div>
          )}

          {!showCatalog && (
            <>
              <ExerciseListSearch
                label="Search exercises"
                placeholder="Search/add exercises"
                value={searchQuery}
                onChange={setSearchQuery}
              />
              <form.AppField
                name="exercises"
                mode="array"
                children={(field) => (
                  <ExerciseList
                    entries={filteredExercises}
                    emptyMessage={`No exercises found matching "${searchQuery}"`}
                    showCount={searchQuery.length > 0}
                    onSelect={(exercise) => {
                      field.pushValue({ name: exercise.name, sets: [] });
                      onAddExercise(field.state.value.length - 1, true);
                    }}
                  />
                )}
              />
            </>
          )}
        </div>
      </main>
    );
  },
});
