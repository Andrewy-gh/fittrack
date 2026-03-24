import { afterEach, describe, expect, it, vi } from 'vitest';

afterEach(() => {
  vi.resetModules();
  vi.restoreAllMocks();
});

describe('local-storage module', () => {
  it('falls back when browser localStorage access throws during module init', async () => {
    vi.resetModules();

    const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => undefined);
    vi.spyOn(window, 'localStorage', 'get').mockImplementation(() => {
      throw new DOMException(
        'Access is denied for this document.',
        'SecurityError'
      );
    });

    const module = await import('./local-storage');

    expect(module.workoutDraftStorage.load('user-123')).toBeNull();
    expect(() =>
      module.workoutDraftStorage.save(
        {
          date: new Date('2026-03-24T10:30:00.000Z'),
          notes: '',
          workoutFocus: '',
          exercises: [],
        },
        'user-123'
      )
    ).not.toThrow();
    expect(() => module.workoutDraftStorage.clear('user-123')).not.toThrow();
    expect(warnSpy).toHaveBeenCalledWith(
      'Failed to access localStorage:',
      expect.objectContaining({ name: 'SecurityError' })
    );
  });
});
