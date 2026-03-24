import type { ComponentType, CSSProperties, ReactNode } from 'react';
import { Link, useRouter } from '@tanstack/react-router';
import {
  DndContext,
  KeyboardSensor,
  MouseSensor,
  TouchSensor,
  closestCenter,
  type DragEndEvent,
  useSensor,
  useSensors,
} from '@dnd-kit/core';
import { restrictToVerticalAxis } from '@dnd-kit/modifiers';
import {
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { PencilLine, Plus, Save, Trash2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import type { WorkoutExerciseInput } from '@/client';
import type { WorkoutFocus } from '@/lib/api/workouts';
import { cn } from '@/lib/utils';
import type { ReorderableExercise } from './use-exercise-reorder';

export type WorkoutExerciseCard = Pick<WorkoutExerciseInput, 'name' | 'sets'>;
type SortableHandleAttributes = ReturnType<typeof useSortable>['attributes'];
type SortableHandleListeners = ReturnType<typeof useSortable>['listeners'];
type SortableHandleRef = ReturnType<typeof useSortable>['setActivatorNodeRef'];
type WorkoutFormSectionApi = {
  AppField: ComponentType<any>;
  Subscribe: ComponentType<any>;
};

type WorkoutMetadataFieldsProps = {
  form: WorkoutFormSectionApi;
  workoutsFocus: WorkoutFocus[];
};

type WorkoutExerciseCardsProps = {
  exercises: ReorderableExercise<WorkoutExerciseCard>[];
  dataTestId: string;
  canEditOrder: boolean;
  hasPendingOrderChanges: boolean;
  isReorderMode: boolean;
  onCancelOrder: () => void;
  onEditOrder: () => void;
  onRemoveExercise: (index: number) => void;
  onReorderExercises: (activeId: string, overId: string) => void;
  onSaveOrder: () => void;
  formatVolume: (volume: number) => string;
  renderNameSupplement?: (exercise: WorkoutExerciseCard) => ReactNode;
  renderMetrics?: (exercise: WorkoutExerciseCard) => ReactNode;
};

type WorkoutFormActionsProps = {
  form: WorkoutFormSectionApi;
  isReorderMode: boolean;
};

function getExerciseVolume(exercise: WorkoutExerciseCard): number {
  return exercise.sets.reduce(
    (total, set) => total + (set.reps || 0) * (set.weight || 0),
    0
  );
}

export function WorkoutMetadataFields({
  form,
  workoutsFocus,
}: WorkoutMetadataFieldsProps) {
  return (
    <div className="grid grid-cols-2 gap-4 mb-4">
      <form.AppField
        name="date"
        children={(field: any) => <field.DatePicker />}
      />
      <form.AppField
        name="workoutFocus"
        children={(field: any) => (
          <field.WorkoutFocusCombobox workoutsFocus={workoutsFocus} />
        )}
      />
      <div className="col-span-2">
        <form.AppField
          name="notes"
          children={(field: any) => <field.NotesTextarea />}
        />
      </div>
    </div>
  );
}

export function WorkoutExerciseCards({
  exercises,
  dataTestId,
  canEditOrder,
  hasPendingOrderChanges,
  isReorderMode,
  onCancelOrder,
  onEditOrder,
  onRemoveExercise,
  onReorderExercises,
  onSaveOrder,
  formatVolume,
  renderNameSupplement,
  renderMetrics,
}: WorkoutExerciseCardsProps) {
  const router = useRouter();
  const sensors = useSensors(
    useSensor(MouseSensor, {
      activationConstraint: { distance: 4 },
    }),
    useSensor(TouchSensor, {
      activationConstraint: { delay: 120, tolerance: 8 },
    }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  );

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;

    if (!over || active.id === over.id) {
      return;
    }

    onReorderExercises(String(active.id), String(over.id));
  };

  const exerciseList = (
    <div className="space-y-3">
      {exercises.map(({ exercise, id }, exerciseIndex) => {
        if (isReorderMode) {
          return (
            <SortableWorkoutExerciseCard
              key={id}
              dataTestId={dataTestId}
              exercise={exercise}
              exerciseIndex={exerciseIndex}
              formatVolume={formatVolume}
              id={id}
              renderMetrics={renderMetrics}
              renderNameSupplement={renderNameSupplement}
            />
          );
        }

        return (
          <WorkoutExerciseCardContent
            key={id}
            dataTestId={dataTestId}
            exercise={exercise}
            exerciseIndex={exerciseIndex}
            formatVolume={formatVolume}
            onOpenExercise={() => {
              router.navigate({ to: '.', search: { exerciseIndex } });
            }}
            onRemoveExercise={onRemoveExercise}
            renderMetrics={renderMetrics}
            renderNameSupplement={renderNameSupplement}
          />
        );
      })}
    </div>
  );

  return (
    <section className="space-y-3">
      <div className="flex items-start justify-between gap-3">
        <div className="space-y-1">
          <p className="text-sm font-semibold text-card-foreground">Exercises</p>
          <p className="text-sm text-muted-foreground">
            {isReorderMode
              ? 'Drag the handle to lock in the exercise order, then save it.'
              : 'Tap an exercise card to manage its sets and details.'}
          </p>
        </div>
        {isReorderMode ? (
          <div className="flex items-center gap-2">
            <Button type="button" variant="outline" size="sm" onClick={onCancelOrder}>
              Cancel
            </Button>
            <Button
              type="button"
              size="sm"
              onClick={onSaveOrder}
              disabled={!hasPendingOrderChanges}
              data-testid="save-exercise-order"
            >
              Save
            </Button>
          </div>
        ) : (
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={onEditOrder}
            disabled={!canEditOrder}
            data-testid="edit-exercise-order"
          >
            <PencilLine className="mr-1.5 h-4 w-4" />
            Edit order
          </Button>
        )}
      </div>

      {isReorderMode ? (
        <DndContext
          collisionDetection={closestCenter}
          modifiers={[restrictToVerticalAxis]}
          onDragEnd={handleDragEnd}
          sensors={sensors}
        >
          <SortableContext
            items={exercises.map(({ id }) => id)}
            strategy={verticalListSortingStrategy}
          >
            {exerciseList}
          </SortableContext>
        </DndContext>
      ) : (
        exerciseList
      )}
    </section>
  );
}

function WorkoutExerciseCardContent({
  dataTestId,
  exercise,
  exerciseIndex,
  formatVolume,
  isDragging = false,
  isReorderMode = false,
  onOpenExercise,
  onRemoveExercise,
  renderMetrics,
  renderNameSupplement,
  sortableAttributes,
  sortableListeners,
  sortableHandleRef,
}: {
  dataTestId: string;
  exercise: WorkoutExerciseCard;
  exerciseIndex: number;
  formatVolume: (volume: number) => string;
  isDragging?: boolean;
  isReorderMode?: boolean;
  onOpenExercise?: () => void;
  onRemoveExercise?: (index: number) => void;
  renderNameSupplement?: (exercise: WorkoutExerciseCard) => ReactNode;
  renderMetrics?: (exercise: WorkoutExerciseCard) => ReactNode;
  sortableAttributes?: SortableHandleAttributes;
  sortableListeners?: SortableHandleListeners;
  sortableHandleRef?: SortableHandleRef;
}) {
  const volume = getExerciseVolume(exercise);

  return (
    <Card
      className={cn(
        'p-4 transition-all duration-200',
        isReorderMode
          ? 'border-primary/30 bg-primary/5 shadow-sm workout-card-wiggle'
          : 'hover:shadow-md',
        isDragging && 'opacity-80 shadow-lg ring-1 ring-primary/30'
      )}
      style={
        isReorderMode
          ? ({
              '--wiggle-index': exerciseIndex,
            } as CSSProperties)
          : undefined
      }
    >
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0 flex-1">
          <div className="flex items-center justify-between gap-3">
            {isReorderMode ? (
              <div className="mb-2 flex min-w-0 items-center gap-2">
                <div className="h-2 w-2 rounded-full bg-primary"></div>
                <div className="min-w-0">
                  <span className="text-sm font-medium text-primary">{exercise.name}</span>
                  {renderNameSupplement?.(exercise)}
                </div>
              </div>
            ) : (
              <button
                type="button"
                aria-label={`Edit ${exercise.name}`}
                data-testid={dataTestId}
                className="mb-2 flex min-w-0 flex-1 items-center gap-2 text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                onClick={onOpenExercise}
              >
                <div className="h-2 w-2 rounded-full bg-primary"></div>
                <div className="min-w-0">
                  <span className="text-sm font-medium text-primary">{exercise.name}</span>
                  {renderNameSupplement?.(exercise)}
                </div>
              </button>
            )}
            {isReorderMode ? (
              <button
                type="button"
                ref={sortableHandleRef}
                className="flex h-10 w-10 shrink-0 cursor-grab items-center justify-center rounded-full border border-border/70 bg-background/90 text-muted-foreground transition hover:border-primary/40 hover:text-primary active:cursor-grabbing"
                aria-label={`Reorder ${exercise.name}`}
                {...sortableAttributes}
                {...sortableListeners}
              >
                <span className="flex flex-col gap-1" aria-hidden="true">
                  <span className="h-0.5 w-4 rounded-full bg-current"></span>
                  <span className="h-0.5 w-4 rounded-full bg-current"></span>
                  <span className="h-0.5 w-4 rounded-full bg-current"></span>
                </span>
              </button>
            ) : (
              <Button
                type="button"
                variant="ghost"
                size="icon"
                className="h-8 w-8 text-primary hover:bg-primary/10 hover:text-primary/80"
                aria-label={`Delete ${exercise.name}`}
                onClick={(event) => {
                  event.stopPropagation();
                  event.preventDefault();
                  onRemoveExercise?.(exerciseIndex);
                }}
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            )}
          </div>

          <div className="flex items-end justify-between gap-3">
            {isReorderMode ? (
              <>
                <div>
                  <div className="text-lg font-bold text-card-foreground">{exercise.sets.length}</div>
                  <div className="text-sm font-semibold uppercase tracking-tight text-muted-foreground">
                    sets
                  </div>
                </div>

                <div className="flex items-end gap-4">
                  <div className="text-right">
                    <div className="text-lg font-bold text-card-foreground">
                      {formatVolume(volume)}
                    </div>
                    <div className="text-sm font-semibold uppercase tracking-tight text-muted-foreground">
                      volume
                    </div>
                  </div>
                  {renderMetrics?.(exercise)}
                </div>
              </>
            ) : (
              <button
                type="button"
                aria-label={`${exercise.name}: ${exercise.sets.length} sets, ${formatVolume(volume)} volume`}
                className="flex w-full items-end justify-between gap-3 text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                onClick={onOpenExercise}
              >
                <div>
                  <div className="text-lg font-bold text-card-foreground">{exercise.sets.length}</div>
                  <div className="text-sm font-semibold uppercase tracking-tight text-muted-foreground">
                    sets
                  </div>
                </div>

                <div className="flex items-end gap-4">
                  <div className="text-right">
                    <div className="text-lg font-bold text-card-foreground">
                      {formatVolume(volume)}
                    </div>
                    <div className="text-sm font-semibold uppercase tracking-tight text-muted-foreground">
                      volume
                    </div>
                  </div>
                  {renderMetrics?.(exercise)}
                </div>
              </button>
            )}
          </div>
        </div>
      </div>
    </Card>
  );
}

function SortableWorkoutExerciseCard({
  dataTestId,
  exercise,
  exerciseIndex,
  formatVolume,
  id,
  renderMetrics,
  renderNameSupplement,
}: {
  dataTestId: string;
  exercise: WorkoutExerciseCard;
  exerciseIndex: number;
  formatVolume: (volume: number) => string;
  id: string;
  renderNameSupplement?: (exercise: WorkoutExerciseCard) => ReactNode;
  renderMetrics?: (exercise: WorkoutExerciseCard) => ReactNode;
}) {
  const {
    attributes,
    isDragging,
    listeners,
    setActivatorNodeRef,
    setNodeRef,
    transform,
    transition,
  } = useSortable({ id });

  return (
    <div
      ref={setNodeRef}
      style={{
        transform: CSS.Transform.toString(transform),
        transition,
      }}
    >
      <WorkoutExerciseCardContent
        dataTestId={dataTestId}
        exercise={exercise}
        exerciseIndex={exerciseIndex}
        formatVolume={formatVolume}
        isDragging={isDragging}
        isReorderMode
        renderMetrics={renderMetrics}
        renderNameSupplement={renderNameSupplement}
        sortableAttributes={attributes}
        sortableHandleRef={setActivatorNodeRef}
        sortableListeners={listeners}
      />
    </div>
  );
}

export function WorkoutFormActions({ form, isReorderMode }: WorkoutFormActionsProps) {
  return (
    <>
      <div className="py-6">
        {isReorderMode ? (
          <Button
            type="button"
            variant="outline"
            className="w-full rounded-lg text-base font-semibold"
            disabled
          >
            <Plus className="mr-2 h-5 w-5" />
            Save order before adding exercises
          </Button>
        ) : (
          <Button
            type="button"
            variant="outline"
            className="w-full rounded-lg text-base font-semibold"
            asChild
          >
            <Link
              to="."
              search={{ addExercise: true }}
              data-testid="add-exercise"
            >
              <Plus className="mr-2 h-5 w-5" />
              Add Exercise
            </Link>
          </Button>
        )}
      </div>
      <div className="mt-8">
        <form.Subscribe
          selector={(state: any) => [state.canSubmit, state.isSubmitting]}
          children={([canSubmit, isSubmitting]: [boolean, boolean]) => (
            <Button
              type="submit"
              disabled={!canSubmit || isReorderMode}
              className="w-full rounded-lg text-base font-semibold"
              data-testid="save-workout"
            >
              <Save className="mr-1.5 h-3.5 w-3.5" />
              {isSubmitting ? 'Saving...' : 'Save'}
            </Button>
          )}
        />
      </div>
    </>
  );
}
