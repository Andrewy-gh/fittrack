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

1. Rename ambiguous AI metric wording from `runtime` to `provider` or `model` so it is not confused with the Go runtime. Done in code as `ai_chat_model_duration_seconds`.
2. Add an internal metrics listener on a non-public port that serves Prometheus metrics without public Basic Auth. Done in code on `METRICS_PORT`, default `9091`.
3. Add Fly `[metrics]` config pointing at the internal metrics port and path. Done in `fly.toml` with `port = 9091` and `path = "/metrics"`.
4. Deploy and verify:
   - public `/metrics` remains Basic Auth protected
   - Fly/Grafana receives custom metrics
   - the app still auto-stops under `min_machines_running = 0`
   - a stopped Machine is not kept awake by metrics scraping
5. Create dashboard queries for route health, DB pool health, AI chat client outcomes, AI chat stream milestones, provider/model duration, and persistence duration.
6. Add alert routing to Discord and email once the alerting backend is chosen.

Parallel work lanes used for this rollout:

- Metrics naming: rename `ai_chat_runtime_*` wording to model/provider naming, then update focused tests and dashboard docs.
- Internal scraping: add the separate metrics listener and Fly `[metrics]` config without changing public `/metrics` auth.
- Dashboard and alert prep: prepare PromQL panels and alert candidates without adding new code or exposing sensitive labels.
- Verification: keep deploy, auth, custom metric ingestion, and auto-stop behavior checks as explicit rollout gates.

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

Route 5xx rate by route:

```promql
sum by (path) (rate(http_requests_total{status=~"5.."}[15m]))
```

Route p95 latency:

```promql
histogram_quantile(0.95, sum by (le, path) (rate(http_request_duration_seconds_bucket[15m])))
```

DB active connections:

```promql
max(db_connections_active)
```

DB idle connections:

```promql
max(db_connections_idle)
```

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

Stream milestone p95 by milestone:

```promql
histogram_quantile(0.95, sum by (le, milestone) (rate(ai_chat_stream_milestone_duration_seconds_bucket[15m])))
```

Stream lifecycle events:

```promql
sum by (event) (rate(ai_chat_stream_events_total[15m]))
```

Model/provider p95 duration:

```promql
histogram_quantile(0.95, sum by (le, operation, result) (rate(ai_chat_model_duration_seconds_bucket[15m])))
```

Persistence p95 duration:

```promql
histogram_quantile(0.95, sum by (le, operation, result) (rate(ai_chat_persistence_duration_seconds_bucket[15m])))
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

Alerting backend decision:

- First choice: Fly managed Grafana, if contact points and alert rules are available for the organization and can send Discord plus email.
- Fallback: Grafana Cloud Free with the Fly Prometheus datasource and Discord/email contact points.
- Last-resort fallback: a small Alertmanager-style service only if managed options cannot route both channels without keeping extra compute awake.

Alert text must name only metric categories, outcomes, stages, and cohort. Do not include prompts, generated workout content, user IDs, workout IDs, emails, or request IDs.

## Verification Checklist

After deploy:

1. Confirm public metrics still require Basic Auth:

```sh
curl -i https://fittrack.fly.dev/metrics
curl -i -u "$METRICS_USERNAME:$METRICS_PASSWORD" https://fittrack.fly.dev/metrics
```

2. Confirm Fly receives custom metrics after at least one scrape interval:

```sh
flyctl orgs list
TOKEN=$(flyctl auth token)
curl "https://api.fly.io/prometheus/$ORG_SLUG/api/v1/query" \
  --data-urlencode 'query=ai_chat_model_duration_seconds_count' \
  -H "Authorization: Bearer $TOKEN"
```

3. Confirm auto-stop still works with `min_machines_running = 0`:

```sh
flyctl status --app fittrack
flyctl machine list --app fittrack
```

4. Validate whether metrics scraping wakes stopped Machines:

- Let the app go idle until all Machines are stopped.
- Do not send public traffic.
- Wait at least two custom metrics scrape intervals.
- Re-run `flyctl machine list --app fittrack`.
- Passing result: Machines remain stopped, or no new app request traffic appears.
- Failing result: a stopped Machine starts without public traffic; stop rollout and revisit the scrape approach.

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
