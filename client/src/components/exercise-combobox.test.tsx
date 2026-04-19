import { createEvent, fireEvent, render, screen } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import { ExerciseCombobox } from '@/components/exercise-combobox';

let mediaQueryMatches = false;

beforeAll(() => {
  class ResizeObserverMock {
    observe() {}
    unobserve() {}
    disconnect() {}
  }

  Object.defineProperty(globalThis, 'ResizeObserver', {
    value: ResizeObserverMock,
    writable: true,
  });

  Object.defineProperty(HTMLElement.prototype, 'scrollIntoView', {
    value: () => {},
    writable: true,
  });

  Object.defineProperty(globalThis, 'matchMedia', {
    value: (query: string) =>
      ({
        matches: mediaQueryMatches,
        media: query,
        onchange: null,
        addEventListener: () => {},
        removeEventListener: () => {},
        addListener: () => {},
        removeListener: () => {},
        dispatchEvent: () => false,
      }) as MediaQueryList,
    writable: true,
  });
});

beforeEach(() => {
  mediaQueryMatches = false;
});

async function renderMobileExerciseCombobox(onCreate = vi.fn()) {
  render(
    <ExerciseCombobox
      options={[{ id: 1, name: 'Squat' }]}
      selected=""
      onChange={vi.fn()}
      onCreate={onCreate}
    />
  );

  const trigger = screen.getByText('Select exercise...').closest('button');

  if (!trigger) {
    throw new Error('Expected exercise combobox trigger button');
  }

  fireEvent.click(trigger);
  fireEvent.change(screen.getByPlaceholderText('Search exercises...'), {
    target: { value: 'Bench' },
  });

  const [createRow] = await screen.findAllByText('Create "Bench"');

  return { createRow, onCreate };
}

describe('ExerciseCombobox create row', () => {
  it('creates an option on touch release in the mobile drawer path', async () => {
    const { createRow, onCreate } = await renderMobileExerciseCombobox();

    fireEvent.touchStart(createRow, {
      touches: [{ clientX: 8, clientY: 12 }],
      changedTouches: [{ clientX: 8, clientY: 12 }],
    });

    const touchEnd = createEvent.touchEnd(createRow, {
      changedTouches: [{ clientX: 8, clientY: 12 }],
    });

    fireEvent(createRow, touchEnd);

    expect(onCreate).toHaveBeenCalledTimes(1);
    expect(onCreate).toHaveBeenCalledWith('Bench');
    expect(touchEnd.defaultPrevented).toBe(true);
  });

  it('does not create an option after touch tracking is canceled', async () => {
    const { createRow, onCreate } = await renderMobileExerciseCombobox();

    fireEvent.touchStart(createRow, {
      touches: [{ clientX: 8, clientY: 12 }],
      changedTouches: [{ clientX: 8, clientY: 12 }],
    });
    fireEvent.touchCancel(createRow);
    fireEvent.touchEnd(createRow, {
      changedTouches: [{ clientX: 8, clientY: 12 }],
    });

    expect(onCreate).not.toHaveBeenCalled();
  });

  it('does not create a duplicate option after a follow-up click', async () => {
    const { createRow, onCreate } = await renderMobileExerciseCombobox();

    fireEvent.touchStart(createRow, {
      touches: [{ clientX: 8, clientY: 12 }],
      changedTouches: [{ clientX: 8, clientY: 12 }],
    });
    fireEvent.touchEnd(createRow, {
      changedTouches: [{ clientX: 8, clientY: 12 }],
    });
    fireEvent.click(createRow);

    expect(onCreate).toHaveBeenCalledTimes(1);
  });
});
