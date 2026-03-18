import { render } from '@testing-library/react';
import { beforeAll, describe, expect, it } from 'vitest';
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
});
