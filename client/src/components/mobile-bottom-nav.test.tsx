import { render, screen } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import { MobileBottomNav } from '@/components/mobile-bottom-nav';

vi.mock('@/components/mobile-nav-drawer', () => ({
  MobileNavDrawer: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="mobile-nav-drawer">{children}</div>
  ),
}));

vi.mock('@/components/custom-user-button', () => ({
  CustomUserButton: () => (
    <button type="button" aria-label="Signed-in user menu">
      user
    </button>
  ),
}));

vi.mock('@/components/guest-user-button', () => ({
  GuestUserButton: () => (
    <button type="button" aria-label="Guest user menu">
      guest
    </button>
  ),
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

    expect(
      await screen.findByRole('button', { name: 'Guest user menu' })
    ).toBeInTheDocument();
    expect(
      screen.queryByRole('button', { name: 'Signed-in user menu' })
    ).not.toBeInTheDocument();
  });

  it('renders the signed-in button for authenticated sessions', async () => {
    render(<MobileBottomNav includeChat isAuthenticated />);

    expect(
      await screen.findByRole('button', { name: 'Signed-in user menu' })
    ).toBeInTheDocument();
    expect(
      screen.queryByRole('button', { name: 'Guest user menu' })
    ).not.toBeInTheDocument();
  });
});
