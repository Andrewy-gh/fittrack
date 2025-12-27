# Client Error Handling Implementation Plan

**Project:** FitTrack Frontend Error Handling
**Created:** 2025-12-27
**Status:** Planning

---

## Overview

This plan addresses frontend error handling to align with the backend error handling overhaul completed in `server-error-handling-implementation-plan.md`.

**Backend Error Response Format:**
```json
{
  "message": "sanitized error message",
  "request_id": "uuid-v4-string"  // optional
}
```

**Key Goals:**
1. Properly consume backend error shape `{message, request_id?}`
2. Implement tiered error display strategy (toast/dialog/page)
3. Add global error handling infrastructure
4. Remove fragile string-based error detection
5. Enable Sonner toasts (installed but unused)

---

## Current State Analysis

### What Exists
- **Sonner** - Installed, configured (`sonner.tsx`), but NOT added to app root
- **ErrorBoundary** - Component exists (`error-boundary.tsx`), never used
- **Alert/Dialog** - Available for prominent errors
- **ContributionGraphError** - Example of inline component error
- **Generated API types** - Missing `request_id` field

### Problems Found
| Issue | Location | Impact |
|-------|----------|--------|
| Sonner not in DOM | `main.tsx` | Toasts don't work |
| No response error interceptor | `client-config.ts` | No centralized error handling |
| String-based error detection | `exercise-edit-dialog.tsx:66` | Fragile, breaks if message changes |
| `alert()` for errors | `new.tsx:73`, `edit.tsx:76` | Poor UX |
| Missing `request_id` in type | `types.gen.ts:67-69` | Can't use for debugging |
| No route error component | `__root.tsx` | Route errors crash app |
| TODO: handle error | `delete-dialog.tsx:40`, `exercise-delete-dialog.tsx:40` | Unimplemented |

---

## Error Display Strategy

### Decision Matrix

| Error Type | Display Method | Rationale |
|------------|----------------|-----------|
| **Form validation** | Inline red text | Immediate context, user can fix |
| **Mutation failure** (recoverable) | Toast | Non-blocking, user can retry |
| **Auth error (401)** | Redirect to login | Session expired |
| **Permission error (403)** | Toast or Dialog | User needs to know |
| **Not found (404)** | Route error page | Critical failure |
| **Conflict (409)** | Inline + Toast | e.g., duplicate name |
| **Server error (500/503)** | Toast + optional dialog | Network/server issue |
| **Route load failure** | Route error component | Page can't render |
| **Component render crash** | Error boundary | React error |

### Recommended Approach

**Tier 1: Toast (Sonner)** - Default for most API errors
- Quick, non-blocking feedback
- Auto-dismisses
- User can continue working
- Best for: mutations, transient failures, network errors

**Tier 2: Inline Error** - Form/component context
- Shows next to relevant input
- Doesn't interrupt flow
- Best for: validation, field-specific errors

**Tier 3: Alert Dialog** - Critical/blocking errors
- Requires user acknowledgment
- Best for: destructive action failures, data loss warnings

**Tier 4: Error Page** - Unrecoverable route errors
- TanStack Router `errorComponent`
- Best for: page load failures, 404s

**Tier 5: Error Boundary** - React crashes
- Catches render errors
- Shows fallback UI
- Best for: unexpected component crashes

---

## Implementation Tasks

### Task 1: Enable Sonner Toast Infrastructure

**Priority:** Critical (blocks other tasks)
**Files:**
- `client/src/main.tsx` OR `client/src/app.tsx`

#### Subtasks
- [ ] 1.1 Import `Toaster` from `@/components/ui/sonner`
- [ ] 1.2 Add `<Toaster />` to app root (after all providers)
- [ ] 1.3 Test toast works: `toast.error('test')`
- [ ] 1.4 Configure toast position/duration if needed

---

### Task 2: Update API Error Type Definition

**Priority:** High
**Files:**
- `server/api/openapi.yaml` (source of truth)
- `client/src/client/types.gen.ts` (regenerated)

#### Subtasks
- [ ] 2.1 Update OpenAPI spec `ErrorResponse` to include `request_id`
```yaml
ErrorResponse:
  type: object
  properties:
    message:
      type: string
    request_id:
      type: string
```
- [ ] 2.2 Regenerate client types: `bun run openapi-ts`
- [ ] 2.3 Verify `ResponseErrorResponse` now has `request_id?: string`

---

### Task 3: Create Error Utilities

**Priority:** High
**Files:**
- `client/src/lib/errors.ts` (new)

#### Subtasks
- [ ] 3.1 Create error utility file
- [ ] 3.2 Define `ApiError` type matching backend shape
```typescript
export interface ApiError {
  message: string;
  request_id?: string;
}
```
- [ ] 3.3 Create `isApiError(error: unknown): error is ApiError` type guard
- [ ] 3.4 Create `getErrorMessage(error: unknown): string` extractor
- [ ] 3.5 Create `showErrorToast(error: unknown, fallback?: string)` helper
- [ ] 3.6 Export all utilities

---

### Task 4: Add Response Error Interceptor

**Priority:** High
**Files:**
- `client/src/lib/api/client-config.ts`

#### Subtasks
- [ ] 4.1 Add response interceptor to handle errors globally
```typescript
client.interceptors.response.use(async (response) => {
  if (!response.ok) {
    const error = await response.json();
    // Transform to consistent error shape
  }
  return response;
});
```
- [ ] 4.2 Consider global toast for 500/503 errors
- [ ] 4.3 Handle 401 - redirect to login or refresh token
- [ ] 4.4 Log request_id for debugging (console in dev, service in prod)

---

### Task 5: Add TanStack Router Error Component

**Priority:** Medium
**Files:**
- `client/src/routes/__root.tsx`
- `client/src/components/route-error.tsx` (new)

#### Subtasks
- [ ] 5.1 Create `RouteError` component
  - Show error message
  - Show request_id if available (for support)
  - Provide "Go back" and "Try again" buttons
  - Include `reset` function from props
- [ ] 5.2 Add `errorComponent` to root route
```typescript
export const Route = createRootRouteWithContext<RouteContext>()({
  component: RootComponent,
  errorComponent: RouteError,
});
```
- [ ] 5.3 Test by throwing error in route loader

---

### Task 6: Configure React Query Global Error Handler

**Priority:** Medium
**Files:**
- `client/src/lib/api/api.ts`

#### Subtasks
- [ ] 6.1 Add `onError` callback to QueryClient
```typescript
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
    },
    mutations: {
      onError: (error) => {
        showErrorToast(error);
      },
    },
  },
});
```
- [ ] 6.2 Decide: global mutation error toast vs per-mutation handling
- [ ] 6.3 Test mutation errors show toast

---

### Task 7: Refactor Exercise Edit Dialog Errors

**Priority:** Medium
**Files:**
- `client/src/routes/_layout/exercises/-components/exercise-edit-dialog.tsx`

#### Subtasks
- [ ] 7.1 Replace string matching with proper error type checking
```typescript
// Before
if (err.message.includes('already exists')) { ... }

// After
import { isApiError, getErrorMessage } from '@/lib/errors';
if (isApiError(err) && err.message.includes('duplicate')) { ... }
// OR use status code from mutation error
```
- [ ] 7.2 Use `toast.error()` for unexpected errors
- [ ] 7.3 Keep inline error for validation/duplicate name
- [ ] 7.4 Test: edit to duplicate name, test server error

---

### Task 8: Implement Exercise Delete Dialog Error Handling

**Priority:** Medium
**Files:**
- `client/src/routes/_layout/exercises/-components/exercise-delete-dialog.tsx`

#### Subtasks
- [ ] 8.1 Remove TODO comment
- [ ] 8.2 Add try-catch with proper error handling
- [ ] 8.3 Show toast for delete failure
- [ ] 8.4 Show inline error if makes sense
- [ ] 8.5 Test: simulate delete failure

---

### Task 9: Implement Workout Delete Dialog Error Handling

**Priority:** Medium
**Files:**
- `client/src/routes/_layout/workouts/-components/delete-dialog.tsx`

#### Subtasks
- [ ] 9.1 Remove TODO comment
- [ ] 9.2 Add try-catch with proper error handling
- [ ] 9.3 Show toast for delete failure
- [ ] 9.4 Test: simulate delete failure

---

### Task 10: Refactor Workout New/Edit Alert Errors

**Priority:** Medium
**Files:**
- `client/src/routes/_layout/workouts/new.tsx`
- `client/src/routes/_layout/workouts/$workoutId/edit.tsx`

#### Subtasks
- [ ] 10.1 Replace `alert(error)` with `toast.error()`
- [ ] 10.2 Use `getErrorMessage()` helper
- [ ] 10.3 Consider inline error for form validation
- [ ] 10.4 Test: create/edit workout failures

---

### Task 11: Add Error Boundary to Critical Routes

**Priority:** Low
**Files:**
- `client/src/routes/_layout.tsx`
- `client/src/components/error-boundary.tsx`

#### Subtasks
- [ ] 11.1 Wrap Suspense boundaries with ErrorBoundary
- [ ] 11.2 Create route-specific fallback UIs
- [ ] 11.3 Test: force component crash

---

### Task 12: Consider Alert Dialog for Destructive Failures

**Priority:** Low
**Files:**
- Various delete flows

#### Subtasks
- [ ] 12.1 Evaluate if delete failures need AlertDialog
- [ ] 12.2 If yes, implement pattern for "Delete failed" dialog
- [ ] 12.3 Include retry option in dialog

---

## Trade-offs Analysis

### Toast (Sonner) Pros/Cons
**Pros:**
- Non-blocking - user can continue
- Consistent location (corner)
- Auto-dismisses - no cleanup
- Low code complexity

**Cons:**
- Can be missed if user not looking
- Limited information space
- Queue management for multiple errors

### Alert Dialog Pros/Cons
**Pros:**
- Guaranteed visibility
- Can include more detail
- Forces acknowledgment

**Cons:**
- Blocks user flow
- Jarring UX for minor errors
- More code per error case

### Inline Error Pros/Cons
**Pros:**
- Immediate context
- Clear association with action
- Doesn't interrupt flow

**Cons:**
- Requires per-component implementation
- Layout shifts
- More boilerplate

### Route Error Page Pros/Cons
**Pros:**
- Full page for serious errors
- Can include rich recovery options
- Consistent pattern via router

**Cons:**
- Complete flow interruption
- User loses current state
- Only for route-level failures

### Recommendation

**Default to Toast** for:
- Network errors
- Server errors (500s)
- Most mutation failures
- Transient issues

**Use Inline** for:
- Form validation
- Specific field errors
- Duplicate/conflict where user must fix input

**Use Dialog** sparingly:
- Data loss warnings
- Session expiration (with redirect)
- Critical failures needing explanation

**Route Error** for:
- Page load failures
- 404s
- Auth redirects

---

## Implementation Order

### Phase 1: Foundation
1. **Task 1**: Enable Sonner (FIRST - unblocks everything)
2. **Task 3**: Create error utilities
3. **Task 2**: Update API error type

### Phase 2: Global Handlers
4. **Task 4**: Response error interceptor
5. **Task 6**: React Query global error handler
6. **Task 5**: TanStack Router error component

### Phase 3: Component Fixes
7. **Task 7**: Exercise edit dialog
8. **Task 8**: Exercise delete dialog
9. **Task 9**: Workout delete dialog
10. **Task 10**: Workout new/edit alerts

### Phase 4: Polish
11. **Task 11**: Error boundaries
12. **Task 12**: Alert dialog consideration

---

## Testing Checklist

- [ ] Toast appears on mutation error
- [ ] Toast shows user-friendly message (not raw error)
- [ ] Request ID available in dev tools for debugging
- [ ] 401 redirects to login
- [ ] Route error shows error page not crash
- [ ] Form validation shows inline error
- [ ] Duplicate name shows appropriate message
- [ ] Network offline shows error
- [ ] Server 500 shows generic message
- [ ] Error boundary catches component crash

---

## Questions to Resolve

1. **Global mutation toast vs per-mutation?**
   - Option A: All mutations show toast on error (simpler)
   - Option B: Per-mutation handlers (more control)
   - Recommendation: Global default + override when needed

2. **Toast position?**
   - Top-right is common
   - Bottom-right is less intrusive
   - Recommendation: Bottom-right to match mobile patterns

3. **Request ID display to user?**
   - Option A: Always show (helps support)
   - Option B: Only in dev mode
   - Option C: Behind "details" expansion
   - Recommendation: Show in dev, hide in prod (log to console)

4. **Should 401 auto-redirect or show dialog?**
   - Depends on auth strategy
   - If token refresh exists, try that first
   - Recommendation: Check with auth flow

---

## Success Criteria

- [ ] No more `alert()` calls in codebase
- [ ] All mutations have error handling
- [ ] Sonner toasts working
- [ ] Route errors handled gracefully
- [ ] Error messages user-friendly (no raw errors)
- [ ] Request IDs logged for debugging
- [ ] No TODO error handling comments remaining

---

## Files to Create/Modify

### New Files
- `client/src/lib/errors.ts` - Error utilities
- `client/src/components/route-error.tsx` - Route error component

### Modified Files
- `client/src/main.tsx` OR `client/src/app.tsx` - Add Toaster
- `client/src/lib/api/client-config.ts` - Response interceptor
- `client/src/lib/api/api.ts` - QueryClient config
- `client/src/routes/__root.tsx` - Error component
- `client/src/routes/_layout/exercises/-components/exercise-edit-dialog.tsx`
- `client/src/routes/_layout/exercises/-components/exercise-delete-dialog.tsx`
- `client/src/routes/_layout/workouts/-components/delete-dialog.tsx`
- `client/src/routes/_layout/workouts/new.tsx`
- `client/src/routes/_layout/workouts/$workoutId/edit.tsx`
- `server/api/openapi.yaml` - Add request_id to ErrorResponse

---

**End of Implementation Plan**
