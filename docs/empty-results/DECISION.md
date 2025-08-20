# ADR: Empty Results vs 404 Not Found

**Status**: Accepted  
**Date**: 2025-08-20  
**Context**: PR 49 - Inconsistent empty results handling  

## Decision

**For this codebase, we adopt the following policy:**

### Collections & Sub-Collections
- **MUST return `200 OK` with `[]` when empty**
- **NEVER return `404` based solely on emptiness**
- Applies to: `/workouts`, `/exercises`, `/exercises/{id}/recent-sets`

### Singular Resources
- **Return `404 Not Found` ONLY when resource is missing or not owned by user**
- **Return `200 OK` with object when found** (nested arrays may be `[]`)
- Applies to: `/workouts/{id}` (as object)

### Array-of-Rows Endpoints (Special Case)
- **Treat as filtered collections → return `200 []` when empty**
- Applies to: `/exercises/{id}` (current array implementation)
- **Future improvement**: Migrate to object-based responses

## Immediate Resolution for PR 49

### Problem
`GetExerciseWithSets` returns 404 when no sets found, while other collection endpoints return 200 `[]`. This creates inconsistent API behavior.

### Solution
1. **Remove the empty-404 check** in `internal/exercise/handler.go:105-108`
2. **Always return 200** with array (may be empty)
3. **Update Swagger docs** to reflect empty-array behavior
4. **Add test cases** for empty results returning 200

### Breaking Change Notice
Clients previously expecting 404 for empty `GetExerciseWithSets` must handle 200 `[]` instead.

## Rationale

1. **HTTP Semantics**: 404 = resource not found; empty collections are successful responses with payload `[]`
2. **Developer Experience**: Consistent behavior, easier iteration over arrays
3. **Industry Standards**: GitHub, Google APIs follow this pattern
4. **Cache Behavior**: 200 responses cacheable; 404s may cache differently

## Enforcement

- Added to `WARP.md` as enforceable review rules
- Comprehensive guide in `docs/empty-results/README.md`
- All future collection endpoints must follow this pattern

## References

- [Empty Results vs Not Found README](./README.md)
- [WARP.md Section 4](../../WARP.md#4-empty-results-handling---be-consistent)
- [RFC 9110: HTTP Semantics](https://tools.ietf.org/rfc/rfc9110.txt)
