import { toast } from "sonner";

/**
 * API Error shape matching backend response.ErrorResponse
 */
export interface ApiError {
  message: string;
  request_id?: string;
}

/**
 * Type guard to check if an error is an ApiError
 */
export function isApiError(error: unknown): error is ApiError {
  return (
    typeof error === "object" &&
    error !== null &&
    "message" in error &&
    typeof (error as ApiError).message === "string"
  );
}

/**
 * Extract error message from various error types
 */
export function getErrorMessage(error: unknown, fallback = "An unexpected error occurred"): string {
  // Check if it's an ApiError with a non-empty message
  if (isApiError(error) && error.message) {
    return error.message;
  }

  // Check if it's a standard Error object with a non-empty message
  if (error instanceof Error && error.message) {
    return error.message;
  }

  // Check if it's a non-empty string
  if (typeof error === "string" && error.trim()) {
    return error;
  }

  // Return fallback for unknown/empty error types
  return fallback;
}

/**
 * Show error toast with proper message extraction
 */
export function showErrorToast(error: unknown, fallback?: string): void {
  const message = getErrorMessage(error, fallback);
  toast.error(message);

  // Log request_id if available (dev debugging)
  if (isApiError(error) && error.request_id) {
    console.error(`[Error] request_id: ${error.request_id}`, error);
  }
}
