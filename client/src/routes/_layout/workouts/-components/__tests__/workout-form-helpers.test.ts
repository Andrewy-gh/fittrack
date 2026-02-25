import { describe, expect, it } from 'vitest';
import {
  isSetEmptyForDismiss,
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
});
