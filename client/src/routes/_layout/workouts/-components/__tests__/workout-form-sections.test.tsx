import type {
  ButtonHTMLAttributes,
  HTMLAttributes,
  ReactNode,
} from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const {
  mockUseSensor,
  mockUseSortable,
  mockSortableKeyboardCoordinates,
  keyboardSensorToken,
  mouseSensorToken,
  touchSensorToken,
} = vi.hoisted(() => ({
  mockUseSensor: vi.fn(),
  mockUseSortable: vi.fn(),
  mockSortableKeyboardCoordinates: vi.fn(),
  keyboardSensorToken: { name: 'KeyboardSensor' },
  mouseSensorToken: { name: 'MouseSensor' },
  touchSensorToken: { name: 'TouchSensor' },
}));

vi.mock('@tanstack/react-router', () => ({
  useRouter: () => ({
    navigate: vi.fn(),
  }),
}));

vi.mock('@dnd-kit/core', () => ({
  DndContext: ({ children }: { children: ReactNode }) => (
    <div data-testid="dnd-context">{children}</div>
  ),
  KeyboardSensor: keyboardSensorToken,
  MouseSensor: mouseSensorToken,
  TouchSensor: touchSensorToken,
  closestCenter: vi.fn(),
  useSensor: mockUseSensor,
  useSensors: (...sensors: unknown[]) => sensors,
}));

vi.mock('@dnd-kit/modifiers', () => ({
  restrictToVerticalAxis: {},
}));

vi.mock('@dnd-kit/sortable', () => ({
  SortableContext: ({ children }: { children: ReactNode }) => (
    <div data-testid="sortable-context">{children}</div>
  ),
  arrayMove: <T,>(items: T[], oldIndex: number, newIndex: number) => {
    const nextItems = [...items];
    const [movedItem] = nextItems.splice(oldIndex, 1);
    nextItems.splice(newIndex, 0, movedItem);
    return nextItems;
  },
  sortableKeyboardCoordinates: mockSortableKeyboardCoordinates,
  useSortable: mockUseSortable,
  verticalListSortingStrategy: {},
}));

vi.mock('@dnd-kit/utilities', () => ({
  CSS: {
    Transform: {
      toString: () => undefined,
    },
  },
}));

vi.mock('@/components/ui/button', () => ({
  Button: ({
    children,
    ...props
  }: ButtonHTMLAttributes<HTMLButtonElement>) => (
    <button {...props}>{children}</button>
  ),
}));

vi.mock('@/components/ui/card', () => ({
  Card: ({
    children,
    className,
    ...props
  }: HTMLAttributes<HTMLDivElement>) => (
    <div className={className} {...props}>
      {children}
    </div>
  ),
}));

import { WorkoutExerciseCards, type WorkoutExerciseCard } from '../workout-form-sections';
import { useExerciseReorder } from '../use-exercise-reorder';
import { useAppForm } from '@/hooks/form';

function createExercise(name: string): WorkoutExerciseCard {
  return {
    name,
    sets: [],
  };
}

function WorkoutExerciseCardsHarness() {
  const form = useAppForm({
    defaultValues: {
      date: '',
      notes: '',
      exercises: [createExercise('Bench Press'), createExercise('Barbell Squat')],
    },
    onSubmit: async () => undefined,
  });

  return (
    <form.AppField
      name="exercises"
      mode="array"
      children={(field) => <WorkoutExerciseCardsField field={field} />}
    />
  );
}

function WorkoutExerciseCardsField({ field }: { field: any }) {
  const exerciseReorder = useExerciseReorder(
    field.state.value as WorkoutExerciseCard[]
  );

  return (
    <>
      <button
        type="button"
        data-testid="force-reorder"
        onClick={() => {
          const [firstEntry, secondEntry] = exerciseReorder.displayEntries;

          if (!firstEntry || !secondEntry) {
            return;
          }

          exerciseReorder.moveExercise(secondEntry.id, firstEntry.id);
        }}
      >
        Force reorder
      </button>

      <WorkoutExerciseCards
        exercises={exerciseReorder.displayEntries}
        dataTestId="exercise-card"
        canEditOrder={exerciseReorder.canReorder}
        hasPendingOrderChanges={exerciseReorder.hasPendingOrderChanges}
        isReorderMode={exerciseReorder.isReorderMode}
        onCancelOrder={exerciseReorder.cancelReorder}
        onEditOrder={exerciseReorder.startReorder}
        onRemoveExercise={field.removeValue}
        onReorderExercises={exerciseReorder.moveExercise}
        onSaveOrder={() => {
          field.handleChange(exerciseReorder.commitReorder());
        }}
        formatVolume={() => '0 lb'}
      />
    </>
  );
}

describe('WorkoutExerciseCards', () => {
  beforeEach(() => {
    mockUseSensor.mockImplementation((sensor, options) => ({ sensor, options }));
    mockUseSortable.mockReturnValue({
      attributes: {},
      isDragging: false,
      listeners: {},
      setActivatorNodeRef: vi.fn(),
      setNodeRef: vi.fn(),
      transform: null,
      transition: undefined,
    });
  });

  it('configures keyboard reordering with sortable keyboard coordinates', () => {
    const exercises: Array<{
      id: string;
      exercise: WorkoutExerciseCard;
    }> = [
      { id: 'exercise-0', exercise: { name: 'Bench Press', sets: [] } },
      { id: 'exercise-1', exercise: { name: 'Barbell Squat', sets: [] } },
    ];

    render(
      <WorkoutExerciseCards
        exercises={exercises}
        dataTestId="exercise-card"
        canEditOrder
        hasPendingOrderChanges={false}
        isReorderMode
        onCancelOrder={vi.fn()}
        onEditOrder={vi.fn()}
        onRemoveExercise={vi.fn()}
        onReorderExercises={vi.fn()}
        onSaveOrder={vi.fn()}
        formatVolume={() => '0 lb'}
      />
    );

    expect(
      mockUseSensor.mock.calls.find(([sensor]) => sensor === keyboardSensorToken)
    ).toEqual([
      keyboardSensorToken,
      { coordinateGetter: mockSortableKeyboardCoordinates },
    ]);
    expect(screen.getAllByLabelText(/reorder /i)).toHaveLength(2);
  });

  it('exits reorder mode after saving a reordered list', () => {
    render(<WorkoutExerciseCardsHarness />);

    fireEvent.click(screen.getByTestId('edit-exercise-order'));
    fireEvent.click(screen.getByTestId('force-reorder'));
    fireEvent.click(screen.getByTestId('save-exercise-order'));

    expect(screen.queryAllByLabelText(/reorder /i)).toHaveLength(0);
    expect(screen.getByTestId('edit-exercise-order')).toBeInTheDocument();
    expect(
      screen
        .getAllByRole('button')
        .find((button) => button.textContent?.includes('Barbell Squat'))
    ).toHaveTextContent(
      'Barbell Squat'
    );
  });
});
