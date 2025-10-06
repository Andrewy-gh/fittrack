## Playwright E2E Testing Steps

Check `playwright-e2e-decision-analysis.md` for details and decisions

1. [x] Run `bunx playwright install chromium --with-deps`
2. [x] Create `playwright.config.ts`
3. [ ] Convert the 6 test files from Vitest browser API to Playwright API
4. [ ] Test locally with `bunx playwright test --ui`
5. [ ] Remove any unneeded testing dependencies from `package.json` and `vitest.config.ts`if it is no longer needed.
6. [ ] Update CI workflow
7. [ ] Document in README

### Important
- Confirm user approval before proceeding with steps 2-7
- Run `bun run tsc` to ensure types are up to date and changes compile if we made any code changes.
- Remove any unnecessary comments from your new code