# AI Chat Eval Baseline - 2026-07-07 d310b20

Official Phase 1 baseline for the AI chat scenario sweep on commit `d310b20f326ed8016e1c4d6321a4d29768a252ae` (`d310b20`).

## Run Metadata

- Date: 2026-07-07
- Model: `googleai/gemini-2.5-flash`
- Mode: `two_turn`
- Default scenario set: 19 scenarios
- Fixture scenario set: 27 scenarios with `-with-data-fixtures`
- Server precondition: `git status --porcelain -- server/` was empty before the sweeps.
- Working tree note: unrelated uncommitted files existed under `client/`; they were not touched.
- Sweep ledger path: `server/tmp/ai-chat-scenario-sweeps/fittrack-ai-chat-scenario-sweep-runs.jsonl`
- Temporary reports:
  - `server/tmp/ai-chat-scenario-sweeps/baseline-two-turn-d310b20.json`
  - `server/tmp/ai-chat-scenario-sweeps/fixtures-two-turn-d310b20.json`

## Commands

Run from `server/`:

```powershell
$env:FITTRACK_AI_CHAT_SWEEP_OUT='tmp/ai-chat-scenario-sweeps/baseline-two-turn-d310b20.json'; go run ./cmd/ai-chat-scenario-sweep -mode two_turn -scenario-delay 30s -timeout 45m
```

```powershell
$env:FITTRACK_AI_CHAT_SWEEP_OUT='tmp/ai-chat-scenario-sweeps/fixtures-two-turn-d310b20.json'; go run ./cmd/ai-chat-scenario-sweep -mode two_turn -with-data-fixtures -scenario-delay 30s -timeout 60m
```

## Summary

| Sweep | Pass | Fail | Unscored | Operational errors |
| --- | ---: | ---: | ---: | ---: |
| `baseline-two-turn.json` | 16 | 3 | 0 | 0 |
| `fixtures-two-turn.json` | 24 | 3 | 0 | 0 |

No operational-error reruns were needed.

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
| prompt-13 | fail | expected the second turn to generate a structured draft |
| prompt-14 | pass |  |
| prompt-15 | pass |  |
| prompt-16 | pass |  |
| prompt-17 | pass |  |
| prompt-18 | fail | expected the first turn to ask the user to choose one workout or session |
| prompt-19 | pass |  |

## Fixture Two-Turn Results

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
| prompt-09 | fail | expected one follow-up question before generating |
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
| data-01 | pass |  |
| data-02 | pass |  |
| data-03 | pass |  |
| data-04 | pass |  |
| data-05 | pass |  |
| data-06 | pass |  |
| data-07 | pass |  |
| data-08 | pass |  |

## Spliced Reruns

None. Both full sweeps completed with `operational_error_count: 0`.

## Known Noise And Current Failure Labels

- `prompt-18` failed in both sweeps. This matches the historical consistent failure and is not treated as a desired behavior contract.
- `prompt-13` failed only in the default sweep after an empty-response retry warning and a null text-only second turn. Label this as flaky model/provider output for this run, not a Phase 1 contract.
- `prompt-14` passed in both sweeps; the historical flakiness did not appear in this baseline.
- `prompt-16` passed in both sweeps; the historical occasional draft-model hallucination did not appear in this baseline.
- `prompt-03` failed in both sweeps with a scored `error` status, not an `operational_error`: the inner draft model chose "Goblet Squat" despite the knee-pain context and the workout quality validator rejected it after the repair retry. This is the same pre-existing inner-draft-model quality flakiness class as the historical `prompt-16` hallucination; that code path was not modified in this commit.
- `prompt-09` failed only in the fixture sweep by generating immediately instead of asking one follow-up question. Treat it as a current baseline failure unless later evidence proves it is pre-existing noise.

## Run Anomalies

- Default sweep: one retry-guard warning on `prompt-13` for an empty response with no tool activity.
- Fixture sweep: retry-guard warnings on `prompt-05` and `prompt-11` for empty responses with no tool activity.
- No rate-limit wave was observed in command output.
