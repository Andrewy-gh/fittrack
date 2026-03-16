import { describe, expect, it } from 'vitest';
import {
  hasWorkoutDraftContent,
  isSetEmptyForDismiss,
  shouldShowRecentFocusAreaCard,
  shouldDiscardNewExerciseAfterSetRemoval,
  validateSetReps,
} from '../workout-form-helpers';

describe('workout-form-helpers', () => {
  describe('validateSetReps', () => {
    it('returns error when reps are below minimum', () => {
      expect(validateSetReps(0)).toBe('Reps must be at least 1');
    });

    it('returns undefined when reps are within range', () => {
      expect(validateSetReps(8)).toBeUndefined();
    });
  });

  describe('isSetEmptyForDismiss', () => {
    it('treats default zeroed working set as empty', () => {
      expect(
        isSetEmptyForDismiss({
          reps: 0,
          weight: 0,
          setType: 'working',
        })
      ).toBe(true);
    });

    it('treats set with reps as non-empty', () => {
      expect(
        isSetEmptyForDismiss({
          reps: 5,
          weight: 0,
          setType: 'working',
        })
      ).toBe(false);
    });
  });

  describe('shouldDiscardNewExerciseAfterSetRemoval', () => {
    it('discards when removing the only set from a new exercise', () => {
      expect(shouldDiscardNewExerciseAfterSetRemoval(true, 1)).toBe(true);
    });

    it('does not discard existing exercise', () => {
      expect(shouldDiscardNewExerciseAfterSetRemoval(false, 1)).toBe(false);
    });
  });

  describe('hasWorkoutDraftContent', () => {
    it('returns false for an untouched draft', () => {
      expect(
        hasWorkoutDraftContent({
          notes: '',
          workoutFocus: '',
          exercises: [],
        })
      ).toBe(false);
    });

    it('returns true when notes have been typed but not persisted yet', () => {
      expect(
        hasWorkoutDraftContent({
          notes: 'Felt strong today',
          workoutFocus: '',
          exercises: [],
        })
      ).toBe(true);
    });

    it('returns true when exercises are present', () => {
      expect(
        hasWorkoutDraftContent({
          notes: '',
          workoutFocus: '',
          exercises: [
            {
              name: 'Bench Press',
              sets: [{ reps: 5, weight: 185, setType: 'working' }],
            },
          ],
        })
      ).toBe(true);
    });
  });

  describe('shouldShowRecentFocusAreaCard', () => {
    it('shows the card for an untouched empty draft when templates exist', () => {
      expect(
        shouldShowRecentFocusAreaCard({
          focusAreaTemplateCount: 2,
          isDirty: false,
          value: {
            notes: '',
            workoutFocus: '',
            exercises: [],
          },
        })
      ).toBe(true);
    });

    it('hides the card after the user starts typing', () => {
      expect(
        shouldShowRecentFocusAreaCard({
          focusAreaTemplateCount: 2,
          isDirty: true,
          value: {
            notes: 'Leg day',
            workoutFocus: '',
            exercises: [],
          },
        })
      ).toBe(false);
    });

    it('keeps the card hidden for restored draft content even if the form is not dirty', () => {
      expect(
        shouldShowRecentFocusAreaCard({
          focusAreaTemplateCount: 2,
          isDirty: false,
          value: {
            notes: '',
            workoutFocus: 'Upper body',
            exercises: [],
          },
        })
      ).toBe(false);
    });

    it('shows the card again after the draft is cleared', () => {
      expect(
        shouldShowRecentFocusAreaCard({
          focusAreaTemplateCount: 2,
          isDirty: false,
          value: {
            notes: '',
            workoutFocus: '',
            exercises: [],
          },
        })
      ).toBe(true);
    });
  });
});
