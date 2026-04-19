import { createEvent, fireEvent, render, screen } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import { GenericCombobox } from '@/components/generic-combobox';

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

async function renderGenericCombobox({
  isDesktop = false,
  onCreate = vi.fn(),
}: {
  isDesktop?: boolean;
  onCreate?: ReturnType<typeof vi.fn>;
} = {}) {
  mediaQueryMatches = isDesktop;

  render(
    <GenericCombobox
      options={[{ name: 'Squat' }]}
      selected=""
      ariaLabel="Exercise type"
      inputAriaLabel="Search options"
      onChange={vi.fn()}
      onCreate={onCreate}
    />
  );

  fireEvent.click(screen.getByRole('combobox', { name: 'Exercise type' }));
  fireEvent.change(screen.getByPlaceholderText('Search options...'), {
    target: { value: 'Bench' },
  });

  const [createRow] = await screen.findAllByRole('option', {
    name: 'Create "Bench"',
  });

  return { createRow, onCreate };
}

describe('GenericCombobox create row', () => {
  it('creates an option on touch release in the mobile drawer path', async () => {
    const { createRow, onCreate } = await renderGenericCombobox();

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

  it('does not create an option after a touch drag gesture', async () => {
    const { createRow, onCreate } = await renderGenericCombobox({ isDesktop: true });

    fireEvent.touchStart(createRow, {
      touches: [{ clientX: 8, clientY: 12 }],
      changedTouches: [{ clientX: 8, clientY: 12 }],
    });
    fireEvent.touchMove(createRow, {
      touches: [{ clientX: 8, clientY: 40 }],
      changedTouches: [{ clientX: 8, clientY: 40 }],
    });
    fireEvent.touchEnd(createRow, {
      changedTouches: [{ clientX: 8, clientY: 40 }],
    });

    expect(onCreate).not.toHaveBeenCalled();
  });

  it('does not create a duplicate option after a follow-up click', async () => {
    const { createRow, onCreate } = await renderGenericCombobox();

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
