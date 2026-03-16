import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { LastWorkoutNoteSection } from './last-workout-note-section';

describe('LastWorkoutNoteSection', () => {
  it('is collapsed by default and reveals the note on demand', () => {
    render(
      <LastWorkoutNoteSection
        title="Last Workout Note"
        note="Felt better after the second warm-up set."
        dateLabel="2026-03-10"
      />
    );

    expect(
      screen.getByRole('button', { name: /show last workout note/i })
    ).toBeInTheDocument();
    expect(
      screen.queryByText('Felt better after the second warm-up set.')
    ).not.toBeInTheDocument();

    fireEvent.click(
      screen.getByRole('button', { name: /show last workout note/i })
    );

    expect(
      screen.getByRole('button', { name: /hide last workout note/i })
    ).toBeInTheDocument();
    expect(
      screen.getByText('Felt better after the second warm-up set.')
    ).toBeInTheDocument();
    expect(screen.getByText('Last Workout Note')).toBeInTheDocument();
  });

  it('renders nothing when the note is blank', () => {
    const { container } = render(
      <LastWorkoutNoteSection title="Last Workout Note" note="   " />
    );

    expect(container.firstChild).toBeNull();
  });
});
