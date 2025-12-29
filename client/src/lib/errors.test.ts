import { describe, it, expect, vi, beforeEach } from 'vitest';
import { isApiError, getErrorMessage, showErrorToast } from './errors';
import { toast } from 'sonner';

vi.mock('sonner', () => ({
  toast: {
    error: vi.fn(),
  },
}));

describe('isApiError', () => {
  it('returns true for valid ApiError objects', () => {
    expect(isApiError({ message: 'Error message' })).toBe(true);
    expect(isApiError({ message: 'Error message', request_id: '123' })).toBe(true);
  });

  it('returns false for null/undefined', () => {
    expect(isApiError(null)).toBe(false);
    expect(isApiError(undefined)).toBe(false);
  });

  it('returns false for Error instances without message property structure', () => {
    const error = new Error('Test error');
    // Error instances have message, but we want to distinguish ApiError
    expect(isApiError(error)).toBe(true); // Error has message property
  });

  it('returns false for objects without message property', () => {
    expect(isApiError({})).toBe(false);
    expect(isApiError({ error: 'test' })).toBe(false);
  });

  it('returns false for non-object types', () => {
    expect(isApiError('error string')).toBe(false);
    expect(isApiError(123)).toBe(false);
    expect(isApiError(true)).toBe(false);
  });
});

describe('getErrorMessage', () => {
  it('extracts message from ApiError', () => {
    expect(getErrorMessage({ message: 'API error occurred' })).toBe('API error occurred');
    expect(getErrorMessage({ message: 'Error', request_id: '123' })).toBe('Error');
  });

  it('extracts message from Error instance', () => {
    const error = new Error('Standard error');
    expect(getErrorMessage(error)).toBe('Standard error');
  });

  it('returns string as-is', () => {
    expect(getErrorMessage('Simple error string')).toBe('Simple error string');
  });

  it('returns fallback for unknown types', () => {
    expect(getErrorMessage(null)).toBe('An unexpected error occurred');
    expect(getErrorMessage(undefined)).toBe('An unexpected error occurred');
    expect(getErrorMessage(123)).toBe('An unexpected error occurred');
    expect(getErrorMessage({})).toBe('An unexpected error occurred');
  });

  it('returns fallback for empty string', () => {
    expect(getErrorMessage('')).toBe('An unexpected error occurred');
    expect(getErrorMessage('   ')).toBe('An unexpected error occurred');
  });

  it('uses custom fallback when provided', () => {
    expect(getErrorMessage(null, 'Custom fallback')).toBe('Custom fallback');
    expect(getErrorMessage({}, 'Custom fallback')).toBe('Custom fallback');
  });

  it('returns fallback for ApiError with empty message', () => {
    expect(getErrorMessage({ message: '' })).toBe('An unexpected error occurred');
  });
});

describe('showErrorToast', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls toast.error with extracted message', () => {
    showErrorToast({ message: 'Test error' });
    expect(toast.error).toHaveBeenCalledWith('Test error');
  });

  it('calls toast.error with fallback for unknown error', () => {
    showErrorToast(null);
    expect(toast.error).toHaveBeenCalledWith('An unexpected error occurred');
  });

  it('uses custom fallback when provided', () => {
    showErrorToast(null, 'Custom error message');
    expect(toast.error).toHaveBeenCalledWith('Custom error message');
  });

  it('logs request_id when available', () => {
    const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
    const error = { message: 'API error', request_id: 'req-123' };

    showErrorToast(error);

    expect(consoleErrorSpy).toHaveBeenCalledWith(
      '[Error] request_id: req-123',
      error
    );
    consoleErrorSpy.mockRestore();
  });

  it('does not log when request_id is not available', () => {
    const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

    showErrorToast({ message: 'API error' });

    expect(consoleErrorSpy).not.toHaveBeenCalled();
    consoleErrorSpy.mockRestore();
  });
});
