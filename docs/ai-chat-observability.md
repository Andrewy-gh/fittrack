# AI Chat Observability and Rollout Gate

This repo records client-observed AI chat outcomes into Prometheus via `ai_chat_client_outcomes_total{category,outcome,stage,cohort}`.

Outcome taxonomy:

- `stream`: `completed`, `server_error`, `transport_ended_pre_terminal`, `client_aborted`
- `recovery`: `recovered_completed`, `recovered_failed`, `recovery_timeout`, `recovery_aborted`
- `load`: `load_completed`, `load_failed`, `load_aborted_stale`
- `ux`: `failure_toast_shown`, `failure_toast_suppressed_due_to_successful_recovery`

For `stream` events, `stage` is one of `pre_start`, `post_start`, or `terminal`. Other categories use `stage="n/a"`.

## Dashboard Panels

True stream failure rate:

```promql
sum(rate(ai_chat_client_outcomes_total{category="stream",outcome="server_error",cohort="beta"}[15m]))
/
clamp_min(sum(rate(ai_chat_client_outcomes_total{category="stream",outcome=~"completed|server_error|transport_ended_pre_terminal|client_aborted",cohort="beta"}[15m])), 0.001)
```

User or navigation abort rate:

```promql
sum(rate(ai_chat_client_outcomes_total{category="stream",outcome="client_aborted",cohort="beta"}[15m]))
/
clamp_min(sum(rate(ai_chat_client_outcomes_total{category="stream",outcome=~"completed|server_error|transport_ended_pre_terminal|client_aborted",cohort="beta"}[15m])), 0.001)
```

Recovery success rate after interrupted streams:

```promql
sum(rate(ai_chat_client_outcomes_total{category="recovery",outcome="recovered_completed",cohort="beta"}[30m]))
/
clamp_min(sum(rate(ai_chat_client_outcomes_total{category="recovery",outcome=~"recovered_completed|recovered_failed|recovery_timeout",cohort="beta"}[30m])), 0.001)
```

Pre-start disconnect rate:

```promql
sum(rate(ai_chat_client_outcomes_total{category="stream",outcome="transport_ended_pre_terminal",stage="pre_start",cohort="beta"}[15m]))
/
clamp_min(sum(rate(ai_chat_client_outcomes_total{category="stream",outcome=~"completed|server_error|transport_ended_pre_terminal|client_aborted",cohort="beta"}[15m])), 0.001)
```

Recovery timeout rate:

```promql
sum(rate(ai_chat_client_outcomes_total{category="recovery",outcome="recovery_timeout",cohort="beta"}[30m]))
/
clamp_min(sum(rate(ai_chat_client_outcomes_total{category="recovery",outcome=~"recovered_completed|recovered_failed|recovery_timeout",cohort="beta"}[30m])), 0.001)
```

Client-visible error toast rate:

```promql
sum(rate(ai_chat_client_outcomes_total{category="ux",outcome="failure_toast_shown",cohort="beta"}[15m]))
/
clamp_min(sum(rate(ai_chat_client_outcomes_total{category="stream",outcome=~"completed|server_error|transport_ended_pre_terminal|client_aborted",cohort="beta"}[15m])), 0.001)
```

Beta vs non-beta comparison:

```promql
sum by (cohort, outcome) (rate(ai_chat_client_outcomes_total{category="stream"}[15m]))
```

## Alerting Rules

Page on:

- rising `stream=server_error`
- rising `recovery=recovery_timeout`
- rising `load=load_failed`
- degraded recovery success rate using only `recovered_completed|recovered_failed|recovery_timeout`

Do not page on:

- `stream=client_aborted`
- `load=load_aborted_stale`
- `recovery=recovery_aborted`
- `ux=failure_toast_suppressed_due_to_successful_recovery`

## Failure Drills

1. POST succeeds, SSE dies before first `start`, persisted conversation later completes.
Expected signals:
- `stream=transport_ended_pre_terminal,stage=pre_start`
- `recovery=recovered_completed`
- `ux=failure_toast_suppressed_due_to_successful_recovery`

2. Recovery polling is in flight, then the user navigates away, clears chat, or starts a new chat.
Expected signals:
- `recovery=recovery_aborted`
- no `ux=failure_toast_shown`

3. Slow initial conversation GET resolves after the user switches or clears.
Expected signals:
- `load=load_aborted_stale`
- no paging

4. Recovery polling times out.
Expected signals:
- `recovery=recovery_timeout`
- `ux=failure_toast_shown`

5. Interrupted SSE plus overlapping navigation or reconnect churn.
Expected signals:
- interruptions show up as `stream=transport_ended_pre_terminal`
- user-driven leaves show up as `stream=client_aborted`
- rollout remains gated on recovery health, not aggregate error counts

## Rollout Gate

Do not widen access on aggregate “chat error rate” alone.

Widen only when all are true:

- `stream=server_error` stays within the agreed beta threshold
- `recovery=recovered_completed` remains healthy relative to `recovered_failed|recovery_timeout`
- `load=load_failed` is stable
- `ux=failure_toast_shown` is not climbing faster than underlying true failures

Known limitation:

- A literal server-side `start` write failure still aborts the prepared run immediately in the current backend. The drill above is fully represented in telemetry from the client side, but the server start-write path itself has not been rewritten in this phase.
