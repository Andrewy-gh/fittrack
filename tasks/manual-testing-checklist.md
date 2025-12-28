# Client Error Handling - Manual Testing Checklist

**Project:** FitTrack Frontend Error Handling
**Created:** 2025-12-27
**Purpose:** Verify all error handling implementations are working correctly

---

## How to Use This Checklist

1. Test each item in order
2. Check off `[ ]` items as you complete them
3. Note any issues found in the "Issues Found" section at the bottom
4. Some tests require temporary code changes - examples are provided

---

## Testing Checklist

### 1. Toast Infrastructure

#### [X] 1.1 Toast Appears on Mutation Error

**What to test:** Verify Sonner toast appears when a mutation fails

**How to test:**

1. Navigate to any workout or exercise
2. Trigger a delete operation error (see code below)
3. Observe toast notification in bottom-right corner

**Expected result:** Red error toast appears with error message

**Code modification (temporary):**

Insert in `client/src/routes/_layout/workouts/-components/delete-dialog.tsx` at line 77:

```typescript
} catch (err) {
  // TESTING: Force an error
  const testError = new Error('Test error message');
  toast.error(getErrorMessage(testError, 'Failed to delete workout. Please try again.'));
  setError(getErrorMessage(testError, 'Failed to delete workout. Please try again.'));
  return; // Prevent actual deletion

  // Original code below (remove after testing):
  // const errorMessage = getErrorMessage(err, 'Failed to delete workout. Please try again.');
```

---

#### [ ] 1.2 Toast Shows User-Friendly Message (Not Raw Error)

**What to test:** Verify error messages are sanitized and user-friendly

**How to test:** Same as 1.1, but check the message content

**Expected result:**

- ✅ "Failed to delete workout. Please try again."
- ❌ NOT raw error like "Error: Network request failed at line 123"

**Code modification:** Use the same modification from 1.1

---

### 2. Request ID Debugging

#### [ ] 2.1 Request ID Available in Dev Tools

**What to test:** Verify request_id is logged to console in development

**How to test:**

1. Open browser DevTools console
2. Trigger any API error (use code below)
3. Check console for request_id log

**Expected result:** Console shows `[API Error] request_id: <uuid>` (if backend provides it)

**Code modification (temporary):**

Insert in `client/src/lib/api/client-config.ts` response interceptor (around line 20):

```typescript
client.interceptors.response.use(async (response) => {
  if (!response.ok) {
    // TESTING: Force log a test request_id
    console.log('[API Error Test] request_id:', 'test-uuid-12345');

    // Continue with original interceptor logic...
  }
  return response;
});
```

---

### 3. Authentication Errors

#### [X] 3.1 401 Redirects to Login

**What to test:** Verify 401 Unauthorized triggers redirect to login

**How to test:**

1. Temporarily modify the response interceptor to simulate 401
2. Trigger any API call
3. Check if redirect happens

**Expected result:** User is redirected to login page

**Code modification (temporary):**

Insert in `client/src/lib/api/client-config.ts` response interceptor:

```typescript
client.interceptors.response.use(async (response) => {
  // TESTING: Force 401 error
  if (response.url.includes('/workouts') || response.url.includes('/exercises')) {
    const mockResponse = new Response(
      JSON.stringify({ message: 'Unauthorized' }),
      { status: 401, headers: response.headers }
    );
    // This should trigger the 401 handling in the interceptor
    response = mockResponse;
  }

  if (!response.ok) {
    // Original interceptor logic continues...
```

**Alternative testing method:**

1. Clear authentication cookies/localStorage
2. Try to access a protected route
3. Should redirect to login

---

### 4. Route Error Handling

#### [X] 4.1 Route Error Shows Error Page (Not Crash)

**What to test:** Verify route loader errors show error component instead of crashing

**How to test:**

1. Temporarily throw an error in a route loader
2. Navigate to that route
3. Check if RouteError component appears

**Expected result:**

- Custom error page with "Something went wrong" message
- "Go back" and "Try again" buttons visible
- App doesn't crash or show blank page

**Code modification (temporary):**

Insert in `client/src/routes/_layout/workouts/index.tsx` loader (if it exists), or create a test route:

```typescript
// In the route component or loader
export const Route = createFileRoute('/_layout/workouts/')({
  component: WorkoutsPage,
  loader: async () => {
    // TESTING: Force route error
    throw new Error('Test route loader error');

    // Original loader logic below...
  },
});
```

**OR** insert at the top of the component function:

```typescript
function WorkoutsPage() {
  // TESTING: Force component error
  throw new Error('Test component render error');

  // Original component code...
}
```

---

### 5. Form Validation

#### [ ] 5.1 Form Validation Shows Inline Error

**What to test:** Verify form field validation shows red error text below field

**How to test:**

1. Open any form with validation (e.g., add/edit exercise)
2. Try to submit with empty required field
3. Check for inline error message

**Expected result:**

- Red error text appears below the field
- Submit button is disabled
- No toast appears (validation errors should be inline only)

**Code modification:** Not needed - validation should already be implemented

**Where to check:**

- Exercise name field in exercise dialogs
- Form fields should have `field.state.meta.errors` rendering

---

#### [ ] 5.2 Duplicate Name Shows Inline Error

**What to test:** Verify duplicate name error shows inline (not just toast)

**How to test:**

1. Create an exercise with name "Test Exercise"
2. Try to create/edit another exercise with the same name
3. Check for inline error

**Expected result:**

- Inline error message appears near the name field
- Toast may also appear as backup
- User can immediately fix the name without closing/reopening dialog

**Code modification (temporary):**

Insert in `client/src/routes/_layout/exercises/-components/exercise-edit-dialog.tsx` at line 77:

```typescript
} catch (err) {
  // TESTING: Simulate duplicate name error
  const duplicateError = { message: 'Exercise with this name already exists' };
  setError('Exercise with this name already exists');
  return; // Prevent normal error handling

  // Original error handling below...
```

---

### 6. Network Errors

#### [ ] 6.1 Network Offline Shows Error

**What to test:** Verify offline state triggers appropriate error handling

**How to test:**

1. Open DevTools Network tab
2. Set throttling to "Offline"
3. Try to perform any mutation (create/edit/delete)
4. Observe error handling

**Expected result:**

- Toast appears with network error message
- Error message is user-friendly (e.g., "Network error. Please check your connection.")
- App doesn't crash

**Code modification:** Not needed - just use browser DevTools

**Alternative:** Insert in any mutation onError:

```typescript
onError: (error) => {
  // TESTING: Simulate network error
  const networkError = new Error('Failed to fetch');
  toast.error('Network error. Please check your connection.');
  return;

  // Original error handling...
```

---

### 7. Server Errors

#### [X] 7.1 Server 500 Shows Generic Message

**What to test:** Verify 500 errors show user-friendly message (not raw server error)

**How to test:**

1. Temporarily simulate a 500 error in the response interceptor
2. Trigger any API call
3. Check error message

**Expected result:**

- Toast shows generic message like "Something went wrong. Please try again later."
- NOT raw server error details
- request_id logged to console (if available)

**Code modification (temporary):**

Insert in `client/src/lib/api/client-config.ts` response interceptor:

```typescript
client.interceptors.response.use(async (response) => {
  // TESTING: Force 500 error
  if (response.url.includes('/workouts') || response.url.includes('/exercises')) {
    const mockResponse = new Response(
      JSON.stringify({ message: 'Internal server error' }),
      { status: 500, headers: response.headers }
    );
    response = mockResponse;
  }

  if (!response.ok) {
    // Original interceptor logic continues...
```

---

### 8. Error Boundary

#### [X] 8.1 Error Boundary Catches Component Crash

**What to test:** Verify ErrorBoundary catches React render errors

**How to test:**

1. Temporarily throw an error in a component render
2. Check if error boundary fallback UI appears
3. Verify app doesn't completely crash

**Expected result:**

- Error boundary fallback UI appears (FullScreenError or InlineError)
- Rest of the app outside the boundary continues to work
- User can navigate away from the error

**Code modification (temporary):**

Insert in any component wrapped in ErrorBoundary (e.g., `client/src/routes/_layout/workouts/new.tsx`):

```typescript
function NewWorkoutPage() {
  // TESTING: Force component error
  if (Math.random() > -1) {
    // Always true
    throw new Error('Test React component error');
  }

  // Original component code...
}
```

**Where to check:** Components wrapped in `<ErrorBoundary>` in:

- `client/src/routes/_layout.tsx`
- Dialog components

---

## Success Criteria Verification

### [X] 9.1 No More `alert()` Calls

**What to verify:** No JavaScript `alert()` calls remain in the codebase

**How to verify:**

```bash
# Run this command in the client directory:
cd client
grep -r "alert(" src/ --include="*.tsx" --include="*.ts"
```

**Expected result:** No matches found (or only comments)

**Files that previously had alert():**

- `client/src/routes/_layout/workouts/new.tsx:73` - Should be `toast.error()` now
- `client/src/routes/_layout/workouts/$workoutId/edit.tsx:76` - Should be `toast.error()` now

---

### [X] 9.2 All Mutations Have Error Handling

**What to verify:** Every `useMutation` has either global or local error handling

**How to verify:**

```bash
# Search for useMutation calls:
cd client
grep -n "useMutation" src/ -R --include="*.tsx" --include="*.ts"
```

**Expected result:** Each mutation either:

- Relies on global `QueryClient` onError (defined in `client/src/lib/api/api.ts`)
- Has explicit `onError` callback
- Has try-catch in the component

**Files to check:**

- Delete dialogs (exercise-delete-dialog.tsx, delete-dialog.tsx)
- Edit dialogs (exercise-edit-dialog.tsx)
- New/edit pages (new.tsx, edit.tsx)

---

### [X] 9.3 Sonner Toasts Working

**What to verify:** Toaster is mounted and toasts appear correctly

**How to verify:**

1. Check that `<Toaster />` is in the DOM
2. Trigger any error from the tests above
3. Verify toast appears in bottom-right

**Expected result:** Toast system fully functional

**Where to check:** `client/src/app.tsx` or `client/src/main.tsx` should have `<Toaster />` component

---

### [X] 9.4 Route Errors Handled Gracefully

**What to verify:** Route errors show error page, not crash

**How to verify:** Use test from 4.1 above

**Expected result:** `errorComponent` is set in `__root.tsx` route

---

### [X] 9.5 Error Messages User-Friendly

**What to verify:** No raw error objects or stack traces shown to users

**How to verify:**

1. Trigger various errors from tests above
2. Check all error messages displayed to users
3. Verify they're friendly and actionable

**Expected result:**

- ✅ "Failed to delete workout. Please try again."
- ✅ "Exercise with this name already exists"
- ✅ "Network error. Please check your connection."
- ❌ NOT "Error: Cannot read property 'data' of undefined"
- ❌ NOT stack traces

---

### [ ] 9.6 Request IDs Logged for Debugging

**What to verify:** request_id is logged in dev console when available

**How to verify:** Use test from 2.1 above

**Expected result:** Console logs show request_id in development mode

---

### [X] 9.7 No TODO Error Handling Comments

**What to verify:** No "TODO: handle error" comments remain

**How to verify:**

```bash
# Search for TODO comments related to error handling:
cd client
grep -n "TODO.*error" src/ -R --include="*.tsx" --include="*.ts" -i
```

**Expected result:** No matches found

**Files that previously had TODOs:**

- `client/src/routes/_layout/workouts/-components/delete-dialog.tsx:40`
- `client/src/routes/_layout/exercises/-components/exercise-delete-dialog.tsx:40`

---

## Issues Found

**Template for reporting issues:**

```
### Issue #1: [Brief description]

**Test:** [Which test item]
**Expected:** [What should happen]
**Actual:** [What actually happened]
**Severity:** [High/Medium/Low]
**File:** [Affected file]
**Screenshot:** [If applicable]
```

---

## Testing Summary

**Total Tests:** 16 items
**Passed:** **\_ / 16
**Failed:** \_** / 16
**Blocked:** \_\_\_ / 16

**Overall Status:** [ ] All tests passed ✅ | [ ] Issues found ⚠️ | [ ] Blocked ❌

---

## Notes

- Remember to **remove all temporary testing code** after completing the checklist
- Run `git diff` to verify no testing code is committed
- Run `bun run tsc` to ensure no TypeScript errors after removing test code
- Consider running the full test suite: `bun run test` (if available)

---

**End of Manual Testing Checklist**
