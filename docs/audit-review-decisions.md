# Audit decisions for user review

Related tracker: [audit-remediation-plan.md](audit-remediation-plan.md).
Working branch: `fix/audit-remediation`.

Originally **64 findings required review**: 13 high-confidence gated findings and 51 medium/low-confidence deferred findings. A103/A104 removal is now approved; A094 has conditional cleanup approval; A088 is awaiting a decision after explanation. The inventory below preserves all original decision items. You do not need to review the 42 ready findings before agents begin. Every implemented behavior still requires tests/evidence and independent code review.

## Review these decisions first

Record answers by A-ID below or reply in chat; the CHIEF will transfer your decisions into the main tracker. Recommendations are proposals, not approvals.

| Priority | IDs | Decision needed | Suggested starting point (not yet approved) |
| --- | --- | --- | --- |
| 1 | A014, A037 | Should Delete account erase the authentication identity too, and what happens to an unpaid subscription? | Treat deletion as full account closure with explicit confirmation and retryable external cleanup; agree retention/legal requirements first. |
| 2 | A033 | Is a trial one-time per customer/identity, or renewed on each subscription? | One trial per identity/customer, not per checkout. |
| 3 | A020 | Should a provider failure consume a trial prompt? | Do not charge when no useful response is delivered; define partial-response handling and safe refunds. |
| 4 | A008, A012 | Keep workouts after their last exercise is deleted? Sort by workout date or creation time? | Keep an existing empty workout distinct from a missing one; sort by workout date with a stable tie-breaker. |
| 5 | A013 | Which time zone defines a training day across charts, drafts and chat? | User's configured local time zone consistently; preserve date-only workout values. |
| 6 | A015, A017 | What reps/weight/name limits should be accepted? | Bound inputs to safe storage/calculation ranges; use the existing trimmed 256-character exercise-name convention. Exact numeric limits need engineering evidence. |
| 7 | A019, A021–A024 | Approve fixing failed-start busy state and recovery timeouts? What should happen to bad AI drafts or failed repair attempts? | Terminal failure on failed start; recovery covers server duration; preserve old drafts and offer recoverable errors instead of losing the conversation. |
| 8 | A030–A032, A063 | Should metrics be private, directory listings disabled, JWT issuer/audience enforced and internal access notes hidden? | Yes where deployment/provider evidence supports it; agents must confirm monitoring/auth contracts before implementation. |
| 9 | A056–A058 | Must demo match the API? Should goals be per-user with preferred units? Standardize confirmation dialogs? | Align semantic demo behavior; keep optional latency simulation explicit; isolate goals by user and use preferred units. |
| 10 | A068–A069 | Must deployment wait for checks on the exact commit, and should Goose be version-verified everywhere? | Yes; agree any emergency-release exception separately. |
| 11 | A088, A094, A103–A104 | Keep parked training-evidence work, manual scripts, copied docs and old verification artifacts? | Inventory purpose/references first; archive or delete only items confirmed retired. |
| 12 | A106 | May any binaries, environment files, old PostgreSQL data or backup directories be deleted? | No deletion by default. Require exact-path approval, ownership check and verified backups for data. |

## Technical gates the CHIEF can investigate, but not silently approve

- **A045:** Determine whether migration gaps/history are intentional. Never renumber or rewrite applied migrations.
- **A059–A060, A074:** Assess type/conversion/helper consolidation without changing validation or trimming semantics.
- **A071–A073:** Measure chat polling/chunk persistence and set insertion; propose changes preserving recovery, authorization and atomicity. A073 logging cleanup can proceed separately.
- **A075–A076, A078–A080:** Validate index/query/environment optimizations using query plans and timings, not assumptions.
- **A034–A036, A042–A043, A061–A062, A081–A082:** Establish operational/API/concurrency contracts before choosing fixes.

These investigations do not authorize changing deferred behavior. Any consequential decision returns to you. Low-priority cosmetic cleanup can remain deferred while correctness work proceeds.

## Complete gated/deferred inventory

This is a decision worksheet, not a second implementation tracker. Main A-ID entries remain authoritative for progress. Every gated/deferred finding appears once below. A decision may be **Approve proposed approach**, **Keep current behavior**, **Investigate first**, or **Defer**, with intended behavior/constraints.

| ID | Gate | Original finding (nested details remain in main tracker) | Your decision |
| --- | --- | --- | --- |
| [A008](audit-remediation-plan.md#a008) | Deferred | M A workout whose sets were all removed by deleting an exercise looks the same as a missing one. | Pending |
| [A012](audit-remediation-plan.md#a012) | Deferred | M The workout list is re-sorted by creation time on the client (workouts-filters.ts:24) while the server sorts by workout date, so backdated workouts appear newest. | Pending |
| [A013](audit-remediation-plan.md#a013) | Deferred | M Day boundaries use UTC or server time in the activity map, metrics history, the chat "current date", draft dates and chat tool dates. The assessment endpoints take an explicit time zone. | Pending |
| [A014](audit-remediation-plan.md#a014) | Deferred | M "Delete account" deletes app data and signs out but never deletes the Stack Auth identity (account-settings-page.tsx:85). Signing in again creates a fresh empty account. This may be intended. | Pending |
| [A015](audit-remediation-plan.md#a015) | Deferred | M Reps have no upper bound on the server. Extreme reps × weight overflow NUMERIC(10,1) and numeric(8,2) casts and return 500 on save or read. | Pending |
| [A016](audit-remediation-plan.md#a016) | Deferred | M navigate() is called during render for an out-of-range exercise index in both workout pages. | Pending |
| [A017](audit-remediation-plan.md#a017) | Deferred | L The first-run exercise name has no length limit or trim, unlike the 256 limit elsewhere. | Pending |
| [A019](audit-remediation-plan.md#a019) | Deferred | M If generation fails to start, the run is left "streaming" with no abort (aichat/handler.go:442). The conversation stays busy until the stale timeout. | Pending |
| [A020](audit-remediation-plan.md#a020) | Deferred | M A trial prompt is consumed before generation, so a provider failure still costs the user one. | Pending |
| [A021](audit-remediation-plan.md#a021) | Deferred | M The client recovery poll gives up at 55 s (ai-chat.ts:211) while the server may legitimately generate for 75 s. | Pending |
| [A022](audit-remediation-plan.md#a022) | Deferred | L The model's structured output type is the full create-workout request, including catalog_id and recommendations. A hallucinated value passes draft validation and fails at save with a 500. | Pending |
| [A023](audit-remediation-plan.md#a023) | Deferred | L Stored drafts are re-validated on every read. Tightening validation later would make old conversations fail to load. | Pending |
| [A024](audit-remediation-plan.md#a024) | Deferred | L Draft quality checks are substring-based ("back" matches "coming back from a break"), and a second failed repair fails the whole turn. | Pending |
| [A030](audit-remediation-plan.md#a030) | Deferred | M /metrics is also mounted on the public port and is open when no credentials are set (routes.go:46). | Pending |
| [A031](audit-remediation-plan.md#a031) | Deferred | M The static handler serves directory listings for paths such as /assets/ (routes.go:184). Not run. | Pending |
| [A032](audit-remediation-plan.md#a032) | Deferred | M Tokens are checked for signature and expiry only, not issuer or audience (auth/middleware.go:176). | Pending |
| [A033](audit-remediation-plan.md#a033) | Deferred | M Every checkout grants a new 7-day trial and a new prompt allowance (billing/service.go:100), so cancel-and-resubscribe repeats the trial. | Pending |
| [A034](audit-remediation-plan.md#a034) | Deferred | M The Stripe webhook returns 400 for database errors, and events that can never resolve are retried for days. | Pending |
| [A035](audit-remediation-plan.md#a035) | Deferred | M fly-deploy.yml:40 appends ?default_query_exec_mode=… to the database URL, which breaks if the secret already has a query string. | Pending |
| [A036](audit-remediation-plan.md#a036) | Deferred | M Error-message sanitising uses substring matches ("token", "table", "connection"), so unrelated errors can be rewritten to "unauthorized". | Pending |
| [A037](audit-remediation-plan.md#a037) | Deferred | L An unpaid subscription is not cancelled on account deletion. | Pending |
| [A038](audit-remediation-plan.md#a038) | Deferred | L The runtime panic log always names GOOGLE_API_KEY (runtime.go:76). | Pending |
| [A042](audit-remediation-plan.md#a042) | Deferred | M Chat streaming and stop use their own fetch, auth headers and base URL, bypassing the shared 401 handling. | Pending |
| [A043](audit-remediation-plan.md#a043) | Deferred | M About 400 lines of chat parsers duplicate generated types. | Pending |
| [A045](audit-remediation-plan.md#a045) | Gated | H Migration 00029 is missing, 00031 creates a table that 00032 drops, and the down step of 00004 is empty. | Pending |
| [A046](audit-remediation-plan.md#a046) | Deferred | M Catalog IDs live in four places: the JSON catalog, a migration check, schema.sql and the muscle-role map. | Pending |
| [A056](audit-remediation-plan.md#a056) | Deferred | M Demo mode differs from the API: | Pending |
| [A057](audit-remediation-plan.md#a057) | Deferred | M Exercise goals are stored in one global localStorage key, not per user, and hard-code "lb". | Pending |
| [A058](audit-remediation-plan.md#a058) | Deferred | L The edit page uses window.confirm where the new page uses a dialog. | Pending |
| [A059](audit-remediation-plan.md#a059) | Gated | H The workout request has four parallel type sets (create, update, draft, Postgres). | Pending |
| [A060](audit-remediation-plan.md#a060) | Gated | H Small helpers are re-implemented in four packages with different trimming behaviour. | Pending |
| [A061](audit-remediation-plan.md#a061) | Deferred | M Gemini and trace settings are read straight from the environment, bypassing the validated config. | Pending |
| [A062](audit-remediation-plan.md#a062) | Deferred | M The alert rules are duplicated in ops/prometheus and ops/grafana-cloud, differing by one line. | Pending |
| [A063](audit-remediation-plan.md#a063) | Deferred | M GET /api/features/access returns internal granted_by and note fields to end users. | Pending |
| [A064](audit-remediation-plan.md#a064) | Deferred | L Error-message casing varies between handlers. | Pending |
| [A065](audit-remediation-plan.md#a065) | Deferred | L response.Error and response.ErrorResponse are identical types used interchangeably. | Pending |
| [A066](audit-remediation-plan.md#a066) | Deferred | L The muscle-contribution handler reports "Could not load training assessment". | Pending |
| [A067](audit-remediation-plan.md#a067) | Deferred | L The telemetry cohort labels are still "beta" and "non_beta". | Pending |
| [A068](audit-remediation-plan.md#a068) | Deferred | M Tests run only on pull requests to main, and the deploy has no test gate. | Pending |
| [A069](audit-remediation-plan.md#a069) | Deferred | M The deploy and test workflows do not verify the Goose version, which the preview workflow does. | Pending |
| [A070](audit-remediation-plan.md#a070) | Deferred | L go mod tidy -diff fails on a Windows checkout only because go.sum is checked out with CRLF. | Pending |
| [A071](audit-remediation-plan.md#a071) | Gated | H Live chat is delivered by database polling: about five queries every 250 ms per open stream (service_stream.go:80), plus an ownership check every 250 ms per run. The pool is 15 connections, 3 in preview. | Pending |
| [A072](audit-remediation-plan.md#a072) | Gated | H Each model chunk is its own transaction and rewrites the full message text so far (repository_stream.go:316). | Pending |
| [A073](audit-remediation-plan.md#a073) | Gated | H Sets are inserted one query at a time, with an info log per set and several per exercise (repository_utils.go:71). | Pending |
| [A074](audit-remediation-plan.md#a074) | Gated | H Workout create converts request to internal to request to internal to Postgres types (repository.go:237). | Pending |
| [A075](audit-remediation-plan.md#a075) | Gated | H Three indexes are redundant prefixes of others (schema.sql:322: idx_set_user_id, idx_workout_user_id, idx_exercise_user_id). | Pending |
| [A076](audit-remediation-plan.md#a076) | Gated | H Three near-identical 50-line metrics queries, and three identical mappers, exist where one parameterised query would do. | Pending |
| [A078](audit-remediation-plan.md#a078) | Gated | H Vitest reports that creating jsdom 99 times is 38% of the 188 s client test run. | Pending |
| [A079](audit-remediation-plan.md#a079) | Deferred | M No index exists on exercise.historical_1rm_source_workout_id, which is queried on every workout update and delete. | Pending |
| [A080](audit-remediation-plan.md#a080) | Deferred | M Every API request does a user lookup, and every pool checkout does a set_config round trip. | Pending |
| [A081](audit-remediation-plan.md#a081) | Deferred | M A 401 produces three log lines, and all 4xx responses are logged at error level. | Pending |
| [A082](audit-remediation-plan.md#a082) | Deferred | M stripe.Key is reassigned on every request (billing/service.go:80), an unsynchronised global write. | Pending |
| [A083](audit-remediation-plan.md#a083) | Deferred | L The draft is written to localStorage on every keystroke and parsed on every workout-list render. | Pending |
| [A084](audit-remediation-plan.md#a084) | Deferred | L Each telemetry event costs a database query. | Pending |
| [A085](audit-remediation-plan.md#a085) | Deferred | L Static files are opened twice per request. | Pending |
| [A086](audit-remediation-plan.md#a086) | Deferred | L make dev and make test-db-ready duplicate about 60 lines. | Pending |
| [A088](audit-remediation-plan.md#a088) | Gated | H The training-evidence service and its query are reachable only from tests (training_evidence.go:115). This may be parked on purpose. | Explanation requested; no deletion approved. |
| [A094](audit-remediation-plan.md#a094) | Gated | H Six scripts in server/scripts are referenced by nothing, and two of the SQL files differ only in their title. | Conditional approval: explain purpose; delete only scripts with no useful current purpose. |
| [A095](audit-remediation-plan.md#a095) | Deferred | M workout/validator.go and two redundant ID-validation structs in the exercise package. | Pending |
| [A096](audit-remediation-plan.md#a096) | Deferred | L Nil checks in routes.go that are always true. | Pending |
| [A103](audit-remediation-plan.md#a103) | Deferred | M Three client/docs files are copies of upstream TanStack docs. | Approved: remove copied docs. |
| [A104](audit-remediation-plan.md#a104) | Deferred | M output/ holds 2.3 MB of one-off verification screenshots and JSON, referenced by one doc. | Approved: remove output artifacts and fix references. |
| [A105](audit-remediation-plan.md#a105) | Deferred | L .gitignore has a redundant server/tmp/ai-chat-scenario-sweeps/ entry, and the Makefile's GOGET variable is unused. | Pending |
| [A106](audit-remediation-plan.md#a106) | Gated | H, your call Untracked local leftovers: server/api.exe (55 MB, from August), server/bin/, server/.setenv.sh, the Postgres 16 data directory and two backup directories. | Pending |
