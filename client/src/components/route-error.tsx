import { useRouter } from '@tanstack/react-router';
import type { ErrorComponentProps } from '@tanstack/react-router';
import { Button } from '@/components/ui/button';
import { isApiError, getErrorMessage } from '@/lib/errors';
import { AlertCircle } from 'lucide-react';

export function RouteError({ error, reset }: ErrorComponentProps) {
  const router = useRouter();

  const errorMessage = getErrorMessage(error);

  // Extract request_id if it's an API error
  const errorAsUnknown = error as unknown;
  let requestId: string | undefined = undefined;
  if (isApiError(errorAsUnknown)) {
    requestId = errorAsUnknown.request_id;
  }

  return (
    <div className="flex min-h-screen items-center justify-center p-4">
      <div className="w-full max-w-md space-y-6 rounded-lg border border-destructive/50 bg-background p-6 shadow-lg">
        <div className="flex items-center gap-3">
          <AlertCircle className="h-8 w-8 text-destructive" />
          <h1 className="text-2xl font-semibold">Something went wrong</h1>
        </div>

        <div className="space-y-2">
          <p className="text-sm text-muted-foreground">
            We encountered an error while loading this page:
          </p>
          <p className="rounded-md bg-muted p-3 text-sm font-medium">
            {errorMessage}
          </p>
          {requestId && import.meta.env.DEV && (
            <p className="text-xs text-muted-foreground">
              Request ID: <code className="rounded bg-muted px-1">{requestId}</code>
            </p>
          )}
        </div>

        <div className="flex flex-col gap-2 sm:flex-row">
          <Button
            onClick={() => router.history.back()}
            variant="outline"
            className="flex-1"
          >
            Go Back
          </Button>
          <Button
            onClick={() => reset()}
            className="flex-1"
          >
            Try Again
          </Button>
        </div>
      </div>
    </div>
  );
}
