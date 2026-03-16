import type { ComponentPropsWithoutRef } from 'react';
import { render, screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

vi.mock('@tanstack/react-router', () => ({
  Link: ({
    children,
    preload: _preload,
    ...props
  }: ComponentPropsWithoutRef<'a'> & { preload?: boolean }) => (
    <a {...props}>{children}</a>
  ),
  createFileRoute: () => () => ({
    useRouteContext: () => ({ user: null }),
  }),
}));

import { HomePage } from './index';

describe('HomePage feature cards', () => {
  it('uses the favicon for the navbar brand mark', () => {
    render(<HomePage user={null} />);

    const brandText = screen.getByText('FITTRACK');
    const nav = brandText.closest('nav');

    expect(nav).not.toBeNull();
    expect((nav as HTMLElement).querySelector('img')).toHaveAttribute(
      'src',
      '/favicon.svg'
    );
  });

  it('removes highlight badges from the grounded features cards', () => {
    render(<HomePage user={null} />);

    const groundedFeaturesHeading = screen.getByRole('heading', {
      name: /grounded features/i,
    });
    const groundedFeaturesSection = groundedFeaturesHeading.closest('section');

    expect(groundedFeaturesSection).not.toBeNull();

    const section = within(groundedFeaturesSection as HTMLElement);

    expect(
      section.getByRole('heading', { name: 'Fast workout logging' })
    ).toBeInTheDocument();
    expect(
      section.getByRole('heading', { name: 'Repeat what you already did' })
    ).toBeInTheDocument();
    expect(section.queryByText('QUICK ENTRY')).not.toBeInTheDocument();
    expect(section.queryByText('REPEAT LAST')).not.toBeInTheDocument();
    expect(section.queryByText('CLEAR HISTORY')).not.toBeInTheDocument();
    expect(section.queryByText('PLAIN SUMMARY')).not.toBeInTheDocument();
  });
});
