# Client Testing Notes

Use these rules for new or updated client tests.

## Priorities

- Put the most weight on user-critical flows in this repo:
  - workout creation and editing
  - form validation behavior
  - analytics rendering, filtering, and state transitions
  - drag-and-drop and interaction-heavy UI
  - loading, error, and empty states
- Prefer adding or improving tests at the layer that gives the most confidence with the least mocking.

## Selector Policy

- Prefer accessible selectors first in both React Testing Library and Playwright:
  - `getByRole` / `findByRole`
  - `getByLabelText`
  - `within(...)` scoped by landmarks, dialogs, forms, or regions
- Query elements the way a user would find them: by role, name, label, visible text, and dialog context.
- Treat `data-testid` as a last resort, not a default.
- If a test cannot find an element semantically, fix the markup before adding a test id when practical.
- If `data-testid` is still necessary for a non-semantic target, keep usage narrow and note why in the test.

## Assertion Policy

- Assert on user-visible behavior and business outcomes:
  - what renders
  - what becomes enabled or disabled
  - what validation message appears
  - what state transitions after a user action
- Do not assert on implementation details like internal handler calls, query factory plumbing, or component internals unless that wiring is the feature.
- Do not assert on request bodies unless request construction itself is the behavior under test.

## Mocking and Boundaries

- Keep app logic real when possible:
  - validation rules
  - analytics transformations
  - routing behavior
  - query and mutation side effects
- Mock the app boundary when needed:
  - server and API adapters
  - analytics or data clients
  - persistence and storage
  - time and date providers
- Be careful with mocks that only prove mocked collaborators agree with each other. Replace those with integration tests where practical.

## Dependency Injection

- Prefer small DI seams for:
  - analytics and data clients
  - persistence and storage
  - time and date providers
  - server and API adapters
- Keep DI targeted. Add the seam that unlocks testing; do not introduce framework-heavy indirection.
- Do not add DI for tiny pure helpers, formatting utilities, or local-only UI helpers.

## Layer Guidance

- Unit tests:
  - use for pure domain logic and deterministic transforms
- Integration tests:
  - default investment area for routed UI, forms, validation, loading states, and analytics state changes
  - keep router, query client, and component logic real
  - mock only the external boundary
- E2E tests:
  - keep a small set of golden paths for workout CRUD, drag-and-drop, analytics filters, and major failure paths

## Review Check

Before keeping a new test, ask:

- Would this still pass if the UI became less accessible?
- Am I asserting what the user sees and can do?
- Am I mocking a boundary, or am I mocking away the behavior I actually need confidence in?
