import { useState } from 'react';
import { Button } from '@/components/ui/button';

/**
 * Test component to verify ErrorBoundary works correctly.
 *
 * Usage:
 * 1. Import this component in a route
 * 2. Wrap it with ErrorBoundary
 * 3. Click "Trigger Error" to test error handling
 *
 * Example:
 * ```tsx
 * import { ErrorBoundary, FullScreenErrorFallback } from '@/components/error-boundary';
 * import { ErrorBoundaryTest } from '@/components/error-boundary-test';
 *
 * <ErrorBoundary fallback={<FullScreenErrorFallback message="Component crashed!" />}>
 *   <ErrorBoundaryTest />
 * </ErrorBoundary>
 * ```
 */
export function ErrorBoundaryTest() {
  const [shouldError, setShouldError] = useState(false);

  if (shouldError) {
    throw new Error('Test error: ErrorBoundary is working!');
  }

  return (
    <div className="p-8 max-w-md mx-auto">
      <div className="space-y-4 border rounded-lg p-6 bg-card">
        <h2 className="text-xl font-semibold">ErrorBoundary Test</h2>
        <p className="text-sm text-muted-foreground">
          Click the button below to trigger a React error and test the ErrorBoundary.
        </p>
        <Button
          onClick={() => setShouldError(true)}
          variant="destructive"
          className="w-full"
        >
          Trigger Error
        </Button>
      </div>
    </div>
  );
}
