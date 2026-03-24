import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { useAppForm } from '@/hooks/form';
import { AddSetDialog } from '../add-set-dialog';

type SetValue = {
  weight: number;
  reps: number;
  setType: 'warmup' | 'working';
};

function AddSetDialogHarness({
  initialSet,
  onClose = vi.fn(),
  onRemoveSet = vi.fn(),
  onSaveSet = vi.fn(),
}: {
  initialSet: SetValue;
  onClose?: ReturnType<typeof vi.fn>;
  onRemoveSet?: ReturnType<typeof vi.fn>;
  onSaveSet?: ReturnType<typeof vi.fn>;
}) {
  const form = useAppForm({
    defaultValues: {
      date: '2026-03-24T10:30:00.000Z',
      notes: '',
      workoutFocus: '',
      exercises: [
        {
          name: 'Bench Press',
          sets: [initialSet],
        },
      ],
    },
    onSubmit: async () => undefined,
  });

  return (
    <AddSetDialog
      form={form as any}
      exerciseIndex={0}
      setIndex={0}
      onClose={onClose}
      onRemoveSet={onRemoveSet}
      onSaveSet={onSaveSet}
    />
  );
}

describe('AddSetDialog', () => {
  it('shows reps validation errors and only enables save after the value is valid', async () => {
    const user = userEvent.setup();
    const onSaveSet = vi.fn();

    render(
      <AddSetDialogHarness
        initialSet={{ weight: 0, reps: 0, setType: 'working' }}
        onSaveSet={onSaveSet}
      />
    );

    const saveButton = await screen.findByRole('button', { name: 'Save Set' });
    await waitFor(() => {
      expect(screen.getByText('Reps must be at least 1')).toBeInTheDocument();
    });
    expect(saveButton).toBeDisabled();
    expect(screen.getByRole('button', { name: 'Cancel' })).toBeInTheDocument();

    const repsInput = screen.getByLabelText('Reps');
    await user.clear(repsInput);
    await user.type(repsInput, '8');

    await waitFor(() => {
      expect(
        screen.queryByText('Reps must be at least 1')
      ).not.toBeInTheDocument();
    });
    expect(saveButton).toBeEnabled();

    await user.click(saveButton);

    expect(onSaveSet).toHaveBeenCalledTimes(1);
  });

  it('dismisses an empty set by removing it', async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();
    const onRemoveSet = vi.fn();

    render(
      <AddSetDialogHarness
        initialSet={{ weight: 0, reps: 0, setType: 'working' }}
        onClose={onClose}
        onRemoveSet={onRemoveSet}
      />
    );

    await user.click(await screen.findByRole('button', { name: 'Close' }));

    expect(onRemoveSet).toHaveBeenCalledTimes(1);
    expect(onClose).not.toHaveBeenCalled();
  });

  it('dismisses a populated set without removing it', async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();
    const onRemoveSet = vi.fn();

    render(
      <AddSetDialogHarness
        initialSet={{ weight: 135, reps: 5, setType: 'working' }}
        onClose={onClose}
        onRemoveSet={onRemoveSet}
      />
    );

    expect(
      await screen.findByRole('button', { name: 'Remove Set' })
    ).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: 'Close' }));

    expect(onClose).toHaveBeenCalledTimes(1);
    expect(onRemoveSet).not.toHaveBeenCalled();
  });
});
