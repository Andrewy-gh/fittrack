# Client Error Handling Implementation Plan

**Project:** FitTrack Frontend Error Handling
**Created:** 2025-12-27
**Status:** Planning
**Tasks:** 13 total (6 foundation, 5 component fixes, 2 polish)

---

## Overview

This plan addresses frontend error handling to align with the backend error handling overhaul completed in `server-error-handling-implementation-plan.md`.

**Backend Error Response Format:**

```json
{
  "message": "sanitized error message",
  "request_id": "uuid-v4-string" // optional
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

| Issue                         | Location                                                | Impact                             |
| ----------------------------- | ------------------------------------------------------- | ---------------------------------- |
| Sonner not in DOM             | `main.tsx`                                              | Toasts don't work                  |
| No response error interceptor | `client-config.ts`                                      | No centralized error handling      |
| String-based error detection  | `exercise-edit-dialog.tsx:66`                           | Fragile, breaks if message changes |
| `alert()` for errors          | `new.tsx:73`, `edit.tsx:76`                             | Poor UX                            |
| Missing `request_id` in type  | `types.gen.ts:67-69`                                    | Can't use for debugging            |
| No route error component      | `__root.tsx`                                            | Route errors crash app             |
| TODO: handle error            | `delete-dialog.tsx:40`, `exercise-delete-dialog.tsx:40` | Unimplemented                      |

---

## Error Display Strategy

### Decision Matrix

| Error Type                         | Display Method          | Rationale                       |
| ---------------------------------- | ----------------------- | ------------------------------- |
| **Form validation**                | Inline red text         | Immediate context, user can fix |
| **Mutation failure** (recoverable) | Toast                   | Non-blocking, user can retry    |
| **Auth error (401)**               | Redirect to login       | Session expired                 |
| **Permission error (403)**         | Toast or Dialog         | User needs to know              |
| **Not found (404)**                | Route error page        | Critical failure                |
| **Conflict (409)**                 | Inline + Toast          | e.g., duplicate name            |
| **Server error (500/503)**         | Toast + optional dialog | Network/server issue            |
| **Route load failure**             | Route error component   | Page can't render               |
| **Component render crash**         | Error boundary          | React error                     |

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

**IMPORTANT:** Mark each task as completed when you finish it. Add the end of each high level task run `bun run tsc` to make sure there are no typescript errors.

### ✅ Task 1: Enable Sonner Toast Infrastructure

**Priority:** Critical (blocks other tasks)
**Files:**

- `client/src/main.tsx` OR `client/src/app.tsx`

#### Subtasks

- [x] 1.1 Import `Toaster` from `@/components/ui/sonner`
- [x] 1.2 Add `<Toaster />` to app root (after all providers)
- [x] 1.3 Test toast works: `toast.error('test')`
- [x] 1.4 Configure toast position/duration if needed

---

### ✅ Task 2: Update API Error Type Definition (OPTIONAL)

**Priority:** Low (skip unless you want request_id tracking)
**Files:**

- `server/internal/response/types.go` (updated ErrorResponse struct)
- `server/docs/swagger.json` (regenerated)
- `client/src/client/types.gen.ts` (regenerated)

**Decision:** ✅ Implemented - request_id now available for debugging

#### Subtasks (if implementing)

- [x] 2.1 Update OpenAPI spec `ErrorResponse` to include `request_id`

```yaml
ErrorResponse:
  type: object
  properties:
    message:
      type: string
    request_id:
      type: string
```

- [x] 2.2 Regenerate client types: `bun run openapi-ts`
- [x] 2.3 Verify `ResponseErrorResponse` now has `request_id?: string`

---

### ✅ Task 3: Create Error Utilities

**Priority:** High
**Files:**

- `client/src/lib/errors.ts` (new)

#### Subtasks

- [x] 3.1 Create error utility file
- [x] 3.2 Define `ApiError` type matching backend shape

```typescript
export interface ApiError {
  message: string;
  request_id?: string;
}
```

- [x] 3.3 Create `isApiError(error: unknown): error is ApiError` type guard
- [x] 3.4 Create `getErrorMessage(error: unknown): string` extractor
- [x] 3.5 Create `showErrorToast(error: unknown, fallback?: string)` helper
- [x] 3.6 Export all utilities

---

### ✅ Task 4: Add Response Error Interceptor

**Priority:** High
**Files:**

- `client/src/lib/api/client-config.ts`

#### Subtasks

- [x] 4.1 Add response interceptor to handle errors globally

```typescript
client.interceptors.response.use(async (response) => {
  if (!response.ok) {
    const error = await response.json();
    // Transform to consistent error shape
  }
  return response;
});
```

- [x] 4.2 Consider global toast for 500/503 errors
- [x] 4.3 Handle 401 - redirect to login or refresh token
- [x] 4.4 Log request_id for debugging (console in dev, service in prod)

---

### ✅ Task 5: Add TanStack Router Error Component

**Priority:** Medium
**Files:**

- `client/src/routes/__root.tsx`
- `client/src/components/route-error.tsx` (new)

#### Subtasks

- [x] 5.1 Create `RouteError` component
  - Show error message
  - Show request_id if available (for support)
  - Provide "Go back" and "Try again" buttons
  - Include `reset` function from props
- [x] 5.2 Add `errorComponent` to root route

```typescript
export const Route = createRootRouteWithContext<RouteContext>()({
  component: RootComponent,
  errorComponent: RouteError,
});
```

- [x] 5.3 Test by throwing error in route loader

---

### ✅ Task 6: Configure React Query Global Error Handler

**Priority:** Medium
**Files:**

- `client/src/lib/api/api.ts`

#### Subtasks

- [x] 6.1 Add `onError` callback to QueryClient

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

- [x] 6.2 Decide: global mutation error toast vs per-mutation handling
- [x] 6.3 Test mutation errors show toast

---

### ✅ Task 7: Add TanStack Form Client-Side Validation

**Priority:** Medium
**Files:**

- `client/src/hooks/form.ts` (existing)
- Form field components
- Dialogs using forms

**Context:** You already have TanStack Form set up with `createFormHook` pattern. Add validators to fields.

#### How TanStack Form Validation Works

```typescript
// Field-level validation
<form.AppField
  name="exerciseName"
  validators={{
    onChange: ({ value }) => {
      if (!value.trim()) return 'Exercise name is required';
      if (value.length > 100) return 'Name must be 100 characters or less';
      return undefined;
    },
    // Run on blur for less intrusive UX
    onBlur: ({ value }) => {
      if (!value.trim()) return 'Exercise name is required';
      return undefined;
    },
  }}
>
  {(field) => (
    <>
      <input
        value={field.state.value}
        onChange={(e) => field.handleChange(e.target.value)}
        onBlur={field.handleBlur}
      />
      {field.state.meta.errors.length > 0 && (
        <span className="text-sm text-red-600">
          {field.state.meta.errors.join(', ')}
        </span>
      )}
    </>
  )}
</form.AppField>
```

#### Subtasks

- [x] 7.1 Add `validators` prop to existing form fields
- [x] 7.2 Create reusable validation functions (e.g., `required`, `maxLength`)
- [x] 7.3 Display `field.state.meta.errors` in field components
- [x] 7.4 Use `onBlur` validation for better UX (less intrusive)
- [x] 7.5 Use `form.Subscribe` with `canSubmit` to disable submit button
- [ ] 7.6 Consider Zod schema integration for complex validation

---

### ✅ Task 8: Refactor Exercise Edit Dialog Errors

**Priority:** Medium
**Files:**

- `client/src/routes/_layout/exercises/-components/exercise-edit-dialog.tsx`

#### Subtasks

- [x] 8.1 Replace string matching with proper error type checking

```typescript
// Before
if (err.message.includes('already exists')) { ... }

// After
import { isApiError, getErrorMessage } from '@/lib/errors';
if (isApiError(err) && err.message.includes('duplicate')) { ... }
// OR use status code from mutation error
```

- [x] 8.2 Use `toast.error()` for unexpected errors
- [x] 8.3 Keep inline error for validation/duplicate name
- [x] 8.4 Test: edit to duplicate name, test server error

---

### ✅ Task 9: Implement Exercise Delete Dialog Error Handling

**Priority:** Medium
**Files:**

- `client/src/routes/_layout/exercises/-components/exercise-delete-dialog.tsx`

#### Subtasks

- [x] 9.1 Remove TODO comment
- [x] 9.2 Add try-catch with proper error handling
- [x] 9.3 Show toast for delete failure
- [x] 9.4 Show inline error if makes sense
- [x] 9.5 Test: simulate delete failure

---

### ✅ Task 10: Implement Workout Delete Dialog Error Handling

**Priority:** Medium
**Files:**

- `client/src/routes/_layout/workouts/-components/delete-dialog.tsx`

#### Subtasks

- [x] 10.1 Remove TODO comment
- [x] 10.2 Add try-catch with proper error handling
- [x] 10.3 Show toast for delete failure
- [x] 10.4 Test: simulate delete failure

---

### ✅ Task 11: Refactor Workout New/Edit Alert Errors

**Priority:** Medium
**Files:**

- `client/src/routes/_layout/workouts/new.tsx`
- `client/src/routes/_layout/workouts/$workoutId/edit.tsx`

#### Subtasks

- [x] 11.1 Replace `alert(error)` with `toast.error()`
- [x] 11.2 Use `getErrorMessage()` helper
- [x] 11.3 Consider inline error for form validation (N/A - toast is appropriate for save/update errors)
- [x] 11.4 Test: create/edit workout failures

---

### ✅ Task 12: Add Error Boundary to Critical Routes

**Priority:** Low
**Files:**

- `client/src/routes/_layout.tsx`
- `client/src/components/error-boundary.tsx`

#### Subtasks

- [x] 12.1 Wrap Suspense boundaries with ErrorBoundary
- [x] 12.2 Create route-specific fallback UIs
- [x] 12.3 Test: force component crash

---

### ✅ Task 13: Consider Alert Dialog for Destructive Failures

**Priority:** Low
**Files:**

- `client/src/routes/_layout/exercises/-components/exercise-delete-dialog.tsx` (reviewed)
- `client/src/routes/_layout/workouts/-components/delete-dialog.tsx` (reviewed)

**Decision:** ❌ Additional AlertDialog is NOT needed - current implementation is optimal

#### Evaluation Summary

**Current Implementation:**
Both delete dialogs already use a hybrid approach:
- Toast notification for quick feedback
- Inline error displayed in the confirmation AlertDialog (exercise-delete-dialog.tsx:92, delete-dialog.tsx:95)
- Dialog stays open for immediate retry
- User can cancel if needed

**Why Additional "Delete Failed" Dialog is NOT Recommended:**
- ❌ Would create confusing dialog stacking (confirmation → error → retry flow)
- ❌ Current inline error + toast approach is clearer and more contextual
- ❌ Retry is already available (user clicks Delete button again)
- ❌ Adds unnecessary code complexity
- ✅ Current implementation provides all needed functionality

#### Subtasks

- [x] 13.1 Evaluate if delete failures need AlertDialog
- [N/A] 13.2 If yes, implement pattern for "Delete failed" dialog - Not needed per evaluation
- [N/A] 13.3 Include retry option in dialog - Already implemented via inline error approach

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
3. **Task 2**: Update API error type (OPTIONAL - skip if not using request_id)

### Phase 2: Global Handlers

4. **Task 4**: Response error interceptor
5. **Task 6**: React Query global error handler
6. **Task 5**: TanStack Router error component

### Phase 3: Validation & Component Fixes

7. **Task 7**: TanStack Form client-side validation
8. **Task 8**: Exercise edit dialog errors
9. **Task 9**: Exercise delete dialog errors
10. **Task 10**: Workout delete dialog errors
11. **Task 11**: Workout new/edit alert errors

### Phase 4: Polish

12. **Task 12**: Error boundaries
13. **Task 13**: Alert dialog consideration

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

## Questions Resolved

### 1. Request ID - Is It Needed?

**Short answer: Nice-to-have, not must-have.**

**When request_id is useful:**

- Correlating user-reported errors with backend logs
- Support tickets: "Error with ID abc-123"
- Debugging production issues across client/server boundary

**When it's NOT needed:**

- You don't have centralized logging (Sentry, DataDog, etc.)
- Console logs in deployment are your only debugging tool
- Small/personal project without support workflow

**Does NOT including it cause data drift?**
No. It's metadata for debugging, not data. Frontend doesn't need to store or track it.

**Recommendation for your case:**

- Skip request_id display to users
- Log it to console in dev for debugging
- Add it later if you add error monitoring service (Sentry, etc.)
- Don't update OpenAPI spec unless you want to use it

**Complexity cost:** Minimal if added, but adds noise if unused.

### 2. Toast Position

**Answer:** Bottom-right

### 3. Request ID Display

**Answer:** Dev only (console.log), hide in prod

### 4. 401 Handling

**Answer:** Follow StackFrame auth flow - likely auto-redirect to login

---

## Resolved Question: Global vs Per-Mutation Error Handling

**Pattern: Global default + override**

```typescript
// lib/api/api.ts - Global default
const queryClient = new QueryClient({
  defaultOptions: {
    mutations: {
      onError: (error) => {
        // Default: show toast for all mutations
        toast.error(getErrorMessage(error));
      },
    },
  },
});

// In component - Override for specific case
const mutation = useMutation({
  ...saveMutation(),
  onError: (error) => {
    // Override: show inline error instead of toast
    setError(getErrorMessage(error));
  },
});
```

**When to override (use per-mutation):**

- Form validation errors → show inline instead of toast
- Duplicate name (409) → show inline next to input
- Custom retry logic needed
- Need error in local state for UI

**When to use global (toast):**

- Network errors
- 500s
- Unexpected failures
- Delete operations (no form to show inline error)

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
- `client/src/lib/validation.ts` - Reusable validation functions (optional)
- `client/src/components/route-error.tsx` - Route error component

### Modified Files

- `client/src/main.tsx` OR `client/src/app.tsx` - Add Toaster
- `client/src/lib/api/client-config.ts` - Response interceptor
- `client/src/lib/api/api.ts` - QueryClient config
- `client/src/routes/__root.tsx` - Error component
- `client/src/hooks/form.ts` - Add validation to field components
- `client/src/components/form/*.tsx` - Add error display to form fields
- `client/src/routes/_layout/exercises/-components/exercise-edit-dialog.tsx`
- `client/src/routes/_layout/exercises/-components/exercise-delete-dialog.tsx`
- `client/src/routes/_layout/workouts/-components/delete-dialog.tsx`
- `client/src/routes/_layout/workouts/new.tsx`
- `client/src/routes/_layout/workouts/$workoutId/edit.tsx`

### Optional (if using request_id)

- `server/api/openapi.yaml` - Add request_id to ErrorResponse

---

**End of Implementation Plan**
