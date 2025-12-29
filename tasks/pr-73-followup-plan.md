# PR #73 Follow-up Implementation Plan

**Status:** Approved
**Date:** 2025-12-28
**Acknowledgment:** All recommendations in this plan have been reviewed and approved for implementation.

---

## Summary of Research Findings

### 1. TKDodo Blog: `mutate` vs `mutateAsync`

- **Use `mutate` in almost all cases** - React Query handles errors internally
- Only use `mutateAsync` for concurrent mutations or dependent mutation chains
- Callbacks on `mutate` won't fire if component unmounts before mutation finishes
- Recommendation: Keep `mutateAsync` in delete-dialog.tsx (needs await for navigation), consider `mutate` elsewhere

### 2. Backend 401 Analysis

- All API endpoints require authentication and can return 401
- No "non-critical" endpoints that might 401 in the background
- Current behavior is acceptable: 401 = invalid/expired token = sign out

### 3. QueryClient Error Handling (Better Approach)

**Current**: Using `defaultOptions.mutations.onError` - gets overwritten by local `onError`
**Recommended**: Use `MutationCache` with `onError` - always runs, never overwritten

```typescript
// Current approach (can be overwritten)
new QueryClient({
  defaultOptions: {
    mutations: { onError: showErrorToast },
  },
});

// Better approach (always runs)
new QueryClient({
  mutationCache: new MutationCache({
    onError: (error, _variables, _context, mutation) => {
      // Check meta to skip specific mutations
      if (mutation.options.meta?.skipGlobalErrorHandler) return;
      showErrorToast(error);
    },
  }),
});
```

### 4. TanStack Form Validation Modes

| Mode                      | When Fires                 | Use Case                          |
| ------------------------- | -------------------------- | --------------------------------- |
| `onChange`                | Every keystroke            | Real-time feedback, short fields  |
| `onBlur`                  | When field loses focus     | Longer inputs, less intrusive     |
| `onChangeAsyncDebounceMs` | Debounced async validation | API calls (username availability) |

**Form State:**

- `isValid`: Form-level - true if NO errors across all fields
- `canSubmit`: Form-level - false when invalid AND touched
- Field validation only affects `canSubmit` when THAT field has been touched

**Recommendation**: Use `onChange` for reps field (number input, immediate feedback)

### 5. E2E Testing with Stack Auth (Research Complete)

**Stack Auth uses `@stackframe/react` with cookie-based token storage.**

**No official E2E testing docs found.** Options:

1. **Unit tests**: Mock `useUser` hook with Vitest's `vi.mock('@stackframe/react')`
2. **E2E with demo mode**: Test unauthenticated flow (already works)
3. **E2E with auth**: Would require Playwright's `storageState` to inject cookies manually

**Recommended approach**:

- Focus on unit tests for components (mock `useUser`)
- Use demo mode for E2E flows
- Defer full authenticated E2E until Stack Auth provides testing utilities

---

## Implementation Tasks

### Task 1: Refactor QueryClient to use MutationCache

**File:** `client/src/lib/api/api.ts`

**Changes:**

- Import `MutationCache` from `@tanstack/react-query`
- Replace `defaultOptions.mutations.onError` with `mutationCache` callback
- Add `meta.skipGlobalErrorHandler` check for silent mutations
- Remove 5xx toast from client-config.ts (MutationCache handles it)

**Benefits:**

- Global error handler always runs (not overwritten)
- Cleaner "silent" pattern using `meta` instead of empty `onError`
- Removes duplicate toast issue for 5xx errors

### Task 2: Update Silent Mutations Pattern

**Files:**

- `client/src/lib/api/workouts.ts`
- `client/src/lib/api/exercises.ts`
- `client/src/lib/demo-data/query-options.ts`

**Changes:**

- Remove `onError: () => {}` from silent mutations
- Add `meta: { skipGlobalErrorHandler: true }` instead
- Rename `useDeleteWorkoutMutationSilent` → `useDeleteWorkoutMutation` (remove "Silent" suffix)

### Task 3: Update Response Interceptor (401 toast + remove 5xx toast)

**File:** `client/src/lib/api/client-config.ts`

**Changes:**

- Remove `showErrorToast` call for 500/503 responses (lines 58-62) - MutationCache handles it
- Add toast notification before 401 sign-out: `toast.error('Session expired. Please log in again.')`
- Keep 401 sign-out logic
- Keep error parsing and throwing

### Task 4: Fix Validation in add-set-dialog.tsx

**File:** `client/src/routes/_layout/workouts/-components/add-set-dialog.tsx`

**Changes:**

- Change validation from `onBlur` to `onChange` for immediate feedback
- Remove `required` validator (minValue(1) handles it for numbers)
- Remove `repsIsTouched` from disable logic
- Use `!canSubmit || !isValid` only (or add check for reps value > 0)

### Task 5: Add Unit Tests for errors.ts

**File:** `client/src/lib/errors.test.ts` (new file)

**Tests:**

```typescript
describe('isApiError', () => {
  it('returns true for valid ApiError objects');
  it('returns false for null/undefined');
  it('returns false for Error instances without message property structure');
});

describe('getErrorMessage', () => {
  it('extracts message from ApiError');
  it('extracts message from Error instance');
  it('returns string as-is');
  it('returns fallback for unknown types');
  it('returns fallback for empty string');
});

describe('showErrorToast', () => {
  it('calls toast.error with extracted message');
  it('logs request_id when available');
});
```

### Task 6: Add Unit Tests for validation.ts

**File:** `client/src/lib/validation.test.ts` (new file)

**Tests:**

```typescript
describe('required', () => {
  it('returns error for empty string');
  it('returns undefined for non-empty string');
  it('returns error for null/undefined');
  it('returns undefined for numbers (including 0)');
});

describe('minValue/maxValue', () => {
  it('returns error when value below/above threshold');
  it('returns undefined when value within range');
  it('handles non-number inputs gracefully');
});

describe('compose', () => {
  it('returns first error from multiple validators');
  it('returns undefined when all pass');
});
```

### Task 7: Add Unit Test for ErrorBoundary

**File:** `client/src/components/error-boundary.test.tsx` (new file)

**Tests:**

```typescript
describe('ErrorBoundary', () => {
  it('renders children when no error');
  it('renders fallback when child throws');
  it('logs error to console');
});

describe('FullScreenErrorFallback', () => {
  it('renders error message');
  it('calls onAction when button clicked');
  it('reloads page when no onAction provided');
});
```

### Task 8: Delete Documentation File

**File:** `client/docs/tanstack-handling-errors.md`

**Action:** Delete file (copy-pasted TanStack docs, adds bloat)

### Task 9: Update CLAUDE.local.md

**File:** `client/CLAUDE.local.md`

**Changes:**

- Update to reflect MutationCache pattern
- Update silent mutation pattern (meta instead of onError)

---

## Files to Modify

| File                                                                | Action                         |
| ------------------------------------------------------------------- | ------------------------------ |
| `client/src/lib/api/api.ts`                                         | Refactor to MutationCache      |
| `client/src/lib/api/client-config.ts`                               | Remove 5xx toast               |
| `client/src/lib/api/workouts.ts`                                    | Update silent mutation pattern |
| `client/src/lib/api/exercises.ts`                                   | Update silent mutation pattern |
| `client/src/lib/demo-data/query-options.ts`                         | Update silent mutation pattern |
| `client/src/routes/_layout/workouts/-components/add-set-dialog.tsx` | Fix validation                 |
| `client/src/lib/errors.test.ts`                                     | Create (new)                   |
| `client/src/lib/validation.test.ts`                                 | Create (new)                   |
| `client/src/components/error-boundary.test.tsx`                     | Create (new)                   |
| `client/docs/tanstack-handling-errors.md`                           | Delete                         |
| `client/CLAUDE.local.md`                                            | Update documentation           |

---

## Execution Order

1. [x] Task 1 + Task 3 (QueryClient refactor + remove interceptor toast)
2. [x] Task 2 (Update silent mutations to use meta pattern)
3. [x] Task 4 (Fix add-set-dialog validation)
4. [x] Task 5 + Task 6 + Task 7 (Unit tests - parallel)
5. [ ] Task 8 (Delete docs file)
6. [ ] Task 9 (Update CLAUDE.local.md)
7. [ ] Run tests, verify build
