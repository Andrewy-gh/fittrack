# AI Chat Eval Baseline - 2026-07-08 4965cda

Official Phase 1 baseline for the AI chat scenario sweep on commit `4965cdad4d9e2eac2923d0b1f84c9bdb66bf71a8` (`4965cda`), the squash merge of PR #231 on `main`.

## Run Metadata

- Date: 2026-07-08 local time; report `generated_at` timestamps are UTC on 2026-07-09.
- Model: `googleai/gemini-2.5-flash`
- Mode: `two_turn`
- Default scenario set: 20 scenarios
- Fixture scenario set: 22 scenarios with `-with-data-fixtures`
- Server precondition: `git status --porcelain -- server/` was empty before the sweeps.
- Working tree note: the worktree was clean before the sweeps.
- Sweep ledger path: `server/tmp/ai-chat-scenario-sweeps/fittrack-ai-chat-scenario-sweep-runs.jsonl`
- Default report path: `server/tmp/ai-chat-scenario-sweeps/fittrack-ai-chat-scenario-sweep.json`
- Copied reports:
  - `server/evals/baselines/2026-07-08-4965cda-gemini-2.5-flash/baseline-two-turn.json`
  - `server/evals/baselines/2026-07-08-4965cda-gemini-2.5-flash/fixtures-two-turn.json`

## Commands

Run from `server/`:

```powershell
go run ./cmd/ai-chat-scenario-sweep -mode two_turn -scenario-delay 15s -timeout 30m
```

The default report was copied immediately to `baseline-two-turn.json`.

```powershell
go run ./cmd/ai-chat-scenario-sweep -mode two_turn -with-data-fixtures -scenario-delay 15s -timeout 45m
```

The default report was copied immediately to `fixtures-two-turn.json`.

Behavior-label reruns, also from `server/`:

```powershell
$env:FITTRACK_AI_CHAT_SWEEP_OUT='tmp/ai-chat-scenario-sweeps/baseline-two-turn-4965cda-rerun-prompt-03-prompt-18.json'; go run ./cmd/ai-chat-scenario-sweep -mode two_turn -scenarios prompt-03,prompt-18 -scenario-delay 15s -timeout 15m
```

```powershell
$env:FITTRACK_AI_CHAT_SWEEP_OUT='tmp/ai-chat-scenario-sweeps/fixtures-two-turn-4965cda-rerun-prompt-18.json'; go run ./cmd/ai-chat-scenario-sweep -mode two_turn -with-data-fixtures -scenarios prompt-18 -scenario-delay 15s -timeout 15m
```

## Summary

| Sweep | Pass | Fail | Unscored | Operational errors |
| --- | ---: | ---: | ---: | ---: |
| `baseline-two-turn.json` | 18 | 2 | 0 | 0 |
| `fixtures-two-turn.json` | 21 | 1 | 0 | 0 |

No operational-error reruns or splices were needed.

## Default Two-Turn Results

| id | score_status | reason for non-pass |
| --- | --- | --- |
| prompt-01 | pass |  |
| prompt-02 | pass |  |
| prompt-03 | fail | assistant behavior ended in an error before meeting the expected outcome |
| prompt-04 | pass |  |
| prompt-05 | pass |  |
| prompt-06 | pass |  |
| prompt-07 | pass |  |
| prompt-08 | pass |  |
| prompt-09 | pass |  |
| prompt-10 | pass |  |
| prompt-11 | pass |  |
| prompt-12 | pass |  |
| prompt-13 | pass |  |
| prompt-14 | pass |  |
| prompt-15 | pass |  |
| prompt-16 | pass |  |
| prompt-17 | pass |  |
| prompt-18 | fail | expected the first turn to ask the user to choose one workout or session |
| prompt-19 | pass |  |
| prompt-20 | pass |  |

## Fixture Two-Turn Results

| id | score_status | reason for non-pass |
| --- | --- | --- |
| prompt-05 | pass |  |
| prompt-11 | pass |  |
| prompt-12 | pass |  |
| prompt-14 | pass |  |
| prompt-15 | pass |  |
| prompt-16 | pass |  |
| prompt-17 | pass |  |
| prompt-18 | fail | expected the first turn to ask the user to choose one workout or session |
| data-01 | pass |  |
| data-02 | pass |  |
| data-03 | pass |  |
| data-04 | pass |  |
| data-05 | pass |  |
| data-06 | pass |  |
| data-07 | pass |  |
| data-08 | pass |  |
| data-09 | pass |  |
| data-10 | pass |  |
| data-11 | pass |  |
| profile-01 | pass |  |
| profile-02 | pass |  |
| profile-03 | pass |  |

## Spliced Reruns

None. Both full sweeps completed with `operational_error_count: 0`, so the checked-in reports are direct copies of the full-run outputs.

## Known Noise And Current Failure Labels

- `prompt-03` failed in the default sweep and failed again in the targeted rerun with `operational_error_count: 0`. The behavior matches the 2026-07-07 `d310b20` failure class: the first turn asked a follow-up, then the inner draft path rejected the generated workout because "Dumbbell Goblet Squat" conflicted with the knee-pain context. Treat this as a consistent, pre-existing model/draft-quality failure, not a new 4965cda regression.
- `prompt-18` failed in the default sweep and failed again in the default targeted rerun with `operational_error_count: 0`. This also matches the 2026-07-07 consistent failure class: the assistant narrows away from meal planning, but does not satisfy the scorer's expected first-turn "choose one workout or session" behavior. Treat this as a consistent, pre-existing behavior failure.
- `prompt-18` failed in the fixture sweep but passed the fixture targeted rerun with `operational_error_count: 0`. Label the fixture-pack failure as flaky model behavior for this run; keep the official failed result in the artifact.
- `prompt-13` passed in the default sweep. This is an improvement from the 2026-07-07 default baseline, where it failed after empty-response noise.

## Run Anomalies

- Default sweep: retry-guard warnings on `prompt-08` and `prompt-15` for empty responses with no tool activity. Both scenarios still passed in the recorded artifact.
- Fixture sweep: retry-guard warnings on `data-02` and `data-04` for empty responses with no tool activity. Both scenarios still passed in the recorded artifact.
- No rate-limit wave was observed in command output.

## Comparison Against 2026-07-07 d310b20

- Default pack: 18/20 pass on `4965cda` versus 16/19 pass on `d310b20`. The shared 19 scenarios are mostly comparable; `prompt-13` changed from fail to pass, and `prompt-20` is new.
- Fixture pack: 21/22 pass on `4965cda` versus 24/27 pass on `d310b20`. The raw counts are not directly comparable because fixture mode now excludes BaseOnly ask-first prompt scenarios.
- Fixture prompt removals from the comparable set: `prompt-01`, `prompt-02`, `prompt-03`, `prompt-04`, `prompt-06`, `prompt-07`, `prompt-08`, `prompt-09`, `prompt-10`, `prompt-13`, and `prompt-19`.
- Fixture additions: `data-09`, `data-10`, `data-11`, `profile-01`, `profile-02`, and `profile-03`; all six passed.
- Among the 16 shared fixture scenarios, no score statuses changed from the 2026-07-07 baseline.
