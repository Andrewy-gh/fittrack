import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { ErrorBoundary, FullScreenErrorFallback, InlineErrorFallback } from './error-boundary';
import userEvent from '@testing-library/user-event';

// Component that throws an error
function ThrowError({ shouldThrow }: { shouldThrow: boolean }) {
  if (shouldThrow) {
    throw new Error('Test error');
  }
  return <div>No error</div>;
}

describe('ErrorBoundary', () => {
  beforeEach(() => {
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  it('renders children when no error', () => {
    render(
      <ErrorBoundary fallback={<div>Error occurred</div>}>
        <div>Child content</div>
      </ErrorBoundary>
    );

    expect(screen.getByText('Child content')).toBeInTheDocument();
    expect(screen.queryByText('Error occurred')).not.toBeInTheDocument();
  });

  it('renders fallback when child throws', () => {
    render(
      <ErrorBoundary fallback={<div>Error occurred</div>}>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    );

    expect(screen.queryByText('No error')).not.toBeInTheDocument();
    expect(screen.getByText('Error occurred')).toBeInTheDocument();
  });

  it('logs error to console', () => {
    const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

    render(
      <ErrorBoundary fallback={<div>Error occurred</div>}>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    );

    expect(consoleErrorSpy).toHaveBeenCalled();
    consoleErrorSpy.mockRestore();
  });
});

describe('FullScreenErrorFallback', () => {
  it('renders error message', () => {
    render(<FullScreenErrorFallback message="Test error message" />);

    expect(screen.getByRole('heading', { name: /something went wrong/i })).toBeInTheDocument();
    expect(screen.getByText('Test error message')).toBeInTheDocument();
  });

  it('calls onAction when button clicked', async () => {
    const user = userEvent.setup();
    const onAction = vi.fn();

    render(
      <FullScreenErrorFallback
        message="Test error"
        onAction={onAction}
        actionLabel="Retry"
      />
    );

    const button = screen.getByRole('button', { name: 'Retry' });
    await user.click(button);

    expect(onAction).toHaveBeenCalledTimes(1);
  });

  it('reloads page when no onAction provided', async () => {
    const user = userEvent.setup();
    const reloadSpy = vi.fn();
    Object.defineProperty(window, 'location', {
      value: { reload: reloadSpy },
      writable: true,
    });

    render(<FullScreenErrorFallback message="Test error" />);

    const button = screen.getByRole('button', { name: 'Reload Page' });
    await user.click(button);

    expect(reloadSpy).toHaveBeenCalledTimes(1);
  });

  it('uses default action label when not provided', () => {
    render(<FullScreenErrorFallback message="Test error" />);

    expect(screen.getByRole('button', { name: 'Reload Page' })).toBeInTheDocument();
  });

  it('uses custom action label when provided', () => {
    render(
      <FullScreenErrorFallback
        message="Test error"
        actionLabel="Try Again"
      />
    );

    expect(screen.getByRole('button', { name: 'Try Again' })).toBeInTheDocument();
  });
});

describe('InlineErrorFallback', () => {
  it('renders error message', () => {
    render(<InlineErrorFallback message="Failed to load data" />);

    expect(screen.getByText('Failed to load data')).toBeInTheDocument();
  });

  it('displays with correct styling classes', () => {
    const { container } = render(<InlineErrorFallback message="Test error" />);

    const errorDiv = container.querySelector('.bg-destructive\\/10');
    expect(errorDiv).toBeInTheDocument();
  });
});
