# AI Chat Alerting Runbook

This runbook makes the AI chat client-outcome alerts loadable into Grafana Cloud Metrics, which uses Mimir for Prometheus-compatible rule evaluation.

The repo-owned rule file is `ops/grafana-cloud/ai-chat-alerts.mimir.yml`. It keeps the same paging semantics as `ops/prometheus/ai-chat-alerts.yml`:

- page on true server stream failures, recovery timeouts, conversation load failures, and recovery success regression
- do not page on user-driven stream aborts, stale load aborts, recovery aborts, or successful-recovery toast suppression

## Required Access

Keep credentials out of the repo. Set them only in the shell used to verify or load rules:

```sh
export MIMIR_ADDRESS="https://prometheus-<region>.grafana.net"
export MIMIR_API_USER="<grafana-cloud-instance-id>"
export MIMIR_API_KEY="<grafana-cloud-access-policy-token>"
export MIMIR_QUERY_URL="${MIMIR_ADDRESS}/prometheus/api/v1/query"
```

The token needs enough Grafana Cloud Metrics permissions to query metrics and manage rules.

## Confirm Fly Is Scraping The Metric

Confirm Fly still has the internal metrics scrape configured:

```sh
flyctl config show -a fittrack --toml
```

Expected config:

```toml
[metrics]
  port = 9091
  path = "/metrics"
```

After using the beta AI chat flow, query Grafana Cloud Metrics for the counter from Grafana Explore or the Prometheus-compatible query endpoint:

```sh
curl -u "${MIMIR_API_USER}:${MIMIR_API_KEY}" \
  --get "${MIMIR_QUERY_URL}" \
  --data-urlencode 'query=ai_chat_client_outcomes_total'
```

The result should include `category`, `outcome`, `stage`, and `cohort` labels.

## Load And Verify Rules

Validate the rule file before loading:

```sh
mimirtool rules lint ops/grafana-cloud/ai-chat-alerts.mimir.yml
```

Preview the remote change:

```sh
mimirtool rules diff ops/grafana-cloud/ai-chat-alerts.mimir.yml \
  --address="${MIMIR_ADDRESS}" \
  --id="${MIMIR_API_USER}" \
  --key="${MIMIR_API_KEY}"
```

Load the rules:

```sh
mimirtool rules load ops/grafana-cloud/ai-chat-alerts.mimir.yml \
  --address="${MIMIR_ADDRESS}" \
  --id="${MIMIR_API_USER}" \
  --key="${MIMIR_API_KEY}"
```

Confirm the namespace and group are present:

```sh
mimirtool rules list \
  --address="${MIMIR_ADDRESS}" \
  --id="${MIMIR_API_USER}" \
  --key="${MIMIR_API_KEY}"
```

Expected entry:

```text
ai_chat_rollout | ai-chat-rollout
```

Confirm Mimir is evaluating the loaded alerts:

```sh
mimirtool rules print \
  --address="${MIMIR_ADDRESS}" \
  --id="${MIMIR_API_USER}" \
  --key="${MIMIR_API_KEY}"
```

In Grafana Cloud, open Alerting, then Alert rules, and confirm the `ai_chat_rollout` namespace shows the four AI chat rollout alerts.

## Notification Routing

These rules label page-worthy alerts with `severity="page"`. Configure notification routing in Grafana Cloud Alerting so that this label reaches the intended on-call contact point.

Before enabling a production page route, confirm:

- the `severity="page"` matcher routes to the expected contact point
- the contact point is owned by the current FitTrack operator
- no broad catch-all route sends beta rollout alerts to an unintended audience

Do not commit contact point secrets, webhook URLs, phone numbers, or API keys.
