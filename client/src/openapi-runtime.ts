import { OpenAPI } from '@/generated';

// Set OpenAPI base URL at runtime.
// - In development (localhost/127.0.0.1), call the Go API on port 8080 directly
// - In production, use same-origin '/api' to avoid CORS and mixed content
if (typeof window !== 'undefined') {
  const host = window.location.hostname;
  const isLocal = host === 'localhost' || host === '127.0.0.1';
  OpenAPI.BASE = isLocal ? 'http://localhost:8080/api' : '/api';
}

