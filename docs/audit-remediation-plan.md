# FitTrack audit remediation tracker

## Scope and counting rules

Source: `C:\Users\lenny\Desktop\fittrack-findings.txt` (FitTrack audit).

**106 source findings = 106 independently tracked entries, A001–A106, in original order.** The source did not assign IDs; these IDs are now stable. Each entry quotes the original finding, including nested bullets and source confidence. Nested bullets remain part of their parent finding, not extra findings. Coverage notes and the two closing offers are not findings.

The original F/G/D identifiers are retained **only as implementation guidance**, not as additional tracking entries. Multiple findings may share guidance; one finding may require multiple work packages. Completing one package does not automatically close every mapped finding.

Fourteen findings are accepted after independent review and appropriate verification: A001–A007, A009–A010, A029 (superseded by removal), A050, A094, A103–A104. Other findings are not complete unless their individual entries say Done. The audit's remaining claims have not all been independently reproduced.

Implementation branch: `fix/audit-remediation`. See [user decision worksheet](audit-review-decisions.md) and [orchestration log](audit-orchestration.md). Wave 1, the approved cleanup and A005–A007 passed independent review. A011 fabricated card charts is next. A088 remains unchanged pending the user's decision. See individual entries for evidence.

## Progress summary

| Disposition | Source confidence | Findings | Done |
| --- | --- | ---: | ---: |
| Ready to implement (including approved A094) | H | 43 | 12 |
| User-approved cleanup | M | 2 | 2 |
| Gated on evidence, design or approval | H | 12 | 0 |
| Deferred for user review | M | 31 | 0 |
| Deferred for user review | L | 18 | 0 |
| **Total** | **55 H / 33 M / 18 L** | **106** | **14** |

A073 contains both ready logging cleanup and a gated insertion optimization. It counts once, under Gated, and cannot be closed by completing only its logging work.

### How to update this tracker

- Update **Status**, **Evidence / commit / PR**, and **Decision / blocker** on the individual A-ID, not on the shared guidance.
- Statuses: **Not started**, **In progress**, **Blocked**, **Done**, **Already fixed**, **Not reproduced**, **Intentionally retained**.
- Use **Done** only after all parts of that source finding meet the mapped completion criteria. Record partial completion in its evidence field.
- For **Already fixed**, **Not reproduced**, or **Intentionally retained**, record evidence or the explicit decision; these are dispositions, not implemented fixes. Update the summary when statuses/dispositions change.
- Ready means the intended correction is clear, not that the source has been freshly verified. Reconfirm before editing.
- Gated items require the specified evidence/approval. Deferred items await user review; lower confidence does not mean lower risk.
- Do not renumber IDs or merge source entries. New findings after this audit should receive separate IDs beyond A106 and must not alter the original-audit total.

### Source coverage

| Original section | IDs | Count |
| --- | --- | ---: |
| Broken — Data and behaviour | A001–A017 | 17 |
| Broken — AI chat | A018–A024 | 7 |
| Broken — Server, security and operations | A025–A038 | 14 |
| Inconsistent — API contract | A039–A043 | 5 |
| Inconsistent — Schema and migrations | A044–A046 | 3 |
| Inconsistent — Docs versus reality | A047–A053 | 7 |
| Inconsistent — Client setup | A054–A058 | 5 |
| Inconsistent — Server code | A059–A067 | 9 |
| Inconsistent — CI | A068–A070 | 3 |
| Inefficient | A071–A086 | 16 |
| No longer needed — Server | A087–A096 | 10 |
| No longer needed — Client | A097–A103 | 7 |
| No longer needed — Repository | A104–A106 | 3 |
| **Total** | **A001–A106** | **106** |

The source reports passing type/lint/format/unused-code checks, 489 client tests, Go build/vet and fast tests. Those results were **not rerun for this plan**. Database integration and live E2E were not run in the source audit. Planning spot-checks confirmed client scripts, the server Makefile and CI test sequence; this is not an independent re-audit.

Existing unrelated changes in `docs/media/how-it-works.gif` and `docs/design/` must remain untouched.

## Execution order

| Batch | Scope | Prerequisites / completion gate |
| --- | --- | --- |
| 1 | F01–F04: workout data preservation | Persistence tests, including reload and nullable values |
| 2 | F05–F10: cache consistency and truthful API/UI results | Batch 1; mutation/read and API regression checks |
| 3 | F11–F16: resource bounds, middleware and operations | Proxy configuration verified; real middleware-chain tests |
| 4 | F17–F19: contract and schema correctness | Regenerate from source; migrated disposable DB comparison |
| 5 | F20–F24: docs, PWA and test configuration | Establish authoritative config before deleting duplicates |
| 6 | F25–F32: safe dead-code cleanup and duplicate CI work | Prove no production/external callers; full relevant checks |
| 7 | G01–G09: measured optimization and design decisions | Explicit evidence/decision gates below |

Keep batches in focused changes, not one giant refactor. Fix behavior before restructuring its types or SQL. A gated or deferred item must not block independent ready fixes.

## Source finding register — authoritative progress tracking

### Broken / Data and behaviour

#### A001

> H Editing a workout cannot clear notes or focus. The client sends undefined (edit-workout-page.tsx:121) and the query keeps the old value via COALESCE (server/queries/workouts.sql:82).

- **Disposition:** Ready
- **Status:** Done
- **Plan:** `F01` — Batch 1
- **Evidence / commit / PR:** Integrated workout commit `9f56129` (worker `729e0cb`). Head ran full exercise/workout suites on isolated migrated PostgreSQL 15.19: omitted/null retain, independent empty-string clears, reload and unrelated updates passed. Frontend worker regressions passed; combined client checks and fresh independent review passed. Latest checkpoint: 495 client tests, tsc/api:check and Go build/vet passed; full isolated exercise/workout suites and separate retained restricted-role SQL checks passed. A010 head repair is committed as `3e52255`; see orchestration log for exact evidence. Reviewer independently ran 19 focused client tests plus scoped lint/format and verified supplied DB/full-client evidence; no blocking issues.
- **Decision / blocker:** Accepted after AUDIT-W1-COMBINED-REVIEW (Pi / Sol / high, target `355d722..3e52255`), reconciled by head. No blocking finding. No remaining acceptance blocker for this finding.

#### A002

> H Editing any workout turns no-weight sets into weight 0 (workouts.ts:183, weight || 0).

- **Disposition:** Ready
- **Status:** Done
- **Plan:** `F02` — Batch 1
- **Evidence / commit / PR:** Integrated `9f56129`. Head's real HTTP/database regression passed null, zero and positive weight round trips through update/reload. Worker reproduced the original client coercion and its regression passed. Combined client checks and fresh independent review passed. Latest checkpoint: 495 client tests, tsc/api:check and Go build/vet passed; full isolated exercise/workout suites and separate retained restricted-role SQL checks passed. A010 head repair is committed as `3e52255`; see orchestration log for exact evidence. Reviewer independently ran 19 focused client tests plus scoped lint/format and verified supplied DB/full-client evidence; no blocking issues.
- **Decision / blocker:** Accepted after AUDIT-W1-COMBINED-REVIEW (Pi / Sol / high, target `355d722..3e52255`), reconciled by head. No blocking finding. No remaining acceptance blocker for this finding.

#### A003

> H The default workout date is computed once at module load (form-options.ts:14). In a long-lived tab or the installed PWA, a new workout defaults to when the app was opened.

- **Disposition:** Ready
- **Status:** Done
- **Plan:** `F03` — Batch 1
- **Evidence / commit / PR:** Integrated `9f56129`. Worker simulated date rollover, selected/restored dates, draft clear and save reset successfully in caller-facing tests. Combined client checks and fresh independent review passed. Latest checkpoint: 495 client tests, tsc/api:check and Go build/vet passed; full isolated exercise/workout suites and separate retained restricted-role SQL checks passed. A010 head repair is committed as `3e52255`; see orchestration log for exact evidence. Reviewer independently ran 19 focused client tests plus scoped lint/format and verified supplied DB/full-client evidence; no blocking issues.
- **Decision / blocker:** Accepted after AUDIT-W1-COMBINED-REVIEW (Pi / Sol / high, target `355d722..3e52255`), reconciled by head. No blocking finding. No remaining acceptance blocker for this finding.

#### A004

> H Reloading an in-progress draft drops the catalog link of a newly added catalog exercise. add-exercise-screen.tsx:138 sets catalog_id, but the storage schema at local-storage.ts:25 strips it.

- **Disposition:** Ready
- **Status:** Done
- **Plan:** `F04` — Batch 1
- **Evidence / commit / PR:** Integrated `9f56129`. Worker verified catalog selection → storage reload → save request retains catalog_id; legacy/custom drafts remain readable. Combined client checks and fresh independent review passed. Latest checkpoint: 495 client tests, tsc/api:check and Go build/vet passed; full isolated exercise/workout suites and separate retained restricted-role SQL checks passed. A010 head repair is committed as `3e52255`; see orchestration log for exact evidence. Reviewer independently ran 19 focused client tests plus scoped lint/format and verified supplied DB/full-client evidence; no blocking issues.
- **Decision / blocker:** Accepted after AUDIT-W1-COMBINED-REVIEW (Pi / Sol / high, target `355d722..3e52255`), reconciled by head. No blocking finding. No remaining acceptance blocker for this finding.

#### A005

> H Saving an AI workout draft from chat refreshes no caches (chat-session-workout-draft.ts:22). The workout list and analytics stay stale, and route loaders use staleTime: "static".

- **Disposition:** Ready
- **Status:** Done
- **Plan:** `F05` — Batch 2
- **Evidence / commit / PR:** Worker `d349756` integrated as `ef677d5`. Shared training-data completion policy refreshes generated workout/exercise families and initiating-user analytics; mutable route loaders honor invalidation. Worker reports 71 focused tests, tsc/lint/format and negative controls passing. Head inspected implementation and behavioral suite; repaired shared chat fixture and asynchronous route-test assumption in `2f8cd33`. Full combined gate passed 513 tests / 102 files plus format/lint/both knip modes (session35 log). No live browser/account-switch isolation proof.
- **Decision / blocker:** Accepted after AUDIT-W2-CACHES-REVIEW, Pi / Sol / high, target `e0fbcbc..2f8cd33`: no actionable introduced regressions. Reviewer independently passed 44 tests / 2 files and inspected combined evidence. Background refresh can briefly render stale data; live account-transition isolation remains unproven.

#### A006

> H Other cache gaps:
>
> - Workout create, update and delete never refresh exercise detail, recent sets or metrics history.
>
> - Update does not refresh the exercise list.
>
> - Exercise delete does not refresh workouts or the activity map.
>
> - The TODO at workouts.ts:77 assumes a low stale time; it is 5 minutes.

- **Disposition:** Ready
- **Status:** Done
- **Plan:** `F06` — Batch 2
- **Evidence / commit / PR:** Worker `d349756` integrated as `ef677d5`. Shared training-data completion policy refreshes generated workout/exercise families and initiating-user analytics; mutable route loaders honor invalidation. Worker reports 71 focused tests, tsc/lint/format and negative controls passing. Head inspected implementation and behavioral suite; repaired shared chat fixture and asynchronous route-test assumption in `2f8cd33`. Full combined gate passed 513 tests / 102 files plus format/lint/both knip modes (session35 log). No live browser/account-switch isolation proof.
- **Decision / blocker:** Accepted after AUDIT-W2-CACHES-REVIEW, Pi / Sol / high, target `e0fbcbc..2f8cd33`: no actionable introduced regressions. Reviewer independently passed 44 tests / 2 files and inspected combined evidence. Background refresh can briefly render stale data; live account-transition isolation remains unproven.

#### A007

> H GET /api/workouts/{id} returns 200 with empty sets for a missing workout (workout/service.go:129); PUT and DELETE return 404.

- **Disposition:** Ready
- **Status:** Done
- **Plan:** `F07` — Batch 2
- **Evidence / commit / PR:** Worker `2f3a712` integrated as `ac12780`. Existing ownership-scoped lookup now distinguishes missing/foreign IDs (404), owned empty/nonempty workouts (200), and database errors (safe 500). Worker reported intended pre-fix HTTP failures; head reran full isolated exercise/workout suites and retained restricted-role SQL checks successfully. Log: `C:/Users/lenny/AppData/Local/Temp/pi-unified-exec-20-81eb5ef6.log`. Existing owner-mode direct-RLS subtest skipped; HTTP ownership cases passed. A008 policy unchanged.
- **Decision / blocker:** Accepted after AUDIT-W2-WORKOUT-LOOKUP-REVIEW, Pi / Sol / high, target `2fa1225..ac12780`: no actionable findings. Reviewer independently ran uncached handler/service checks and inspected head DB evidence. No live token-middleware/outage-injection proof; A008 remains deferred.

#### A008

> M A workout whose sets were all removed by deleting an exercise looks the same as a missing one.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D01` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A009

> H Any database error is reported as 404 in exercise/historical_1rm.go:31 and exercise/metrics_history.go:40.

- **Disposition:** Ready
- **Status:** Done
- **Plan:** `F08` — Batch 2
- **Evidence / commit / PR:** Implemented in `7da9a8f`. Independent Sol/high review cleared correctness; reviewer reran both handler suites uncached (nine subtests passed). Head also ran managed short tests, Go build/vet and, after restoring isolated DB setup, the full exercise suite: error-mapping cases passed.
- **Decision / blocker:** Accepted separately from A010. Reviewer report AUDIT-W1-EXERCISE-REVIEW-1; no remaining A009 blocker.

#### A010

> H Exercise list and create return created_at/updated_at as null although the API contract marks them required (exercise/repository.go:54).

- **Disposition:** Ready
- **Status:** Done
- **Plan:** `F09` — Batch 2
- **Evidence / commit / PR:** Head corrected persisted timestamps with new migrations 00034/00035 and schema default/NOT NULL. Replaced the initially mis-authenticated test with POST/list/detail JSON, distinct existing timestamps, transaction/catalog persistence and real migration-backfill coverage. Red run failed on null updated_at; green full exercise/workout suites passed. Logs: pi-unified-exec-31-e9f3ac82.log (red), pi-unified-exec-37-c8b6c187.log (green), under the user's Local/Temp directory. Latest checkpoint: 495 client tests, tsc/api:check and Go build/vet passed; full isolated exercise/workout suites and separate retained restricted-role SQL checks passed. A010 head repair is committed as `3e52255`; see orchestration log for exact evidence. Reviewer independently ran 19 focused client tests plus scoped lint/format and verified supplied DB/full-client evidence; no blocking issues.
- **Decision / blocker:** Accepted after AUDIT-W1-COMBINED-REVIEW (Pi / Sol / high, target `355d722..3e52255`), reconciled by head. No blocking finding. No remaining acceptance blocker for this finding.

#### A011

> H The mini chart on every exercise card in the new and edit pages shows hard-coded numbers (new-workout-page.tsx:100).

- **Disposition:** Ready
- **Status:** In progress
- **Plan:** `F10` — Batch 2
- **Evidence / commit / PR:** Worker `863bcb3` integrated as `caf2b2b`: removed both fabricated mini-charts and unused component/prop plumbing; real set count/volume retained. Worker reports intended red and seven focused tests passing. Head reran five page tests, tsc and both knip modes successfully (session42), then two card-section tests passed (session43). Live desktop/mobile layout not verified; worker only ran browser bridge fixture preflight.
- **Decision / blocker:** Fresh independent review pending; not accepted yet.

#### A012

> M The workout list is re-sorted by creation time on the client (workouts-filters.ts:24) while the server sorts by workout date, so backdated workouts appear newest.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D02` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A013

> M Day boundaries use UTC or server time in the activity map, metrics history, the chat "current date", draft dates and chat tool dates. The assessment endpoints take an explicit time zone.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D03` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A014

> M "Delete account" deletes app data and signs out but never deletes the Stack Auth identity (account-settings-page.tsx:85). Signing in again creates a fresh empty account. This may be intended.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D04` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A015

> M Reps have no upper bound on the server. Extreme reps × weight overflow NUMERIC(10,1) and numeric(8,2) casts and return 500 on save or read.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D05` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A016

> M navigate() is called during render for an out-of-range exercise index in both workout pages.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D06` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A017

> L The first-run exercise name has no length limit or trim, unlike the 256 limit elsewhere.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D07` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

### Broken / AI chat

#### A018

> H There is no size limit on a chat prompt (aichat/handler_helpers.go:36), and the whole conversation history is sent to the model every turn (aichat/runtime.go:178).

- **Disposition:** Ready
- **Status:** Not started
- **Plan:** `F11` — Batch 3
- **Evidence / commit / PR:** —
- **Decision / blocker:** Reconfirm against current code, implement and verify using the mapped completion criteria.

#### A019

> M If generation fails to start, the run is left "streaming" with no abort (aichat/handler.go:442). The conversation stays busy until the stale timeout.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D08` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A020

> M A trial prompt is consumed before generation, so a provider failure still costs the user one.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D09` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A021

> M The client recovery poll gives up at 55 s (ai-chat.ts:211) while the server may legitimately generate for 75 s.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D10` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A022

> L The model's structured output type is the full create-workout request, including catalog_id and recommendations. A hallucinated value passes draft validation and fails at save with a 500.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D11` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A023

> L Stored drafts are re-validated on every read. Tightening validation later would make old conversations fail to load.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D11` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A024

> L Draft quality checks are substring-based ("back" matches "coming back from a break"), and a second failed repair fails the whole turn.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D11` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

### Broken / Server, security and operations

#### A025

> H, reproduced Metrics and the request log sit outside the auth middleware, which copies the request (app.go:323, auth/middleware.go:126).
>
> - Every authenticated API route is labelled /api/*.
>
> - user_present is always false.
>
> - The middleware tests set the pattern by hand, so they miss it.

- **Disposition:** Ready
- **Status:** Not started
- **Plan:** `F12` — Batch 3
- **Evidence / commit / PR:** —
- **Decision / blocker:** Reconfirm against current code, implement and verify using the mapped completion criteria.

#### A026

> H The CORS allow-list omits PATCH (middleware/cors.go:30) although PATCH routes exist. It only matters cross-origin. There is also no Vary: Origin.

- **Disposition:** Ready
- **Status:** Not started
- **Plan:** `F13` — Batch 3
- **Evidence / commit / PR:** —
- **Decision / blocker:** Reconfirm against current code, implement and verify using the mapped completion criteria.

#### A027

> H HSTS is only sent when the Go server itself terminates TLS (middleware/security.go:28), which never happens behind Fly.

- **Disposition:** Ready
- **Status:** Not started
- **Plan:** `F14` — Batch 3
- **Evidence / commit / PR:** —
- **Decision / blocker:** Reconfirm against current code, implement and verify using the mapped completion criteria.

#### A028

> H The AI runtime, trace and tool code log through the global slog logger, which is never configured. Those lines ignore LOG_LEVEL and the JSON format.

- **Disposition:** Ready
- **Status:** Not started
- **Plan:** `F15` — Batch 3
- **Evidence / commit / PR:** —
- **Decision / blocker:** Reconfirm against current code, implement and verify using the mapped completion criteria.

#### A029

> H server/scripts/run-complete-rls-audit.sh cannot run: it uses PowerShell Get-Content inside bash and a container name that no longer exists.

- **Disposition:** Ready
- **Status:** Done
- **Plan:** `F16` — Batch 3
- **Evidence / commit / PR:** Worker proved obsolete runner defects and removed it with superseded fixtures/probes in `b3a7ec2` (source `5adeefeb`), under user-authorized A094 cleanup. Maintained runtime-role tools retained. This supersedes repair; it does not claim the historical audit passes. Reviewer independently ran 19 focused client tests plus scoped lint/format and verified supplied DB/full-client evidence; no blocking issues.
- **Decision / blocker:** Accepted after AUDIT-W1-COMBINED-REVIEW (Pi / Sol / high, target `355d722..3e52255`), reconciled by head. No blocking finding. Resolution: obsolete runner removed, so repair is superseded; no claim the retired audit passes.

#### A030

> M /metrics is also mounted on the public port and is open when no credentials are set (routes.go:46).

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D12` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A031

> M The static handler serves directory listings for paths such as /assets/ (routes.go:184). Not run.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D13` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A032

> M Tokens are checked for signature and expiry only, not issuer or audience (auth/middleware.go:176).

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D14` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A033

> M Every checkout grants a new 7-day trial and a new prompt allowance (billing/service.go:100), so cancel-and-resubscribe repeats the trial.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D15` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A034

> M The Stripe webhook returns 400 for database errors, and events that can never resolve are retried for days.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D16` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A035

> M fly-deploy.yml:40 appends ?default_query_exec_mode=… to the database URL, which breaks if the secret already has a query string.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D17` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A036

> M Error-message sanitising uses substring matches ("token", "table", "connection"), so unrelated errors can be rewritten to "unauthorized".

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D18` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A037

> L An unpaid subscription is not cancelled on account deletion.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D19` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A038

> L The runtime panic log always names GOOGLE_API_KEY (runtime.go:76).

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D20` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

### Inconsistent / API contract

#### A039

> H The OpenAPI spec has no billing endpoints and no DELETE /account, so the client hand-types them.

- **Disposition:** Ready
- **Status:** Not started
- **Plan:** `F17`, `F18` — Batch 4
- **Evidence / commit / PR:** —
- **Decision / blocker:** Reconfirm against current code, implement and verify using the mapped completion criteria.

#### A040

> H Three client modules hand-type endpoints that are in the spec: muscle contributions, training assessment and training profile.

- **Disposition:** Ready
- **Status:** Not started
- **Plan:** `F18` — Batch 4
- **Evidence / commit / PR:** —
- **Decision / blocker:** Reconfirm against current code, implement and verify using the mapped completion criteria.

#### A041

> H Doc-only response structs have drifted from the real ones (UpdateWorkoutRequestSwagger, the exercise response types).

- **Disposition:** Ready
- **Status:** Not started
- **Plan:** `F17` — Batch 4
- **Evidence / commit / PR:** —
- **Decision / blocker:** Reconfirm against current code, implement and verify using the mapped completion criteria.

#### A042

> M Chat streaming and stop use their own fetch, auth headers and base URL, bypassing the shared 401 handling.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D21` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A043

> M About 400 lines of chat parsers duplicate generated types.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D22` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

### Inconsistent / Schema and migrations

#### A044

> H schema.sql has drifted from the migrations: two indexes, the stopped statuses, three foreign keys and a check constraint are missing.

- **Disposition:** Ready
- **Status:** Not started
- **Plan:** `F19` — Batch 4
- **Evidence / commit / PR:** —
- **Decision / blocker:** Reconfirm against current code, implement and verify using the mapped completion criteria.

#### A045

> H Migration 00029 is missing, 00031 creates a table that 00032 drops, and the down step of 00004 is empty.

- **Disposition:** Gated
- **Status:** Not started
- **Plan:** `G01` — Gated work
- **Evidence / commit / PR:** —
- **Decision / blocker:** Satisfy the mapped evidence/approval gate before closing this finding.

#### A046

> M Catalog IDs live in four places: the JSON catalog, a migration check, schema.sql and the muscle-role map.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D23` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

### Inconsistent / Docs versus reality

#### A047

> H docs/development.md:209 lists a make docker-up target that does not exist.

- **Disposition:** Ready
- **Status:** Not started
- **Plan:** `F20` — Batch 5
- **Evidence / commit / PR:** —
- **Decision / blocker:** Reconfirm against current code, implement and verify using the mapped completion criteria.

#### A048

> H client/README.md:97 says to regenerate with bun run openapi-ts, which skips the Swagger step.

- **Disposition:** Ready
- **Status:** Not started
- **Plan:** `F20` — Batch 5
- **Evidence / commit / PR:** —
- **Decision / blocker:** Reconfirm against current code, implement and verify using the mapped completion criteria.

#### A049

> H client/README.md:131 says the service worker caches API responses.

- **Disposition:** Ready
- **Status:** Not started
- **Plan:** `F20` — Batch 5
- **Evidence / commit / PR:** —
- **Decision / blocker:** Reconfirm against current code, implement and verify using the mapped completion criteria.

#### A050

> H server/scripts/README.md describes a workflow file and secrets that do not exist.

- **Disposition:** Ready
- **Status:** Done
- **Plan:** `F21` — Batch 5
- **Evidence / commit / PR:** Rewrote server/scripts/README.md in `b3a7ec2` to describe real CI/local callers, credential boundaries and limitations; removed fictitious workflow/secrets guidance. Worker verified all eight links. Reviewer independently ran 19 focused client tests plus scoped lint/format and verified supplied DB/full-client evidence; no blocking issues.
- **Decision / blocker:** Accepted after AUDIT-W1-COMBINED-REVIEW (Pi / Sol / high, target `355d722..3e52255`), reconciled by head. No blocking finding. No remaining acceptance blocker for this finding.

#### A051

> H docs/environment.md lists Stripe secrets that no workflow uses.

- **Disposition:** Ready
- **Status:** Not started
- **Plan:** `F21` — Batch 5
- **Evidence / commit / PR:** —
- **Decision / blocker:** Reconfirm against current code, implement and verify using the mapped completion criteria.

#### A052

> H client/docs/api-type-generation.md describes a different generator.

- **Disposition:** Ready
- **Status:** Not started
- **Plan:** `F20` — Batch 5
- **Evidence / commit / PR:** —
- **Decision / blocker:** Reconfirm against current code, implement and verify using the mapped completion criteria.

#### A053

> H .cursorrules says pnpx, and the pre-commit hook falls back to pnpm with no pnpm lockfile.

- **Disposition:** Ready
- **Status:** Not started
- **Plan:** `F21` — Batch 5
- **Evidence / commit / PR:** —
- **Decision / blocker:** Reconfirm against current code, implement and verify using the mapped completion criteria.

### Inconsistent / Client setup

#### A054

> H There are two PWA manifests with different names: public/manifest.json and the generated one.

- **Disposition:** Ready
- **Status:** Not started
- **Plan:** `F22` — Batch 5
- **Evidence / commit / PR:** —
- **Decision / blocker:** Reconfirm against current code, implement and verify using the mapped completion criteria.

#### A055

> H vite.config.js:76 has a test block that Vitest ignores, and there are two test setup files.

- **Disposition:** Ready
- **Status:** Not started
- **Plan:** `F23` — Batch 5
- **Evidence / commit / PR:** —
- **Decision / blocker:** Reconfirm against current code, implement and verify using the mapped completion criteria.

#### A056

> M Demo mode differs from the API:
>
> - 10 recent sets rather than 3.
>
> - 0-based rather than 1-based ordering.
>
> - A case-insensitive rename check.
>
> - Simulated 50–200 ms delays on every query.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D24` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A057

> M Exercise goals are stored in one global localStorage key, not per user, and hard-code "lb".

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D25` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A058

> L The edit page uses window.confirm where the new page uses a dialog.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D26` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

### Inconsistent / Server code

#### A059

> H The workout request has four parallel type sets (create, update, draft, Postgres).

- **Disposition:** Gated
- **Status:** Not started
- **Plan:** `G02` — Gated work
- **Evidence / commit / PR:** —
- **Decision / blocker:** Satisfy the mapped evidence/approval gate before closing this finding.

#### A060

> H Small helpers are re-implemented in four packages with different trimming behaviour.

- **Disposition:** Gated
- **Status:** Not started
- **Plan:** `G03` — Gated work
- **Evidence / commit / PR:** —
- **Decision / blocker:** Satisfy the mapped evidence/approval gate before closing this finding.

#### A061

> M Gemini and trace settings are read straight from the environment, bypassing the validated config.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D27` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A062

> M The alert rules are duplicated in ops/prometheus and ops/grafana-cloud, differing by one line.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D28` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A063

> M GET /api/features/access returns internal granted_by and note fields to end users.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D29` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A064

> L Error-message casing varies between handlers.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D30` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A065

> L response.Error and response.ErrorResponse are identical types used interchangeably.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D30` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A066

> L The muscle-contribution handler reports "Could not load training assessment".

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D30` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A067

> L The telemetry cohort labels are still "beta" and "non_beta".

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D30` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

### Inconsistent / CI

#### A068

> M Tests run only on pull requests to main, and the deploy has no test gate.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D31` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A069

> M The deploy and test workflows do not verify the Goose version, which the preview workflow does.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D32` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A070

> L go mod tidy -diff fails on a Windows checkout only because go.sum is checked out with CRLF.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D33` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

### Inefficient

#### A071

> H Live chat is delivered by database polling: about five queries every 250 ms per open stream (service_stream.go:80), plus an ownership check every 250 ms per run. The pool is 15 connections, 3 in preview.

- **Disposition:** Gated
- **Status:** Not started
- **Plan:** `G05` — Gated work
- **Evidence / commit / PR:** —
- **Decision / blocker:** Satisfy the mapped evidence/approval gate before closing this finding.

#### A072

> H Each model chunk is its own transaction and rewrites the full message text so far (repository_stream.go:316).

- **Disposition:** Gated
- **Status:** Not started
- **Plan:** `G05` — Gated work
- **Evidence / commit / PR:** —
- **Decision / blocker:** Satisfy the mapped evidence/approval gate before closing this finding.

#### A073

> H Sets are inserted one query at a time, with an info log per set and several per exercise (repository_utils.go:71).

- **Disposition:** Gated
- **Status:** Not started
- **Plan:** `F32`, `G04` — Batch 6; Gated work. Logging cleanup (F32) may proceed; bulk insertion (G04) remains gated.
- **Evidence / commit / PR:** —
- **Decision / blocker:** Satisfy the mapped evidence/approval gate before closing this finding.

#### A074

> H Workout create converts request to internal to request to internal to Postgres types (repository.go:237).

- **Disposition:** Gated
- **Status:** Not started
- **Plan:** `G02` — Gated work
- **Evidence / commit / PR:** —
- **Decision / blocker:** Satisfy the mapped evidence/approval gate before closing this finding.

#### A075

> H Three indexes are redundant prefixes of others (schema.sql:322: idx_set_user_id, idx_workout_user_id, idx_exercise_user_id).

- **Disposition:** Gated
- **Status:** Not started
- **Plan:** `G06` — Gated work
- **Evidence / commit / PR:** —
- **Decision / blocker:** Satisfy the mapped evidence/approval gate before closing this finding.

#### A076

> H Three near-identical 50-line metrics queries, and three identical mappers, exist where one parameterised query would do.

- **Disposition:** Gated
- **Status:** Not started
- **Plan:** `G06` — Gated work
- **Evidence / commit / PR:** —
- **Decision / blocker:** Satisfy the mapped evidence/approval gate before closing this finding.

#### A077

> H CI runs go test -short ./... and then the full suite, which is a superset (test.yml:125).

- **Disposition:** Ready
- **Status:** Not started
- **Plan:** `F31` — Batch 6
- **Evidence / commit / PR:** —
- **Decision / blocker:** Reconfirm against current code, implement and verify using the mapped completion criteria.

#### A078

> H Vitest reports that creating jsdom 99 times is 38% of the 188 s client test run.

- **Disposition:** Gated
- **Status:** Not started
- **Plan:** `G07` — Gated work
- **Evidence / commit / PR:** —
- **Decision / blocker:** Satisfy the mapped evidence/approval gate before closing this finding.

#### A079

> M No index exists on exercise.historical_1rm_source_workout_id, which is queried on every workout update and delete.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D34` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A080

> M Every API request does a user lookup, and every pool checkout does a set_config round trip.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D35` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A081

> M A 401 produces three log lines, and all 4xx responses are logged at error level.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D36` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A082

> M stripe.Key is reassigned on every request (billing/service.go:80), an unsynchronised global write.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D37` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A083

> L The draft is written to localStorage on every keystroke and parsed on every workout-list render.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D38` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A084

> L Each telemetry event costs a database query.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D39` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A085

> L Static files are opened twice per request.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D40` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A086

> L make dev and make test-db-ready duplicate about 60 lines.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D41` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

### No longer needed / Server

#### A087

> H The "Phase-0" validation flow: two unrouted handlers plus their service methods, runtime methods, prompts and types.

- **Disposition:** Ready
- **Status:** Not started
- **Plan:** `F25` — Batch 6
- **Evidence / commit / PR:** —
- **Decision / blocker:** Reconfirm against current code, implement and verify using the mapped completion criteria.

#### A088

> H The training-evidence service and its query are reachable only from tests (training_evidence.go:115). This may be parked on purpose.

- **Disposition:** Gated
- **Status:** Not started
- **Plan:** `G08` — Gated work
- **Evidence / commit / PR:** Head inspected training_evidence.go, ListTrainingEvidence SQL, documentation and callers. Dedicated service/query are test-only; shared period/types/helpers are used by live assessment and muscle-contribution features.
- **Decision / blocker:** User requested explanation, not deletion. Preserve service/query and live shared helpers until a decision.

#### A089

> H EnsureAIChatPromptAllowed and CancelCurrentSubscriptionRenewal (billing/service.go:303, :243) have no callers, which also strands one repository method and one query.

- **Disposition:** Ready
- **Status:** Not started
- **Plan:** `F26` — Batch 6
- **Evidence / commit / PR:** —
- **Decision / blocker:** Reconfirm against current code, implement and verify using the mapped completion criteria.

#### A090

> H Four queries are never called: GetSet, ListSets, UpdateSet, DeleteSetsByWorkoutAndExercise.

- **Disposition:** Ready
- **Status:** Not started
- **Plan:** `F26` — Batch 6
- **Evidence / commit / PR:** —
- **Decision / blocker:** Reconfirm against current code, implement and verify using the mapped completion criteria.

#### A091

> H Unreachable or no-op functions:
>
> - FormatValidationErrors
>
> - isClientSafeError
>
> - normalizeStreamFailure (returns its input on every path)
>
> - IsForeignKeyConstraintError
>
> - IsRLSContextError
>
> - two lifecycle predicates
>
> - aichateval.ScoreResult
>
> - config.StripeConfigured
>
> - ErrUnsupportedStripeEvent

- **Disposition:** Ready
- **Status:** Not started
- **Plan:** `F27` — Batch 6
- **Evidence / commit / PR:** —
- **Decision / blocker:** Reconfirm against current code, implement and verify using the mapped completion criteria.

#### A092

> H workout_test_helpers.go is not a _test.go file, so its mocks and test database helpers compile into the production package.

- **Disposition:** Ready
- **Status:** Not started
- **Plan:** `F28` — Batch 6
- **Evidence / commit / PR:** —
- **Decision / blocker:** Reconfirm against current code, implement and verify using the mapped completion criteria.

#### A093

> H server/cmd/test-jwk is a scratch program with placeholder values.

- **Disposition:** Ready
- **Status:** Not started
- **Plan:** `F29` — Batch 6
- **Evidence / commit / PR:** —
- **Decision / blocker:** Reconfirm against current code, implement and verify using the mapped completion criteria.

#### A094

> H Six scripts in server/scripts are referenced by nothing, and two of the SQL files differ only in their title.

- **Disposition:** Ready
- **Status:** Done
- **Plan:** `G08` — Gated work
- **Evidence / commit / PR:** User-authorized purpose inventory complete. Removed six obsolete/unsafe scripts in `b3a7ec2`; four useful SQL tools unchanged. Source corrections: scripts had manual/internal callers and two privilege files were not identical. Worker reproduced false success/incorrect cross-user expectation with network-free Bash mocks. Reviewer independently ran 19 focused client tests plus scoped lint/format and verified supplied DB/full-client evidence; no blocking issues.
- **Decision / blocker:** Accepted after AUDIT-W1-COMBINED-REVIEW (Pi / Sol / high, target `355d722..3e52255`), reconciled by head. No blocking finding. No remaining acceptance blocker for this finding.

#### A095

> M workout/validator.go and two redundant ID-validation structs in the exercise package.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D42` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A096

> L Nil checks in routes.go that are always true.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D42` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

### No longer needed / Client

#### A097

> H Five components have no importers and are hidden from knip by ignoreFiles: date-picker-base, dialog-base, error-boundary-test, mode-toggle, ui/table.

- **Disposition:** Ready
- **Status:** Not started
- **Plan:** `F29` — Batch 6
- **Evidence / commit / PR:** —
- **Decision / blocker:** Reconfirm against current code, implement and verify using the mapped completion criteria.

#### A098

> H Starter-template leftovers: reportWebVitals.ts (called with no callback, main.tsx:69), the web-vitals dependency, src/logo.svg, .cta.json, and a commented-out block in lib/utils.ts.

- **Disposition:** Ready
- **Status:** Not started
- **Plan:** `F30` — Batch 6
- **Evidence / commit / PR:** —
- **Decision / blocker:** Reconfirm against current code, implement and verify using the mapped completion criteria.

#### A099

> H @vitest/browser 3.2.4 sits beside Vitest 5.0.1 and is referenced only in tsconfig.json.

- **Disposition:** Ready
- **Status:** Not started
- **Plan:** `F23` — Batch 5
- **Evidence / commit / PR:** —
- **Decision / blocker:** Reconfirm against current code, implement and verify using the mapped completion criteria.

#### A100

> H src/test/setup.ts is never loaded.

- **Disposition:** Ready
- **Status:** Not started
- **Plan:** `F23` — Batch 5
- **Evidence / commit / PR:** —
- **Decision / blocker:** Reconfirm against current code, implement and verify using the mapped completion criteria.

#### A101

> H The service-worker caching rule at vite.config.js:51 matches https://api.…, which the app never calls.

- **Disposition:** Ready
- **Status:** Not started
- **Plan:** `F22` — Batch 5
- **Evidence / commit / PR:** —
- **Decision / blocker:** Reconfirm against current code, implement and verify using the mapped completion criteria.

#### A102

> H Two pull-to-refresh docs describe a feature that does not exist.

- **Disposition:** Ready
- **Status:** Not started
- **Plan:** `F24` — Batch 5
- **Evidence / commit / PR:** —
- **Decision / blocker:** Reconfirm against current code, implement and verify using the mapped completion criteria.

#### A103

> M Three client/docs files are copies of upstream TanStack docs.

- **Disposition:** Ready
- **Status:** Done
- **Plan:** `D43` — User review
- **Evidence / commit / PR:** User-approved three copied TanStack documents removed in `6269904`. No remaining inbound non-audit references. Commit checks passed format/lint, both knip modes and all 489 client tests; closeout review passed. Reviewer independently ran 19 focused client tests plus scoped lint/format and verified supplied DB/full-client evidence; no blocking issues.
- **Decision / blocker:** Accepted after AUDIT-W1-COMBINED-REVIEW (Pi / Sol / high, target `355d722..3e52255`), reconciled by head. No blocking finding. No remaining acceptance blocker for this finding.

### No longer needed / Repository

#### A104

> M output/ holds 2.3 MB of one-off verification screenshots and JSON, referenced by one doc.

- **Disposition:** Ready
- **Status:** Done
- **Plan:** `D44` — User review
- **Evidence / commit / PR:** User-approved tracked output artifacts removed in `6269904`. Two evidence references corrected in docs/training-assessment-policy.md, retaining historical caveats. Non-audit reference scan clean; closeout review passed. Reviewer independently ran 19 focused client tests plus scoped lint/format and verified supplied DB/full-client evidence; no blocking issues.
- **Decision / blocker:** Accepted after AUDIT-W1-COMBINED-REVIEW (Pi / Sol / high, target `355d722..3e52255`), reconciled by head. No blocking finding. No remaining acceptance blocker for this finding.

#### A105

> L .gitignore has a redundant server/tmp/ai-chat-scenario-sweeps/ entry, and the Makefile's GOGET variable is unused.

- **Disposition:** Deferred
- **Status:** Not started
- **Plan:** `D41` — User review
- **Evidence / commit / PR:** —
- **Decision / blocker:** Awaiting user review; see the mapped decision guidance.

#### A106

> H, your call Untracked local leftovers: server/api.exe (55 MB, from August), server/bin/, server/.setenv.sh, the Postgres 16 data directory and two backup directories.

- **Disposition:** Gated
- **Status:** Not started
- **Plan:** `G09` — Gated work
- **Evidence / commit / PR:** —
- **Decision / blocker:** Satisfy the mapped evidence/approval gate before closing this finding.

## Using the implementation guidance

The following 85 F/G/D entries describe how to address the 106 findings. They deliberately have no progress checkboxes. Apply only the relevant portion of shared guidance to each A-ID and record progress in that A-ID's entry. Source order and confidence remain unchanged above.

## Implementation guidance — ready work packages

### Batch 1 — preserve workout data

- **F01 — Allow clearing workout notes and focus.** Inspect `edit-workout-page.tsx`, request validation/types and `server/queries/workouts.sql`. Define omitted = retain and explicit empty/null = clear as appropriate to the existing API; replace `COALESCE` behavior where it prevents clearing. Regenerate SQL/API artifacts when necessary. **Done when:** populated values can be cleared and remain cleared after reload; omitted fields still preserve values; unrelated updates remain intact.
- **F02 — Preserve no-weight sets on edit.** Remove the `weight || 0` fallback in `workouts.ts`; preserve absent/null weight through request, database and response without conflating it with explicit zero. **Done when:** editing another field round-trips no-weight, zero-weight and weighted sets distinctly according to the contract.
- **F03 — Compute the default date when starting a workout.** Replace the module-load date in `form-options.ts` with a per-new-form/draft default. Do not overwrite restored drafts or user-selected dates. **Done when:** starting a second workout after a day rollover uses the current date without reloading the app; existing drafts retain their dates.
- **F04 — Persist catalog exercise links in drafts.** Add optional `catalog_id` to the local-storage schema and preserve it in serialization/restoration. Keep older drafts without it readable. **Done when:** add catalog exercise → reload → save retains the catalog identity; custom exercises and legacy drafts still work.

### Batch 2 — fresh caches, accurate errors and honest UI

- **F05 — Refresh data after saving a chat workout draft.** Route `chat-session-workout-draft.ts` through the same mutation completion policy as other workout creation, without adding a pass-through abstraction. **Done when:** workout list, analytics and subsequently opened routes show the new workout without manual reload.
- **F06 — Complete the mutation/cache dependency map.** In `workouts.ts` and exercise mutations, cover create/update/delete → exercise detail, recent sets and metrics history; workout update → exercise list; exercise delete → workouts and activity map. Include existing analytics/list invalidations, affected user/query parameters, and inactive caches. Remove the incorrect low-stale-time TODO. Inspect both Query and Router caches: invalidation must actually refresh routes using `staleTime: "static"`, not merely mark unused keys stale. **Done when:** each mutation followed by navigation/read shows authoritative server state, with no unrelated-user invalidation or redundant refetch storm.
- **F07 — Return 404 for a missing workout.** In `workout/service.go` and its repository boundary, check workout existence/ownership independently of returned sets. **Done when:** nonexistent and other-user IDs return 404; an existing workout with no sets is not accidentally classified as missing. Product treatment of empty workouts remains D01.
- **F08 — Stop masking database failures as 404.** In `exercise/historical_1rm.go` and `exercise/metrics_history.go`, map only genuine missing-resource errors to 404; use the existing safe server-error response/logging path for database failures. **Done when:** missing resources return 404, injected database failures return 5xx without leaking SQL or secrets, and successful reads are unchanged.
- **F09 — Populate exercise timestamps.** Fix list/create mappings in `exercise/repository.go`, including the actual query projections if needed. **Done when:** persisted `created_at` and `updated_at` serialize as valid required timestamps on list/create and agree with detail responses.
- **F10 — Remove fabricated exercise-card charts.** Remove hard-coded numbers from both new/edit workout pages. Use an existing real-data source only if its meaning is already established; otherwise omit the chart or show an honest empty state. **Done when:** no production card presents synthetic history as user data, including users with no history.

### Batch 3 — bounds, observability and deployment correctness

- **F11 — Bound chat input and model context.** At `aichat/handler_helpers.go`, apply a request-body bound and prompt-length validation before storing a run or consuming quota. At `aichat/runtime.go`, enforce the selected model's context/input budget, reserving output capacity and retaining required system/tool-message structure. Prefer an explicit safe rejection to silent history loss until truncation policy is decided. Document configurable limits and useful user-facing errors. **Done when:** oversized requests and accumulated-history overflow fail predictably before provider calls/quota consumption; boundary-size normal chats work. Context summarization/truncation UX belongs to D11, not this minimum safety fix.
- **F12 — Correct route and authenticated-user observability.** Fix propagation across the copied request in `app.go` / `auth/middleware.go`; preserve intended authentication ordering. **Done when:** requests through the real application router label metrics/logs with bounded route templates rather than `/api/*`, and authenticated requests report `user_present=true`; unauthenticated requests remain false. Do not expose raw user IDs or concrete URL IDs as metric labels. Do not prove this only by manually setting a pattern in a middleware unit test.
- **F13 — Correct CORS responses.** Add PATCH to allowed methods and correctly append `Vary: Origin` without discarding existing Vary values; check preflight handling and allowed/disallowed origins. **Done when:** allowed cross-origin PATCH preflights succeed and intermediary caches cannot reuse one origin's response incorrectly.
- **F14 — Make HSTS effective behind Fly.** Verify existing edge headers and trust boundaries first. Set HSTS at the trusted TLS edge or in explicitly configured production middleware, rather than relying only on `r.TLS`. Never trust arbitrary forwarded headers from public clients. **Done when:** public HTTPS responses include the intended HSTS header and local development is unaffected; record live verification or mark deployment verification blocked.
- **F15 — Configure logging consistently.** Configure the default `slog` logger at startup, or consistently pass the existing configured logger into AI runtime/trace/tool code. Preserve redaction. **Done when:** representative AI logs honor LOG_LEVEL and JSON formatting just like HTTP logs, without logging prompt contents or credentials inadvertently.
- **F16 — Repair the complete RLS audit script.** Replace PowerShell `Get-Content` in `server/scripts/run-complete-rls-audit.sh` with Bash-compatible reading and discover the database through the maintained Compose project/service or explicit connection configuration, not a retired container name. **Done when:** syntax checks and the complete audit pass against a disposable database using the current setup; failures propagate a nonzero exit. Never target or remove the legacy Postgres data directories.

### Batch 4 — API contract and schema

- **F17 — Document missing API endpoints and eliminate contract drift.** Add billing endpoints and DELETE `/account` to authoritative Swagger annotations; reconcile `UpdateWorkoutRequestSwagger` and exercise response structs with real handlers. Generate with `bun run --cwd client api:generate` (also available via `make -C server swagger`), not the client-only generator. **Done when:** generated contracts match real request/response behavior and `api:check` passes.
- **F18 — Use generated types for documented APIs.** After F17, replace duplicated endpoint types in muscle contributions, training assessment and training profile; migrate billing/account handwritten types where the completed contract permits. Preserve intentional UI/domain transformations. **Done when:** callers compile using generated API types and behavior remains unchanged.
- **F19 — Reconcile `schema.sql` with applied migrations.** Inventory the reported two missing indexes, stopped statuses, three foreign keys and check constraint. Compare a database built from migrations with the schema representation, then update the canonical schema and regenerate SQL bindings. If the deployed database itself needs a change, add a new forward migration. **Done when:** both representations agree on intended constraints/indexes/statuses, fresh migrations succeed, and relevant persistence tests pass. Do not rewrite applied migration history; G01 covers historical anomalies.

### Batch 5 — documentation and client configuration

- **F20 — Repair setup/API/PWA documentation.** Correct the nonexistent `make docker-up` in `docs/development.md` using the current supported setup; correct `client/README.md` to use `api:generate` and describe actual service-worker caching; rewrite `client/docs/api-type-generation.md` around the actual generator. **Done when:** every documented command exists and the API workflow produces fresh server and client artifacts.
- **F21 — Align scripts, environment and package-manager guidance.** Reconcile `server/scripts/README.md` with existing workflows, remove or clearly label obsolete Stripe workflow-secret requirements in `docs/environment.md` without removing runtime-required settings, replace `.cursorrules` pnpx guidance, and remove the unsupported pnpm fallback from the pre-commit hook in favor of the repo's Bun workflow and an actionable missing-tool error. **Done when:** docs name real workflows/settings and hook verification works with the supported toolchain.
- **F22 — Make the PWA manifest and caching configuration authoritative.** Inspect HTML manifest links and generated output; keep one authoritative manifest, preserving the intended app identity/start URL/scope to avoid installed-PWA regressions. Remove the unused `https://api.…` runtime caching rule; do not introduce API caching as a replacement. **Done when:** production output links to one correct manifest and normal online/offline behavior is unchanged except removal of dead configuration. Manifest identity choices that cannot be inferred move to user review.
- **F23 — Consolidate active test configuration.** Remove ignored test settings from `vite.config.js` and the unloaded `src/test/setup.ts`, first moving any genuinely required setup into the active Vitest config/setup. Remove unused `@vitest/browser` and its tsconfig reference if no browser-mode tests depend on it; update the lockfile. **Done when:** the suite uses the intended setup once, test discovery is unchanged, and type checking passes.
- **F24 — Retire nonexistent-feature documentation.** Remove or explicitly archive the two pull-to-refresh documents and repair inbound links; do not imply the feature is implemented. **Done when:** current documentation no longer promises unavailable behavior.

### Batch 6 — provably dead code and duplicate work

For every removal: search production code, tests, scripts, generation configuration and documented external entry points. An unreferenced executable/script is not automatically unused. Stop and move to G08 if evidence is ambiguous.

- **F25 — Remove the unrouted Phase-0 validation flow.** Remove the two obsolete handlers and exclusively reachable service/runtime methods, prompts and types. **Done when:** routed chat flows and eval tooling still build/test; generated contracts contain no obsolete endpoints.
- **F26 — Remove unreachable billing methods and set queries.** Remove `EnsureAIChatPromptAllowed`, `CancelCurrentSubscriptionRenewal`, their exclusively stranded repository method/query, and unused `GetSet`, `ListSets`, `UpdateSet`, `DeleteSetsByWorkoutAndExercise`; regenerate SQL code. **Done when:** subscription/account flows still pass and generated query code is fresh. Do not alter active billing policy.
- **F27 — Remove unreachable/no-op server helpers.** Verify and remove `FormatValidationErrors`, `isClientSafeError`, `IsForeignKeyConstraintError`, `IsRLSContextError`, the two unused lifecycle predicates, `aichateval.ScoreResult`, `config.StripeConfigured`, and `ErrUnsupportedStripeEvent`. Replace uses of `normalizeStreamFailure` with the unchanged input before deleting it. **Done when:** no callers remain and affected packages/build/vet pass; preserve useful error classifications still in use.
- **F28 — Exclude test helpers from production.** Move `workout_test_helpers.go` to an appropriate `_test.go` file/package; account for any cross-package test use without exposing test-only production APIs. **Done when:** production builds exclude mocks/database test helpers and all affected tests still compile and run.
- **F29 — Remove confirmed scratch and unused UI code.** Remove `server/cmd/test-jwk` after verifying it is only the placeholder scratch tool; remove `date-picker-base`, `dialog-base`, `error-boundary-test`, `mode-toggle`, and `ui/table` after confirming no importers. Remove corresponding knip ignore entries. **Done when:** builds and both knip modes pass without replacement blanket ignores.
- **F30 — Remove starter-template leftovers.** Remove `reportWebVitals.ts`, its no-callback call in `main.tsx`, `web-vitals`, unused `src/logo.svg`, `.cta.json` and the commented block in `lib/utils.ts`; update lockfiles/references. **Done when:** the production build and client checks pass with no broken assets.
- **F31 — Stop running the same Go tests twice in CI.** Consolidate the `go test -short ./...` and full-suite steps in `.github/workflows/test.yml` when their coverage/environment is equivalent, retaining serialized `go test -p 1 -v ./...` for shared-database safety. Preserve separate RLS/runtime-role checks. **Done when:** CI still exercises unit, integration and runtime-role behavior once at the appropriate stage; do not trade coverage for speed.
- **F32 — Remove excessive per-set insertion logs.** In `repository_utils.go`, replace info-per-set/exercise chatter with a bounded operation summary or appropriate debug logging, preserving actionable failures and safe fields. **Done when:** a large workout produces bounded info logs and failed writes remain diagnosable. Bulk SQL insertion is a separate gated optimization (G04).

## Implementation guidance — gated work packages

These are real observations in the source, but they do not establish that a particular rewrite or deletion is safe.

| ID | Finding | Evidence/decision required and proposed next step |
| --- | --- | --- |
| G01 | Missing migration 00029; 00031 creates what 00032 drops; 00004 has no down step | Inspect history and deployed migration versions. Numbering gaps and later removal are not themselves defects. Do not invent 00029, renumber history, or delete applied migrations. Decide whether rollback support is required; document intentional gaps/irreversible steps or add a forward repair only for a demonstrated schema defect. |
| G02 | Four workout type sets and request → internal → request → internal → Postgres conversions | After F01/F02/F17, map actual boundary differences. Remove proven redundant conversions first; preserve distinct create/update/draft semantics and persistence types where useful. Accept only if behavior stays stable and readers need fewer representations, not merely fewer files. |
| G03 | Helpers duplicated across four packages with different trimming | Inventory callers and expected whitespace semantics. Consolidate only identical behavior; do not silently change input normalization. Approve a shared boundary only if it hides real complexity. |
| G04 | Per-set SQL insertion | Measure query count/latency for representative large workouts. If material, use a transaction-safe bulk insert retaining ordering, IDs, validation, RLS and atomic rollback. Compare results and timings to the old path. |
| G05 | Chat polls about five queries/250 ms, plus ownership polling; every chunk rewrites full text in its own transaction | Benchmark concurrent streams against the production/preview pool sizes (15/3). First consider bounded chunk coalescing and shared poll reads; notifications or a new transport need an explicit design. Preserve reconnect cursor semantics, authorization/revocation, stop behavior, crash recovery and final flush. Set measurable latency/query/pool targets before implementation. |
| G06 | Three redundant-prefix indexes; three nearly identical metrics queries/mappers | Check full index definitions, uniqueness, predicates, FK needs and query plans before a forward migration drops indexes. Compare all metrics query semantics and plans before parameterizing them; retain result ordering, numeric precision and useful index access. Benchmark before/after. |
| G07 | jsdom startup is 38% of a reported 188-second test run | Measure current timings. Move genuinely DOM-free tests to a Node environment using explicit test-project/include rules; retain jsdom for browser-dependent behavior. Do not weaken isolation just to improve speed. Full suite and discovery must remain equivalent. |
| G08 | Training-evidence service/query reachable only from tests; six unreferenced server scripts; two SQL files differ only by title | Ask whether the evidence service is parked intentionally and whether scripts are manually used runbooks. Inventory exact paths and external references. Delete only confirmed retired assets; preserve useful tests and docs or migrate their coverage first. |
| G09 | Untracked `server/api.exe`, `server/bin/`, `.setenv.sh`, Postgres 16 data and two backup directories | **Explicit deletion approval required.** Inventory paths/size without exposing secrets; confirm ownership and backups. Never delete database/backup/env files as ordinary code cleanup. |

## Implementation guidance — deferred decisions

All M/L findings are retained below. Select a decision before promoting them to ready work. Review security, billing and data-integrity entries first even though their confidence is lower.

### Data and workout behavior

- **D01 (M): Empty workouts after exercise deletion.** Decide whether to retain an empty workout, prevent deletion of its last set, or remove it explicitly. Keep this separate from F07's unambiguous missing-ID behavior.
- **D02 (M): Workout list sorts by creation rather than workout date.** Choose intended default chronology, direction and tie-breaker; align server/client and cover backdated entries.
- **D03 (M): Mixed day/time-zone boundaries.** Choose the canonical user-zone policy for activity maps, metrics, chat current date, draft dates and tool dates; preserve date-only values and test midnight/DST behavior.
- **D04 (M): Delete account leaves Stack Auth identity.** Decide whether this is app-data reset or identity deletion; define reauthentication, external-provider failure recovery and messaging before changing it.
- **D05 (M): Unbounded reps overflow numeric calculations.** Reproduce boundary values; choose consistent accepted reps/weight limits or wider numeric storage/calculation types. Reject invalid input cleanly rather than allowing a later read failure.
- **D06 (M): Navigation during render in both workout pages.** Confirm out-of-range behavior under the router/React version; move redirect to a route guard or effect if needed, covering removal/reordering of exercises.
- **D07 (L): First-run name lacks trim/length validation.** Confirm onboarding policy; then align client/server trimming and the 256-character limit if intended.

### AI chat

- **D08 (M): Failed generation start leaves a streaming run.** Reproduce provider-start failure; terminate the run and clear busy state exactly once, without racing a valid generation.
- **D09 (M): Provider failure consumes a trial prompt.** Decide when quota is charged and whether/how to refund; make retries/idempotency safe before implementation.
- **D10 (M): Recovery ends at 55 seconds versus 75-second server generation.** Align recovery with the complete server lifecycle, not just one timeout; verify slow success, stop and stale-run cases.
- **D11 (L + F11 policy follow-up): Model output uses full create-workout request; stored drafts revalidate on reads; substring quality checks and repeated repair can fail the turn.** Decide a model-facing draft schema, catalog-ID resolution, versioned stored-draft compatibility, bounded context selection/summarization policy and graceful repair fallback. Keep the immediate prompt/context safety bound in F11 independent.

### Security, billing and operations — priority user review

- **D12 (M): Public `/metrics` may be open without credentials.** Confirm exposure and monitoring requirements; prefer a private listener or explicit authentication/deny-by-default configuration if unintended.
- **D13 (M): Static handler may list directories.** Reproduce `/assets/` and other directories; disable listing while preserving SPA fallback and asset serving if confirmed.
- **D14 (M): JWT issuer/audience are not validated.** Check Stack Auth's actual claims and trust contract; configure expected issuer/audience and negative-token cases without breaking valid sessions.
- **D15 (M): Recheckout grants repeated trials/prompts.** Decide whether trials are once per identity/customer/subscription; persist eligibility and verify cancellation/resubscription and concurrent checkout behavior.
- **D16 (M): Stripe webhook error classification/retries.** Separate transient database failures from permanently unresolvable events; define safe acknowledgement, logging and retry/idempotency policy.
- **D17 (M): Deploy appends `?` to database URLs with existing queries.** Confirm current secret shape without printing it; use URL-aware parameter merging and cover existing query/encoded values.
- **D18 (M): Substring-based error sanitizing misclassifies errors.** Replace with typed/code-based mappings once the current externally visible error contract is understood.
- **D19 (L): Unpaid subscriptions survive account deletion.** Confirm intended billing cleanup policy and subscription states; coordinate with D04 and ensure retryable provider cleanup.
- **D20 (L): Panic log always names GOOGLE_API_KEY.** Use provider-aware diagnostics after confirming runtime configuration paths; never include key values.

### API, schema and client setup

- **D21 (M): Chat stream/stop bypass shared fetch/auth/401 handling.** Preserve streaming/abort semantics while deciding how to share authentication, base URL and unauthorized-session handling.
- **D22 (M): Roughly 400 lines of chat parsers duplicate generated types.** Distinguish runtime validation from compile-time types; remove only duplication that does not eliminate validation of untrusted data.
- **D23 (M): Catalog IDs exist in JSON, migration check, schema and muscle-role map.** Choose an authoritative source and generation/consistency checks; do not rewrite historical migration snapshots to chase current catalog changes.
- **D24 (M): Demo/API differences.** Decide which are intentional: 10 versus 3 recent sets, zero- versus one-based ordering, case-insensitive rename checks and simulated 50–200 ms delays. Align only intended API parity and cover demo flows.
- **D25 (M): Exercise goals use global localStorage and hard-coded pounds.** Define user scoping, sign-out behavior, migration of existing goals and preferred units; never silently attach another user's data to a new account.
- **D26 (L): Edit uses window.confirm while new uses a dialog.** Decide a consistent destructive-action interaction and accessibility expectations.

### Server design and CI

- **D27 (M): Gemini/trace settings bypass validated config.** Inventory defaults and secret handling; move to validated startup config if behavior can be preserved.
- **D28 (M): Alert rule copies differ by one line.** Determine whether deployment-specific differences are intentional; generate from one source only if it simplifies ownership.
- **D29 (M): Feature access exposes internal granted_by/note.** Confirm whether these are intended user-facing fields and whether sensitive content can be present; restrict the public DTO if not.
- **D30 (L): Error casing, identical response.Error/ErrorResponse, wrong muscle-contribution error message, beta/non_beta cohort names.** Review externally consumed contracts/dashboard dependencies before standardizing names/messages or types.
- **D31 (M): Tests only run on PRs to main; deployment has no test gate.** Confirm release policy and branch protections; decide whether deploy must depend on verified checks for the exact commit.
- **D32 (M): Test/deploy do not verify Goose version.** Align tool-version checks with preview after confirming the intended pinned version and cache behavior.
- **D33 (L): Windows CRLF causes go mod tidy -diff noise.** Reproduce without rewriting unrelated files; decide targeted `.gitattributes` normalization. Treat the source's old-staticcheck issue as toolchain compatibility, not a product bug.

### Performance and remaining cleanup

- **D34 (M): No index on historical_1rm_source_workout_id.** Measure plans on representative data before proposing a forward index migration.
- **D35 (M): Per-request user lookup and per-checkout set_config round trip.** Measure cost; retain RLS isolation and fresh authorization. Do not cache identities or reuse session state blindly.
- **D36 (M): Three logs per 401 and all 4xx logged as errors.** Agree status/severity ownership and retain security/audit visibility while reducing duplicates.
- **D37 (M): stripe.Key is written on each request.** Confirm SDK usage and concurrency risk; use configured client instances or startup initialization as appropriate, with race checks.
- **D38 (L): Draft writes on every keystroke and parses on every workout-list render.** Profile before optimizing; any debounce must flush before navigation/reload and preserve crash recovery.
- **D39 (L): One DB query per telemetry event.** Measure throughput; define durability/privacy requirements before batching or sampling.
- **D40 (L): Static files opened twice per request.** Confirm actual hot-path cost before consolidating file lifecycle handling.
- **D41 (L): Duplicated Makefile setup, unused GOGET, redundant gitignore entry.** Consolidate only where it reduces maintenance without hiding meaningful dev/test differences.
- **D42 (M/L): workout/validator.go, two exercise ID-validation structs, always-true nil checks in routes.go.** Prove reachability and invariants before removing guards or validators.
- **D43 (M): Three copied upstream TanStack docs.** Decide whether these are intentional offline references; archive or replace with versioned links if not.
- **D44 (M): output/ has 2.3 MB of one-off artifacts referenced by one doc.** Check evidence/retention needs and repair links before removal; no blind cleanup.

## Verification and closeout rules

1. Reproduce each behavioral issue through the caller-facing boundary before fixing it. Follow the repository's test-audit guidance when writing/changing tests; avoid tests tied only to private call order.
2. Run the narrowest meaningful tests first. Client scripts confirmed in `client/package.json`: `bun run test -- <test-path>`, `bun run tsc`, `bun run lint`, `bun run fmt:check`, `bun run knip`, `bun run knip:production` (run from `client/`). Use the full suite/build for shared configuration or broad cleanup.
3. Server checks: focused `go test` package(s), `go build ./...`, `go vet ./...`, and `make test-fast` from `server/` as relevant. Database-backed changes require an isolated disposable database and full relevant integration tests, not just mocks or fast tests. Keep shared-database package execution serialized, matching CI.
4. Before starting services/tests in linked worktrees, follow `docs/worktrees.md` for separate ports/database configuration. Read local setup before using `make test-integration`; do not migrate a shared or production database accidentally.
5. Contract changes: regenerate through `bun run --cwd client api:generate`, review the generated diff, then run `bun run --cwd client api:check`. SQL changes: regenerate using the documented SQLC workflow and inspect generated code.
6. Verify PWA, cache/navigation, proxy headers and cross-origin behavior in representative runtime conditions. Record any missing browser, proxy or database environment explicitly rather than marking the task complete. Do not install/switch browser-control surfaces silently.
7. For performance changes, record before/after query counts, latency and resource use under equivalent workloads; passing tests alone does not prove an optimization.
8. Record commit/PR, checks and any deployment-only verification in the corresponding A-ID entry above. Do not mark a fix complete while its required verification is blocked. Ship/release and destructive cleanup are separate from this planning document.

## Non-remediation requests from the source

- An autoresearch note for the staticcheck/CRLF tooling issues is optional and separate; create one only after confirming a reusable lesson and checking for existing notes.
- Publishing the list to an external page is not needed for this local Markdown tracker and has not been done.
