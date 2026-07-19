# AI Chat Health Triage

Use this read-only runbook with a lookback such as `15m` or `1h`. Grafana Cloud Metrics (Mimir) is the metrics source of truth. Keep credentials only in the existing `MIMIR_ADDRESS`, `MIMIR_API_USER`, and `MIMIR_API_KEY` environment variables; never paste or print them. See [AI Chat Observability](ai-chat-observability.md) for metric semantics and dashboard queries, [AI Chat Alerting Runbook](ai-chat-alerting-runbook.md) for access details, and [Production Triage](production-triage.md) for fuller Fly guidance.

This is symptom triage, not full root-cause attribution: without traces or request correlation, do not claim that correlated client, HTTP, deploy, or runtime evidence proves a cause.

## Collect evidence

Set the requested lookback in `WINDOW` (for example, `15m`) in your shell without recording its environment. Set `MIMIR_QUERY_URL` to `${MIMIR_ADDRESS}/prometheus/api/v1/query` as described in the alerting runbook. The following queries and commands are read-only.

Query scrape health, client outcomes, and server HTTP failures. Save the command/query and summarize what its result showed; do not include credentials in the report.

```sh
curl --fail --silent --show-error -u "${MIMIR_API_USER}:${MIMIR_API_KEY}" --get "${MIMIR_QUERY_URL}" --data-urlencode 'query=up'

curl --fail --silent --show-error -u "${MIMIR_API_USER}:${MIMIR_API_KEY}" --get "${MIMIR_QUERY_URL}" --data-urlencode "query=sum by (category, outcome, stage, cohort) (increase(ai_chat_client_outcomes_total[${WINDOW}]))"

curl --fail --silent --show-error -u "${MIMIR_API_USER}:${MIMIR_API_KEY}" --get "${MIMIR_QUERY_URL}" --data-urlencode "query=sum by (path, status) (increase(http_requests_total{status=~\"5..\"}[${WINDOW}]))"
```

`http_requests_total{method,path,status}` and `http_request_duration_seconds{method,path,status}` are the server HTTP metrics implemented today. Use the `path` label to distinguish AI chat routes; do not invent route-specific metric names. Interpret an absent client series carefully: it can mean no traffic, a scrape problem, or missing instrumentation. The outcome taxonomy is:

- `stream`: `completed`, `server_error`, `transport_ended_pre_terminal`, `client_aborted`; stage is `pre_start`, `post_start`, or `terminal`.
- `recovery`: `recovered_completed`, `recovered_failed`, `recovery_timeout`, `recovery_aborted`.
- `load`: `load_completed`, `load_failed`, `load_aborted_stale`.
- `ux`: `failure_toast_shown`, `failure_toast_suppressed_due_to_successful_recovery`.

Check runtime and recent deployment state:

```sh
flyctl status -a fittrack
flyctl checks list -a fittrack
flyctl machine list -a fittrack
flyctl releases -a fittrack
gh run list --workflow fly-deploy.yml
flyctl logs -a fittrack --no-tail
```

Focus logs and deploy history on the requested window where the tools permit, and report that limitation otherwise. In Grafana Cloud Alerting, `severity="page"` alerts route to the `fittrack-email-page` email contact point; a page means one of those page-worthy rules fired, not that root cause is known.

## Choose exactly one category

Use the strongest direct evidence and select exactly one value:

- `healthy`: scrape target is up, Fly checks/machines are healthy, no relevant deployment failure, and client outcomes plus HTTP 5xx show no meaningful degradation.
- `client_load_problem`: `load_failed` rises materially (distinguish `load_aborted_stale`, which is not itself a failure).
- `stream_problem`: `server_error` or `transport_ended_pre_terminal` rises materially; use stage and recovery outcomes to describe impact.
- `backend_http_problem`: relevant AI chat routes show HTTP 5xx growth, especially when client evidence is absent or secondary.
- `metrics_or_scrape_problem`: `up == 0`, the target/series is absent unexpectedly, or metrics disappear while Fly application health remains good.
- `infra_runtime_problem`: a machine/check is down, a release or deploy workflow failed, or runtime logs show an availability failure.
- `unknown_needs_manual_debug`: evidence is missing, contradictory, ambiguous, or requires request-level correlation.

Do not force `healthy` from an empty series or choose multiple categories. Mention secondary signals in evidence, not as additional categories.

## Output template

```text
Window: <15m|1h|...>
Overall health: <healthy|degraded|unavailable|indeterminate, with one-sentence summary>
Likely user impact: <who experienced what; say none observed or unknown when appropriate>
Evidence:
- <exact read-only query/command> -> <what it showed>
- <exact read-only query/command> -> <what it showed>
Category: <exactly one allowed category>
Next actions:
1. <one concrete action>
2. <optional second action>
Limit: Symptom triage only; no full root-cause attribution without traces/request correlation.
```
