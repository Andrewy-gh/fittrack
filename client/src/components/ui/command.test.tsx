import { fireEvent, render } from '@testing-library/react';
import { beforeAll, describe, expect, it, vi } from 'vitest';
import { Command, CommandGroup, CommandItem, CommandList } from '@/components/ui/command';

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
});

describe('CommandList', () => {
  it('contains vertical touch scrolling inside the list', () => {
    const { container } = render(
      <Command>
        <CommandList>
          <CommandGroup>
            <CommandItem value="squat">Squat</CommandItem>
          </CommandGroup>
        </CommandList>
      </Command>
    );

    const list = container.querySelector('[data-slot="command-list"]');

    expect(list).not.toBeNull();
    if (!list) {
      throw new Error('Expected command list to render');
    }

    expect(list).toHaveClass('overscroll-contain');
    expect(list).toHaveClass('touch-pan-y');
  });

  it('selects an item on touch release', () => {
    const onSelect = vi.fn();
    const { getByRole } = render(
      <Command>
        <CommandList>
          <CommandGroup>
            <CommandItem value="squat" onSelect={onSelect}>
              Squat
            </CommandItem>
          </CommandGroup>
        </CommandList>
      </Command>
    );

    const option = getByRole('option', { name: 'Squat' });

    fireEvent.touchStart(option, {
      touches: [{ clientX: 8, clientY: 12 }],
      changedTouches: [{ clientX: 8, clientY: 12 }],
    });
    fireEvent.touchEnd(option, {
      changedTouches: [{ clientX: 8, clientY: 12 }],
    });

    expect(onSelect).toHaveBeenCalledTimes(1);
    expect(onSelect).toHaveBeenCalledWith('squat');
  });

  it('does not select an item after a drag gesture', () => {
    const onSelect = vi.fn();
    const { getByRole } = render(
      <Command>
        <CommandList>
          <CommandGroup>
            <CommandItem value="squat" onSelect={onSelect}>
              Squat
            </CommandItem>
          </CommandGroup>
        </CommandList>
      </Command>
    );

    const option = getByRole('option', { name: 'Squat' });

    fireEvent.touchStart(option, {
      touches: [{ clientX: 8, clientY: 12 }],
      changedTouches: [{ clientX: 8, clientY: 12 }],
    });
    fireEvent.touchMove(option, {
      touches: [{ clientX: 8, clientY: 40 }],
      changedTouches: [{ clientX: 8, clientY: 40 }],
    });
    fireEvent.touchEnd(option, {
      changedTouches: [{ clientX: 8, clientY: 40 }],
    });

    expect(onSelect).not.toHaveBeenCalled();
  });
});
