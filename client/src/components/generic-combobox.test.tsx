import { createEvent, fireEvent, render, screen } from '@testing-library/react';
import { afterAll, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import { GenericCombobox } from '@/components/generic-combobox';

let mediaQueryMatches = false;
const originalResizeObserver = globalThis.ResizeObserver;
const originalScrollIntoView = HTMLElement.prototype.scrollIntoView;
const originalMatchMedia = globalThis.matchMedia;

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

afterAll(() => {
  Object.defineProperty(globalThis, 'ResizeObserver', {
    value: originalResizeObserver,
    writable: true,
  });

  Object.defineProperty(HTMLElement.prototype, 'scrollIntoView', {
    value: originalScrollIntoView,
    writable: true,
  });

  Object.defineProperty(globalThis, 'matchMedia', {
    value: originalMatchMedia,
    writable: true,
  });
});

function touchStart(target: Element, x = 8, y = 12) {
  fireEvent.touchStart(target, {
    touches: [{ clientX: x, clientY: y }],
    changedTouches: [{ clientX: x, clientY: y }],
  });
}

function touchEnd(target: Element, x = 8, y = 12) {
  const event = createEvent.touchEnd(target, {
    changedTouches: [{ clientX: x, clientY: y }],
  });

  fireEvent(target, event);
  return event;
}

function touchTap(target: Element, x = 8, y = 12) {
  touchStart(target, x, y);
  return touchEnd(target, x, y);
}

async function renderGenericCombobox({
  isDesktop = false,
  onChange = vi.fn(),
  onCreate = vi.fn(),
}: {
  isDesktop?: boolean;
  onChange?: ReturnType<typeof vi.fn>;
  onCreate?: ReturnType<typeof vi.fn>;
} = {}) {
  mediaQueryMatches = isDesktop;

  render(
    <GenericCombobox
      options={[{ name: 'Squat' }]}
      selected=""
      ariaLabel="Exercise type"
      inputAriaLabel="Search options"
      onChange={onChange}
      onCreate={onCreate}
    />
  );

  fireEvent.click(screen.getByRole('combobox', { name: 'Exercise type' }));

  return { onChange, onCreate };
}

describe('GenericCombobox touch activation', () => {
  it('selects an existing option on touch release in the mobile drawer path', async () => {
    const onChange = vi.fn();
    await renderGenericCombobox({ onChange });

    const option = await screen.findByRole('option', { name: 'Squat' });
    const releasedTouch = touchTap(option);

    expect(onChange).toHaveBeenCalledTimes(1);
    expect(onChange).toHaveBeenCalledWith({ name: 'Squat' });
    expect(releasedTouch.defaultPrevented).toBe(true);
  });

  it('does not select an existing option after a drag gesture', async () => {
    const onChange = vi.fn();
    await renderGenericCombobox({ onChange });

    const option = await screen.findByRole('option', { name: 'Squat' });

    touchStart(option);
    fireEvent.touchMove(option, {
      touches: [{ clientX: 8, clientY: 40 }],
      changedTouches: [{ clientX: 8, clientY: 40 }],
    });
    touchEnd(option, 8, 40);

    expect(onChange).not.toHaveBeenCalled();
  });

  it('does not select an existing option twice after a follow-up click', async () => {
    const onChange = vi.fn();
    await renderGenericCombobox({ onChange });

    const option = await screen.findByRole('option', { name: 'Squat' });

    touchTap(option);
    fireEvent.click(option);

    expect(onChange).toHaveBeenCalledTimes(1);
  });

  it('creates an option on touch release in the mobile drawer path', async () => {
    const { onCreate } = await renderGenericCombobox();

    fireEvent.change(screen.getByPlaceholderText('Search options...'), {
      target: { value: 'Bench' },
    });

    const [createRow] = await screen.findAllByRole('option', {
      name: 'Create "Bench"',
    });
    const releasedTouch = touchTap(createRow);

    expect(onCreate).toHaveBeenCalledTimes(1);
    expect(onCreate).toHaveBeenCalledWith('Bench');
    expect(releasedTouch.defaultPrevented).toBe(true);
  });

  it('does not create an option after touch tracking is canceled', async () => {
    const { onCreate } = await renderGenericCombobox();

    fireEvent.change(screen.getByPlaceholderText('Search options...'), {
      target: { value: 'Bench' },
    });

    const [createRow] = await screen.findAllByRole('option', {
      name: 'Create "Bench"',
    });

    touchStart(createRow);
    fireEvent.touchCancel(createRow);
    touchEnd(createRow);

    expect(onCreate).not.toHaveBeenCalled();
  });

  it('does not create a duplicate option after a follow-up click', async () => {
    const { onCreate } = await renderGenericCombobox();

    fireEvent.change(screen.getByPlaceholderText('Search options...'), {
      target: { value: 'Bench' },
    });

    const [createRow] = await screen.findAllByRole('option', {
      name: 'Create "Bench"',
    });

    touchTap(createRow);
    fireEvent.click(createRow);

    expect(onCreate).toHaveBeenCalledTimes(1);
  });
});
