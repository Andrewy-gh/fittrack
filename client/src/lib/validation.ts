export function minValue(min: number) {
  return (value: unknown, fieldName = 'This field'): string | undefined => {
    if (typeof value === 'number' && value < min) {
      return `${fieldName} must be at least ${min}`;
    }
    return undefined;
  };
}

export function maxValue(max: number) {
  return (value: unknown, fieldName = 'This field'): string | undefined => {
    if (typeof value === 'number' && value > max) {
      return `${fieldName} must be at most ${max}`;
    }
    return undefined;
  };
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
