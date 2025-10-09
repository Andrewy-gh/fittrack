# Pre-commit E2E Testing Strategy

## Current Situation
- E2E tests take ~28 seconds locally
- CI e2e tests are hanging (see issue #TBD)
- Need to decide on pre-commit/pre-push testing strategy

## Pre-commit E2E Testing: Not Recommended ❌

### Why it's problematic:
- **Slow**: Even 28 seconds per commit is disruptive to workflow
- **Flaky**: E2E tests can be flaky, blocking commits unnecessarily
- **Overkill**: Most commits don't affect UI behavior
- **Developer friction**: Developers will skip hooks (`--no-verify`) if too slow

## Better Alternatives ✅

### Option 1: Pre-push Hook (Better balance)
Run e2e tests before pushing to remote:

```bash
# .husky/pre-push
bun run test:e2e
```

**Pros:**
- Runs before pushing to remote
- Less frequent than commits
- Still catches issues before CI

**Cons:**
- Still adds 28s delay to push
- May be skipped with `--no-verify`

### Option 2: Selective Pre-commit (Smart approach)
Only run checks on changed files:

```bash
# .husky/pre-commit
# Only run if frontend files changed
if git diff --cached --name-only | grep -q "^client/"; then
  echo "Frontend changes detected, running quick checks..."
  cd client && bun run tsc
fi
```

**Pros:**
- Type-check only (fast ~2s)
- Catches type errors immediately
- Minimal disruption

**Cons:**
- Doesn't catch runtime issues

### Option 3: GitHub Action on PR (Current approach - Best)
Keep e2e in CI only, use hooks for fast checks:

```bash
# .husky/pre-commit
cd client && bun run tsc && echo "✓ Type check passed"
```

**Pros:**
- Fast feedback (~2s)
- E2E runs comprehensively in CI
- No developer friction

**Cons:**
- E2E issues found later (in CI)

## Recommendation

**Use Option 3**: Keep e2e tests in CI only, add pre-push hook for type-checking:

```bash
# .husky/pre-push
cd client && bun run tsc && echo "✓ Type check passed"
```

### Why This Works:
1. **Fast**: Type-checking is ~2s vs 28s for e2e
2. **Effective**: Catches most issues before push
3. **Non-disruptive**: Developers won't skip hooks
4. **Comprehensive**: Full e2e coverage in CI

### Implementation Steps:
1. Create `.husky/pre-push` hook
2. Add type-check command
3. Ensure CI e2e tests work (currently blocked)
4. Document hook behavior in README

## Future Improvements
- Fix CI e2e hanging issue
- Consider running subset of critical e2e tests locally
- Add unit tests for complex components (faster than e2e)
