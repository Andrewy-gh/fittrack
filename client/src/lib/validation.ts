/**
 * Reusable validation functions for TanStack Form
 */

/**
 * Validates that a value is not empty
 */
export function required(value: unknown, fieldName = 'This field'): string | undefined {
  if (typeof value === 'string') {
    return value.trim() === '' ? `${fieldName} is required` : undefined;
  }
  if (typeof value === 'number') {
    return undefined; // Numbers are always valid if provided
  }
  if (value === null || value === undefined) {
    return `${fieldName} is required`;
  }
  return undefined;
}

/**
 * Validates maximum string length
 */
export function maxLength(max: number) {
  return (value: unknown, fieldName = 'This field'): string | undefined => {
    if (typeof value === 'string' && value.length > max) {
      return `${fieldName} must be ${max} characters or less`;
    }
    return undefined;
  };
}

/**
 * Validates minimum string length
 */
export function minLength(min: number) {
  return (value: unknown, fieldName = 'This field'): string | undefined => {
    if (typeof value === 'string' && value.trim().length < min) {
      return `${fieldName} must be at least ${min} characters`;
    }
    return undefined;
  };
}

/**
 * Validates minimum number value
 */
export function minValue(min: number) {
  return (value: unknown, fieldName = 'This field'): string | undefined => {
    if (typeof value === 'number' && value < min) {
      return `${fieldName} must be at least ${min}`;
    }
    return undefined;
  };
}

/**
 * Validates maximum number value
 */
export function maxValue(max: number) {
  return (value: unknown, fieldName = 'This field'): string | undefined => {
    if (typeof value === 'number' && value > max) {
      return `${fieldName} must be at most ${max}`;
    }
    return undefined;
  };
}

/**
 * Validates email format
 */
export function email(value: unknown, fieldName = 'Email'): string | undefined {
  if (typeof value !== 'string') return undefined;
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(value) ? undefined : `${fieldName} must be a valid email address`;
}

/**
 * Validates URL format
 */
export function url(value: unknown, fieldName = 'URL'): string | undefined {
  if (typeof value !== 'string') return undefined;
  try {
    new URL(value);
    return undefined;
  } catch {
    return `${fieldName} must be a valid URL`;
  }
}

/**
 * Compose multiple validators - returns first error found
 */
export function compose(...validators: Array<(value: unknown, fieldName?: string) => string | undefined>) {
  return (value: unknown, fieldName?: string): string | undefined => {
    for (const validator of validators) {
      const error = validator(value, fieldName);
      if (error) return error;
    }
    return undefined;
  };
}
