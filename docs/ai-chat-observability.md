# AI Chat Observability and Rollout Gate

This repo records client-observed AI chat outcomes into Prometheus via `ai_chat_client_outcomes_total{category,outcome,stage,cohort}`.

## 2026-06 Observability Decisions

Product goals:

- Keep Fly compute low and preserve `min_machines_running = 0`.
- Prefer free/default Fly-managed retention before adding another vendor or always-on collector.
- Keep prompts, generated workout content, user IDs, workout IDs, emails, and request IDs out of metric labels and alert text.
- Use Discord and email for alerts when alerting is wired; Discord is the preferred push channel.
- Do not block a small private/beta AI chat test on every dashboard, but do gate broader rollout on retained production metrics.

Current decision:

- Use Prometheus-native metrics first, not OpenTelemetry, because the immediate need is retained route and AI-chat health with minimal runtime and cost overhead.
- Use Fly custom metrics for retained dashboards if it can be wired without weakening public metrics auth or keeping Machines awake.
- Keep the public `/metrics` endpoint Basic Auth protected.
- Add a separate internal-only metrics listener/port for Fly scraping because Fly's documented `[metrics]` config only supports `port` and `path`; Basic Auth or custom scrape headers are not documented.
- Treat Fly retention/cost details as best-effort platform behavior, not a guaranteed 30-day free retention contract, unless current Fly docs state otherwise.
- Validate in production or staging that Fly custom metrics scraping does not wake stopped Machines or materially increase compute.
- Plan alerting separately if Fly managed Grafana does not support contact points/rules. Grafana Cloud Free or Alertmanager are fallback paths for Discord/email.

Implementation steps:

1. Rename ambiguous AI metric wording from `runtime` to `provider` or `model` so it is not confused with the Go runtime.
2. Add an internal metrics listener on a non-public port that serves Prometheus metrics without public Basic Auth.
3. Add Fly `[metrics]` config pointing at the internal metrics port and path.
4. Deploy and verify:
   - public `/metrics` remains Basic Auth protected
   - Fly/Grafana receives custom metrics
   - the app still auto-stops under `min_machines_running = 0`
   - a stopped Machine is not kept awake by metrics scraping
5. Create dashboard queries for route health, DB pool health, AI chat client outcomes, AI chat stream milestones, provider/model duration, and persistence duration.
6. Add alert routing to Discord and email once the alerting backend is chosen.

Open validation items:

- Fly docs do not clearly guarantee whether custom metrics scrapes wake auto-stopped Machines; verify empirically.
- Fly docs do not clearly guarantee 30-day free retention for managed Prometheus/Grafana; accept free/default retention unless a no-cost longer option is confirmed.
- Fly managed Grafana may be dashboard-only for this use case; confirm alerting support before relying on it.

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

Do not widen access on aggregate "chat error rate" alone.

Widen only when all are true:

- `stream=server_error` stays within the agreed beta threshold
- `recovery=recovered_completed` remains healthy relative to `recovered_failed|recovery_timeout`
- `load=load_failed` is stable
- `ux=failure_toast_shown` is not climbing faster than underlying true failures
