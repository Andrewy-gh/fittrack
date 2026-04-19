import { render, screen } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import { MobileBottomNav } from '@/components/mobile-bottom-nav';

vi.mock('@/components/mobile-nav-drawer', () => ({
  MobileNavDrawer: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="mobile-nav-drawer">{children}</div>
  ),
}));

vi.mock('@/components/custom-user-button', () => ({
  CustomUserButton: () => <div data-testid="custom-user-button">user</div>,
}));

vi.mock('@/components/guest-user-button', () => ({
  GuestUserButton: () => <div data-testid="guest-user-button">guest</div>,
}));

beforeAll(() => {
  Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: (query: string) => ({
      matches:
        query === '(pointer: coarse)' ||
        query === '(max-width: 1024px)',
      media: query,
      onchange: null,
      addEventListener: () => {},
      removeEventListener: () => {},
      addListener: () => {},
      removeListener: () => {},
      dispatchEvent: () => false,
    }),
  });
});

beforeEach(() => {
  Object.defineProperty(navigator, 'maxTouchPoints', {
    configurable: true,
    value: 1,
  });
});

describe('MobileBottomNav', () => {
  it('renders the guest button in demo mode', async () => {
    render(<MobileBottomNav includeChat isAuthenticated={false} />);

    expect(await screen.findByTestId('guest-user-button')).toBeInTheDocument();
    expect(screen.queryByTestId('custom-user-button')).not.toBeInTheDocument();
  });

  it('renders the signed-in button for authenticated sessions', async () => {
    render(<MobileBottomNav includeChat isAuthenticated />);

    expect(await screen.findByTestId('custom-user-button')).toBeInTheDocument();
    expect(screen.queryByTestId('guest-user-button')).not.toBeInTheDocument();
  });
});
