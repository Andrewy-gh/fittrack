import { describe, it, expect } from 'vitest';
import {
  required,
  minValue,
  maxValue,
  minLength,
  maxLength,
  email,
  url,
  compose,
} from './validation';

describe('required', () => {
  it('returns error for empty string', () => {
    expect(required('')).toBe('This field is required');
    expect(required('   ')).toBe('This field is required');
  });

  it('returns undefined for non-empty string', () => {
    expect(required('test')).toBeUndefined();
    expect(required('  test  ')).toBeUndefined();
  });

  it('returns error for null/undefined', () => {
    expect(required(null)).toBe('This field is required');
    expect(required(undefined)).toBe('This field is required');
  });

  it('returns undefined for numbers (including 0)', () => {
    expect(required(0)).toBeUndefined();
    expect(required(1)).toBeUndefined();
    expect(required(-1)).toBeUndefined();
    expect(required(100)).toBeUndefined();
  });

  it('uses custom field name in error message', () => {
    expect(required('', 'Email')).toBe('Email is required');
    expect(required(null, 'Username')).toBe('Username is required');
  });
});

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

describe('minLength', () => {
  it('returns error when string too short', () => {
    const validator = minLength(5);
    expect(validator('test')).toBe('This field must be at least 5 characters');
    expect(validator('ab')).toBe('This field must be at least 5 characters');
  });

  it('returns undefined when string meets minimum', () => {
    const validator = minLength(5);
    expect(validator('hello')).toBeUndefined();
    expect(validator('hello world')).toBeUndefined();
  });

  it('trims whitespace before checking length', () => {
    const validator = minLength(5);
    expect(validator('  hi  ')).toBe('This field must be at least 5 characters');
    expect(validator('  hello  ')).toBeUndefined();
  });

  it('handles non-string inputs gracefully', () => {
    const validator = minLength(5);
    expect(validator(123)).toBeUndefined();
    expect(validator(null)).toBeUndefined();
  });
});

describe('maxLength', () => {
  it('returns error when string too long', () => {
    const validator = maxLength(5);
    expect(validator('toolong')).toBe('This field must be 5 characters or less');
    expect(validator('hello world')).toBe('This field must be 5 characters or less');
  });

  it('returns undefined when string within limit', () => {
    const validator = maxLength(5);
    expect(validator('hi')).toBeUndefined();
    expect(validator('hello')).toBeUndefined();
  });

  it('does not trim whitespace for max length', () => {
    const validator = maxLength(5);
    expect(validator('  hi  ')).toBeUndefined(); // 6 chars including spaces
  });

  it('handles non-string inputs gracefully', () => {
    const validator = maxLength(5);
    expect(validator(123)).toBeUndefined();
    expect(validator(null)).toBeUndefined();
  });
});

describe('email', () => {
  it('returns undefined for valid email addresses', () => {
    expect(email('test@example.com')).toBeUndefined();
    expect(email('user.name+tag@example.co.uk')).toBeUndefined();
    expect(email('user_name@example-domain.com')).toBeUndefined();
  });

  it('returns error for invalid email addresses', () => {
    expect(email('invalid')).toBe('Email must be a valid email address');
    expect(email('test@')).toBe('Email must be a valid email address');
    expect(email('@example.com')).toBe('Email must be a valid email address');
    expect(email('test@example')).toBe('Email must be a valid email address');
  });

  it('handles non-string inputs gracefully', () => {
    expect(email(123)).toBeUndefined();
    expect(email(null)).toBeUndefined();
  });

  it('uses custom field name in error message', () => {
    expect(email('invalid', 'User email')).toBe('User email must be a valid email address');
  });
});

describe('url', () => {
  it('returns undefined for valid URLs', () => {
    expect(url('https://example.com')).toBeUndefined();
    expect(url('http://example.com')).toBeUndefined();
    expect(url('https://example.com/path?query=value')).toBeUndefined();
  });

  it('returns error for invalid URLs', () => {
    expect(url('invalid')).toBe('URL must be a valid URL');
    expect(url('example.com')).toBe('URL must be a valid URL');
    expect(url('htp://example.com')).toBe('URL must be a valid URL');
  });

  it('handles non-string inputs gracefully', () => {
    expect(url(123)).toBeUndefined();
    expect(url(null)).toBeUndefined();
  });

  it('uses custom field name in error message', () => {
    expect(url('invalid', 'Website')).toBe('Website must be a valid URL');
  });
});

describe('compose', () => {
  it('returns first error from multiple validators', () => {
    const validator = compose(required, minLength(5));
    expect(validator('')).toBe('This field is required');
    expect(validator('ab')).toBe('This field must be at least 5 characters');
  });

  it('returns undefined when all pass', () => {
    const validator = compose(required, minLength(3), maxLength(10));
    expect(validator('hello')).toBeUndefined();
  });

  it('stops at first error', () => {
    const validator = compose(
      required,
      minLength(5),
      maxLength(10)
    );
    // Should return the required error, not check length validators
    expect(validator('')).toBe('This field is required');
  });

  it('works with custom validators returning functions', () => {
    const validator = compose(minValue(1), maxValue(100));
    expect(validator(0)).toBe('This field must be at least 1');
    expect(validator(101)).toBe('This field must be at most 100');
    expect(validator(50)).toBeUndefined();
  });

  it('passes field name to all validators', () => {
    const validator = compose(required, minLength(5));
    expect(validator('', 'Username')).toBe('Username is required');
    expect(validator('ab', 'Username')).toBe('Username must be at least 5 characters');
  });
});
