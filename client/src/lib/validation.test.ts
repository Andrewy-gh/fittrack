import { describe, it, expect } from 'vitest';
import {
  minValue,
  maxValue,
  compose,
} from './validation';

describe('minValue', () => {
  it('returns error when value below threshold', () => {
    const validator = minValue(10);
    expect(validator(5)).toBe('This field must be at least 10');
    expect(validator(0)).toBe('This field must be at least 10');
    expect(validator(-5)).toBe('This field must be at least 10');
  });

  it('returns undefined when value within range', () => {
    const validator = minValue(10);
    expect(validator(10)).toBeUndefined();
    expect(validator(11)).toBeUndefined();
    expect(validator(100)).toBeUndefined();
  });

  it('handles non-number inputs gracefully', () => {
    const validator = minValue(10);
    expect(validator('test')).toBeUndefined();
    expect(validator(null)).toBeUndefined();
    expect(validator(undefined)).toBeUndefined();
  });

  it('uses custom field name in error message', () => {
    const validator = minValue(1);
    expect(validator(0, 'Reps')).toBe('Reps must be at least 1');
  });
});

describe('maxValue', () => {
  it('returns error when value above threshold', () => {
    const validator = maxValue(100);
    expect(validator(101)).toBe('This field must be at most 100');
    expect(validator(200)).toBe('This field must be at most 100');
  });

  it('returns undefined when value within range', () => {
    const validator = maxValue(100);
    expect(validator(100)).toBeUndefined();
    expect(validator(99)).toBeUndefined();
    expect(validator(0)).toBeUndefined();
    expect(validator(-10)).toBeUndefined();
  });

  it('handles non-number inputs gracefully', () => {
    const validator = maxValue(100);
    expect(validator('test')).toBeUndefined();
    expect(validator(null)).toBeUndefined();
    expect(validator(undefined)).toBeUndefined();
  });

  it('uses custom field name in error message', () => {
    const validator = maxValue(1000);
    expect(validator(1001, 'Reps')).toBe('Reps must be at most 1000');
  });
});

describe('compose', () => {
  it('returns first error from multiple validators', () => {
    const validator = compose(minValue(5), maxValue(10));
    expect(validator(4)).toBe('This field must be at least 5');
    expect(validator(11)).toBe('This field must be at most 10');
  });

  it('returns undefined when all pass', () => {
    const validator = compose(minValue(3), maxValue(10));
    expect(validator(5)).toBeUndefined();
  });

  it('works with custom validators returning functions', () => {
    const validator = compose(minValue(1), maxValue(100));
    expect(validator(0)).toBe('This field must be at least 1');
    expect(validator(101)).toBe('This field must be at most 100');
    expect(validator(50)).toBeUndefined();
  });

  it('passes field name to all validators', () => {
    const validator = compose(minValue(5), maxValue(10));
    expect(validator(4, 'Score')).toBe('Score must be at least 5');
    expect(validator(11, 'Score')).toBe('Score must be at most 10');
  });
});
