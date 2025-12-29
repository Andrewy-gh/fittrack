import React, { Component } from 'react';
import type { ReactNode } from 'react';
import { Button } from '@/components/ui/button';
import { AlertCircle } from 'lucide-react';

interface ErrorBoundaryProps {
  children: ReactNode;
  fallback: ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error?: Error;
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return this.props.fallback;
    }

    return this.props.children;
  }
}

// Reusable fallback UI components

interface FullScreenErrorFallbackProps {
  message: string;
  onAction?: () => void;
  actionLabel?: string;
}

export function FullScreenErrorFallback({
  message,
  onAction,
  actionLabel = 'Reload Page',
}: FullScreenErrorFallbackProps) {
  const handleAction = onAction || (() => window.location.reload());

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-background">
      <div className="text-center space-y-4 max-w-md">
        <div className="flex justify-center">
          <AlertCircle className="w-12 h-12 text-destructive" />
        </div>
        <h2 className="text-xl font-semibold text-foreground">Something went wrong</h2>
        <p className="text-destructive font-medium">{message}</p>
        <Button onClick={handleAction} className="mt-4">
          {actionLabel}
        </Button>
      </div>
    </div>
  );
}

interface InlineErrorFallbackProps {
  message: string;
}

export function InlineErrorFallback({ message }: InlineErrorFallbackProps) {
  return (
    <div className="flex items-center gap-2 p-4 rounded-lg bg-destructive/10 border border-destructive/20">
      <AlertCircle className="w-5 h-5 text-destructive flex-shrink-0" />
      <p className="text-sm text-destructive font-medium">{message}</p>
    </div>
  );
}
