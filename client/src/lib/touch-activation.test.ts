import {
  beginTouchTapTracking,
  finishTouchTapTracking,
  hasRecentTouchActivation,
  markRecentTouchActivation,
  updateTouchTapTracking,
} from '@/lib/touch-activation';
import { describe, expect, it } from 'vitest';

function createTouchEvent(x: number, y: number) {
  return {
    touches: [{ clientX: x, clientY: y }],
    changedTouches: [{ clientX: x, clientY: y }],
  } as unknown as TouchEvent;
}

describe('touch activation helpers', () => {
  it('treats a short touch as a tap', () => {
    const element = document.createElement('div');

    beginTouchTapTracking(element, createTouchEvent(10, 20));

    expect(finishTouchTapTracking(element, createTouchEvent(14, 24))).toBe(true);
  });

  it('treats a moved touch as a drag', () => {
    const element = document.createElement('div');

    beginTouchTapTracking(element, createTouchEvent(10, 20));
    updateTouchTapTracking(element, createTouchEvent(10, 42));

    expect(finishTouchTapTracking(element, createTouchEvent(10, 42))).toBe(false);
  });

  it('tracks recent touch activation windows', () => {
    const element = document.createElement('div');

    markRecentTouchActivation(element, 100);

    expect(hasRecentTouchActivation(element, 200)).toBe(true);
    expect(hasRecentTouchActivation(element, 900)).toBe(false);
  });
});
