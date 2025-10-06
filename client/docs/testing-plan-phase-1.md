# Testing Plan for Phase 1 + 1.5: `/workouts/new` Route

**Created**: 2025-10-06
**Status**: Ready to Implement
**Scope**: Automated testing for unified route implementation (Phase 1 + 1.5)

---

## Executive Summary

This document outlines the automated testing strategy for verifying the `/workouts/new` route implementation with demo and authenticated user support, including the critical RecentSets component fix.

---

## Testing Infrastructure

### Available Tools ✅

1. **Vitest** v3.0.5
   - Unit and integration testing framework
   - React component testing support
   - Fast, modern test runner

2. **React Testing Library** v16.2.0
   - Component testing utilities
   - User-centric testing approach
   - Best practices for React testing

3. **JSDOM** v26.0.0
   - Browser environment simulation
   - DOM API support for tests

4. **TypeScript Compiler** v5.9.2
   - Static type checking
   - Compile-time error detection

---

## Test Strategy

### Test Pyramid Approach

```
           ┌─────────────┐
           │   E2E (0)   │ ← Future: Playwright tests
           ├─────────────┤
           │Integration  │ ← Route & component integration
           │   Tests     │
           ├─────────────┤
           │   Unit      │ ← Factory functions, utilities
           │   Tests     │
           └─────────────┘
```

### Coverage Goals

- **Factory Functions**: 100% coverage
- **Critical Paths**: 80%+ coverage
  - Workout creation flow
  - RecentSets rendering
  - Draft persistence
- **Edge Cases**: All covered
  - Null user handling
  - Empty data states
  - Type safety

---

## Test Files to Create

### 1. Vitest Configuration
**File**: `client/vitest.config.ts`

**Purpose**: Configure Vitest for React component testing

**Configuration**:
```typescript
import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/test/setup.ts'],
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
})
```

---

### 2. Test Setup File
**File**: `client/src/test/setup.ts`

**Purpose**: Global test configuration and mocks

**Setup**:
- Mock `@stackframe/react` auth hooks
- Mock `@tanstack/react-router` navigation
- Configure testing-library cleanup
- Mock localStorage

---

### 3. Factory Functions Unit Tests
**File**: `client/src/lib/api/__tests__/unified-query-options.test.ts`

**Purpose**: Verify factory functions return correct query options based on user state

**Test Cases**:
```typescript
describe('getExercisesQueryOptions', () => {
  it('returns API query options when user is authenticated')
  it('returns demo query options when user is null')
})

describe('getRecentSetsQueryOptions', () => {
  it('returns API query options with exerciseId when user is authenticated')
  it('returns demo query options with exerciseId when user is null')
})

describe('getWorkoutsQueryOptions', () => {
  it('returns API query options when user is authenticated')
  it('returns demo query options when user is null')
})

describe('getWorkoutByIdQueryOptions', () => {
  it('returns API query options with workoutId when user is authenticated')
  it('returns demo query options with workoutId when user is null')
})

describe('getWorkoutsFocusQueryOptions', () => {
  it('returns API query options when user is authenticated')
  it('returns demo query options when user is null')
})
```

**Approach**:
- Mock both API and demo query option functions
- Test return values match expected query options
- Verify exerciseId/workoutId are passed correctly

---

### 4. RecentSets Component Tests
**File**: `client/src/routes/workouts/-components/__tests__/recent-sets-display.test.tsx`

**Purpose**: Verify RecentSets component works with both demo and auth users

**Test Cases**:
```typescript
describe('RecentSets', () => {
  describe('with null exerciseId', () => {
    it('renders nothing when exerciseId is null')
  })

  describe('with demo user (null)', () => {
    it('renders loading state initially')
    it('renders recent sets from demo data')
    it('renders empty state when no recent sets exist')
  })

  describe('with authenticated user', () => {
    it('renders loading state initially')
    it('renders recent sets from API data')
    it('renders empty state when no recent sets exist')
  })
})

describe('RecentSetsDisplay', () => {
  it('displays recent sets with correct formatting')
  it('handles exercise switching correctly')
  it('uses correct query options based on user state')
})
```

**Mocking Strategy**:
- Mock `useSuspenseQuery` to return test data
- Mock `getRecentSetsQueryOptions` factory function
- Test both demo and auth code paths

---

### 5. Route Integration Tests
**File**: `client/src/routes/workouts/__tests__/new.test.tsx`

**Purpose**: Verify complete workflow for workout creation in both modes

**Test Cases**:
```typescript
describe('/workouts/new route', () => {
  describe('Demo mode (unauthenticated)', () => {
    it('loads route with demo exercises')
    it('displays form with demo data')
    it('uses demo mutation on save')
    it('passes null user to RecentSets component')
    it('persists draft to localStorage without userId')
  })

  describe('Authenticated mode', () => {
    it('loads route with API exercises')
    it('displays form with API data')
    it('uses API mutation on save')
    it('passes user to RecentSets component')
    it('persists draft to localStorage with userId')
  })

  describe('Loader behavior', () => {
    it('initializes demo data when user is null')
    it('loads API data when user is authenticated')
    it('pre-fetches exercises and focus values')
  })
})

describe('WorkoutTracker component', () => {
  describe('Mutation selection', () => {
    it('uses demo mutation when user is null')
    it('uses API mutation when user is authenticated')
  })

  describe('Draft persistence', () => {
    it('saves draft with user?.id')
    it('clears draft with user?.id on save')
    it('handles undefined userId correctly')
  })
})
```

**Mocking Strategy**:
- Mock TanStack Query hooks (`useSuspenseQuery`, `useMutation`)
- Mock route context (`useRouteContext`)
- Mock localStorage functions
- Mock form submission

---

### 6. TypeScript Compilation Test
**Script**: `bun run tsc`

**Purpose**: Verify no TypeScript errors exist

**Success Criteria**: Exit code 0, no compilation errors

---

## Test Execution Plan

### Phase 1: Setup (30 minutes)
1. Create `vitest.config.ts`
2. Create `client/src/test/setup.ts`
3. Create test directory structure:
   ```
   client/src/
   ├── lib/
   │   └── api/
   │       └── __tests__/
   │           └── unified-query-options.test.ts
   └── routes/
       └── workouts/
           ├── __tests__/
           │   └── new.test.tsx
           └── -components/
               └── __tests__/
                   └── recent-sets-display.test.tsx
   ```

### Phase 2: Unit Tests (45 minutes)
1. Write factory function tests
2. Run tests: `bun run test unified-query-options`
3. Verify 100% coverage for factory functions
4. Fix any issues

### Phase 3: Component Tests (60 minutes)
1. Write RecentSets component tests
2. Run tests: `bun run test recent-sets-display`
3. Verify demo and auth modes work
4. Fix any issues

### Phase 4: Integration Tests (90 minutes)
1. Write route integration tests
2. Run tests: `bun run test new.test`
3. Verify complete workflows
4. Fix any issues

### Phase 5: Type Checking (5 minutes)
1. Run TypeScript compiler: `bun run tsc`
2. Verify 0 errors
3. Fix any type issues

### Phase 6: Full Test Suite (15 minutes)
1. Run all tests: `bun run test`
2. Verify all tests pass
3. Review test coverage report
4. Document results

---

## Success Criteria

### Functional Requirements ✅
- [ ] All factory functions tested (5/5 functions)
- [ ] RecentSets component tested (demo + auth)
- [ ] Route loader tested (demo + auth)
- [ ] Mutation selection tested (demo + auth)
- [ ] Draft persistence tested (both user types)
- [ ] TypeScript compilation: 0 errors

### Test Quality Requirements ✅
- [ ] Unit tests: 100% coverage for factory functions
- [ ] Integration tests: 80%+ coverage for critical paths
- [ ] All tests pass consistently
- [ ] No test flakiness
- [ ] Clear test descriptions
- [ ] Proper mocking strategy

### Code Quality Requirements ✅
- [ ] Tests follow React Testing Library best practices
- [ ] Tests are maintainable and readable
- [ ] Tests document expected behavior
- [ ] Tests catch regressions

---

## Test Running Commands

```bash
# Run all tests
bun run test

# Run specific test file
bun run test unified-query-options

# Run tests in watch mode
bun run test --watch

# Run tests with coverage
bun run test --coverage

# Type check
bun run tsc
```

---

## Mocking Strategy

### Required Mocks

1. **@stackframe/react**
   ```typescript
   vi.mock('@stackframe/react', () => ({
     useUser: vi.fn(),
   }))
   ```

2. **@tanstack/react-router**
   ```typescript
   vi.mock('@tanstack/react-router', () => ({
     useRouteContext: vi.fn(),
     Route: { useRouteContext: vi.fn() },
   }))
   ```

3. **@tanstack/react-query**
   ```typescript
   vi.mock('@tanstack/react-query', () => ({
     useSuspenseQuery: vi.fn(),
     useMutation: vi.fn(),
   }))
   ```

4. **localStorage**
   ```typescript
   const localStorageMock = {
     getItem: vi.fn(),
     setItem: vi.fn(),
     removeItem: vi.fn(),
     clear: vi.fn(),
   }
   global.localStorage = localStorageMock as any
   ```

---

## Expected Test Results

### Factory Function Tests
```
✓ getExercisesQueryOptions (5 tests)
✓ getRecentSetsQueryOptions (2 tests)
✓ getWorkoutsQueryOptions (2 tests)
✓ getWorkoutByIdQueryOptions (2 tests)
✓ getWorkoutsFocusQueryOptions (2 tests)

Tests: 13 passed (13 total)
```

### RecentSets Component Tests
```
✓ RecentSets (7 tests)
✓ RecentSetsDisplay (3 tests)

Tests: 10 passed (10 total)
```

### Route Integration Tests
```
✓ /workouts/new route (10 tests)
✓ WorkoutTracker component (6 tests)

Tests: 16 passed (16 total)
```

### TypeScript Compilation
```
$ bun run tsc
✓ No errors found
```

---

## Timeline Estimate

| Phase | Task | Time | Cumulative |
|-------|------|------|------------|
| 1 | Setup test infrastructure | 30 min | 30 min |
| 2 | Factory function unit tests | 45 min | 1h 15min |
| 3 | RecentSets component tests | 60 min | 2h 15min |
| 4 | Route integration tests | 90 min | 3h 45min |
| 5 | Type checking | 5 min | 3h 50min |
| 6 | Full test suite run | 15 min | 4h 5min |

**Total Estimated Time**: 4 hours

---

## Completion Criteria for Line 47

Line 47 in `summary-unified-routes.md` can be marked complete when:

1. ✅ All test files created
2. ✅ All tests passing (`bun run test`)
3. ✅ TypeScript compilation successful (`bun run tsc`)
4. ✅ Test coverage meets goals (80%+ for critical paths)
5. ✅ No test flakiness or intermittent failures
6. ✅ Test results documented

---

## Future Enhancements

### After Phase 1 + 1.5 Complete

1. **E2E Tests with Playwright**
   - Full user workflows
   - Auth state transitions
   - Multi-tab scenarios

2. **Visual Regression Tests**
   - Component snapshots
   - UI consistency checks

3. **Performance Tests**
   - Query performance benchmarks
   - Render performance monitoring

4. **Mutation Tests**
   - Test mutation logic independently
   - Verify optimistic updates

---

## References

- [Vitest Documentation](https://vitest.dev/)
- [React Testing Library](https://testing-library.com/react)
- [TanStack Query Testing Guide](https://tanstack.com/query/latest/docs/react/guides/testing)
- [Phase 1 Implementation Plan](./unified-route-implementation-plan.md)
- [Phase 1 Summary](./summary-unified-routes.md)

---

## Appendix: Test Template Examples

### Factory Function Test Template
```typescript
import { describe, it, expect, vi } from 'vitest'
import { getExercisesQueryOptions } from '../unified-query-options'
import * as apiExercises from '../exercises'
import * as demoQueryOptions from '@/lib/demo-data/query-options'

vi.mock('../exercises')
vi.mock('@/lib/demo-data/query-options')

describe('getExercisesQueryOptions', () => {
  it('returns API query options when user is authenticated', () => {
    const mockUser = { id: 'user123' }
    const mockApiOptions = { queryKey: ['exercises'], queryFn: vi.fn() }

    vi.mocked(apiExercises.exercisesQueryOptions).mockReturnValue(mockApiOptions)

    const result = getExercisesQueryOptions(mockUser)

    expect(apiExercises.exercisesQueryOptions).toHaveBeenCalled()
    expect(result).toBe(mockApiOptions)
  })

  it('returns demo query options when user is null', () => {
    const mockDemoOptions = { queryKey: ['demo_exercises'], queryFn: vi.fn() }

    vi.mocked(demoQueryOptions.getDemoExercisesQueryOptions).mockReturnValue(mockDemoOptions)

    const result = getExercisesQueryOptions(null)

    expect(demoQueryOptions.getDemoExercisesQueryOptions).toHaveBeenCalled()
    expect(result).toBe(mockDemoOptions)
  })
})
```

### Component Test Template
```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { RecentSets } from '../recent-sets-display'

vi.mock('@tanstack/react-query', () => ({
  useSuspenseQuery: vi.fn(),
}))

describe('RecentSets', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders nothing when exerciseId is null', () => {
    const { container } = render(
      <RecentSets exerciseId={null} user={null} />
    )

    expect(container.firstChild).toBeNull()
  })

  it('renders recent sets from demo data', async () => {
    const mockRecentSets = [
      { id: 1, reps: 10, weight: 135, date: '2025-10-01' },
    ]

    vi.mocked(useSuspenseQuery).mockReturnValue({
      data: mockRecentSets,
    })

    render(<RecentSets exerciseId={1} user={null} />)

    expect(screen.getByText('Recent Sets')).toBeInTheDocument()
    expect(screen.getByText('10 reps')).toBeInTheDocument()
    expect(screen.getByText('135 lbs')).toBeInTheDocument()
  })
})
```

---

**End of Testing Plan**
