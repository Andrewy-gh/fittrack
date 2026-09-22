# Browser proof

Actual AnalyticsDashboard and TrainingAssessment/MuscleContributions components from `feature/hypertrophy-evidence`, mounted by a temporary Vite harness on localhost:5198. Harness was removed after verification. API responses and authenticated user are fixtures; this is not live authenticated end-to-end proof.

Desktop 1280×1000 and mobile 390×844, dark theme, America/New_York timezone. Verified hypertrophy-only goal renders no strength section, same backdated week is requested, direct/indirect contributions remain distinct, missing mappings remain visible, coverage details open, next-week request reaches current partial week and forward navigation disables. No page errors or horizontal overflow. See `proof.json`; screenshots show the completed-week state.

Frontend routed post-save regression tests additionally cover a saved backdated workout landing both strength and muscle requests on the correct week. Backend tests use actual migrated disposable PostgreSQL for ownership, boundaries and relinking. Live provider-authenticated browser-to-backend verification remains unperformed.

Surrounding workout totals/charts use empty fixtures while muscle contributions use the worked mixed-coverage example. Those surrounding totals are intentionally not a consistent live account dataset.

Validation: 32 frontend tests across Analytics, classification, mutation refresh and workouts passed; TypeScript, oxlint, formatting and production build passed. Go workout/app short tests and vet passed. The dedicated PostgreSQL ownership/window/reclassification test passed. Disposable database container removed after checks.
