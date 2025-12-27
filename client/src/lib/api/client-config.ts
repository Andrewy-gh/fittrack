import { client } from '@/client/client.gen';
import { stackClientApp } from '@/stack';
import type { ApiError } from '@/lib/errors';
import { showErrorToast } from '@/lib/errors';

const BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api';

client.setConfig({
  baseUrl: BASE_URL,
});

client.interceptors.request.use(async (request) => {
  const user = await stackClientApp.getUser();
  if (user) {
    const { accessToken } = await user.getAuthJson();
    if (accessToken) {
      request.headers.set('x-stack-access-token', accessToken);
    }
  }
  return request;
});

client.interceptors.response.use(async (response) => {
  // If response is ok, return as-is
  if (response.ok) {
    return response;
  }

  // Try to parse error response
  try {
    const error: ApiError = await response.json();

    // Log request_id in development for debugging
    if (import.meta.env.DEV && error.request_id) {
      console.error(`[API Error] ${response.status} ${response.statusText}`, {
        message: error.message,
        request_id: error.request_id,
        url: response.url,
      });
    }

    // Handle specific status codes
    switch (response.status) {
      case 401:
        // Unauthorized - session expired or invalid
        // Sign out and let Stack auth handle re-authentication
        stackClientApp.getUser().then((user) => {
          if (user) {
            user.signOut().catch((err: unknown) => {
              console.error('Error signing out on 401:', err);
            });
          }
        }).catch((err: unknown) => {
          console.error('Error getting user on 401:', err);
        });
        // The error will still be thrown for the mutation to handle
        break;
      case 500:
      case 503:
        // Show toast for server errors - user can't fix these
        showErrorToast(error, 'Server error. Please try again later.');
        break;
    }

    // Throw the structured error for mutation handlers to catch
    throw error;
  } catch (err) {
    // If JSON parsing fails or error already thrown, re-throw
    if (err && typeof err === 'object' && 'message' in err) {
      throw err;
    }

    // Fallback error if response body isn't valid JSON
    throw {
      message: `${response.status} ${response.statusText}`,
    } as ApiError;
  }
});
